package service

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/jinzhu/copier"
)

type GroupService struct{}

type IGroupService interface {
	List(req dto.OperateByType) ([]dto.GroupInfo, error)
	Create(req dto.GroupCreate) error
	Update(req dto.GroupUpdate) error
	Delete(id uint) error
}

func NewIGroupService() IGroupService {
	return &GroupService{}
}

func (u *GroupService) List(req dto.OperateByType) ([]dto.GroupInfo, error) {
	options := []repo.DBOption{
		repo.WithOrderBy("is_default desc"),
		repo.WithOrderBy("created_at desc"),
	}
	if len(req.Type) != 0 {
		options = append(options, repo.WithByType(req.Type))
	}
	var (
		groups []model.Group
		err    error
	)
	groups, err = groupRepo.GetList(options...)
	if err != nil {
		return nil, buserr.New("ErrRecordNotFound")
	}
	var dtoUsers []dto.GroupInfo
	for _, group := range groups {
		var item dto.GroupInfo
		if err := copier.Copy(&item, &group); err != nil {
			return nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
		}
		dtoUsers = append(dtoUsers, item)
	}
	return dtoUsers, err
}

func (u *GroupService) Create(req dto.GroupCreate) error {
	group, _ := groupRepo.Get(repo.WithByName(req.Name), repo.WithByType(req.Type))
	if group.ID != 0 {
		return buserr.New("ErrRecordExist")
	}
	if err := copier.Copy(&group, &req); err != nil {
		return buserr.WithDetail("ErrStructTransform", err.Error(), nil)
	}
	if err := groupRepo.Create(&group); err != nil {
		return err
	}
	return nil
}

func (u *GroupService) Delete(id uint) error {
	group, _ := groupRepo.Get(repo.WithByID(id))
	if group.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	if group.IsDefault {
		return buserr.New("ErrGroupIsDefault")
	}
	if group.Type == "website" {
		websites, _ := websiteRepo.List(websiteRepo.WithGroupID(group.ID))
		if len(websites) > 0 {
			return buserr.New("ErrGroupIsInWebsiteUse")
		}
	}
	return groupRepo.Delete(repo.WithByID(id))
}

func (u *GroupService) Update(req dto.GroupUpdate) error {
	upMap := make(map[string]interface{})
	upMap["name"] = req.Name
	upMap["is_default"] = req.IsDefault
	return groupRepo.Update(req.ID, upMap)
}
