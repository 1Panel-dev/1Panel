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
