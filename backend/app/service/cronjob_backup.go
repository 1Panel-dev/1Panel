package service

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
)

func (u *CronjobService) handleApp(cronjob model.Cronjob, startTime time.Time) error {
	var apps []model.AppInstall
	if cronjob.AppID == "all" {
		apps, _ = appInstallRepo.ListBy()
	} else {
		itemID, _ := (strconv.Atoi(cronjob.AppID))
		app, err := appInstallRepo.GetFirst(commonRepo.WithByID(uint(itemID)))
		if err != nil {
			return err
		}
		apps = append(apps, app)
	}
	accountMap, err := u.loadClientMap(cronjob.TargetAccountIDs)
	if err != nil {
		return err
	}
	for _, app := range apps {
		var record model.BackupRecord
		record.From = "cronjob"
		record.Type = "app"
		record.CronjobID = cronjob.ID
		record.Name = app.App.Key
		record.DetailName = app.Name
		record.Source, record.BackupType = loadRecordPath(cronjob, accountMap)
		backupDir := path.Join(global.CONF.System.TmpDir, fmt.Sprintf("app/%s/%s", app.App.Key, app.Name))
		record.FileName = fmt.Sprintf("app_%s_%s.tar.gz", app.Name, startTime.Format("20060102150405"))
		if err := handleAppBackup(&app, backupDir, record.FileName); err != nil {
			return err
		}
		if err := backupRepo.CreateRecord(&record); err != nil {
			global.LOG.Errorf("save backup record failed, err: %v", err)
			return err
		}
		downloadPath, err := u.uploadCronjobBackFile(cronjob, accountMap, path.Join(backupDir, record.FileName))
		if err != nil {
			return err
		}
		record.FileDir = path.Dir(downloadPath)
		u.removeExpiredBackup(cronjob, accountMap, record)
	}
	return nil
}

func (u *CronjobService) handleWebsite(cronjob model.Cronjob, startTime time.Time) error {
	webs := loadWebsForJob(cronjob)
	accountMap, err := u.loadClientMap(cronjob.TargetAccountIDs)
	if err != nil {
		return err
	}
	for _, web := range webs {
		var record model.BackupRecord
		record.From = "cronjob"
		record.Type = "website"
		record.CronjobID = cronjob.ID
		record.Name = web.PrimaryDomain
		record.DetailName = web.Alias
		record.Source, record.BackupType = loadRecordPath(cronjob, accountMap)
		backupDir := path.Join(global.CONF.System.TmpDir, fmt.Sprintf("website/%s", web.PrimaryDomain))
		record.FileName = fmt.Sprintf("website_%s_%s.tar.gz", web.PrimaryDomain, startTime.Format("20060102150405"))
		if err := handleWebsiteBackup(&web, backupDir, record.FileName); err != nil {
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
	}
	return nil
}

func (u *CronjobService) handleDatabase(cronjob model.Cronjob, startTime time.Time) error {
	dbs := loadDbsForJob(cronjob)
	accountMap, err := u.loadClientMap(cronjob.TargetAccountIDs)
	if err != nil {
		return err
	}
	for _, dbInfo := range dbs {
		var record model.BackupRecord
		record.From = "cronjob"
		record.Type = dbInfo.DBType
		record.CronjobID = cronjob.ID
		record.Name = dbInfo.Database
		record.DetailName = dbInfo.Name
		record.Source, record.BackupType = loadRecordPath(cronjob, accountMap)

		backupDir := path.Join(global.CONF.System.TmpDir, fmt.Sprintf("database/%s/%s/%s", dbInfo.DBType, record.Name, dbInfo.Name))
		record.FileName = fmt.Sprintf("db_%s_%s.sql.gz", dbInfo.Name, startTime.Format("20060102150405"))
		if cronjob.DBType == "mysql" || cronjob.DBType == "mariadb" {
			if err := handleMysqlBackup(dbInfo.Database, dbInfo.Name, backupDir, record.FileName); err != nil {
				return err
			}
		} else {
			if err := handlePostgresqlBackup(dbInfo.Database, dbInfo.Name, backupDir, record.FileName); err != nil {
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
	}
	return nil
}

func (u *CronjobService) handleDirectory(cronjob model.Cronjob, startTime time.Time) error {
	accountMap, err := u.loadClientMap(cronjob.TargetAccountIDs)
	if err != nil {
		return err
	}
	fileName := fmt.Sprintf("directory%s_%s.tar.gz", strings.ReplaceAll(cronjob.SourceDir, "/", "_"), startTime.Format("20060102150405"))
	backupDir := path.Join(global.CONF.System.TmpDir, fmt.Sprintf("%s/%s", cronjob.Type, cronjob.Name))
	if err := handleTar(cronjob.SourceDir, backupDir, fileName, cronjob.ExclusionRules); err != nil {
		return err
	}
	var record model.BackupRecord
	record.From = "cronjob"
	record.Type = "directory"
	record.CronjobID = cronjob.ID
	record.Name = cronjob.Name
	record.Source, record.BackupType = loadRecordPath(cronjob, accountMap)
	downloadPath, err := u.uploadCronjobBackFile(cronjob, accountMap, path.Join(backupDir, fileName))
	if err != nil {
		return err
	}
	record.FileDir = path.Dir(downloadPath)
	record.FileName = fileName

	if err := backupRepo.CreateRecord(&record); err != nil {
		global.LOG.Errorf("save backup record failed, err: %v", err)
		return err
	}
	u.removeExpiredBackup(cronjob, accountMap, record)
	return nil
}

func (u *CronjobService) handleSystemLog(cronjob model.Cronjob, startTime time.Time) error {
	accountMap, err := u.loadClientMap(cronjob.TargetAccountIDs)
	if err != nil {
		return err
	}
	fileName := fmt.Sprintf("system_log_%s.tar.gz", startTime.Format("20060102150405"))
	backupDir := path.Join(global.CONF.System.TmpDir, "log", startTime.Format("20060102150405"))
	if err := handleBackupLogs(backupDir, fileName); err != nil {
		return err
	}
	var record model.BackupRecord
	record.From = "cronjob"
	record.Type = "log"
	record.CronjobID = cronjob.ID
	record.Name = cronjob.Name
	record.Source, record.BackupType = loadRecordPath(cronjob, accountMap)
	downloadPath, err := u.uploadCronjobBackFile(cronjob, accountMap, path.Join(path.Dir(backupDir), fileName))
	if err != nil {
		return err
	}
	record.FileDir = path.Dir(downloadPath)
	record.FileName = fileName

	if err := backupRepo.CreateRecord(&record); err != nil {
		global.LOG.Errorf("save backup record failed, err: %v", err)
		return err
	}
	u.removeExpiredBackup(cronjob, accountMap, record)
	return nil
}

func (u *CronjobService) handleSnapshot(cronjob model.Cronjob, startTime time.Time, logPath string) error {
	accountMap, err := u.loadClientMap(cronjob.TargetAccountIDs)
	if err != nil {
		return err
	}

	var record model.BackupRecord
	record.From = "cronjob"
	record.Type = "directory"
	record.CronjobID = cronjob.ID
	record.Name = cronjob.Name
	record.Source, record.BackupType = loadRecordPath(cronjob, accountMap)
	record.FileDir = "system_snapshot"

	req := dto.SnapshotCreate{
		From: record.BackupType,
	}
	name, err := NewISnapshotService().HandleSnapshot(true, logPath, req, startTime.Format("20060102150405"))
	if err != nil {
		return err
	}
	record.FileName = name + ".tar.gz"

	if err := backupRepo.CreateRecord(&record); err != nil {
		global.LOG.Errorf("save backup record failed, err: %v", err)
		return err
	}
	u.removeExpiredBackup(cronjob, accountMap, record)
	return nil
}

type databaseHelper struct {
	DBType   string
	Database string
	Name     string
}

func loadDbsForJob(cronjob model.Cronjob) []databaseHelper {
	var dbs []databaseHelper
	if cronjob.DBName == "all" {
		if cronjob.DBType == "mysql" || cronjob.DBType == "mariadb" {
			mysqlItems, _ := mysqlRepo.List()
			for _, mysql := range mysqlItems {
				dbs = append(dbs, databaseHelper{
					DBType:   cronjob.DBType,
					Database: mysql.MysqlName,
					Name:     mysql.Name,
				})
			}
		} else {
			pgItems, _ := postgresqlRepo.List()
			for _, pg := range pgItems {
				dbs = append(dbs, databaseHelper{
					DBType:   cronjob.DBType,
					Database: pg.PostgresqlName,
					Name:     pg.Name,
				})
			}
		}
		return dbs
	}
	itemID, _ := (strconv.Atoi(cronjob.DBName))
	if cronjob.DBType == "mysql" || cronjob.DBType == "mariadb" {
		mysqlItem, _ := mysqlRepo.Get(commonRepo.WithByID(uint(itemID)))
		dbs = append(dbs, databaseHelper{
			DBType:   cronjob.DBType,
			Database: mysqlItem.MysqlName,
			Name:     mysqlItem.Name,
		})
	} else {
		pgItem, _ := postgresqlRepo.Get(commonRepo.WithByID(uint(itemID)))
		dbs = append(dbs, databaseHelper{
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
	itemID, _ := (strconv.Atoi(cronjob.Website))
	webItem, _ := websiteRepo.GetFirst(commonRepo.WithByID(uint(itemID)))
	if webItem.ID != 0 {
		weblist = append(weblist, webItem)
	}
	return weblist
}

func loadRecordPath(cronjob model.Cronjob, accountMap map[string]cronjobUploadHelper) (string, string) {
	source := accountMap[fmt.Sprintf("%v", cronjob.TargetDirID)].backType
	targets := strings.Split(cronjob.TargetAccountIDs, ",")
	var itemAccounts []string
	for _, target := range targets {
		if len(target) == 0 {
			continue
		}
		if len(accountMap[target].backType) != 0 {
			itemAccounts = append(itemAccounts, accountMap[target].backType)
		}
	}
	backupType := strings.Join(itemAccounts, ",")
	return source, backupType
}

func handleBackupLogs(targetDir, fileName string) error {
	websites, err := websiteRepo.List()
	if err != nil {
		return err
	}
	if len(websites) != 0 {
		nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
		if err != nil {
			return err
		}
		webItem := path.Join(nginxInstall.GetPath(), "www/sites")
		for _, website := range websites {
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
						_ = cpBinary([]string{path.Join(itemDir, logFiles[i].Name())}, dirItem)
					}
				}
			}
			itemDir2 := path.Join(global.CONF.System.Backup, "log/website", website.Alias)
			logFiles2, _ := os.ReadDir(itemDir2)
			if len(logFiles2) != 0 {
				for i := 0; i < len(logFiles2); i++ {
					if !logFiles2[i].IsDir() {
						_ = cpBinary([]string{path.Join(itemDir2, logFiles2[i].Name())}, dirItem)
					}
				}
			}
		}
		global.LOG.Debug("backup website log successful!")
	}

	systemLogDir := path.Join(global.CONF.System.BaseDir, "1panel/log")
	systemDir := path.Join(targetDir, "system")
	if _, err := os.Stat(systemDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(systemDir, os.ModePerm); err != nil {
			return err
		}
	}
	systemLogFiles, _ := os.ReadDir(systemLogDir)
	if len(systemLogFiles) != 0 {
		for i := 0; i < len(systemLogFiles); i++ {
			if !systemLogFiles[i].IsDir() {
				_ = cpBinary([]string{path.Join(systemLogDir, systemLogFiles[i].Name())}, systemDir)
			}
		}
	}
	global.LOG.Debug("backup system log successful!")

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
				_ = cpBinary([]string{path.Join("/var/log", loginLogFiles[i].Name())}, loginDir)
			}
		}
	}
	global.LOG.Debug("backup ssh log successful!")

	if err := handleTar(targetDir, path.Dir(targetDir), fileName, ""); err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(targetDir)
	}()
	return nil
}
