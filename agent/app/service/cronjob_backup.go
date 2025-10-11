package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
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
	apps := loadAppsForJob(cronjob)
	if len(apps) == 0 {
		addSkipTask("App", taskItem)
		return nil
	}
	accountMap := NewBackupClientMap(strings.Split(cronjob.SourceAccountIDs, ","))
	if !accountMap[fmt.Sprintf("%d", cronjob.DownloadAccountID)].isOk {
		return errors.New(i18n.GetMsgWithDetail("LoadBackupFailed", accountMap[fmt.Sprintf("%d", cronjob.DownloadAccountID)].message))
	}
	for _, app := range apps {
		retry := 0
		taskItem.AddSubTaskWithOps(task.GetTaskName(app.Name, task.TaskBackup, task.TaskScopeCronjob), func(task *task.Task) error {
			var record model.BackupRecord
			record.Status = constant.StatusSuccess
			record.From = "cronjob"
			record.Type = "app"
			record.CronjobID = cronjob.ID
			record.Name = app.App.Key
			record.DetailName = app.Name
			record.DownloadAccountID, record.SourceAccountIDs = cronjob.DownloadAccountID, cronjob.SourceAccountIDs
			backupDir := path.Join(global.Dir.LocalBackupDir, fmt.Sprintf("tmp/app/%s/%s", app.App.Key, app.Name))
			record.FileName = simplifiedFileName(fmt.Sprintf("app_%s_%s.tar.gz", app.Name, startTime.Format(constant.DateTimeSlimLayout)+common.RandStrAndNum(5)))
			if err := doAppBackup(&app, task, backupDir, record.FileName, cronjob.ExclusionRules, cronjob.Secret); err != nil {
				if retry < int(cronjob.RetryTimes) || !cronjob.IgnoreErr {
					retry++
					return err
				} else {
					task.Log(i18n.GetMsgWithDetail("IgnoreBackupErr", err.Error()))
					cleanAccountMap(accountMap)
					return nil
				}
			}

			src := path.Join(backupDir, record.FileName)
			dst := strings.TrimPrefix(src, global.Dir.LocalBackupDir+"/tmp/")
			if err := uploadWithMap(*task, accountMap, src, dst, cronjob.SourceAccountIDs, cronjob.DownloadAccountID, cronjob.RetryTimes); err != nil {
				if retry < int(cronjob.RetryTimes) || !cronjob.IgnoreErr {
					retry++
					return err
				}
				task.Log(i18n.GetMsgWithDetail("IgnoreUploadErr", err.Error()))
				cleanAccountMap(accountMap)
				return nil
			}
			record.FileDir = path.Dir(dst)
			if err := backupRepo.CreateRecord(&record); err != nil {
				global.LOG.Errorf("save backup record failed, err: %v", err)
				return err
			}
			u.removeExpiredBackup(cronjob, accountMap, record)
			cleanAccountMap(accountMap)
			return nil
		}, nil, int(cronjob.RetryTimes), time.Duration(cronjob.Timeout)*time.Second)
	}
	return nil
}

func (u *CronjobService) handleWebsite(cronjob model.Cronjob, startTime time.Time, taskItem *task.Task) error {
	webs := loadWebsForJob(cronjob)
	if len(webs) == 0 {
		addSkipTask("Website", taskItem)
		return nil
	}
	accountMap := NewBackupClientMap(strings.Split(cronjob.SourceAccountIDs, ","))
	if !accountMap[fmt.Sprintf("%d", cronjob.DownloadAccountID)].isOk {
		return errors.New(i18n.GetMsgWithDetail("LoadBackupFailed", accountMap[fmt.Sprintf("%d", cronjob.DownloadAccountID)].message))
	}
	for _, web := range webs {
		retry := 0
		taskItem.AddSubTaskWithOps(task.GetTaskName(web.Alias, task.TaskBackup, task.TaskScopeCronjob), func(task *task.Task) error {
			var record model.BackupRecord
			record.Status = constant.StatusSuccess
			record.From = "cronjob"
			record.Type = "website"
			record.CronjobID = cronjob.ID
			record.Name = web.Alias
			record.DetailName = web.Alias
			record.DownloadAccountID, record.SourceAccountIDs = cronjob.DownloadAccountID, cronjob.SourceAccountIDs
			backupDir := path.Join(global.Dir.LocalBackupDir, fmt.Sprintf("tmp/website/%s", web.Alias))
			record.FileName = simplifiedFileName(fmt.Sprintf("website_%s_%s.tar.gz", web.Alias, startTime.Format(constant.DateTimeSlimLayout)+common.RandStrAndNum(5)))

			if err := doWebsiteBackup(&web, taskItem, backupDir, record.FileName, cronjob.ExclusionRules, cronjob.Secret); err != nil {
				if retry < int(cronjob.RetryTimes) || !cronjob.IgnoreErr {
					retry++
					return err
				} else {
					task.Log(i18n.GetMsgWithDetail("IgnoreBackupErr", err.Error()))
					cleanAccountMap(accountMap)
					return nil
				}
			}

			src := path.Join(backupDir, record.FileName)
			dst := strings.TrimPrefix(src, global.Dir.LocalBackupDir+"/tmp/")
			if err := uploadWithMap(*task, accountMap, src, dst, cronjob.SourceAccountIDs, cronjob.DownloadAccountID, cronjob.RetryTimes); err != nil {
				if retry < int(cronjob.RetryTimes) || !cronjob.IgnoreErr {
					retry++
					return err
				}
				task.Log(i18n.GetMsgWithDetail("IgnoreUploadErr", err.Error()))
				cleanAccountMap(accountMap)
				return nil
			}
			record.FileDir = path.Dir(dst)
			if err := backupRepo.CreateRecord(&record); err != nil {
				global.LOG.Errorf("save backup record failed, err: %v", err)
				return err
			}
			u.removeExpiredBackup(cronjob, accountMap, record)
			cleanAccountMap(accountMap)
			return nil
		}, nil, int(cronjob.RetryTimes), time.Duration(cronjob.Timeout)*time.Second)
	}
	return nil
}

func (u *CronjobService) handleDatabase(cronjob model.Cronjob, startTime time.Time, taskItem *task.Task) error {
	dbs := loadDbsForJob(cronjob)
	if len(dbs) == 0 {
		addSkipTask("Database", taskItem)
		return nil
	}
	accountMap := NewBackupClientMap(strings.Split(cronjob.SourceAccountIDs, ","))
	if !accountMap[fmt.Sprintf("%d", cronjob.DownloadAccountID)].isOk {
		return errors.New(i18n.GetMsgWithDetail("LoadBackupFailed", accountMap[fmt.Sprintf("%d", cronjob.DownloadAccountID)].message))
	}
	for _, dbInfo := range dbs {
		retry := 0
		itemName := fmt.Sprintf("%s[%s] - %s", dbInfo.Database, dbInfo.DBType, dbInfo.Name)
		taskItem.AddSubTaskWithOps(task.GetTaskName(itemName, task.TaskBackup, task.TaskScopeCronjob), func(task *task.Task) error {
			var record model.BackupRecord
			record.Status = constant.StatusSuccess
			record.From = "cronjob"
			record.Type = dbInfo.DBType
			record.CronjobID = cronjob.ID
			record.Name = dbInfo.Database
			record.DetailName = dbInfo.Name
			record.DownloadAccountID, record.SourceAccountIDs = cronjob.DownloadAccountID, cronjob.SourceAccountIDs

			backupDir := path.Join(global.Dir.LocalBackupDir, fmt.Sprintf("tmp/database/%s/%s/%s", dbInfo.DBType, record.Name, dbInfo.Name))
			record.FileName = simplifiedFileName(fmt.Sprintf("db_%s_%s.sql.gz", dbInfo.Name, startTime.Format(constant.DateTimeSlimLayout)+common.RandStrAndNum(5)))
			if cronjob.DBType == "mysql" || cronjob.DBType == "mariadb" || cronjob.DBType == "mysql-cluster" {
				if err := doMysqlBackup(dbInfo, backupDir, record.FileName); err != nil {
					if retry < int(cronjob.RetryTimes) || !cronjob.IgnoreErr {
						retry++
						return err
					} else {
						task.Log(i18n.GetMsgWithDetail("IgnoreBackupErr", err.Error()))
						cleanAccountMap(accountMap)
						return nil
					}
				}
			} else {
				if err := doPostgresqlgBackup(dbInfo, backupDir, record.FileName); err != nil {
					if retry < int(cronjob.RetryTimes) || !cronjob.IgnoreErr {
						retry++
						return err
					} else {
						task.Log(i18n.GetMsgWithDetail("IgnoreBackupErr", err.Error()))
						cleanAccountMap(accountMap)
						return nil
					}
				}
			}

			src := path.Join(backupDir, record.FileName)
			dst := strings.TrimPrefix(src, global.Dir.LocalBackupDir+"/tmp/")
			if err := uploadWithMap(*task, accountMap, src, dst, cronjob.SourceAccountIDs, cronjob.DownloadAccountID, cronjob.RetryTimes); err != nil {
				if retry < int(cronjob.RetryTimes) || !cronjob.IgnoreErr {
					retry++
					return err
				}
				task.Log(i18n.GetMsgWithDetail("IgnoreUploadErr", err.Error()))
				cleanAccountMap(accountMap)
				return nil
			}
			record.FileDir = path.Dir(dst)
			if err := backupRepo.CreateRecord(&record); err != nil {
				global.LOG.Errorf("save backup record failed, err: %v", err)
				return err
			}
			u.removeExpiredBackup(cronjob, accountMap, record)
			cleanAccountMap(accountMap)
			return nil
		}, nil, int(cronjob.RetryTimes), time.Duration(cronjob.Timeout)*time.Second)
	}
	return nil
}

func (u *CronjobService) handleDirectory(cronjob model.Cronjob, startTime time.Time, taskItem *task.Task) error {
	accountMap := NewBackupClientMap(strings.Split(cronjob.SourceAccountIDs, ","))
	if !accountMap[fmt.Sprintf("%d", cronjob.DownloadAccountID)].isOk {
		return errors.New(i18n.GetMsgWithDetail("LoadBackupFailed", accountMap[fmt.Sprintf("%d", cronjob.DownloadAccountID)].message))
	}
	taskItem.AddSubTaskWithOps(task.GetTaskName(cronjob.SourceDir, task.TaskBackup, task.TaskScopeCronjob), func(task *task.Task) error {
		fileName := fmt.Sprintf("%s.tar.gz", startTime.Format(constant.DateTimeSlimLayout)+common.RandStrAndNum(2))
		if cronjob.IsDir || len(strings.Split(cronjob.SourceDir, ",")) == 1 {
			fileName = loadFileName(cronjob.SourceDir)
		}
		fileName = simplifiedFileName(fileName)
		backupDir := path.Join(global.Dir.LocalBackupDir, fmt.Sprintf("tmp/%s/%s", cronjob.Type, cronjob.Name))

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
		record.Status = constant.StatusSuccess
		record.From = "cronjob"
		record.Type = "directory"
		record.CronjobID = cronjob.ID
		record.Name = cronjob.Name
		record.DownloadAccountID, record.SourceAccountIDs = cronjob.DownloadAccountID, cronjob.SourceAccountIDs

		src := path.Join(backupDir, fileName)
		dst := strings.TrimPrefix(src, global.Dir.LocalBackupDir+"/tmp/")
		if err := uploadWithMap(*task, accountMap, src, dst, cronjob.SourceAccountIDs, cronjob.DownloadAccountID, cronjob.RetryTimes); err != nil {
			return err
		}
		record.FileDir = path.Dir(dst)
		record.FileName = fileName
		if err := backupRepo.CreateRecord(&record); err != nil {
			return err
		}
		u.removeExpiredBackup(cronjob, accountMap, record)
		return nil
	}, nil, int(cronjob.RetryTimes), time.Duration(cronjob.Timeout)*time.Second)
	return nil
}

func (u *CronjobService) handleSystemLog(cronjob model.Cronjob, startTime time.Time, taskItem *task.Task) error {
	accountMap := NewBackupClientMap(strings.Split(cronjob.SourceAccountIDs, ","))
	if !accountMap[fmt.Sprintf("%d", cronjob.DownloadAccountID)].isOk {
		return errors.New(i18n.GetMsgWithDetail("LoadBackupFailed", accountMap[fmt.Sprintf("%d", cronjob.DownloadAccountID)].message))
	}
	taskItem.AddSubTaskWithOps(task.GetTaskName(i18n.GetMsgByKey("SystemLog"), task.TaskBackup, task.TaskScopeCronjob), func(task *task.Task) error {
		nameItem := startTime.Format(constant.DateTimeSlimLayout) + common.RandStrAndNum(5)
		fileName := fmt.Sprintf("system_log_%s.tar.gz", nameItem)
		backupDir := path.Join(global.Dir.LocalBackupDir, "tmp/log", nameItem)
		if err := handleBackupLogs(taskItem, backupDir, fileName, cronjob.Secret); err != nil {
			return err
		}
		var record model.BackupRecord
		record.Status = constant.StatusSuccess
		record.From = "cronjob"
		record.Type = "log"
		record.CronjobID = cronjob.ID
		record.Name = cronjob.Name
		record.DownloadAccountID, record.SourceAccountIDs = cronjob.DownloadAccountID, cronjob.SourceAccountIDs

		src := path.Join(path.Dir(backupDir), fileName)
		dst := strings.TrimPrefix(src, global.Dir.LocalBackupDir+"/tmp/")
		if err := uploadWithMap(*task, accountMap, src, dst, cronjob.SourceAccountIDs, cronjob.DownloadAccountID, cronjob.RetryTimes); err != nil {
			return err
		}
		record.FileDir = path.Dir(dst)
		record.FileName = fileName
		if err := backupRepo.CreateRecord(&record); err != nil {
			return err
		}
		u.removeExpiredBackup(cronjob, accountMap, record)
		return nil
	}, nil, int(cronjob.RetryTimes), time.Duration(cronjob.Timeout)*time.Second)
	return nil
}

func (u *CronjobService) handleSnapshot(cronjob model.Cronjob, jobRecord model.JobRecords, taskItem *task.Task) error {
	accountMap := NewBackupClientMap(strings.Split(cronjob.SourceAccountIDs, ","))
	if !accountMap[fmt.Sprintf("%d", cronjob.DownloadAccountID)].isOk {
		return errors.New(i18n.GetMsgWithDetail("LoadBackupFailed", accountMap[fmt.Sprintf("%d", cronjob.DownloadAccountID)].message))
	}
	var record model.BackupRecord
	record.Status = constant.StatusSuccess
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

	itemData, err := loadSnapWithRule(cronjob)
	if err != nil {
		return err
	}
	req := dto.SnapshotCreate{
		Name:    fmt.Sprintf("snapshot-1panel-%s-%s-linux-%s-%s", scope, versionItem.Value, loadOs(), jobRecord.StartTime.Format(constant.DateTimeSlimLayout)+common.RandStrAndNum(5)),
		Secret:  cronjob.Secret,
		TaskID:  jobRecord.TaskID,
		Timeout: cronjob.Timeout,

		SourceAccountIDs:  record.SourceAccountIDs,
		DownloadAccountID: cronjob.DownloadAccountID,
		AppData:           itemData.AppData,
		PanelData:         itemData.PanelData,
		BackupData:        itemData.BackupData,
		WithDockerConf:    true,
		WithMonitorData:   true,
		WithLoginLog:      true,
		WithOperationLog:  true,
		WithSystemLog:     true,
		WithTaskLog:       true,
		IgnoreFiles:       strings.Split(cronjob.ExclusionRules, ","),
	}

	if err := NewISnapshotService().SnapshotCreate(taskItem, req, jobRecord.ID, cronjob.RetryTimes); err != nil {
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

func loadAppsForJob(cronjob model.Cronjob) []model.AppInstall {
	var apps []model.AppInstall
	if cronjob.AppID == "all" {
		apps, _ = appInstallRepo.ListBy(context.Background())
	} else {
		appIds := strings.Split(cronjob.AppID, ",")
		var idItems []uint
		for i := 0; i < len(appIds); i++ {
			itemID, _ := strconv.Atoi(appIds[i])
			idItems = append(idItems, uint(itemID))
		}
		appItems, _ := appInstallRepo.ListBy(context.Background(), repo.WithByIDs(idItems))
		apps = appItems
	}
	return apps
}

type DatabaseHelper struct {
	ID       uint
	DBType   string
	Database string
	Name     string
}

func addSkipTask(source string, taskItem *task.Task) {
	taskItem.AddSubTask(task.GetTaskName(i18n.GetMsgByKey(source), task.TaskBackup, task.TaskScopeCronjob), func(task *task.Task) error {
		taskItem.Log(i18n.GetMsgByKey("NoSuchResource"))
		return nil
	}, nil)
}

func loadDbsForJob(cronjob model.Cronjob) []DatabaseHelper {
	var dbs []DatabaseHelper
	if cronjob.DBName == "all" {
		if cronjob.DBType == "mysql" || cronjob.DBType == "mariadb" || cronjob.DBType == "mysql-cluster" {
			databaseService := NewIDatabaseService()
			mysqlItems, _ := databaseService.LoadItems(cronjob.DBType)
			for _, mysql := range mysqlItems {
				dbs = append(dbs, DatabaseHelper{
					ID:       mysql.ID,
					DBType:   cronjob.DBType,
					Database: mysql.Database,
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
	dbNames := strings.Split(cronjob.DBName, ",")
	for _, name := range dbNames {
		itemID, _ := strconv.Atoi(name)
		if cronjob.DBType == "mysql" || cronjob.DBType == "mariadb" || cronjob.DBType == "mysql-cluster" {
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
	}
	return dbs
}

func loadWebsForJob(cronjob model.Cronjob) []model.Website {
	var weblist []model.Website
	if cronjob.Website == "all" {
		weblist, _ = websiteRepo.List()
		return weblist
	}
	websites := strings.Split(cronjob.Website, ",")
	var idItems []uint
	for i := 0; i < len(websites); i++ {
		itemID, _ := strconv.Atoi(websites[i])
		idItems = append(idItems, uint(itemID))
	}
	weblist, _ = websiteRepo.GetBy(repo.WithByIDs(idItems))
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
			itemDir2 := path.Join(global.Dir.LocalBackupDir, "tmp/log/website", website.Alias)
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
	taskItem.Logf("%s Website logs...", i18n.GetMsgByKey("TaskBackup"))

	systemDir := path.Join(targetDir, "system")
	if _, err := os.Stat(systemDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(systemDir, os.ModePerm); err != nil {
			return err
		}
	}

	systemLogFiles, _ := os.ReadDir(global.Dir.LogDir)
	if len(systemLogFiles) != 0 {
		for i := 0; i < len(systemLogFiles); i++ {
			if !systemLogFiles[i].IsDir() {
				_ = fileOp.CopyFile(path.Join(global.Dir.LogDir, systemLogFiles[i].Name()), systemDir)
			}
		}
	}
	taskItem.Logf("%s System logs...", i18n.GetMsgByKey("TaskBackup"))

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
	taskItem.Logf("%s SSH logs...", i18n.GetMsgByKey("TaskBackup"))

	if err := fileOp.TarGzCompressPro(true, targetDir, path.Join(path.Dir(targetDir), fileName), secret, ""); err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(targetDir)
	}()
	return nil
}

func loadSnapWithRule(cronjob model.Cronjob) (dto.SnapshotData, error) {
	itemData, err := NewISnapshotService().LoadSnapshotData()
	if err != nil {
		return itemData, err
	}

	if len(cronjob.SnapshotRule) == 0 {
		return itemData, nil
	}
	var snapRule dto.SnapshotRule
	if err := json.Unmarshal([]byte(cronjob.SnapshotRule), &snapRule); err != nil {
		return itemData, err
	}
	if len(snapRule.IgnoreAppIDs) == 0 && !snapRule.WithImage {
		return itemData, nil
	}

	var ignoreApps []model.AppInstall
	if len(snapRule.IgnoreAppIDs) != 0 {
		ignoreApps, _ = appInstallRepo.ListBy(context.Background(), repo.WithByIDs(snapRule.IgnoreAppIDs))
	}
	if len(ignoreApps) == 0 && !snapRule.WithImage {
		return itemData, nil
	}
	for i := 0; i < len(itemData.AppData); i++ {
		isIgnore := false
		for _, ignore := range ignoreApps {
			if ignore.App.Key == itemData.AppData[i].Key && ignore.Name == itemData.AppData[i].Name {
				isIgnore = true
				itemData.AppData[i].IsCheck = false
				for j := 0; j < len(itemData.AppData[i].Children); j++ {
					if itemData.AppData[i].Children[j].Label == "appData" {
						itemData.AppData[i].Children[j].IsCheck = false
					}
				}
				break
			}
		}
		if snapRule.WithImage && !isIgnore {
			for j := 0; j < len(itemData.AppData[i].Children); j++ {
				if itemData.AppData[i].Children[j].Label == "appImage" {
					itemData.AppData[i].Children[j].IsCheck = true
				}
			}
		}
	}
	return itemData, nil
}

func loadFileName(src string) string {
	dirs := strings.Split(filepath.ToSlash(src), "/")
	var keyPart string
	if len(dirs) >= 3 {
		keyPart = filepath.Join(dirs[len(dirs)-3], dirs[len(dirs)-2], dirs[len(dirs)-1])
	}
	cleanName := strings.ReplaceAll(keyPart, string(filepath.Separator), "_")
	timestamp := time.Now().Format(constant.DateTimeSlimLayout)
	return fmt.Sprintf("%s_%s_%s.tar.gz", cleanName, timestamp, common.RandStrAndNum(2))
}

func simplifiedFileName(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")
	return name
}

func cleanAccountMap(accountMap map[string]backupClientHelper) {
	for key, val := range accountMap {
		val.hasBackup = false
		accountMap[key] = val
	}
}
