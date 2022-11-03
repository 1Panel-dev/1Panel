package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WebSiteDomainRepo struct {
}

func (w WebSiteDomainRepo) WithWebSiteId(websiteId uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("web_site_id = ?", websiteId)
	}
}

func (w WebSiteDomainRepo) WithPort(port int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("port = ?", port)
	}
}
func (w WebSiteDomainRepo) WithDomain(domain string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("domain = ?", domain)
	}
}
func (w WebSiteDomainRepo) Page(page, size int, opts ...DBOption) (int64, []model.WebSiteDomain, error) {
	var domains []model.WebSiteDomain
	db := getDb(opts...).Model(&model.WebSiteDomain{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Debug().Limit(size).Offset(size * (page - 1)).Find(&domains).Error
	return count, domains, err
}

func (w WebSiteDomainRepo) GetFirst(opts ...DBOption) (model.WebSiteDomain, error) {
	var domain model.WebSiteDomain
	db := getDb(opts...).Model(&model.WebSiteDomain{})
	if err := db.First(&domain).Error; err != nil {
		return domain, err
	}
	return domain, nil
}

func (w WebSiteDomainRepo) GetBy(opts ...DBOption) ([]model.WebSiteDomain, error) {
	var domains []model.WebSiteDomain
	db := getDb(opts...).Model(&model.WebSiteDomain{})
	if err := db.Find(&domains).Error; err != nil {
		return domains, err
	}
	return domains, nil
}

func (w WebSiteDomainRepo) BatchCreate(ctx context.Context, domains []model.WebSiteDomain) error {
	return getTx(ctx).Model(&model.WebSiteDomain{}).Create(&domains).Error
}

func (w WebSiteDomainRepo) Create(ctx context.Context, app *model.WebSiteDomain) error {
	return getTx(ctx).Omit(clause.Associations).Create(app).Error
}

func (w WebSiteDomainRepo) Save(ctx context.Context, app *model.WebSiteDomain) error {
	return getTx(ctx).Omit(clause.Associations).Save(app).Error
}

func (w WebSiteDomainRepo) DeleteBy(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.WebSiteDomain{}).Error
}
