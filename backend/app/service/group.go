package service

import (
	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type GroupService struct{}

type IGroupService interface {
	Search() ([]model.Group, error)
	Create(groupDto dto.GroupCreate) error
	Update(id uint, upMap map[string]interface{}) error
	Delete(name string) error
}

func NewIGroupService() IGroupService {
	return &GroupService{}
}

func (u *GroupService) Search() ([]model.Group, error) {
	groups, err := groupRepo.GetList()
	if err != nil {
		return nil, constant.ErrRecordNotFound
	}
	return groups, err
}

func (u *GroupService) Create(groupDto dto.GroupCreate) error {
	group, _ := groupRepo.Get(commonRepo.WithByName(groupDto.Name), commonRepo.WithByName(groupDto.Name))
	if group.ID != 0 {
		return constant.ErrRecordExist
	}
	if err := copier.Copy(&group, &groupDto); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	if err := groupRepo.Create(&group); err != nil {
		return err
	}
	return nil
}

func (u *GroupService) Delete(name string) error {
	group, _ := groupRepo.Get(commonRepo.WithByName(name))
	if group.ID == 0 {
		return constant.ErrRecordNotFound
	}
	return groupRepo.Delete(commonRepo.WithByID(group.ID))
}

func (u *GroupService) Update(id uint, upMap map[string]interface{}) error {
	return groupRepo.Update(id, upMap)
}
