package repo

import "github.com/1Panel-dev/1Panel/backend/app/model"

type WebsiteAcmeAccountRepo struct {
}

func (w WebsiteAcmeAccountRepo) Page(page, size int, opts ...DBOption) (int64, []model.WebsiteAcmeAccount, error) {
	var accounts []model.WebsiteAcmeAccount
	db := getDb(opts...).Model(&model.WebsiteAcmeAccount{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Debug().Limit(size).Offset(size * (page - 1)).Find(&accounts).Error
	return count, accounts, err
}

func (w WebsiteAcmeAccountRepo) Create(account model.WebsiteAcmeAccount) error {
	return getDb().Create(&account).Error
}

func (w WebsiteAcmeAccountRepo) Save(account model.WebsiteAcmeAccount) error {
	return getDb().Save(&account).Error
}

func (w WebsiteAcmeAccountRepo) DeleteBy(opts ...DBOption) error {
	return getDb(opts...).Debug().Delete(&model.WebsiteAcmeAccount{}).Error
}
