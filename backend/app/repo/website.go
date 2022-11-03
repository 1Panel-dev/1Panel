package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WebSiteRepo struct {
}

func (w WebSiteRepo) WithAppInstallId(appInstallId uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("app_install_id = ?", appInstallId)
	}
}

func (w WebSiteRepo) Page(page, size int, opts ...DBOption) (int64, []model.WebSite, error) {
	var websites []model.WebSite
	db := getDb(opts...).Model(&model.WebSite{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Debug().Limit(size).Offset(size * (page - 1)).Find(&websites).Error
	return count, websites, err
}

func (w WebSiteRepo) GetFirst(opts ...DBOption) (model.WebSite, error) {
	var website model.WebSite
	db := getDb(opts...).Model(&model.WebSite{})
	if err := db.First(&website).Error; err != nil {
		return website, err
	}
	return website, nil
}

func (w WebSiteRepo) GetBy(opts ...DBOption) ([]model.WebSite, error) {
	var websites []model.WebSite
	db := getDb(opts...).Model(&model.WebSite{})
	if err := db.Find(&websites).Error; err != nil {
		return websites, err
	}
	return websites, nil
}

func (w WebSiteRepo) Create(ctx context.Context, app *model.WebSite) error {
	return getTx(ctx).Omit(clause.Associations).Create(app).Error
}

func (w WebSiteRepo) Save(ctx context.Context, app *model.WebSite) error {
	return getTx(ctx).Omit(clause.Associations).Save(app).Error
}

func (w WebSiteRepo) DeleteBy(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.WebSite{}).Error
}
