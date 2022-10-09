package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/global"
	"gorm.io/gorm"
)

type AppInstallRepo struct{}

func (a AppInstallRepo) WithDetailIdsIn(detailIds []uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("app_detail_id in (?)", detailIds)
	}
}
func (a AppInstallRepo) WithAppId(appId uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("app_id = ?", appId)
	}
}
func (a AppInstallRepo) WithStatus(status string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("status = ?", status)
	}
}

func (a AppInstallRepo) GetBy(opts ...DBOption) ([]model.AppInstall, error) {
	db := global.DB.Model(&model.AppInstall{})
	for _, opt := range opts {
		db = opt(db)
	}
	var install []model.AppInstall
	err := db.Preload("App").Preload("Containers").Find(&install).Error
	return install, err
}

func (a AppInstallRepo) GetFirst(opts ...DBOption) (model.AppInstall, error) {
	db := global.DB.Model(&model.AppInstall{})
	for _, opt := range opts {
		db = opt(db)
	}
	var install model.AppInstall
	err := db.Preload("App").Preload("Containers").First(&install).Error
	return install, err
}

func (a AppInstallRepo) Create(ctx context.Context, install *model.AppInstall) error {
	db := ctx.Value("db").(*gorm.DB).Model(&model.AppInstall{})
	return db.Create(&install).Error
}

func (a AppInstallRepo) Save(install model.AppInstall) error {
	db := global.DB
	return db.Save(&install).Error
}

func (a AppInstallRepo) DeleteBy(opts ...DBOption) error {
	db := global.DB.Model(&model.AppInstall{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.AppInstall{}).Error
}

func (a AppInstallRepo) Delete(install model.AppInstall) error {
	db := global.DB
	return db.Delete(&install).Error
}

func (a AppInstallRepo) Page(page, size int, opts ...DBOption) (int64, []model.AppInstall, error) {
	var apps []model.AppInstall
	db := global.DB.Model(&model.AppInstall{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Debug().Limit(size).Offset(size * (page - 1)).Preload("App").Preload("Containers").Find(&apps).Error
	return count, apps, err
}

func (a AppInstallRepo) BatchUpdateBy(update model.AppInstall, opts ...DBOption) error {
	db := global.DB.Model(model.AppInstall{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Updates(update).Error
}
