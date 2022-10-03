package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/global"
	"gorm.io/gorm"
)

type AppDetailRepo struct {
}

func (a AppDetailRepo) WithVersion(version string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("version = ?", version)
	}
}
func (a AppDetailRepo) WithAppId(id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("app_id = ?", id)
	}
}

func (a AppDetailRepo) GetFirst(opts ...DBOption) (model.AppDetail, error) {
	var detail model.AppDetail
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&detail).Error
	return detail, err
}

func (a AppDetailRepo) Update(ctx context.Context, detail model.AppDetail) error {
	db := ctx.Value("db").(*gorm.DB)
	return db.Save(&detail).Error
}

func (a AppDetailRepo) BatchCreate(ctx context.Context, details []model.AppDetail) error {
	db := ctx.Value("db").(*gorm.DB)
	return db.Model(&model.AppDetail{}).Create(&details).Error
}

func (a AppDetailRepo) DeleteByAppIds(ctx context.Context, appIds []uint) error {
	db := ctx.Value("db").(*gorm.DB)
	return db.Where("app_id in (?)", appIds).Delete(&model.AppDetail{}).Error
}

func (a AppDetailRepo) GetBy(opts ...DBOption) ([]model.AppDetail, error) {
	var details []model.AppDetail
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&details).Error
	return details, err
}

func (a AppDetailRepo) BatchUpdateBy(update model.AppDetail, opts ...DBOption) error {
	db := global.DB.Model(model.AppDetail{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Updates(update).Error
}
