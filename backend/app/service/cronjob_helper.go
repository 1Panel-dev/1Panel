package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
}

func (u *CronjobService) HandleBackup(cronjob *model.Cronjob, startTime time.Time) (string, error) {
	var (
		baseDir   string
		backupDir string
		fileName  string
	)
	backup, err := backupRepo.Get(commonRepo.WithByID(uint(cronjob.TargetDirID)))
	if err != nil {
		return "", err
	}

	global.LOG.Infof("start to backup %s %s to %s", cronjob.Type, cronjob.Name, backup.Type)
	if cronjob.KeepLocal || cronjob.Type != "LOCAL" {
		localDir, err := loadLocalDir()
		if err != nil {
			return "", err
		}
		baseDir = localDir
	} else {
		baseDir = global.CONF.System.TmpDir
	}

	switch cronjob.Type {
	case "database":
		app, err := appInstallRepo.LoadBaseInfo("mysql", "")
		if err != nil {
			return "", err
		}
		fileName = fmt.Sprintf("db_%s_%s.sql.gz", cronjob.DBName, startTime.Format("20060102150405"))
		backupDir = fmt.Sprintf("database/mysql/%s/%s", app.Name, cronjob.DBName)
		if err = backupMysql(backup.Type, baseDir, backupDir, app.Name, cronjob.DBName, fileName); err != nil {
			return "", err
		}
	case "website":
		fileName = fmt.Sprintf("website_%s_%s", cronjob.Website, startTime.Format("20060102150405"))
		backupDir = fmt.Sprintf("website/%s", cronjob.Website)
		if err := handleWebsiteBackup(backup.Type, baseDir, backupDir, cronjob.Website, fileName); err != nil {
			return "", err
		}
		fileName = fileName + ".tar.gz"
	default:
		fileName = fmt.Sprintf("%s.tar.gz", startTime.Format("20060102150405"))
		backupDir = fmt.Sprintf("%s/%s", cronjob.Type, cronjob.Name)
		global.LOG.Infof("handle tar %s to %s", backupDir, fileName)
		if err := handleTar(cronjob.SourceDir, baseDir+"/"+backupDir, fileName, cronjob.ExclusionRules); err != nil {
			return "", err
		}
	}

	if backup.Type == "LOCAL" {
		u.HandleRmExpired(backup.Type, baseDir, backupDir, cronjob, nil)
		return baseDir + "/" + backupDir + "/" + fileName, nil
	}

	cloudFile := baseDir + "/" + backupDir + "/" + fileName
	if !cronjob.KeepLocal {
		cloudFile = backupDir + "/" + fileName
	}
	client, err := NewIBackupService().NewClient(&backup)
	if err != nil {
		return cloudFile, err
	}
	if _, err = client.Upload(baseDir+"/"+backupDir+"/"+fileName, backupDir+"/"+fileName); err != nil {
		return cloudFile, err
	}
	u.HandleRmExpired(backup.Type, baseDir, backupDir, cronjob, client)
	return cloudFile, nil
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

func (u *CronjobService) HandleRmExpired(backType, baseDir, backupDir string, cronjob *model.Cronjob, backClient cloud_storage.CloudStorageClient) {
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
	files, err := ioutil.ReadDir(baseDir + "/" + backupDir)
	if err != nil {
		global.LOG.Errorf("read dir %s failed, err: %v", baseDir+"/"+backupDir, err)
		return
	}
	if len(files) == 0 {
		return
	}
	if cronjob.Type == "database" {
		dbCopies := uint64(0)
		for i := len(files) - 1; i >= 0; i-- {
			if strings.HasPrefix(files[i].Name(), "db_") {
				dbCopies++
				if dbCopies > cronjob.RetainCopies {
					_ = os.Remove(baseDir + "/" + backupDir + "/" + files[i].Name())
					_ = backupRepo.DeleteRecord(context.Background(), backupRepo.WithByFileName(files[i].Name()))
				}
			}
		}
	} else {
		for i := 0; i < len(files)-int(cronjob.RetainCopies); i++ {
			_ = os.Remove(baseDir + "/" + backupDir + "/" + files[i].Name())
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
	exStr := []string{}
	exStr = append(exStr, "zcvf")
	exStr = append(exStr, targetDir+"/"+name)
	excludes := strings.Split(exclusionRules, ";")
	for _, exclude := range excludes {
		if len(exclude) == 0 {
			continue
		}
		exStr = append(exStr, "--exclude")
		exStr = append(exStr, exclude)
	}
	if len(strings.Split(sourceDir, "/")) > 3 {
		exStr = append(exStr, "-C")
		itemDir := strings.ReplaceAll(sourceDir[strings.LastIndex(sourceDir, "/"):], "/", "")
		aheadDir := strings.ReplaceAll(sourceDir, itemDir, "")
		exStr = append(exStr, aheadDir)
		exStr = append(exStr, itemDir)
	} else {
		exStr = append(exStr, sourceDir)
	}
	cmd := exec.Command("tar", exStr...)
	stdout, err := cmd.CombinedOutput()
	fmt.Println(string(stdout))
	if err != nil {
		return errors.New(string(stdout))
	}
	return nil
}

func handleUnTar(sourceFile, targetDir string) error {
	if _, err := os.Stat(targetDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return err
		}
	}

	cmd := exec.Command("tar", "zxvfC", sourceFile, targetDir)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(stdout))
	}
	return nil
}
