package migrations

import (
	"fmt"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"

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
			Name: "localhost", Addr: "127.0.0.1", User: "root", Port: 22, AuthMode: "password", GroupID: group.ID,
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
		encryptKey := common.RandStr(16)
		if err := tx.Create(&model.Setting{Key: "UserName", Value: global.CONF.System.Username}).Error; err != nil {
			return err
		}
		global.CONF.System.EncryptKey = encryptKey
		pass, _ := encrypt.StringEncrypt(global.CONF.System.Password)
		if err := tx.Create(&model.Setting{Key: "Password", Value: pass}).Error; err != nil {
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

		if err := tx.Create(&model.Setting{Key: "ServerPort", Value: global.CONF.System.Port}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SecurityEntrance", Value: global.CONF.System.Entrance}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "JWTSigningKey", Value: common.RandStr(16)}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "EncryptKey", Value: encryptKey}).Error; err != nil {
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
		if err := tx.Create(&model.Setting{Key: "SystemVersion", Value: global.CONF.System.Version}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SystemStatus", Value: "Free"}).Error; err != nil {
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
		if err := tx.AutoMigrate(&model.Website{}, &model.WebsiteDomain{}, &model.WebsiteDnsAccount{}, &model.WebsiteSSL{}, &model.WebsiteAcmeAccount{}); err != nil {
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

var AddDefaultGroup = &gormigrate.Migration{
	ID: "2023022-change-default-group",
	Migrate: func(tx *gorm.DB) error {
		defaultGroup := &model.Group{
			Name:      "默认",
			IsDefault: true,
			Type:      "website",
		}
		if err := tx.Create(defaultGroup).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Group{}).Where("name = ? AND type = ?", "default", "host").Update("name", "默认").Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Website{}).Where("1 = 1").Update("website_group_id", defaultGroup.ID).Error; err != nil {
			return err
		}
		return tx.Migrator().DropTable("website_groups")
	},
}

var AddTableRuntime = &gormigrate.Migration{
	ID: "20230406-add-table-runtime",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.Runtime{})
	},
}

var UpdateTableApp = &gormigrate.Migration{
	ID: "20230408-update-table-app",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.App{}); err != nil {
			return err
		}
		return nil
	},
}

var UpdateTableHost = &gormigrate.Migration{
	ID: "20230410-update-table-host",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Host{}); err != nil {
			return err
		}
		return nil
	},
}

var UpdateTableWebsite = &gormigrate.Migration{
	ID: "20230418-update-table-website",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Website{}); err != nil {
			return err
		}
		if err := tx.Model(&model.Website{}).Where("1 = 1").Update("site_dir", "/").Error; err != nil {
			return err
		}
		return nil
	},
}

var AddEntranceAndSSL = &gormigrate.Migration{
	ID: "20230414-add-entrance-and-ssl",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Model(&model.Setting{}).
			Where("key = ? AND value = ?", "SecurityEntrance", "onepanel").
			Updates(map[string]interface{}{"value": ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SSLType", Value: "self"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SSLID", Value: "0"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SSL", Value: "disable"}).Error; err != nil {
			return err
		}
		return tx.AutoMigrate(&model.Website{})
	},
}

var UpdateTableSetting = &gormigrate.Migration{
	ID: "20200516-update-table-setting",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "AppStoreLastModified", Value: "0"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var UpdateTableAppDetail = &gormigrate.Migration{
	ID: "20200517-update-table-app-detail",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.App{}); err != nil {
			return err
		}
		if err := tx.AutoMigrate(&model.AppDetail{}); err != nil {
			return err
		}
		return nil
	},
}

var AddBindAndAllowIPs = &gormigrate.Migration{
	ID: "20230517-add-bind-and-allow",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "BindDomain", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "AllowIPs", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "TimeZone", Value: common.LoadTimeZoneByCmd()}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "NtpSite", Value: "pool.ntp.org"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "MonitorInterval", Value: "1"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var UpdateCronjobWithSecond = &gormigrate.Migration{
	ID: "20200524-update-table-cronjob",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Cronjob{}); err != nil {
			return err
		}
		var jobs []model.Cronjob
		if err := tx.Where("exclusion_rules != ?", "").Find(&jobs).Error; err != nil {
			return err
		}
		for _, job := range jobs {
			if strings.Contains(job.ExclusionRules, ";") {
				newRules := strings.ReplaceAll(job.ExclusionRules, ";", ",")
				if err := tx.Model(&model.Cronjob{}).Where("id = ?", job.ID).Update("exclusion_rules", newRules).Error; err != nil {
					return err
				}
			}
		}
		return nil
	},
}

var UpdateWebsite = &gormigrate.Migration{
	ID: "20200530-update-table-website",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Website{}); err != nil {
			return err
		}
		return nil
	},
}

var AddBackupAccountDir = &gormigrate.Migration{
	ID: "20200620-add-backup-dir",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.BackupAccount{}, &model.Cronjob{}); err != nil {
			return err
		}
		return nil
	},
}

var AddMfaInterval = &gormigrate.Migration{
	ID: "20230625-add-mfa-interval",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "MFAInterval", Value: "30"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SystemIP", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "OneDriveID", Value: "MDEwOTM1YTktMWFhOS00ODU0LWExZGMtNmU0NWZlNjI4YzZi"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "OneDriveSc", Value: "akpuOFF+YkNXOU1OLWRzS1ZSRDdOcG1LT2ZRM0RLNmdvS1RkVWNGRA=="}).Error; err != nil {
			return err
		}
		return nil
	},
}

var UpdateAppDetail = &gormigrate.Migration{
	ID: "20230704-update-app-detail",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.AppDetail{}); err != nil {
			return err
		}
		if err := tx.Model(&model.AppDetail{}).Where("1 = 1").Update("ignore_upgrade", "0").Error; err != nil {
			return err
		}
		return nil
	},
}

var EncryptHostPassword = &gormigrate.Migration{
	ID: "20230703-encrypt-host-password",
	Migrate: func(tx *gorm.DB) error {
		var hosts []model.Host
		if err := tx.Where("1 = 1").Find(&hosts).Error; err != nil {
			return err
		}

		var encryptSetting model.Setting
		if err := tx.Where("key = ?", "EncryptKey").Find(&encryptSetting).Error; err != nil {
			return err
		}
		global.CONF.System.EncryptKey = encryptSetting.Value

		for _, host := range hosts {
			if len(host.Password) != 0 {
				pass, err := encrypt.StringEncrypt(host.Password)
				if err != nil {
					return err
				}
				if err := tx.Model(&model.Host{}).Where("id = ?", host.ID).Update("password", pass).Error; err != nil {
					return err
				}
			}
			if len(host.PrivateKey) != 0 {
				key, err := encrypt.StringEncrypt(host.PrivateKey)
				if err != nil {
					return err
				}
				if err := tx.Model(&model.Host{}).Where("id = ?", host.ID).Update("private_key", key).Error; err != nil {
					return err
				}
			}
			if len(host.PassPhrase) != 0 {
				pass, err := encrypt.StringEncrypt(host.PassPhrase)
				if err != nil {
					return err
				}
				if err := tx.Model(&model.Host{}).Where("id = ?", host.ID).Update("pass_phrase", pass).Error; err != nil {
					return err
				}
			}
		}
		return nil
	},
}
