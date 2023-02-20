package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
)

type WebsiteGroupService struct {
}

func (w WebsiteGroupService) CreateGroup(create request.WebsiteGroupCreate) error {
	groups, _ := websiteGroupRepo.GetBy(commonRepo.WithByName(create.Name))
	if len(groups) > 0 {
		return buserr.New(constant.ErrNameIsExist)
	}
	return websiteGroupRepo.Create(&model.WebsiteGroup{
		Name: create.Name,
	})
}

func (w WebsiteGroupService) GetGroups() ([]model.WebsiteGroup, error) {
	return websiteGroupRepo.GetBy()
}

func (w WebsiteGroupService) UpdateGroup(update request.WebsiteGroupUpdate) error {
	if update.Default {
		if err := websiteGroupRepo.CancelDefault(); err != nil {
			return err
		}
		return websiteGroupRepo.Save(&model.WebsiteGroup{
			BaseModel: model.BaseModel{
				ID: update.ID,
			},
			Name:    update.Name,
			Default: true,
		})
	} else {
		exists, _ := websiteGroupRepo.GetBy(commonRepo.WithByName(update.Name))
		for _, exist := range exists {
			if exist.ID != update.ID {
				return buserr.New(constant.ErrNameIsExist)
			}
		}
		return websiteGroupRepo.Save(&model.WebsiteGroup{
			BaseModel: model.BaseModel{
				ID: update.ID,
			},
			Name: update.Name,
		})
	}
}

func (w WebsiteGroupService) DeleteGroup(id uint) error {
	return websiteGroupRepo.DeleteBy(commonRepo.WithByID(id))
}
