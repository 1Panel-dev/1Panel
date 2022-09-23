package service

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
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
	switch cronjobDto.Type {
	case "shell":
		entryID, err := u.AddShellJob(&cronjob)
		if err != nil {
			return err
		}
		_ = cronjobRepo.Update(cronjob.ID, map[string]interface{}{"entry_id": entryID})
	}
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
			cronjobRepo.EndRecords(record, constant.StatusFailed, err.Error(), "ERR_GENERAGE_STDOUT")
			return
		}
		record.Records, err = mkdirAndWriteFile(cronjob.ID, cronjob.Name, record.StartTime, stdout)
		if err != nil {
			record.Records = "ERR_CREATE_FILE"
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

func mkdirAndWriteFile(id uint, name string, startTime time.Time, msg []byte) (string, error) {
	dir := fmt.Sprintf("/opt/1Panel/data/cron/%s%v", name, id)
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return "", err
		}
	}

	path := fmt.Sprintf("%s/%s", dir, startTime.Format("20060102150405"))
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
