package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"gorm.io/gorm"
)

func NewISSLRepo() ISSLRepo {
	return &WebsiteSSLRepo{}
}

type ISSLRepo interface {
	WithByAlias(alias string) DBOption
	WithByAcmeAccountId(acmeAccountId uint) DBOption
	WithByDnsAccountId(dnsAccountId uint) DBOption
	WithByCAID(caID uint) DBOption
	Page(page, size int, opts ...DBOption) (int64, []model.WebsiteSSL, error)
	GetFirst(opts ...DBOption) (*model.WebsiteSSL, error)
	List(opts ...DBOption) ([]model.WebsiteSSL, error)
	Create(ctx context.Context, ssl *model.WebsiteSSL) error
	Save(ssl *model.WebsiteSSL) error
	DeleteBy(opts ...DBOption) error
	SaveByMap(ssl *model.WebsiteSSL, params map[string]interface{}) error
}

type WebsiteSSLRepo struct {
}

func (w WebsiteSSLRepo) WithByAlias(alias string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("alias = ?", alias)
	}
}

func (w WebsiteSSLRepo) WithByAcmeAccountId(acmeAccountId uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("acme_account_id = ?", acmeAccountId)
	}
}

func (w WebsiteSSLRepo) WithByDnsAccountId(dnsAccountId uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("dns_account_id = ?", dnsAccountId)
	}
}

func (w WebsiteSSLRepo) WithByCAID(caID uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("ca_id = ?", caID)
	}
}
func (w WebsiteSSLRepo) Page(page, size int, opts ...DBOption) (int64, []model.WebsiteSSL, error) {
	var sslList []model.WebsiteSSL
	db := getDb(opts...).Model(&model.WebsiteSSL{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Preload("AcmeAccount").Preload("DnsAccount").Preload("Websites").Find(&sslList).Error
	return count, sslList, err
}

func (w WebsiteSSLRepo) GetFirst(opts ...DBOption) (*model.WebsiteSSL, error) {
	var website *model.WebsiteSSL
	db := getDb(opts...).Model(&model.WebsiteSSL{})
	if err := db.Preload("AcmeAccount").Preload("DnsAccount").First(&website).Error; err != nil {
		return website, err
	}
	return website, nil
}

func (w WebsiteSSLRepo) List(opts ...DBOption) ([]model.WebsiteSSL, error) {
	var websites []model.WebsiteSSL
	db := getDb(opts...).Model(&model.WebsiteSSL{})
	if err := db.Preload("AcmeAccount").Preload("DnsAccount").Find(&websites).Error; err != nil {
		return websites, err
	}
	return websites, nil
}

func (w WebsiteSSLRepo) Create(ctx context.Context, ssl *model.WebsiteSSL) error {
	return getTx(ctx).Create(ssl).Error
}

func (w WebsiteSSLRepo) Save(ssl *model.WebsiteSSL) error {
	return getDb().Model(&model.WebsiteSSL{BaseModel: model.BaseModel{
		ID: ssl.ID,
	}}).Save(&ssl).Error
}

func (w WebsiteSSLRepo) SaveByMap(ssl *model.WebsiteSSL, params map[string]interface{}) error {
	return getDb().Model(&model.WebsiteSSL{BaseModel: model.BaseModel{
		ID: ssl.ID,
	}}).Updates(params).Error
}

func (w WebsiteSSLRepo) DeleteBy(opts ...DBOption) error {
	return getDb(opts...).Delete(&model.WebsiteSSL{}).Error
}
