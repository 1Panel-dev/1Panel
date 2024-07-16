package repo

import (
	"context"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IWebsiteRepo interface {
	WithAppInstallId(appInstallId uint) DBOption
	WithDomain(domain string) DBOption
	WithAlias(alias string) DBOption
	WithWebsiteSSLID(sslId uint) DBOption
	WithGroupID(groupId uint) DBOption
	WithDefaultServer() DBOption
	WithDomainLike(domain string) DBOption
	WithRuntimeID(runtimeID uint) DBOption
	WithIDs(ids []uint) DBOption
	Page(page, size int, opts ...DBOption) (int64, []model.Website, error)
	List(opts ...DBOption) ([]model.Website, error)
	GetFirst(opts ...DBOption) (model.Website, error)
	GetBy(opts ...DBOption) ([]model.Website, error)
	Save(ctx context.Context, app *model.Website) error
	SaveWithoutCtx(app *model.Website) error
	DeleteBy(ctx context.Context, opts ...DBOption) error
	Create(ctx context.Context, app *model.Website) error
	DeleteAll(ctx context.Context) error
}

func NewIWebsiteRepo() IWebsiteRepo {
	return &WebsiteRepo{}
}

type WebsiteRepo struct {
}

func (w *WebsiteRepo) WithAppInstallId(appInstallID uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("app_install_id = ?", appInstallID)
	}
}

func (w *WebsiteRepo) WithIDs(ids []uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id in (?)", ids)
	}
}

func (w *WebsiteRepo) WithRuntimeID(runtimeID uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("runtime_id = ?", runtimeID)
	}
}

func (w *WebsiteRepo) WithDomain(domain string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("primary_domain = ?", domain)
	}
}

func (w *WebsiteRepo) WithDomainLike(domain string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("primary_domain like ?", "%"+domain+"%")
	}
}

func (w *WebsiteRepo) WithAlias(alias string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("alias = ?", alias)
	}
}

func (w *WebsiteRepo) WithWebsiteSSLID(sslId uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("website_ssl_id = ?", sslId)
	}
}

func (w *WebsiteRepo) WithGroupID(groupId uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("website_group_id = ?", groupId)
	}
}

func (w *WebsiteRepo) WithDefaultServer() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("default_server = 1")
	}
}

func (w *WebsiteRepo) Page(page, size int, opts ...DBOption) (int64, []model.Website, error) {
	var websites []model.Website
	db := getDb(opts...).Model(&model.Website{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Debug().Limit(size).Offset(size * (page - 1)).Preload("WebsiteSSL").Find(&websites).Error
	return count, websites, err
}

func (w *WebsiteRepo) List(opts ...DBOption) ([]model.Website, error) {
	var websites []model.Website
	err := getDb(opts...).Model(&model.Website{}).Preload("Domains").Preload("WebsiteSSL").Find(&websites).Error
	return websites, err
}

func (w *WebsiteRepo) GetFirst(opts ...DBOption) (model.Website, error) {
	var website model.Website
	db := getDb(opts...).Model(&model.Website{})
	if err := db.Preload("Domains").First(&website).Error; err != nil {
		return website, err
	}
	return website, nil
}

func (w *WebsiteRepo) GetBy(opts ...DBOption) ([]model.Website, error) {
	var websites []model.Website
	db := getDb(opts...).Model(&model.Website{})
	if err := db.Find(&websites).Error; err != nil {
		return websites, err
	}
	return websites, nil
}

func (w *WebsiteRepo) Create(ctx context.Context, app *model.Website) error {
	return getTx(ctx).Omit(clause.Associations).Create(app).Error
}

func (w *WebsiteRepo) Save(ctx context.Context, app *model.Website) error {
	return getTx(ctx).Omit(clause.Associations).Save(app).Error
}

func (w *WebsiteRepo) SaveWithoutCtx(website *model.Website) error {
	return global.DB.Save(website).Error
}

func (w *WebsiteRepo) DeleteBy(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.Website{}).Error
}

func (w *WebsiteRepo) DeleteAll(ctx context.Context) error {
	return getTx(ctx).Where("1 = 1 ").Delete(&model.Website{}).Error
}
