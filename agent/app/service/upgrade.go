package service

import (
	"fmt"
	"os"
	"path"
	"sort"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
)

type UpgradeService struct{}

type IUpgradeService interface {
	Upgrade(req dto.Upgrade) error
	Rollback(req dto.Rollback) error
}

func NewIUpgradeService() IUpgradeService {
	return &UpgradeService{}
}

func (u *UpgradeService) Upgrade(req dto.Upgrade) error {
	fileOP := files.NewFileOp()
	if !fileOP.Stat(req.UpgradePath) {
		global.LOG.Errorf("handle upgrade 1panel to %s failed, err: no such file %s", req.Version, req.UpgradePath)
		return fmt.Errorf("no such upgrade file %s", req.UpgradePath)
	}

	_, _ = cmd.Execf("rm -rf %s", path.Join(global.CONF.System.BaseDir, "1panel/tmp/upgrade", req.Version, "upgrade_*"))

	backupDir := path.Join(global.CONF.System.BaseDir, "1panel/tmp/upgrade", req.Version, fmt.Sprintf("upgrade_%s", time.Now().Format(constant.DateTimeSlimLayout)))
	_ = os.MkdirAll(backupDir, os.ModePerm)
	if err := u.handleBackup(backupDir, fileOP); err != nil {
		return err
	}
	if err := fileOP.CopyFile(path.Join(req.UpgradePath, "1panel-agent.service"), "/etc/systemd/system/1panel-agent.service"); err != nil {
		_ = u.handleRollback(backupDir, 1, fileOP)
		global.LOG.Errorf("handle upgrade 1panel service file failed, err: %v", err)
		return err
	}
	if err := fileOP.CopyFile(path.Join(req.UpgradePath, "1panel-agent"), "/usr/local/bin"); err != nil {
		_ = u.handleRollback(backupDir, 2, fileOP)
		global.LOG.Errorf("handle upgrade 1panel failed, err: %v", err)
		return err
	}
	global.LOG.Info("upgrade successful!")
	_ = settingRepo.Update("SystemVersion", req.Version)
	_, _ = cmd.ExecWithTimeOut("systemctl daemon-reload && systemctl restart 1panel.service", 1*time.Minute)
	return nil
}

func (u *UpgradeService) Rollback(req dto.Rollback) error {
	rollbackDir := path.Join(global.CONF.System.BaseDir, "1panel/tmp/upgrade", req.Version)
	pathItem := loadRestorePath(rollbackDir)
	if len(pathItem) == 0 {
		return fmt.Errorf("no such rollback file in %s", rollbackDir)
	}
	fileOP := files.NewFileOp()
	return u.handleRollback(pathItem, 2, fileOP)
}

func (u *UpgradeService) handleBackup(backupDir string, fileOP files.FileOp) error {
	if err := fileOP.CopyDir(path.Join(global.CONF.System.BaseDir, "1panel/db"), backupDir); err != nil {
		global.LOG.Errorf("handle backup original db file failed, err: %v", err)
		return err
	}
	if err := fileOP.CopyFile("/usr/local/bin/1panel-agent", backupDir); err != nil {
		global.LOG.Errorf("handle backup 1panel binary file failed, err: %v", err)
		return err
	}
	if err := fileOP.CopyFile("/etc/systemd/system/1panel-agent.service", backupDir); err != nil {
		global.LOG.Errorf("handle backup 1panel service file failed, err: %v", err)
		return err
	}
	return nil
}

func (u *UpgradeService) handleRollback(backupDir string, errStep int, fileOP files.FileOp) error {
	if errStep == 1 {
		if err := fileOP.CopyFile(path.Join(backupDir, "1panel-agent.service"), "/etc/systemd/system/1panel-agent.service"); err != nil {
			global.LOG.Errorf("handle recover 1panel service file failed, err: %v", err)
			return err
		}
		return nil
	}
	if errStep == 2 {
		if err := fileOP.CopyFile(path.Join(backupDir, "1panel-agent"), "/usr/local/bin/1panel-agent"); err != nil {
			global.LOG.Errorf("handle recover 1panel service file failed, err: %v", err)
			return err
		}
		if err := fileOP.CopyDir(path.Join(backupDir, "db"), path.Join(global.CONF.System.BaseDir, "1panel")); err != nil {
			global.LOG.Errorf("handle recover 1panel db file failed, err: %v", err)
			return err
		}
	}
	return nil
}

func loadRestorePath(upgradeDir string) string {
	if _, err := os.Stat(upgradeDir); err != nil && os.IsNotExist(err) {
		return ""
	}
	files, err := os.ReadDir(upgradeDir)
	if err != nil {
		return ""
	}
	var folders []string
	for _, file := range files {
		if file.IsDir() {
			folders = append(folders, file.Name())
		}
	}
	if len(folders) == 0 {
		return ""
	}
	sort.Slice(folders, func(i, j int) bool {
		return folders[i] > folders[j]
	})
	return folders[0]
}
