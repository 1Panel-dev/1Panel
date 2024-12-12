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

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	httpUtil "github.com/1Panel-dev/1Panel/backend/utils/http"
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
	DeveloperMode, err := settingRepo.Get(settingRepo.WithByKey("DeveloperMode"))
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
	if (global.CONF.System.Mode == "dev" || DeveloperMode.Value == "enable") && len(upgrade.TestVersion) != 0 {
		itemVersion = upgrade.TestVersion
	}
	if len(itemVersion) == 0 {
		return &upgrade, nil
	}
	mode := global.CONF.System.Mode
	if strings.Contains(itemVersion, "beta") {
		mode = "beta"
	}
	notes, err := u.loadReleaseNotes(fmt.Sprintf("%s/%s/%s/release/1panel-%s-release-notes", global.CONF.System.RepoUrl, mode, itemVersion, itemVersion))
	if err != nil {
		return nil, fmt.Errorf("load releases-notes of version %s failed, err: %v", itemVersion, err)
	}
	upgrade.ReleaseNote = notes
	return &upgrade, nil
}

func (u *UpgradeService) LoadNotes(req dto.Upgrade) (string, error) {
	mode := global.CONF.System.Mode
	if strings.Contains(req.Version, "beta") {
		mode = "beta"
	}
	notes, err := u.loadReleaseNotes(fmt.Sprintf("%s/%s/%s/release/1panel-%s-release-notes", global.CONF.System.RepoUrl, mode, req.Version, req.Version))
	if err != nil {
		return "", fmt.Errorf("load releases-notes of version %s failed, err: %v", req.Version, err)
	}
	return notes, nil
}

func (u *UpgradeService) Upgrade(req dto.Upgrade) error {
	global.LOG.Info("start to upgrade now...")
	fileOp := files.NewFileOp()
	timeStr := time.Now().Format(constant.DateTimeSlimLayout)
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

	mode := global.CONF.System.Mode
	if strings.Contains(req.Version, "beta") {
		mode = "beta"
	}
	downloadPath := fmt.Sprintf("%s/%s/%s/release", global.CONF.System.RepoUrl, mode, req.Version)
	fileName := fmt.Sprintf("1panel-%s-%s-%s.tar.gz", req.Version, "linux", itemArch)
	_ = settingRepo.Update("SystemStatus", "Upgrading")
	go func() {
		_ = global.Cron.Stop()
		defer func() {
			global.Cron.Start()
		}()
		if err := fileOp.DownloadFileWithProxy(downloadPath+"/"+fileName, rootDir+"/"+fileName); err != nil {
			global.LOG.Errorf("download service file failed, err: %v", err)
			_ = settingRepo.Update("SystemStatus", "Free")
			return
		}
		global.LOG.Info("download all file successful!")
		defer func() {
			_ = os.Remove(rootDir)
		}()
		if err := handleUnTar(rootDir+"/"+fileName, rootDir, ""); err != nil {
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

		if err := common.CopyFile(path.Join(tmpDir, "1panel"), "/usr/local/bin"); err != nil {
			global.LOG.Errorf("upgrade 1panel failed, err: %v", err)
			u.handleRollback(originalDir, 1)
			return
		}

		if err := common.CopyFile(path.Join(tmpDir, "1pctl"), "/usr/local/bin"); err != nil {
			global.LOG.Errorf("upgrade 1pctl failed, err: %v", err)
			u.handleRollback(originalDir, 2)
			return
		}
		_, _ = cmd.Execf("cp -r %s /usr/local/bin", path.Join(tmpDir, "lang"))
		geoPath := path.Join(global.CONF.System.BaseDir, "1panel/geo")
		_, _ = cmd.Execf("mkdir %s && cp %s %s/", geoPath, path.Join(tmpDir, "GeoIP.mmdb"), geoPath)

		if _, err := cmd.Execf("sed -i -e 's#BASE_DIR=.*#BASE_DIR=%s#g' /usr/local/bin/1pctl", global.CONF.System.BaseDir); err != nil {
			global.LOG.Errorf("upgrade basedir in 1pctl failed, err: %v", err)
			u.handleRollback(originalDir, 2)
			return
		}

		if err := common.CopyFile(path.Join(tmpDir, "1panel.service"), "/etc/systemd/system"); err != nil {
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
	_, _ = cmd.Execf("cp -r /usr/local/bin/lang %s", originalDir)
	_, _ = cmd.Execf("cp %s %s", path.Join(global.CONF.System.BaseDir, "1panel/geo/GeoIP.mmdb"), originalDir)
	checkPointOfWal()
	if err := handleTar(path.Join(global.CONF.System.BaseDir, "1panel/db"), originalDir, "db.tar.gz", "db/1Panel.db-*", ""); err != nil {
		return err
	}
	return nil
}

func (u *UpgradeService) handleRollback(originalDir string, errStep int) {
	_ = settingRepo.Update("SystemStatus", "Free")

	checkPointOfWal()
	if _, err := os.Stat(path.Join(originalDir, "1Panel.db")); err == nil {
		if err := common.CopyFile(path.Join(originalDir, "1Panel.db"), global.CONF.System.DbPath); err != nil {
			global.LOG.Errorf("rollback 1panel db failed, err: %v", err)
		}
	}
	if _, err := os.Stat(path.Join(originalDir, "db.tar.gz")); err == nil {
		if err := handleUnTar(path.Join(originalDir, "db.tar.gz"), global.CONF.System.DbPath, ""); err != nil {
			global.LOG.Errorf("rollback 1panel db failed, err: %v", err)
		}
	}
	if err := common.CopyFile(path.Join(originalDir, "1panel"), "/usr/local/bin"); err != nil {
		global.LOG.Errorf("rollback 1pctl failed, err: %v", err)
	}
	if errStep == 1 {
		return
	}
	if err := common.CopyFile(path.Join(originalDir, "1pctl"), "/usr/local/bin"); err != nil {
		global.LOG.Errorf("rollback 1panel failed, err: %v", err)
	}
	_, _ = cmd.Execf("cp -r %s /usr/local/bin", path.Join(originalDir, "lang"))
	geoPath := path.Join(global.CONF.System.BaseDir, "1panel/geo")
	_, _ = cmd.Execf("mkdir %s && cp %s %s/", geoPath, path.Join(originalDir, "GeoIP.mmdb"), geoPath)

	if errStep == 2 {
		return
	}
	if err := common.CopyFile(path.Join(originalDir, "1panel.service"), "/etc/systemd/system"); err != nil {
		global.LOG.Errorf("rollback 1panel failed, err: %v", err)
	}
}

func (u *UpgradeService) loadVersionByMode(developer, currentVersion string) (string, string, string) {
	var current, latest string
	if global.CONF.System.Mode == "dev" {
		betaVersionLatest := u.loadVersion(true, currentVersion, "beta")
		devVersionLatest := u.loadVersion(true, currentVersion, "dev")
		if common.ComparePanelVersion(betaVersionLatest, devVersionLatest) {
			return betaVersionLatest, "", ""
		}
		return devVersionLatest, "", ""
	}

	betaVersionLatest := ""
	latest = u.loadVersion(true, currentVersion, "stable")
	current = u.loadVersion(false, currentVersion, "stable")
	if developer == "enable" {
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
	path := fmt.Sprintf("%s/%s/latest", global.CONF.System.RepoUrl, mode)
	if !isLatest {
		path = fmt.Sprintf("%s/%s/latest.current", global.CONF.System.RepoUrl, mode)
	}
	_, latestVersionRes, err := httpUtil.HandleGet(path, http.MethodGet, constant.TimeOut20s)
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
	_, releaseNotes, err := httpUtil.HandleGet(path, http.MethodGet, constant.TimeOut20s)
	if err != nil {
		return "", err
	}
	return string(releaseNotes), nil
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
