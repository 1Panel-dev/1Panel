package migrations

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	providercatalog "github.com/1Panel-dev/1Panel/agent/app/provider"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	migrationutils "github.com/1Panel-dev/1Panel/agent/init/migration/migrations/utils"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/copier"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/ssh"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var AddTable = &gormigrate.Migration{
	ID: "20250930-add-table",
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
			&model.DatabaseUser{},
			&model.DatabaseMongodb{},
			&model.DatabasePostgresql{},
			&model.Favorite{},
			&model.FileShare{},
			&model.Firewall{},
			&model.Host{},
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
		nodeInfo, err := xpack.MultiNodeProvider.LoadNodeInfo(true)
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
		lang := common.LoadParamsWithoutPanic("LANGUAGE")
		if err := tx.Create(&model.Setting{Key: "Language", Value: lang}).Error; err != nil {
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
			KeyType:          "EC256",
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
				Type:       "sms",
				Title:      "xpack.alert.smsConfig",
				Status:     "Enable",
				Config:     `{"alertDailyNum":50}`,
				CreateUser: "system",
				UpdateUser: "system",
			},
			{
				Type:       "common",
				Title:      "xpack.alert.commonConfig",
				Status:     "Enable",
				Config:     `{"isOffline":"Disable","alertSendTimeRange":{"noticeAlert":{"sendTimeRange":"08:00:00 - 23:59:59","type":["ssl","siteEndTime","panelPwdEndTime","panelUpdate"]},"resourceAlert":{"sendTimeRange":"00:00:00 - 23:59:59","type":["clams","cronJob","cpu","memory","load","disk"]}}}`,
				CreateUser: "system",
				UpdateUser: "system",
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

var MigrateAlertMethodConfigIDs = &gormigrate.Migration{
	ID: "20251001-migrate-alert-method-config-ids",
	Migrate: func(tx *gorm.DB) error {
		if err := global.AlertDB.AutoMigrate(&model.Alert{}, &model.AlertLog{}, &model.AlertTask{}, &model.AlertConfig{}); err != nil {
			return err
		}
		if err := migrateAlertMethodConfigIDs(tx); err != nil {
			return err
		}
		return nil
	},
}

var MigrateAlertLogTaskMethodConfigIDs = &gormigrate.Migration{
	ID: "20260608-migrate-alert-log-task-method-config-ids",
	Migrate: func(tx *gorm.DB) error {
		if err := global.AlertDB.AutoMigrate(&model.AlertLog{}, &model.AlertTask{}, &model.AlertConfig{}); err != nil {
			return err
		}
		if err := migrateAlertMethodRecords(tx, &model.AlertLog{}); err != nil {
			return err
		}
		if err := migrateAlertMethodRecords(tx, &model.AlertTask{}); err != nil {
			return err
		}
		return nil
	},
}

var AddAlertAuditUser = &gormigrate.Migration{
	ID: "20260602-add-alert-audit-user",
	Migrate: func(tx *gorm.DB) error {
		return global.AlertDB.AutoMigrate(&model.Alert{}, &model.AlertConfig{})
	},
}

func migrateAlertMethodConfigIDs(tx *gorm.DB) error {
	if err := tx.Model(&model.AlertConfig{}).Where("type = ?", "mail").Update("type", constant.EmailConfig).Error; err != nil {
		return err
	}

	configIDs, err := loadAlertConfigIDs(tx)
	if err != nil {
		return err
	}

	var alerts []model.Alert
	if err := tx.Find(&alerts).Error; err != nil {
		return err
	}
	for _, alert := range alerts {
		method := migrateAlertMethodValue(alert.Method, alertLegacyMethodTypeMap(), configIDs)
		if method == alert.Method {
			continue
		}
		if err := tx.Model(&model.Alert{}).Where("id = ?", alert.ID).Update("method", method).Error; err != nil {
			return err
		}
	}
	return nil
}

func migrateAlertMethodRecords(tx *gorm.DB, modelValue interface{}) error {
	configIDs, err := loadAlertConfigIDs(tx)
	if err != nil {
		return err
	}

	switch modelValue.(type) {
	case *model.AlertLog:
		var logs []model.AlertLog
		if err := tx.Find(&logs).Error; err != nil {
			return err
		}
		for _, item := range logs {
			method := migrateAlertMethodValue(item.Method, alertLegacyMethodTypeMap(), configIDs)
			if method == item.Method {
				continue
			}
			if err := tx.Model(&model.AlertLog{}).Where("id = ?", item.ID).Update("method", method).Error; err != nil {
				return err
			}
		}
	case *model.AlertTask:
		var tasks []model.AlertTask
		if err := tx.Find(&tasks).Error; err != nil {
			return err
		}
		for _, item := range tasks {
			method := migrateAlertMethodValue(item.Method, alertLegacyMethodTypeMap(), configIDs)
			if method == item.Method {
				continue
			}
			if err := tx.Model(&model.AlertTask{}).Where("id = ?", item.ID).Update("method", method).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func loadAlertConfigIDs(tx *gorm.DB) (map[string]string, error) {
	configIDs := map[string]string{}
	for _, configType := range alertLegacyMethodTypeMap() {
		if _, ok := configIDs[configType]; ok {
			continue
		}
		var config model.AlertConfig
		if err := tx.Where("type = ?", configType).Order("id ASC").First(&config).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}
		configIDs[configType] = strconv.Itoa(int(config.ID))
	}
	return configIDs, nil
}

func alertLegacyMethodTypeMap() map[string]string {
	return map[string]string{
		"mail":            constant.Email,
		constant.Email:    constant.Email,
		constant.SMS:      constant.SMS,
		constant.Bark:     constant.Bark,
		constant.WeChat:   constant.WeCom,
		constant.WeCom:    constant.WeCom,
		constant.DingTalk: constant.DingTalk,
		constant.FeiShu:   constant.FeiShu,
	}
}

func migrateAlertMethodValue(method string, typeMap map[string]string, configIDs map[string]string) string {
	items := strings.Split(method, ",")
	next := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		configType, ok := typeMap[item]
		if ok {
			if id, exists := configIDs[configType]; exists {
				item = id
			}
		}
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		next = append(next, item)
	}
	return strings.Join(next, ",")
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

var UpdateMcpServerGatewayConfig = &gormigrate.Migration{
	ID: "20260615-update-mcp-server-gateway-config",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.McpServer{}); err != nil {
			return err
		}
		if err := tx.Model(&model.McpServer{}).Where("gateway_image IS NULL OR gateway_image = ''").Where("type = ?", "uvx").Update("gateway_image", "supercorp/supergateway:uvx").Error; err != nil {
			return err
		}
		if err := tx.Model(&model.McpServer{}).Where("gateway_image IS NULL OR gateway_image = ''").Where("type <> ? OR type IS NULL OR type = ''", "uvx").Update("gateway_image", "supercorp/supergateway:3.4.3").Error; err != nil {
			return err
		}
		if err := tx.Model(&model.McpServer{}).Where("protocol_version IS NULL OR protocol_version = ''").Update("protocol_version", "2025-06-18").Error; err != nil {
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
			_ = service.NewISSHService().CreateRootCert(dto.RootCertOperate{EncryptionMode: "ed25519", Name: "id_ed25519_1panel", Description: "1Panel Terminal"})
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

var InitRecordStatus = &gormigrate.Migration{
	ID: "20250910-init-record-status",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.BackupRecord{}); err != nil {
			return err
		}
		if err := tx.Model(&model.BackupRecord{}).Where("1 == 1").Updates(map[string]interface{}{"status": constant.StatusSuccess}).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddShowNameForQuickJump = &gormigrate.Migration{
	ID: "20250918-add-show-name-for-quick-jump",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.QuickJump{})
	},
}

var AddAgentQuickJump = &gormigrate.Migration{
	ID: "20260312-add-agent-quick-jump",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.QuickJump{}); err != nil {
			return err
		}

		var quicks []model.QuickJump
		if err := tx.Find(&quicks).Error; err != nil {
			return err
		}

		var (
			cronjob  *model.QuickJump
			database *model.QuickJump
			showList []*model.QuickJump
		)
		for i := range quicks {
			switch quicks[i].Name {
			case "Cronjob":
				cronjob = &quicks[i]
			case "Database":
				database = &quicks[i]
			}
			if quicks[i].IsShow {
				showList = append(showList, &quicks[i])
			}
		}

		showCount := len(showList)
		updatedIDs := make(map[uint]struct{})
		kickedCronjob := false
		if showCount >= 4 && cronjob != nil && cronjob.IsShow {
			cronjob.IsShow = false
			updatedIDs[cronjob.ID] = struct{}{}
			kickedCronjob = true
		}
		if !kickedCronjob && showCount >= 4 && database != nil && database.IsShow {
			database.IsShow = false
			updatedIDs[database.ID] = struct{}{}
		}

		for _, item := range quicks {
			if _, ok := updatedIDs[item.ID]; !ok {
				continue
			}
			if err := tx.Model(&model.QuickJump{}).Where("id = ?", item.ID).Update("is_show", item.IsShow).Error; err != nil {
				return err
			}
		}

		return tx.Create(&model.QuickJump{
			Name:      "Agent",
			Title:     "aiTools.agents.agent",
			Recommend: 1,
			IsShow:    true,
			Router:    "/ai/agents/agent",
		}).Error
	},
}

var AddTimeoutForClam = &gormigrate.Migration{
	ID: "20250922-add-timeout-for-clam",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Clam{}); err != nil {
			return err
		}
		if err := tx.Model(&model.Clam{}).Where("1 == 1").Updates(map[string]interface{}{"timeout": 18000}).Error; err != nil {
			return err
		}
		return nil
	},
}

var UpdateCronjobSpec = &gormigrate.Migration{
	ID: "20250925-update-cronjob-spec",
	Migrate: func(tx *gorm.DB) error {
		var cronjobs []model.Cronjob
		if err := tx.Where("1 == 1").Find(&cronjobs).Error; err != nil {
			return err
		}
		for _, item := range cronjobs {
			if !strings.Contains(item.Spec, ",") {
				continue
			}
			if err := tx.Model(&model.Cronjob{}).Where("id = ?", item.ID).Updates(
				map[string]interface{}{"spec": strings.ReplaceAll(item.Spec, ",", "&&")}).Error; err != nil {
				return err
			}
		}
		return nil
	},
}

var UpdateWebsiteSSLAddColumn = &gormigrate.Migration{
	ID: "20250928-update-website-ssl",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.WebsiteSSL{}); err != nil {
			return err
		}
		return nil
	},
}

var AddTensorRTLLMModel = &gormigrate.Migration{
	ID: "20251018-add-tensorrt-llm-model",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.TensorRTLLM{})
	},
}

var UpdateMonitorInterval = &gormigrate.Migration{
	ID: "20251026-update-monitor-interval",
	Migrate: func(tx *gorm.DB) error {
		var monitorInterval model.Setting
		if err := tx.Where("key = ?", "MonitorInterval").First(&monitorInterval).Error; err != nil {
			return err
		}
		interval, _ := strconv.Atoi(monitorInterval.Value)
		if interval == 0 {
			interval = 300
		}
		if err := tx.Model(&model.Setting{}).
			Where("key = ?", "MonitorInterval").
			Updates(map[string]interface{}{"value": fmt.Sprintf("%v", interval*60)}).
			Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "DefaultIO", Value: "all"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddIptablesFilterRuleTable = &gormigrate.Migration{
	ID: "20251106-add-iptables-filter-rule-table",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Firewall{}); err != nil {
			return err
		}
		var firewalls []model.Firewall
		_ = tx.Where("1 = 1").Find(&firewalls).Error

		firewallType := ""
		client, err := firewall.NewFirewallClient()
		if err == nil {
			firewallType = client.Name()
		}
		for _, item := range firewalls {
			if err := tx.Model(&model.Firewall{}).
				Where("id = ?", item.ID).
				Updates(map[string]interface{}{"dst_port": item.Port, "src_ip": item.Address, "firewall_type": firewallType}); err != nil {
				global.LOG.Errorf("update firewall failed, err: %v", err)
			}
		}
		return nil
	},
}

var AddMonitorProcess = &gormigrate.Migration{
	ID: "20251030-add-monitor-process",
	Migrate: func(tx *gorm.DB) error {
		return global.MonitorDB.AutoMigrate(&model.MonitorBase{})
	},
}

var UpdateCronJob = &gormigrate.Migration{
	ID: "20251105-update-cronjob",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.Cronjob{})
	},
}

var UpdateTensorrtLLM = &gormigrate.Migration{
	ID: "20251110-update-tensorrt-llm",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.TensorRTLLM{})
	},
}

var AddCommonDescription = &gormigrate.Migration{
	ID: "20251117-add-common-description",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.CommonDescription{})
	},
}

var UpdateDatabase = &gormigrate.Migration{
	ID: "20251117-update-database",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.Database{})
	},
}

var AddGPUMonitor = &gormigrate.Migration{
	ID: "20251122-add-gpu-monitor",
	Migrate: func(tx *gorm.DB) error {
		return global.GPUMonitorDB.AutoMigrate(&model.MonitorGPU{})
	},
}

var UpdateDatabaseMysql = &gormigrate.Migration{
	ID: "20251126-update-database-mysql",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.DatabaseMysql{}); err != nil {
			return err
		}
		var data []model.DatabaseMysql
		_ = tx.Where("1 = 1").Find(&data).Error
		for _, item := range data {
			if len(item.Collation) == 0 {
				collation := ""
				switch item.Format {
				case "utf8":
					collation = "utf8_general_ci"
				case "utf8mb4":
					collation = "utf8mb4_general_ci"
				case "gbk":
					collation = "gbk_chinese_ci"
				case "big5":
					collation = "big5_chinese_ci"
				default:
					collation = "utf8mb4_general_ci"
				}
				_ = tx.Model(&model.DatabaseMysql{}).Where("id = ?", item.ID).Updates(map[string]interface{}{"collation": collation}).Error
			}
		}
		return nil
	},
}

var AddDatabaseMongodb = &gormigrate.Migration{
	ID: "20260413-add-database-mongodb",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.DatabaseMongodb{})
	},
}

var InitIptablesStatus = &gormigrate.Migration{
	ID: "20251201-init-iptables-status",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "IptablesStatus", Value: constant.StatusDisable}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "IptablesForwardStatus", Value: constant.StatusDisable}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "IptablesInputStatus", Value: constant.StatusDisable}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "IptablesOutputStatus", Value: constant.StatusDisable}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: constant.FirewallPortWhiteList, Value: constant.FirewallPortWhiteListValue}).Error; err != nil {
			return err
		}
		return nil
	},
}

var InitFirewallPortWhiteList = &gormigrate.Migration{
	ID: "20260601-init-firewall-port-whitelist",
	Migrate: func(tx *gorm.DB) error {
		var setting model.Setting
		if err := tx.Where("key = ?", constant.FirewallPortWhiteList).First(&setting).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return tx.Create(&model.Setting{Key: constant.FirewallPortWhiteList, Value: constant.FirewallPortWhiteListValue}).Error
			}
			return err
		}
		return nil
	},
}

var UpdateWebsite = &gormigrate.Migration{
	ID: "20251203-update-website",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.Website{})
	},
}

var AddisIPtoWebsiteSSL = &gormigrate.Migration{
	ID: "20251223-update-website-ssl",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.WebsiteSSL{})
	},
}

var InitPingStatus = &gormigrate.Migration{
	ID: "20251201-init-ping-status",
	Migrate: func(tx *gorm.DB) error {
		status := firewall.LoadPingStatus()
		if err := tx.Create(&model.Setting{Key: "BanPing", Value: status}).Error; err != nil {
			return err
		}
		return nil
	},
}

var UpdateApp = &gormigrate.Migration{
	ID: "20251228-update-app",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.App{})
	},
}

var AddCronjobArgs = &gormigrate.Migration{
	ID: "20260106-add-cronjob-args",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.Cronjob{})
	},
}

var AddWebsiteAcmeAccountColumn = &gormigrate.Migration{
	ID: "20260110-add-website-acme-account",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.WebsiteAcmeAccount{})
	},
}

var AddAgentTables = &gormigrate.Migration{
	ID: "20260205-add-agent-tables",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			&model.Agent{},
			&model.AgentAccount{},
		)
	},
}

var AddAgentCustomModelFields = &gormigrate.Migration{
	ID: "20260224-add-agent-custom-model-fields",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Agent{}, &model.AgentAccount{}); err != nil {
			return err
		}
		if err := tx.Model(&model.AgentAccount{}).Where("api_type = '' OR api_type IS NULL").Update("api_type", "openai-completions").Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Agent{}).Where("api_type = '' OR api_type IS NULL").Update("api_type", "openai-completions").Error; err != nil {
			return err
		}
		return nil
	},
}

var AddAppInstallSortOrder = &gormigrate.Migration{
	ID: "20260222-add-app-install-sort-order",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.AppInstall{})
	},
}

var AddAgentAccountRememberAPIKey = &gormigrate.Migration{
	ID: "20260225-add-agent-account-remember-api-key",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.AgentAccount{}); err != nil {
			return err
		}
		return tx.Model(&model.AgentAccount{}).Where("remember_api_key IS NULL").Update("remember_api_key", true).Error
	},
}

var AddEditionSetting = &gormigrate.Migration{
	ID: "20260224-add-edition-setting",
	Migrate: func(tx *gorm.DB) error {
		var setting model.Setting
		edition := common.LoadParamsWithoutPanic("PANEL_EDITION")
		if err := tx.Where("key = ?", "Edition").First(&setting).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return tx.Create(&model.Setting{Key: "Edition", Value: edition}).Error
			}
			return err
		}
		if setting.Value == "" {
			return tx.Model(&model.Setting{}).Where("key = ?", "Edition").Update("value", edition).Error
		}
		return nil
	},
}

var AddAgentTypeForAgents = &gormigrate.Migration{
	ID: "20260302-add-agent-type-for-agents",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Agent{}); err != nil {
			return err
		}
		if err := tx.Model(&model.Agent{}).Where("agent_type = '' OR agent_type IS NULL").Update("agent_type", constant.AppOpenclaw).Error; err != nil {
			return err
		}
		return tx.Exec(
			"UPDATE agents SET agent_type = ? WHERE app_install_id IN (SELECT ai.id FROM app_installs ai JOIN apps a ON ai.app_id = a.id WHERE a.key = ?)",
			constant.AppCopaw,
			constant.AppCopaw,
		).Error
	},
}

var NormalizeAgentAccountVerifiedStatus = &gormigrate.Migration{
	ID: "20260303-normalize-agent-account-verified-status",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.AgentAccount{}); err != nil {
			return err
		}
		return tx.Model(&model.AgentAccount{}).
			Where("provider IN ?", []string{"custom", "ollama", "kimi-coding"}).
			Update("verified", false).Error
	},
}

var NormalizeOllamaAccountAPIType = &gormigrate.Migration{
	ID: "20260304-normalize-ollama-account-api-type",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.AgentAccount{}); err != nil {
			return err
		}
		return tx.Model(&model.AgentAccount{}).
			Where("provider = ?", "ollama").
			Update("api_type", "openai-responses").Error
	},
}

var InitAgentAccountModelPool = &gormigrate.Migration{
	ID: "20260319-init-agent-account-model-pool",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.AgentAccountModel{}); err != nil {
			return err
		}
		return migrationutils.MigrateAgentAccountModelPool(tx)
	},
}

var AddAgentAccountMasterID = &gormigrate.Migration{
	ID: "20260401-add-agent-account-master-id",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.AgentAccount{})
	},
}

var NormalizeAgentAccountModelIDs = &gormigrate.Migration{
	ID: "20260716-normalize-agent-account-model-ids",
	Migrate: func(tx *gorm.DB) error {
		return migrationutils.NormalizeAgentAccountModelIDs(tx)
	},
}

var AddAgentAccountVerifyModel = &gormigrate.Migration{
	ID: "20260716-add-agent-account-verify-model",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.AgentAccount{}); err != nil {
			return err
		}
		var accounts []model.AgentAccount
		if err := tx.Where("verify_model = '' OR verify_model IS NULL").Find(&accounts).Error; err != nil {
			return err
		}
		for _, account := range accounts {
			var accountModel model.AgentAccountModel
			err := tx.Where("account_id = ?", account.ID).Order("sort_order ASC, id ASC").First(&accountModel).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			if err != nil {
				return err
			}
			if err := tx.Model(&model.AgentAccount{}).Where("id = ?", account.ID).Update("verify_model", accountModel.Model).Error; err != nil {
				return err
			}
		}
		return nil
	},
}

var AddAgentAccountAuthMode = &gormigrate.Migration{
	ID: "20260716-add-agent-account-auth-mode",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.AgentAccount{}); err != nil {
			return err
		}
		if err := tx.Model(&model.AgentAccount{}).
			Where("api_type = ? AND (auth_mode IS NULL OR auth_mode = '') AND provider IN ?", "anthropic-messages", []string{"bailian-coding-plan", "ark-coding-plan", "xiaomi"}).
			Update("auth_mode", providercatalog.AuthModeBearer).Error; err != nil {
			return err
		}
		return tx.Model(&model.AgentAccount{}).
			Where("api_type = ? AND (auth_mode IS NULL OR auth_mode = '')", "anthropic-messages").
			Update("auth_mode", providercatalog.AuthModeXAPIKey).Error
	},
}

var AddHostTable = &gormigrate.Migration{
	ID: "20260318-add-host-table",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Host{}); err != nil {
			return err
		}
		if global.CoreDB == nil || !global.CoreDB.Migrator().HasTable("hosts") {
			if err := tx.Create(&model.Group{Name: "Default", Type: "host", IsDefault: true}).Error; err != nil {
				return err
			}
			return nil
		}

		var encryptSetting model.Setting
		if err := global.CoreDB.Where("key = ?", "EncryptKey").First(&encryptSetting).Error; err != nil {
			global.LOG.Errorf("failed to get encrypt key from core db, err: %v", err)
			return nil
		}
		coreEncryptKey := strings.TrimSpace(encryptSetting.Value)
		if coreEncryptKey == "" {
			global.LOG.Error("encrypt key from core db is empty")
			return nil
		}

		groupIDMap := make(map[uint]uint)
		defaultGroupID := uint(0)
		var coreGroups []model.Group
		if err := global.CoreDB.Where("type = ?", "host").Order("id asc").Find(&coreGroups).Error; err != nil {
			return err
		}
		for _, coreGroup := range coreGroups {
			agentGroup := model.Group{
				Name:      coreGroup.Name,
				Type:      "host",
				IsDefault: coreGroup.IsDefault,
			}
			if agentGroup.IsDefault {
				defaultGroupID = coreGroup.ID
			}
			if err := tx.Create(&agentGroup).Error; err != nil {
				global.LOG.Errorf("failed to create group, group id: %v, err: %v", coreGroup.ID, err)
				continue
			}
			groupIDMap[coreGroup.ID] = agentGroup.ID
		}

		var coreHosts []model.Host
		if err := global.CoreDB.Order("id asc").Find(&coreHosts).Error; err != nil {
			return err
		}
		for _, coreHost := range coreHosts {
			password, err := encrypt.StringDecryptWithKey(coreHost.Password, coreEncryptKey)
			if err != nil {
				global.LOG.Errorf("failed to decrypt host password, host id: %v, err: %v", coreHost.ID, err)
				continue
			}
			privateKey, err := encrypt.StringDecryptWithKey(coreHost.PrivateKey, coreEncryptKey)
			if err != nil {
				global.LOG.Errorf("failed to decrypt host private key, host id: %v, err: %v", coreHost.ID, err)
				continue
			}
			passPhrase, err := encrypt.StringDecryptWithKey(coreHost.PassPhrase, coreEncryptKey)
			if err != nil {
				global.LOG.Errorf("failed to decrypt host pass phrase, host id: %v, err: %v", coreHost.ID, err)
				continue
			}

			encryptedPassword, err := encrypt.StringEncrypt(password)
			if err != nil {
				global.LOG.Errorf("failed to encrypt host password, host id: %v, err: %v", coreHost.ID, err)
				continue
			}
			encryptedPrivateKey, err := encrypt.StringEncrypt(privateKey)
			if err != nil {
				global.LOG.Errorf("failed to encrypt host private key, host id: %v, err: %v", coreHost.ID, err)
				continue
			}
			encryptedPassPhrase, err := encrypt.StringEncrypt(passPhrase)
			if err != nil {
				global.LOG.Errorf("failed to encrypt host pass phrase, host id: %v, err: %v", coreHost.ID, err)
				continue
			}

			groupID := defaultGroupID
			if mappedGroupID, ok := groupIDMap[coreHost.GroupID]; ok && mappedGroupID != 0 {
				groupID = mappedGroupID
			}
			host := model.Host{
				GroupID:          groupID,
				Name:             coreHost.Name,
				Addr:             coreHost.Addr,
				Port:             coreHost.Port,
				User:             coreHost.User,
				AuthMode:         coreHost.AuthMode,
				Password:         encryptedPassword,
				PrivateKey:       encryptedPrivateKey,
				PassPhrase:       encryptedPassPhrase,
				RememberPassword: coreHost.RememberPassword,
				Description:      coreHost.Description,
			}
			if err := tx.Create(&host).Error; err != nil {
				global.LOG.Errorf("failed to create host, host id: %v, err: %v", coreHost.ID, err)
				continue
			}
		}
		return nil
	},
}

var AddAITerminalSettings = &gormigrate.Migration{
	ID: "20260318-add-ai-terminal-settings",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "AIStatus", Value: constant.StatusDisable}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "AIAccountID", Value: ""}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "AIPrefix", Value: constant.DefaultTerminalAIPrefix}).Error; err != nil {
			return err
		}
		return tx.Create(&model.Setting{
			Key:   "AIRiskCommands",
			Value: constant.DefaultTerminalAIRiskCommands,
		}).Error
	},
}

var UpdateAgentQuickJumpTitle = &gormigrate.Migration{
	ID: "20260324-update-agent-quick-jump-title",
	Migrate: func(tx *gorm.DB) error {
		return tx.Model(&model.QuickJump{}).
			Where("title = ?", "aiTools.agents.agents").
			Update("title", "aiTools.agents.agent").Error
	},
}

var FixOpenclaw20260323HTTPPort = &gormigrate.Migration{
	ID: "20260325-fix-openclaw-20260323-http-port",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec(
			`UPDATE app_installs
SET http_port = https_port,
    https_port = 0
WHERE version = ?
  AND https_port > 0
  AND app_id IN (SELECT id FROM apps WHERE key = ?)`,
			"2026.3.24",
			constant.AppOpenclaw,
		).Error
	},
}

var AddAgentRemarkColumn = &gormigrate.Migration{
	ID: "20260330-add-agent-remark-column",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.Agent{})
	},
}

var AddAgentWebsiteBinding = &gormigrate.Migration{
	ID: "20260403-add-agent-website-binding",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Agent{}); err != nil {
			return err
		}

		var agents []model.Agent
		if err := tx.Find(&agents).Error; err != nil {
			return err
		}
		if len(agents) == 0 {
			return nil
		}

		var websites []model.Website
		if err := tx.Where("type = ? AND app_install_id > 0", constant.Deployment).Find(&websites).Error; err != nil {
			return err
		}
		websiteMap := service.UniqueDeploymentWebsiteMapByAppInstall(websites)
		for _, agent := range agents {
			if agent.WebsiteID != 0 || agent.AppInstallID == 0 {
				continue
			}
			website, ok := websiteMap[agent.AppInstallID]
			if !ok {
				continue
			}
			if err := tx.Model(&model.Agent{}).Where("id = ?", agent.ID).Update("website_id", website.ID).Error; err != nil {
				return err
			}
		}
		return nil
	},
}

var AddFileManageAISettings = &gormigrate.Migration{
	ID: "20260330-add-file-manage-ai-settings",
	Migrate: func(tx *gorm.DB) error {
		rows := []model.Setting{
			{Key: "FileAIStatus", Value: constant.StatusDisable},
			{Key: "FileAIAccountID", Value: ""},
		}
		for i := range rows {
			var exist model.Setting
			if err := tx.Where("`key` = ?", rows[i].Key).First(&exist).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					if err := tx.Create(&rows[i]).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}
		}
		return nil
	},
}

var AddFileShareTable = &gormigrate.Migration{
	ID: "20260410-add-file-share-table",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.FileShare{})
	},
}

var AddFileHistoryTable = &gormigrate.Migration{
	ID: "20260414-add-file-history-table",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.FileHistory{}); err != nil {
			return err
		}
		defaultRows := []model.Setting{
			{Key: "FileHistoryStatus", Value: constant.StatusEnable},
			{Key: "FileHistoryMaxPerPath", Value: "20"},
			{Key: "FileHistoryDiskQuotaMB", Value: "1024"},
		}
		for i := range defaultRows {
			var exist model.Setting
			if err := tx.Where("`key` = ?", defaultRows[i].Key).First(&exist).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					if err := tx.Create(&defaultRows[i]).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}
		}
		return nil
	},
}

func loadDatabaseUserTypesForMigration(tx *gorm.DB, mysqlName string) []string {
	var database model.Database
	if err := tx.Where("name = ?", mysqlName).First(&database).Error; err == nil && len(database.Type) != 0 {
		return []string{database.Type}
	}

	var appKey string
	_ = tx.Table("app_installs").
		Select("apps.`key`").
		Joins("JOIN apps ON apps.id = app_installs.app_id").
		Where("app_installs.name = ?", mysqlName).
		Scan(&appKey).Error
	switch appKey {
	case constant.AppMariaDB, constant.AppMysqlCluster:
		return []string{appKey}
	case constant.AppMysql:
		return []string{constant.AppMysql}
	default:
		return []string{constant.AppMysql, constant.AppMariaDB}
	}
}

func normalizeDatabaseUserHostsForMigration(permission string) []string {
	hostSet := make(map[string]struct{})
	for _, host := range strings.Split(permission, ",") {
		host = strings.TrimSpace(host)
		if len(host) == 0 {
			continue
		}
		hostSet[host] = struct{}{}
	}
	if len(hostSet) == 0 {
		return []string{"%"}
	}
	hosts := make([]string, 0, len(hostSet))
	for host := range hostSet {
		hosts = append(hosts, host)
	}
	return hosts
}

func isDatabaseSystemUserForMigration(username string) bool {
	switch strings.ToLower(username) {
	case "root",
		"mysql.session", "mysql.sys", "mysql.infoschema", "mysqlxsys",
		"mariadb.sys", "mariadb-sys",
		"debian-sys-maint":
		return true
	default:
		return false
	}
}

func migrateDatabaseUsers(tx *gorm.DB) error {
	var mysqls []model.DatabaseMysql
	if err := tx.Find(&mysqls).Error; err != nil {
		return err
	}
	for _, item := range mysqls {
		if len(item.Username) == 0 || len(item.Password) == 0 || isDatabaseSystemUserForMigration(item.Username) {
			continue
		}
		for _, dbType := range loadDatabaseUserTypesForMigration(tx, item.MysqlName) {
			if len(dbType) == 0 {
				continue
			}
			for _, host := range normalizeDatabaseUserHostsForMigration(item.Permission) {
				var old model.DatabaseUser
				err := tx.Where("`type` = ? AND database = ? AND username = ? AND host = ?", dbType, item.MysqlName, item.Username, host).First(&old).Error
				if err == nil {
					continue
				}
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				if err := tx.Create(&model.DatabaseUser{
					Type:     dbType,
					Database: item.MysqlName,
					Username: item.Username,
					Host:     host,
					Password: item.Password,
				}).Error; err != nil {
					return err
				}
			}
		}
	}
	return nil
}

var AddDatabaseUserTable = &gormigrate.Migration{
	ID: "20260703-add-database-user-table",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.DatabaseUser{}); err != nil {
			return err
		}
		return migrateDatabaseUsers(tx)
	},
}

// MigrateLegoV5 normalizes data persisted under lego v4 so that lego v5 can read it.
//
// Two things changed in lego v5 that affect existing rows:
//
//  1. certcrypto.KeyType string values were renamed:
//     "P256"->"EC256", "P384"->"EC384",
//     "2048"->"RSA2048", "3072"->"RSA3072", "4096"->"RSA4096", "8192"->"RSA8192".
//     We update every key_type column we own to the new form.
//
//  2. DnsPod provider was removed from upstream lego v5; the frontend already
//     marked it deprecated. Existing DnsPod website_dns_accounts rows are kept
//     so the user can decide what to do, but any website_ssls still pointing
//     at a DnsPod account would fail to renew. We log a warning row count and
//     leave deletion to the operator.
var MigrateLegoV5 = &gormigrate.Migration{
	ID: "20260523-migrate-lego-v5",
	Migrate: func(tx *gorm.DB) error {
		keyTypeMap := map[string]string{
			"P256": "EC256",
			"P384": "EC384",
			"2048": "RSA2048",
			"3072": "RSA3072",
			"4096": "RSA4096",
			"8192": "RSA8192",
		}
		for old, neu := range keyTypeMap {
			if err := tx.Model(&model.WebsiteAcmeAccount{}).
				Where("key_type = ?", old).
				Update("key_type", neu).Error; err != nil {
				return fmt.Errorf("migrate WebsiteAcmeAccount.key_type %s->%s: %w", old, neu, err)
			}
			if err := tx.Model(&model.WebsiteSSL{}).
				Where("key_type = ?", old).
				Update("key_type", neu).Error; err != nil {
				return fmt.Errorf("migrate WebsiteSSL.key_type %s->%s: %w", old, neu, err)
			}
			if err := tx.Model(&model.WebsiteCA{}).
				Where("key_type = ?", old).
				Update("key_type", neu).Error; err != nil {
				return fmt.Errorf("migrate WebsiteCA.key_type %s->%s: %w", old, neu, err)
			}
		}

		var dnsPodCount int64
		_ = tx.Model(&model.WebsiteDnsAccount{}).Where("type = ?", "DnsPod").Count(&dnsPodCount).Error
		if dnsPodCount > 0 {
			global.LOG.Warnf("lego v5 removed the DnsPod provider; %d existing DnsPod DNS account(s) will not be usable for renewal -- please switch them to TencentCloud", dnsPodCount)
		}

		return nil
	},
}

var AddMcpServerGatewayArgs = &gormigrate.Migration{
	ID: "20260714-add-mcp-server-gateway-args",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&model.McpServer{})
	},
}
