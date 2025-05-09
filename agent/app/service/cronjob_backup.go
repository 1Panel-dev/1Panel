package service

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/i18n"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/pkg/errors"
)

func (u *CronjobService) handleApp(cronjob model.Cronjob, startTime time.Time, taskItem *task.Task) error {
	var apps []model.AppInstall
	if cronjob.AppID == "all" {
		apps, _ = appInstallRepo.ListBy(context.Background())
	} else {
		itemID, _ := strconv.Atoi(cronjob.AppID)
		app, err := appInstallRepo.GetFirst(repo.WithByID(uint(itemID)))
		if err != nil {
			return err
		}
		apps = append(apps, app)
	}
	if len(apps) == 0 {
		return errors.New("no such app in database!")
	}
	accountMap, err := NewBackupClientMap(strings.Split(cronjob.SourceAccountIDs, ","))
	if err != nil {
		return err
	}
	for _, app := range apps {
		taskItem.AddSubTaskWithOps(task.GetTaskName(app.Name, task.TaskBackup, task.TaskScopeCronjob), func(task *task.Task) error {
			var record model.BackupRecord
			record.From = "cronjob"
			record.Type = "app"
			record.CronjobID = cronjob.ID
			record.Name = app.App.Key
			record.DetailName = app.Name
			record.DownloadAccountID, record.SourceAccountIDs = cronjob.DownloadAccountID, cronjob.SourceAccountIDs
			backupDir := path.Join(global.Dir.TmpDir, fmt.Sprintf("app/%s/%s", app.App.Key, app.Name))
			record.FileName = fmt.Sprintf("app_%s_%s.tar.gz", app.Name, startTime.Format(constant.DateTimeSlimLayout)+common.RandStrAndNum(5))
			if err := doAppBackup(&app, task, backupDir, record.FileName, cronjob.ExclusionRules, cronjob.Secret); err != nil {
				return err
			}
			downloadPath, err := u.uploadCronjobBackFile(cronjob, accountMap, path.Join(backupDir, record.FileName))
			if err != nil {
				return err
			}
			record.FileDir = path.Dir(downloadPath)
			if err := backupRepo.CreateRecord(&record); err != nil {
				global.LOG.Errorf("save backup record failed, err: %v", err)
				return err
			}
			u.removeExpiredBackup(cronjob, accountMap, record)
			return nil
		}, nil, int(cronjob.RetryTimes), time.Duration(cronjob.Timeout)*time.Second)
	}
	return nil
}

func (u *CronjobService) handleWebsite(cronjob model.Cronjob, startTime time.Time, taskItem *task.Task) error {
	webs := loadWebsForJob(cronjob)
	if len(webs) == 0 {
		return errors.New("no such website in database!")
	}
	accountMap, err := NewBackupClientMap(strings.Split(cronjob.SourceAccountIDs, ","))
	if err != nil {
		return err
	}
	for _, web := range webs {
		taskItem.AddSubTaskWithOps(task.GetTaskName(web.Alias, task.TaskBackup, task.TaskScopeCronjob), func(task *task.Task) error {
			var record model.BackupRecord
			record.From = "cronjob"
			record.Type = "website"
			record.CronjobID = cronjob.ID
			record.Name = web.Alias
			record.DetailName = web.Alias
			record.DownloadAccountID, record.SourceAccountIDs = cronjob.DownloadAccountID, cronjob.SourceAccountIDs
			backupDir := path.Join(global.Dir.TmpDir, fmt.Sprintf("website/%s", web.Alias))
			record.FileName = fmt.Sprintf("website_%s_%s.tar.gz", web.Alias, startTime.Format(constant.DateTimeSlimLayout)+common.RandStrAndNum(5))

			if err := doWebsiteBackup(&web, taskItem, backupDir, record.FileName, cronjob.ExclusionRules, cronjob.Secret); err != nil {
				return err
			}

			downloadPath, err := u.uploadCronjobBackFile(cronjob, accountMap, path.Join(backupDir, record.FileName))
			if err != nil {
				return err
			}
			record.FileDir = path.Dir(downloadPath)
			if err := backupRepo.CreateRecord(&record); err != nil {
				global.LOG.Errorf("save backup record failed, err: %v", err)
				return err
			}
			u.removeExpiredBackup(cronjob, accountMap, record)
			return nil
		}, nil, int(cronjob.RetryTimes), time.Duration(cronjob.Timeout)*time.Second)
		return nil
	}
	return nil
}

func (u *CronjobService) handleDatabase(cronjob model.Cronjob, startTime time.Time, taskItem *task.Task) error {
	dbs := loadDbsForJob(cronjob)
	if len(dbs) == 0 {
		return errors.New("no such db in database!")
	}
	accountMap, err := NewBackupClientMap(strings.Split(cronjob.SourceAccountIDs, ","))
	if err != nil {
		return err
	}
	for _, dbInfo := range dbs {
		itemName := fmt.Sprintf("%s[%s] - %s", dbInfo.Database, dbInfo.DBType, dbInfo.Name)
		taskItem.AddSubTaskWithOps(task.GetTaskName(itemName, task.TaskBackup, task.TaskScopeCronjob), func(task *task.Task) error {
			var record model.BackupRecord
			record.From = "cronjob"
			record.Type = dbInfo.DBType
			record.CronjobID = cronjob.ID
			record.Name = dbInfo.Database
			record.DetailName = dbInfo.Name
			record.DownloadAccountID, record.SourceAccountIDs = cronjob.DownloadAccountID, cronjob.SourceAccountIDs

			backupDir := path.Join(global.Dir.TmpDir, fmt.Sprintf("database/%s/%s/%s", dbInfo.DBType, record.Name, dbInfo.Name))
			record.FileName = fmt.Sprintf("db_%s_%s.sql.gz", dbInfo.Name, startTime.Format(constant.DateTimeSlimLayout)+common.RandStrAndNum(5))
			if cronjob.DBType == "mysql" || cronjob.DBType == "mariadb" {
				if err := doMysqlBackup(dbInfo, backupDir, record.FileName); err != nil {
					return err
				}
			} else {
				if err := doPostgresqlgBackup(dbInfo, backupDir, record.FileName); err != nil {
					return err
				}
			}

			downloadPath, err := u.uploadCronjobBackFile(cronjob, accountMap, path.Join(backupDir, record.FileName))
			if err != nil {
				return err
			}
			record.FileDir = path.Dir(downloadPath)
			if err := backupRepo.CreateRecord(&record); err != nil {
				global.LOG.Errorf("save backup record failed, err: %v", err)
				return err
			}
			u.removeExpiredBackup(cronjob, accountMap, record)
			return nil
		}, nil, int(cronjob.RetryTimes), time.Duration(cronjob.Timeout)*time.Second)
	}
	return nil
}

func (u *CronjobService) handleDirectory(cronjob model.Cronjob, startTime time.Time, taskItem *task.Task) error {
	taskItem.AddSubTaskWithOps(task.GetTaskName(i18n.GetMsgByKey("BackupFileOrDir"), task.TaskBackup, task.TaskScopeCronjob), func(task *task.Task) error {
		accountMap, err := NewBackupClientMap(strings.Split(cronjob.SourceAccountIDs, ","))
		if err != nil {
			return err
		}
		fileName := fmt.Sprintf("%s.tar.gz", startTime.Format(constant.DateTimeSlimLayout)+common.RandStrAndNum(5))
		backupDir := path.Join(global.Dir.TmpDir, fmt.Sprintf("%s/%s", cronjob.Type, cronjob.Name))

		fileOp := files.NewFileOp()
		if cronjob.IsDir {
			taskItem.Logf("Dir: %s, Excludes: %s", cronjob.SourceDir, cronjob.ExclusionRules)
			if err := fileOp.TarGzCompressPro(true, cronjob.SourceDir, path.Join(backupDir, fileName), cronjob.Secret, cronjob.ExclusionRules); err != nil {
				return err
			}
		} else {
			taskItem.Logf("Files: %s", cronjob.SourceDir)
			fileLists := strings.Split(cronjob.SourceDir, ",")
			if err := fileOp.TarGzFilesWithCompressPro(fileLists, path.Join(backupDir, fileName), cronjob.Secret); err != nil {
				return err
			}
		}
		var record model.BackupRecord
		record.From = "cronjob"
		record.Type = "directory"
		record.CronjobID = cronjob.ID
		record.Name = cronjob.Name
		record.DownloadAccountID, record.SourceAccountIDs = cronjob.DownloadAccountID, cronjob.SourceAccountIDs
		downloadPath, err := u.uploadCronjobBackFile(cronjob, accountMap, path.Join(backupDir, fileName))
		if err != nil {
			taskItem.LogFailedWithErr("Upload backup file", err)
			return err
		}
		record.FileDir = path.Dir(downloadPath)
		record.FileName = fileName
		if err := backupRepo.CreateRecord(&record); err != nil {
			taskItem.LogFailedWithErr("Save record", err)
			return err
		}
		u.removeExpiredBackup(cronjob, accountMap, record)
		return nil
	}, nil, int(cronjob.RetryTimes), time.Duration(cronjob.Timeout)*time.Second)
	return nil
}

func (u *CronjobService) handleSystemLog(cronjob model.Cronjob, startTime time.Time, taskItem *task.Task) error {
	taskItem.AddSubTaskWithOps(task.GetTaskName(i18n.GetMsgByKey("BackupSystemLog"), task.TaskBackup, task.TaskScopeCronjob), func(task *task.Task) error {
		accountMap, err := NewBackupClientMap(strings.Split(cronjob.SourceAccountIDs, ","))
		if err != nil {
			return err
		}
		nameItem := startTime.Format(constant.DateTimeSlimLayout) + common.RandStrAndNum(5)
		fileName := fmt.Sprintf("system_log_%s.tar.gz", nameItem)
		backupDir := path.Join(global.Dir.TmpDir, "log", nameItem)
		if err := handleBackupLogs(taskItem, backupDir, fileName, cronjob.Secret); err != nil {
			return err
		}
		var record model.BackupRecord
		record.From = "cronjob"
		record.Type = "log"
		record.CronjobID = cronjob.ID
		record.Name = cronjob.Name
		record.DownloadAccountID, record.SourceAccountIDs = cronjob.DownloadAccountID, cronjob.SourceAccountIDs
		downloadPath, err := u.uploadCronjobBackFile(cronjob, accountMap, path.Join(path.Dir(backupDir), fileName))
		if err != nil {
			taskItem.LogFailedWithErr("Upload backup file", err)
			return err
		}
		record.FileDir = path.Dir(downloadPath)
		record.FileName = fileName
		if err := backupRepo.CreateRecord(&record); err != nil {
			taskItem.LogFailedWithErr("Save record", err)
			return err
		}
		u.removeExpiredBackup(cronjob, accountMap, record)
		return nil
	}, nil, int(cronjob.RetryTimes), time.Duration(cronjob.Timeout)*time.Second)
	return nil
}

func (u *CronjobService) handleSnapshot(cronjob model.Cronjob, jobRecord model.JobRecords, taskItem *task.Task) error {
	accountMap, err := NewBackupClientMap(strings.Split(cronjob.SourceAccountIDs, ","))
	if err != nil {
		return err
	}
	itemData, err := NewISnapshotService().LoadSnapshotData()
	if err != nil {
		return err
	}

	var record model.BackupRecord
	record.From = "cronjob"
	record.Type = "snapshot"
	record.CronjobID = cronjob.ID
	record.Name = cronjob.Name
	record.DownloadAccountID, record.SourceAccountIDs = cronjob.DownloadAccountID, cronjob.SourceAccountIDs
	record.FileDir = "system_snapshot"

	versionItem, _ := settingRepo.Get(settingRepo.WithByKey("SystemVersion"))
	scope := "core"
	if !global.IsMaster {
		scope = "agent"
	}
	req := dto.SnapshotCreate{
		Name:   fmt.Sprintf("snapshot-1panel-%s-%s-linux-%s-%s", scope, versionItem.Value, loadOs(), jobRecord.StartTime.Format(constant.DateTimeSlimLayout)+common.RandStrAndNum(5)),
		Secret: cronjob.Secret,
		TaskID: jobRecord.TaskID,

		SourceAccountIDs:  record.SourceAccountIDs,
		DownloadAccountID: cronjob.DownloadAccountID,
		AppData:           itemData.AppData,
		PanelData:         itemData.PanelData,
		BackupData:        itemData.BackupData,
		WithMonitorData:   true,
		WithLoginLog:      true,
		WithOperationLog:  true,
		WithSystemLog:     true,
		WithTaskLog:       true,
	}

	if err := NewISnapshotService().SnapshotCreate(taskItem, req, jobRecord.ID, cronjob.RetryTimes, cronjob.Timeout); err != nil {
		return err
	}
	record.FileName = req.Name + ".tar.gz"

	if err := backupRepo.CreateRecord(&record); err != nil {
		global.LOG.Errorf("save backup record failed, err: %v", err)
		return err
	}
	u.removeExpiredBackup(cronjob, accountMap, record)
	return nil
}

type DatabaseHelper struct {
	ID       uint
	DBType   string
	Database string
	Name     string
}

func loadDbsForJob(cronjob model.Cronjob) []DatabaseHelper {
	var dbs []DatabaseHelper
	if cronjob.DBName == "all" {
		if cronjob.DBType == "mysql" || cronjob.DBType == "mariadb" {
			mysqlItems, _ := mysqlRepo.List()
			for _, mysql := range mysqlItems {
				dbs = append(dbs, DatabaseHelper{
					ID:       mysql.ID,
					DBType:   cronjob.DBType,
					Database: mysql.MysqlName,
					Name:     mysql.Name,
				})
			}
		} else {
			pgItems, _ := postgresqlRepo.List()
			for _, pg := range pgItems {
				dbs = append(dbs, DatabaseHelper{
					ID:       pg.ID,
					DBType:   cronjob.DBType,
					Database: pg.PostgresqlName,
					Name:     pg.Name,
				})
			}
		}
		return dbs
	}
	itemID, _ := strconv.Atoi(cronjob.DBName)
	if cronjob.DBType == "mysql" || cronjob.DBType == "mariadb" {
		mysqlItem, _ := mysqlRepo.Get(repo.WithByID(uint(itemID)))
		dbs = append(dbs, DatabaseHelper{
			ID:       mysqlItem.ID,
			DBType:   cronjob.DBType,
			Database: mysqlItem.MysqlName,
			Name:     mysqlItem.Name,
		})
	} else {
		pgItem, _ := postgresqlRepo.Get(repo.WithByID(uint(itemID)))
		dbs = append(dbs, DatabaseHelper{
			ID:       pgItem.ID,
			DBType:   cronjob.DBType,
			Database: pgItem.PostgresqlName,
			Name:     pgItem.Name,
		})
	}
	return dbs
}

func loadWebsForJob(cronjob model.Cronjob) []model.Website {
	var weblist []model.Website
	if cronjob.Website == "all" {
		weblist, _ = websiteRepo.List()
		return weblist
	}
	itemID, _ := strconv.Atoi(cronjob.Website)
	webItem, _ := websiteRepo.GetFirst(repo.WithByID(uint(itemID)))
	if webItem.ID != 0 {
		weblist = append(weblist, webItem)
	}
	return weblist
}

func handleBackupLogs(taskItem *task.Task, targetDir, fileName string, secret string) error {
	fileOp := files.NewFileOp()
	websites, err := websiteRepo.List()
	if err != nil {
		return err
	}
	if len(websites) != 0 {
		webItem := GetOpenrestyDir(SitesRootDir)
		for _, website := range websites {
			taskItem.Logf("%s Website logs %s...", i18n.GetMsgByKey("TaskBackup"), website.Alias)
			dirItem := path.Join(targetDir, "website", website.Alias)
			if _, err := os.Stat(dirItem); err != nil && os.IsNotExist(err) {
				if err = os.MkdirAll(dirItem, os.ModePerm); err != nil {
					return err
				}
			}
			itemDir := path.Join(webItem, website.Alias, "log")
			logFiles, _ := os.ReadDir(itemDir)
			if len(logFiles) != 0 {
				for i := 0; i < len(logFiles); i++ {
					if !logFiles[i].IsDir() {
						_ = fileOp.CopyFile(path.Join(itemDir, logFiles[i].Name()), dirItem)
					}
				}
			}
			itemDir2 := path.Join(global.Dir.LocalBackupDir, "log/website", website.Alias)
			logFiles2, _ := os.ReadDir(itemDir2)
			if len(logFiles2) != 0 {
				for i := 0; i < len(logFiles2); i++ {
					if !logFiles2[i].IsDir() {
						_ = fileOp.CopyFile(path.Join(itemDir2, logFiles2[i].Name()), dirItem)
					}
				}
			}
		}
	}

	systemDir := path.Join(targetDir, "system")
	if _, err := os.Stat(systemDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(systemDir, os.ModePerm); err != nil {
			return err
		}
	}

	taskItem.Logf("%s System logs...", i18n.GetMsgByKey("TaskBackup"))
	systemLogFiles, _ := os.ReadDir(global.Dir.LogDir)
	if len(systemLogFiles) != 0 {
		for i := 0; i < len(systemLogFiles); i++ {
			if !systemLogFiles[i].IsDir() {
				_ = fileOp.CopyFile(path.Join(global.Dir.LogDir, systemLogFiles[i].Name()), systemDir)
			}
		}
	}

	taskItem.Logf("%s SSH logs...", i18n.GetMsgByKey("TaskBackup"))
	loginLogFiles, _ := os.ReadDir("/var/log")
	loginDir := path.Join(targetDir, "login")
	if _, err := os.Stat(loginDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(loginDir, os.ModePerm); err != nil {
			return err
		}
	}
	if len(loginLogFiles) != 0 {
		for i := 0; i < len(loginLogFiles); i++ {
			if !loginLogFiles[i].IsDir() && (strings.HasPrefix(loginLogFiles[i].Name(), "secure") || strings.HasPrefix(loginLogFiles[i].Name(), "auth.log")) {
				_ = fileOp.CopyFile(path.Join("/var/log", loginLogFiles[i].Name()), loginDir)
			}
		}
	}
	taskItem.Log("backup ssh log successful!")

	if err := fileOp.TarGzCompressPro(true, targetDir, path.Join(path.Dir(targetDir), fileName), secret, ""); err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(targetDir)
	}()
	return nil
}
