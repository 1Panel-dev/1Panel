package repo

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
)

type WebsiteDnsAccountRepo struct {
}

func (w WebsiteDnsAccountRepo) Page(page, size int, opts ...DBOption) (int64, []model.WebsiteDnsAccount, error) {
	var accounts []model.WebsiteDnsAccount
	db := getDb(opts...).Model(&model.WebsiteDnsAccount{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Debug().Limit(size).Offset(size * (page - 1)).Find(&accounts).Error
	return count, accounts, err
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
