package service

import (
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/pkg/errors"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
)

func (u *BackupService) WebsiteBackup(req dto.CommonBackup) error {
	website, err := websiteRepo.GetFirst(websiteRepo.WithAlias(req.DetailName))
	if err != nil {
		return err
	}

	timeNow := time.Now().Format(constant.DateTimeSlimLayout)
	itemDir := fmt.Sprintf("website/%s", req.Name)
	backupDir := path.Join(global.CONF.System.Backup, itemDir)
	fileName := fmt.Sprintf("%s_%s.tar.gz", website.Alias, timeNow+common.RandStrAndNum(5))

	go func() {
		if err = handleWebsiteBackup(&website, backupDir, fileName, "", req.Secret, req.TaskID); err != nil {
			global.LOG.Errorf("backup website %s failed, err: %v", website.Alias, err)
			return
		}
		record := &model.BackupRecord{
			Type:              "website",
			Name:              website.Alias,
			DetailName:        req.DetailName,
			SourceAccountIDs:  "1",
			DownloadAccountID: 1,
			FileDir:           itemDir,
			FileName:          fileName,
		}
		if err = backupRepo.CreateRecord(record); err != nil {
			global.LOG.Errorf("save backup record failed, err: %v", err)
			return
		}
	}()
	return nil
}

func (u *BackupService) WebsiteRecover(req dto.CommonRecover) error {
	website, err := websiteRepo.GetFirst(websiteRepo.WithAlias(req.DetailName))
	if err != nil {
		return err
	}
	go func() {
		if err := handleWebsiteRecover(&website, req.File, false, req.Secret, req.TaskID); err != nil {
			global.LOG.Errorf("recover website %s failed, err: %v", website.Alias, err)
		}
	}()
	return nil
}

func handleWebsiteRecover(website *model.Website, recoverFile string, isRollback bool, secret, taskID string) error {
	recoverTask, err := task.NewTaskWithOps(website.PrimaryDomain, task.TaskRecover, task.TaskScopeWebsite, taskID, website.ID)
	if err != nil {
		return err
	}
	recoverTask.AddSubTask(task.GetTaskName(website.PrimaryDomain, task.TaskRecover, task.TaskScopeWebsite), func(t *task.Task) error {
		fileOp := files.NewFileOp()
		tmpPath := strings.ReplaceAll(recoverFile, ".tar.gz", "")
		t.Log(i18n.GetWithName("DeCompressFile", recoverFile))
		if err = handleUnTar(recoverFile, path.Dir(recoverFile), secret); err != nil {
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
		if err = json.Unmarshal(websiteJson, &oldWebsite); err != nil {
			return fmt.Errorf("unmarshal app.json failed, err: %v", err)
		}

		if err = checkValidOfWebsite(&oldWebsite, website); err != nil {
			t.Log(i18n.GetWithName("ErrCheckValid", err.Error()))
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
			if err := handleWebsiteBackup(website, path.Dir(rollbackFile), path.Base(rollbackFile), "", "", ""); err != nil {
				return fmt.Errorf("backup website %s for rollback before recover failed, err: %v", website.Alias, err)
			}
			defer func() {
				if !isOk {
					t.LogStart(i18n.GetMsgByKey("Rollback"))
					if err := handleWebsiteRecover(website, rollbackFile, true, "", taskID); err != nil {
						t.LogFailedWithErr(i18n.GetMsgByKey("Rollback"), err)
						return
					}
					t.LogSuccess(i18n.GetMsgByKey("Rollback"))
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
		if err = fileOp.CopyFile(fmt.Sprintf("%s/%s.conf", tmpPath, website.Alias), GetOpenrestyDir(SiteConfDir)); err != nil {
			return err
		}

		switch website.Type {
		case constant.Deployment:
			app, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
			if err != nil {
				return err
			}
			taskName := task.GetTaskName(app.Name, task.TaskRecover, task.TaskScopeApp)
			t.LogStart(taskName)
			if err := handleAppRecover(&app, recoverTask, fmt.Sprintf("%s/%s.app.tar.gz", tmpPath, website.Alias), true, "", ""); err != nil {
				t.LogFailedWithErr(taskName, err)
				return err
			}
			t.LogSuccess(taskName)
			if _, err = compose.DownAndUp(fmt.Sprintf("%s/%s/%s/docker-compose.yml", constant.AppInstallDir, app.App.Key, app.Name)); err != nil {
				t.LogFailedWithErr("Run", err)
				return err
			}
		case constant.Runtime:
			runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(website.RuntimeID))
			if err != nil {
				return err
			}
			taskName := task.GetTaskName(runtime.Name, task.TaskRecover, task.TaskScopeRuntime)
			t.LogStart(taskName)
			if err := handleRuntimeRecover(runtime, fmt.Sprintf("%s/%s.runtime.tar.gz", tmpPath, website.Alias), true, ""); err != nil {
				t.LogFailedWithErr(taskName, err)
				return err
			}
			t.LogSuccess(taskName)
			if oldWebsite.DbID > 0 {
				if err := recoverWebsiteDatabase(t, oldWebsite.DbID, oldWebsite.DbType, tmpPath, website.Alias); err != nil {
					return err
				}
			}
		case constant.Static:
			if oldWebsite.DbID > 0 {
				if err := recoverWebsiteDatabase(t, oldWebsite.DbID, oldWebsite.DbType, tmpPath, website.Alias); err != nil {
					return err
				}
			}
		}
		taskName := i18n.GetMsgByKey("TaskRecover") + i18n.GetMsgByKey("websiteDir")
		t.Log(taskName)
		if err = handleUnTar(fmt.Sprintf("%s/%s.web.tar.gz", tmpPath, website.Alias), GetSitePath(*website, SiteDir), ""); err != nil {
			t.LogFailedWithErr(taskName, err)
			return err
		}
		stdout, err := cmd.Execf("docker exec -i %s nginx -s reload", nginxInfo.ContainerName)
		if err != nil {
			return errors.New(stdout)
		}
		oldWebsite.ID = website.ID
		if err := websiteRepo.SaveWithoutCtx(&oldWebsite); err != nil {
			return err
		}
		isOk = true
		return nil
	}, nil)
	return recoverTask.Execute()
}

func handleWebsiteBackup(website *model.Website, backupDir, fileName, excludes, secret, taskID string) error {
	backupTask, err := task.NewTaskWithOps(website.Alias, task.TaskBackup, task.TaskScopeWebsite, taskID, website.ID)
	if err != nil {
		return err
	}
	backupTask.AddSubTask(task.GetTaskName(website.Alias, task.TaskBackup, task.TaskScopeWebsite), func(t *task.Task) error {
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
		if err = fileOp.SaveFile(tmpDir+"/website.json", string(remarkInfo), fs.ModePerm); err != nil {
			return err
		}
		nginxConfFile := GetSitePath(*website, SiteConf)
		if err = fileOp.CopyFile(nginxConfFile, tmpDir); err != nil {
			return err
		}
		t.Log(i18n.GetMsgByKey("BackupNginxConfig"))

		switch website.Type {
		case constant.Deployment:
			app, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
			if err != nil {
				return err
			}
			t.LogStart(task.GetTaskName(app.Name, task.TaskBackup, task.TaskScopeApp))
			if err = handleAppBackup(&app, backupTask, tmpDir, fmt.Sprintf("%s.app.tar.gz", website.Alias), excludes, "", ""); err != nil {
				return err
			}
			t.LogSuccess(task.GetTaskName(app.Name, task.TaskBackup, task.TaskScopeApp))
		case constant.Runtime:
			runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(website.RuntimeID))
			if err != nil {
				return err
			}
			t.LogStart(task.GetTaskName(runtime.Name, task.TaskBackup, task.TaskScopeRuntime))
			if err = handleRuntimeBackup(runtime, tmpDir, fmt.Sprintf("%s.runtime.tar.gz", website.Alias), excludes, ""); err != nil {
				return err
			}
			t.LogSuccess(task.GetTaskName(runtime.Name, task.TaskBackup, task.TaskScopeRuntime))
			if website.DbID > 0 {
				if err = backupDatabaseWithTask(t, website.DbType, tmpDir, website.Alias, website.DbID); err != nil {
					return err
				}
			}
		case constant.Static:
			if website.DbID > 0 {
				if err = backupDatabaseWithTask(t, website.DbType, tmpDir, website.Alias, website.DbID); err != nil {
					return err
				}
			}
		}

		websiteDir := GetSitePath(*website, SiteDir)
		t.LogStart(i18n.GetMsgByKey("CompressDir"))
		if err = handleTar(websiteDir, tmpDir, fmt.Sprintf("%s.web.tar.gz", website.Alias), excludes, ""); err != nil {
			return err
		}
		if err = handleTar(tmpDir, backupDir, fileName, "", secret); err != nil {
			return err
		}
		t.Log(i18n.GetWithName("CompressFileSuccess", fileName))
		return nil
	}, nil)
	return backupTask.Execute()
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

func recoverWebsiteDatabase(t *task.Task, dbID uint, dbType, tmpPath, websiteKey string) error {
	switch dbType {
	case constant.AppPostgresql:
		db, err := postgresqlRepo.Get(commonRepo.WithByID(dbID))
		if err != nil {
			return err
		}
		taskName := task.GetTaskName(db.Name, task.TaskRecover, task.TaskScopeDatabase)
		t.LogStart(taskName)
		if err := handlePostgresqlRecover(dto.CommonRecover{
			Name:       db.PostgresqlName,
			DetailName: db.Name,
			File:       fmt.Sprintf("%s/%s.sql.gz", tmpPath, websiteKey),
		}, true); err != nil {
			t.LogFailedWithErr(taskName, err)
			return err
		}
		t.LogSuccess(taskName)
	case constant.AppMysql, constant.AppMariaDB:
		db, err := mysqlRepo.Get(commonRepo.WithByID(dbID))
		if err != nil {
			return err
		}
		taskName := task.GetTaskName(db.Name, task.TaskRecover, task.TaskScopeDatabase)
		t.LogStart(taskName)
		if err := handleMysqlRecover(dto.CommonRecover{
			Name:       db.MysqlName,
			DetailName: db.Name,
			File:       fmt.Sprintf("%s/%s.sql.gz", tmpPath, websiteKey),
		}, true); err != nil {
			t.LogFailedWithErr(taskName, err)
			return err
		}
		t.LogSuccess(taskName)
	}
	return nil
}
