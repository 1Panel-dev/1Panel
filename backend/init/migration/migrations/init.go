package migrations

import (
	"time"

	"github.com/1Panel-dev/1Panel/app/model"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

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

var AddTableMonitor = &gormigrate.Migration{
	ID: "20200905-add-table-monitor",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.MonitorBase{}, &model.MonitorIO{}, &model.MonitorNetwork{})
	},
}

var AddTableSetting = &gormigrate.Migration{
	ID: "20200908-add-table-setting",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Setting{}); err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "UserName", Value: "admin"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "Password", Value: "5WYEZ4XcitdomVvAyimt9WwJwBJJSbTTHncZoqyOraQ="}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "Email", Value: ""}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "PanelName", Value: "1Panel"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "Language", Value: "zh"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "Theme", Value: "auto"}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "SessionTimeout", Value: "86400"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "LocalTime", Value: ""}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "ServerPort", Value: "4004"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SecurityEntrance", Value: "onepanel"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "PasswordTimeOut", Value: time.Now().AddDate(0, 0, 10).Format("2016.01.02 15:04:05")}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "ComplexityVerification", Value: "enable"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "MFAStatus", Value: "disable"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "MFASecret", Value: ""}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "MonitorStatus", Value: "enable"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "MonitorStoreDays", Value: "30"}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "MessageType", Value: "none"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "EmailVars", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "WeChatVars", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "DingVars", Value: ""}).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddTableBackupAccount = &gormigrate.Migration{
	ID: "20200916-add-table-backup",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.BackupAccount{}); err != nil {
			return err
		}
		item := &model.BackupAccount{
			Type: "LOCAL",
			Vars: "{\"dir\":\"/opt/1Panel/backup\"}",
		}
		if err := tx.Create(item).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddTableCronjob = &gormigrate.Migration{
	ID: "20200921-add-table-cronjob",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.Cronjob{}, &model.JobRecords{})
	},
}
