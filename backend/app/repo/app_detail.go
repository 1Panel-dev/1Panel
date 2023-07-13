package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AppDetailRepo struct {
}

type IAppDetailRepo interface {
	WithVersion(version string) DBOption
	WithAppId(id uint) DBOption
	WithIgnored() DBOption
	GetFirst(opts ...DBOption) (model.AppDetail, error)
	Update(ctx context.Context, detail model.AppDetail) error
	BatchCreate(ctx context.Context, details []model.AppDetail) error
	DeleteByAppIds(ctx context.Context, appIds []uint) error
	GetBy(opts ...DBOption) ([]model.AppDetail, error)
	BatchUpdateBy(maps map[string]interface{}, opts ...DBOption) error
	BatchDelete(ctx context.Context, appDetails []model.AppDetail) error
}

func NewIAppDetailRepo() IAppDetailRepo {
	return &AppDetailRepo{}
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

func (a AppDetailRepo) WithIgnored() DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("ignore_upgrade = 1")
	}
}

func (a AppDetailRepo) GetFirst(opts ...DBOption) (model.AppDetail, error) {
	var detail model.AppDetail
	err := getDb(opts...).Model(&model.AppDetail{}).Find(&detail).Error
	return detail, err
}

func (a AppDetailRepo) Update(ctx context.Context, detail model.AppDetail) error {
	return getTx(ctx).Save(&detail).Error
}

func (a AppDetailRepo) BatchCreate(ctx context.Context, details []model.AppDetail) error {
	return getTx(ctx).Model(&model.AppDetail{}).Create(&details).Error
}

func (a AppDetailRepo) DeleteByAppIds(ctx context.Context, appIds []uint) error {
	return getTx(ctx).Where("app_id in (?)", appIds).Delete(&model.AppDetail{}).Error
}

func (a AppDetailRepo) GetBy(opts ...DBOption) ([]model.AppDetail, error) {
	var details []model.AppDetail
	err := getDb(opts...).Find(&details).Error
	return details, err
}

func (a AppDetailRepo) BatchUpdateBy(maps map[string]interface{}, opts ...DBOption) error {
	db := getDb(opts...).Model(&model.AppDetail{})
	if len(opts) == 0 {
		db = db.Where("1=1")
	}
	return db.Updates(&maps).Error
}

func (a AppDetailRepo) BatchDelete(ctx context.Context, appDetails []model.AppDetail) error {
	return getTx(ctx).Omit(clause.Associations).Delete(&appDetails).Error
}
