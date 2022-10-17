package repo

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
)

type SettingRepo struct{}

type ISettingRepo interface {
	GetList(opts ...DBOption) ([]model.Setting, error)
	Get(opts ...DBOption) (model.Setting, error)
	Update(key, value string) error
	WithByKey(key string) DBOption
}

func NewISettingRepo() ISettingRepo {
	return &SettingRepo{}
}

func (u *SettingRepo) GetList(opts ...DBOption) ([]model.Setting, error) {
	var settings []model.Setting
	db := global.DB.Model(&model.Setting{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&settings).Error
	return settings, err
}

func (u *SettingRepo) Get(opts ...DBOption) (model.Setting, error) {
	var settings model.Setting
	db := global.DB.Model(&model.Setting{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&settings).Error
	return settings, err
}

func (c *SettingRepo) WithByKey(key string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("key = ?", key)
	}
}

func (u *SettingRepo) Update(key, value string) error {
	return global.DB.Model(&model.Setting{}).Where("key = ?", key).Updates(map[string]interface{}{"value": value}).Error
}
