package service

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/utils/cloud_storage"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

const (
	errRecord = "errRecord"
	errHandle = "errHandle"
)

type CronjobService struct{}

type ICronjobService interface {
	SearchWithPage(search dto.SearchWithPage) (int64, interface{}, error)
	SearchRecords(search dto.SearchRecord) (int64, interface{}, error)
	Create(cronjobDto dto.CronjobCreate) error
	Save(id uint, req dto.CronjobUpdate) error
	Delete(ids []uint) error
}

func NewICronjobService() ICronjobService {
	return &CronjobService{}
}

func (u *CronjobService) SearchWithPage(search dto.SearchWithPage) (int64, interface{}, error) {
	total, cronjobs, err := cronjobRepo.Page(search.Page, search.PageSize, commonRepo.WithLikeName(search.Info))
	var dtoCronjobs []dto.CronjobInfo
	for _, cronjob := range cronjobs {
		var item dto.CronjobInfo
		if err := copier.Copy(&item, &cronjob); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		if item.Type == "website" || item.Type == "database" || item.Type == "directory" {
			backup, _ := backupRepo.Get(commonRepo.WithByID(uint(item.TargetDirID)))
			if len(backup.Type) != 0 {
				item.TargetDir = backup.Type
			}
		} else {
			item.TargetDir = "-"
		}
		dtoCronjobs = append(dtoCronjobs, item)
	}
	return total, dtoCronjobs, err
}

func (u *CronjobService) SearchRecords(search dto.SearchRecord) (int64, interface{}, error) {
	total, records, err := cronjobRepo.PageRecords(
		search.Page,
		search.PageSize,
		commonRepo.WithByStatus(search.Status),
		cronjobRepo.WithByJobID(search.CronjobID),
		cronjobRepo.WithByDate(search.StartTime, search.EndTime))
	var dtoCronjobs []dto.Record
	for _, record := range records {
		var item dto.Record
		if err := copier.Copy(&item, &record); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtoCronjobs = append(dtoCronjobs, item)
	}
	return total, dtoCronjobs, err
}

func (u *CronjobService) Create(cronjobDto dto.CronjobCreate) error {
	cronjob, _ := cronjobRepo.Get(commonRepo.WithByName(cronjobDto.Name))
	if cronjob.ID != 0 {
		return constant.ErrRecordExist
	}
	if err := copier.Copy(&cronjob, &cronjobDto); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	cronjob.Status = constant.StatusEnable
	cronjob.Spec = loadSpec(cronjob)

	if err := cronjobRepo.Create(&cronjob); err != nil {
		return err
	}
	var (
		entryID int
		err     error
	)
	switch cronjobDto.Type {
	case "shell":
		entryID, err = u.AddShellJob(&cronjob)
	case "curl":
		entryID, err = u.AddCurlJob(&cronjob)
	case "directory":
		entryID, err = u.AddDirectoryJob(&cronjob)
	}

	if err != nil {
		return err
	}
	_ = cronjobRepo.Update(cronjob.ID, map[string]interface{}{"entry_id": entryID})
	return nil
}

func (u *CronjobService) Delete(ids []uint) error {
	if len(ids) == 1 {
		cronjob, _ := cronjobRepo.Get(commonRepo.WithByID(ids[0]))
		if cronjob.ID == 0 {
			return constant.ErrRecordNotFound
		}
		return cronjobRepo.Delete(commonRepo.WithByID(ids[0]))
	}
	return cronjobRepo.Delete(commonRepo.WithIdsIn(ids))
}

func (u *CronjobService) Save(id uint, req dto.CronjobUpdate) error {
	var cronjob model.Cronjob
	if err := copier.Copy(&cronjob, &req); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	return cronjobRepo.Save(id, cronjob)
}

func (u *CronjobService) AddShellJob(cronjob *model.Cronjob) (int, error) {
	addFunc := func() {
		record := cronjobRepo.StartRecords(cronjob.ID, "")

		cmd := exec.Command(cronjob.Script)
		stdout, err := cmd.Output()
		if err != nil {
			record.Records = errHandle
			cronjobRepo.EndRecords(record, constant.StatusFailed, err.Error(), errHandle)
			return
		}
		record.Records, err = mkdirAndWriteFile(cronjob, record.StartTime, stdout)
		if err != nil {
			record.Records = errRecord
			global.LOG.Errorf("save file %s failed, err: %v", record.Records, err)
		}
		cronjobRepo.EndRecords(record, constant.StatusSuccess, "", record.Records)
	}
	entryID, err := global.Cron.AddFunc(cronjob.Spec, addFunc)
	if err != nil {
		return 0, err
	}
	return int(entryID), nil
}

func (u *CronjobService) AddCurlJob(cronjob *model.Cronjob) (int, error) {
	addFunc := func() {
		record := cronjobRepo.StartRecords(cronjob.ID, "")
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
		stdout, _ := ioutil.ReadAll(response.Body)

		record.Records, err = mkdirAndWriteFile(cronjob, record.StartTime, stdout)
		if err != nil {
			record.Records = errRecord
			global.LOG.Errorf("save file %s failed, err: %v", record.Records, err)
		}
		cronjobRepo.EndRecords(record, constant.StatusSuccess, "", record.Records)
	}
	entryID, err := global.Cron.AddFunc(cronjob.Spec, addFunc)
	if err != nil {
		return 0, err
	}
	return int(entryID), nil
}

func (u *CronjobService) AddDirectoryJob(cronjob *model.Cronjob) (int, error) {
	addFunc := func() {
		record := cronjobRepo.StartRecords(cronjob.ID, "")
		if len(cronjob.SourceDir) == 0 {
			return
		}
		message, err := tarWithExclude(cronjob, record.StartTime)
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
	entryID, err := global.Cron.AddFunc(cronjob.Spec, addFunc)
	if err != nil {
		return 0, err
	}
	return int(entryID), nil
}

func (u *CronjobService) AddWebSiteJob(cronjob *model.Cronjob) (int, error) {
	addFunc := func() {
		record := cronjobRepo.StartRecords(cronjob.ID, "")
		if len(cronjob.URL) == 0 {
			return
		}
		message, err := tarWithExclude(cronjob, record.StartTime)
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
	entryID, err := global.Cron.AddFunc(cronjob.Spec, addFunc)
	if err != nil {
		return 0, err
	}
	return int(entryID), nil
}

func mkdirAndWriteFile(cronjob *model.Cronjob, startTime time.Time, msg []byte) (string, error) {
	dir := fmt.Sprintf("%s%s/%s-%v", constant.TaskDir, cronjob.Type, cronjob.Name, cronjob.ID)
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return "", err
		}
	}

	path := fmt.Sprintf("%s/%s.log", dir, startTime.Format("20060102150405"))
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(string(msg))
	write.Flush()
	return path, nil
}

func tarWithExclude(cronjob *model.Cronjob, startTime time.Time) ([]byte, error) {
	varMaps, targetdir, err := loadTargetInfo(cronjob)
	if err != nil {
		return nil, fmt.Errorf("load target dir failed, err: %v", err)
	}

	excludes := strings.Split(cronjob.ExclusionRules, ";")
	name := fmt.Sprintf("%s/%s.tar.gz", targetdir, startTime.Format("20060102150405"))
	exStr := []string{"-zcvPf", name}
	for _, exclude := range excludes {
		exStr = append(exStr, "--exclude")
		exStr = append(exStr, exclude)
	}
	exStr = append(exStr, cronjob.SourceDir)
	cmd := exec.Command("tar", exStr...)
	stdout, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("tar zcvPf failed, err: %v", err)
	}

	if varMaps["type"] != "LOCAL" {
		backClient, err := cloud_storage.NewCloudStorageClient(varMaps)
		if err != nil {
			return stdout, fmt.Errorf("new cloud storage client failed, err: %v", err)
		}
		isOK, err := backClient.Upload(name, strings.Replace(name, constant.TmpDir, "", -1))
		if !isOK {
			return nil, fmt.Errorf("cloud storage upload failed, err: %v", err)
		}
	}
	return stdout, nil
}

func loadTargetInfo(cronjob *model.Cronjob) (map[string]interface{}, string, error) {
	backup, err := backupRepo.Get(commonRepo.WithByID(uint(cronjob.TargetDirID)))
	if err != nil {
		return nil, "", err
	}
	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(backup.Vars), &varMap); err != nil {
		return nil, "", err
	}
	dir := ""
	if backup.Type != "LOCAL" {
		varMap["type"] = backup.Type
		varMap["bucket"] = backup.Bucket
		switch backup.Type {
		case constant.Sftp:
			varMap["password"] = backup.Credential
		case constant.OSS, constant.S3, constant.MinIo:
			varMap["secretKey"] = backup.Credential
		}
		dir = fmt.Sprintf("%s%s/%s", constant.TmpDir, cronjob.Type, cronjob.Name)
	} else {
		if _, ok := varMap["dir"]; !ok {
			return nil, "", errors.New("load local backup dir failed")
		}
		dir = fmt.Sprintf("%v/%s/%s", varMap["dir"], cronjob.Type, cronjob.Name)
	}

	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			if err != nil {
				return nil, "", fmt.Errorf("mkdir %s failed, err: %v", dir, err)
			}
		}
	}
	return varMap, dir, nil
}

func loadSpec(cronjob model.Cronjob) string {
	switch cronjob.SpecType {
	case "perMonth":
		return fmt.Sprintf("%v %v %v * *", cronjob.Minute, cronjob.Hour, cronjob.Day)
	case "perWeek":
		return fmt.Sprintf("%v %v * * %v", cronjob.Minute, cronjob.Hour, cronjob.Week)
	case "perNDay":
		return fmt.Sprintf("%v %v */%v * *", cronjob.Minute, cronjob.Hour, cronjob.Day)
	case "perNHour":
		return fmt.Sprintf("%v */%v * * *", cronjob.Minute, cronjob.Hour)
	case "perHour":
		return fmt.Sprintf("%v * * * *", cronjob.Minute)
	case "perNMinute":
		return fmt.Sprintf("@every %vm", cronjob.Minute)
	default:
		return ""
	}
}
