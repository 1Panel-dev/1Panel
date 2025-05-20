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

	"github.com/1Panel-dev/1Panel/agent/app/repo"

	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/i18n"

	"github.com/1Panel-dev/1Panel/agent/buserr"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/pkg/errors"
)

func (u *BackupService) AppBackup(req dto.CommonBackup) (*model.BackupRecord, error) {
	app, err := appRepo.GetFirst(appRepo.WithKey(req.Name))
	if err != nil {
		return nil, err
	}
	install, err := appInstallRepo.GetFirst(repo.WithByName(req.DetailName), appInstallRepo.WithAppId(app.ID))
	if err != nil {
		return nil, err
	}
	timeNow := time.Now().Format(constant.DateTimeSlimLayout)
	itemDir := fmt.Sprintf("app/%s/%s", req.Name, req.DetailName)
	backupDir := path.Join(global.Dir.LocalBackupDir, itemDir)

	fileName := req.FileName
	if req.FileName == "" {
		fileName = fmt.Sprintf("%s_%s.tar.gz", req.DetailName, timeNow+common.RandStrAndNum(5))
	}

	backupApp := func() (*model.BackupRecord, error) {
		if err = handleAppBackup(&install, nil, backupDir, fileName, "", req.Secret, req.TaskID); err != nil {
			global.LOG.Errorf("backup app %s failed, err: %v", req.DetailName, err)
			return nil, err
		}
		record := &model.BackupRecord{
			Type:              "app",
			Name:              req.Name,
			DetailName:        req.DetailName,
			SourceAccountIDs:  "1",
			DownloadAccountID: 1,
			FileDir:           itemDir,
			FileName:          fileName,
			Description:       req.Description,
		}
		if err := backupRepo.CreateRecord(record); err != nil {
			global.LOG.Errorf("save backup record failed, err: %v", err)
			return nil, err
		}
		return record, nil
	}

	if req.TaskID != "" {
		go func() {
			_, _ = backupApp()
		}()
	} else {
		return backupApp()
	}
	return nil, nil
}

func (u *BackupService) AppRecover(req dto.CommonRecover) error {
	app, err := appRepo.GetFirst(appRepo.WithKey(req.Name))
	if err != nil {
		return err
	}
	install, err := appInstallRepo.GetFirst(repo.WithByName(req.DetailName), appInstallRepo.WithAppId(app.ID))
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
	go func() {
		if err := handleAppRecover(&install, nil, req.File, false, req.Secret, req.TaskID); err != nil {
			global.LOG.Errorf("recover app %s failed, err: %v", req.DetailName, err)
		}
	}()
	return nil
}

func backupDatabaseWithTask(parentTask *task.Task, resourceKey, tmpDir, name string, databaseID uint) error {
	switch resourceKey {
	case constant.AppMysql, constant.AppMariaDB:
		db, err := mysqlRepo.Get(repo.WithByID(databaseID))
		if err != nil {
			return err
		}
		parentTask.LogStart(task.GetTaskName(db.Name, task.TaskBackup, task.TaskScopeDatabase))
		databaseHelper := DatabaseHelper{Database: db.MysqlName, DBType: resourceKey, Name: db.Name}
		if err := handleMysqlBackup(databaseHelper, parentTask, tmpDir, fmt.Sprintf("%s.sql.gz", name), ""); err != nil {
			return err
		}
		parentTask.LogSuccess(task.GetTaskName(db.Name, task.TaskBackup, task.TaskScopeDatabase))
	case constant.AppPostgresql:
		db, err := postgresqlRepo.Get(repo.WithByID(databaseID))
		if err != nil {
			return err
		}
		parentTask.LogStart(task.GetTaskName(db.Name, task.TaskBackup, task.TaskScopeDatabase))
		databaseHelper := DatabaseHelper{Database: db.PostgresqlName, DBType: resourceKey, Name: db.Name}
		if err := handlePostgresqlBackup(databaseHelper, parentTask, tmpDir, fmt.Sprintf("%s.sql.gz", name), ""); err != nil {
			return err
		}
		parentTask.LogSuccess(task.GetTaskName(db.Name, task.TaskBackup, task.TaskScopeDatabase))
	}
	return nil
}

func handleAppBackup(install *model.AppInstall, parentTask *task.Task, backupDir, fileName, excludes, secret, taskID string) error {
	var (
		err        error
		backupTask *task.Task
	)
	backupTask = parentTask
	if parentTask == nil {
		backupTask, err = task.NewTaskWithOps(install.Name, task.TaskBackup, task.TaskScopeApp, taskID, install.ID)
		if err != nil {
			return err
		}
	}

	itemHandler := func() error { return doAppBackup(install, backupTask, backupDir, fileName, excludes, secret) }
	backupTask.AddSubTask(task.GetTaskName(install.Name, task.TaskBackup, task.TaskScopeApp), func(t *task.Task) error { return itemHandler() }, nil)
	if parentTask != nil {
		return itemHandler()
	}
	return backupTask.Execute()
}

func handleAppRecover(install *model.AppInstall, parentTask *task.Task, recoverFile string, isRollback bool, secret, taskID string) error {
	var (
		err          error
		recoverTask  *task.Task
		isOk         = false
		rollbackFile string
	)
	recoverTask = parentTask
	if parentTask == nil {
		recoverTask, err = task.NewTaskWithOps(install.Name, task.TaskRecover, task.TaskScopeApp, taskID, install.ID)
		if err != nil {
			return err
		}
	}

	recoverApp := func(t *task.Task) error {
		fileOp := files.NewFileOp()
		if !isRollback {
			rollbackFile = path.Join(global.Dir.TmpDir, fmt.Sprintf("app/%s_%s.tar.gz", install.Name, time.Now().Format(constant.DateTimeSlimLayout)))
			if err := handleAppBackup(install, recoverTask, path.Dir(rollbackFile), path.Base(rollbackFile), "", "", taskID); err != nil {
				t.Log(fmt.Sprintf("backup app %s for rollback before recover failed, err: %v", install.Name, err))
			}
		}

		if err := fileOp.TarGzExtractPro(recoverFile, path.Dir(recoverFile), secret); err != nil {
			return err
		}
		tmpPath := strings.ReplaceAll(recoverFile, ".tar.gz", "")
		defer func() {
			_, _ = compose.Up(install.GetComposePath())
			_ = os.RemoveAll(strings.ReplaceAll(recoverFile, ".tar.gz", ""))
		}()

		if !fileOp.Stat(tmpPath+"/app.json") || !fileOp.Stat(tmpPath+"/app.tar.gz") {
			return errors.New(i18n.GetMsgByKey("AppBackupFileIncomplete"))
		}
		var oldInstall model.AppInstall
		appJson, err := os.ReadFile(tmpPath + "/app.json")
		if err != nil {
			return err
		}
		if err := json.Unmarshal(appJson, &oldInstall); err != nil {
			return fmt.Errorf("unmarshal app.json failed, err: %v", err)
		}
		if oldInstall.App.Key != install.App.Key || oldInstall.Name != install.Name {
			return errors.New(i18n.GetMsgByKey("AppAttributesNotMatch"))
		}

		newEnvFile := ""
		resources, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithAppInstallId(install.ID))
		for _, resource := range resources {
			var database model.Database
			switch resource.From {
			case constant.AppResourceRemote:
				database, err = databaseRepo.Get(repo.WithByID(resource.LinkId))
				if err != nil {
					return err
				}
			case constant.AppResourceLocal:
				resourceApp, err := appInstallRepo.GetFirst(repo.WithByID(resource.LinkId))
				if err != nil {
					return err
				}
				database, err = databaseRepo.Get(databaseRepo.WithAppInstallID(resourceApp.ID), repo.WithByType(resource.Key), repo.WithByFrom(constant.AppResourceLocal), repo.WithByName(resourceApp.Name))
				if err != nil {
					return err
				}
			}
			switch database.Type {
			case constant.AppPostgresql:
				db, err := postgresqlRepo.Get(repo.WithByID(resource.ResourceId))
				if err != nil {
					return err
				}
				taskName := task.GetTaskName(db.Name, task.TaskRecover, task.TaskScopeDatabase)
				t.LogStart(taskName)
				if err := handlePostgresqlRecover(dto.CommonRecover{
					Name:       database.Name,
					DetailName: db.Name,
					File:       fmt.Sprintf("%s/%s.sql.gz", tmpPath, install.Name),
				}, parentTask, true); err != nil {
					t.LogFailedWithErr(taskName, err)
					return err
				}
				t.LogSuccess(taskName)
			case constant.AppMysql, constant.AppMariaDB:
				db, err := mysqlRepo.Get(repo.WithByID(resource.ResourceId))
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
				_ = appInstallResourceRepo.BatchUpdateBy(map[string]interface{}{"resource_id": newDB.ID}, repo.WithByID(resource.ID))
				taskName := task.GetTaskName(db.Name, task.TaskRecover, task.TaskScopeDatabase)
				t.LogStart(taskName)
				if err := handleMysqlRecover(dto.CommonRecover{
					Name:       newDB.MysqlName,
					DetailName: newDB.Name,
					File:       fmt.Sprintf("%s/%s.sql.gz", tmpPath, install.Name),
				}, parentTask, true, ""); err != nil {
					t.LogFailedWithErr(taskName, err)
					return err
				}
				t.LogSuccess(taskName)
			}
		}

		appDir := install.GetPath()
		backPath := fmt.Sprintf("%s_bak", appDir)
		_ = fileOp.Rename(appDir, backPath)
		_ = fileOp.CreateDir(appDir, constant.DirPerm)

		deCompressName := i18n.GetWithName("DeCompressFile", "app.tar.gz")
		t.LogStart(deCompressName)
		if err := fileOp.TarGzExtractPro(tmpPath+"/app.tar.gz", install.GetAppPath(), ""); err != nil {
			t.LogFailedWithErr(deCompressName, err)
			_ = fileOp.DeleteDir(appDir)
			_ = fileOp.Rename(backPath, appDir)
			return err
		}
		t.LogSuccess(deCompressName)
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

	rollBackApp := func(t *task.Task) {
		if isRollback {
			return
		}
		if !isOk {
			t.Log(i18n.GetMsgByKey("RecoverFailedStartRollBack"))
			if err := handleAppRecover(install, t, rollbackFile, true, "", ""); err != nil {
				t.LogFailedWithErr(i18n.GetMsgByKey("Rollback"), err)
				return
			}
			t.LogSuccess(i18n.GetMsgByKey("Rollback"))
			_ = os.RemoveAll(rollbackFile)
		} else {
			_ = os.RemoveAll(rollbackFile)
		}
	}

	recoverTask.AddSubTask(task.GetTaskName(install.Name, task.TaskRecover, task.TaskScopeApp), recoverApp, rollBackApp)
	if parentTask != nil {
		return recoverApp(parentTask)
	}
	return recoverTask.Execute()
}

func doAppBackup(install *model.AppInstall, parentTask *task.Task, backupDir, fileName, excludes, secret string) error {
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
	if err := fileOp.TarGzCompressPro(true, appPath, path.Join(tmpDir, "app.tar.gz"), "", excludes); err != nil {
		return err
	}

	resources, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithAppInstallId(install.ID))
	for _, resource := range resources {
		if err := backupDatabaseWithTask(parentTask, resource.Key, tmpDir, install.Name, resource.ResourceId); err != nil {
			return err
		}
	}
	parentTask.LogStart(i18n.GetMsgByKey("CompressDir"))
	if err := fileOp.TarGzCompressPro(true, tmpDir, path.Join(backupDir, fileName), secret, ""); err != nil {
		return err
	}
	parentTask.Log(i18n.GetWithName("CompressFileSuccess", fileName))
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
