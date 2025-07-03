package migrations

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/dto"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/init/migration/helper"
	"github.com/1Panel-dev/1Panel/core/utils/cmd"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddTable = &gormigrate.Migration{
	ID: "20240506-add-table",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			&model.OperationLog{},
			&model.LoginLog{},
			&model.Setting{},
			&model.BackupAccount{},
			&model.Group{},
			&model.Host{},
			&model.Command{},
			&model.UpgradeLog{},
			&model.ScriptLibrary{},
		)
	},
}

var InitSetting = &gormigrate.Migration{
	ID: "20200908-add-table-setting",
	Migrate: func(tx *gorm.DB) error {
		encryptKey := common.RandStr(16)
		if err := tx.Create(&model.Setting{Key: "UserName", Value: global.CONF.Base.Username}).Error; err != nil {
			return err
		}
		global.CONF.Base.EncryptKey = encryptKey
		pass, _ := encrypt.StringEncrypt(global.CONF.Base.Password)
		language := "en"
		if global.CONF.Base.Language == "zh" {
			language = "zh"
		}
		if err := tx.Create(&model.Setting{Key: "Password", Value: pass}).Error; err != nil {
			return err
		}
		_, _ = cmd.RunDefaultWithStdoutBashCf("%s sed -i -e 's#ORIGINAL_PASSWORD=.*#ORIGINAL_PASSWORD=**********#g' /usr/local/bin/1pctl", cmd.SudoHandleCmd())
		if err := tx.Create(&model.Setting{Key: "Theme", Value: "light"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "MenuTabs", Value: constant.StatusDisable}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "PanelName", Value: "1Panel"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "Language", Value: language}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SessionTimeout", Value: "86400"}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "SSLType", Value: "self"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SSLID", Value: "0"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SSL", Value: constant.StatusDisable}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "DeveloperMode", Value: constant.StatusDisable}).Error; err != nil {
			return err
		}
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
		if err := tx.Create(&model.Setting{Key: "HideMenu", Value: helper.LoadMenus()}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "ServerPort", Value: global.CONF.Conn.Port}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SecurityEntrance", Value: global.CONF.Conn.Entrance}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "EncryptKey", Value: encryptKey}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "ExpirationTime", Value: time.Now().AddDate(0, 0, 10).Format(constant.DateTimeLayout)}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "ExpirationDays", Value: "0"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "ComplexityVerification", Value: constant.StatusEnable}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "MFAStatus", Value: constant.StatusDisable}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "MFASecret", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "MFAInterval", Value: "30"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SystemVersion", Value: global.CONF.Base.Version}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SystemStatus", Value: "Free"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "BindAddress", Value: "0.0.0.0"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "Ipv6", Value: constant.StatusDisable}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "BindDomain", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "AllowIPs", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "NoAuthSetting", Value: "200"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "ApiInterfaceStatus", Value: constant.StatusDisable}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "ApiKey", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "IpWhiteList", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "ApiKeyValidityTime", Value: "120"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "ScriptVersion", Value: ""}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "UninstallDeleteImage", Value: constant.StatusDisable}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "UpgradeBackup", Value: "Enable"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "UninstallDeleteBackup", Value: constant.StatusDisable}).Error; err != nil {
			return err
		}
		return nil
	},
}

var InitTerminalSetting = &gormigrate.Migration{
	ID: "20240814-init-terminal-setting",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "LineHeight", Value: "1.2"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "LetterSpacing", Value: "0"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "FontSize", Value: "12"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "CursorBlink", Value: constant.StatusEnable}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "CursorStyle", Value: "block"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "Scrollback", Value: "1000"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "ScrollSensitivity", Value: "6"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var InitHost = &gormigrate.Migration{
	ID: "20240816-init-host",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Group{Name: "Default", Type: "host", IsDefault: true}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Group{Name: "Default", Type: "node", IsDefault: true}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Group{Name: "Default", Type: "command", IsDefault: true}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Group{Name: "Default", Type: "website", IsDefault: true}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Group{Name: "Default", Type: "redis", IsDefault: true}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Group{Name: "Default", Type: "script", IsDefault: true}).Error; err != nil {
			return err
		}
		return nil
	},
}

var InitOneDrive = &gormigrate.Migration{
	ID: "20240808-init-one-drive",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "OneDriveID", Value: "MDEwOTM1YTktMWFhOS00ODU0LWExZGMtNmU0NWZlNjI4YzZi"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "OneDriveSc", Value: "akpuOFF+YkNXOU1OLWRzS1ZSRDdOcG1LT2ZRM0RLNmdvS1RkVWNGRA=="}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.BackupAccount{
			Name: "localhost",
			Type: "LOCAL",
			Vars: fmt.Sprintf("{\"dir\":\"%s\"}", path.Join(global.CONF.Base.InstallDir, "1panel/backup")),
		}).Error; err != nil {
			return err
		}
		return nil
	},
}

var InitGoogle = &gormigrate.Migration{
	ID: "20241111-init-google",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "GoogleID", Value: "NTU2NTQ3NDYwMTQtY2Q0bGR0dDk2aGNsNWcxYWtwdmJhZTFmcjJlZ2Y0MXAuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20K"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "GoogleSc", Value: "R09DU1BYLXRibXg0QVdVZ3d3Ykc2QW1XTHQ3YUdaZElVeE4K"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddTaskDB = &gormigrate.Migration{
	ID: "20241125-add-task-table",
	Migrate: func(tx *gorm.DB) error {
		return global.TaskDB.AutoMigrate(
			&model.Task{},
		)
	},
}

var AddXpackHideMenu = &gormigrate.Migration{
	ID: "20250529-add-xpack-hide-menu",
	Migrate: func(tx *gorm.DB) error {
		var menuJSON string
		if err := tx.Model(&model.Setting{}).Where("key = ?", "HideMenu").Pluck("value", &menuJSON).Error; err != nil {
			return err
		}
		if strings.Contains(menuJSON, `"XApp"`) && strings.Contains(menuJSON, `"/xpack/app"`) {
			return nil
		}

		var menus []dto.ShowMenu
		if err := json.Unmarshal([]byte(menuJSON), &menus); err != nil {
			return tx.Model(&model.Setting{}).
				Where("key = ?", "HideMenu").
				Update("value", helper.LoadMenus()).Error
		}

		newItem := dto.ShowMenu{
			ID:       "118",
			Disabled: false,
			Title:    "xpack.app.app",
			IsShow:   true,
			Label:    "XApp",
			Path:     "/xpack/app",
		}

		for i, menu := range menus {
			if menu.ID == "11" {
				exists := false
				for _, child := range menu.Children {
					if child.ID == newItem.ID {
						exists = true
						break
					}
				}
				if !exists {
					menus[i].Children = append([]dto.ShowMenu{newItem}, menus[i].Children...)
				}
				break
			}
		}

		updatedJSON, err := json.Marshal(menus)
		if err != nil {
			return tx.Model(&model.Setting{}).
				Where("key = ?", "HideMenu").
				Update("value", helper.LoadMenus()).Error
		}

		return tx.Model(&model.Setting{}).Where("key = ?", "HideMenu").Update("value", string(updatedJSON)).Error
	},
}

var UpdateXpackHideMenu = &gormigrate.Migration{
	ID: "20250617-update-xpack-hide-menu",
	Migrate: func(tx *gorm.DB) error {
		var menuJSON string
		if err := tx.Model(&model.Setting{}).Where("key = ?", "HideMenu").Pluck("value", &menuJSON).Error; err != nil {
			return err
		}
		var menus []dto.ShowMenu
		if err := json.Unmarshal([]byte(menuJSON), &menus); err != nil {
			return tx.Model(&model.Setting{}).
				Where("key = ?", "HideMenu").
				Update("value", helper.LoadMenus()).Error
		}
		newItem := dto.ShowMenu{
			ID:       "119",
			Disabled: false,
			Title:    "xpack.upage",
			IsShow:   true,
			Label:    "Upage",
			Path:     "/xpack/upage",
		}

		for i, menu := range menus {
			if menu.ID == "11" {
				exists := false
				for _, child := range menu.Children {
					if child.ID == newItem.ID {
						exists = true
						break
					}
				}
				if exists {
					break
				}

				insertIndex := -1
				for j, child := range menu.Children {
					if child.ID == "111" {
						insertIndex = j
						break
					}
				}

				if insertIndex != -1 {
					children := menu.Children
					menus[i].Children = append(children[:insertIndex+1], append([]dto.ShowMenu{newItem}, children[insertIndex+1:]...)...)
				} else {
					menus[i].Children = append([]dto.ShowMenu{newItem}, menus[i].Children...)
				}
				break
			}
		}

		for i, menu := range menus {
			if menu.ID == "11" {
				existsIndex := -1
				for j, child := range menu.Children {
					if child.ID == "118" {
						existsIndex = j
						break
					}
				}

				if existsIndex == 0 {
					break
				}

				var item118 dto.ShowMenu
				if existsIndex != -1 {
					item118 = menu.Children[existsIndex]
					menus[i].Children = append(menu.Children[:existsIndex], menu.Children[existsIndex+1:]...)
				} else {
					item118 = dto.ShowMenu{
						ID:       "118",
						Disabled: false,
						Title:    "xpack.app.app",
						IsShow:   true,
						Label:    "XApp",
						Path:     "/xpack/app",
					}
				}

				menus[i].Children = append([]dto.ShowMenu{item118}, menus[i].Children...)
				break
			}
		}

		var idx9, idx10 = -1, -1
		for i, menu := range menus {
			if menu.ID == "9" && menu.Path == "/toolbox" {
				idx9 = i
			}
			if menu.ID == "10" && menu.Path == "/cronjobs" {
				idx10 = i
			}
		}
		if idx9 != -1 && idx10 != -1 && idx10 > idx9 {
			menus[idx9], menus[idx10] = menus[idx10], menus[idx9]
		}

		for i, menu := range menus {
			if menu.ID == "7" {
				for j, child := range menu.Children {
					if child.ID == "75" {
						if child.Title != "menu.processManage" {
							menus[i].Children[j].Title = "menu.processManage"
						}
						break
					}
				}
				break
			}
		}

		updatedJSON, err := json.Marshal(menus)
		if err != nil {
			return tx.Model(&model.Setting{}).
				Where("key = ?", "HideMenu").
				Update("value", helper.LoadMenus()).Error
		}

		return tx.Model(&model.Setting{}).Where("key = ?", "HideMenu").Update("value", string(updatedJSON)).Error
	},
}

var UpdateGoogle = &gormigrate.Migration{
	ID: "20250616-update-google",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Model(&model.Setting{}).
			Where("key = ?", "GoogleID").
			Update("value", "NTU2NTQ3NDYwMTQtY2Q0bGR0dDk2aGNsNWcxYWtwdmJhZTFmcjJlZ2Y0MXAuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20=").Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Setting{}).
			Where("key = ?", "GoogleSc").
			Update("value", "R09DU1BYLXRibXg0QVdVZ3d3Ykc2QW1XTHQ3YUdaZElVeE4=").Error; err != nil {
			return err
		}
		return nil
	},
}
