package repo

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
)

type WebsiteDnsAccountRepo struct {
}

type IWebsiteDnsAccountRepo interface {
	Page(page, size int, opts ...DBOption) (int64, []model.WebsiteDnsAccount, error)
	GetFirst(opts ...DBOption) (*model.WebsiteDnsAccount, error)
	List(opts ...DBOption) ([]model.WebsiteDnsAccount, error)
	Create(account model.WebsiteDnsAccount) error
	Save(account model.WebsiteDnsAccount) error
	DeleteBy(opts ...DBOption) error
}

func NewIWebsiteDnsAccountRepo() IWebsiteDnsAccountRepo {
	return &WebsiteDnsAccountRepo{}
}

func (w WebsiteDnsAccountRepo) Page(page, size int, opts ...DBOption) (int64, []model.WebsiteDnsAccount, error) {
	var accounts []model.WebsiteDnsAccount
	db := getDb(opts...).Model(&model.WebsiteDnsAccount{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&accounts).Error
	return count, accounts, err
}

func (w WebsiteDnsAccountRepo) GetFirst(opts ...DBOption) (*model.WebsiteDnsAccount, error) {
	var account model.WebsiteDnsAccount
	db := getDb(opts...).Model(&model.WebsiteDnsAccount{})
	if err := db.First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (w WebsiteDnsAccountRepo) List(opts ...DBOption) ([]model.WebsiteDnsAccount, error) {
	var accounts []model.WebsiteDnsAccount
	db := getDb(opts...).Model(&model.WebsiteDnsAccount{})
	if err := db.Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (w WebsiteDnsAccountRepo) Create(account model.WebsiteDnsAccount) error {
	return getDb().Create(&account).Error
}

func (w WebsiteDnsAccountRepo) Save(account model.WebsiteDnsAccount) error {
	return getDb().Save(&account).Error
}

func (w WebsiteDnsAccountRepo) DeleteBy(opts ...DBOption) error {
	return getDb(opts...).Delete(&model.WebsiteDnsAccount{}).Error
}
