package repo

import (
	"context"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
)

type AppInstallResourceRpo struct {
}

func (a AppInstallResourceRpo) WithAppInstallId(appInstallId uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("app_install_id = ?", appInstallId)
	}
}

func (a AppInstallResourceRpo) WithLinkId(linkId uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("link_id = ?", linkId)
	}
}

func (a AppInstallResourceRpo) WithResourceId(resourceId uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("resource_id = ?", resourceId)
	}
}

func (a AppInstallResourceRpo) GetBy(opts ...DBOption) ([]model.AppInstallResource, error) {
	db := global.DB.Model(&model.AppInstallResource{})
	var resources []model.AppInstallResource
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&resources).Error
	return resources, err
}

func (a AppInstallResourceRpo) GetFirst(opts ...DBOption) (model.AppInstallResource, error) {
	db := global.DB.Model(&model.AppInstallResource{})
	var resources model.AppInstallResource
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&resources).Error
	return resources, err
}

func (a AppInstallResourceRpo) Create(ctx context.Context, resource *model.AppInstallResource) error {
	db := getTx(ctx).Model(&model.AppInstallResource{})
	return db.Create(&resource).Error
}

func (a AppInstallResourceRpo) DeleteBy(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.AppInstallResource{}).Error
}
