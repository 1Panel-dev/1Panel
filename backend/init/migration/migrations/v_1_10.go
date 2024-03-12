package migrations

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddSnapshotIgnore = &gormigrate.Migration{
	ID: "20240311-add-snapshot-ignore",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "SnapshotIgnore", Value: "*.sock"}).Error; err != nil {
			return err
		}
		return nil
	},
}
