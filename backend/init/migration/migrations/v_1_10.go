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

var AddDatabaseIsDelete = &gormigrate.Migration{
	ID: "20240314-add-database-is-delete",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.DatabaseMysql{}, &model.DatabasePostgresql{}); err != nil {
			return err
		}
		return nil
	},
}

var AddXpackHideMenu = &gormigrate.Migration{
	ID: "20240328-add-xpack-hide-menu",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "XpackHideMenu", Value: "{\"id\":\"1\",\"label\":\"/xpack\",\"isCheck\":true,\"title\":\"xpack.menu\",\"children\":[{\"id\":\"2\",\"title\":\"xpack.waf.name\",\"path\":\"/xpack/waf/dashboard\",\"label\":\"Dashboard\",\"isCheck\":true},{\"id\":\"3\",\"title\":\"xpack.tamper.tamper\",\"path\":\"/xpack/tamper\",\"label\":\"Tamper\",\"isCheck\":true},{\"id\":\"4\",\"title\":\"xpack.setting.setting\",\"path\":\"/xpack/setting\",\"label\":\"XSetting\",\"isCheck\":true}]}"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddCronjobCommand = &gormigrate.Migration{
	ID: "20240403-add-cronjob-command",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Cronjob{}); err != nil {
			return err
		}
		return nil
	},
}