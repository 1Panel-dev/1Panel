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
		if err := tx.Create(&model.Setting{Key: "XpackHideMenu", Value: "{\n    \"id\": \"1\",\n    \"label\": \"/xpack\",\n    \"isCheck\": false,\n    \"title\": \"xpack.menu\",\n    \"children\": [\n        {\n            \"id\": \"2\",\n            \"title\": \"xpack.waf.name\",\n            \"path\": \"/xpack/waf/dashboard\",\n            \"label\": \"Dashboard\",\n            \"isCheck\": false\n        },\n        {\n            \"id\": \"3\",\n            \"title\": \"xpack.tamper.tamper\",\n            \"path\": \"/xpack/tamper\",\n            \"label\": \"Tamper\",\n            \"isCheck\": true\n        },\n        {\n            \"id\": \"4\",\n            \"title\": \"xpack.setting.setting\",\n            \"path\": \"/xpack/setting\",\n            \"label\": \"XSetting\",\n            \"isCheck\": true\n        }\n    ]\n}"}).Error; err != nil {
			return err
		}
		return nil
	},
}
