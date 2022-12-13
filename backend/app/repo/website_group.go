package repo

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm/clause"
)

type WebsiteGroupRepo struct {
}

func (w WebsiteGroupRepo) Page(page, size int, opts ...DBOption) (int64, []model.WebsiteGroup, error) {
	var groups []model.WebsiteGroup
	db := getDb(opts...).Model(&model.WebsiteGroup{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Order("`default` desc").Find(&groups).Error
	return count, groups, err
}

func (w WebsiteGroupRepo) GetBy(opts ...DBOption) ([]model.WebsiteGroup, error) {
	var groups []model.WebsiteGroup
	db := getDb(opts...).Model(&model.WebsiteGroup{})
	if err := db.Order("`default` desc").Find(&groups).Error; err != nil {
		return groups, err
	}
	return groups, nil
}

func (w WebsiteGroupRepo) Create(app *model.WebsiteGroup) error {
	return getDb().Omit(clause.Associations).Create(app).Error
}

func (w WebsiteGroupRepo) Save(app *model.WebsiteGroup) error {
	return getDb().Omit(clause.Associations).Save(app).Error
}

func (w WebsiteGroupRepo) DeleteBy(opts ...DBOption) error {
	return getDb(opts...).Delete(&model.WebsiteGroup{}).Error
}

func (w WebsiteGroupRepo) CancelDefault() error {
	return global.DB.Model(&model.WebsiteGroup{}).Where("`default` = 1").Updates(map[string]interface{}{"default": 0}).Error
}
