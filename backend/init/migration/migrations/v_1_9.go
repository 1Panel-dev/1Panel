package migrations

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
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
	ID: "20231126-update-website-ssl",
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
