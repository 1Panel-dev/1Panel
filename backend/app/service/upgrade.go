package service

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
)

type UpgradeService struct{}

type IUpgradeService interface {
	Upgrade(req dto.Upgrade) error
	SearchUpgrade() (*dto.UpgradeInfo, error)
}

func NewIUpgradeService() IUpgradeService {
	return &UpgradeService{}
}

func (u *UpgradeService) SearchUpgrade() (*dto.UpgradeInfo, error) {
	var upgrade dto.UpgradeInfo
	currentVersion, err := settingRepo.Get(settingRepo.WithByKey("SystemVersion"))
	if err != nil {
		return nil, err
	}

	versionRes, err := http.Get(fmt.Sprintf("%s/%s/latest", global.CONF.System.RepoUrl, global.CONF.System.Mode))
	if err != nil {
		return nil, err
	}
	defer versionRes.Body.Close()
	version, err := ioutil.ReadAll(versionRes.Body)
	if err != nil {
		return nil, err
	}
	isNew, err := compareVersion(currentVersion.Value, string(version))
	if !isNew || err != nil {
		return nil, err
	}

	upgrade.NewVersion = string(version)

	releaseNotes, err := http.Get(fmt.Sprintf("%s/%s/%s/release/1panel-%s-release-notes", global.CONF.System.RepoUrl, global.CONF.System.Mode, upgrade.NewVersion, upgrade.NewVersion))
	if err != nil {
		return nil, err
	}
	defer releaseNotes.Body.Close()
	release, err := ioutil.ReadAll(releaseNotes.Body)
	if err != nil {
		return nil, err
	}
	upgrade.ReleaseNote = string(release)

	return &upgrade, nil
}

func (u *UpgradeService) Upgrade(req dto.Upgrade) error {
	global.LOG.Info("start to upgrade now...")
	fileOp := files.NewFileOp()
	timeStr := time.Now().Format("20060102150405")
	rootDir := fmt.Sprintf("%s/upgrade_%s/downloads", global.CONF.System.TmpDir, timeStr)
	originalDir := fmt.Sprintf("%s/upgrade_%s/original", global.CONF.System.TmpDir, timeStr)
	if err := os.MkdirAll(rootDir, os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(originalDir, os.ModePerm); err != nil {
		return err
	}

	downloadPath := fmt.Sprintf("%s/%s/%s/release", global.CONF.System.RepoUrl, global.CONF.System.Mode, req.Version)
	fileName := fmt.Sprintf("1panel-%s-%s-%s.tar.gz", req.Version, "linux", runtime.GOARCH)
	_ = settingRepo.Update("SystemStatus", "Upgrading")
	go func() {
		if err := fileOp.DownloadFile(downloadPath+"/"+fileName, rootDir+"/"+fileName); err != nil {
			global.LOG.Errorf("download service file failed, err: %v", err)
			_ = settingRepo.Update("SystemStatus", "Free")
			return
		}
		global.LOG.Info("download all file successful!")
		defer func() {
			_ = os.Remove(rootDir)
		}()
		if err := handleUnTar(rootDir+"/"+fileName, rootDir); err != nil {
			global.LOG.Errorf("decompress file failed, err: %v", err)
			_ = settingRepo.Update("SystemStatus", "Free")
			return
		}
		tmpDir := rootDir + "/" + strings.ReplaceAll(fileName, ".tar.gz", "")

		if err := u.handleBackup(fileOp, originalDir); err != nil {
			global.LOG.Errorf("handle backup original file failed, err: %v", err)
			_ = settingRepo.Update("SystemStatus", "Free")
			return
		}
		global.LOG.Info("backup original data successful, now start to upgrade!")

		if err := cpBinary(tmpDir+"/1panel", "/usr/local/bin/1panel"); err != nil {
			u.handleRollback(fileOp, originalDir, 1)
			global.LOG.Errorf("upgrade 1panel failed, err: %v", err)
			return
		}

		if err := cpBinary(tmpDir+"/1pctl", "/usr/local/bin/1pctl"); err != nil {
			u.handleRollback(fileOp, originalDir, 2)
			global.LOG.Errorf("upgrade 1pctl failed, err: %v", err)
			return
		}
		if _, err := cmd.Execf("sed -i -e 's#BASE_DIR=.*#BASE_DIR=%s#g' /usr/local/bin/1pctl", global.CONF.System.BaseDir); err != nil {
			u.handleRollback(fileOp, originalDir, 2)
			global.LOG.Errorf("upgrade basedir in 1pctl failed, err: %v", err)
			return
		}

		if err := cpBinary(tmpDir+"/1panel.service", "/etc/systemd/system/1panel.service"); err != nil {
			u.handleRollback(fileOp, originalDir, 3)
			global.LOG.Errorf("upgrade 1panel.service failed, err: %v", err)
			return
		}

		global.LOG.Info("upgrade successful!")
		go writeLogs(req.Version)
		_ = settingRepo.Update("SystemVersion", req.Version)
		_ = settingRepo.Update("SystemStatus", "Free")
		_, _ = cmd.Exec("systemctl daemon-reload && systemctl restart 1panel.service")
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
	_ = settingRepo.Update("SystemStatus", "Free")
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

func compareVersion(version, newVersion string) (bool, error) {
	if version == newVersion {
		return false, nil
	}
	if len(version) == 0 || len(newVersion) == 0 {
		return false, fmt.Errorf("incorrect version or new version entered %v -- %v", version, newVersion)
	}
	versions := strings.Split(strings.ReplaceAll(version, "v", ""), ".")
	if len(versions) != 3 {
		return false, fmt.Errorf("incorrect version input %v", version)
	}
	newVersions := strings.Split(strings.ReplaceAll(newVersion, "v", ""), ".")
	if len(newVersions) != 3 {
		return false, fmt.Errorf("incorrect newVersions input %v", version)
	}
	version1, _ := strconv.Atoi(versions[0])
	newVersion1, _ := strconv.Atoi(newVersions[0])
	if newVersion1 > version1 {
		return true, nil
	} else if newVersion1 == version1 {
		version2, _ := strconv.Atoi(versions[1])
		newVersion2, _ := strconv.Atoi(newVersions[1])
		if newVersion2 > version2 {
			return true, nil
		} else if newVersion2 == version2 {
			version3, _ := strconv.Atoi(versions[2])
			newVersion3, _ := strconv.Atoi(newVersions[2])
			if newVersion3 > version3 {
				return true, nil
			} else {
				return false, nil
			}
		} else {
			return false, nil
		}
	} else {
		return false, nil
	}
}
