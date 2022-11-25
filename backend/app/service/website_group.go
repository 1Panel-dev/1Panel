package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
)

type WebsiteGroupService struct {
}

func (w WebsiteGroupService) CreateGroup(create dto.WebSiteGroupCreate) error {
	return websiteGroupRepo.Create(&model.WebSiteGroup{
		Name: create.Name,
	})
}

func (w WebsiteGroupService) GetGroups() ([]model.WebSiteGroup, error) {
	return websiteGroupRepo.GetBy()
}

func (w WebsiteGroupService) UpdateGroup(update dto.WebSiteGroupUpdate) error {

	if update.Default {
		if err := websiteGroupRepo.CancelDefault(); err != nil {
			return err
		}
		return websiteGroupRepo.Save(&model.WebSiteGroup{
			BaseModel: model.BaseModel{
				ID: update.ID,
			},
			Name:    update.Name,
			Default: true,
		})
	} else {
		return websiteGroupRepo.Save(&model.WebSiteGroup{
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
