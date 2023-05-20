package service

import (
	"context"
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
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/pkg/errors"
)

func (u *BackupService) AppBackup(req dto.CommonBackup) error {
	localDir, err := loadLocalDir()
	if err != nil {
		return err
	}
	app, err := appRepo.GetFirst(appRepo.WithKey(req.Name))
	if err != nil {
		return err
	}
	install, err := appInstallRepo.GetFirst(commonRepo.WithByName(req.DetailName), appInstallRepo.WithAppId(app.ID))
	if err != nil {
		return err
	}
	timeNow := time.Now().Format("20060102150405")

	backupDir := fmt.Sprintf("%s/app/%s/%s", localDir, req.Name, req.DetailName)

	fileName := fmt.Sprintf("%s_%s.tar.gz", req.DetailName, timeNow)
	if err := handleAppBackup(&install, backupDir, fileName); err != nil {
		return err
	}

	record := &model.BackupRecord{
		Type:       "app",
		Name:       req.Name,
		DetailName: req.DetailName,
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

func (u *BackupService) AppRecover(req dto.CommonRecover) error {
	app, err := appRepo.GetFirst(appRepo.WithKey(req.Name))
	if err != nil {
		return err
	}
	install, err := appInstallRepo.GetFirst(commonRepo.WithByName(req.DetailName), appInstallRepo.WithAppId(app.ID))
	if err != nil {
		return err
	}

	fileOp := files.NewFileOp()
	if !fileOp.Stat(req.File) {
		return errors.New(fmt.Sprintf("%s file is not exist", req.File))
	}
	if _, err := compose.Down(install.GetComposePath()); err != nil {
		return err
	}
	if err := handleAppRecover(&install, req.File, false); err != nil {
		return err
	}
	return nil
}

func handleAppBackup(install *model.AppInstall, backupDir, fileName string) error {
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

	remarkInfo, _ := json.Marshal(install)
	remarkInfoPath := fmt.Sprintf("%s/app.json", tmpDir)
	if err := fileOp.SaveFile(remarkInfoPath, string(remarkInfo), fs.ModePerm); err != nil {
		return err
	}

	appPath := fmt.Sprintf("%s/%s", install.GetPath(), install.Name)
	if err := handleTar(appPath, tmpDir, "app.tar.gz", ""); err != nil {
		return err
	}

	resource, _ := appInstallResourceRepo.GetFirst(appInstallResourceRepo.WithAppInstallId(install.ID))
	if resource.ID != 0 && resource.ResourceId != 0 {
		mysqlInfo, err := appInstallRepo.LoadBaseInfo(constant.AppMysql, "")
		if err != nil {
			return err
		}
		db, err := mysqlRepo.Get(commonRepo.WithByID(resource.ResourceId))
		if err != nil {
			return err
		}
		if err := handleMysqlBackup(mysqlInfo, tmpDir, db.Name, fmt.Sprintf("%s.sql.gz", install.Name)); err != nil {
			return err
		}
	}

	if err := handleTar(tmpDir, backupDir, fileName, ""); err != nil {
		return err
	}
	return nil
}

func handleAppRecover(install *model.AppInstall, recoverFile string, isRollback bool) error {
	isOk := false
	fileOp := files.NewFileOp()
	if err := handleUnTar(recoverFile, path.Dir(recoverFile)); err != nil {
		return err
	}
	tmpPath := strings.ReplaceAll(recoverFile, ".tar.gz", "")
	defer func() {
		_, _ = compose.Up(install.GetComposePath())
		_ = os.RemoveAll(strings.ReplaceAll(recoverFile, ".tar.gz", ""))
	}()

	if !fileOp.Stat(tmpPath+"/app.json") || !fileOp.Stat(tmpPath+"/app.tar.gz") {
		return errors.New("the wrong recovery package does not have app.json or app.tar.gz files")
	}
	var oldInstall model.AppInstall
	appjson, err := os.ReadFile(tmpPath + "/app.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(appjson, &oldInstall); err != nil {
		return fmt.Errorf("unmarshal app.json failed, err: %v", err)
	}
	if oldInstall.App.Key != install.App.Key || oldInstall.Name != install.Name || oldInstall.Version != install.Version || oldInstall.ID != install.ID {
		return errors.New("the current backup file does not match the application")
	}

	if !isRollback {
		rollbackFile := fmt.Sprintf("%s/original/app/%s_%s.tar.gz", global.CONF.System.BaseDir, install.Name, time.Now().Format("20060102150405"))
		if err := handleAppBackup(install, path.Dir(rollbackFile), path.Base(rollbackFile)); err != nil {
			return fmt.Errorf("backup app %s for rollback before recover failed, err: %v", install.Name, err)
		}
		defer func() {
			if !isOk {
				global.LOG.Info("recover failed, start to rollback now")
				if err := handleAppRecover(install, rollbackFile, true); err != nil {
					global.LOG.Errorf("rollback app %s from %s failed, err: %v", install.Name, rollbackFile, err)
					return
				}
				global.LOG.Infof("rollback app %s from %s successful", install.Name, rollbackFile)
				_ = os.RemoveAll(rollbackFile)
			} else {
				_ = os.RemoveAll(rollbackFile)
			}
		}()
	}

	resource, _ := appInstallResourceRepo.GetFirst(appInstallResourceRepo.WithAppInstallId(install.ID))
	if resource.ID != 0 && install.App.Key != "mysql" {
		mysqlInfo, err := appInstallRepo.LoadBaseInfo(resource.Key, "")
		if err != nil {
			return err
		}
		db, err := mysqlRepo.Get(commonRepo.WithByID(resource.ResourceId))
		if err != nil {
			return err
		}
		if err := handleMysqlRecover(mysqlInfo, tmpPath, db.Name, fmt.Sprintf("%s.sql.gz", install.Name), true); err != nil {
			global.LOG.Errorf("handle recover from sql.gz failed, err: %v", err)
			return err
		}
	}

	if err := handleUnTar(tmpPath+"/app.tar.gz", fmt.Sprintf("%s/%s", constant.AppInstallDir, install.App.Key)); err != nil {
		global.LOG.Errorf("handle recover from app.tar.gz failed, err: %v", err)
		return err
	}

	oldInstall.Status = constant.Running
	if err := appInstallRepo.Save(context.Background(), install); err != nil {
		global.LOG.Errorf("save db app install failed, err: %v", err)
		return err
	}
	isOk = true
	return nil
}
