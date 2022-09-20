package service

import (
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
