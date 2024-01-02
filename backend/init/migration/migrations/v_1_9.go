package migrations

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/app/service"
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
