package migrations

import (
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddAPIKeys = &gormigrate.Migration{
	ID:      "20260915-api-keys",
	Migrate: func(tx *gorm.DB) error { return tx.AutoMigrate(&model.APIKey{}, &model.LegacyAPIKeyPolicy{}) },
}

var AddOperationLogAPIKey = &gormigrate.Migration{
	ID:      "20260915-add-operation-log-api-key",
	Migrate: func(tx *gorm.DB) error { return tx.AutoMigrate(&model.OperationLog{}) },
}
