package migrations

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/copier"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/1Panel-dev/1Panel/agent/utils/ssh"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddTable = &gormigrate.Migration{
	ID: "20250902-add-table",
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
			&model.RootCert{},
			&model.ClamRecord{},
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
	ID: "20250703-add-snapshot-rule",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			&model.Cronjob{},
		)
	},
}
var UpdatePHPRuntime = &gormigrate.Migration{
	ID: "20250702-update-php-runtime",
	Migrate: func(tx *gorm.DB) error {
		service.HandleOldPHPRuntime()
		return nil
	},
}
var AddSnapshotIgnore = &gormigrate.Migration{
	ID: "20250716-add-snapshot-ignore",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			&model.Snapshot{},
		)
	},
}

var InitAppLauncher = &gormigrate.Migration{
	ID: "20250702-init-app-launcher",
	Migrate: func(tx *gorm.DB) error {
		launchers := []string{"openresty", "mysql", "halo", "redis", "maxkb", "wordpress"}
		for _, val := range launchers {
			var item model.AppLauncher
			_ = tx.Model(&model.AppLauncher{}).Where("key = ?", val).First(&item).Error
			if item.ID == 0 {
				item.Key = val
				_ = tx.Create(&item).Error
			}
		}
		return nil
	},
}

var AddTableAlert = &gormigrate.Migration{
	ID: "20250122-add-table-alert",
	Migrate: func(tx *gorm.DB) error {
		return global.AlertDB.AutoMigrate(&model.Alert{}, &model.AlertTask{}, model.AlertLog{}, model.AlertConfig{})
	},
}

var InitAlertConfig = &gormigrate.Migration{
	ID: "20250705-init-alert-config",
	Migrate: func(tx *gorm.DB) error {
		records := []model.AlertConfig{
			{
				Type:   "sms",
				Title:  "xpack.alert.smsConfig",
				Status: "Enable",
				Config: `{"alertDailyNum":50}`,
			},
			{
				Type:   "common",
				Title:  "xpack.alert.commonConfig",
				Status: "Enable",
				Config: `{"isOffline":"Disable","alertSendTimeRange":{"noticeAlert":{"sendTimeRange":"08:00:00 - 23:59:59","type":["ssl","siteEndTime","panelPwdEndTime","panelUpdate"]},"resourceAlert":{"sendTimeRange":"00:00:00 - 23:59:59","type":["clams","cronJob","cpu","memory","load","disk"]}}}`,
			},
		}
		for _, r := range records {
			if err := global.AlertDB.Model(&model.AlertConfig{}).Create(&r).Error; err != nil {
				return err
			}
		}
		return nil
	},
}

var AddMethodToAlertLog = &gormigrate.Migration{
	ID: "20250713-add-method-to-alert_log",
	Migrate: func(tx *gorm.DB) error {
		if err := global.AlertDB.AutoMigrate(&model.AlertLog{}); err != nil {
			return err
		}
		if err := global.AlertDB.Model(&model.AlertLog{}).Where("method IS NULL OR method = ''").Update("method", "sms").Error; err != nil {
			return err
		}
		return nil
	},
}

var AddMethodToAlertTask = &gormigrate.Migration{
	ID: "20250723-add-method-to-alert_task",
	Migrate: func(tx *gorm.DB) error {
		if err := global.AlertDB.AutoMigrate(&model.AlertTask{}); err != nil {
			return err
		}
		if err := global.AlertDB.Model(&model.AlertTask{}).Where("method IS NULL OR method = ''").Update("method", "sms").Error; err != nil {
			return err
		}
		return nil
	},
}

var UpdateMcpServer = &gormigrate.Migration{
	ID: "20250729-update-mcp-server",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.McpServer{}); err != nil {
			return err
		}
		if err := tx.Model(&model.McpServer{}).Where("1=1").Update("output_transport", "sse").Error; err != nil {
			return err
		}
		return nil
	},
}

var InitCronjobGroup = &gormigrate.Migration{
	ID: "20250805-init-cronjob-group",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Cronjob{}); err != nil {
			return err
		}
		if err := tx.Model(&model.Cronjob{}).Where("1=1").Updates(map[string]interface{}{"group_id": 0}).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddColumnToAlert = &gormigrate.Migration{
	ID: "20250729-add-column-to-alert",
	Migrate: func(tx *gorm.DB) error {
		if err := global.AlertDB.AutoMigrate(&model.Alert{}); err != nil {
			return err
		}
		if err := global.AlertDB.Model(&model.Alert{}).
			Where("advanced_params IS NULL").
			Update("advanced_params", "").Error; err != nil {
			return err
		}
		return nil
	},
}

var UpdateWebsiteSSL = &gormigrate.Migration{
	ID: "20250819-update-website-ssl",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.WebsiteSSL{}); err != nil {
			return err
		}
		return nil
	},
}

var AddQuickJump = &gormigrate.Migration{
	ID: "20250901-add-quick-jump",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.QuickJump{}); err != nil {
			return err
		}
		if err := tx.Create(&model.QuickJump{Name: "Website", Title: "menu.website", Recommend: 10, IsShow: true, Router: "/websites"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.QuickJump{Name: "Database", Title: "home.database", Recommend: 30, IsShow: true, Router: "/databases"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.QuickJump{Name: "Cronjob", Title: "menu.cronjob", Recommend: 50, IsShow: true, Router: "/cronjobs"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.QuickJump{Name: "AppInstalled", Title: "home.appInstalled", Recommend: 70, IsShow: true, Router: "/apps/installed"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.QuickJump{Name: "File", Detail: "/", Title: "home.quickDir", Recommend: 90, IsShow: false, Router: "/hosts/files"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var UpdateMcpServerAddType = &gormigrate.Migration{
	ID: "20250904-update-mcp-server",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.McpServer{}); err != nil {
			return err
		}
		if err := tx.Model(&model.McpServer{}).Where("1=1").Update("type", "npx").Error; err != nil {
			return err
		}
		return nil
	},
}

var InitLocalSSHConn = &gormigrate.Migration{
	ID: "20250905-init-local-ssh",
	Migrate: func(tx *gorm.DB) error {
		itemPath := ""
		currentInfo, _ := user.Current()
		if len(currentInfo.HomeDir) == 0 {
			itemPath = "/root/.ssh/id_ed25519_1panel"
		} else {
			itemPath = path.Join(currentInfo.HomeDir, ".ssh/id_ed25519_1panel")
		}
		if _, err := os.Stat(itemPath); err != nil {
			_ = service.NewISSHService().CreateRootCert(dto.CreateRootCert{EncryptionMode: "ed25519", Name: "id_ed25519_1panel", Description: "1Panel Terminal"})
		}
		privateKey, _ := os.ReadFile(itemPath)
		connWithKey := ssh.ConnInfo{
			Addr:       "127.0.0.1",
			User:       "root",
			Port:       22,
			AuthMode:   "key",
			PrivateKey: privateKey,
		}
		if _, err := ssh.NewClient(connWithKey); err != nil {
			return nil
		}
		var conn model.LocalConnInfo
		_ = copier.Copy(&conn, &connWithKey)
		conn.PrivateKey = string(privateKey)
		conn.PassPhrase = ""
		localConn, _ := json.Marshal(&conn)
		connAfterEncrypt, _ := encrypt.StringEncrypt(string(localConn))
		if err := tx.Model(&model.Setting{}).Where("key = ?", "LocalSSHConn").Updates(map[string]interface{}{"value": connAfterEncrypt}).Error; err != nil {
			return err
		}
		return nil
	},
}

var InitLocalSSHShow = &gormigrate.Migration{
	ID: "20250908-init-local-ssh-show",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "LocalSSHConnShow", Value: constant.StatusEnable}).Error; err != nil {
			return err
		}
		return nil
	},
}
