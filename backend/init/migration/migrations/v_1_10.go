package migrations

import (
	"encoding/json"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
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

var NewMonitorDB = &gormigrate.Migration{
	ID: "20240408-new-monitor-db",
	Migrate: func(tx *gorm.DB) error {
		var (
			bases    []model.MonitorBase
			ios      []model.MonitorIO
			networks []model.MonitorNetwork
		)
		_ = tx.Find(&bases).Error
		_ = tx.Find(&ios).Error
		_ = tx.Find(&networks).Error

		if err := global.MonitorDB.AutoMigrate(&model.MonitorBase{}, &model.MonitorIO{}, &model.MonitorNetwork{}); err != nil {
			return err
		}
		_ = global.MonitorDB.Exec("DELETE FROM monitor_bases").Error
		_ = global.MonitorDB.Exec("DELETE FROM monitor_ios").Error
		_ = global.MonitorDB.Exec("DELETE FROM monitor_networks").Error

		if len(bases) != 0 {
			for i := 0; i <= len(bases)/200; i++ {
				var itemData []model.MonitorBase
				if 200*(i+1) <= len(bases) {
					itemData = bases[200*i : 200*(i+1)]
				} else {
					itemData = bases[200*i:]
				}
				if len(itemData) != 0 {
					if err := global.MonitorDB.Create(&itemData).Error; err != nil {
						return err
					}
				}
			}
		}
		if len(ios) != 0 {
			for i := 0; i <= len(ios)/200; i++ {
				var itemData []model.MonitorIO
				if 200*(i+1) <= len(ios) {
					itemData = ios[200*i : 200*(i+1)]
				} else {
					itemData = ios[200*i:]
				}
				if len(itemData) != 0 {
					if err := global.MonitorDB.Create(&itemData).Error; err != nil {
						return err
					}
				}
			}
		}
		if len(networks) != 0 {
			for i := 0; i <= len(networks)/200; i++ {
				var itemData []model.MonitorNetwork
				if 200*(i+1) <= len(networks) {
					itemData = networks[200*i : 200*(i+1)]
				} else {
					itemData = networks[200*i:]
				}
				if len(itemData) != 0 {
					if err := global.MonitorDB.Create(&itemData).Error; err != nil {
						return err
					}
				}
			}
		}
		return nil
	},
}

var AddNoAuthSetting = &gormigrate.Migration{
	ID: "20240328-add-no-auth-setting",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "NoAuthSetting", Value: "200"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var UpdateXpackHideMenu = &gormigrate.Migration{
	ID: "20240411-update-xpack-hide-menu",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Model(&model.Setting{}).Where("key", "XpackHideMenu").Updates(map[string]interface{}{"value": "{\"id\":\"1\",\"label\":\"/xpack\",\"isCheck\":true,\"title\":\"xpack.menu\",\"children\":[{\"id\":\"2\",\"title\":\"xpack.waf.name\",\"path\":\"/xpack/waf/dashboard\",\"label\":\"Dashboard\",\"isCheck\":true},{\"id\":\"3\",\"title\":\"xpack.tamper.tamper\",\"path\":\"/xpack/tamper\",\"label\":\"Tamper\",\"isCheck\":true},{\"id\":\"4\",\"title\":\"xpack.gpu.gpu\",\"path\":\"/xpack/gpu\",\"label\":\"GPU\",\"isCheck\":true},{\"id\":\"5\",\"title\":\"xpack.setting.setting\",\"path\":\"/xpack/setting\",\"label\":\"XSetting\",\"isCheck\":true}]}"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddMenuTabsSetting = &gormigrate.Migration{
	ID: "20240415-add-menu-tabs-setting",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "MenuTabs", Value: "disable"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddDeveloperSetting = &gormigrate.Migration{
	ID: "20240423-add-developer-setting",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "DeveloperMode", Value: "disable"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddWebsiteSSLColumn = &gormigrate.Migration{
	ID: "20240508-update-website-ssl",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.WebsiteSSL{}); err != nil {
			return err
		}
		return nil
	},
}

var AddRedisCommand = &gormigrate.Migration{
	ID: "20240515-add-redis-command",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.RedisCommand{}); err != nil {
			return err
		}
		return nil
	},
}

var AddMonitorMenu = &gormigrate.Migration{
	ID: "20240517-update-xpack-hide-menu",
	Migrate: func(tx *gorm.DB) error {
		var (
			setting model.Setting
			menu    dto.XpackHideMenu
		)
		tx.Model(&model.Setting{}).Where("key", "XpackHideMenu").First(&setting)
		if err := json.Unmarshal([]byte(setting.Value), &menu); err != nil {
			return err
		}
		menu.Children = append(menu.Children, dto.XpackHideMenu{
			ID:      "6",
			Title:   "xpack.monitor.name",
			Path:    "/xpack/monitor/dashboard",
			Label:   "MonitorDashboard",
			IsCheck: true,
		})
		data, err := json.Marshal(menu)
		if err != nil {
			return err
		}
		return tx.Model(&model.Setting{}).Where("key", "XpackHideMenu").Updates(map[string]interface{}{"value": string(data)}).Error
	},
}

var AddFtp = &gormigrate.Migration{
	ID: "20240521-add-ftp",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Ftp{}, model.Website{}); err != nil {
			return err
		}
		return nil
	},
}

var AddProxy = &gormigrate.Migration{
	ID: "20240528-add-proxy",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "ProxyType", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "ProxyUrl", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "ProxyPort", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "ProxyUser", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "ProxyPasswd", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "ProxyPasswdKeep", Value: ""}).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddForward = &gormigrate.Migration{
	ID: "202400611-add-forward",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Forward{}); err != nil {
			return err
		}
		return nil
	},
}

var AddCronJobColumn = &gormigrate.Migration{
	ID: "20240524-add-cronjob-command",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Cronjob{}); err != nil {
			return err
		}
		return nil
	},
}

var AddShellColumn = &gormigrate.Migration{
	ID: "20240620-update-website-ssl",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.WebsiteSSL{}); err != nil {
			return err
		}
		return nil
	},
}

var AddClam = &gormigrate.Migration{
	ID: "20240701-add-clam",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Clam{}); err != nil {
			return err
		}
		return nil
	},
}

var AddClamStatus = &gormigrate.Migration{
	ID: "20240716-add-clam-status",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Clam{}); err != nil {
			return err
		}
		return nil
	},
}

var AddAlertMenu = &gormigrate.Migration{
	ID: "20240706-update-xpack-hide-menu",
	Migrate: func(tx *gorm.DB) error {
		var (
			setting model.Setting
			menu    dto.XpackHideMenu
		)
		tx.Model(&model.Setting{}).Where("key", "XpackHideMenu").First(&setting)
		if err := json.Unmarshal([]byte(setting.Value), &menu); err != nil {
			return err
		}
		menu.Children = append(menu.Children, dto.XpackHideMenu{
			ID:      "7",
			Title:   "xpack.alert.alert",
			Path:    "/xpack/alert/dashboard",
			Label:   "XAlertDashboard",
			IsCheck: true,
		})
		data, err := json.Marshal(menu)
		if err != nil {
			return err
		}
		return tx.Model(&model.Setting{}).Where("key", "XpackHideMenu").Updates(map[string]interface{}{"value": string(data)}).Error
	},
}

var AddComposeColumn = &gormigrate.Migration{
	ID: "20240906-add-compose-command",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Compose{}); err != nil {
			return err
		}
		return nil
	},
}

var AddAutoRestart = &gormigrate.Migration{
	ID: "20241021-add-auto-restart",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "AutoRestart", Value: "enable"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddApiInterfaceConfig = &gormigrate.Migration{
	ID: "202411-add-api-interface-config",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "ApiInterfaceStatus", Value: "disable"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "ApiKey", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "IpWhiteList", Value: ""}).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddApiKeyValidityTime = &gormigrate.Migration{
	ID: "20241226-add-api-key-validity-time",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "ApiKeyValidityTime", Value: "120"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var UpdateAppTag = &gormigrate.Migration{
	ID: "20250114-update-app-tag",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Tag{}); err != nil {
			return err
		}
		return nil
	},
}

var UpdateApp = &gormigrate.Migration{
	ID: "20250213-update-app",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.App{}); err != nil {
			return err
		}
		return nil
	},
}
