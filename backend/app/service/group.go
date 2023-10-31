package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type GroupService struct{}

type IGroupService interface {
	List(req dto.GroupSearch) ([]dto.GroupInfo, error)
	Create(req dto.GroupCreate) error
	Update(req dto.GroupUpdate) error
	Delete(id uint) error
}

func NewIGroupService() IGroupService {
	return &GroupService{}
}

func (u *GroupService) List(req dto.GroupSearch) ([]dto.GroupInfo, error) {
	groups, err := groupRepo.GetList(commonRepo.WithByType(req.Type), commonRepo.WithOrderBy("is_default desc"), commonRepo.WithOrderBy("created_at desc"))
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

func (u *GroupService) Create(req dto.GroupCreate) error {
	group, _ := groupRepo.Get(commonRepo.WithByName(req.Name), commonRepo.WithByType(req.Type))
	if group.ID != 0 {
		return constant.ErrRecordExist
	}
	if err := copier.Copy(&group, &req); err != nil {
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
	switch group.Type {
	case "website":
		websites, _ := websiteRepo.GetBy(websiteRepo.WithGroupID(id))
		if len(websites) > 0 {
			return buserr.New(constant.ErrGroupIsUsed)
		}
	case "command":
		commands, _ := commandRepo.GetList(commonRepo.WithByGroupID(id))
		if len(commands) > 0 {
			return buserr.New(constant.ErrGroupIsUsed)
		}
	case "host":
		hosts, _ := hostRepo.GetList(commonRepo.WithByGroupID(id))
		if len(hosts) > 0 {
			return buserr.New(constant.ErrGroupIsUsed)
		}
	}
	return groupRepo.Delete(commonRepo.WithByID(id))
}

func (u *GroupService) Update(req dto.GroupUpdate) error {
	if req.IsDefault {
		if err := groupRepo.CancelDefault(req.Type); err != nil {
			return err
		}
	}
	upMap := make(map[string]interface{})
	upMap["name"] = req.Name
	upMap["is_default"] = req.IsDefault

	return groupRepo.Update(req.ID, upMap)
}
