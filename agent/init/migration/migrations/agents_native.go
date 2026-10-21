package migrations

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddNativeAgentInstalls = &gormigrate.Migration{
	ID: "20261020-add-deepseek-harness-octop-agent-installs",
	Migrate: func(tx *gorm.DB) error {
		var installs []struct {
			model.AppInstall
			AgentType string
		}
		if err := tx.Table("app_installs").
			Select("app_installs.*, apps.key AS agent_type").
			Joins("JOIN apps ON apps.id = app_installs.app_id").
			Where("apps.key IN ?", []string{constant.AppDeepSeekHarness, constant.AppOctop}).
			Where("NOT EXISTS (SELECT 1 FROM agents WHERE agents.app_install_id = app_installs.id)").
			Order("app_installs.id").
			Scan(&installs).Error; err != nil {
			return err
		}
		for _, install := range installs {
			var count int64
			if err := tx.Model(&model.Agent{}).Where("LOWER(name) = LOWER(?)", install.Name).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				global.LOG.Warnf("skip agent registration for app install %d: name already exists", install.ID)
				continue
			}
			agent := model.Agent{
				BaseModel: model.BaseModel{CreatedAt: install.CreatedAt, UpdatedAt: install.UpdatedAt},
				Name:      install.Name, AgentType: install.AgentType, AppInstallID: install.ID,
				Status: install.Status, Message: install.Message,
			}
			if err := tx.Create(&agent).Error; err != nil {
				return err
			}
		}
		return nil
	},
}
