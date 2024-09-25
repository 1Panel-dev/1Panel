package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/pkg/errors"
)

type snapRecoverHelper struct {
	FileOp files.FileOp
	Task   *task.Task
}

func (u *SnapshotService) SnapshotRecover(req dto.SnapshotRecover) error {
	global.LOG.Info("start to recover panel by snapshot now")
	snap, err := snapshotRepo.Get(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if hasOs(snap.Name) && !strings.Contains(snap.Name, loadOs()) {
		errInfo := fmt.Sprintf("restoring snapshots(%s) between different server architectures(%s) is not supported", snap.Name, loadOs())
		_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"recover_status": constant.StatusFailed, "recover_message": errInfo})
		return errors.New(errInfo)
	}
	if len(snap.RollbackStatus) != 0 && snap.RollbackStatus != constant.StatusSuccess {
		req.IsNew = true
	}
	if !req.IsNew && (snap.InterruptStep == "RecoverDownload" || snap.InterruptStep == "RecoverDecompress" || snap.InterruptStep == "BackupBeforeRecover") {
		req.IsNew = true
	}
	_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"recover_status": constant.StatusWaiting})
	_ = settingRepo.Update("SystemStatus", "Recovering")

	if len(snap.InterruptStep) == 0 {
		req.IsNew = true
	}
	if len(snap.TaskRecoverID) != 0 {
		req.TaskID = snap.TaskRecoverID
	} else {
		_ = snapshotRepo.Update(snap.ID, map[string]interface{}{"task_recover_id": req.TaskID})
	}
	taskItem, err := task.NewTaskWithOps(snap.Name, task.TaskRecover, task.TaskScopeSnapshot, req.TaskID, snap.ID)
	if err != nil {
		global.LOG.Errorf("new task for create snapshot failed, err: %v", err)
		return err
	}
	rootDir := path.Join(global.CONF.System.TmpDir, "system", snap.Name)
	if _, err := os.Stat(rootDir); err != nil && os.IsNotExist(err) {
		_ = os.MkdirAll(rootDir, os.ModePerm)
	}
	itemHelper := snapRecoverHelper{Task: taskItem, FileOp: files.NewFileOp()}

	go func() {
		_ = global.Cron.Stop()
		defer func() {
			global.Cron.Start()
		}()

		if req.IsNew || snap.InterruptStep == "RecoverDownload" || req.ReDownload {
			taskItem.AddSubTaskWithAlias(
				"RecoverDownload",
				func(t *task.Task) error { return handleDownloadSnapshot(&itemHelper, snap, rootDir) },
				nil,
			)
			req.IsNew = true
		}
		if req.IsNew || snap.InterruptStep == "RecoverDecompress" {
			taskItem.AddSubTaskWithAlias(
				"RecoverDecompress",
				func(t *task.Task) error {
					itemHelper.Task.Log("######################## 2 / 10 ########################")
					itemHelper.Task.LogStart(i18n.GetWithName("RecoverDecompress", snap.Name))
					err := itemHelper.FileOp.TarGzExtractPro(fmt.Sprintf("%s/%s.tar.gz", rootDir, snap.Name), rootDir, req.Secret)
					itemHelper.Task.LogWithStatus(i18n.GetMsgByKey("Decompress"), err)
					return err
				},
				nil,
			)
			req.IsNew = true
		}
		if req.IsNew || snap.InterruptStep == "BackupBeforeRecover" {
			taskItem.AddSubTaskWithAlias(
				"BackupBeforeRecover",
				func(t *task.Task) error { return backupBeforeRecover(snap.Name, &itemHelper) },
				nil,
			)
			req.IsNew = true
		}

		var snapJson SnapshotJson
		taskItem.AddSubTaskWithAlias(
			"Readjson",
			func(t *task.Task) error {
				snapJson, err = readFromJson(path.Join(rootDir, snap.Name), &itemHelper)
				return err
			},
			nil,
		)
		if req.IsNew || snap.InterruptStep == "RecoverApp" {
			taskItem.AddSubTaskWithAlias(
				"RecoverApp",
				func(t *task.Task) error { return recoverAppData(path.Join(rootDir, snap.Name), &itemHelper) },
				nil,
			)
			req.IsNew = true
		}
		if req.IsNew || snap.InterruptStep == "RecoverBaseData" {
			taskItem.AddSubTaskWithAlias(
				"RecoverBaseData",
				func(t *task.Task) error { return recoverBaseData(path.Join(rootDir, snap.Name, "base"), &itemHelper) },
				nil,
			)
			req.IsNew = true
		}
		if req.IsNew || snap.InterruptStep == "RecoverDBData" {
			taskItem.AddSubTaskWithAlias(
				"RecoverDBData",
				func(t *task.Task) error { return recoverDBData(path.Join(rootDir, snap.Name, "db"), &itemHelper) },
				nil,
			)
			req.IsNew = true
		}
		if req.IsNew || snap.InterruptStep == "RecoverBackups" {
			taskItem.AddSubTaskWithAlias(
				"RecoverBackups",
				func(t *task.Task) error {
					itemHelper.Task.Log("######################## 8 / 10 ########################")
					itemHelper.Task.LogStart(i18n.GetWithName("RecoverBackups", snap.Name))
					err := itemHelper.FileOp.TarGzExtractPro(path.Join(rootDir, snap.Name, "/1panel_backup.tar.gz"), snapJson.BackupDataDir, "")
					itemHelper.Task.LogWithStatus(i18n.GetMsgByKey("Decompress"), err)
					return err
				},
				nil,
			)
			req.IsNew = true
		}
		if req.IsNew || snap.InterruptStep == "RecoverPanelData" {
			taskItem.AddSubTaskWithAlias(
				"RecoverPanelData",
				func(t *task.Task) error {
					itemHelper.Task.Log("######################## 9 / 10 ########################")
					itemHelper.Task.LogStart(i18n.GetWithName("RecoverPanelData", snap.Name))
					err := itemHelper.FileOp.TarGzExtractPro(path.Join(rootDir, snap.Name, "/1panel_data.tar.gz"), path.Join(snapJson.BaseDir, "1panel"), "")
					itemHelper.Task.LogWithStatus(i18n.GetMsgByKey("Decompress"), err)
					return err
				},
				nil,
			)
			req.IsNew = true
		}
		taskItem.AddSubTaskWithAlias(
			"RecoverDBData",
			func(t *task.Task) error {
				return restartCompose(path.Join(snapJson.BaseDir, "1panel/docker/compose"), &itemHelper)
			},
			nil,
		)

		if err := taskItem.Execute(); err != nil {
			_ = settingRepo.Update("SystemStatus", "Free")
			_ = snapshotRepo.Update(req.ID, map[string]interface{}{"recover_status": constant.StatusFailed, "recover_message": err.Error(), "interrupt_step": taskItem.Task.CurrentStep})
			return
		}
		_ = os.RemoveAll(rootDir)
		_, _ = cmd.Exec("systemctl daemon-reload && systemctl restart 1panel.service")
	}()
	return nil
}

func handleDownloadSnapshot(itemHelper *snapRecoverHelper, snap model.Snapshot, targetDir string) error {
	itemHelper.Task.Log("######################## 1 / 10 ########################")
	itemHelper.Task.LogStart(i18n.GetMsgByKey("RecoverDownload"))

	account, client, err := NewBackupClientWithID(snap.DownloadAccountID)
	itemHelper.Task.LogWithStatus(i18n.GetWithName("RecoverDownloadAccount", fmt.Sprintf("%s - %s", account.Type, account.Name)), err)
	pathItem := account.BackupPath
	if account.BackupPath != "/" {
		pathItem = strings.TrimPrefix(account.BackupPath, "/")
	}
	filePath := fmt.Sprintf("%s/%s.tar.gz", targetDir, snap.Name)
	_ = os.RemoveAll(filePath)
	_, err = client.Download(path.Join(pathItem, fmt.Sprintf("system_snapshot/%s.tar.gz", snap.Name)), filePath)
	itemHelper.Task.LogWithStatus(i18n.GetMsgByKey("Download"), err)
	return err
}

func backupBeforeRecover(name string, itemHelper *snapRecoverHelper) error {
	itemHelper.Task.Log("######################## 3 / 10 ########################")
	itemHelper.Task.LogStart(i18n.GetMsgByKey("BackupBeforeRecover"))

	rootDir := fmt.Sprintf("%s/1panel_original/original_%s", global.CONF.System.BaseDir, name)
	baseDir := path.Join(rootDir, "base")
	if _, err := os.Stat(baseDir); err != nil {
		_ = os.MkdirAll(baseDir, os.ModePerm)
	}

	err := itemHelper.FileOp.CopyDirWithExclude(path.Join(global.CONF.System.BaseDir, "1panel"), rootDir, []string{"cache", "tmp"})
	itemHelper.Task.LogWithStatus(i18n.GetWithName("SnapCopy", path.Join(global.CONF.System.BaseDir, "1panel")), err)
	if err != nil {
		return err
	}
	err = itemHelper.FileOp.CopyDirWithExclude(global.CONF.System.Backup, rootDir, []string{"system_snapshot"})
	itemHelper.Task.LogWithStatus(i18n.GetWithName("SnapCopy", global.CONF.System.Backup), err)
	if err != nil {
		return err
	}
	err = itemHelper.FileOp.CopyFile("/usr/local/bin/1pctl", baseDir)
	itemHelper.Task.LogWithStatus(i18n.GetWithName("SnapCopy", "/usr/local/bin/1pctl"), err)
	if err != nil {
		return err
	}
	err = itemHelper.FileOp.CopyFile("/usr/local/bin/1panel", baseDir)
	itemHelper.Task.LogWithStatus(i18n.GetWithName("SnapCopy", "/usr/local/bin/1panel"), err)
	if err != nil {
		return err
	}
	err = itemHelper.FileOp.CopyFile("/usr/local/bin/1panel_agent", baseDir)
	itemHelper.Task.LogWithStatus(i18n.GetWithName("SnapCopy", "/usr/local/bin/1panel_agent"), err)
	if err != nil {
		return err
	}
	err = itemHelper.FileOp.CopyFile("/etc/systemd/system/1panel.service", baseDir)
	itemHelper.Task.LogWithStatus(i18n.GetWithName("SnapCopy", "/etc/systemd/system/1panel.service"), err)
	if err != nil {
		return err
	}
	err = itemHelper.FileOp.CopyFile("/etc/systemd/system/1panel_agent.service", baseDir)
	itemHelper.Task.LogWithStatus(i18n.GetWithName("SnapCopy", "/etc/systemd/system/1panel_agent.service"), err)
	if err != nil {
		return err
	}
	err = itemHelper.FileOp.CopyFile("/etc/docker/daemon.json", baseDir)
	itemHelper.Task.LogWithStatus(i18n.GetWithName("SnapCopy", "/etc/docker/daemon.json"), err)
	if err != nil {
		return err
	}
	return nil
}

func readFromJson(rootDir string, itemHelper *snapRecoverHelper) (SnapshotJson, error) {
	itemHelper.Task.Log("######################## 4 / 10 ########################")
	itemHelper.Task.LogStart(i18n.GetMsgByKey("Readjson"))

	snapJsonPath := path.Join(rootDir, "base/snapshot.json")
	var snap SnapshotJson
	_, err := os.Stat(snapJsonPath)
	itemHelper.Task.LogWithStatus(i18n.GetMsgByKey("ReadjsonPath"), err)
	if err != nil {
		return snap, err
	}
	fileByte, err := os.ReadFile(snapJsonPath)
	itemHelper.Task.LogWithStatus(i18n.GetMsgByKey("ReadjsonContent"), err)
	if err != nil {
		return snap, err
	}
	err = json.Unmarshal(fileByte, &snap)
	itemHelper.Task.LogWithStatus(i18n.GetMsgByKey("ReadjsonMarshal"), err)
	if err != nil {
		return snap, err
	}
	return snap, nil
}

func recoverAppData(src string, itemHelper *snapRecoverHelper) error {
	itemHelper.Task.Log("######################## 5 / 10 ########################")
	itemHelper.Task.LogStart(i18n.GetMsgByKey("RecoverApp"))

	if _, err := os.Stat(path.Join(src, "images.tar.gz")); err != nil {
		itemHelper.Task.Log(i18n.GetMsgByKey("RecoverAppEmpty"))
		return nil
	} else {
		std, err := cmd.Execf("docker load < %s", path.Join(src, "images.tar.gz"))
		if err != nil {
			itemHelper.Task.LogFailedWithErr(i18n.GetMsgByKey("RecoverAppImage"), errors.New(std))
			return fmt.Errorf("docker load images failed, err: %v", err)
		}
		itemHelper.Task.LogSuccess(i18n.GetMsgByKey("RecoverAppImage"))
	}

	appInstalls, err := appInstallRepo.ListBy()
	itemHelper.Task.LogWithStatus(i18n.GetMsgByKey("RecoverAppList"), err)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for i := 0; i < len(appInstalls); i++ {
		wg.Add(1)
		appInstalls[i].Status = constant.Rebuilding
		_ = appInstallRepo.Save(context.Background(), &appInstalls[i])
		go func(app model.AppInstall) {
			defer wg.Done()
			dockerComposePath := app.GetComposePath()
			out, err := compose.Down(dockerComposePath)
			if err != nil {
				_ = handleErr(app, err, out)
				return
			}
			out, err = compose.Up(dockerComposePath)
			if err != nil {
				_ = handleErr(app, err, out)
				return
			}
			app.Status = constant.Running
			_ = appInstallRepo.Save(context.Background(), &app)
		}(appInstalls[i])
	}
	wg.Wait()
	return nil
}

func recoverBaseData(src string, itemHelper *snapRecoverHelper) error {
	itemHelper.Task.Log("######################## 6 / 10 ########################")
	itemHelper.Task.LogStart(i18n.GetMsgByKey("SnapBaseInfo"))

	err := itemHelper.FileOp.CopyFile(path.Join(src, "1pctl"), "/usr/local/bin")
	itemHelper.Task.LogWithStatus(i18n.GetWithName("SnapCopy", "/usr/local/bin/1pctl"), err)
	if err != nil {
		return err
	}

	err = itemHelper.FileOp.CopyFile(path.Join(src, "1panel"), "/usr/local/bin")
	itemHelper.Task.LogWithStatus(i18n.GetWithName("SnapCopy", "/usr/local/bin/1panel"), err)
	if err != nil {
		return err
	}
	err = itemHelper.FileOp.CopyFile(path.Join(src, "1panel_agent"), "/usr/local/bin")
	itemHelper.Task.LogWithStatus(i18n.GetWithName("SnapCopy", "/usr/local/bin/1panel_agent"), err)
	if err != nil {
		return err
	}
	err = itemHelper.FileOp.CopyFile(path.Join(src, "1panel.service"), "/etc/systemd/system")
	itemHelper.Task.LogWithStatus(i18n.GetWithName("SnapCopy", "/etc/systemd/system/1panel.service"), err)
	if err != nil {
		return err
	}
	err = itemHelper.FileOp.CopyFile(path.Join(src, "1panel_agent.service"), "/etc/systemd/system")
	itemHelper.Task.LogWithStatus(i18n.GetWithName("SnapCopy", "/etc/systemd/system/1panel_agent.service"), err)
	if err != nil {
		return err
	}

	daemonJsonPath := "/etc/docker/daemon.json"
	_, errSrc := os.Stat(path.Join(src, "docker/daemon.json"))
	_, errPath := os.Stat(daemonJsonPath)
	if os.IsNotExist(errSrc) && os.IsNotExist(errPath) {
		itemHelper.Task.Log(i18n.GetMsgByKey("RecoverDaemonJsonEmpty"))
		return nil
	}
	if errSrc == nil {
		err = itemHelper.FileOp.CopyFile(path.Join(src, "docker/daemon.json"), "/etc/docker")
		itemHelper.Task.Log(i18n.GetMsgByKey("RecoverDaemonJson"))
		if err != nil {
			return fmt.Errorf("recover docker daemon.json failed, err: %v", err)
		}
	}

	_, _ = cmd.Exec("systemctl restart docker")
	return nil
}

func recoverDBData(src string, itemHelper *snapRecoverHelper) error {
	itemHelper.Task.Log("######################## 7 / 10 ########################")
	itemHelper.Task.LogStart(i18n.GetMsgByKey("RecoverDBData"))
	err := itemHelper.FileOp.CopyDirWithExclude(src, path.Join(global.CONF.System.BaseDir, "1panel"), nil)

	itemHelper.Task.LogWithStatus(i18n.GetMsgByKey("RecoverDBData"), err)
	return err
}

func restartCompose(composePath string, itemHelper *snapRecoverHelper) error {
	itemHelper.Task.Log("######################## 10 / 10 ########################")
	itemHelper.Task.LogStart(i18n.GetMsgByKey("RecoverCompose"))

	composes, err := composeRepo.ListRecord()
	itemHelper.Task.LogWithStatus(i18n.GetMsgByKey("RecoverComposeList"), err)
	if err != nil {
		return err
	}

	for _, compose := range composes {
		pathItem := path.Join(composePath, compose.Name, "docker-compose.yml")
		if _, err := os.Stat(pathItem); err != nil {
			continue
		}
		upCmd := fmt.Sprintf("docker compose -f %s up -d", pathItem)
		stdout, err := cmd.Exec(upCmd)
		if err != nil {
			itemHelper.Task.LogFailedWithErr(i18n.GetMsgByKey("RecoverCompose"), errors.New(stdout))
			continue
		}
		itemHelper.Task.LogSuccess(i18n.GetWithName("RecoverComposeItem", pathItem))
	}
	return nil
}
