package service

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/utils/cloud_storage"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

func (u *CronjobService) HandleJob(cronjob *model.Cronjob) {
	var (
		message []byte
		err     error
	)
	record := cronjobRepo.StartRecords(cronjob.ID, "")
	switch cronjob.Type {
	case "shell":
		cmd := exec.Command(cronjob.Script)
		message, err = cmd.CombinedOutput()
	case "website":
		message, err = u.HandleBackup(cronjob, record.StartTime)
	case "database":
		message, err = u.HandleBackup(cronjob, record.StartTime)
	case "directory":
		if len(cronjob.SourceDir) == 0 {
			return
		}
		message, err = u.HandleBackup(cronjob, record.StartTime)
	case "curl":
		if len(cronjob.URL) == 0 {
			return
		}
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Timeout: 1 * time.Second, Transport: tr}
		request, _ := http.NewRequest("GET", cronjob.URL, nil)
		response, err := client.Do(request)
		if err != nil {
			record.Records = errHandle
			cronjobRepo.EndRecords(record, constant.StatusFailed, err.Error(), errHandle)
		}
		defer response.Body.Close()
		message, _ = ioutil.ReadAll(response.Body)
	}
	if err != nil {
		record.Records = errHandle
		cronjobRepo.EndRecords(record, constant.StatusFailed, err.Error(), errHandle)
		return
	}
	record.Records, err = mkdirAndWriteFile(cronjob, record.StartTime, message)
	if err != nil {
		record.Records = errRecord
		global.LOG.Errorf("save file %s failed, err: %v", record.Records, err)
	}
	cronjobRepo.EndRecords(record, constant.StatusSuccess, "", record.Records)
}

func (u *CronjobService) HandleBackup(cronjob *model.Cronjob, startTime time.Time) ([]byte, error) {
	var stdout []byte
	backup, err := backupRepo.Get(commonRepo.WithByID(uint(cronjob.TargetDirID)))
	if err != nil {
		return nil, err
	}
	commonDir := fmt.Sprintf("%s/%s/", cronjob.Type, cronjob.Name)
	name := fmt.Sprintf("%s.gz", startTime.Format("20060102150405"))
	if cronjob.Type != "database" {
		name = fmt.Sprintf("%s.tar.gz", startTime.Format("20060102150405"))
	}
	if backup.Type == "LOCAL" {
		varMap := make(map[string]interface{})
		if err := json.Unmarshal([]byte(backup.Vars), &varMap); err != nil {
			return nil, err
		}
		if _, ok := varMap["dir"]; !ok {
			return nil, errors.New("load local backup dir failed")
		}
		baseDir := varMap["dir"].(string)
		if _, err := os.Stat(baseDir); err != nil && os.IsNotExist(err) {
			if err = os.MkdirAll(baseDir, os.ModePerm); err != nil {
				if err != nil {
					return nil, fmt.Errorf("mkdir %s failed, err: %v", baseDir, err)
				}
			}
		}
		stdout, err = handleTar(cronjob.SourceDir, fmt.Sprintf("%s/%s", baseDir, commonDir), name, cronjob.ExclusionRules)
		if err != nil {
			return stdout, err
		}
		u.HandleRmExpired(backup.Type, fmt.Sprintf("%s/%s", baseDir, commonDir), cronjob, nil)
		return stdout, nil
	}
	targetDir := constant.TmpDir + commonDir
	client, err := NewIBackupService().NewClient(&backup)
	if err != nil {
		return nil, err
	}
	if cronjob.Type != "database" {
		stdout, err = handleTar(cronjob.SourceDir, targetDir, name, cronjob.ExclusionRules)
		if err != nil {
			return stdout, err
		}
	}
	if _, err = client.Upload(targetDir+name, commonDir+name); err != nil {
		return nil, err
	}
	u.HandleRmExpired(backup.Type, commonDir+name, cronjob, client)
	return stdout, nil
}

func (u *CronjobService) HandleDelete(id uint) error {
	cronjob, _ := cronjobRepo.Get(commonRepo.WithByID(id))
	if cronjob.ID == 0 {
		return errors.New("find cronjob in db failed")
	}
	commonDir := fmt.Sprintf("%s/%s/", cronjob.Type, cronjob.Name)
	global.Cron.Remove(cron.EntryID(cronjob.EntryID))
	_ = cronjobRepo.DeleteRecord(cronjobRepo.WithByJobID(int(id)))

	if err := os.RemoveAll(fmt.Sprintf("%s/%s-%v", constant.TaskDir, commonDir, cronjob.ID)); err != nil {
		global.LOG.Errorf("rm file %s/%s-%v failed, err: %v", constant.TaskDir, commonDir, cronjob.ID, err)
	}
	return nil
}

func (u *CronjobService) HandleRmExpired(backType, path string, cronjob *model.Cronjob, backClient cloud_storage.CloudStorageClient) {
	if backType != "LOCAL" {
		commonDir := fmt.Sprintf("%s/%s/", cronjob.Type, cronjob.Name)
		currentObjs, err := backClient.ListObjects(commonDir)
		if err != nil {
			global.LOG.Errorf("list bucket object %s failed, err: %v", commonDir, err)
			return
		}
		for i := 0; i < len(currentObjs)-int(cronjob.RetainCopies); i++ {
			_, _ = backClient.Delete(currentObjs[i].(string))
		}
		return
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		global.LOG.Errorf("read dir %s failed, err: %v", path, err)
		return
	}
	for i := 0; i < len(files)-int(cronjob.RetainCopies); i++ {
		_ = os.Remove(path + "/" + files[i].Name())
	}
	records, _ := cronjobRepo.ListRecord(cronjobRepo.WithByJobID(int(cronjob.ID)))
	if len(records) > int(cronjob.RetainCopies) {
		for i := int(cronjob.RetainCopies); i < len(records); i++ {
			_ = cronjobRepo.DeleteRecord(cronjobRepo.WithByJobID(int(records[i].ID)))
		}
	}
}

func handleTar(sourceDir, targetDir, name, exclusionRules string) ([]byte, error) {
	if _, err := os.Stat(targetDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return nil, err
		}
	}
	exStr := []string{}
	exStr = append(exStr, "zcvf")
	exStr = append(exStr, targetDir+name)
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
	return (cmd.CombinedOutput())
}
