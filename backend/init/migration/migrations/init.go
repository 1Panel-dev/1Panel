package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var InitTable = &gormigrate.Migration{
	ID: "20220803-init-table",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&entity.User{})
	},
}

var user = entity.User{
	Name: "admin", Email: "admin@fit2cloud.com", NickName: "admin", Password: "Calong@2015",
}

var AddData = &gormigrate.Migration{
	ID: "20200803-add-data",
	Migrate: func(tx *gorm.DB) error {
		return tx.Create(&user).Error
	},
}
