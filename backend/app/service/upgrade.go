package service

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gitee.com/openeuler/go-gitee/gitee"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/google/go-github/github"
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
	currentVerion, err := settingRepo.Get(settingRepo.WithByKey("SystemVersion"))
	if err != nil {
		return nil, err
	}

	var releaseInfo dto.UpgradeInfo
	isGiteeOK := checkValid("https://gitee.com/wanghe-fit2cloud/1Panel")
	if isGiteeOK {
		releaseInfo, err = u.loadLatestFromGitee()
		if err != nil {
			global.LOG.Error(err)
		}
	}
	if len(releaseInfo.NewVersion) == 0 {
		isGithubOK := checkValid("https://gitee.com/1Panel-dev/1Panel")
		if isGithubOK {
			releaseInfo, err = u.loadLatestFromGithub()
			if err != nil {
				global.LOG.Error(err)
				return nil, err
			}
		}
	}
	if len(releaseInfo.NewVersion) != 0 {
		isNew, err := compareVersion(currentVerion.Value, releaseInfo.NewVersion)
		if !isNew && err != nil {
			return nil, err
		}
		return &releaseInfo, nil
	}

	return nil, errors.New("both gitee and github were unavailable")
}

func (u *UpgradeService) Upgrade(req dto.Upgrade) error {
	global.LOG.Info("start to upgrade now...")
	fileOp := files.NewFileOp()
	timeStr := time.Now().Format("20060102150405")
	rootDir := fmt.Sprintf("%s/upgrade_%s/downloads", constant.TmpDir, timeStr)
	originalDir := fmt.Sprintf("%s/upgrade_%s/original", constant.TmpDir, timeStr)
	if err := os.MkdirAll(rootDir, os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(originalDir, os.ModePerm); err != nil {
		return err
	}

	downloadPath := fmt.Sprintf("https://gitee.com/%s/%s/releases/download/%s/", "wanghe-fit2cloud", "1Panel", req.Version)
	isGiteeOK := checkValid(downloadPath)
	if !isGiteeOK {
		downloadPath = fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/", "wanghe-fit2cloud", "1Panel", req.Version)
		isGithubOK := checkValid(downloadPath)
		if !isGithubOK {
			return errors.New("both gitee and github were unavailabl")
		}
	}

	panelName := fmt.Sprintf("1panel-%s-%s", "linux", runtime.GOARCH)
	fileName := fmt.Sprintf("1panel-online-installer-%s.tar.gz", req.Version)
	_ = settingRepo.Update("SystemStatus", "Upgrading")
	go func() {
		if err := fileOp.DownloadFile(downloadPath+panelName, rootDir+"/1panel"); err != nil {
			global.LOG.Errorf("download panel file failed, err: %v", err)
			return
		}
		if err := fileOp.DownloadFile(downloadPath+fileName, rootDir+"/service.tar.gz"); err != nil {
			global.LOG.Errorf("download service file failed, err: %v", err)
			return
		}
		global.LOG.Info("download all file successful!")
		defer func() {
			_ = os.Remove(rootDir)
		}()
		if err := fileOp.Decompress(rootDir+"/service.tar.gz", rootDir, files.TarGz); err != nil {
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
		if err := fileOp.Chmod("/usr/local/bin/1panel", 0755); err != nil {
			u.handleRollback(fileOp, originalDir, 1)
			global.LOG.Errorf("chmod 1panel failed, err: %v", err)
			return
		}
		if err := cpBinary(fmt.Sprintf("%s/1panel-online-installer-%s/1pctl", rootDir, req.Version), "/usr/local/bin/1pctl"); err != nil {
			u.handleRollback(fileOp, originalDir, 2)
			global.LOG.Errorf("upgrade 1pctl failed, err: %v", err)
			return
		}
		if err := fileOp.Chmod("/usr/local/bin/1pctl", 0755); err != nil {
			u.handleRollback(fileOp, originalDir, 1)
			global.LOG.Errorf("chmod 1pctl failed, err: %v", err)
			return
		}
		if err := cpBinary(fmt.Sprintf("%s/1panel-online-installer-%s/1panel/conf/1panel.service", rootDir, req.Version), "/etc/systemd/system/1panel.service"); err != nil {
			u.handleRollback(fileOp, originalDir, 3)
			global.LOG.Errorf("upgrade 1panel.service failed, err: %v", err)
			return
		}

		global.LOG.Info("upgrade successful!")
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

func (u *UpgradeService) loadLatestFromGithub() (dto.UpgradeInfo, error) {
	var info dto.UpgradeInfo
	client := github.NewClient(nil)
	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()
	stats, res, err := client.Repositories.GetLatestRelease(ctx, "wanghe-fit2cloud", "1Panel")
	if res.StatusCode != 200 || err != nil {
		return info, fmt.Errorf("load upgrade info from github failed, err: %v", err)
	}
	info.NewVersion = string(*stats.Name)
	info.ReleaseNote = string(*stats.Body)
	info.CreatedAt = stats.PublishedAt.Add(8 * time.Hour).Format("2006-01-02 15:04:05")
	return info, nil
}

func (u *UpgradeService) loadLatestFromGitee() (dto.UpgradeInfo, error) {
	var info dto.UpgradeInfo
	client := gitee.NewAPIClient(gitee.NewConfiguration())
	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()
	stats, res, err := client.RepositoriesApi.GetV5ReposOwnerRepoReleasesLatest(ctx, "wanghe-fit2cloud", "1Panel", &gitee.GetV5ReposOwnerRepoReleasesLatestOpts{})
	if res.StatusCode != 200 || err != nil {
		return info, fmt.Errorf("load upgrade info from gitee failed, err: %v", err)
	}
	info.NewVersion = string(stats.Name)
	info.ReleaseNote = string(stats.Body)
	info.CreatedAt = stats.CreatedAt.Format("2006-01-02 15:04:05")
	return info, nil
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

func checkValid(addr string) bool {
	timeout := time.Duration(2 * time.Second)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Transport: tr,
		Timeout:   timeout,
	}
	if _, err := client.Get(addr); err != nil {
		return false
	}
	return true
}
