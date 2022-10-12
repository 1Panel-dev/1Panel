package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/app/model"
	"gorm.io/gorm"
)

type AppInstallRepo struct{}

func (a AppInstallRepo) WithDetailIdsIn(detailIds []uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("app_detail_id in (?)", detailIds)
	}
}

func (a AppInstallRepo) WithDetailIdNotIn(detailIds []uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("app_detail_id not in (?)", detailIds)
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

func (a AppInstallRepo) WithServiceName(serviceName string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("service_name = ?", serviceName)
	}
}

func (a AppInstallRepo) GetBy(opts ...DBOption) ([]model.AppInstall, error) {
	var install []model.AppInstall
	db := getDb(opts...).Model(&model.AppInstall{})
	err := db.Preload("App").Find(&install).Error
	return install, err
}

func (a AppInstallRepo) GetFirst(opts ...DBOption) (model.AppInstall, error) {
	var install model.AppInstall
	db := getDb(opts...).Model(&model.AppInstall{})
	err := db.Preload("App").First(&install).Error
	return install, err
}

func (a AppInstallRepo) Create(ctx context.Context, install *model.AppInstall) error {
	db := getTx(ctx).Model(&model.AppInstall{})
	return db.Create(&install).Error
}

func (a AppInstallRepo) Save(install model.AppInstall) error {
	return getDb().Save(&install).Error
}

func (a AppInstallRepo) DeleteBy(opts ...DBOption) error {
	return getDb(opts...).Delete(&model.AppInstall{}).Error
}

func (a AppInstallRepo) Delete(ctx context.Context, install model.AppInstall) error {
	db := getTx(ctx).Model(&model.AppInstall{})
	return db.Delete(&install).Error
}

func (a AppInstallRepo) Page(page, size int, opts ...DBOption) (int64, []model.AppInstall, error) {
	var apps []model.AppInstall
	db := getDb(opts...).Model(&model.AppInstall{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Debug().Limit(size).Offset(size * (page - 1)).Preload("App").Find(&apps).Error
	return count, apps, err
}

func (a AppInstallRepo) BatchUpdateBy(maps map[string]interface{}, opts ...DBOption) error {

	db := getDb(opts...).Model(&model.AppInstall{})
	if len(opts) == 0 {
		db = db.Where("1=1")
	}
	return db.Updates(&maps).Error
}
