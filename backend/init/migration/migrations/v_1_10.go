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
		if err := tx.Create(&model.Setting{Key: "XpackHideMenu", Value: "{\n\t\"id\": \"1\",\n\t\"label\": \"/xpack\",\n\t\"isCheck\": false,\n\t\"title\": \"xpack.name\",\n\t\"children\": [{\n\t\t\t\"id\": \"2\",\n\t\t\t\"title\": \"xpack.waf.name\",\n\t\t\t\"path\": \"/xpack/waf\",\n\t\t\t\"label\": \"WAF\",\n\t\t\t\"isCheck\": true\n\t\t},\n\t\t{\n\t\t\t\"id\": \"3\",\n\t\t\t\"title\": \"xpack.monitor.name\",\n\t\t\t\"path\": \"/xpack/monitor\",\n\t\t\t\"label\": \"webMonitor\",\n\t\t\t\"isCheck\": true\n\t\t},\n\t\t{\n\t\t\t\"id\": \"4\",\n\t\t\t\"title\": \"xpack.tamper.tamper\",\n\t\t\t\"path\": \"/xpack/tamper\",\n\t\t\t\"label\": \"Tamper\",\n\t\t\t\"isCheck\": false\n\t\t}\n\t]\n}"}).Error; err != nil {
			return err
		}
		return nil
	},
}
