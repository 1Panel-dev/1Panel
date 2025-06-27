package migrations

import (
	"fmt"
	"path"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddTable = &gormigrate.Migration{
	ID: "20250507-add-table",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			&model.AppDetail{},
			&model.AppInstallResource{},
			&model.AppInstall{},
			&model.AppTag{},
			&model.Tag{},
			&model.App{},
			&model.AppLauncher{},
			&model.OllamaModel{},
			&model.BackupAccount{},
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
			&model.ScriptLibrary{},
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
			&model.Group{},
			&model.AppIgnoreUpgrade{},
			&model.McpServer{},
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
		global.CONF.Base.EncryptKey = common.RandStr(16)
		nodeInfo, err := xpack.LoadNodeInfo(true)
		if err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "BaseDir", Value: nodeInfo.BaseDir}).Error; err != nil {
			return err
		}
		itemKey, _ := encrypt.StringEncrypt(nodeInfo.ServerKey)
		if err := tx.Create(&model.Setting{Key: "ServerKey", Value: itemKey}).Error; err != nil {
			return err
		}
		itemCrt, _ := encrypt.StringEncrypt(nodeInfo.ServerCrt)
		if err := tx.Create(&model.Setting{Key: "ServerCrt", Value: itemCrt}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "NodeScope", Value: nodeInfo.Scope}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "NodePort", Value: fmt.Sprintf("%v", nodeInfo.NodePort)}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SystemVersion", Value: nodeInfo.Version}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "EncryptKey", Value: global.CONF.Base.EncryptKey}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "DockerSockPath", Value: "unix:///var/run/docker.sock"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SystemStatus", Value: "Free"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "Language", Value: "zh"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SystemIP", Value: ""}).Error; err != nil {
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
		if err := tx.Create(&model.Setting{Key: "MonitorStatus", Value: constant.StatusEnable}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "MonitorStoreDays", Value: "7"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "MonitorInterval", Value: "5"}).Error; err != nil {
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

		if err := tx.Create(&model.Setting{Key: "AppStoreVersion", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "AppStoreSyncStatus", Value: "SyncSuccess"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "AppStoreLastModified", Value: "0"}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "FileRecycleBin", Value: constant.StatusEnable}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "SnapshotIgnore", Value: "*.sock"}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "LocalSSHConn", Value: ""}).Error; err != nil {
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
		phpExtensions := []model.PHPExtensions{
			{
				Name:       "Default",
				Extensions: "bcmath,ftp,gd,gettext,intl,mysqli,pcntl,pdo_mysql,shmop,soap,sockets,sysvsem,xmlrpc,zip",
			},
			{
				Name:       "WordPress",
				Extensions: "exif,igbinary,imagick,intl,zip,apcu,memcached,opcache,redis,shmop,mysqli,pdo_mysql,gd",
			},
			{
				Name:       "Flarum",
				Extensions: "curl,gd,pdo_mysql,mysqli,bz2,exif,yaf,imap",
			},
			{
				Name:       "SeaCMS",
				Extensions: "mysqli,pdo_mysql,gd,curl",
			},
			{
				Name:       "Dev",
				Extensions: "bcmath,ftp,gd,gettext,intl,mysqli,pcntl,pdo_mysql,shmop,soap,sockets,sysvsem,xmlrpc,zip,exif,igbinary,imagick,apcu,memcached,opcache,redis,bc,image,dom,iconv,mbstring,mysqlnd,openssl,pdo,tokenizer,xml,curl,bz2,yaf,imap,xdebug,swoole,pdo_pgsql,fileinfo,pgsql,calendar,gmp",
			},
		}

		for _, ext := range phpExtensions {
			if err := tx.Create(&ext).Error; err != nil {
				return err
			}
		}
		return nil
	},
}

var AddTaskTable = &gormigrate.Migration{
	ID: "20241226-add-task",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			&model.Task{},
		)
	},
}

var InitBackup = &gormigrate.Migration{
	ID: "20241226-init-backup",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.BackupAccount{
			Name:       "localhost",
			Type:       "LOCAL",
			BackupPath: path.Join(global.Dir.DataDir, "backup"),
		}).Error; err != nil {
			return err
		}
		return nil
	},
}

var InitDefault = &gormigrate.Migration{
	ID: "20250301-init-default",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Group{Name: "Default", Type: "website", IsDefault: true}).Error; err != nil {
			return err
		}
		return nil
	},
}

var UpdateWebsiteExpireDate = &gormigrate.Migration{
	ID: "20250304-update-website",
	Migrate: func(tx *gorm.DB) error {
		targetDate := time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)

		if err := tx.Model(&model.Website{}).
			Where("expire_date = ?", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)).
			Update("expire_date", targetDate).Error; err != nil {
			return err
		}
		return nil
	},
}

var UpdateRuntime = &gormigrate.Migration{
	ID: "20250624-update-runtime",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			&model.Runtime{},
		)
	},
}

var AddSnapshotRule = &gormigrate.Migration{
	ID: "20250627-add-snapshot-rule",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			&model.Cronjob{},
		)
	},
}

var UpdatePHPRuntime = &gormigrate.Migration{
	ID: "20250624-update-php-runtime",
	Migrate: func(tx *gorm.DB) error {
		service.HandleOldPHPRuntime()
		return nil
	},
}
