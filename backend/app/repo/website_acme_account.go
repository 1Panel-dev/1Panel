package repo

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"gorm.io/gorm"
)

type IAcmeAccountRepo interface {
	Page(page, size int, opts ...DBOption) (int64, []model.WebsiteAcmeAccount, error)
	GetFirst(opts ...DBOption) (*model.WebsiteAcmeAccount, error)
	Create(account model.WebsiteAcmeAccount) error
	Save(account model.WebsiteAcmeAccount) error
	DeleteBy(opts ...DBOption) error
	WithEmail(email string) DBOption
	WithType(acType string) DBOption
}

func NewIAcmeAccountRepo() IAcmeAccountRepo {
	return &WebsiteAcmeAccountRepo{}
}

type WebsiteAcmeAccountRepo struct {
}

func (w *WebsiteAcmeAccountRepo) WithEmail(email string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("email = ?", email)
	}
}
func (w *WebsiteAcmeAccountRepo) WithType(acType string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("type = ?", acType)
	}
}

func (w *WebsiteAcmeAccountRepo) Page(page, size int, opts ...DBOption) (int64, []model.WebsiteAcmeAccount, error) {
	var accounts []model.WebsiteAcmeAccount
	db := getDb(opts...).Model(&model.WebsiteAcmeAccount{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&accounts).Error
	return count, accounts, err
}

func (w *WebsiteAcmeAccountRepo) GetFirst(opts ...DBOption) (*model.WebsiteAcmeAccount, error) {
	var account model.WebsiteAcmeAccount
	db := getDb(opts...).Model(&model.WebsiteAcmeAccount{})
	if err := db.First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (w *WebsiteAcmeAccountRepo) Create(account model.WebsiteAcmeAccount) error {
	return getDb().Create(&account).Error
}

func (w *WebsiteAcmeAccountRepo) Save(account model.WebsiteAcmeAccount) error {
	return getDb().Save(&account).Error
}

func (w *WebsiteAcmeAccountRepo) DeleteBy(opts ...DBOption) error {
	return getDb(opts...).Debug().Delete(&model.WebsiteAcmeAccount{}).Error
}
