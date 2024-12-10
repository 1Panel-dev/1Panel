package service

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/pkg/errors"
)

func (u *SnapshotService) HandleSnapshotRecover(snap model.Snapshot, isRecover bool, req dto.SnapshotRecover) {
	_ = global.Cron.Stop()
	defer func() {
		global.Cron.Start()
	}()

	snapFileDir := ""
	if isRecover {
		baseDir := path.Join(global.CONF.System.TmpDir, fmt.Sprintf("system/%s", snap.Name))
		if _, err := os.Stat(baseDir); err != nil && os.IsNotExist(err) {
			_ = os.MkdirAll(baseDir, os.ModePerm)
		}
		if req.IsNew || snap.InterruptStep == "Download" || req.ReDownload {
			if err := handleDownloadSnapshot(snap, baseDir); err != nil {
				updateRecoverStatus(snap.ID, isRecover, "Backup", constant.StatusFailed, err.Error())
				return
			}
			global.LOG.Debugf("download snapshot file to %s successful!", baseDir)
			req.IsNew = true
		}
		if req.IsNew || snap.InterruptStep == "Decompress" {
			if err := handleUnTar(fmt.Sprintf("%s/%s.tar.gz", baseDir, snap.Name), baseDir, req.Secret); err != nil {
				updateRecoverStatus(snap.ID, isRecover, "Decompress", constant.StatusFailed, fmt.Sprintf("decompress file failed, err: %v", err))
				return
			}
			global.LOG.Debug("decompress snapshot file successful!", baseDir)
			req.IsNew = true
		}
		if req.IsNew || snap.InterruptStep == "Backup" {
			if err := backupBeforeRecover(snap); err != nil {
				updateRecoverStatus(snap.ID, isRecover, "Backup", constant.StatusFailed, fmt.Sprintf("handle backup before recover failed, err: %v", err))
				return
			}
			global.LOG.Debug("handle backup before recover successful!")
			req.IsNew = true
		}
		snapFileDir = fmt.Sprintf("%s/%s", baseDir, snap.Name)
		if _, err := os.Stat(snapFileDir); err != nil {
			snapFileDir = baseDir
		}
	} else {
		snapFileDir = fmt.Sprintf("%s/1panel_original/original_%s", global.CONF.System.BaseDir, snap.Name)
		if _, err := os.Stat(snapFileDir); err != nil {
			updateRecoverStatus(snap.ID, isRecover, "", constant.StatusFailed, fmt.Sprintf("cannot find the backup file %s, please try to recover again.", snapFileDir))
			return
		}
	}
	snapJson, err := u.readFromJson(fmt.Sprintf("%s/snapshot.json", snapFileDir))
	if err != nil {
		updateRecoverStatus(snap.ID, isRecover, "Readjson", constant.StatusFailed, fmt.Sprintf("decompress file failed, err: %v", err))
		return
	}
	if snap.InterruptStep == "Readjson" {
		req.IsNew = true
	}
	if isRecover && (req.IsNew || snap.InterruptStep == "AppData") {
		if err := recoverAppData(snapFileDir); err != nil {
			updateRecoverStatus(snap.ID, isRecover, "DockerDir", constant.StatusFailed, fmt.Sprintf("handle recover app data failed, err: %v", err))
			return
		}
		global.LOG.Debug("recover app data from snapshot file successful!")
		req.IsNew = true
	}
	if req.IsNew || snap.InterruptStep == "DaemonJson" {
		fileOp := files.NewFileOp()
		if err := recoverDaemonJson(snapFileDir, fileOp); err != nil {
			updateRecoverStatus(snap.ID, isRecover, "DaemonJson", constant.StatusFailed, err.Error())
			return
		}
		global.LOG.Debug("recover daemon.json from snapshot file successful!")
		req.IsNew = true
	}

	if req.IsNew || snap.InterruptStep == "1PanelBinary" {
		if err := recoverPanel(path.Join(snapFileDir, "1panel/1panel"), "/usr/local/bin"); err != nil {
			updateRecoverStatus(snap.ID, isRecover, "1PanelBinary", constant.StatusFailed, err.Error())
			return
		}
		global.LOG.Debug("recover 1panel binary from snapshot file successful!")
		req.IsNew = true
	}
	if req.IsNew || snap.InterruptStep == "1PctlBinary" {
		if err := recoverPanel(path.Join(snapFileDir, "1panel/1pctl"), "/usr/local/bin"); err != nil {
			updateRecoverStatus(snap.ID, isRecover, "1PctlBinary", constant.StatusFailed, err.Error())
			return
		}
		_, _ = cmd.Execf("cp -r %s %s", path.Join(snapFileDir, "1panel/lang"), "/usr/local/bin/")
		global.LOG.Debug("recover 1pctl from snapshot file successful!")
		req.IsNew = true
	}
	if req.IsNew || snap.InterruptStep == "1PanelService" {
		if err := recoverPanel(path.Join(snapFileDir, "1panel/1panel.service"), "/etc/systemd/system"); err != nil {
			updateRecoverStatus(snap.ID, isRecover, "1PanelService", constant.StatusFailed, err.Error())
			return
		}
		global.LOG.Debug("recover 1panel service from snapshot file successful!")
		req.IsNew = true
	}

	if req.IsNew || snap.InterruptStep == "1PanelBackups" {
		if err := u.handleUnTar(path.Join(snapFileDir, "/1panel/1panel_backup.tar.gz"), snapJson.BackupDataDir, ""); err != nil {
			updateRecoverStatus(snap.ID, isRecover, "1PanelBackups", constant.StatusFailed, err.Error())
			return
		}
		global.LOG.Debug("recover 1panel backups from snapshot file successful!")
		req.IsNew = true
	}

	if req.IsNew || snap.InterruptStep == "1PanelData" {
		checkPointOfWal()
		if err := u.handleUnTar(path.Join(snapFileDir, "/1panel/1panel_data.tar.gz"), path.Join(snapJson.BaseDir, "1panel"), ""); err != nil {
			updateRecoverStatus(snap.ID, isRecover, "1PanelData", constant.StatusFailed, err.Error())
			return
		}
		global.LOG.Debug("recover 1panel data from snapshot file successful!")
		req.IsNew = true
	}
	_ = rebuildAllAppInstall()
	restartCompose(path.Join(snapJson.BaseDir, "1panel/docker/compose"))

	global.LOG.Info("recover successful")
	if !isRecover {
		oriPath := fmt.Sprintf("%s/1panel_original/original_%s", global.CONF.System.BaseDir, snap.Name)
		global.LOG.Debugf("remove the file %s after the operation is successful", oriPath)
		_ = os.RemoveAll(oriPath)
	} else {
		global.LOG.Debugf("remove the file %s after the operation is successful", path.Dir(snapFileDir))
		_ = os.RemoveAll(path.Dir(snapFileDir))
	}
	_, _ = cmd.Exec("systemctl daemon-reload && systemctl restart 1panel.service")
}

func backupBeforeRecover(snap model.Snapshot) error {
	baseDir := fmt.Sprintf("%s/1panel_original/original_%s", global.CONF.System.BaseDir, snap.Name)
	var wg sync.WaitGroup
	var status model.SnapshotStatus
	itemHelper := snapHelper{SnapID: 0, Status: &status, Wg: &wg, FileOp: files.NewFileOp(), Ctx: context.Background()}

	jsonItem := SnapshotJson{
		BaseDir:       global.CONF.System.BaseDir,
		BackupDataDir: global.CONF.System.Backup,
		PanelDataDir:  path.Join(global.CONF.System.BaseDir, "1panel"),
	}
	_ = os.MkdirAll(path.Join(baseDir, "1panel"), os.ModePerm)
	_ = os.MkdirAll(path.Join(baseDir, "docker"), os.ModePerm)

	wg.Add(4)
	itemHelper.Wg = &wg
	go snapJson(itemHelper, jsonItem, baseDir)
	go snapPanel(itemHelper, path.Join(baseDir, "1panel"))
	go snapDaemonJson(itemHelper, path.Join(baseDir, "docker"))
	go snapBackup(itemHelper, global.CONF.System.Backup, path.Join(baseDir, "1panel"))
	wg.Wait()
	itemHelper.Status.AppData = constant.StatusDone

	allDone, msg := checkAllDone(status)
	if !allDone {
		return errors.New(msg)
	}
	snapPanelData(itemHelper, global.CONF.System.BaseDir, path.Join(baseDir, "1panel"))
	if status.PanelData != constant.StatusDone {
		return errors.New(status.PanelData)
	}
	return nil
}

func handleDownloadSnapshot(snap model.Snapshot, targetDir string) error {
	backup, err := backupRepo.Get(commonRepo.WithByType(snap.DefaultDownload))
	if err != nil {
		return err
	}
	client, err := NewIBackupService().NewClient(&backup)
	if err != nil {
		return err
	}
	pathItem := backup.BackupPath
	if backup.BackupPath != "/" {
		pathItem = strings.TrimPrefix(backup.BackupPath, "/")
	}
	filePath := fmt.Sprintf("%s/%s.tar.gz", targetDir, snap.Name)
	_ = os.RemoveAll(filePath)
	ok, err := client.Download(path.Join(pathItem, fmt.Sprintf("system_snapshot/%s.tar.gz", snap.Name)), filePath)
	if err != nil || !ok {
		return fmt.Errorf("download file %s from %s failed, err: %v", snap.Name, backup.Type, err)
	}
	return nil
}

func recoverAppData(src string) error {
	if _, err := os.Stat(path.Join(src, "docker/docker_image.tar")); err != nil {
		global.LOG.Debug("no such docker images in snapshot")
		return nil
	}
	std, err := cmd.Execf("docker load < %s", path.Join(src, "docker/docker_image.tar"))
	if err != nil {
		return errors.New(std)
	}
	return err
}

func recoverDaemonJson(src string, fileOp files.FileOp) error {
	daemonJsonPath := "/etc/docker/daemon.json"
	_, errSrc := os.Stat(path.Join(src, "docker/daemon.json"))
	_, errPath := os.Stat(daemonJsonPath)
	if os.IsNotExist(errSrc) && os.IsNotExist(errPath) {
		global.LOG.Debug("the daemon.json file does not exist, nothing happens.")
		return nil
	}
	if errSrc == nil {
		if err := fileOp.CopyFile(path.Join(src, "docker/daemon.json"), "/etc/docker"); err != nil {
			return fmt.Errorf("recover docker daemon.json failed, err: %v", err)
		}
	}

	if err := restartDocker(); err != nil {
		return err
	}
	return nil
}

func recoverPanel(src string, dst string) error {
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("file is not found in %s, err: %v", src, err)
	}
	if err := common.CopyFile(src, dst); err != nil {
		return fmt.Errorf("cp file failed, err: %v", err)
	}
	return nil
}

func restartCompose(composePath string) {
	composes, err := composeRepo.ListRecord()
	if err != nil {
		return
	}
	for _, compose := range composes {
		pathItem := path.Join(composePath, compose.Name, "docker-compose.yml")
		if _, err := os.Stat(pathItem); err != nil {
			continue
		}
		upCmd := fmt.Sprintf("docker-compose -f %s up -d", pathItem)
		stdout, err := cmd.Exec(upCmd)
		if err != nil {
			global.LOG.Debugf("%s failed, err: %v", upCmd, stdout)
		}
	}
	global.LOG.Debug("restart all compose successful!")
}
