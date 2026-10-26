package migrations

import (
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddPanelMCP = &gormigrate.Migration{
	ID: "20261008-panel-mcp",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.MCPCredential{}, &model.MCPGrant{}, &model.MCPInvocation{}, &model.MCPSetting{})
	},
}
