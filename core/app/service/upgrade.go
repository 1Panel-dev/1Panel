package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/cmd"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/1Panel-dev/1Panel/core/utils/files"
	"github.com/1Panel-dev/1Panel/core/utils/req_helper"
)

type UpgradeService struct{}

type IUpgradeService interface {
	Upgrade(req dto.Upgrade) error
	Rollback(req dto.OperateByID) error
	LoadNotes(req dto.Upgrade) (string, error)
	SearchUpgrade() (*dto.UpgradeInfo, error)
}

func NewIUpgradeService() IUpgradeService {
	return &UpgradeService{}
}

func (u *UpgradeService) SearchUpgrade() (*dto.UpgradeInfo, error) {
	var upgrade dto.UpgradeInfo
	currentVersion, err := settingRepo.Get(repo.WithByKey("SystemVersion"))
	if err != nil {
		return nil, err
	}
	DeveloperMode, err := settingRepo.Get(repo.WithByKey("DeveloperMode"))
	if err != nil {
		return nil, err
	}

	upgrade.TestVersion, upgrade.NewVersion, upgrade.LatestVersion = u.loadVersionByMode(DeveloperMode.Value, currentVersion.Value)
	var itemVersion string
	if len(upgrade.LatestVersion) != 0 {
		itemVersion = upgrade.LatestVersion
	}
	if len(upgrade.NewVersion) != 0 {
		itemVersion = upgrade.NewVersion
	}
	if (global.CONF.Base.Mode == "dev" || DeveloperMode.Value == constant.StatusEnable) && len(upgrade.TestVersion) != 0 {
		itemVersion = upgrade.TestVersion
	}
	if len(itemVersion) == 0 {
		return &upgrade, nil
	}
	mode := global.CONF.Base.Mode
	if strings.Contains(itemVersion, "beta") {
		mode = "beta"
	}
	notes, err := u.loadReleaseNotes(fmt.Sprintf("%s/%s/%s/release/1panel-%s-release-notes", global.CONF.RemoteURL.RepoUrl, mode, itemVersion, itemVersion))
	if err != nil {
		return nil, fmt.Errorf("load releases-notes of version %s failed, err: %v", itemVersion, err)
	}
	upgrade.ReleaseNote = notes
	return &upgrade, nil
}

func (u *UpgradeService) LoadNotes(req dto.Upgrade) (string, error) {
	mode := global.CONF.Base.Mode
	if strings.Contains(req.Version, "beta") {
		mode = "beta"
	}
	notes, err := u.loadReleaseNotes(fmt.Sprintf("%s/%s/%s/release/1panel-%s-release-notes", global.CONF.RemoteURL.RepoUrl, mode, req.Version, req.Version))
	if err != nil {
		return "", fmt.Errorf("load releases-notes of version %s failed, err: %v", req.Version, err)
	}
	return notes, nil
}

func (u *UpgradeService) Upgrade(req dto.Upgrade) error {
	global.LOG.Info("start to upgrade now...")
	baseDir := path.Join(global.CONF.Base.InstallDir, fmt.Sprintf("1panel/tmp/upgrade/%s", req.Version))
	downloadDir := path.Join(baseDir, "downloads")
	_ = os.RemoveAll(baseDir)
	originalDir := path.Join(baseDir, "original")
	if err := os.MkdirAll(downloadDir, os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(originalDir, os.ModePerm); err != nil {
		return err
	}
	itemArch, err := loadArch()
	if err != nil {
		return err
	}

	mode := global.CONF.Base.Mode
	if strings.Contains(req.Version, "beta") {
		mode = "beta"
	}
	downloadPath := fmt.Sprintf("%s/%s/%s/release", global.CONF.RemoteURL.RepoUrl, mode, req.Version)
	fileName := fmt.Sprintf("1panel-%s-%s-%s.tar.gz", req.Version, "linux", itemArch)
	_ = settingRepo.Update("SystemStatus", "Upgrading")
	go func() {
		if err := files.DownloadFileWithProxy(downloadPath+"/"+fileName, downloadDir+"/"+fileName); err != nil {
			global.LOG.Errorf("download service file failed, err: %v", err)
			_ = settingRepo.Update("SystemStatus", "Free")
			return
		}
		global.LOG.Info("download all file successful!")
		defer func() {
			_ = os.Remove(downloadDir)
		}()
		if err := files.HandleUnTar(downloadDir+"/"+fileName, downloadDir, ""); err != nil {
			global.LOG.Errorf("decompress file failed, err: %v", err)
			_ = settingRepo.Update("SystemStatus", "Free")
			return
		}
		tmpDir := downloadDir + "/" + strings.ReplaceAll(fileName, ".tar.gz", "")

		if err := u.handleBackup(originalDir); err != nil {
			global.LOG.Errorf("handle backup original file failed, err: %v", err)
			_ = settingRepo.Update("SystemStatus", "Free")
			return
		}
		itemLog := model.UpgradeLog{NodeID: 0, OldVersion: global.CONF.Base.Version, NewVersion: req.Version, BackupFile: baseDir}
		_ = upgradeLogRepo.Create(&itemLog)

		global.LOG.Info("backup original data successful, now start to upgrade!")

		if err := files.CopyItem(false, true, path.Join(tmpDir, "1panel-core"), "/usr/local/bin"); err != nil {
			global.LOG.Errorf("upgrade 1panel-core failed, err: %v", err)
			_ = settingRepo.Update("SystemStatus", "Free")
			u.handleRollback(originalDir, 1)
			return
		}
		if err := files.CopyItem(false, true, path.Join(tmpDir, "1panel-agent"), "/usr/local/bin"); err != nil {
			global.LOG.Errorf("upgrade 1panel-agent failed, err: %v", err)
			_ = settingRepo.Update("SystemStatus", "Free")
			u.handleRollback(originalDir, 1)
			return
		}

		if err := files.CopyItem(false, true, path.Join(tmpDir, "1pctl"), "/usr/local/bin"); err != nil {
			global.LOG.Errorf("upgrade 1pctl failed, err: %v", err)
			_ = settingRepo.Update("SystemStatus", "Free")
			u.handleRollback(originalDir, 2)
			return
		}
		if _, err := cmd.RunDefaultWithStdoutBashCf("sed -i -e 's#BASE_DIR=.*#BASE_DIR=%s#g' /usr/local/bin/1pctl", global.CONF.Base.InstallDir); err != nil {
			global.LOG.Errorf("upgrade basedir in 1pctl failed, err: %v", err)
			u.handleRollback(originalDir, 2)
			return
		}

		if err := files.CopyItem(false, true, path.Join(tmpDir, "1panel-core.service"), "/etc/systemd/system"); err != nil {
			global.LOG.Errorf("upgrade 1panel.service failed, err: %v", err)
			_ = settingRepo.Update("SystemStatus", "Free")
			u.handleRollback(originalDir, 3)
			return
		}
		if err := files.CopyItem(false, true, path.Join(tmpDir, "1panel-agent.service"), "/etc/systemd/system"); err != nil {
			global.LOG.Errorf("upgrade 1panel.service failed, err: %v", err)
			_ = settingRepo.Update("SystemStatus", "Free")
			u.handleRollback(originalDir, 3)
			return
		}

		global.LOG.Info("upgrade successful!")
		go writeLogs(req.Version)
		_ = settingRepo.Update("SystemVersion", req.Version)
		_ = global.AgentDB.Model(&model.Setting{}).Where("key = ?", "SystemVersion").Updates(map[string]interface{}{"value": req.Version}).Error
		global.CONF.Base.Version = req.Version
		_ = settingRepo.Update("SystemStatus", "Free")

		cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(10 * time.Second))
		_, _ = cmdMgr.RunWithStdoutBashC("systemctl daemon-reload")
		_, _ = cmdMgr.RunWithStdoutBashC("systemctl restart 1panel-agent.service")
		_, _ = cmdMgr.RunWithStdoutBashC("systemctl restart 1panel-core.service")
	}()
	return nil
}

func (u *UpgradeService) Rollback(req dto.OperateByID) error {
	log, _ := upgradeLogRepo.Get(repo.WithByID(req.ID))
	if log.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	u.handleRollback(log.BackupFile, 3)
	return nil
}

func (u *UpgradeService) handleBackup(originalDir string) error {
	if err := files.CopyItem(false, true, "/usr/local/bin/1panel-core", originalDir); err != nil {
		return err
	}
	if err := files.CopyItem(false, true, "/usr/local/bin/1panel-agent", originalDir); err != nil {
		return err
	}
	if err := files.CopyItem(false, true, "/usr/local/bin/1pctl", originalDir); err != nil {
		return err
	}
	if err := files.CopyItem(false, true, "/etc/systemd/system/1panel-core.service", originalDir); err != nil {
		return err
	}
	if err := files.CopyItem(false, true, "/etc/systemd/system/1panel-agent.service", originalDir); err != nil {
		return err
	}
	if err := files.CopyItem(true, true, path.Join(global.CONF.Base.InstallDir, "1panel/db"), originalDir); err != nil {
		return err
	}
	return nil
}

func (u *UpgradeService) handleRollback(originalDir string, errStep int) {
	_ = settingRepo.Update("SystemStatus", "Free")

	dbPath := path.Join(global.CONF.Base.InstallDir, "1panel/db")
	if _, err := os.Stat(path.Join(originalDir, "db")); err == nil {
		if err := files.CopyItem(true, true, path.Join(originalDir, "db"), dbPath); err != nil {
			global.LOG.Errorf("rollback 1panel db failed, err: %v", err)
		}
	}
	if err := files.CopyItem(false, true, path.Join(originalDir, "1panel-core"), "/usr/local/bin"); err != nil {
		global.LOG.Errorf("rollback 1panel-core failed, err: %v", err)
	}
	if err := files.CopyItem(false, true, path.Join(originalDir, "1panel-agent"), "/usr/local/bin"); err != nil {
		global.LOG.Errorf("rollback 1panel-agent failed, err: %v", err)
	}
	if errStep == 1 {
		return
	}
	if err := files.CopyItem(false, true, path.Join(originalDir, "1pctl"), "/usr/local/bin"); err != nil {
		global.LOG.Errorf("rollback 1pctl failed, err: %v", err)
	}
	if errStep == 2 {
		return
	}
	if err := files.CopyItem(false, true, path.Join(originalDir, "1panel-core.service"), "/etc/systemd/system"); err != nil {
		global.LOG.Errorf("rollback 1panel-core.service failed, err: %v", err)
	}
	if err := files.CopyItem(false, true, path.Join(originalDir, "1panel-agent.service"), "/etc/systemd/system"); err != nil {
		global.LOG.Errorf("rollback 1panel-agent.service failed, err: %v", err)
	}
}

func (u *UpgradeService) loadVersionByMode(developer, currentVersion string) (string, string, string) {
	var current, latest string
	if global.CONF.Base.Mode == "dev" {
		devVersionLatest := u.loadVersion(true, currentVersion, "dev")
		return devVersionLatest, "", ""
	}

	betaVersionLatest := ""
	latest = u.loadVersion(true, currentVersion, "stable")
	current = u.loadVersion(false, currentVersion, "stable")
	if developer == constant.StatusEnable {
		betaVersionLatest = u.loadVersion(true, currentVersion, "beta")
	}
	if current != latest {
		return betaVersionLatest, current, latest
	}

	versionPart := strings.Split(current, ".")
	if len(versionPart) < 3 {
		return betaVersionLatest, current, latest
	}
	num, _ := strconv.Atoi(versionPart[1])
	if num == 0 {
		return betaVersionLatest, current, latest
	}
	if num >= 10 {
		if current[:6] == currentVersion[:6] {
			return betaVersionLatest, current, ""
		}
		return betaVersionLatest, "", latest
	}
	if current[:5] == currentVersion[:5] {
		return betaVersionLatest, current, ""
	}
	return betaVersionLatest, "", latest
}

func (u *UpgradeService) loadVersion(isLatest bool, currentVersion, mode string) string {
	path := fmt.Sprintf("%s/%s/latest", global.CONF.RemoteURL.RepoUrl, mode)
	if !isLatest {
		path = fmt.Sprintf("%s/%s/latest.current", global.CONF.RemoteURL.RepoUrl, mode)
	}
	_, latestVersionRes, err := req_helper.HandleRequestWithProxy(path, http.MethodGet, constant.TimeOut20s)
	if err != nil {
		global.LOG.Errorf("load latest version from oss failed, err: %v", err)
		return ""
	}
	version := string(latestVersionRes)
	if strings.Contains(version, "<") {
		global.LOG.Errorf("load latest version from oss failed, err: %v", version)
		return ""
	}
	if isLatest {
		return u.checkVersion(version, currentVersion)
	}

	versionMap := make(map[string]string)
	if err := json.Unmarshal(latestVersionRes, &versionMap); err != nil {
		global.LOG.Errorf("load latest version from oss failed (error unmarshal), err: %v", err)
		return ""
	}

	versionPart := strings.Split(currentVersion, ".")
	if len(versionPart) < 3 {
		global.LOG.Errorf("current version is error format: %s", currentVersion)
		return ""
	}
	num, _ := strconv.Atoi(versionPart[1])
	if num == 0 {
		global.LOG.Errorf("current version is error format: %s", currentVersion)
		return ""
	}
	if num >= 10 {
		if version, ok := versionMap[currentVersion[0:5]]; ok {
			return u.checkVersion(version, currentVersion)
		}
		return ""
	}
	if version, ok := versionMap[currentVersion[0:4]]; ok {
		return u.checkVersion(version, currentVersion)
	}
	return ""
}

func (u *UpgradeService) checkVersion(v2, v1 string) string {
	addSuffix := false
	if !strings.Contains(v1, "-") {
		v1 = v1 + "-lts"
	}
	if !strings.Contains(v2, "-") {
		addSuffix = true
		v2 = v2 + "-lts"
	}
	if common.ComparePanelVersion(v2, v1) {
		if addSuffix {
			return strings.TrimSuffix(v2, "-lts")
		}
		return v2
	}
	return ""
}

func (u *UpgradeService) loadReleaseNotes(path string) (string, error) {
	_, releaseNotes, err := req_helper.HandleRequestWithProxy(path, http.MethodGet, constant.TimeOut20s)
	if err != nil {
		return "", err
	}
	return string(releaseNotes), nil
}

func loadArch() (string, error) {
	std, err := cmd.RunDefaultWithStdoutBashC("uname -a")
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
