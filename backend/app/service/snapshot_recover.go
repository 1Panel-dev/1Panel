package service

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/1Panel-dev/1Panel/backend/utils/systemctl"
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
			if err := u.handleUnTar(fmt.Sprintf("%s/%s.tar.gz", baseDir, snap.Name), baseDir, req.Secret); err != nil {
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
	if err := u.recoverCriticalComponents(snap, isRecover, req, snapFileDir); err != nil {
		updateRecoverStatus(snap.ID, isRecover, "CoreComponents", constant.StatusFailed, err.Error())
		return
	}

	_ = rebuildAllAppInstall()
	restartCompose(path.Join(snapJson.BaseDir, "1panel/docker/compose"))

	global.LOG.Info("recover successful")
	cleanupAfterRecover(snap, isRecover, snapFileDir)

	if err := systemctl.SystemRestart(); err != nil {
		global.LOG.Errorf("1Panel service restart failed: %v", err)
		updateRecoverStatus(snap.ID, isRecover, "FinalRestart", constant.StatusFailed, err.Error())
	} else {
		global.LOG.Info("1Panel service restarted successfully")
	}
}

// 辅助函数：动态恢复核心组件
func (u *SnapshotService) recoverCriticalComponents(snap model.Snapshot, isRecover bool, req dto.SnapshotRecover, snapFileDir string) error {
	// 创建服务句柄
	h, err := systemctl.NewServiceHandle(systemctl.PanelService)
	if err != nil {
		return fmt.Errorf("service handle creation failed: %w", err)
	}

	targetPaths := struct {
		Binary  string
		Ctl     string
		Service string
	}{
		Binary:  "/usr/local/bin/1panel",
		Ctl:     "/usr/local/bin/1pctl",
		Service: "",
	}

	// 获取服务文件路径
	servicePath, err := h.GetServicePath()
	if err != nil {
		return fmt.Errorf("service path resolution failed: %w", err)
	}
	targetPaths.Service = servicePath

	// 构建恢复映射
	criticalFiles := []struct {
		StepName   string
		SrcPath    string
		DestPath   string
		IsCritical bool
	}{
		{"1PanelBinary", path.Join(snapFileDir, "1panel/1panel"), targetPaths.Binary, true},
		{"1PctlBinary", path.Join(snapFileDir, "1panel/1pctl"), targetPaths.Ctl, true},
		{"1PanelService", path.Join(snapFileDir, "1panel/"+filepath.Base(servicePath)), servicePath, true},
	}

	// 批量恢复关键文件
	for _, file := range criticalFiles {
		if req.IsNew || snap.InterruptStep == file.StepName {
			if err := recoverPanel(file.SrcPath, file.DestPath); err != nil {
				return fmt.Errorf("%s recovery failed: %w", file.StepName, err)
			}
			global.LOG.Debugf("recover %s to %s successful!", filepath.Base(file.SrcPath), file.DestPath)
			req.IsNew = true

			// 特殊处理服务文件
			if file.StepName == "1PanelService" {
				if err := postProcessService(h); err != nil {
					global.LOG.Warnf("Service post-processing failed: %v", err)
				}
			}
		}
	}

	// 恢复非关键文件（语言文件）
	if req.IsNew || snap.InterruptStep == "LangFiles" {
		srcLang := path.Join(snapFileDir, "1panel/lang")
		if _, err := os.Stat(srcLang); !os.IsNotExist(err) {
			if _, err := cmd.Execf("cp -r %s %s", srcLang, "/usr/local/bin/"); err != nil {
				global.LOG.Warnf("Lang files recovery warning: %v", err)
			}
		}
	}

	return nil
}

// 辅助函数：服务恢复后处理
func postProcessService(h *systemctl.ServiceHandle) error {
	// 管理器感知操作
	switch h.ManagerName() {
	case "systemd":
		if _, err := systemctl.RunSystemCtl("daemon-reload"); err != nil {
			return fmt.Errorf("systemd daemon-reload failed: %w", err)
		}
	case "openrc":
		if _, err := cmd.Execf("rc-update add %s", systemctl.PanelService.ServiceName["openrc"]); err != nil {
			return fmt.Errorf("openrc service registration failed: %w", err)
		}

	case "sysvinit":
		if _, err := cmd.Execf("/etc/init.d/%s enable", systemctl.PanelService.ServiceName["sysvinit"]); err != nil {
			return fmt.Errorf("sysvinit service registration failed: %w", err)
		}
	}
	return nil
}

// 辅助函数：清理临时文件
func cleanupAfterRecover(snap model.Snapshot, isRecover bool, snapFileDir string) {
	if !isRecover {
		oriPath := fmt.Sprintf("%s/1panel_original/original_%s", global.CONF.System.BaseDir, snap.Name)
		global.LOG.Debugf("remove original backup files: %s", oriPath)
		_ = os.RemoveAll(oriPath)
	} else {
		global.LOG.Debugf("remove temporary files: %s", path.Dir(snapFileDir))
		_ = os.RemoveAll(path.Dir(snapFileDir))
	}
}
func backupBeforeRecover(snap model.Snapshot) error {
	baseDir := fmt.Sprintf("%s/1panel_original/original_%s", global.CONF.System.BaseDir, snap.Name)
	var wg sync.WaitGroup
	var status model.SnapshotStatus
	itemHelper := &snapHelper{SnapID: 0, Status: &status, Wg: &wg, FileOp: files.NewFileOp(), Ctx: context.Background()}

	jsonItem := SnapshotJson{
		BaseDir:       global.CONF.System.BaseDir,
		BackupDataDir: global.CONF.System.Backup,
		PanelDataDir:  path.Join(global.CONF.System.BaseDir, "1panel"),
	}
	_ = os.MkdirAll(path.Join(baseDir, "1panel"), os.ModePerm)
	_ = os.MkdirAll(path.Join(baseDir, "docker"), os.ModePerm)

	wg.Add(4)
	itemHelper.Wg = &wg
	go snapJson(*itemHelper, jsonItem, baseDir)
	go snapPanel(*itemHelper, path.Join(baseDir, "1panel"))
	go snapDaemonJson(*itemHelper, path.Join(baseDir, "docker"))
	go snapBackup(*itemHelper, global.CONF.System.Backup, path.Join(baseDir, "1panel"))
	wg.Wait()
	itemHelper.Status.AppData = constant.StatusDone

	allDone, msg := checkAllDone(status)
	if !allDone {
		return errors.New(msg)
	}
	snapPanelData(*itemHelper, global.CONF.System.BaseDir, path.Join(baseDir, "1panel"))
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
