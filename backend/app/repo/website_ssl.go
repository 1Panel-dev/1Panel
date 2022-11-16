package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"gorm.io/gorm"
)

type WebsiteSSLRepo struct {
}

func (w WebsiteSSLRepo) ByAlias(alias string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("alias = ?", alias)
	}
}

func (w WebsiteSSLRepo) Page(page, size int, opts ...DBOption) (int64, []model.WebSiteSSL, error) {
	var sslList []model.WebSiteSSL
	db := getDb(opts...).Model(&model.WebSiteSSL{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Debug().Limit(size).Offset(size * (page - 1)).Find(&sslList).Error
	return count, sslList, err
}

func (w WebsiteSSLRepo) GetFirst(opts ...DBOption) (model.WebSiteSSL, error) {
	var website model.WebSiteSSL
	db := getDb(opts...).Model(&model.WebSiteSSL{})
	if err := db.First(&website).Error; err != nil {
		return website, err
	}
	return website, nil
}

func (w WebsiteSSLRepo) Create(ctx context.Context, ssl *model.WebSiteSSL) error {
	return getTx(ctx).Create(ssl).Error
}

func (w WebsiteSSLRepo) Save(ssl model.WebSiteSSL) error {
	return getDb().Save(&ssl).Error
}

func (w WebsiteSSLRepo) DeleteBy(opts ...DBOption) error {
	return getDb(opts...).Delete(&model.WebSiteSSL{}).Error
}
