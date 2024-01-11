package service

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/i18n"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cloud_storage"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/1Panel-dev/1Panel/backend/utils/ntp"
	"github.com/pkg/errors"
)

func (u *CronjobService) HandleJob(cronjob *model.Cronjob) {
	var (
		message []byte
		err     error
	)
	record := cronjobRepo.StartRecords(cronjob.ID, cronjob.KeepLocal, "")
	go func() {
		switch cronjob.Type {
		case "shell":
			if len(cronjob.Script) == 0 {
				return
			}
			if len(cronjob.ContainerName) != 0 {
				message, err = u.handleShell(cronjob.Type, cronjob.Name, fmt.Sprintf("docker exec %s %s", cronjob.ContainerName, cronjob.Script))
			} else {
				message, err = u.handleShell(cronjob.Type, cronjob.Name, cronjob.Script)
			}
			u.HandleRmExpired("LOCAL", "", "", cronjob, nil)
		case "snapshot":
			messageItem := ""
			messageItem, record.File, err = u.handleSnapshot(cronjob, record.StartTime)
			message = []byte(messageItem)
		case "curl":
			if len(cronjob.URL) == 0 {
				return
			}
			message, err = u.handleShell(cronjob.Type, cronjob.Name, fmt.Sprintf("curl '%s'", cronjob.URL))
			u.HandleRmExpired("LOCAL", "", "", cronjob, nil)
		case "ntp":
			err = u.handleNtpSync()
			u.HandleRmExpired("LOCAL", "", "", cronjob, nil)
		case "website", "database", "app":
			record.File, err = u.handleBackup(cronjob, record.StartTime)
		case "directory":
			if len(cronjob.SourceDir) == 0 {
				return
			}
			record.File, err = u.handleBackup(cronjob, record.StartTime)
		case "cutWebsiteLog":
			var messageItem []string
			messageItem, record.File, err = u.handleCutWebsiteLog(cronjob, record.StartTime)
			message = []byte(strings.Join(messageItem, "\n"))
		case "clean":
			messageItem := ""
			messageItem, err = u.handleSystemClean()
			message = []byte(messageItem)
			u.HandleRmExpired("LOCAL", "", "", cronjob, nil)
		case "log":
			record.File, err = u.handleSystemLog(*cronjob, record.StartTime)
		}

		if err != nil {
			cronjobRepo.EndRecords(record, constant.StatusFailed, err.Error(), string(message))
			return
		}
		if len(message) != 0 {
			record.Records, err = mkdirAndWriteFile(cronjob, record.StartTime, message)
			if err != nil {
				global.LOG.Errorf("save file %s failed, err: %v", record.Records, err)
			}
		}
		cronjobRepo.EndRecords(record, constant.StatusSuccess, "", record.Records)
	}()
}

func (u *CronjobService) handleShell(cronType, cornName, script string) ([]byte, error) {
	handleDir := fmt.Sprintf("%s/task/%s/%s", constant.DataDir, cronType, cornName)
	if _, err := os.Stat(handleDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(handleDir, os.ModePerm); err != nil {
			return nil, err
		}
	}
	stdout, err := cmd.ExecCronjobWithTimeOut(script, handleDir, 24*time.Hour)
	if err != nil {
		return []byte(stdout), err
	}
	return []byte(stdout), nil
}

func (u *CronjobService) handleNtpSync() error {
	ntpServer, err := settingRepo.Get(settingRepo.WithByKey("NtpSite"))
	if err != nil {
		return err
	}
	ntime, err := ntp.GetRemoteTime(ntpServer.Value)
	if err != nil {
		return err
	}
	if err := ntp.UpdateSystemTime(ntime.Format("2006-01-02 15:04:05")); err != nil {
		return err
	}
	return nil
}

func (u *CronjobService) handleBackup(cronjob *model.Cronjob, startTime time.Time) (string, error) {
	backup, err := backupRepo.Get(commonRepo.WithByID(uint(cronjob.TargetDirID)))
	if err != nil {
		return "", err
	}
	localDir, err := loadLocalDir()
	if err != nil {
		return "", err
	}
	global.LOG.Infof("start to backup %s %s to %s", cronjob.Type, cronjob.Name, backup.Type)

	switch cronjob.Type {
	case "database":
		paths, err := u.handleDatabase(*cronjob, backup, startTime)
		return strings.Join(paths, ","), err
	case "app":
		paths, err := u.handleApp(*cronjob, backup, startTime)
		return strings.Join(paths, ","), err
	case "website":
		paths, err := u.handleWebsite(*cronjob, backup, startTime)
		return strings.Join(paths, ","), err
	default:
		fileName := fmt.Sprintf("directory%s_%s.tar.gz", strings.ReplaceAll(cronjob.SourceDir, "/", "_"), startTime.Format("20060102150405"))
		backupDir := path.Join(localDir, fmt.Sprintf("%s/%s", cronjob.Type, cronjob.Name))
		itemFileDir := fmt.Sprintf("%s/%s", cronjob.Type, cronjob.Name)
		global.LOG.Infof("handle tar %s to %s", backupDir, fileName)
		if err := handleTar(cronjob.SourceDir, backupDir, fileName, cronjob.ExclusionRules); err != nil {
			return "", err
		}
		var client cloud_storage.CloudStorageClient
		if backup.Type != "LOCAL" {
			if !cronjob.KeepLocal {
				defer func() {
					_ = os.RemoveAll(fmt.Sprintf("%s/%s", backupDir, fileName))
				}()
			}
			client, err = NewIBackupService().NewClient(&backup)
			if err != nil {
				return "", err
			}
			if len(backup.BackupPath) != 0 {
				itemFileDir = path.Join(strings.TrimPrefix(backup.BackupPath, "/"), itemFileDir)
			}
			if _, err = client.Upload(backupDir+"/"+fileName, itemFileDir+"/"+fileName); err != nil {
				return "", err
			}
		}
		u.HandleRmExpired(backup.Type, backup.BackupPath, localDir, cronjob, client)
		if backup.Type == "LOCAL" || cronjob.KeepLocal {
			return fmt.Sprintf("%s/%s", backupDir, fileName), nil
		} else {
			return fmt.Sprintf("%s/%s", itemFileDir, fileName), nil
		}
	}
}

func (u *CronjobService) HandleRmExpired(backType, backupPath, localDir string, cronjob *model.Cronjob, backClient cloud_storage.CloudStorageClient) {
	global.LOG.Infof("start to handle remove expired, retain copies: %d", cronjob.RetainCopies)
	records, _ := cronjobRepo.ListRecord(cronjobRepo.WithByJobID(int(cronjob.ID)), commonRepo.WithOrderBy("created_at desc"))
	if len(records) <= int(cronjob.RetainCopies) {
		return
	}
	for i := int(cronjob.RetainCopies); i < len(records); i++ {
		if len(records[i].File) != 0 {
			files := strings.Split(records[i].File, ",")
			for _, file := range files {
				_ = os.Remove(file)
				_ = backupRepo.DeleteRecord(context.TODO(), backupRepo.WithByFileName(path.Base(file)))
				if backType == "LOCAL" {
					continue
				}

				fileItem := file
				if cronjob.KeepLocal {
					if len(backupPath) != 0 {
						fileItem = path.Join(strings.TrimPrefix(backupPath, "/") + strings.TrimPrefix(file, localDir+"/"))
					} else {
						fileItem = strings.TrimPrefix(file, localDir+"/")
					}
				}

				if cronjob.Type == "snapshot" {
					_ = snapshotRepo.Delete(commonRepo.WithByName(strings.TrimSuffix(path.Base(fileItem), ".tar.gz")))
				}
				_, _ = backClient.Delete(fileItem)
			}
		}
		_ = cronjobRepo.DeleteRecord(commonRepo.WithByID(uint(records[i].ID)))
		_ = os.Remove(records[i].Records)
	}
}

func handleTar(sourceDir, targetDir, name, exclusionRules string) error {
	if _, err := os.Stat(targetDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return err
		}
	}

	excludes := strings.Split(exclusionRules, ",")
	excludeRules := ""
	excludes = append(excludes, "*.sock")
	for _, exclude := range excludes {
		if len(exclude) == 0 {
			continue
		}
		excludeRules += " --exclude " + exclude
	}
	path := ""
	if strings.Contains(sourceDir, "/") {
		itemDir := strings.ReplaceAll(sourceDir[strings.LastIndex(sourceDir, "/"):], "/", "")
		aheadDir := sourceDir[:strings.LastIndex(sourceDir, "/")]
		if len(aheadDir) == 0 {
			aheadDir = "/"
		}
		path += fmt.Sprintf("-C %s %s", aheadDir, itemDir)
	} else {
		path = sourceDir
	}

	commands := fmt.Sprintf("tar --warning=no-file-changed --ignore-failed-read -zcf %s %s %s", targetDir+"/"+name, excludeRules, path)
	global.LOG.Debug(commands)
	stdout, err := cmd.ExecWithTimeOut(commands, 24*time.Hour)
	if err != nil {
		if len(stdout) != 0 {
			global.LOG.Errorf("do handle tar failed, stdout: %s, err: %v", stdout, err)
			return fmt.Errorf("do handle tar failed, stdout: %s, err: %v", stdout, err)
		}
	}
	return nil
}

func handleUnTar(sourceFile, targetDir string) error {
	if _, err := os.Stat(targetDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return err
		}
	}

	commands := fmt.Sprintf("tar zxvfC %s %s", sourceFile, targetDir)
	global.LOG.Debug(commands)
	stdout, err := cmd.ExecWithTimeOut(commands, 24*time.Hour)
	if err != nil {
		global.LOG.Errorf("do handle untar failed, stdout: %s, err: %v", stdout, err)
		return errors.New(stdout)
	}
	return nil
}

func (u *CronjobService) handleDatabase(cronjob model.Cronjob, backup model.BackupAccount, startTime time.Time) ([]string, error) {
	var paths []string
	localDir, err := loadLocalDir()
	if err != nil {
		return paths, err
	}

	dbs := loadDbsForJob(cronjob)

	var client cloud_storage.CloudStorageClient
	if backup.Type != "LOCAL" {
		client, err = NewIBackupService().NewClient(&backup)
		if err != nil {
			return paths, err
		}
	}

	for _, dbInfo := range dbs {
		var record model.BackupRecord
		record.Type = dbInfo.DBType
		record.Source = "LOCAL"
		record.BackupType = backup.Type

		record.Name = dbInfo.Database
		backupDir := path.Join(localDir, fmt.Sprintf("database/%s/%s/%s", dbInfo.DBType, record.Name, dbInfo.Name))
		record.FileName = fmt.Sprintf("db_%s_%s.sql.gz", dbInfo.Name, startTime.Format("20060102150405"))
		if cronjob.DBType == "mysql" || cronjob.DBType == "mariadb" {
			if err = handleMysqlBackup(dbInfo.Database, dbInfo.Name, backupDir, record.FileName); err != nil {
				return paths, err
			}
		} else {
			if err = handlePostgresqlBackup(dbInfo.Database, dbInfo.Name, backupDir, record.FileName); err != nil {
				return paths, err
			}
		}

		record.DetailName = dbInfo.Name
		record.FileDir = backupDir
		itemFileDir := strings.TrimPrefix(backupDir, localDir+"/")
		if !cronjob.KeepLocal && backup.Type != "LOCAL" {
			record.Source = backup.Type
			record.FileDir = itemFileDir
		}

		if err := backupRepo.CreateRecord(&record); err != nil {
			global.LOG.Errorf("save backup record failed, err: %v", err)
			return paths, err
		}
		if backup.Type != "LOCAL" {
			if !cronjob.KeepLocal {
				defer func() {
					_ = os.RemoveAll(fmt.Sprintf("%s/%s", backupDir, record.FileName))
				}()
			}
			if len(backup.BackupPath) != 0 {
				itemFileDir = path.Join(strings.TrimPrefix(backup.BackupPath, "/"), itemFileDir)
			}
			if _, err = client.Upload(backupDir+"/"+record.FileName, itemFileDir+"/"+record.FileName); err != nil {
				return paths, err
			}
		}
		if backup.Type == "LOCAL" || cronjob.KeepLocal {
			paths = append(paths, fmt.Sprintf("%s/%s", record.FileDir, record.FileName))
		} else {
			paths = append(paths, fmt.Sprintf("%s/%s", itemFileDir, record.FileName))
		}
	}
	u.HandleRmExpired(backup.Type, backup.BackupPath, localDir, &cronjob, client)
	return paths, nil
}

func (u *CronjobService) handleCutWebsiteLog(cronjob *model.Cronjob, startTime time.Time) ([]string, string, error) {
	var (
		err       error
		filePaths []string
		msgs      []string
	)
	websites := loadWebsForJob(*cronjob)
	nginx, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return msgs, "", nil
	}
	baseDir := path.Join(nginx.GetPath(), "www", "sites")
	fileOp := files.NewFileOp()
	var wg sync.WaitGroup
	wg.Add(len(websites))
	for _, website := range websites {
		websiteLogDir := path.Join(baseDir, website.Alias, "log")
		srcAccessLogPath := path.Join(websiteLogDir, "access.log")
		srcErrorLogPath := path.Join(websiteLogDir, "error.log")
		dstLogDir := path.Join(global.CONF.System.Backup, "log", "website", website.Alias)
		if !fileOp.Stat(dstLogDir) {
			_ = os.MkdirAll(dstLogDir, 0755)
		}

		dstName := fmt.Sprintf("%s_log_%s.gz", website.PrimaryDomain, startTime.Format("20060102150405"))
		dstFilePath := path.Join(dstLogDir, dstName)
		filePaths = append(filePaths, dstFilePath)

		if err = backupLogFile(dstFilePath, websiteLogDir, fileOp); err != nil {
			websiteErr := buserr.WithNameAndErr("ErrCutWebsiteLog", website.PrimaryDomain, err)
			err = websiteErr
			msgs = append(msgs, websiteErr.Error())
			global.LOG.Error(websiteErr.Error())
			continue
		} else {
			_ = fileOp.WriteFile(srcAccessLogPath, strings.NewReader(""), 0755)
			_ = fileOp.WriteFile(srcErrorLogPath, strings.NewReader(""), 0755)
		}
		msg := i18n.GetMsgWithMap("CutWebsiteLogSuccess", map[string]interface{}{"name": website.PrimaryDomain, "path": dstFilePath})
		global.LOG.Infof(msg)
		msgs = append(msgs, msg)
	}
	wg.Wait()
	u.HandleRmExpired("LOCAL", "", "", cronjob, nil)
	return msgs, strings.Join(filePaths, ","), err
}

func backupLogFile(dstFilePath, websiteLogDir string, fileOp files.FileOp) error {
	if err := cmd.ExecCmd(fmt.Sprintf("tar -czf %s -C %s %s", dstFilePath, websiteLogDir, strings.Join([]string{"access.log", "error.log"}, " "))); err != nil {
		dstDir := path.Dir(dstFilePath)
		if err = fileOp.Copy(path.Join(websiteLogDir, "access.log"), dstDir); err != nil {
			return err
		}
		if err = fileOp.Copy(path.Join(websiteLogDir, "error.log"), dstDir); err != nil {
			return err
		}
		if err = cmd.ExecCmd(fmt.Sprintf("tar -czf %s -C %s %s", dstFilePath, dstDir, strings.Join([]string{"access.log", "error.log"}, " "))); err != nil {
			return err
		}
		_ = fileOp.DeleteFile(path.Join(dstDir, "access.log"))
		_ = fileOp.DeleteFile(path.Join(dstDir, "error.log"))
		return nil
	}
	return nil
}

func (u *CronjobService) handleApp(cronjob model.Cronjob, backup model.BackupAccount, startTime time.Time) ([]string, error) {
	var paths []string
	localDir, err := loadLocalDir()
	if err != nil {
		return paths, err
	}

	var applist []model.AppInstall
	if cronjob.AppID == "all" {
		applist, err = appInstallRepo.ListBy()
		if err != nil {
			return paths, err
		}
	} else {
		itemID, _ := (strconv.Atoi(cronjob.AppID))
		app, err := appInstallRepo.GetFirst(commonRepo.WithByID(uint(itemID)))
		if err != nil {
			return paths, err
		}
		applist = append(applist, app)
	}

	var client cloud_storage.CloudStorageClient
	if backup.Type != "LOCAL" {
		client, err = NewIBackupService().NewClient(&backup)
		if err != nil {
			return paths, err
		}
	}

	for _, app := range applist {
		var record model.BackupRecord
		record.Type = "app"
		record.Name = app.App.Key
		record.DetailName = app.Name
		record.Source = "LOCAL"
		record.BackupType = backup.Type
		backupDir := path.Join(localDir, fmt.Sprintf("app/%s/%s", app.App.Key, app.Name))
		record.FileDir = backupDir
		itemFileDir := strings.TrimPrefix(backupDir, localDir+"/")
		if !cronjob.KeepLocal && backup.Type != "LOCAL" {
			record.Source = backup.Type
			record.FileDir = strings.TrimPrefix(backupDir, localDir+"/")
		}
		record.FileName = fmt.Sprintf("app_%s_%s.tar.gz", app.Name, startTime.Format("20060102150405"))
		if err := handleAppBackup(&app, backupDir, record.FileName); err != nil {
			return paths, err
		}
		if err := backupRepo.CreateRecord(&record); err != nil {
			global.LOG.Errorf("save backup record failed, err: %v", err)
			return paths, err
		}
		if backup.Type != "LOCAL" {
			if !cronjob.KeepLocal {
				defer func() {
					_ = os.RemoveAll(fmt.Sprintf("%s/%s", backupDir, record.FileName))
				}()
			}
			if len(backup.BackupPath) != 0 {
				itemFileDir = path.Join(strings.TrimPrefix(backup.BackupPath, "/"), itemFileDir)
			}
			if _, err = client.Upload(backupDir+"/"+record.FileName, itemFileDir+"/"+record.FileName); err != nil {
				return paths, err
			}
		}
		if backup.Type == "LOCAL" || cronjob.KeepLocal {
			paths = append(paths, fmt.Sprintf("%s/%s", record.FileDir, record.FileName))
		} else {
			paths = append(paths, fmt.Sprintf("%s/%s", itemFileDir, record.FileName))
		}
	}
	u.HandleRmExpired(backup.Type, backup.BackupPath, localDir, &cronjob, client)
	return paths, nil
}

func (u *CronjobService) handleWebsite(cronjob model.Cronjob, backup model.BackupAccount, startTime time.Time) ([]string, error) {
	var paths []string
	localDir, err := loadLocalDir()
	if err != nil {
		return paths, err
	}

	weblist := loadWebsForJob(cronjob)
	var client cloud_storage.CloudStorageClient
	if backup.Type != "LOCAL" {
		client, err = NewIBackupService().NewClient(&backup)
		if err != nil {
			return paths, err
		}
	}

	for _, websiteItem := range weblist {
		var record model.BackupRecord
		record.Type = "website"
		record.Name = websiteItem.PrimaryDomain
		record.DetailName = websiteItem.Alias
		record.Source = "LOCAL"
		record.BackupType = backup.Type
		backupDir := path.Join(localDir, fmt.Sprintf("website/%s", websiteItem.PrimaryDomain))
		record.FileDir = backupDir
		itemFileDir := strings.TrimPrefix(backupDir, localDir+"/")
		if !cronjob.KeepLocal && backup.Type != "LOCAL" {
			record.Source = backup.Type
			record.FileDir = strings.TrimPrefix(backupDir, localDir+"/")
		}
		record.FileName = fmt.Sprintf("website_%s_%s.tar.gz", websiteItem.PrimaryDomain, startTime.Format("20060102150405"))
		if err := handleWebsiteBackup(&websiteItem, backupDir, record.FileName); err != nil {
			return paths, err
		}
		if err := backupRepo.CreateRecord(&record); err != nil {
			global.LOG.Errorf("save backup record failed, err: %v", err)
			return paths, err
		}
		if backup.Type != "LOCAL" {
			if !cronjob.KeepLocal {
				defer func() {
					_ = os.RemoveAll(fmt.Sprintf("%s/%s", backupDir, record.FileName))
				}()
			}
			if len(backup.BackupPath) != 0 {
				itemFileDir = path.Join(strings.TrimPrefix(backup.BackupPath, "/"), itemFileDir)
			}
			if _, err = client.Upload(backupDir+"/"+record.FileName, itemFileDir+"/"+record.FileName); err != nil {
				return paths, err
			}
		}
		if backup.Type == "LOCAL" || cronjob.KeepLocal {
			paths = append(paths, fmt.Sprintf("%s/%s", record.FileDir, record.FileName))
		} else {
			paths = append(paths, fmt.Sprintf("%s/%s", itemFileDir, record.FileName))
		}
	}
	u.HandleRmExpired(backup.Type, backup.BackupPath, localDir, &cronjob, client)
	return paths, nil
}

func (u *CronjobService) handleSnapshot(cronjob *model.Cronjob, startTime time.Time) (string, string, error) {
	backup, err := backupRepo.Get(commonRepo.WithByID(uint(cronjob.TargetDirID)))
	if err != nil {
		return "", "", err
	}
	client, err := NewIBackupService().NewClient(&backup)
	if err != nil {
		return "", "", err
	}

	req := dto.SnapshotCreate{
		From: backup.Type,
	}
	message, name, err := NewISnapshotService().HandleSnapshot(true, req, startTime.Format("20060102150405"))
	if err != nil {
		return message, "", err
	}

	path := path.Join(strings.TrimPrefix(backup.BackupPath, "/"), "system_snapshot", name+".tar.gz")

	u.HandleRmExpired(backup.Type, backup.BackupPath, "", cronjob, client)
	return message, path, nil
}

func (u *CronjobService) handleSystemClean() (string, error) {
	return NewIDeviceService().CleanForCronjob()
}

func (u *CronjobService) handleSystemLog(cronjob model.Cronjob, startTime time.Time) (string, error) {
	backup, err := backupRepo.Get(commonRepo.WithByID(uint(cronjob.TargetDirID)))
	if err != nil {
		return "", err
	}

	pathItem := path.Join(global.CONF.System.BaseDir, "1panel/tmp/log", startTime.Format("20060102150405"))
	websites, err := websiteRepo.List()
	if err != nil {
		return "", err
	}
	if len(websites) != 0 {
		nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
		if err != nil {
			return "", err
		}
		webItem := path.Join(nginxInstall.GetPath(), "www/sites")
		for _, website := range websites {
			dirItem := path.Join(pathItem, "website", website.Alias)
			if _, err := os.Stat(dirItem); err != nil && os.IsNotExist(err) {
				if err = os.MkdirAll(dirItem, os.ModePerm); err != nil {
					return "", err
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
	systemDir := path.Join(pathItem, "system")
	if _, err := os.Stat(systemDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(systemDir, os.ModePerm); err != nil {
			return "", err
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
	loginDir := path.Join(pathItem, "login")
	if _, err := os.Stat(loginDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(loginDir, os.ModePerm); err != nil {
			return "", err
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

	fileName := fmt.Sprintf("system_log_%s.tar.gz", startTime.Format("20060102150405"))
	targetDir := path.Dir(pathItem)
	if err := handleTar(pathItem, targetDir, fileName, ""); err != nil {
		return "", err
	}
	defer func() {
		os.RemoveAll(pathItem)
		os.RemoveAll(path.Join(targetDir, fileName))
	}()

	client, err := NewIBackupService().NewClient(&backup)
	if err != nil {
		return "", err
	}

	targetPath := "log/" + fileName
	if len(backup.BackupPath) != 0 {
		targetPath = strings.TrimPrefix(backup.BackupPath, "/") + "/log/" + fileName
	}

	if _, err = client.Upload(path.Join(targetDir, fileName), targetPath); err != nil {
		return "", err
	}

	u.HandleRmExpired(backup.Type, backup.BackupPath, "", &cronjob, client)
	return targetPath, nil
}

func hasBackup(cronjobType string) bool {
	return cronjobType == "app" || cronjobType == "database" || cronjobType == "website" || cronjobType == "directory" || cronjobType == "snapshot" || cronjobType == "log"
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
