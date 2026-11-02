package migrations

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddContainerImageCheckUpdates = &gormigrate.Migration{
	ID: "20261028-container-image-updates",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.ContainerImageCheckUpdate{}); err != nil {
			return err
		}
		return tx.Where("key = ?", "ContainerImageUpdateEnabled").FirstOrCreate(&model.Setting{Key: "ContainerImageUpdateEnabled", Value: "Disable"}).Error
	},
}
