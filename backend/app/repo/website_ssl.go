package repo

import "github.com/1Panel-dev/1Panel/backend/app/model"

type WebsiteSSLRepo struct {
}

func (w WebsiteSSLRepo) Page(page, size int, opts ...DBOption) (int64, []model.WebSiteSSL, error) {
	var sslList []model.WebSiteSSL
	db := getDb(opts...).Model(&model.WebSiteSSL{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Debug().Limit(size).Offset(size * (page - 1)).Find(&sslList).Error
	return count, sslList, err
}

func (w WebsiteSSLRepo) Create(ssl model.WebSiteSSL) error {
	return getDb().Create(&ssl).Error
}

func (w WebsiteSSLRepo) Save(ssl model.WebSiteSSL) error {
	return getDb().Save(&ssl).Error
}

func (w WebsiteSSLRepo) DeleteBy(opts ...DBOption) error {
	return getDb(opts...).Delete(&model.WebSiteSSL{}).Error
}
