package migrations

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/service"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/common"

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
			&model.BackupAccount{},
			&model.BackupRecord{},
			&model.Clam{},
			&model.Command{},
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
			&model.Group{},
			&model.Host{},
			&model.ImageRepo{},
			&model.JobRecords{},
			&model.MonitorBase{},
			&model.MonitorIO{},
			&model.MonitorNetwork{},
			&model.PHPExtensions{},
			&model.RedisCommand{},
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

var InitHost = &gormigrate.Migration{
	ID: "20240722-init-host",
	Migrate: func(tx *gorm.DB) error {
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

var InitSetting = &gormigrate.Migration{
	ID: "20240722-init-setting",
	Migrate: func(tx *gorm.DB) error {
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

		if err := tx.Create(&model.Setting{Key: "DefaultNetwork", Value: "all"}).Error; err != nil {
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

		if err := tx.Create(&model.Setting{Key: "OneDriveID", Value: "MDEwOTM1YTktMWFhOS00ODU0LWExZGMtNmU0NWZlNjI4YzZi"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "OneDriveSc", Value: "akpuOFF+YkNXOU1OLWRzS1ZSRDdOcG1LT2ZRM0RLNmdvS1RkVWNGRA=="}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "FileRecycleBin", Value: "enable"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SnapshotIgnore", Value: "*.sock"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var InitBackupAccount = &gormigrate.Migration{
	ID: "20240722-init-backup",
	Migrate: func(tx *gorm.DB) error {
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

var InitDefaultGroup = &gormigrate.Migration{
	ID: "20240722-init-default-group",
	Migrate: func(tx *gorm.DB) error {
		websiteGroup := &model.Group{
			Name:      "默认",
			IsDefault: true,
			Type:      "website",
		}
		if err := tx.Create(websiteGroup).Error; err != nil {
			return err
		}
		commandGroup := &model.Group{IsDefault: true, Name: "默认", Type: "command"}
		if err := tx.Create(commandGroup).Error; err != nil {
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
