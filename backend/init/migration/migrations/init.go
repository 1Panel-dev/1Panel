package migrations

import (
	"1Panel/app/model"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var InitTable = &gormigrate.Migration{
	ID: "20220803-init-table",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.User{})
	},
}

var user = model.User{
	Name: "admin", Email: "admin@fit2cloud.com", Password: "Calong@2015",
}

var AddData = &gormigrate.Migration{
	ID: "20200803-add-data",
	Migrate: func(tx *gorm.DB) error {
		return tx.Create(&user).Error
	},
}
