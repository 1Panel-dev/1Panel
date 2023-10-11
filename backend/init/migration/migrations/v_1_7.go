package migrations

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddDefaultNetwork = &gormigrate.Migration{
	ID: "20230928-add-default-network",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "DefaultNetwork", Value: "all"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "LastCleanTime", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "LastCleanSize", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "LastCleanData", Value: ""}).Error; err != nil {
			return err
		}
		return nil
	},
}

var UpdateRuntime = &gormigrate.Migration{
	ID: "20230927-update-runtime",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Runtime{}); err != nil {
			return err
		}
		return nil
	},
}

var UpdateTag = &gormigrate.Migration{
	ID: "20231008-update-tag",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Tag{}); err != nil {
			return err
		}
		return nil
	},
}
