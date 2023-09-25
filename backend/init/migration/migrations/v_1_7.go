package migrations

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddDefaultNetwork = &gormigrate.Migration{
	ID: "20230918-add-default-network",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "DefaultNetwork", Value: ""}).Error; err != nil {
			return err
		}
		return nil
	},
}

var UpdateRuntime = &gormigrate.Migration{
	ID: "20230920-update-runtime",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Runtime{}); err != nil {
			return err
		}
		return nil
	},
}
