package service

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/pkg/errors"
)

func (u *SnapshotService) HandleSnapshotRecover(snap model.Snapshot, req dto.SnapshotRecover) {
	_ = global.Cron.Stop()
	defer func() {
		global.Cron.Start()
	}()

	fileOp := files.NewFileOp()
	baseDir := path.Join(global.CONF.System.TmpDir, fmt.Sprintf("system/%s", snap.Name))
	if _, err := os.Stat(baseDir); err != nil && os.IsNotExist(err) {
		_ = os.MkdirAll(baseDir, os.ModePerm)
	}
	if req.IsNew || snap.InterruptStep == "Download" || req.ReDownload {
		if err := handleDownloadSnapshot(snap, baseDir); err != nil {
			updateRecoverStatus(snap.ID, "Download", constant.StatusFailed, err.Error())
			return
		}
		global.LOG.Debugf("download snapshot file to %s successful!", baseDir)
		req.IsNew = true
	}
	if req.IsNew || snap.InterruptStep == "Decompress" {
		if err := fileOp.TarGzExtractPro(fmt.Sprintf("%s/%s.tar.gz", baseDir, snap.Name), baseDir, req.Secret); err != nil {
			updateRecoverStatus(snap.ID, "Decompress", constant.StatusFailed, fmt.Sprintf("decompress file failed, err: %v", err))
			return
		}
		global.LOG.Debug("decompress snapshot file successful!", baseDir)
		req.IsNew = true
	}
	if req.IsNew || snap.InterruptStep == "Backup" {
		if err := backupBeforeRecover(snap.Name); err != nil {
			updateRecoverStatus(snap.ID, "Backup", constant.StatusFailed, fmt.Sprintf("handle backup before recover failed, err: %v", err))
			return
		}
		global.LOG.Debug("handle backup before recover successful!")
		req.IsNew = true
	}
	snapFileDir := fmt.Sprintf("%s/%s", baseDir, snap.Name)
	if _, err := os.Stat(snapFileDir); err != nil {
		snapFileDir = baseDir
	}
	snapJson, err := u.readFromJson(path.Join(snapFileDir, "base/snapshot.json"))
	if err != nil {
		updateRecoverStatus(snap.ID, "Readjson", constant.StatusFailed, fmt.Sprintf("decompress file failed, err: %v", err))
		return
	}
	if snap.InterruptStep == "Readjson" {
		req.IsNew = true
	}
	if req.IsNew || snap.InterruptStep == "AppImage" {
		if err := recoverAppData(snapFileDir); err != nil {
			updateRecoverStatus(snap.ID, "AppImage", constant.StatusFailed, fmt.Sprintf("handle recover app data failed, err: %v", err))
			return
		}
		global.LOG.Debug("recover app images from snapshot file successful!")
		req.IsNew = true
	}

	if req.IsNew || snap.InterruptStep == "BaseData" {
		if err := recoverBaseData(path.Join(snapFileDir, "base"), fileOp); err != nil {
			updateRecoverStatus(snap.ID, "BaseData", constant.StatusFailed, err.Error())
			return
		}
		global.LOG.Debug("recover base data from snapshot file successful!")
		req.IsNew = true
	}

	if req.IsNew || snap.InterruptStep == "DBData" {
		if err := recoverDBData(path.Join(snapFileDir, "db"), fileOp); err != nil {
			updateRecoverStatus(snap.ID, "DBData", constant.StatusFailed, err.Error())
			return
		}
		global.LOG.Debug("recover db data from snapshot file successful!")
		req.IsNew = true
	}

	if req.IsNew || snap.InterruptStep == "1PanelBackups" {
		if err := fileOp.TarGzExtractPro(path.Join(snapFileDir, "/1panel_backup.tar.gz"), snapJson.BackupDataDir, ""); err != nil {
			updateRecoverStatus(snap.ID, "1PanelBackups", constant.StatusFailed, err.Error())
			return
		}
		global.LOG.Debug("recover 1panel backups from snapshot file successful!")
		req.IsNew = true
	}

	if req.IsNew || snap.InterruptStep == "1PanelData" {
		if err := fileOp.TarGzExtractPro(path.Join(snapFileDir, "/1panel_data.tar.gz"), path.Join(snapJson.BaseDir, "1panel"), ""); err != nil {
			updateRecoverStatus(snap.ID, "1PanelData", constant.StatusFailed, err.Error())
			return
		}
		global.LOG.Debug("recover 1panel data from snapshot file successful!")
		req.IsNew = true
	}
	_ = rebuildAllAppInstall()
	restartCompose(path.Join(snapJson.BaseDir, "1panel/docker/compose"))

	global.LOG.Info("recover successful")
	global.LOG.Debugf("remove the file %s after the operation is successful", path.Dir(snapFileDir))
	_ = os.RemoveAll(path.Dir(snapFileDir))
	_, _ = cmd.Exec("systemctl daemon-reload && systemctl restart 1panel.service")
}

func handleDownloadSnapshot(snap model.Snapshot, targetDir string) error {
	account, client, err := NewBackupClientWithID(snap.DownloadAccountID)
	if err != nil {
		return err
	}
	pathItem := account.BackupPath
	if account.BackupPath != "/" {
		pathItem = strings.TrimPrefix(account.BackupPath, "/")
	}
	filePath := fmt.Sprintf("%s/%s.tar.gz", targetDir, snap.Name)
	_ = os.RemoveAll(filePath)
	ok, err := client.Download(path.Join(pathItem, fmt.Sprintf("system_snapshot/%s.tar.gz", snap.Name)), filePath)
	if err != nil || !ok {
		return fmt.Errorf("download file %s from %s failed, err: %v", snap.Name, account.Name, err)
	}
	return nil
}

func recoverAppData(src string) error {
	if _, err := os.Stat(path.Join(src, "images.tar.gz")); err != nil {
		global.LOG.Debug("no such docker images in snapshot")
		return nil
	}
	std, err := cmd.Execf("docker load < %s", path.Join(src, "images.tar.gz"))
	if err != nil {
		return errors.New(std)
	}
	return err
}

func recoverBaseData(src string, fileOp files.FileOp) error {
	if err := fileOp.CopyFile(path.Join(src, "1pctl"), "/usr/local/bin"); err != nil {
		return err
	}
	if err := fileOp.CopyFile(path.Join(src, "1panel"), "/usr/local/bin"); err != nil {
		return err
	}
	if err := fileOp.CopyFile(path.Join(src, "1panel_agent"), "/usr/local/bin"); err != nil {
		return err
	}
	if err := fileOp.CopyFile(path.Join(src, "1panel.service"), "/etc/systemd/system"); err != nil {
		return err
	}
	if err := fileOp.CopyFile(path.Join(src, "1panel_agent.service"), "/etc/systemd/system"); err != nil {
		return err
	}

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

	_, _ = cmd.Exec("systemctl restart docker")
	return nil
}

func recoverDBData(src string, fileOp files.FileOp) error {
	return fileOp.CopyDir(src, path.Join(global.CONF.System.BaseDir, "1panel", "db"))
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
		upCmd := fmt.Sprintf("docker compose -f %s up -d", pathItem)
		stdout, err := cmd.Exec(upCmd)
		if err != nil {
			global.LOG.Debugf("%s failed, err: %v", upCmd, stdout)
		}
	}
	global.LOG.Debug("restart all compose successful!")
}

func backupBeforeRecover(name string) error {
	rootDir := fmt.Sprintf("%s/1panel_original/original_%s", global.CONF.System.BaseDir, name)
	baseDir := path.Join(rootDir, "base")
	if _, err := os.Stat(baseDir); err != nil {
		_ = os.MkdirAll(baseDir, os.ModePerm)
	}

	FileOp := files.NewFileOp()
	if err := FileOp.CopyDirWithExclude(path.Join(global.CONF.System.BaseDir, "1panel"), rootDir, []string{"cache", "tmp"}); err != nil {
		return err
	}
	if err := FileOp.CopyDir(global.CONF.System.Backup, rootDir); err != nil {
		return err
	}
	if err := FileOp.CopyFile("/usr/local/bin/1pctl", baseDir); err != nil {
		return err
	}
	if err := FileOp.CopyFile("/usr/local/bin/1panel", baseDir); err != nil {
		return err
	}
	if err := FileOp.CopyFile("/usr/local/bin/1panel_agent", baseDir); err != nil {
		return err
	}
	if err := FileOp.CopyFile("/etc/systemd/system/1panel.service", baseDir); err != nil {
		return err
	}
	if err := FileOp.CopyFile("/etc/systemd/system/1panel_agent.service", baseDir); err != nil {
		return err
	}
	if err := FileOp.CopyFile("/etc/docker/daemon.json", baseDir); err != nil {
		return err
	}
	return nil
}

func handleRollback(name string) error {
	rootDir := fmt.Sprintf("%s/1panel_original/original_%s", global.CONF.System.BaseDir, name)
	baseDir := path.Join(rootDir, "base")

	FileOp := files.NewFileOp()
	if err := FileOp.CopyDir(path.Join(rootDir, "1panel"), global.CONF.System.BaseDir); err != nil {
		return err
	}
	if err := FileOp.CopyDir(path.Join(rootDir, "backup"), path.Dir(global.CONF.System.Backup)); err != nil {
		return err
	}
	if err := FileOp.CopyFile(path.Join(baseDir, "1pctl"), "/usr/local/bin/1pctl"); err != nil {
		return err
	}
	if err := FileOp.CopyFile(path.Join(baseDir, "1panel"), "/usr/local/bin/1panel"); err != nil {
		return err
	}
	if err := FileOp.CopyFile(path.Join(baseDir, "1panel_agent"), "/usr/local/bin/1panel_agent"); err != nil {
		return err
	}
	if err := FileOp.CopyFile(path.Join(baseDir, "1panel.service"), "/etc/systemd/system/1panel.service"); err != nil {
		return err
	}
	if err := FileOp.CopyFile(path.Join(baseDir, "1panel_agent.service"), "/etc/systemd/system/1panel_agent.service"); err != nil {
		return err
	}
	if err := FileOp.CopyFile(path.Join(baseDir, "daemon.json"), "/etc/docker/daemon.json"); err != nil {
		return err
	}
	_ = os.RemoveAll(rootDir)
	return nil
}

func updateRecoverStatus(id uint, interruptStep, status, message string) {
	if status != constant.StatusSuccess {
		global.LOG.Errorf("recover failed, err: %s", message)
	}
	if err := snapshotRepo.Update(id, map[string]interface{}{
		"interrupt_step":    interruptStep,
		"recover_status":    status,
		"recover_message":   message,
		"last_recovered_at": time.Now().Format(constant.DateTimeLayout),
	}); err != nil {
		global.LOG.Errorf("update snap recover status failed, err: %v", err)
	}
	_ = settingRepo.Update("SystemStatus", "Free")
}
