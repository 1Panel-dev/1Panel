package migrations

import (
	"github.com/1Panel-dev/1Panel/app/model"

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
	Name: "admin", Email: "admin@fit2cloud.com", Password: "5WYEZ4XcitdomVvAyimt9WwJwBJJSbTTHncZoqyOraQ=",
}

var AddData = &gormigrate.Migration{
	ID: "20200803-add-data",
	Migrate: func(tx *gorm.DB) error {
		return tx.Create(&user).Error
	},
}

var AddTableOperationLog = &gormigrate.Migration{
	ID: "20200809-add-table-operation-log",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.OperationLog{})
	},
}

var AddTableHost = &gormigrate.Migration{
	ID: "20200818-add-table-host",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Host{}); err != nil {
			return err
		}
		if err := tx.AutoMigrate(&model.Group{}); err != nil {
			return err
		}
		if err := tx.AutoMigrate(&model.Command{}); err != nil {
			return err
		}
		group := model.Group{
			Name: "default", Type: "host",
		}
		if err := tx.Create(&group).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddTablemonitor = &gormigrate.Migration{
	ID: "20200905-add-table-monitor",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.MonitorBase{}, &model.MonitorIO{}, &model.MonitorNetwork{})
	},
}
