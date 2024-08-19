package service

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	httpUtils "github.com/1Panel-dev/1Panel/core/utils/http"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
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
		commonRepo.WithByType(req.Type),
		commonRepo.WithOrderBy("is_default desc"),
		commonRepo.WithOrderBy("created_at desc"),
	}
	var (
		groups []model.Group
		err    error
	)
	groups, err = groupRepo.GetList(options...)
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
	if group.IsDefault {
		return buserr.New(constant.ErrGroupIsDefault)
	}
	defaultGroup, err := groupRepo.Get(commonRepo.WithByType(group.Type), groupRepo.WithByDefault(true))
	if err != nil {
		return err
	}
	switch group.Type {
	case "host":
		err = hostRepo.UpdateGroup(id, defaultGroup.ID)
	case "command":
		err = commandRepo.UpdateGroup(id, defaultGroup.ID)
	case "node":
		err = xpack.UpdateGroup("node", id, defaultGroup.ID)
	case "website":
		bodyItem := []byte(fmt.Sprintf(`{"Group":%v, "NewGroup":%v}`, id, defaultGroup.ID))
		if _, err := httpUtils.NewLocalClient("/api/v2/websites/group/change", http.MethodPost, bytes.NewReader(bodyItem)); err != nil {
			return err
		}
		if err := xpack.UpdateGroup("node", id, defaultGroup.ID); err != nil {
			return err
		}
	default:
		return constant.ErrNotSupportType
	}
	if err != nil {
		return err
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
