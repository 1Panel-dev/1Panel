package service

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type HostService struct{}

type IHostService interface {
	GetHostInfo(id uint) (*model.Host, error)
	SearchForTree(search dto.SearchForTree) ([]dto.HostTree, error)
	Create(hostDto dto.HostOperate) (*dto.HostInfo, error)
	Update(id uint, upMap map[string]interface{}) error
	Delete(id uint) error
}

func NewIHostService() IHostService {
	return &HostService{}
}

func (u *HostService) GetHostInfo(id uint) (*model.Host, error) {
	host, err := hostRepo.Get(commonRepo.WithByID(id))
	if err != nil {
		return nil, constant.ErrRecordNotFound
	}
	return &host, err
}

func (u *HostService) SearchForTree(search dto.SearchForTree) ([]dto.HostTree, error) {
	hosts, err := hostRepo.GetList(hostRepo.WithByInfo(search.Info))
	if err != nil {
		return nil, err
	}
	groups, err := groupRepo.GetList()
	if err != nil {
		return nil, err
	}
	var datas []dto.HostTree
	for _, group := range groups {
		var data dto.HostTree
		data.ID = group.ID + 10000
		data.Label = group.Name
		for _, host := range hosts {
			label := fmt.Sprintf("%s@%s:%d", host.User, host.Addr, host.Port)
			if len(host.Name) != 0 {
				label = fmt.Sprintf("%s-%s@%s:%d", host.Name, host.User, host.Addr, host.Port)
			}
			if host.GroupBelong == group.Name {
				data.Children = append(data.Children, dto.TreeChild{ID: host.ID, Label: label})
			}
		}
		datas = append(datas, data)
	}
	return datas, err
}

func (u *HostService) Create(req dto.HostOperate) (*dto.HostInfo, error) {
	host, _ := hostRepo.Get(commonRepo.WithByName(req.Name))
	if host.ID != 0 {
		return nil, constant.ErrRecordExist
	}
	if err := copier.Copy(&host, &req); err != nil {
		return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	var sameHostID uint
	if req.Addr == "127.0.0.1" {
		hostSame, _ := hostRepo.Get(hostRepo.WithByAddr(req.Addr))
		sameHostID = hostSame.ID
	} else {
		hostSame, _ := hostRepo.Get(hostRepo.WithByAddr(req.Addr), hostRepo.WithByUser(req.User), hostRepo.WithByPort(req.Port))
		sameHostID = hostSame.ID
	}
	if sameHostID != 0 {
		host.ID = sameHostID
		upMap := make(map[string]interface{})
		upMap["name"] = req.Name
		upMap["group_belong"] = req.GroupBelong
		upMap["addr"] = req.Addr
		upMap["port"] = req.Port
		upMap["user"] = req.User
		upMap["auth_mode"] = req.AuthMode
		upMap["password"] = req.Password
		upMap["private_key"] = req.PrivateKey
		upMap["description"] = req.Description
		if err := hostRepo.Update(sameHostID, upMap); err != nil {
			return nil, err
		}
		var hostinfo dto.HostInfo
		if err := copier.Copy(&hostinfo, &host); err != nil {
			return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		return &hostinfo, nil
	}

	if err := hostRepo.Create(&host); err != nil {
		return nil, err
	}
	var hostinfo dto.HostInfo
	if err := copier.Copy(&hostinfo, &host); err != nil {
		return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	return &hostinfo, nil
}

func (u *HostService) Delete(id uint) error {
	host, _ := hostRepo.Get(commonRepo.WithByID(id))
	if host.ID == 0 {
		return constant.ErrRecordNotFound
	}
	return hostRepo.Delete(commonRepo.WithByID(id))
}

func (u *HostService) Update(id uint, upMap map[string]interface{}) error {
	return hostRepo.Update(id, upMap)
}
