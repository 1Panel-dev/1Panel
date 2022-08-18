package service

import (
	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type HostService struct{}

type IHostService interface {
	GetConnInfo(id uint) (*model.Host, error)
	Page(search dto.PageInfo) (int64, interface{}, error)
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

func (u *HostService) Page(search dto.PageInfo) (int64, interface{}, error) {
	total, hosts, err := hostRepo.Page(search.Page, search.PageSize)
	var dtoHosts []dto.HostInfo
	for _, host := range hosts {
		var item dto.HostInfo
		if err := copier.Copy(&item, &host); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtoHosts = append(dtoHosts, item)
	}
	return total, dtoHosts, err
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
