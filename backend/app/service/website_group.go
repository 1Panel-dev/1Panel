package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/model"
)

type WebsiteGroupService struct {
}

func (w WebsiteGroupService) CreateGroup(create request.WebsiteGroupCreate) error {
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
