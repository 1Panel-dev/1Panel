package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cloud_storage"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

type CronjobService struct{}

type ICronjobService interface {
	SearchWithPage(search dto.SearchWithPage) (int64, interface{}, error)
	SearchRecords(search dto.SearchRecord) (int64, interface{}, error)
	Create(cronjobDto dto.CronjobCreate) error
	HandleOnce(id uint) error
	Update(id uint, req dto.CronjobUpdate) error
	UpdateStatus(id uint, status string) error
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
		record, _ := cronjobRepo.RecordFirst(cronjob.ID)
		if record.ID != 0 {
			item.LastRecrodTime = record.StartTime.Format("2006-01-02 15:04:05")
		} else {
			item.LastRecrodTime = "-"
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

func (u *CronjobService) Download(down dto.CronjobDownload) (string, error) {
	record, _ := cronjobRepo.GetRecord(commonRepo.WithByID(down.RecordID))
	if record.ID == 0 {
		return "", constant.ErrRecordNotFound
	}
	cronjob, _ := cronjobRepo.Get(commonRepo.WithByID(record.CronjobID))
	if cronjob.ID == 0 {
		return "", constant.ErrRecordNotFound
	}
	backup, _ := backupRepo.Get(commonRepo.WithByID(down.BackupAccountID))
	if cronjob.ID == 0 {
		return "", constant.ErrRecordNotFound
	}
	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(backup.Vars), &varMap); err != nil {
		return "", err
	}
	varMap["type"] = backup.Type
	if backup.Type != "LOCAL" {
		varMap["bucket"] = backup.Bucket
		switch backup.Type {
		case constant.Sftp:
			varMap["password"] = backup.Credential
		case constant.OSS, constant.S3, constant.MinIo:
			varMap["secretKey"] = backup.Credential
		}
		backClient, err := cloud_storage.NewCloudStorageClient(varMap)
		if err != nil {
			return "", fmt.Errorf("new cloud storage client failed, err: %v", err)
		}
		commonDir := fmt.Sprintf("%s/%s/", cronjob.Type, cronjob.Name)
		name := fmt.Sprintf("%s%s.tar.gz", commonDir, record.StartTime.Format("20060102150405"))
		if cronjob.Type == "database" {
			name = fmt.Sprintf("%s%s.gz", commonDir, record.StartTime.Format("20060102150405"))
		}
		tempPath := fmt.Sprintf("%s%s", constant.DownloadDir, commonDir)
		if _, err := os.Stat(tempPath); err != nil && os.IsNotExist(err) {
			if err = os.MkdirAll(tempPath, os.ModePerm); err != nil {
				fmt.Println(err)
			}
		}
		targetPath := tempPath + strings.ReplaceAll(name, commonDir, "")
		if _, err = os.Stat(targetPath); err != nil && os.IsNotExist(err) {
			isOK, err := backClient.Download(name, targetPath)
			if !isOK {
				return "", fmt.Errorf("cloud storage download failed, err: %v", err)
			}
		}
		return targetPath, nil
	}
	if _, ok := varMap["dir"]; !ok {
		return "", errors.New("load local backup dir failed")
	}
	dir := fmt.Sprintf("%v/%s/%s/", varMap["dir"], cronjob.Type, cronjob.Name)
	name := fmt.Sprintf("%s%s.tar.gz", dir, record.StartTime.Format("20060102150405"))
	if cronjob.Type == "database" {
		name = fmt.Sprintf("%s%s.gz", dir, record.StartTime.Format("20060102150405"))
	}
	return name, nil
}

func (u *CronjobService) HandleOnce(id uint) error {
	cronjob, _ := cronjobRepo.Get(commonRepo.WithByID(id))
	if cronjob.ID == 0 {
		return constant.ErrRecordNotFound
	}
	u.HandleJob(&cronjob)
	return nil
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
	if err := u.StartJob(&cronjob); err != nil {
		return err
	}
	return nil
}

func (u *CronjobService) StartJob(cronjob *model.Cronjob) error {
	global.Cron.Remove(cron.EntryID(cronjob.EntryID))
	entryID, err := u.AddCronJob(cronjob)
	if err != nil {
		return err
	}
	_ = cronjobRepo.Update(cronjob.ID, map[string]interface{}{"entry_id": entryID})
	return nil
}

func (u *CronjobService) Delete(ids []uint) error {
	if len(ids) == 1 {
		if err := u.HandleDelete(ids[0]); err != nil {
			return err
		}
		return cronjobRepo.Delete(commonRepo.WithByID(ids[0]))
	}
	cronjobs, err := cronjobRepo.List(commonRepo.WithIdsIn(ids))
	if err != nil {
		return err
	}
	for i := range cronjobs {
		_ = u.HandleDelete(ids[i])
	}
	return cronjobRepo.Delete(commonRepo.WithIdsIn(ids))
}

func (u *CronjobService) Update(id uint, req dto.CronjobUpdate) error {
	var cronjob model.Cronjob
	if err := copier.Copy(&cronjob, &req); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	cronjob.Spec = loadSpec(cronjob)
	if err := u.StartJob(&cronjob); err != nil {
		return err
	}

	upMap := make(map[string]interface{})
	upMap["name"] = req.Name
	upMap["script"] = req.Script
	upMap["spec_type"] = req.SpecType
	upMap["week"] = req.Week
	upMap["day"] = req.Day
	upMap["hour"] = req.Hour
	upMap["minute"] = req.Minute
	upMap["website"] = req.Website
	upMap["exclusion_rules"] = req.ExclusionRules
	upMap["database"] = req.Database
	upMap["url"] = req.URL
	upMap["source_dir"] = req.SourceDir
	upMap["target_dir_id"] = req.TargetDirID
	upMap["retain_days"] = req.RetainCopies
	return cronjobRepo.Update(id, upMap)
}

func (u *CronjobService) UpdateStatus(id uint, status string) error {
	cronjob, _ := cronjobRepo.Get(commonRepo.WithByID(id))
	if cronjob.ID == 0 {
		return errors.WithMessage(constant.ErrRecordNotFound, "record not found")
	}
	if status == constant.StatusEnable {
		if err := u.StartJob(&cronjob); err != nil {
			return err
		}
	} else {
		global.Cron.Remove(cron.EntryID(cronjob.EntryID))
	}
	return cronjobRepo.Update(cronjob.ID, map[string]interface{}{"status": status})
}

func (u *CronjobService) AddCronJob(cronjob *model.Cronjob) (int, error) {
	addFunc := func() {
		u.HandleJob(cronjob)
	}
	global.LOG.Infof("add %s job %s successful", cronjob.Type, cronjob.Name)
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
