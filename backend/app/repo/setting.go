package repo

import (
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/global"
)

type SettingRepo struct{}

type ISettingRepo interface {
	Get(opts ...DBOption) ([]model.Setting, error)
	Update(key, value string) error
}

func NewISettingService() ISettingRepo {
	return &SettingRepo{}
}

func (u *SettingRepo) Get(opts ...DBOption) ([]model.Setting, error) {
	var settings []model.Setting
	db := global.DB.Model(&model.Setting{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&settings).Error
	return settings, err
}

func (u *SettingRepo) Update(key, value string) error {
	return global.DB.Model(&model.Setting{}).Where("key = ?", key).Updates(map[string]interface{}{"value": value}).Error
}
