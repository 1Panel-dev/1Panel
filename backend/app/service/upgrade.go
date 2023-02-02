package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
)

type latestVersion struct {
	Version    string `json:"version"`
	UpdateTime string `json:"update_time"`
}

type UpgradeService struct{}

type IUpgradeService interface {
	Upgrade(version string) error
	SearchUpgrade() (*dto.UpgradeInfo, error)
}

func NewIUpgradeService() IUpgradeService {
	return &UpgradeService{}
}

func (u *UpgradeService) SearchUpgrade() (*dto.UpgradeInfo, error) {
	res, err := http.Get(global.CONF.System.AppOss + "/releases/latest.json")
	if err != nil {
		return nil, err
	}
	resByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var latest latestVersion
	if err := json.Unmarshal(resByte, &latest); err != nil {
		return nil, err
	}
	setting, err := settingRepo.Get(settingRepo.WithByKey("SystemVersion"))
	if err != nil {
		return nil, err
	}
	if latest.Version != setting.Value {
		notes, err := http.Get(global.CONF.System.AppOss + fmt.Sprintf("/releases/%s/release_notes.md", latest.Version))
		if err != nil {
			return nil, err
		}
		noteBytes, err := ioutil.ReadAll(notes.Body)
		if err != nil {
			return nil, err
		}
		return &dto.UpgradeInfo{
			NewVersion:  latest.Version,
			CreatedAt:   latest.UpdateTime,
			ReleaseNote: string(noteBytes),
		}, nil
	}
	return nil, nil
}

func (u *UpgradeService) Upgrade(version string) error {
	global.LOG.Info("start to upgrade now...")
	fileOp := files.NewFileOp()
	timeStr := time.Now().Format("20060102150405")
	filePath := fmt.Sprintf("%s/%s.tar.gz", constant.TmpDir, timeStr)
	rootDir := constant.TmpDir + "/" + timeStr
	originalDir := fmt.Sprintf("%s/%s/original", constant.TmpDir, timeStr)
	downloadPath := fmt.Sprintf("%s/releases/%s/%s.tar.gz", global.CONF.System.AppOss, version, version)

	_ = settingRepo.Update("SystemStatus", "Upgrading")
	go func() {
		if err := os.MkdirAll(originalDir, os.ModePerm); err != nil {
			global.LOG.Error(err.Error())
			return
		}
		if err := fileOp.DownloadFile(downloadPath, filePath); err != nil {
			global.LOG.Errorf("download file failed, err: %v", err)
			return
		}
		global.LOG.Info("download file from oss successful!")
		defer func() {
			_ = os.Remove(filePath)
		}()
		if err := fileOp.Decompress(filePath, rootDir, files.TarGz); err != nil {
			global.LOG.Errorf("decompress file failed, err: %v", err)
			return
		}

		if err := u.handleBackup(fileOp, originalDir); err != nil {
			global.LOG.Errorf("handle backup original file failed, err: %v", err)
			return
		}
		global.LOG.Info("backup original data successful, now start to upgrade!")

		if err := cpBinary(rootDir+"/1panel", "/usr/local/bin/1panel"); err != nil {
			u.handleRollback(fileOp, originalDir, 1)
			global.LOG.Errorf("upgrade 1panel failed, err: %v", err)
			return
		}
		if err := cpBinary(rootDir+"/1pctl", "/usr/local/bin/1pctl"); err != nil {
			u.handleRollback(fileOp, originalDir, 2)
			global.LOG.Errorf("upgrade 1pctl failed, err: %v", err)
			return
		}
		if err := cpBinary(rootDir+"/1panel.service", "/etc/systemd/system/1panel.service"); err != nil {
			u.handleRollback(fileOp, originalDir, 3)
			global.LOG.Errorf("upgrade 1panel.service failed, err: %v", err)
			return
		}

		global.LOG.Info("upgrade successful!")
		_ = settingRepo.Update("SystemStatus", "Upgrade")
		_ = settingRepo.Update("SystemVersion", version)
		_, _ = cmd.Exec("systemctl restart 1panel.service")
	}()
	return nil
}

func (u *UpgradeService) handleBackup(fileOp files.FileOp, originalDir string) error {
	if err := fileOp.Copy("/usr/local/bin/1panel", originalDir); err != nil {
		return err
	}
	if err := fileOp.Copy("/usr/local/bin/1pctl", originalDir); err != nil {
		return err
	}
	if err := fileOp.Copy("/etc/systemd/system/1panel.service", originalDir); err != nil {
		return err
	}
	dbPath := global.CONF.System.DbPath + "/" + global.CONF.System.DbFile
	if err := fileOp.Copy(dbPath, originalDir); err != nil {
		return err
	}
	return nil
}

func (u *UpgradeService) handleRollback(fileOp files.FileOp, originalDir string, errStep int) {
	dbPath := global.CONF.System.DbPath + "/1Panel.db"
	if err := cpBinary(originalDir+"/1Panel.db", dbPath); err != nil {
		global.LOG.Errorf("rollback 1panel failed, err: %v", err)
	}
	if err := cpBinary(originalDir+"/1panel", "/usr/local/bin/1panel"); err != nil {
		global.LOG.Errorf("rollback 1pctl failed, err: %v", err)
	}
	if errStep == 1 {
		return
	}
	if err := cpBinary(originalDir+"/1pctl", "/usr/local/bin/1pctl"); err != nil {
		global.LOG.Errorf("rollback 1panel failed, err: %v", err)
	}
	if errStep == 2 {
		return
	}
	if err := cpBinary(originalDir+"/1panel.service", "/etc/systemd/system/1panel.service"); err != nil {
		global.LOG.Errorf("rollback 1panel failed, err: %v", err)
	}
}
