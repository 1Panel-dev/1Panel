package service

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
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
	Delete(req dto.CronjobBatchDelete) error
	Download(down dto.CronjobDownload) (string, error)
	StartJob(cronjob *model.Cronjob) (int, error)
	CleanRecord(req dto.CronjobClean) error
}

func NewICronjobService() ICronjobService {
	return &CronjobService{}
}

func (u *CronjobService) SearchWithPage(search dto.SearchWithPage) (int64, interface{}, error) {
	total, cronjobs, err := cronjobRepo.Page(search.Page, search.PageSize, commonRepo.WithLikeName(search.Info), commonRepo.WithOrderRuleBy(search.OrderBy, search.Order))
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
			item.LastRecordTime = record.StartTime.Format("2006-01-02 15:04:05")
		} else {
			item.LastRecordTime = "-"
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
		commonRepo.WithByDate(search.StartTime, search.EndTime))
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

func (u *CronjobService) CleanRecord(req dto.CronjobClean) error {
	cronjob, err := cronjobRepo.Get(commonRepo.WithByID(req.CronjobID))
	if err != nil {
		return err
	}
	if req.CleanData && (cronjob.Type == "database" || cronjob.Type == "website" || cronjob.Type == "directory") {
		cronjob.RetainCopies = 0
		backup, err := backupRepo.Get(commonRepo.WithByID(uint(cronjob.TargetDirID)))
		if err != nil {
			return err
		}
		if backup.Type != "LOCAL" {
			localDir, err := loadLocalDir()
			if err != nil {
				return err
			}
			client, err := NewIBackupService().NewClient(&backup)
			if err != nil {
				return err
			}
			u.HandleRmExpired(backup.Type, backup.BackupPath, localDir, &cronjob, client)
		} else {
			u.HandleRmExpired(backup.Type, backup.BackupPath, "", &cronjob, nil)
		}
	}
	delRecords, err := cronjobRepo.ListRecord(cronjobRepo.WithByJobID(int(req.CronjobID)))
	if err != nil {
		return err
	}
	for _, del := range delRecords {
		_ = os.RemoveAll(del.Records)
	}
	if err := cronjobRepo.DeleteRecord(cronjobRepo.WithByJobID(int(req.CronjobID))); err != nil {
		return err
	}
	return nil
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
	if backup.Type == "LOCAL" || record.FromLocal {
		if _, err := os.Stat(record.File); err != nil && os.IsNotExist(err) {
			return "", err
		}
		return record.File, nil
	}
	client, err := NewIBackupService().NewClient(&backup)
	if err != nil {
		return "", err
	}
	tempPath := fmt.Sprintf("%s/download/%s", constant.DataDir, record.File)
	_ = os.MkdirAll(path.Dir(tempPath), os.ModePerm)
	isOK, err := client.Download(record.File, tempPath)
	if !isOK || err != nil {
		return "", err
	}
	return tempPath, nil
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

	global.LOG.Infof("create cronjob %s successful, spec: %s", cronjob.Name, cronjob.Spec)
	entryID, err := u.StartJob(&cronjob)
	if err != nil {
		return err
	}
	cronjob.EntryID = uint64(entryID)
	if err := cronjobRepo.Create(&cronjob); err != nil {
		return err
	}
	return nil
}

func (u *CronjobService) StartJob(cronjob *model.Cronjob) (int, error) {
	if cronjob.EntryID != 0 {
		global.Cron.Remove(cron.EntryID(cronjob.EntryID))
	}
	entryID, err := u.AddCronJob(cronjob)
	if err != nil {
		return 0, err
	}
	return entryID, nil
}

func (u *CronjobService) Delete(req dto.CronjobBatchDelete) error {
	for _, id := range req.IDs {
		cronjob, _ := cronjobRepo.Get(commonRepo.WithByID(id))
		if cronjob.ID == 0 {
			return errors.New("find cronjob in db failed")
		}
		global.Cron.Remove(cron.EntryID(cronjob.EntryID))
		global.LOG.Infof("stop cronjob entryID: %d", cronjob.EntryID)
		if err := u.CleanRecord(dto.CronjobClean{CronjobID: id, CleanData: req.CleanData}); err != nil {
			return err
		}
		if err := cronjobRepo.Delete(commonRepo.WithByID(id)); err != nil {
			return err
		}
	}

	return nil
}

func (u *CronjobService) Update(id uint, req dto.CronjobUpdate) error {
	var cronjob model.Cronjob
	if err := copier.Copy(&cronjob, &req); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	cronModel, err := cronjobRepo.Get(commonRepo.WithByID(id))
	if err != nil {
		return constant.ErrRecordNotFound
	}
	upMap := make(map[string]interface{})
	cronjob.EntryID = cronModel.EntryID
	cronjob.Type = cronModel.Type
	cronjob.Spec = loadSpec(cronjob)
	if cronModel.Status == constant.StatusEnable {
		newEntryID, err := u.StartJob(&cronjob)
		if err != nil {
			return err
		}
		upMap["entry_id"] = newEntryID
	} else {
		global.Cron.Remove(cron.EntryID(cronjob.EntryID))
	}

	upMap["name"] = req.Name
	upMap["spec"] = cronjob.Spec
	upMap["script"] = req.Script
	upMap["container_name"] = req.ContainerName
	upMap["spec_type"] = req.SpecType
	upMap["week"] = req.Week
	upMap["day"] = req.Day
	upMap["hour"] = req.Hour
	upMap["minute"] = req.Minute
	upMap["second"] = req.Second
	upMap["website"] = req.Website
	upMap["exclusion_rules"] = req.ExclusionRules
	upMap["db_name"] = req.DBName
	upMap["url"] = req.URL
	upMap["source_dir"] = req.SourceDir
	upMap["keep_local"] = req.KeepLocal
	upMap["target_dir_id"] = req.TargetDirID
	upMap["retain_copies"] = req.RetainCopies
	return cronjobRepo.Update(id, upMap)
}

func (u *CronjobService) UpdateStatus(id uint, status string) error {
	cronjob, _ := cronjobRepo.Get(commonRepo.WithByID(id))
	if cronjob.ID == 0 {
		return errors.WithMessage(constant.ErrRecordNotFound, "record not found")
	}
	var (
		entryID int
		err     error
	)
	if status == constant.StatusEnable {
		entryID, err = u.StartJob(&cronjob)
		if err != nil {
			return err
		}
	} else {
		global.Cron.Remove(cron.EntryID(cronjob.EntryID))
		global.LOG.Infof("stop cronjob entryID: %d", cronjob.EntryID)
	}
	return cronjobRepo.Update(cronjob.ID, map[string]interface{}{"status": status, "entry_id": entryID})
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
	global.LOG.Infof("start cronjob entryID: %d", entryID)
	return int(entryID), nil
}

func mkdirAndWriteFile(cronjob *model.Cronjob, startTime time.Time, msg []byte) (string, error) {
	dir := fmt.Sprintf("%s/task/%s/%s", constant.DataDir, cronjob.Type, cronjob.Name)
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return "", err
		}
	}

	path := fmt.Sprintf("%s/%s.log", dir, startTime.Format("20060102150405"))
	global.LOG.Infof("cronjob %s has generated some logs %s", cronjob.Name, path)
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
	case "perDay":
		return fmt.Sprintf("%v %v * * *", cronjob.Minute, cronjob.Hour)
	case "perNHour":
		return fmt.Sprintf("%v */%v * * *", cronjob.Minute, cronjob.Hour)
	case "perHour":
		return fmt.Sprintf("%v * * * *", cronjob.Minute)
	case "perNMinute":
		return fmt.Sprintf("@every %vm", cronjob.Minute)
	case "perNSecond":
		return fmt.Sprintf("@every %vs", cronjob.Second)
	default:
		return ""
	}
}
