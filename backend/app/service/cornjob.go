package service

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type CronjobService struct{}

type ICronjobService interface {
	SearchWithPage(search dto.SearchWithPage) (int64, interface{}, error)
	Create(cronjobDto dto.CronjobCreate) error
	Update(id uint, upMap map[string]interface{}) error
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

func (u *CronjobService) Create(cronjobDto dto.CronjobCreate) error {
	cronjob, _ := cronjobRepo.Get(commonRepo.WithByName(cronjobDto.Name))
	if cronjob.ID != 0 {
		return constant.ErrRecordExist
	}
	if err := copier.Copy(&cronjob, &cronjobDto); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	switch cronjobDto.SpecType {
	case "perMonth":
		cronjob.Spec = fmt.Sprintf("%v %v %v * *", cronjobDto.Minute, cronjobDto.Hour, cronjobDto.Day)
	case "perWeek":
		cronjob.Spec = fmt.Sprintf("%v %v * * %v", cronjobDto.Minute, cronjobDto.Hour, cronjobDto.Week)
	case "perNDay":
		cronjob.Spec = fmt.Sprintf("%v %v */%v * *", cronjobDto.Minute, cronjobDto.Hour, cronjobDto.Day)
	case "perNHour":
		cronjob.Spec = fmt.Sprintf("%v */%v * * *", cronjobDto.Minute, cronjobDto.Hour)
	case "perHour":
		cronjob.Spec = fmt.Sprintf("%v * * * *", cronjobDto.Minute)
	case "perNMinute":
		cronjob.Spec = fmt.Sprintf("@every %vm", cronjobDto.Minute)
	}
	if err := cronjobRepo.Create(&cronjob); err != nil {
		return err
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

func (u *CronjobService) Update(id uint, upMap map[string]interface{}) error {
	return cronjobRepo.Update(id, upMap)
}
