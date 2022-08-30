package service

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type HostService struct{}

type IHostService interface {
	GetConnInfo(id uint) (*model.Host, error)
	SearchForTree(search dto.SearchForTree) ([]dto.HostTree, error)
	Create(hostDto dto.HostCreate) (*dto.HostInfo, error)
	Update(id uint, upMap map[string]interface{}) error
	BatchDelete(ids []uint) error
}

func NewIHostService() IHostService {
	return &HostService{}
}

func (u *HostService) GetConnInfo(id uint) (*model.Host, error) {
	host, err := hostRepo.Get(commonRepo.WithByID(id))
	if err != nil {
		return nil, constant.ErrRecordNotFound
	}
	return &host, err
}

func (u *HostService) SearchForTree(search dto.SearchForTree) ([]dto.HostTree, error) {
	hosts, err := hostRepo.GetList(hostRepo.WithByInfo(search.Info))
	distinctMap := make(map[string][]string)
	for _, host := range hosts {
		if _, ok := distinctMap[host.Group]; !ok {
			distinctMap[host.Group] = []string{fmt.Sprintf("%s@%s:%d", host.User, host.Addr, host.Port)}
		} else {
			distinctMap[host.Group] = append(distinctMap[host.Group], fmt.Sprintf("%s@%s:%d", host.User, host.Addr, host.Port))
		}
	}
	var data []dto.HostTree
	for key, value := range distinctMap {
		var children []dto.TreeChild
		for _, label := range value {
			children = append(children, dto.TreeChild{Label: label})
		}
		data = append(data, dto.HostTree{Label: key, Children: children})
	}
	return data, err
}

func (u *HostService) Create(hostDto dto.HostCreate) (*dto.HostInfo, error) {
	host, _ := hostRepo.Get(commonRepo.WithByName(hostDto.Name))
	if host.ID != 0 {
		return nil, constant.ErrRecordExist
	}
	if err := copier.Copy(&host, &hostDto); err != nil {
		return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
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

func (u *HostService) BatchDelete(ids []uint) error {
	if len(ids) == 1 {
		host, _ := hostRepo.Get(commonRepo.WithByID(ids[0]))
		if host.ID == 0 {
			return constant.ErrRecordNotFound
		}
	}
	return hostRepo.Delete(commonRepo.WithIdsIn(ids))
}

func (u *HostService) Update(id uint, upMap map[string]interface{}) error {
	return hostRepo.Update(id, upMap)
}
