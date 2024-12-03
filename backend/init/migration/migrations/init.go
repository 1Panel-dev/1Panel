package migrations

import (
	"fmt"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/service"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
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
		language := "en"
		if global.CONF.System.Language == "zh" {
			language = "zh"
		}

		if err := tx.Create(&model.Setting{Key: "Password", Value: pass}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "Email", Value: ""}).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.Setting{Key: "PanelName", Value: "1Panel"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.Setting{Key: "Language", Value: language}).Error; err != nil {
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

		if err := tx.Create(&model.Setting{Key: "ExpirationTime", Value: time.Now().AddDate(0, 0, 10).Format(constant.DateTimeLayout)}).Error; err != nil {
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
		if err := tx.Create(&model.Setting{Key: "MonitorStoreDays", Value: "7"}).Error; err != nil {
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
			Name:      "Default",
			IsDefault: true,
			Type:      "website",
		}
		if err := tx.Create(defaultGroup).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Group{}).Where("name = ? AND type = ?", "default", "host").Update("name", "Default").Error; err != nil {
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
		if err := tx.Create(&model.Setting{Key: "MonitorInterval", Value: "5"}).Error; err != nil {
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

var AddRemoteDB = &gormigrate.Migration{
	ID: "20230724-add-remote-db",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Database{}, &model.DatabaseMysql{}); err != nil {
			return err
		}
		installRepo := repo.NewIAppInstallRepo()
		mysqlInfo, err := installRepo.LoadBaseInfo("mysql", "")
		if err == nil {
			if err := tx.Create(&model.Database{
				Name:     "local",
				Type:     "mysql",
				Version:  mysqlInfo.Version,
				From:     "local",
				Address:  "127.0.0.1",
				Username: "root",
				Password: mysqlInfo.Password,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	},
}

var UpdateRedisParam = &gormigrate.Migration{
	ID: "20230804-update-redis-param",
	Migrate: func(tx *gorm.DB) error {
		var (
			app        model.App
			appInstall model.AppInstall
		)
		if err := tx.Where("key = ?", "redis").First(&app).Error; err != nil {
			return nil
		}
		if err := tx.Where("app_id = ?", app.ID).First(&appInstall).Error; err != nil {
			return nil
		}
		appInstall.Param = strings.ReplaceAll(appInstall.Param, "PANEL_DB_ROOT_PASSWORD", "PANEL_REDIS_ROOT_PASSWORD")
		appInstall.DockerCompose = strings.ReplaceAll(appInstall.DockerCompose, "PANEL_DB_ROOT_PASSWORD", "PANEL_REDIS_ROOT_PASSWORD")
		appInstall.Env = strings.ReplaceAll(appInstall.Env, "PANEL_DB_ROOT_PASSWORD", "PANEL_REDIS_ROOT_PASSWORD")
		if err := tx.Model(&model.AppInstall{}).Where("id = ?", appInstall.ID).Updates(appInstall).Error; err != nil {
			return err
		}
		return nil
	},
}

var UpdateCronjobWithDb = &gormigrate.Migration{
	ID: "20230809-update-cronjob-with-db",
	Migrate: func(tx *gorm.DB) error {
		var cronjobs []model.Cronjob
		if err := tx.Where("type = ? AND db_name != ?", "database", "all").Find(&cronjobs).Error; err != nil {
			return nil
		}

		for _, job := range cronjobs {
			var db model.DatabaseMysql
			if err := tx.Where("name = ?", job.DBName).First(&db).Error; err != nil {
				continue
			}
			if err := tx.Model(&model.Cronjob{}).
				Where("id = ?", job.ID).
				Updates(map[string]interface{}{"db_name": db.ID}).Error; err != nil {
				continue
			}
		}
		return nil
	},
}

var AddTableFirewall = &gormigrate.Migration{
	ID: "20230908-add-table-firewall",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Firewall{}, model.SnapshotStatus{}, &model.Cronjob{}); err != nil {
			return err
		}
		_ = tx.Exec("alter table remote_dbs rename to databases;").Error
		if err := tx.AutoMigrate(&model.Database{}); err != nil {
			return err
		}
		return nil
	},
}

var AddDatabases = &gormigrate.Migration{
	ID: "20230831-add-databases",
	Migrate: func(tx *gorm.DB) error {
		installRepo := repo.NewIAppInstallRepo()
		_ = tx.Where("name = ? AND address = ?", "local", "127.0.0.1").Delete(&model.Database{}).Error
		mysql := addDatabaseData(tx, installRepo, "mysql")
		if mysql.AppInstallID != 0 {
			if err := tx.Create(mysql).Error; err != nil {
				return err
			}
		}
		mariadb := addDatabaseData(tx, installRepo, "mariadb")
		if mariadb.AppInstallID != 0 {
			if err := tx.Create(mariadb).Error; err != nil {
				return err
			}
		}
		redis := addDatabaseData(tx, installRepo, "redis")
		if redis.AppInstallID != 0 {
			if err := tx.Create(redis).Error; err != nil {
				return err
			}
		}
		postgresql := addDatabaseData(tx, installRepo, "postgresql")
		if postgresql.AppInstallID != 0 {
			if err := tx.Create(postgresql).Error; err != nil {
				return err
			}
		}
		mongodb := addDatabaseData(tx, installRepo, "mongodb")
		if mongodb.AppInstallID != 0 {
			if err := tx.Create(mongodb).Error; err != nil {
				return err
			}
		}
		memcached := addDatabaseData(tx, installRepo, "memcached")
		if memcached.AppInstallID != 0 {
			if err := tx.Create(memcached).Error; err != nil {
				return err
			}
		}

		return nil
	},
}

var UpdateDatabase = &gormigrate.Migration{
	ID: "20230831-update-database",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Model(&model.DatabaseMysql{}).Where("`from` != ?", "local").Updates(map[string]interface{}{
			"from": "remote",
		}).Error; err != nil {
			return err
		}

		var datas []model.Database
		if err := tx.Find(&datas).Error; err != nil {
			return nil
		}
		for _, data := range datas {
			pass, err := encrypt.StringEncrypt(data.Password)
			if err != nil {
				global.LOG.Errorf("encrypt database %s password failed, err: %v", data.Name, err)
				continue
			}
			if err := tx.Model(&model.Database{}).Where("id = ?", data.ID).Updates(map[string]interface{}{
				"password": pass,
			}).Error; err != nil {
				global.LOG.Errorf("updata database %s info failed, err: %v", data.Name, err)
			}
		}

		var mysqls []model.DatabaseMysql
		if err := tx.Find(&mysqls).Error; err != nil {
			return nil
		}
		for _, data := range mysqls {
			pass, err := encrypt.StringEncrypt(data.Password)
			if err != nil {
				global.LOG.Errorf("encrypt database db %s password failed, err: %v", data.Name, err)
				continue
			}
			if err := tx.Model(&model.DatabaseMysql{}).Where("id = ?", data.ID).Updates(map[string]interface{}{
				"password": pass,
			}).Error; err != nil {
				global.LOG.Errorf("updata database db %s info failed, err: %v", data.Name, err)
			}
		}
		return nil
	},
}

var UpdateAppInstallResource = &gormigrate.Migration{
	ID: "20230831-update-app_install_resource",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.AppInstallResource{}); err != nil {
			return err
		}
		if err := tx.Model(&model.AppInstallResource{}).Where("1 = 1").Updates(map[string]interface{}{
			"from": "local",
		}).Error; err != nil {
			return err
		}
		return nil
	},
}

var DropDatabaseLocal = &gormigrate.Migration{
	ID: "20230914-drop-database-local",
	Migrate: func(tx *gorm.DB) error {
		_ = tx.Where("name = ? AND address = ?", "local", "127.0.0.1").Delete(&model.Database{}).Error
		return nil
	},
}

func addDatabaseData(tx *gorm.DB, installRepo repo.IAppInstallRepo, appType string) *model.Database {
	dbInfo, err := installRepo.LoadBaseInfo(appType, "")
	if err != nil {
		return &model.Database{}
	}

	if appType == "mysql" || appType == "redis" || appType == "mariadb" || appType == "memcached" {
		dbInfo.UserName = "root"
	}
	database := &model.Database{
		AppInstallID: dbInfo.ID,
		Name:         dbInfo.Name,
		Type:         appType,
		Version:      dbInfo.Version,
		From:         "local",
		Address:      dbInfo.ServiceName,
		Port:         service.DatabaseKeys[appType],
		Username:     dbInfo.UserName,
		Password:     dbInfo.Password,
	}
	var dbItem model.Database
	_ = global.DB.Where("name = ?", dbInfo.Name).First(&dbItem).Error
	if dbItem.ID != 0 {
		if appType == "mysql" {
			var (
				backups []model.BackupRecord
				mysqls  []model.DatabaseMysql
			)
			_ = tx.Where("name = ? AND type = ?", dbItem.Name, "mysql").Find(&backups)
			_ = tx.Where("`from` = ?", "local").Find(&mysqls)
			for _, item := range backups {
				isLocal := false
				for _, mysql := range mysqls {
					if item.Name == mysql.MysqlName && item.DetailName == mysql.Name {
						isLocal = true
						break
					}
				}
				if !isLocal {
					_ = tx.Model(&model.BackupRecord{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
						"name": "remote-" + dbItem.Name,
					}).Error
				}
			}
		}
		if err := tx.Debug().Model(&model.DatabaseMysql{}).Where("mysql_name = ? AND `from` != ?", dbItem.Name, "local").Updates(map[string]interface{}{
			"mysql_name": "remote-" + dbItem.Name,
		}).Error; err != nil {
			fmt.Println(err)
		}
		if err := tx.Debug().Model(&model.Database{}).Where("name = ?", dbItem.Name).Updates(map[string]interface{}{
			"name": "remote-" + dbItem.Name,
		}).Error; err != nil {
			fmt.Println(err)
		}
	}
	return database
}
