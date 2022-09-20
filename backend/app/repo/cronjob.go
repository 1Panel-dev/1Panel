package repo

import (
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/global"
)

type CronjobRepo struct{}

type ICronjobRepo interface {
	Get(opts ...DBOption) (model.Cronjob, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.Cronjob, error)
	Create(cronjob *model.Cronjob) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error
}

func NewICronjobService() ICronjobRepo {
	return &CronjobRepo{}
}

func (u *CronjobRepo) Get(opts ...DBOption) (model.Cronjob, error) {
	var cronjob model.Cronjob
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&cronjob).Error
	return cronjob, err
}

func (u *CronjobRepo) Page(page, size int, opts ...DBOption) (int64, []model.Cronjob, error) {
	var users []model.Cronjob
	db := global.DB.Model(&model.Cronjob{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&users).Error
	return count, users, err
}

func (u *CronjobRepo) Create(cronjob *model.Cronjob) error {
	return global.DB.Create(cronjob).Error
}

func (u *CronjobRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.Cronjob{}).Where("id = ?", id).Updates(vars).Error
}

func (u *CronjobRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.Cronjob{}).Error
}
