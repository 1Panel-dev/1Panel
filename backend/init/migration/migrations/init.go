package migrations

import (
	"fmt"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/common"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddTableOperationLog = &gormigrate.Migration{
	ID: "20200809-add-table-operation-log",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.OperationLog{}, &model.LoginLog{})
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
			Name: "default", Type: "host", IsDefault: true,
		}
		if err := tx.Create(&group).Error; err != nil {
			return err
		}
		host := model.Host{
			Name: "localhost", Addr: "127.0.0.1", User: "root", Port: 22, GroupID: group.ID,
		}
		if err := tx.Create(&host).Error; err != nil {
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
		if err := tx.Create(&model.Setting{Key: "UserName", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "Password", Value: ""}).Error; err != nil {
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
		if err := tx.Create(&model.Setting{Key: "Theme", Value: "light"}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "SessionTimeout", Value: "86400"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "LocalTime", Value: ""}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "ServerPort", Value: "9999"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SecurityEntrance", Value: "onepanel"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "JWTSigningKey", Value: common.RandStr(16)}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "EncryptKey", Value: common.RandStr(16)}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "ExpirationTime", Value: time.Now().AddDate(0, 0, 10).Format("2006-01-02 15:04:05")}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "ExpirationDays", Value: "0"}).Error; err != nil {
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
		if err := tx.Create(&model.Setting{Key: "SystemVersion", Value: "v1.0.0"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SystemStatus", Value: "Free"}).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Create(&model.Setting{Key: "AppStoreVersion", Value: ""}).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddTableBackupAccount = &gormigrate.Migration{
	ID: "20200916-add-table-backup",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.BackupAccount{}, &model.BackupRecord{}); err != nil {
			return err
		}

		item := &model.BackupAccount{
			Type: "LOCAL",
			Vars: fmt.Sprintf("{\"dir\":\"%s\"}", global.CONF.System.Backup),
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

var AddTableApp = &gormigrate.Migration{
	ID: "20200921-add-table-app",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.App{}, &model.AppDetail{}, &model.Tag{}, &model.AppTag{}, &model.AppInstall{}, &model.AppInstallResource{})
	},
}

var AddTableImageRepo = &gormigrate.Migration{
	ID: "20201009-add-table-imagerepo",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.ImageRepo{}, &model.ComposeTemplate{}, &model.Compose{}); err != nil {
			return err
		}
		item := &model.ImageRepo{
			Name:        "Docker Hub",
			Protocol:    "https",
			DownloadUrl: "docker.io",
			Status:      constant.StatusSuccess,
		}
		if err := tx.Create(item).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddTableDatabaseMysql = &gormigrate.Migration{
	ID: "20201020-add-table-database_mysql",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.DatabaseMysql{})
	},
}

var AddTableWebsite = &gormigrate.Migration{
	ID: "20201009-add-table-website",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Website{}, &model.WebsiteDomain{}, &model.WebsiteGroup{}, &model.WebsiteDnsAccount{}, &model.WebsiteSSL{}, &model.WebsiteAcmeAccount{}); err != nil {
			return err
		}
		item := &model.WebsiteGroup{
			Name:    "默认",
			Default: true,
		}
		if err := tx.Create(item).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddTableSnap = &gormigrate.Migration{
	ID: "20230106-add-table-snap",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Snapshot{}); err != nil {
			return err
		}
		return nil
	},
}
