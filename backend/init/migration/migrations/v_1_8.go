package migrations

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddFavorite = &gormigrate.Migration{
	ID: "20231020-add-favorite",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Favorite{}); err != nil {
			return err
		}
		return nil
	},
}

var AddBindAddress = &gormigrate.Migration{
	ID: "20231024-add-bind-address",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "BindAddress", Value: "0.0.0.0"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "Ipv6", Value: "disable"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddCommandGroup = &gormigrate.Migration{
	ID: "20231030-add-command-group",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Command{}); err != nil {
			return err
		}
		defaultCommand := &model.Group{IsDefault: true, Name: "Default", Type: "command"}
		if err := tx.Create(defaultCommand).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Command{}).Where("1 = 1").Update("group_id", defaultCommand.ID).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddAppSyncStatus = &gormigrate.Migration{
	ID: "20231103-update-table-setting",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "AppStoreSyncStatus", Value: "SyncSuccess"}).Error; err != nil {
			return err
		}
		return nil
	},
}
