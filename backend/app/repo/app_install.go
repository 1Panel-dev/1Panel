package repo

import (
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/global"
)

type AppInstallRepo struct{}

func (a AppInstallRepo) GetBy(opts ...DBOption) ([]model.AppInstall, error) {
	db := global.DB.Model(&model.AppInstall{})
	for _, opt := range opts {
		db = opt(db)
	}
	var install []model.AppInstall
	err := db.Preload("App").Find(&install).Error
	return install, err
}

func (a AppInstallRepo) Create(install model.AppInstall) error {
	db := global.DB.Model(&model.AppInstall{})
	return db.Create(&install).Error
}
func (a AppInstallRepo) Save(install model.AppInstall) error {
	db := global.DB.Model(&model.AppInstall{})
	return db.Save(&install).Error
}

func (a AppInstallRepo) Page(page, size int, opts ...DBOption) (int64, []model.AppInstall, error) {
	var apps []model.AppInstall
	db := global.DB.Model(&model.AppInstall{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Debug().Limit(size).Offset(size * (page - 1)).Preload("App").Find(&apps).Error
	return count, apps, err
}
