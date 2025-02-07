package repo

import (
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
)

type SettingRepo struct{}

type ISettingRepo interface {
	List(opts ...global.DBOption) ([]model.Setting, error)
	Get(opts ...global.DBOption) (model.Setting, error)
	GetValueByKey(key string) (string, error)
	Create(key, value string) error
	Update(key, value string) error
	UpdateOrCreate(key, value string) error
}

func NewISettingRepo() ISettingRepo {
	return &SettingRepo{}
}

func (u *SettingRepo) List(opts ...global.DBOption) ([]model.Setting, error) {
	var settings []model.Setting
	db := global.DB.Model(&model.Setting{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&settings).Error
	return settings, err
}

func (u *SettingRepo) Create(key, value string) error {
	setting := &model.Setting{
		Key:   key,
		Value: value,
	}
	return global.DB.Create(setting).Error
}

func (u *SettingRepo) Get(opts ...global.DBOption) (model.Setting, error) {
	var settings model.Setting
	db := global.DB.Model(&model.Setting{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&settings).Error
	return settings, err
}

func (u *SettingRepo) GetValueByKey(key string) (string, error) {
	var setting model.Setting
	if err := global.DB.Model(&model.Setting{}).Where("key = ?", key).First(&setting).Error; err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (u *SettingRepo) Update(key, value string) error {
	return global.DB.Model(&model.Setting{}).Where("key = ?", key).Updates(map[string]interface{}{"value": value}).Error
}

func (u *SettingRepo) UpdateOrCreate(key, value string) error {
	return global.DB.Model(&model.Setting{}).Where("key = ?", key).Assign(model.Setting{Key: key, Value: value}).FirstOrCreate(&model.Setting{}).Error
}
