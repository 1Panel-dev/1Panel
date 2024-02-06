package migrations

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/app/service"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cloud_storage/client"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var UpdateAcmeAccount = &gormigrate.Migration{
	ID: "20231117-update-acme-account",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.WebsiteAcmeAccount{}); err != nil {
			return err
		}
		return nil
	},
}

var AddWebsiteCA = &gormigrate.Migration{
	ID: "20231125-add-website-ca",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.WebsiteCA{}); err != nil {
			return err
		}
		return nil
	},
}

var UpdateWebsiteSSL = &gormigrate.Migration{
	ID: "20231128-update-website-ssl",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.WebsiteSSL{}); err != nil {
			return err
		}
		return nil
	},
}

var AddDockerSockPath = &gormigrate.Migration{
	ID: "20231128-add-docker-sock-path",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "DockerSockPath", Value: "unix:///var/run/docker.sock"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddDatabaseSSL = &gormigrate.Migration{
	ID: "20231126-add-database-ssl",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Database{}); err != nil {
			return err
		}
		return nil
	},
}

var AddDefaultCA = &gormigrate.Migration{
	ID: "20231129-add-default-ca",
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

var AddSettingRecycleBin = &gormigrate.Migration{
	ID: "20231129-add-setting-recycle-bin",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Create(&model.Setting{Key: "FileRecycleBin", Value: "enable"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var UpdateWebsiteBackupRecord = &gormigrate.Migration{
	ID: "20231218-update-backup-record-for-website",
	Migrate: func(tx *gorm.DB) error {
		backupRepo := repo.NewIBackupRepo()
		websitesBackups, _ := backupRepo.ListRecord(repo.NewCommonRepo().WithByType("website"))
		if len(websitesBackups) > 0 {
			for _, backup := range websitesBackups {
				backup.DetailName = backup.Name
				_ = backupRepo.UpdateRecord(&backup)
			}
		}
		return nil
	},
}

var AddTablePHPExtensions = &gormigrate.Migration{
	ID: "20240102-add-php-extensions",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.PHPExtensions{}); err != nil {
			return err
		}
		if err := tx.Create(&model.PHPExtensions{Name: "默认", Extensions: "bcmath,gd,gettext,intl,pcntl,shmop,soap,sockets,sysvsem,xmlrpc,zip"}).Error; err != nil {
			return err
		}
		return nil
	},
}

var AddTableDatabasePostgresql = &gormigrate.Migration{
	ID: "20231225-add-table-database_postgresql",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.DatabasePostgresql{}); err != nil {
			return err
		}
		if err := tx.AutoMigrate(&model.Cronjob{}); err != nil {
			return err
		}
		var jobs []model.Cronjob
		if err := tx.Where("type == ?", "database").Find(&jobs).Error; err != nil {
			return err
		}
		for _, job := range jobs {
			if job.DBName == "all" {
				if err := tx.Model(&model.Cronjob{}).Where("id = ?", job.ID).Update("db_type", "mysql").Error; err != nil {
					global.LOG.Errorf("update db type of cronjob %s failed, err: %v", job.Name, err)
					continue
				}
			}
			var db model.DatabaseMysql
			if err := tx.Where("id == ?", job.DBName).First(&db).Error; err != nil {
				continue
			}
			var database model.Database
			if err := tx.Where("name == ?", db.MysqlName).First(&database).Error; err != nil {
				continue
			}
			if err := tx.Model(&model.Cronjob{}).Where("id = ?", job.ID).Update("db_type", database.Type).Error; err != nil {
				global.LOG.Errorf("update db type of cronjob %s failed, err: %v", job.Name, err)
				continue
			}
		}
		return nil
	},
}

var AddPostgresqlSuperUser = &gormigrate.Migration{
	ID: "20231225-add-postgresql-super_user",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.DatabasePostgresql{}); err != nil {
			return err
		}
		return nil
	},
}

var UpdateCronjobWithWebsite = &gormigrate.Migration{
	ID: "20230809-update-cronjob-with-website",
	Migrate: func(tx *gorm.DB) error {
		var cronjobs []model.Cronjob
		if err := tx.Where("(type = ? OR type = ?) AND website != ?", "website", "cutWebsiteLog", "all").Find(&cronjobs).Error; err != nil {
			return err
		}

		for _, job := range cronjobs {
			var web model.Website
			if err := tx.Where("primary_domain = ?", job.Website).First(&web).Error; err != nil {
				continue
			}
			if err := tx.Model(&model.Cronjob{}).
				Where("id = ?", job.ID).
				Updates(map[string]interface{}{"website": web.ID}).Error; err != nil {
				continue
			}
		}

		return nil
	},
}

var UpdateOneDriveToken = &gormigrate.Migration{
	ID: "20240117-update-onedrive-token",
	Migrate: func(tx *gorm.DB) error {
		var (
			backup        model.BackupAccount
			clientSetting model.Setting
			secretSetting model.Setting
		)
		_ = tx.Where("type = ?", "OneDrive").First(&backup).Error
		if backup.ID == 0 {
			return nil
		}
		if len(backup.Credential) == 0 {
			global.LOG.Error("OneDrive configuration lacks token information, please rebind.")
			return nil
		}

		_ = tx.Where("key = ?", "OneDriveID").First(&clientSetting).Error
		if clientSetting.ID == 0 {
			global.LOG.Error("system configuration lacks clientID information, please retry.")
			return nil
		}
		_ = tx.Where("key = ?", "OneDriveSc").First(&secretSetting).Error
		if secretSetting.ID == 0 {
			global.LOG.Error("system configuration lacks clientID information, please retry.")
			return nil
		}
		idItem, _ := base64.StdEncoding.DecodeString(clientSetting.Value)
		global.CONF.System.OneDriveID = string(idItem)
		scItem, _ := base64.StdEncoding.DecodeString(secretSetting.Value)
		global.CONF.System.OneDriveSc = string(scItem)

		varMap := make(map[string]interface{})
		varMap["isCN"] = false
		varMap["client_id"] = global.CONF.System.OneDriveID
		varMap["client_secret"] = global.CONF.System.OneDriveSc
		varMap["redirect_uri"] = constant.OneDriveRedirectURI
		varMap["refresh_token"] = backup.Credential
		token, refreshToken, err := client.RefreshToken("refresh_token", varMap)
		varMap["refresh_status"] = constant.StatusSuccess
		varMap["refresh_time"] = time.Now().Format("2006-01-02 15:04:05")
		if err != nil {
			varMap["refresh_msg"] = err.Error()
			varMap["refresh_status"] = constant.StatusFailed
		}
		varMap["refresh_token"] = refreshToken
		itemVars, _ := json.Marshal(varMap)
		if err := tx.Model(&model.BackupAccount{}).
			Where("id = ?", backup.ID).
			Updates(map[string]interface{}{
				"credential": token,
				"vars":       string(itemVars),
			}).Error; err != nil {
			return err
		}

		return nil
	},
}

var UpdateCronjobSpec = &gormigrate.Migration{
	ID: "20240122-update-cronjob-spec",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Cronjob{}); err != nil {
			return err
		}
		if err := tx.AutoMigrate(&model.BackupRecord{}); err != nil {
			return err
		}
		var (
			jobs           []model.Cronjob
			backupAccounts []model.BackupAccount
		)
		mapAccount := make(map[uint]model.BackupAccount)
		if err := tx.Find(&jobs).Error; err != nil {
			return err
		}
		_ = tx.Find(&backupAccounts).Error
		for _, item := range backupAccounts {
			mapAccount[item.ID] = item
		}
		for _, job := range jobs {
			if job.KeepLocal && mapAccount[uint(job.TargetDirID)].Type != constant.Local {
				if err := tx.Model(&model.Cronjob{}).
					Where("id = ?", job.ID).
					Updates(map[string]interface{}{
						"backup_accounts":  fmt.Sprintf("%v,%v", mapAccount[uint(job.TargetDirID)].Type, constant.Local),
						"default_download": constant.Local,
					}).Error; err != nil {
					return err
				}
				job.DefaultDownload = constant.Local
			} else {
				if err := tx.Model(&model.Cronjob{}).
					Where("id = ?", job.ID).
					Updates(map[string]interface{}{
						"backup_accounts":  mapAccount[uint(job.TargetDirID)].Type,
						"default_download": mapAccount[uint(job.TargetDirID)].Type,
					}).Error; err != nil {
					return err
				}
				job.DefaultDownload = mapAccount[uint(job.TargetDirID)].Type
			}
			if job.Type != "directory" && job.Type != "database" && job.Type != "website" && job.Type != "app" && job.Type != "snapshot" && job.Type != "log" {
				continue
			}

			itemPath := mapAccount[uint(job.TargetDirID)].BackupPath
			if job.DefaultDownload == constant.Local || itemPath == "/" {
				itemPath = ""
			} else {
				itemPath = strings.TrimPrefix(itemPath, "/") + "/"
			}

			var records []model.JobRecords
			_ = tx.Where("cronjob_id = ?", job.ID).Find(&records).Error
			for _, record := range records {
				if job.Type == "snapshot" && record.Status == constant.StatusSuccess {
					var snaps []model.Snapshot
					_ = tx.Where("name like ?", "snapshot_"+"%").Find(&snaps).Error
					for _, snap := range snaps {
						item := model.BackupRecord{
							From:       "cronjob",
							CronjobID:  job.ID,
							Type:       "snapshot",
							Name:       job.Name,
							FileDir:    "system_snapshot",
							FileName:   snap.Name + ".tar.gz",
							Source:     snap.From,
							BackupType: snap.From,
							BaseModel: model.BaseModel{
								CreatedAt: record.CreatedAt,
							},
						}
						_ = tx.Create(&item).Error
					}
					continue
				}
				if job.Type == "log" && record.Status == constant.StatusSuccess {
					item := model.BackupRecord{
						From:       "cronjob",
						CronjobID:  job.ID,
						Type:       "log",
						Name:       job.Name,
						FileDir:    path.Dir(strings.TrimPrefix(record.File, itemPath)),
						FileName:   path.Base(record.File),
						Source:     mapAccount[uint(job.TargetDirID)].Type,
						BackupType: mapAccount[uint(job.TargetDirID)].Type,
						BaseModel: model.BaseModel{
							CreatedAt: record.CreatedAt,
						},
					}
					_ = tx.Create(&item).Error
					continue
				}
				if job.Type == "directory" && record.Status == constant.StatusSuccess {
					item := model.BackupRecord{
						From:       "cronjob",
						CronjobID:  job.ID,
						Type:       "directory",
						Name:       job.Name,
						FileDir:    path.Dir(strings.TrimPrefix(record.File, itemPath)),
						FileName:   path.Base(record.File),
						BackupType: mapAccount[uint(job.TargetDirID)].Type,
						BaseModel: model.BaseModel{
							CreatedAt: record.CreatedAt,
						},
					}
					if record.FromLocal {
						item.Source = constant.Local
					} else {
						item.Source = mapAccount[uint(job.TargetDirID)].Type
					}
					_ = tx.Create(&item).Error
					continue
				}
				if strings.Contains(record.File, ",") {
					files := strings.Split(record.File, ",")
					for _, file := range files {
						_ = tx.Model(&model.BackupRecord{}).
							Where("file_dir = ? AND file_name = ?", path.Dir(strings.TrimPrefix(file, itemPath)), path.Base(file)).
							Updates(map[string]interface{}{"cronjob_id": job.ID, "from": "cronjob"}).Error
					}
				} else {
					_ = tx.Model(&model.BackupRecord{}).
						Where("file_dir = ? AND file_name = ?", path.Dir(strings.TrimPrefix(record.File, itemPath)), path.Base(record.File)).
						Updates(map[string]interface{}{"cronjob_id": job.ID, "from": "cronjob"}).Error
				}
			}
		}

		_ = tx.Exec("ALTER TABLE cronjobs DROP COLUMN spec_type;").Error
		_ = tx.Exec("ALTER TABLE cronjobs DROP COLUMN week;").Error
		_ = tx.Exec("ALTER TABLE cronjobs DROP COLUMN day;").Error
		_ = tx.Exec("ALTER TABLE cronjobs DROP COLUMN hour;").Error
		_ = tx.Exec("ALTER TABLE cronjobs DROP COLUMN minute;").Error
		_ = tx.Exec("ALTER TABLE cronjobs DROP COLUMN second;").Error
		_ = tx.Exec("ALTER TABLE cronjobs DROP COLUMN entry_id;").Error

		return nil
	},
}

var UpdateBackupRecordPath = &gormigrate.Migration{
	ID: "20240124-update-cronjob-spec",
	Migrate: func(tx *gorm.DB) error {
		var (
			backupRecords []model.BackupRecord
			localAccount  model.BackupAccount
		)

		_ = tx.Where("type = ?", "LOCAL").First(&localAccount).Error
		if localAccount.ID == 0 {
			return nil
		}
		varMap := make(map[string]string)
		if err := json.Unmarshal([]byte(localAccount.Vars), &varMap); err != nil {
			return err
		}
		dir, ok := varMap["dir"]
		if !ok {
			return errors.New("load local backup dir failed")
		}
		if dir != "/" {
			dir += "/"
		}
		_ = tx.Where("source = ?", "LOCAL").Find(&backupRecords).Error
		for _, record := range backupRecords {
			_ = tx.Model(&model.BackupRecord{}).
				Where("id = ?", record.ID).
				Updates(map[string]interface{}{"file_dir": strings.TrimPrefix(record.FileDir, dir)}).Error
		}
		return nil
	},
}

var UpdateSnapshotRecords = &gormigrate.Migration{
	ID: "20240124-update-snapshot-records",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Snapshot{}); err != nil {
			return err
		}
		var snaps []model.Snapshot
		_ = tx.Find(&snaps).Error
		for _, snap := range snaps {
			_ = tx.Model(&model.Snapshot{}).
				Where("id = ?", snap.ID).
				Updates(map[string]interface{}{"default_download": snap.From}).Error
		}
		return nil
	},
}

var UpdateWebDavConf = &gormigrate.Migration{
	ID: "20240205-update-webdav-conf",
	Migrate: func(tx *gorm.DB) error {
		var backup model.BackupAccount
		_ = tx.Where("type = ?", constant.WebDAV).First(&backup).Error
		if backup.ID == 0 {
			return nil
		}
		varMap := make(map[string]interface{})
		if err := json.Unmarshal([]byte(backup.Vars), &varMap); err != nil {
			return err
		}
		delete(varMap, "addressItem")
		if port, ok := varMap["port"]; ok {
			varMap["address"] = fmt.Sprintf("%s:%v", varMap["address"], port)
			delete(varMap, "port")
		}

		vars, _ := json.Marshal(varMap)
		if err := tx.Model(&model.BackupAccount{}).Where("id = ?", backup.ID).Updates(map[string]interface{}{"vars": string(vars)}).Error; err != nil {
			return err
		}
		return nil
	},
}
