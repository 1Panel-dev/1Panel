package migrations

import (
	"encoding/json"
	"os"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddTable = &gormigrate.Migration{
	ID: "20240722-add-table",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			&model.AppDetail{},
			&model.AppInstallResource{},
			&model.AppInstall{},
			&model.AppTag{},
			&model.Tag{},
			&model.App{},
			&model.BackupRecord{},
			&model.Clam{},
			&model.ComposeTemplate{},
			&model.Compose{},
			&model.Cronjob{},
			&model.Database{},
			&model.DatabaseMysql{},
			&model.DatabasePostgresql{},
			&model.Favorite{},
			&model.Forward{},
			&model.Firewall{},
			&model.Ftp{},
			&model.ImageRepo{},
			&model.JobRecords{},
			&model.MonitorBase{},
			&model.MonitorIO{},
			&model.MonitorNetwork{},
			&model.PHPExtensions{},
			&model.Runtime{},
			&model.Setting{},
			&model.Snapshot{},
			&model.SnapshotStatus{},
			&model.Tag{},
			&model.Website{},
			&model.WebsiteAcmeAccount{},
			&model.WebsiteCA{},
			&model.WebsiteDnsAccount{},
			&model.WebsiteDomain{},
			&model.WebsiteSSL{},
		)
	},
}

var AddMonitorTable = &gormigrate.Migration{
	ID: "20240813-add-monitor-table",
	Migrate: func(tx *gorm.DB) error {
		return global.MonitorDB.AutoMigrate(
			&model.MonitorBase{},
			&model.MonitorIO{},
			&model.MonitorNetwork{},
		)
	},
}

var InitSetting = &gormigrate.Migration{
	ID: "20240722-init-setting",
	Migrate: func(tx *gorm.DB) error {
		encryptKey := common.RandStr(16)
		global.CONF.System.EncryptKey = encryptKey
		if err := tx.Create(&model.Setting{Key: "EncryptKey", Value: encryptKey}).Error; err != nil {
			return err
		}
		currentNode := ""
		if _, err := os.Stat("/opt/1panel/nodeJson"); err == nil {
			type nodeInfo struct {
				MasterAddr  string `json:"masterAddr"`
				Token       string `json:"token"`
				ServerCrt   string `json:"serverCrt"`
				ServerKey   string `json:"serverKey"`
				CurrentNode string `json:"currentNode"`
			}
			nodeJson, err := os.ReadFile("/opt/1panel/nodeJson")
			if err != nil {
				return err
			}
			var node nodeInfo
			if err := json.Unmarshal(nodeJson, &node); err != nil {
				return err
			}
			itemKey, _ := encrypt.StringEncryptWithBase64(node.ServerKey)
			if err := tx.Create(&model.Setting{Key: "ServerKey", Value: itemKey}).Error; err != nil {
				return err
			}
			itemCrt, _ := encrypt.StringEncryptWithBase64(node.ServerCrt)
			if err := tx.Create(&model.Setting{Key: "ServerCrt", Value: itemCrt}).Error; err != nil {
				return err
			}
			itemToken, _ := encrypt.StringEncryptWithBase64(node.Token)
			if err := tx.Create(&model.Setting{Key: "Token", Value: itemToken}).Error; err != nil {
				return err
			}
			if err := tx.Create(&model.Setting{Key: "MasterAddr", Value: node.MasterAddr}).Error; err != nil {
				return err
			}
			global.CONF.System.MasterAddr = node.MasterAddr
			global.CONF.System.MasterToken = itemToken
			global.IsMaster = false
			currentNode = node.CurrentNode
		} else {
			global.IsMaster = true
		}
		if err := tx.Create(&model.Setting{Key: "CurrentNode", Value: currentNode}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "SystemIP", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "DockerSockPath", Value: "unix:///var/run/docker.sock"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SystemVersion", Value: global.CONF.System.Version}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SystemStatus", Value: "Free"}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "LocalTime", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "TimeZone", Value: common.LoadTimeZoneByCmd()}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "NtpSite", Value: "pool.ntp.org"}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "LastCleanTime", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "LastCleanSize", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "LastCleanData", Value: ""}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "DefaultNetwork", Value: "all"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "MonitorStatus", Value: "enable"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "MonitorStoreDays", Value: "7"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "MonitorInterval", Value: "5"}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "AppStoreVersion", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "AppStoreSyncStatus", Value: "SyncSuccess"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "AppStoreLastModified", Value: "0"}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "FileRecycleBin", Value: "enable"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SnapshotIgnore", Value: "*.sock"}).Error; err != nil {
			return err
		}

		_ = os.Remove(("/opt/1panel/nodeJson"))
		return nil
	},
}

var InitImageRepo = &gormigrate.Migration{
	ID: "20240722-init-imagerepo",
	Migrate: func(tx *gorm.DB) error {
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

var InitDefaultCA = &gormigrate.Migration{
	ID: "20240722-init-default-ca",
	Migrate: func(tx *gorm.DB) error {
		caService := service.NewIWebsiteCAService()
		if _, err := caService.Create(request.WebsiteCACreate{
			CommonName:       "1Panel-CA",
			Country:          "CN",
			KeyType:          "P256",
			Name:             "1Panel",
			Organization:     "FIT2CLOUD",
			OrganizationUint: "1Panel",
			Province:         "Beijing",
			City:             "Beijing",
		}); err != nil {
			return err
		}
		return nil
	},
}

var InitPHPExtensions = &gormigrate.Migration{
	ID: "20240722-add-php-extensions",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.PHPExtensions{Name: "默认", Extensions: "bcmath,gd,gettext,intl,pcntl,shmop,soap,sockets,sysvsem,xmlrpc,zip"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.PHPExtensions{Name: "WordPress", Extensions: "exif,igbinary,imagick,intl,zip,apcu,memcached,opcache,redis,bc,image,shmop,mysqli,pdo_mysql,gd"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.PHPExtensions{Name: "Flarum", Extensions: "curl,gd,pdo_mysql,mysqli,bz2,exif,yaf,imap"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.PHPExtensions{Name: "苹果CMS-V10", Extensions: "mysqli,pdo_mysql,zip,gd,redis,memcache,memcached"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.PHPExtensions{Name: "SeaCMS", Extensions: "mysqli,pdo_mysql,gd,curl"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddTask = &gormigrate.Migration{
	ID: "20240802-add-task",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			&model.Task{})
	},
}

var UpdateWebsite = &gormigrate.Migration{
	ID: "20240812-update-website",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			&model.Website{})
	},
}

var UpdateWebsiteDomain = &gormigrate.Migration{
	ID: "20240808-update-website-domain",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			&model.WebsiteDomain{})
	},
}
