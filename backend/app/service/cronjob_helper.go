package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cloud_storage"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

func (u *CronjobService) HandleJob(cronjob *model.Cronjob) {
	var (
		message []byte
		err     error
	)
	record := cronjobRepo.StartRecords(cronjob.ID, "")
	record.FromLocal = cronjob.KeepLocal
	go func() {
		switch cronjob.Type {
		case "shell":
			if len(cronjob.Script) == 0 {
				return
			}
			stdout, errExec := cmd.Exec(cronjob.Script)
			if errExec != nil {
				err = errExec
			}
			message = []byte(stdout)
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
			stdout, errCurl := cmd.Exec("curl " + cronjob.URL)
			if err != nil {
				err = errCurl
			}
			message = []byte(stdout)
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
	var (
		backupDir string
		fileName  string
		record    model.BackupRecord
	)
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
		fileName = fmt.Sprintf("db_%s_%s.sql.gz", cronjob.DBName, startTime.Format("20060102150405"))
		backupDir = fmt.Sprintf("%s/database/mysql/%s/%s", localDir, app.Name, cronjob.DBName)
		if err = handleMysqlBackup(app, backupDir, cronjob.DBName, fileName); err != nil {
			return "", err
		}
		record.Type = "mysql"
		record.Name = app.Name
		record.DetailName = cronjob.DBName
	case "website":
		fileName = fmt.Sprintf("website_%s_%s.tar.gz", cronjob.Website, startTime.Format("20060102150405"))
		backupDir = fmt.Sprintf("%s/website/%s", localDir, cronjob.Website)
		website, err := websiteRepo.GetFirst(websiteRepo.WithDomain(cronjob.Website))
		if err != nil {
			return "", err
		}
		if err := handleWebsiteBackup(&website, backupDir, fileName); err != nil {
			return "", err
		}
		record.Type = "website"
		record.Name = website.PrimaryDomain
	default:
		fileName = fmt.Sprintf("directory%s_%s.tar.gz", strings.ReplaceAll(cronjob.SourceDir, "/", "_"), startTime.Format("20060102150405"))
		backupDir = fmt.Sprintf("%s/%s/%s", localDir, cronjob.Type, cronjob.Name)
		global.LOG.Infof("handle tar %s to %s", backupDir, fileName)
		if err := handleTar(cronjob.SourceDir, backupDir, fileName, cronjob.ExclusionRules); err != nil {
			return "", err
		}
	}
	if len(record.Name) != 0 {
		record.FileName = fileName
		record.FileDir = backupDir
		record.Source = "LOCAL"
		record.BackupType = backup.Type
		if !cronjob.KeepLocal && backup.Type != "LOCAL" {
			record.Source = backup.Type
			record.FileDir = strings.ReplaceAll(backupDir, localDir+"/", "")
		}
		if err := backupRepo.CreateRecord(&record); err != nil {
			global.LOG.Errorf("save backup record failed, err: %v", err)
			return "", err
		}
	}

	fullPath := fmt.Sprintf("%s/%s", record.FileDir, fileName)
	if backup.Type == "LOCAL" {
		u.HandleRmExpired(backup.Type, record.FileDir, cronjob, nil)
		return fullPath, nil
	}

	if !cronjob.KeepLocal {
		defer func() {
			_ = os.RemoveAll(fmt.Sprintf("%s/%s", backupDir, fileName))
		}()
	}
	client, err := NewIBackupService().NewClient(&backup)
	if err != nil {
		return fullPath, err
	}
	if _, err = client.Upload(backupDir+"/"+fileName, fullPath); err != nil {
		return fullPath, err
	}
	u.HandleRmExpired(backup.Type, backupDir, cronjob, client)
	return fullPath, nil
}

func (u *CronjobService) HandleDelete(id uint) error {
	cronjob, _ := cronjobRepo.Get(commonRepo.WithByID(id))
	if cronjob.ID == 0 {
		return errors.New("find cronjob in db failed")
	}
	commonDir := fmt.Sprintf("%s/%s/", cronjob.Type, cronjob.Name)
	global.Cron.Remove(cron.EntryID(cronjob.EntryID))
	global.LOG.Infof("stop cronjob entryID: %d", cronjob.EntryID)
	_ = cronjobRepo.DeleteRecord(cronjobRepo.WithByJobID(int(id)))

	dir := fmt.Sprintf("%s/task/%s/%s", constant.DataDir, cronjob.Type, cronjob.Name)
	if _, err := os.Stat(dir); err == nil {
		if err := os.RemoveAll(dir); err != nil {
			global.LOG.Errorf("rm file %s/task/%s failed, err: %v", constant.DataDir, commonDir, err)
		}
	}
	return nil
}

func (u *CronjobService) HandleRmExpired(backType, backupDir string, cronjob *model.Cronjob, backClient cloud_storage.CloudStorageClient) {
	global.LOG.Infof("start to handle remove expired, retain copies: %d", cronjob.RetainCopies)
	if backType != "LOCAL" {
		currentObjs, err := backClient.ListObjects(backupDir + "/")
		if err != nil {
			global.LOG.Errorf("list bucket object %s failed, err: %v", backupDir, err)
			return
		}
		for i := 0; i < len(currentObjs)-int(cronjob.RetainCopies); i++ {
			_, _ = backClient.Delete(currentObjs[i].(string))
		}
		if !cronjob.KeepLocal {
			return
		}
	}
	files, err := ioutil.ReadDir(backupDir)
	if err != nil {
		global.LOG.Errorf("read dir %s failed, err: %v", backupDir, err)
		return
	}
	if len(files) == 0 {
		return
	}

	prefix := ""
	switch cronjob.Type {
	case "database":
		prefix = "db_"
	case "website":
		prefix = "website_"
	case "directory":
		prefix = "directory_"
	}

	dbCopies := uint64(0)
	for i := len(files) - 1; i >= 0; i-- {
		if strings.HasPrefix(files[i].Name(), prefix) {
			dbCopies++
			if dbCopies > cronjob.RetainCopies {
				_ = os.Remove(backupDir + "/" + files[i].Name())
				_ = backupRepo.DeleteRecord(context.Background(), backupRepo.WithByFileName(files[i].Name()))
			}
		}
	}
	records, _ := cronjobRepo.ListRecord(cronjobRepo.WithByJobID(int(cronjob.ID)))
	if len(records) > int(cronjob.RetainCopies) {
		for i := int(cronjob.RetainCopies); i < len(records); i++ {
			_ = cronjobRepo.DeleteRecord(cronjobRepo.WithByJobID(int(records[i].ID)))
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
	stdout, err := cmd.Exec(commands)
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
	stdout, err := cmd.Exec(commands)
	if err != nil {
		global.LOG.Errorf("do handle untar failed, stdout: %s, err: %v", stdout, err)
		return errors.New(stdout)
	}
	return nil
}
