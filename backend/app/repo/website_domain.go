package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WebsiteDomainRepo struct {
}

type IWebsiteDomainRepo interface {
	WithWebsiteId(websiteId uint) DBOption
	WithPort(port int) DBOption
	WithDomain(domain string) DBOption
	WithDomainLike(domain string) DBOption
	Page(page, size int, opts ...DBOption) (int64, []model.WebsiteDomain, error)
	GetFirst(opts ...DBOption) (model.WebsiteDomain, error)
	GetBy(opts ...DBOption) ([]model.WebsiteDomain, error)
	BatchCreate(ctx context.Context, domains []model.WebsiteDomain) error
	Create(ctx context.Context, app *model.WebsiteDomain) error
	Save(ctx context.Context, app *model.WebsiteDomain) error
	DeleteBy(ctx context.Context, opts ...DBOption) error
	DeleteAll(ctx context.Context) error
}

func NewIWebsiteDomainRepo() IWebsiteDomainRepo {
	return &WebsiteDomainRepo{}
}

func (w WebsiteDomainRepo) WithWebsiteId(websiteId uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("website_id = ?", websiteId)
	}
}

func (w WebsiteDomainRepo) WithPort(port int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("port = ?", port)
	}
}
func (w WebsiteDomainRepo) WithDomain(domain string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("domain = ?", domain)
	}
}
func (w WebsiteDomainRepo) WithDomainLike(domain string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("domain like ?", "%"+domain+"%")
	}
}
func (w WebsiteDomainRepo) Page(page, size int, opts ...DBOption) (int64, []model.WebsiteDomain, error) {
	var domains []model.WebsiteDomain
	db := getDb(opts...).Model(&model.WebsiteDomain{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&domains).Error
	return count, domains, err
}

func (w WebsiteDomainRepo) GetFirst(opts ...DBOption) (model.WebsiteDomain, error) {
	var domain model.WebsiteDomain
	db := getDb(opts...).Model(&model.WebsiteDomain{})
	if err := db.First(&domain).Error; err != nil {
		return domain, err
	}
	return domain, nil
}

func (w WebsiteDomainRepo) GetBy(opts ...DBOption) ([]model.WebsiteDomain, error) {
	var domains []model.WebsiteDomain
	db := getDb(opts...).Model(&model.WebsiteDomain{})
	if err := db.Find(&domains).Error; err != nil {
		return domains, err
	}
	return domains, nil
}

func (w WebsiteDomainRepo) BatchCreate(ctx context.Context, domains []model.WebsiteDomain) error {
	return getTx(ctx).Model(&model.WebsiteDomain{}).Create(&domains).Error
}

func (w WebsiteDomainRepo) Create(ctx context.Context, app *model.WebsiteDomain) error {
	return getTx(ctx).Omit(clause.Associations).Create(app).Error
}

func (w WebsiteDomainRepo) Save(ctx context.Context, app *model.WebsiteDomain) error {
	return getTx(ctx).Omit(clause.Associations).Save(app).Error
}

func (w WebsiteDomainRepo) DeleteBy(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.WebsiteDomain{}).Error
}

func (w WebsiteDomainRepo) DeleteAll(ctx context.Context) error {
	return getTx(ctx).Where("1 = 1 ").Delete(&model.WebsiteDomain{}).Error
}
