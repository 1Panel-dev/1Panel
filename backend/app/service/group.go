package service

import (
	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type GroupService struct{}

type IGroupService interface {
	GetGroupInfo(id uint) (*dto.GroupInfo, error)
	List(req dto.GroupSearch) ([]dto.GroupInfo, error)
	Create(groupDto dto.GroupOperate) error
	Update(id uint, name string) error
	Delete(id uint) error
}

func NewIGroupService() IGroupService {
	return &GroupService{}
}

func (u *GroupService) GetGroupInfo(id uint) (*dto.GroupInfo, error) {
	group, err := groupRepo.Get(commonRepo.WithByID(id))
	if err != nil {
		return nil, constant.ErrRecordNotFound
	}
	var dtoGroup dto.GroupInfo
	if err := copier.Copy(&dtoGroup, &group); err != nil {
		return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	return &dtoGroup, err
}

func (u *GroupService) List(req dto.GroupSearch) ([]dto.GroupInfo, error) {
	groups, err := groupRepo.GetList(groupRepo.WithByType(req.Type))
	if err != nil {
		return nil, constant.ErrRecordNotFound
	}
	var dtoUsers []dto.GroupInfo
	for _, group := range groups {
		var item dto.GroupInfo
		if err := copier.Copy(&item, &group); err != nil {
			return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtoUsers = append(dtoUsers, item)
	}
	return dtoUsers, err
}

func (u *GroupService) Create(groupDto dto.GroupOperate) error {
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

func (u *GroupService) Delete(id uint) error {
	group, _ := groupRepo.Get(commonRepo.WithByID(id))
	if group.ID == 0 {
		return constant.ErrRecordNotFound
	}
	if err := hostRepo.ChangeGroup(group.Name, "default"); err != nil {
		return err
	}
	return groupRepo.Delete(commonRepo.WithByID(id))
}

func (u *GroupService) Update(id uint, name string) error {
	group, _ := groupRepo.Get(commonRepo.WithByID(id))
	if group.ID == 0 {
		return constant.ErrRecordNotFound
	}

	upMap := make(map[string]interface{})
	upMap["name"] = name
	if err := hostRepo.ChangeGroup(group.Name, name); err != nil {
		return err
	}
	return groupRepo.Update(id, upMap)
}
