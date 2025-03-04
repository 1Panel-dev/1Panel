package repo

import (
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
)

type AgentRepo struct{}

type IAgentRepo interface {
	GetWebsiteSSL(opts ...global.DBOption) (model.WebsiteSSL, error)
	GetCA(opts ...global.DBOption) (model.WebsiteCA, error)
}

func NewIAgentRepo() IAgentRepo {
	return &AgentRepo{}
}

func (a *AgentRepo) GetWebsiteSSL(opts ...global.DBOption) (model.WebsiteSSL, error) {
	var ssl model.WebsiteSSL
	db := global.AgentDB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&ssl).Error
	return ssl, err
}

func (a *AgentRepo) GetCA(opts ...global.DBOption) (model.WebsiteCA, error) {
	var ca model.WebsiteCA
	db := global.AgentDB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&ca).Error
	return ca, err
}
