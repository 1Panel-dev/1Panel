package service

import (
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/jinzhu/copier"
)

type ScriptService struct{}

type IScriptService interface {
	Search(req dto.SearchPageWithGroup) (int64, interface{}, error)
	Create(req dto.ScriptOperate) error
	Update(req dto.ScriptOperate) error
	Delete(ids dto.OperateByIDs) error
}

func NewIScriptService() IScriptService {
	return &ScriptService{}
}

func (u *ScriptService) Search(req dto.SearchPageWithGroup) (int64, interface{}, error) {
	options := []global.DBOption{repo.WithOrderBy("created_at desc")}
	if len(req.Info) != 0 {
		options = append(options, scriptRepo.WithByInfo(req.Info))
	}
	list, err := scriptRepo.GetList(options...)
	if err != nil {
		return 0, nil, err
	}
	groups, _ := groupRepo.GetList(repo.WithByType("script"))
	groupMap := make(map[uint]string)
	for _, item := range groups {
		groupMap[item.ID] = item.Name
	}
	var data []dto.ScriptInfo
	for _, itemData := range list {
		var item dto.ScriptInfo
		if err := copier.Copy(&item, &itemData); err != nil {
			global.LOG.Errorf("copy backup account to dto backup info failed, err: %v", err)
		}
		matchGourp := false
		groupIDs := strings.Split(itemData.Groups, ",")
		for _, idItem := range groupIDs {
			id, _ := strconv.Atoi(idItem)
			if id == 0 {
				continue
			}
			if uint(id) == req.GroupID {
				matchGourp = true
			}
			item.GroupList = append(item.GroupList, uint(id))
			item.GroupBelong = append(item.GroupBelong, groupMap[uint(id)])
		}
		if req.GroupID == 0 {
			data = append(data, item)
			continue
		}
		if matchGourp {
			data = append(data, item)
		}
	}
	var records []dto.ScriptInfo
	total, start, end := len(data), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		records = make([]dto.ScriptInfo, 0)
	} else {
		if end >= total {
			end = total
		}
		records = data[start:end]
	}
	return int64(total), records, nil
}

func (u *ScriptService) Create(req dto.ScriptOperate) error {
	itemData, _ := scriptRepo.Get(repo.WithByName(req.Name))
	if itemData.ID != 0 {
		return buserr.New("ErrRecordExist")
	}
	if err := copier.Copy(&itemData, &req); err != nil {
		return buserr.WithDetail("ErrStructTransform", err.Error(), nil)
	}
	if err := scriptRepo.Create(&itemData); err != nil {
		return err
	}
	return nil
}

func (u *ScriptService) Delete(req dto.OperateByIDs) error {
	for _, item := range req.IDs {
		scriptItem, _ := scriptRepo.Get(repo.WithByID(item))
		if scriptItem.ID == 0 || scriptItem.IsSystem {
			continue
		}
		if err := scriptRepo.Delete(repo.WithByID(item)); err != nil {
			return err
		}
	}
	return nil
}

func (u *ScriptService) Update(req dto.ScriptOperate) error {
	itemData, _ := scriptRepo.Get(repo.WithByID(req.ID))
	if itemData.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	updateMap := make(map[string]interface{})
	updateMap["name"] = req.Name
	updateMap["script"] = req.Script
	updateMap["groups"] = req.Groups
	updateMap["description"] = req.Description
	if err := scriptRepo.Update(req.ID, updateMap); err != nil {
		return err
	}
	return nil
}

func LoadScriptInfo(id uint) (model.ScriptLibrary, error) {
	return scriptRepo.Get(repo.WithByID(id))
}
