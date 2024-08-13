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

	"github.com/1Panel-dev/1Panel/backend/buserr"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/pkg/errors"
)

func (u *BackupService) AppBackup(req dto.CommonBackup) (*model.BackupRecord, error) {
	localDir, err := loadLocalDir()
	if err != nil {
		return nil, err
	}
	app, err := appRepo.GetFirst(appRepo.WithKey(req.Name))
	if err != nil {
		return nil, err
	}
	install, err := appInstallRepo.GetFirst(commonRepo.WithByName(req.DetailName), appInstallRepo.WithAppId(app.ID))
	if err != nil {
		return nil, err
	}
	timeNow := time.Now().Format(constant.DateTimeSlimLayout)
	itemDir := fmt.Sprintf("app/%s/%s", req.Name, req.DetailName)
	backupDir := path.Join(localDir, itemDir)

	fileName := req.FileName
	if req.FileName == "" {
		fileName = fmt.Sprintf("%s_%s.tar.gz", req.DetailName, timeNow+common.RandStrAndNum(5))
	}
	if err := handleAppBackup(&install, backupDir, fileName, "", req.Secret); err != nil {
		return nil, err
	}

	record := &model.BackupRecord{
		Type:       "app",
		Name:       req.Name,
		DetailName: req.DetailName,
		Source:     "LOCAL",
		BackupType: "LOCAL",
		FileDir:    itemDir,
		FileName:   fileName,
	}

	if err := backupRepo.CreateRecord(record); err != nil {
		global.LOG.Errorf("save backup record failed, err: %v", err)
		return nil, err
	}
	return record, nil
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
		return buserr.WithName("ErrFileNotFound", req.File)
	}
	if _, err := compose.Down(install.GetComposePath()); err != nil {
		return err
	}
	if err := handleAppRecover(&install, req.File, false, req.Secret); err != nil {
		return err
	}
	return nil
}

func handleAppBackup(install *model.AppInstall, backupDir, fileName string, excludes string, secret string) error {
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

	appPath := install.GetPath()
	if err := handleTar(appPath, tmpDir, "app.tar.gz", excludes, ""); err != nil {
		return err
	}

	resources, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithAppInstallId(install.ID))
	for _, resource := range resources {
		switch resource.Key {
		case constant.AppMysql, constant.AppMariaDB:
			db, err := mysqlRepo.Get(commonRepo.WithByID(resource.ResourceId))
			if err != nil {
				return err
			}
			if err := handleMysqlBackup(db.MysqlName, resource.Key, db.Name, tmpDir, fmt.Sprintf("%s.sql.gz", install.Name)); err != nil {
				return err
			}
		case constant.AppPostgresql:
			db, err := postgresqlRepo.Get(commonRepo.WithByID(resource.ResourceId))
			if err != nil {
				return err
			}
			if err := handlePostgresqlBackup(db.PostgresqlName, db.Name, tmpDir, fmt.Sprintf("%s.sql.gz", install.Name)); err != nil {
				return err
			}
		}
	}

	if err := handleTar(tmpDir, backupDir, fileName, "", secret); err != nil {
		return err
	}
	return nil
}

func handleAppRecover(install *model.AppInstall, recoverFile string, isRollback bool, secret string) error {
	isOk := false
	fileOp := files.NewFileOp()
	if err := handleUnTar(recoverFile, path.Dir(recoverFile), secret); err != nil {
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
	if oldInstall.App.Key != install.App.Key || oldInstall.Name != install.Name {
		return errors.New("the current backup file does not match the application")
	}

	if !isRollback {
		rollbackFile := path.Join(global.CONF.System.TmpDir, fmt.Sprintf("app/%s_%s.tar.gz", install.Name, time.Now().Format(constant.DateTimeSlimLayout)))
		if err := handleAppBackup(install, path.Dir(rollbackFile), path.Base(rollbackFile), "", ""); err != nil {
			return fmt.Errorf("backup app %s for rollback before recover failed, err: %v", install.Name, err)
		}
		defer func() {
			if !isOk {
				global.LOG.Info("recover failed, start to rollback now")
				if err := handleAppRecover(install, rollbackFile, true, secret); err != nil {
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

	newEnvFile := ""
	resources, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithAppInstallId(install.ID))
	for _, resource := range resources {
		var database model.Database
		switch resource.From {
		case constant.AppResourceRemote:
			database, err = databaseRepo.Get(commonRepo.WithByID(resource.LinkId))
			if err != nil {
				return err
			}
		case constant.AppResourceLocal:
			resourceApp, err := appInstallRepo.GetFirst(commonRepo.WithByID(resource.LinkId))
			if err != nil {
				return err
			}
			database, err = databaseRepo.Get(databaseRepo.WithAppInstallID(resourceApp.ID), commonRepo.WithByType(resource.Key), databaseRepo.WithByFrom(constant.AppResourceLocal), commonRepo.WithByName(resourceApp.Name))
			if err != nil {
				return err
			}
		}
		switch database.Type {
		case constant.AppPostgresql:
			db, err := postgresqlRepo.Get(commonRepo.WithByID(resource.ResourceId))
			if err != nil {
				return err
			}
			if err := handlePostgresqlRecover(dto.CommonRecover{
				Name:       database.Name,
				DetailName: db.Name,
				File:       fmt.Sprintf("%s/%s.sql.gz", tmpPath, install.Name),
			}, true); err != nil {
				global.LOG.Errorf("handle recover from sql.gz failed, err: %v", err)
				return err
			}
		case constant.AppMysql, constant.AppMariaDB:
			db, err := mysqlRepo.Get(commonRepo.WithByID(resource.ResourceId))
			if err != nil {
				return err
			}
			newDB, envMap, err := reCreateDB(db.ID, database, oldInstall.Env)
			if err != nil {
				return err
			}
			oldHost := fmt.Sprintf("\"PANEL_DB_HOST\":\"%v\"", envMap["PANEL_DB_HOST"].(string))
			newHost := fmt.Sprintf("\"PANEL_DB_HOST\":\"%v\"", database.Address)
			oldInstall.Env = strings.ReplaceAll(oldInstall.Env, oldHost, newHost)
			envMap["PANEL_DB_HOST"] = database.Address
			newEnvFile, err = coverEnvJsonToStr(oldInstall.Env)
			if err != nil {
				return err
			}
			_ = appInstallResourceRepo.BatchUpdateBy(map[string]interface{}{"resource_id": newDB.ID}, commonRepo.WithByID(resource.ID))

			if err := handleMysqlRecover(dto.CommonRecover{
				Name:       newDB.MysqlName,
				DetailName: newDB.Name,
				File:       fmt.Sprintf("%s/%s.sql.gz", tmpPath, install.Name),
			}, true); err != nil {
				global.LOG.Errorf("handle recover from sql.gz failed, err: %v", err)
				return err
			}
		}
	}

	appDir := install.GetPath()
	backPath := fmt.Sprintf("%s_bak", appDir)
	_ = fileOp.Rename(appDir, backPath)
	_ = fileOp.CreateDir(appDir, 0755)

	if err := handleUnTar(tmpPath+"/app.tar.gz", install.GetAppPath(), ""); err != nil {
		global.LOG.Errorf("handle recover from app.tar.gz failed, err: %v", err)
		_ = fileOp.DeleteDir(appDir)
		_ = fileOp.Rename(backPath, appDir)
		return err
	}
	_ = fileOp.DeleteDir(backPath)

	if len(newEnvFile) != 0 {
		envPath := fmt.Sprintf("%s/%s/.env", install.GetAppPath(), install.Name)
		file, err := os.OpenFile(envPath, os.O_WRONLY|os.O_TRUNC, 0640)
		if err != nil {
			return err
		}
		defer file.Close()
		_, _ = file.WriteString(newEnvFile)
	}

	oldInstall.ID = install.ID
	oldInstall.Status = constant.StatusRunning
	oldInstall.AppId = install.AppId
	oldInstall.AppDetailId = install.AppDetailId
	oldInstall.App.ID = install.AppId
	if err := appInstallRepo.Save(context.Background(), &oldInstall); err != nil {
		global.LOG.Errorf("save db app install failed, err: %v", err)
		return err
	}
	isOk = true

	return nil
}

func reCreateDB(dbID uint, database model.Database, oldEnv string) (*model.DatabaseMysql, map[string]interface{}, error) {
	mysqlService := NewIMysqlService()
	ctx := context.Background()
	_ = mysqlService.Delete(ctx, dto.MysqlDBDelete{ID: dbID, Database: database.Name, Type: database.Type, DeleteBackup: false, ForceDelete: true})

	envMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(oldEnv), &envMap); err != nil {
		return nil, envMap, err
	}
	oldName, _ := envMap["PANEL_DB_NAME"].(string)
	oldUser, _ := envMap["PANEL_DB_USER"].(string)
	oldPassword, _ := envMap["PANEL_DB_USER_PASSWORD"].(string)
	createDB, err := mysqlService.Create(context.Background(), dto.MysqlDBCreate{
		Name:       oldName,
		From:       database.From,
		Database:   database.Name,
		Format:     "utf8mb4",
		Username:   oldUser,
		Password:   oldPassword,
		Permission: "%",
	})
	cronjobs, _ := cronjobRepo.List(cronjobRepo.WithByDbName(fmt.Sprintf("%v", dbID)))
	for _, job := range cronjobs {
		_ = cronjobRepo.Update(job.ID, map[string]interface{}{"db_name": fmt.Sprintf("%v", createDB.ID)})
	}
	if err != nil {
		return nil, envMap, err
	}
	return createDB, envMap, nil
}
