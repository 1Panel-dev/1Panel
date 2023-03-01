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
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/pkg/errors"
)

func (u *BackupService) WebsiteBackup(req dto.CommonBackup) error {
	localDir, err := loadLocalDir()
	if err != nil {
		return err
	}
	website, err := websiteRepo.GetFirst(websiteRepo.WithDomain(req.Name))
	if err != nil {
		return err
	}

	timeNow := time.Now().Format("20060102150405")
	backupDir := fmt.Sprintf("%s/website/%s", localDir, req.Name)
	fileName := fmt.Sprintf("%s_%s.tar.gz", website.PrimaryDomain, timeNow)
	if err := handleWebsiteBackup(&website, backupDir, fileName); err != nil {
		return err
	}

	record := &model.BackupRecord{
		Type:       "website",
		Name:       website.PrimaryDomain,
		DetailName: "",
		Source:     "LOCAL",
		BackupType: "LOCAL",
		FileDir:    backupDir,
		FileName:   fileName,
	}
	if err := backupRepo.CreateRecord(record); err != nil {
		global.LOG.Errorf("save backup record failed, err: %v", err)
		return err
	}
	return nil
}

func (u *BackupService) WebsiteRecoverByUpload(req dto.CommonRecover) error {
	if err := handleUnTar(req.File, path.Dir(req.File)); err != nil {
		return err
	}
	tmpDir := strings.ReplaceAll(req.File, ".tar.gz", "")
	webJson, err := os.ReadFile(fmt.Sprintf("%s/website.json", tmpDir))
	if err != nil {
		return err
	}
	var websiteInfo WebsiteInfo
	if err := json.Unmarshal(webJson, &websiteInfo); err != nil {
		return err
	}
	if websiteInfo.WebsiteName != req.Name {
		return errors.New("the uploaded file does not match the selected website and cannot be recovered")
	}

	website, err := websiteRepo.GetFirst(websiteRepo.WithDomain(req.Name))
	if err != nil {
		return err
	}
	if err := handleWebsiteRecover(&website, tmpDir, false); err != nil {
		return err
	}

	return nil
}

func (u *BackupService) WebsiteRecover(req dto.CommonRecover) error {
	website, err := websiteRepo.GetFirst(websiteRepo.WithDomain(req.Name))
	if err != nil {
		return err
	}
	fileOp := files.NewFileOp()
	if !fileOp.Stat(req.File) {
		return errors.New(fmt.Sprintf("%s file is not exist", req.File))
	}
	global.LOG.Infof("recover website %s from backup file %s", req.Name, req.File)
	if err := handleWebsiteRecover(&website, req.File, false); err != nil {
		return err
	}
	return nil
}

func handleWebsiteRecover(website *model.Website, recoverFile string, isRollback bool) error {
	fileOp := files.NewFileOp()
	fileDir := strings.ReplaceAll(recoverFile, ".tar.gz", "")
	if err := fileOp.Decompress(recoverFile, path.Dir(recoverFile), files.TarGz); err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(fileDir)
	}()

	itemDir := fmt.Sprintf("%s/%s", fileDir, website.Alias)
	if !fileOp.Stat(itemDir+".conf") || !fileOp.Stat(itemDir+".web.tar.gz") {
		return errors.New("the wrong recovery package does not have .conf or .web.tar.gz files")
	}
	if website.Type == constant.Deployment {
		if !fileOp.Stat(itemDir+".sql.gz") || !fileOp.Stat(itemDir+".app.tar.gz") {
			return errors.New("the wrong recovery package does not have .sql.gz or .app.tar.gz files")
		}
	}
	isOk := false
	if !isRollback {
		rollbackFile := fmt.Sprintf("%s/original/website/%s_%s.tar.gz", global.CONF.System.BaseDir, website.Alias, time.Now().Format("20060102150405"))
		if err := handleWebsiteBackup(website, path.Dir(rollbackFile), path.Base(rollbackFile)); err != nil {
			global.LOG.Errorf("backup website %s for rollback before recover failed, err: %v", website.Alias, err)
		}
		defer func() {
			if !isOk {
				if err := handleWebsiteRecover(website, rollbackFile, true); err != nil {
					global.LOG.Errorf("rollback website %s from %s failed, err: %v", website.Alias, rollbackFile, err)
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
	if err := fileOp.CopyFile(fmt.Sprintf("%s/%s.conf", fileDir, website.Alias), nginxConfPath); err != nil {
		return err
	}

	if website.Type == constant.Deployment {
		mysqlInfo, err := appInstallRepo.LoadBaseInfo(constant.AppMysql, "")
		if err != nil {
			return err
		}
		resource, err := appInstallResourceRepo.GetFirst(appInstallResourceRepo.WithAppInstallId(website.AppInstallID))
		if err != nil {
			return err
		}
		db, err := mysqlRepo.Get(commonRepo.WithByID(resource.ResourceId))
		if err != nil {
			return err
		}
		if err := handleMysqlRecover(mysqlInfo, fileDir, db.Name, fmt.Sprintf("%s.sql.gz", website.Alias), isRollback); err != nil {
			return err
		}
		app, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
		if err != nil {
			return err
		}
		if err := handleAppRecover(&app, fmt.Sprintf("%s/%s.app.tar.gz", fileDir, website.Alias), isRollback); err != nil {
			return err
		}
		if _, err := compose.Restart(fmt.Sprintf("%s/%s/%s/docker-compose.yml", constant.AppInstallDir, app.App.Key, app.Name)); err != nil {
			return err
		}
	}
	siteDir := fmt.Sprintf("%s/openresty/%s/www/sites", constant.AppInstallDir, nginxInfo.Name)
	if err := fileOp.Decompress(fmt.Sprintf("%s/%s.web.tar.gz", fileDir, website.Alias), siteDir, files.TarGz); err != nil {
		return err
	}
	stdout, err := cmd.Execf("docker exec -i %s nginx -s reload", nginxInfo.ContainerName)
	if err != nil {
		return errors.New(string(stdout))
	}

	return nil
}

type WebsiteInfo struct {
	WebsiteName string `json:"websiteName"`
	WebsiteType string `json:"websiteType"`
}

func handleWebsiteBackup(website *model.Website, backupDir, fileName string) error {
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

	var websiteInfo WebsiteInfo
	websiteInfo.WebsiteType = website.Type
	websiteInfo.WebsiteName = website.PrimaryDomain
	remarkInfo, _ := json.Marshal(websiteInfo)
	if err := fileOp.SaveFile(tmpDir+"/website.json", string(remarkInfo), fs.ModePerm); err != nil {
		return err
	}
	global.LOG.Info("put websitejson into tmp dir successful")

	nginxInfo, err := appInstallRepo.LoadBaseInfo(constant.AppOpenresty, "")
	if err != nil {
		return err
	}
	nginxConfFile := fmt.Sprintf("%s/openresty/%s/conf/conf.d/%s.conf", constant.AppInstallDir, nginxInfo.Name, website.Alias)
	if err := fileOp.CopyFile(nginxConfFile, tmpDir); err != nil {
		return err
	}
	global.LOG.Info("put openresty conf into tmp dir successful")

	if website.Type == constant.Deployment {
		mysqlInfo, err := appInstallRepo.LoadBaseInfo(constant.AppMysql, "")
		if err != nil {
			return err
		}
		resource, err := appInstallResourceRepo.GetFirst(appInstallResourceRepo.WithAppInstallId(website.AppInstallID))
		if err != nil {
			return err
		}
		db, err := mysqlRepo.Get(commonRepo.WithByID(resource.ResourceId))
		if err != nil {
			return err
		}
		if err := handleMysqlBackup(mysqlInfo, tmpDir, db.Name, fmt.Sprintf("%s.sql.gz", website.Alias)); err != nil {
			return err
		}
		app, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
		if err != nil {
			return err
		}
		if err := handleAppBackup(&app, tmpDir, fmt.Sprintf("%s.app.tar.gz", website.Alias)); err != nil {
			return err
		}
		global.LOG.Info("put app tar into tmp dir successful")
	}
	websiteDir := fmt.Sprintf("%s/openresty/%s/www/sites/%s", constant.AppInstallDir, nginxInfo.Name, website.Alias)
	if err := fileOp.Compress([]string{websiteDir}, tmpDir, fmt.Sprintf("%s.web.tar.gz", website.Alias), files.TarGz); err != nil {
		return err
	}
	global.LOG.Info("put website tar into tmp dir successful, now start to tar tmp dir")
	if err := fileOp.Compress([]string{tmpDir}, backupDir, fileName, files.TarGz); err != nil {
		return err
	}

	return nil
}
