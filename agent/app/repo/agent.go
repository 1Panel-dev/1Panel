package repo

import (
	"context"

	"github.com/1Panel-dev/1Panel/agent/app/model"
)

type AgentRepo struct{}

type IAgentRepo interface {
	Page(page, size int, opts ...DBOption) (int64, []model.Agent, error)
	GetFirst(opts ...DBOption) (*model.Agent, error)
	Create(agent *model.Agent) error
	Save(agent *model.Agent) error
	DeleteByID(id uint) error
	DeleteByAppInstallID(appInstallID uint) error
	DeleteByAppInstallIDWithCtx(ctx context.Context, appInstallID uint) error
	List(opts ...DBOption) ([]model.Agent, error)
}

func NewIAgentRepo() IAgentRepo {
	return &AgentRepo{}
}

func (a AgentRepo) Page(page, size int, opts ...DBOption) (int64, []model.Agent, error) {
	var agents []model.Agent
	db := getDb(opts...).Model(&model.Agent{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&agents).Error
	return count, agents, err
}

func (a AgentRepo) GetFirst(opts ...DBOption) (*model.Agent, error) {
	var agent model.Agent
	if err := getDb(opts...).First(&agent).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (a AgentRepo) Create(agent *model.Agent) error {
	return getDb().Create(agent).Error
}

func (a AgentRepo) Save(agent *model.Agent) error {
	return getDb().Save(agent).Error
}

func (a AgentRepo) DeleteByID(id uint) error {
	return getDb().Delete(&model.Agent{}, id).Error
}

func (a AgentRepo) DeleteByAppInstallID(appInstallID uint) error {
	if appInstallID == 0 {
		return nil
	}
	return getDb().Where("app_install_id = ?", appInstallID).Delete(&model.Agent{}).Error
}

func (a AgentRepo) DeleteByAppInstallIDWithCtx(ctx context.Context, appInstallID uint) error {
	if appInstallID == 0 {
		return nil
	}
	return getTx(ctx).Where("app_install_id = ?", appInstallID).Delete(&model.Agent{}).Error
}

func (a AgentRepo) List(opts ...DBOption) ([]model.Agent, error) {
	var agents []model.Agent
	if err := getDb(opts...).Find(&agents).Error; err != nil {
		return nil, err
	}
	return agents, nil
}
