package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
)

type UpgradeService struct{}

type IUpgradeService interface {
	Upgrade(req dto.Upgrade) error
	LoadNotes(req dto.Upgrade) (string, error)
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

	latestVersion, err := u.loadVersion(true, currentVersion.Value)
	if err != nil {
		global.LOG.Infof("load latest version failed, err: %v", err)
		return nil, err
	}
	if !common.CompareVersion(string(latestVersion), currentVersion.Value) {
		return nil, err
	}
	upgrade.LatestVersion = latestVersion
	if latestVersion[0:4] == currentVersion.Value[0:4] {
		upgrade.NewVersion = ""
	} else {
		newerVersion, err := u.loadVersion(false, currentVersion.Value)
		if err != nil {
			global.LOG.Infof("load newer version failed, err: %v", err)
			return nil, err
		}
		if newerVersion == currentVersion.Value {
			upgrade.NewVersion = ""
		} else {
			upgrade.NewVersion = newerVersion
		}
	}
	itemVersion := upgrade.LatestVersion
	if upgrade.NewVersion != "" {
		itemVersion = upgrade.NewVersion
	}
	notes, err := u.loadReleaseNotes(fmt.Sprintf("%s/%s/%s/release/1panel-%s-release-notes", global.CONF.System.RepoUrl, global.CONF.System.Mode, itemVersion, itemVersion))

	if err != nil {
		return nil, fmt.Errorf("load releases-notes of version %s failed, err: %v", latestVersion, err)
	}
	upgrade.ReleaseNote = notes
	return &upgrade, nil
}

func (u *UpgradeService) LoadNotes(req dto.Upgrade) (string, error) {
	notes, err := u.loadReleaseNotes(fmt.Sprintf("%s/%s/%s/release/1panel-%s-release-notes", global.CONF.System.RepoUrl, global.CONF.System.Mode, req.Version, req.Version))
	if err != nil {
		return "", fmt.Errorf("load releases-notes of version %s failed, err: %v", req.Version, err)
	}
	return notes, nil
}

func (u *UpgradeService) Upgrade(req dto.Upgrade) error {
	global.LOG.Info("start to upgrade now...")
	fileOp := files.NewFileOp()
	timeStr := time.Now().Format("20060102150405")
	rootDir := path.Join(global.CONF.System.TmpDir, fmt.Sprintf("upgrade/upgrade_%s/downloads", timeStr))
	originalDir := path.Join(global.CONF.System.TmpDir, fmt.Sprintf("upgrade/upgrade_%s/original", timeStr))
	if err := os.MkdirAll(rootDir, os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(originalDir, os.ModePerm); err != nil {
		return err
	}
	itemArch, err := loadArch()
	if err != nil {
		return err
	}

	downloadPath := fmt.Sprintf("%s/%s/%s/release", global.CONF.System.RepoUrl, global.CONF.System.Mode, req.Version)
	fileName := fmt.Sprintf("1panel-%s-%s-%s.tar.gz", req.Version, "linux", itemArch)
	_ = settingRepo.Update("SystemStatus", "Upgrading")
	go func() {
		_ = global.Cron.Stop()
		defer func() {
			global.Cron.Start()
		}()
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

		if err := cpBinary([]string{tmpDir + "/1panel"}, "/usr/local/bin/1panel"); err != nil {
			global.LOG.Errorf("upgrade 1panel failed, err: %v", err)
			u.handleRollback(originalDir, 1)
			return
		}

		if err := cpBinary([]string{tmpDir + "/1pctl"}, "/usr/local/bin/1pctl"); err != nil {
			global.LOG.Errorf("upgrade 1pctl failed, err: %v", err)
			u.handleRollback(originalDir, 2)
			return
		}
		if _, err := cmd.Execf("sed -i -e 's#BASE_DIR=.*#BASE_DIR=%s#g' /usr/local/bin/1pctl", global.CONF.System.BaseDir); err != nil {
			global.LOG.Errorf("upgrade basedir in 1pctl failed, err: %v", err)
			u.handleRollback(originalDir, 2)
			return
		}

		if err := cpBinary([]string{tmpDir + "/1panel.service"}, "/etc/systemd/system/1panel.service"); err != nil {
			global.LOG.Errorf("upgrade 1panel.service failed, err: %v", err)
			u.handleRollback(originalDir, 3)
			return
		}

		global.LOG.Info("upgrade successful!")
		go writeLogs(req.Version)
		_ = settingRepo.Update("SystemVersion", req.Version)
		_ = settingRepo.Update("SystemStatus", "Free")
		checkPointOfWal()
		_, _ = cmd.ExecWithTimeOut("systemctl daemon-reload && systemctl restart 1panel.service", 1*time.Minute)
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

func (u *UpgradeService) handleRollback(originalDir string, errStep int) {
	dbPath := global.CONF.System.DbPath + "/1Panel.db"
	_ = settingRepo.Update("SystemStatus", "Free")
	if err := cpBinary([]string{originalDir + "/1Panel.db"}, dbPath); err != nil {
		global.LOG.Errorf("rollback 1panel failed, err: %v", err)
	}
	if err := cpBinary([]string{originalDir + "/1panel"}, "/usr/local/bin/1panel"); err != nil {
		global.LOG.Errorf("rollback 1pctl failed, err: %v", err)
	}
	if errStep == 1 {
		return
	}
	if err := cpBinary([]string{originalDir + "/1pctl"}, "/usr/local/bin/1pctl"); err != nil {
		global.LOG.Errorf("rollback 1panel failed, err: %v", err)
	}
	if errStep == 2 {
		return
	}
	if err := cpBinary([]string{originalDir + "/1panel.service"}, "/etc/systemd/system/1panel.service"); err != nil {
		global.LOG.Errorf("rollback 1panel failed, err: %v", err)
	}
}

func (u *UpgradeService) loadVersion(isLatest bool, currentVersion string) (string, error) {
	path := fmt.Sprintf("%s/%s/latest", global.CONF.System.RepoUrl, global.CONF.System.Mode)
	if !isLatest {
		path = fmt.Sprintf("%s/%s/latest.current", global.CONF.System.RepoUrl, global.CONF.System.Mode)
	}
	latestVersionRes, err := http.Get(path)
	if err != nil {
		return "", buserr.New(constant.ErrOSSConn)
	}
	defer latestVersionRes.Body.Close()
	version, err := io.ReadAll(latestVersionRes.Body)
	if err != nil {
		return "", buserr.New(constant.ErrOSSConn)
	}
	if isLatest {
		return string(version), nil
	}

	versionMap := make(map[string]string)
	if err := json.Unmarshal(version, &versionMap); err != nil {
		return "", buserr.New(constant.ErrOSSConn)
	}

	if len(currentVersion) < 4 {
		return "", fmt.Errorf("current version is error format: %s", currentVersion)
	}
	if version, ok := versionMap[currentVersion[0:4]]; ok {
		return version, nil
	}
	return "", buserr.New(constant.ErrOSSConn)
}

func (u *UpgradeService) loadReleaseNotes(path string) (string, error) {
	releaseNotes, err := http.Get(path)
	if err != nil {
		return "", err
	}
	defer releaseNotes.Body.Close()
	release, err := io.ReadAll(releaseNotes.Body)
	if err != nil {
		return "", err
	}
	return string(release), nil
}

func loadArch() (string, error) {
	std, err := cmd.Exec("uname -a")
	if err != nil {
		return "", fmt.Errorf("std: %s, err: %s", std, err.Error())
	}
	if strings.Contains(std, "x86_64") {
		return "amd64", nil
	}
	if strings.Contains(std, "arm64") || strings.Contains(std, "aarch64") {
		return "arm64", nil
	}
	if strings.Contains(std, "armv7l") {
		return "armv7", nil
	}
	if strings.Contains(std, "ppc64le") {
		return "ppc64le", nil
	}
	if strings.Contains(std, "s390x") {
		return "s390x", nil
	}
	return "", fmt.Errorf("unsupported such arch: %s", std)
}
