package repo

import "github.com/1Panel-dev/1Panel/agent/app/model"

type AgentAccountModelRepo struct{}

type IAgentAccountModelRepo interface {
	List(opts ...DBOption) ([]model.AgentAccountModel, error)
	GetFirst(opts ...DBOption) (*model.AgentAccountModel, error)
	Create(item *model.AgentAccountModel) error
	Save(item *model.AgentAccountModel) error
	DeleteByID(id uint) error
	Delete(opts ...DBOption) error
}

func NewIAgentAccountModelRepo() IAgentAccountModelRepo {
	return &AgentAccountModelRepo{}
}

func (a AgentAccountModelRepo) List(opts ...DBOption) ([]model.AgentAccountModel, error) {
	var list []model.AgentAccountModel
	if err := getDb(opts...).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (a AgentAccountModelRepo) GetFirst(opts ...DBOption) (*model.AgentAccountModel, error) {
	var item model.AgentAccountModel
	if err := getDb(opts...).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (a AgentAccountModelRepo) Create(item *model.AgentAccountModel) error {
	return getDb().Create(item).Error
}

func (a AgentAccountModelRepo) Save(item *model.AgentAccountModel) error {
	return getDb().Save(item).Error
}

func (a AgentAccountModelRepo) DeleteByID(id uint) error {
	return getDb().Delete(&model.AgentAccountModel{}, id).Error
}

func (a AgentAccountModelRepo) Delete(opts ...DBOption) error {
	return getDb(opts...).Delete(&model.AgentAccountModel{}).Error
}
