package service

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cloud_storage"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
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
			stdout, errExec := cmd.ExecWithTimeOut(cronjob.Script, 5*time.Minute)
			if errExec != nil {
				err = errExec
			}
			message = []byte(stdout)
			u.HandleRmExpired("LOCAL", "", cronjob, nil)
		case "website":
			record.File, err = u.HandleBackup(cronjob, record.StartTime)
		case "database":
			record.File, err = u.HandleBackup(cronjob, record.StartTime)
		case "directory":
			if len(cronjob.SourceDir) == 0 {
				return
			}
			record.File, err = u.HandleBackup(cronjob, record.StartTime)
		case "curl":
			if len(cronjob.URL) == 0 {
				return
			}
			stdout, errCurl := cmd.ExecWithTimeOut("curl "+cronjob.URL, 5*time.Minute)
			if err != nil {
				err = errCurl
			}
			message = []byte(stdout)
			u.HandleRmExpired("LOCAL", "", cronjob, nil)
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

func (u *CronjobService) HandleBackup(cronjob *model.Cronjob, startTime time.Time) (string, error) {
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
		app, err := appInstallRepo.LoadBaseInfo("mysql", "")
		if err != nil {
			return "", err
		}
		paths, err := u.handleDatabase(*cronjob, app, backup, startTime)
		return strings.Join(paths, ","), err
	case "website":
		paths, err := u.handleWebsite(*cronjob, backup, startTime)
		return strings.Join(paths, ","), err
	default:
		fileName := fmt.Sprintf("directory%s_%s.tar.gz", strings.ReplaceAll(cronjob.SourceDir, "/", "_"), startTime.Format("20060102150405"))
		backupDir := fmt.Sprintf("%s/%s/%s", localDir, cronjob.Type, cronjob.Name)
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
			if _, err = client.Upload(backupDir+"/"+fileName, itemFileDir+"/"+fileName); err != nil {
				return "", err
			}
		}
		u.HandleRmExpired(backup.Type, localDir, cronjob, client)
		if backup.Type == "LOCAL" || cronjob.KeepLocal {
			return fmt.Sprintf("%s/%s/%s/%s", localDir, cronjob.Type, cronjob.Name, fileName), nil
		}
		return fmt.Sprintf("%s/%s/%s", cronjob.Type, cronjob.Name, fileName), nil
	}
}

func (u *CronjobService) HandleRmExpired(backType, localDir string, cronjob *model.Cronjob, backClient cloud_storage.CloudStorageClient) {
	global.LOG.Infof("start to handle remove expired, retain copies: %d", cronjob.RetainCopies)
	records, _ := cronjobRepo.ListRecord(cronjobRepo.WithByJobID(int(cronjob.ID)), commonRepo.WithOrderBy("created_at desc"))
	if len(records) > int(cronjob.RetainCopies) {
		for i := int(cronjob.RetainCopies); i < len(records); i++ {
			files := strings.Split(records[i].File, ",")
			for _, file := range files {
				if backType != "LOCAL" {
					_, _ = backClient.Delete(strings.ReplaceAll(file, localDir+"/", ""))
					_ = os.Remove(file)
				} else {
					_ = os.Remove(file)
				}
				_ = backupRepo.DeleteRecord(context.TODO(), backupRepo.WithByFileName(path.Base(file)))
			}

			_ = cronjobRepo.DeleteRecord(commonRepo.WithByID(uint(records[i].ID)))
			_ = os.Remove(records[i].Records)
		}
	}
}

func handleTar(sourceDir, targetDir, name, exclusionRules string) error {
	if _, err := os.Stat(targetDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return err
		}
	}

	excludes := strings.Split(exclusionRules, ";")
	excludeRules := ""
	for _, exclude := range excludes {
		if len(exclude) == 0 {
			continue
		}
		excludeRules += (" --exclude " + exclude)
	}
	path := ""
	if strings.Contains(sourceDir, "/") {
		itemDir := strings.ReplaceAll(sourceDir[strings.LastIndex(sourceDir, "/"):], "/", "")
		aheadDir := strings.ReplaceAll(sourceDir, itemDir, "")
		path += fmt.Sprintf("-C %s %s", aheadDir, itemDir)
	} else {
		path = sourceDir
	}

	commands := fmt.Sprintf("tar zcvf %s %s %s", targetDir+"/"+name, excludeRules, path)
	global.LOG.Debug(commands)
	stdout, err := cmd.ExecWithTimeOut(commands, 5*time.Minute)
	if err != nil {
		global.LOG.Errorf("do handle tar failed, stdout: %s, err: %v", stdout, err)
		return errors.New(stdout)
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
	stdout, err := cmd.ExecWithTimeOut(commands, 5*time.Minute)
	if err != nil {
		global.LOG.Errorf("do handle untar failed, stdout: %s, err: %v", stdout, err)
		return errors.New(stdout)
	}
	return nil
}

func (u *CronjobService) handleDatabase(cronjob model.Cronjob, app *repo.RootInfo, backup model.BackupAccount, startTime time.Time) ([]string, error) {
	var paths []string
	localDir, err := loadLocalDir()
	if err != nil {
		return paths, err
	}

	var dblist []string
	if cronjob.DBName == "all" {
		mysqlService := NewIMysqlService()
		dblist, err = mysqlService.ListDBName()
		if err != nil {
			return paths, err
		}
	} else {
		dblist = append(dblist, cronjob.DBName)
	}

	var client cloud_storage.CloudStorageClient
	if backup.Type != "LOCAL" {
		client, err = NewIBackupService().NewClient(&backup)
		if err != nil {
			return paths, err
		}
	}

	for _, dbName := range dblist {
		var record model.BackupRecord

		record.Type = "mysql"
		record.Name = app.Name
		record.Source = "LOCAL"
		record.BackupType = backup.Type

		backupDir := fmt.Sprintf("%s/database/mysql/%s/%s", localDir, app.Name, dbName)
		record.FileName = fmt.Sprintf("db_%s_%s.sql.gz", dbName, startTime.Format("20060102150405"))
		if err = handleMysqlBackup(app, backupDir, dbName, record.FileName); err != nil {
			return paths, err
		}
		record.DetailName = dbName
		record.FileDir = backupDir
		itemFileDir := strings.ReplaceAll(backupDir, localDir+"/", "")
		if !cronjob.KeepLocal && backup.Type != "LOCAL" {
			record.Source = backup.Type
			record.FileDir = itemFileDir
		}
		paths = append(paths, fmt.Sprintf("%s/%s", record.FileDir, record.FileName))

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
			if _, err = client.Upload(backupDir+"/"+record.FileName, itemFileDir+"/"+record.FileName); err != nil {
				return paths, err
			}
		}
	}
	u.HandleRmExpired(backup.Type, localDir, &cronjob, client)
	return paths, nil
}

func (u *CronjobService) handleWebsite(cronjob model.Cronjob, backup model.BackupAccount, startTime time.Time) ([]string, error) {
	var paths []string
	localDir, err := loadLocalDir()
	if err != nil {
		return paths, err
	}

	var weblist []string
	if cronjob.Website == "all" {
		weblist, err = NewIWebsiteService().GetWebsiteOptions()
		if err != nil {
			return paths, err
		}
	} else {
		weblist = append(weblist, cronjob.Website)
	}

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
		record.Name = cronjob.Website
		record.Source = "LOCAL"
		record.BackupType = backup.Type
		website, err := websiteRepo.GetFirst(websiteRepo.WithDomain(websiteItem))
		if err != nil {
			return paths, err
		}
		backupDir := fmt.Sprintf("%s/website/%s", localDir, website.PrimaryDomain)
		record.FileDir = backupDir
		itemFileDir := strings.ReplaceAll(backupDir, localDir+"/", "")
		if !cronjob.KeepLocal && backup.Type != "LOCAL" {
			record.Source = backup.Type
			record.FileDir = strings.ReplaceAll(backupDir, localDir+"/", "")
		}
		record.FileName = fmt.Sprintf("website_%s_%s.tar.gz", website.PrimaryDomain, startTime.Format("20060102150405"))
		paths = append(paths, fmt.Sprintf("%s/%s", record.FileDir, record.FileName))
		if err := handleWebsiteBackup(&website, backupDir, record.FileName); err != nil {
			return paths, err
		}
		record.Name = website.PrimaryDomain
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
			if _, err = client.Upload(backupDir+"/"+record.FileName, itemFileDir+"/"+record.FileName); err != nil {
				return paths, err
			}
		}
	}
	u.HandleRmExpired(backup.Type, localDir, &cronjob, client)
	return paths, nil
}
