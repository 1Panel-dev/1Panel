package repo

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
)

type GPUGroupRepo struct{}

func (r *GPUGroupRepo) List() ([]model.GPUGroup, error) {
	groups := []model.GPUGroup{}
	err := global.DB.Order("id desc").Find(&groups).Error
	return groups, err
}

func (r *GPUGroupRepo) Save(group *model.GPUGroup) error {
	return global.DB.Save(group).Error
}

func (r *GPUGroupRepo) Delete(id uint) error {
	return global.DB.Delete(&model.GPUGroup{}, id).Error
}

func (r *GPUGroupRepo) Deployments() ([]model.AppInstall, error) {
	installs := []model.AppInstall{}
	err := global.DB.Select("id", "name", "docker_compose").Where("docker_compose LIKE ?", "%x-1panel-gpu-group%").Find(&installs).Error
	return installs, err
}
