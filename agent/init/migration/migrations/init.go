package migrations

import (
	"os"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddTable = &gormigrate.Migration{
	ID: "20241009-add-table",
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
			&model.Tag{},
			&model.Website{},
			&model.WebsiteAcmeAccount{},
			&model.WebsiteCA{},
			&model.WebsiteDnsAccount{},
			&model.WebsiteDomain{},
			&model.WebsiteSSL{},
			&model.Task{},
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
		global.CONF.System.EncryptKey = common.RandStr(16)
		isMaster, currentNode, err := xpack.InitNodeData(tx)
		if err != nil {
			return err
		}
		global.IsMaster = isMaster
		if err := tx.Create(&model.Setting{Key: "CurrentNode", Value: currentNode}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "EncryptKey", Value: global.CONF.System.EncryptKey}).Error; err != nil {
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

var AddTaskDB = &gormigrate.Migration{
	ID: "20240822-add-task-table",
	Migrate: func(tx *gorm.DB) error {
		return global.TaskDB.AutoMigrate(
			&model.Task{},
		)
	},
}

var UpdateApp = &gormigrate.Migration{
	ID: "20240826-update-app",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			&model.App{})
	},
}

var UpdateAppInstall = &gormigrate.Migration{
	ID: "20240828-update-app-install",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			&model.AppInstall{})
	},
}

var UpdateSnapshot = &gormigrate.Migration{
	ID: "20240926-update-snapshot",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.Snapshot{})
	},
}

var UpdateCronjob = &gormigrate.Migration{
	ID: "20241011-update-cronjob",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.Cronjob{}, &model.JobRecords{})
	},
}
