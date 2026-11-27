package migrations

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/provider"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddAgentAccountType = &gormigrate.Migration{
	ID: "202611120002",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.AgentAccount{}); err != nil {
			return err
		}
		for key := range provider.All() {
			preset, ok := provider.FindAccountType(key, "")
			if !ok {
				continue
			}
			if err := tx.Model(&model.AgentAccount{}).Where("provider = ? AND (account_type = ? OR account_type IS NULL)", key, "").Update("account_type", preset.Key).Error; err != nil {
				return err
			}
		}
		return nil
	},
}
