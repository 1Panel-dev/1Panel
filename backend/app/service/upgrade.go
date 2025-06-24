package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
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
	"github.com/1Panel-dev/1Panel/backend/utils/systemctl"
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
	serviceHandle, _ := systemctl.DefaultHandler("1panel")
	currentServiceName := serviceHandle.GetServiceName()
	if err := settingRepo.Update("SystemStatus", "Upgrading"); err != nil {
		return fmt.Errorf("update system status failed: %w", err)
	}

	go func() {
		defer func() {
			if err := settingRepo.Update("SystemStatus", "Free"); err != nil {
				global.LOG.Errorf("Reset system status failed: %v", err)
			}
		}()

		_ = global.Cron.Stop()
		defer global.Cron.Start()

		if err := fileOp.DownloadFileWithProxy(
			fmt.Sprintf("%s/%s", downloadPath, fileName),
			path.Join(rootDir, fileName),
		); err != nil {
			global.LOG.Errorf("Failed to download upgrade package: %v", err)
			return
		}
		defer os.RemoveAll(rootDir)

		if err := handleUnTar(path.Join(rootDir, fileName), rootDir, ""); err != nil {
			global.LOG.Errorf("Failed to extract package: %v", err)
			return
		}

		tmpDir := path.Join(rootDir, strings.TrimSuffix(fileName, ".tar.gz"))

		if err := u.handleBackup(fileOp, originalDir); err != nil {
			global.LOG.Errorf("Backup failed: %v", err)
			return
		}

		binDir := systemctl.BinaryPath
		servicePath, _ := serviceHandle.GetServicePath()
		geoPath := path.Join(global.CONF.System.BaseDir, "1panel/geo/GeoIP.mmdb")

		criticalUpdates := []struct {
			src  string
			dest string
			step int
		}{
			{path.Join(tmpDir, "1panel"), path.Join(binDir, "1panel"), 1},
			{path.Join(tmpDir, "1pctl"), path.Join(binDir, "1pctl"), 2},
			{selectInitScript(tmpDir, currentServiceName), servicePath, 3},
		}

		for _, update := range criticalUpdates {
			if err := common.Copy(update.src, update.dest); err != nil {
				global.LOG.Errorf("Update %s failed: %v", path.Base(update.dest), err)
				u.handleRollback(originalDir, update.step)
				return
			}
		}

		if _, err := cmd.Execf("sed -i -e 's#BASE_DIR=.*#BASE_DIR=%s#g' /usr/local/bin/1pctl",
			global.CONF.System.BaseDir); err != nil {
			global.LOG.Errorf("Update base directory failed: %v", err)
			u.handleRollback(originalDir, 2)
			return
		}

		langDir := path.Join(binDir, "lang")
		if err := common.Copy(path.Join(tmpDir, "lang"), langDir); err != nil {
			global.LOG.Errorf("Update language files failed: %v", err)
		}
		if err := common.Copy(path.Join(tmpDir, "GeoIP.mmdb"), geoPath); err != nil {
			global.LOG.Warnf("Update GeoIP database failed: %v", err)
		}

		global.LOG.Info("upgrade successful!")
		go writeLogs(req.Version)
		checkPointOfWal()
		if err := settingRepo.Update("SystemVersion", req.Version); err != nil {
			global.LOG.Errorf("Update system version failed: %v", err)
		}
		if serviceHandle.ManagerName() == "systemd" {
			_, _ = cmd.Exec("systemctl daemon-reload")
		}
		if err := systemctl.Restart("1panel"); err != nil {
			global.LOG.Errorf("Service restart failed: %v", err)
			return
		}
	}()
	return nil
}

func (u *UpgradeService) handleBackup(fileOp files.FileOp, originalDir string) error {
	global.LOG.Info("Initiating backup procedure...")
	h, _ := systemctl.DefaultHandler("1panel")
	binDir := systemctl.BinaryPath
	servicePath, _ := h.GetServicePath()
	geoPath := path.Join(global.CONF.System.BaseDir, "1panel/geo/GeoIP.mmdb")

	backupItems := []struct {
		src  string
		dest string
	}{
		{path.Join(binDir, "1panel"), originalDir},
		{path.Join(binDir, "1pctl"), originalDir},
		{servicePath, originalDir},
		{path.Join(binDir, "lang"), originalDir},
		{geoPath, originalDir},
	}

	for _, item := range backupItems {
		if err := fileOp.Copy(item.src, item.dest); err != nil {
			return fmt.Errorf("backup %s failed: %w", path.Base(item.src), err)
		}
	}

	if err := handleTar(
		path.Join(global.CONF.System.BaseDir, "1panel/db"),
		originalDir,
		"db.tar.gz",
		"db/1Panel.db-*",
		"",
	); err != nil {
		return fmt.Errorf("database backup failed: %w", err)
	}
	return nil
}

func (u *UpgradeService) handleRollback(originalDir string, errStep int) {
	global.LOG.Info("Initiating rollback procedure...")
	h, _ := systemctl.DefaultHandler("1panel")
	binDir := systemctl.BinaryPath
	servicePath, _ := h.GetServicePath()
	geoPath := path.Join(global.CONF.System.BaseDir, "1panel/geo/GeoIP.mmdb")

	rollbackSteps := []struct {
		src  string
		dest string
	}{
		{path.Join(originalDir, "1panel"), path.Join(binDir, "1panel")},
		{path.Join(originalDir, "1pctl"), path.Join(binDir, "1pctl")},
		{path.Join(originalDir, filepath.Base(servicePath)), servicePath},
		{path.Join(originalDir, "lang"), path.Join(binDir, "lang")},
		{path.Join(originalDir, "GeoIP.mmdb"), geoPath},
	}

	for _, step := range rollbackSteps[:errStep] {
		if err := common.CopyFile(step.src, step.dest); err != nil {
			global.LOG.Errorf("Rollback %s failed: %v", path.Base(step.src), err)
		}
	}

	if err := systemctl.Restart("1panel"); err != nil {
		global.LOG.Errorf("Service restart during rollback failed: %v", err)
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
	if strings.Contains(std, "riscv64") {
		return "riscv64", nil
	}
	return "", fmt.Errorf("unsupported such arch: %s", std)
}

func selectInitScript(path string, serviceName string) string {
	path = strings.TrimSuffix(path, "/")
	mgr := systemctl.GetGlobalManager().Name()
	var serviceFileName string
	switch mgr {
	case "systemd":
		serviceFileName = "1panel.service"
	case "openrc":
		serviceFileName = "1paneld.openrc"
	case "sysvinit":
		isWrt := systemctl.FileExist("/etc/rc.common")
		if isWrt {
			serviceFileName = "1paneld.procd"
		} else {
			serviceFileName = "1paneld.init"
		}
	default:
		serviceFileName = serviceName
		global.LOG.Warnf("[%s]unselect InitScript, used default: %s", mgr, serviceName)
	}
	sourcePath := filepath.Join(path, serviceFileName)
	targetPath := filepath.Join(path, serviceName)

	if serviceFileName != serviceName {
		if _, err := cmd.Execf("cp %s %s", sourcePath, targetPath); err != nil {
			global.LOG.Errorf("Failed to copy init script from %s to %s: %v",
				serviceFileName, serviceName, err)
		}
	}

	return targetPath
}
