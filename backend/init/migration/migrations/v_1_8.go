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
