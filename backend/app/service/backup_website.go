package service

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/pkg/errors"
)

func (u *BackupService) WebsiteBackup(req dto.CommonBackup) error {
	localDir, err := loadLocalDir()
	if err != nil {
		return err
	}
	website, err := websiteRepo.GetFirst(websiteRepo.WithAlias(req.DetailName))
	if err != nil {
		return err
	}

	timeNow := time.Now().Format(constant.DateTimeSlimLayout)
	itemDir := fmt.Sprintf("website/%s", req.Name)
	backupDir := path.Join(localDir, itemDir)
	fileName := fmt.Sprintf("%s_%s.tar.gz", website.PrimaryDomain, timeNow+common.RandStrAndNum(5))
	if err := handleWebsiteBackup(&website, backupDir, fileName, "", req.Secret); err != nil {
		return err
	}

	record := &model.BackupRecord{
		Type:       "website",
		Name:       website.PrimaryDomain,
		DetailName: req.DetailName,
		Source:     "LOCAL",
		BackupType: "LOCAL",
		FileDir:    itemDir,
		FileName:   fileName,
	}
	if err := backupRepo.CreateRecord(record); err != nil {
		global.LOG.Errorf("save backup record failed, err: %v", err)
		return err
	}
	return nil
}

func (u *BackupService) WebsiteRecover(req dto.CommonRecover) error {
	fileOp := files.NewFileOp()
	if !fileOp.Stat(req.File) {
		return buserr.WithName("ErrFileNotFound", req.File)
	}
	website, err := websiteRepo.GetFirst(websiteRepo.WithAlias(req.DetailName))
	if err != nil {
		return err
	}
	global.LOG.Infof("recover website %s from backup file %s", req.Name, req.File)
	if err := handleWebsiteRecover(&website, req.File, false, req.Secret); err != nil {
		return err
	}
	return nil
}

func handleWebsiteRecover(website *model.Website, recoverFile string, isRollback bool, secret string) error {
	fileOp := files.NewFileOp()
	tmpPath := strings.ReplaceAll(recoverFile, ".tar.gz", "")
	if err := handleUnTar(recoverFile, path.Dir(recoverFile), secret); err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tmpPath)
	}()

	var oldWebsite model.Website
	websiteJson, err := os.ReadFile(tmpPath + "/website.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(websiteJson, &oldWebsite); err != nil {
		return fmt.Errorf("unmarshal app.json failed, err: %v", err)
	}

	if err := checkValidOfWebsite(&oldWebsite, website); err != nil {
		return err
	}

	temPathWithName := tmpPath + "/" + website.Alias
	if !fileOp.Stat(tmpPath+"/website.json") || !fileOp.Stat(temPathWithName+".conf") || !fileOp.Stat(temPathWithName+".web.tar.gz") {
		return buserr.WithDetail(constant.ErrBackupExist, ".conf or .web.tar.gz", nil)
	}
	if website.Type == constant.Deployment {
		if !fileOp.Stat(temPathWithName + ".app.tar.gz") {
			return buserr.WithDetail(constant.ErrBackupExist, ".app.tar.gz", nil)
		}
	}

	isOk := false
	if !isRollback {
		rollbackFile := path.Join(global.CONF.System.TmpDir, fmt.Sprintf("website/%s_%s.tar.gz", website.Alias, time.Now().Format(constant.DateTimeSlimLayout)))
		if err := handleWebsiteBackup(website, path.Dir(rollbackFile), path.Base(rollbackFile), "", ""); err != nil {
			return fmt.Errorf("backup website %s for rollback before recover failed, err: %v", website.Alias, err)
		}
		defer func() {
			if !isOk {
				global.LOG.Info("recover failed, start to rollback now")
				if err := handleWebsiteRecover(website, rollbackFile, true, ""); err != nil {
					global.LOG.Errorf("rollback website %s from %s failed, err: %v", website.Alias, rollbackFile, err)
					return
				}
				global.LOG.Infof("rollback website %s from %s successful", website.Alias, rollbackFile)
				_ = os.RemoveAll(rollbackFile)
			} else {
				_ = os.RemoveAll(rollbackFile)
			}
		}()
	}

	nginxInfo, err := appInstallRepo.LoadBaseInfo(constant.AppOpenresty, "")
	if err != nil {
		return err
	}
	nginxConfPath := fmt.Sprintf("%s/openresty/%s/conf/conf.d", constant.AppInstallDir, nginxInfo.Name)
	if err := fileOp.CopyFile(fmt.Sprintf("%s/%s.conf", tmpPath, website.Alias), nginxConfPath); err != nil {
		global.LOG.Errorf("handle recover from conf.d failed, err: %v", err)
		return err
	}

	switch website.Type {
	case constant.Deployment:
		app, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
		if err != nil {
			return err
		}
		if err := handleAppRecover(&app, fmt.Sprintf("%s/%s.app.tar.gz", tmpPath, website.Alias), true, ""); err != nil {
			global.LOG.Errorf("handle recover from app.tar.gz failed, err: %v", err)
			return err
		}
		if _, err := compose.Restart(fmt.Sprintf("%s/%s/%s/docker-compose.yml", constant.AppInstallDir, app.App.Key, app.Name)); err != nil {
			global.LOG.Errorf("docker-compose restart failed, err: %v", err)
			return err
		}
	case constant.Runtime:
		runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(website.RuntimeID))
		if err != nil {
			return err
		}
		if runtime.Type == constant.RuntimeNode || runtime.Type == constant.RuntimeJava || runtime.Type == constant.RuntimeGo {
			if err := handleRuntimeRecover(runtime, fmt.Sprintf("%s/%s.runtime.tar.gz", tmpPath, website.Alias), true, ""); err != nil {
				return err
			}
			global.LOG.Info("put runtime.tar.gz into tmp dir successful")
		}
	}

	siteDir := fmt.Sprintf("%s/openresty/%s/www/sites", constant.AppInstallDir, nginxInfo.Name)
	if err := handleUnTar(fmt.Sprintf("%s/%s.web.tar.gz", tmpPath, website.Alias), siteDir, ""); err != nil {
		global.LOG.Errorf("handle recover from web.tar.gz failed, err: %v", err)
		return err
	}
	stdout, err := cmd.Execf("docker exec -i %s nginx -s reload", nginxInfo.ContainerName)
	if err != nil {
		global.LOG.Errorf("nginx -s reload failed, err: %s", stdout)
		return errors.New(string(stdout))
	}

	oldWebsite.ID = website.ID
	if err := websiteRepo.SaveWithoutCtx(&oldWebsite); err != nil {
		global.LOG.Errorf("handle save website data failed, err: %v", err)
		return err
	}
	isOk = true
	return nil
}

func handleWebsiteBackup(website *model.Website, backupDir, fileName string, excludes string, secret string) error {
	fileOp := files.NewFileOp()
	tmpDir := fmt.Sprintf("%s/%s", backupDir, strings.ReplaceAll(fileName, ".tar.gz", ""))
	if !fileOp.Stat(tmpDir) {
		if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
			return fmt.Errorf("mkdir %s failed, err: %v", backupDir, err)
		}
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	remarkInfo, _ := json.Marshal(website)
	if err := fileOp.SaveFile(tmpDir+"/website.json", string(remarkInfo), fs.ModePerm); err != nil {
		return err
	}
	global.LOG.Info("put website.json into tmp dir successful")

	nginxInfo, err := appInstallRepo.LoadBaseInfo(constant.AppOpenresty, "")
	if err != nil {
		return err
	}
	nginxConfFile := fmt.Sprintf("%s/openresty/%s/conf/conf.d/%s.conf", constant.AppInstallDir, nginxInfo.Name, website.Alias)
	if err := fileOp.CopyFile(nginxConfFile, tmpDir); err != nil {
		return err
	}
	global.LOG.Info("put openresty conf into tmp dir successful")

	switch website.Type {
	case constant.Deployment:
		app, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
		if err != nil {
			return err
		}
		if err := handleAppBackup(&app, tmpDir, fmt.Sprintf("%s.app.tar.gz", website.Alias), excludes, ""); err != nil {
			return err
		}
		global.LOG.Info("put app.tar.gz into tmp dir successful")
	case constant.Runtime:
		runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(website.RuntimeID))
		if err != nil {
			return err
		}
		if runtime.Type == constant.RuntimeNode || runtime.Type == constant.RuntimeJava || runtime.Type == constant.RuntimeGo {
			if err := handleRuntimeBackup(runtime, tmpDir, fmt.Sprintf("%s.runtime.tar.gz", website.Alias), excludes, ""); err != nil {
				return err
			}
			global.LOG.Info("put runtime.tar.gz into tmp dir successful")
		}
	}

	websiteDir := fmt.Sprintf("%s/openresty/%s/www/sites/%s", constant.AppInstallDir, nginxInfo.Name, website.Alias)
	if err := handleTar(websiteDir, tmpDir, fmt.Sprintf("%s.web.tar.gz", website.Alias), excludes, ""); err != nil {
		return err
	}
	global.LOG.Info("put web.tar.gz into tmp dir successful, now start to tar tmp dir")
	if err := handleTar(tmpDir, backupDir, fileName, "", secret); err != nil {
		return err
	}

	return nil
}

func checkValidOfWebsite(oldWebsite, website *model.Website) error {
	if oldWebsite.Alias != website.Alias || oldWebsite.Type != website.Type {
		return buserr.WithDetail(constant.ErrBackupMatch, fmt.Sprintf("oldName: %s, oldType: %v", oldWebsite.Alias, oldWebsite.Type), nil)
	}
	if oldWebsite.AppInstallID != 0 {
		_, err := appInstallRepo.GetFirst(commonRepo.WithByID(oldWebsite.AppInstallID))
		if err != nil {
			return buserr.WithDetail(constant.ErrBackupMatch, "app", nil)
		}
	}
	if oldWebsite.RuntimeID != 0 {
		if _, err := runtimeRepo.GetFirst(commonRepo.WithByID(oldWebsite.RuntimeID)); err != nil {
			return buserr.WithDetail(constant.ErrBackupMatch, "runtime", nil)
		}
	}
	if oldWebsite.WebsiteSSLID != 0 {
		if _, err := websiteSSLRepo.GetFirst(commonRepo.WithByID(oldWebsite.WebsiteSSLID)); err != nil {
			return buserr.WithDetail(constant.ErrBackupMatch, "ssl", nil)
		}
	}
	return nil
}
