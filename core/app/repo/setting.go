package repo

import (
	"errors"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/init/migration/helper"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

type SettingRepo struct{}

var (
	settingCache = cache.New(5*time.Minute, 10*time.Minute)
	settingTTL   = 5 * time.Minute
)

type ISettingRepo interface {
	List(opts ...global.DBOption) ([]model.Setting, error)
	Get(opts ...global.DBOption) (model.Setting, error)
	GetValueByKey(key string) (string, error)
	Create(key, value string) error
	Update(key, value string) error
	UpdateOrCreate(key, value string) error
	DefaultMenu() error
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
	if err := global.DB.Create(setting).Error; err != nil {
		return err
	}
	settingCache.Set(key, value, settingTTL)
	return nil
}

func (u *SettingRepo) Get(opts ...global.DBOption) (model.Setting, error) {
	var settings model.Setting
	db := global.DB.Model(&model.Setting{})
	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&settings).Error
	if err == nil && settings.Key != "" {
		settingCache.Set(settings.Key, settings.Value, settingTTL)
	}
	return settings, err
}

func (u *SettingRepo) GetValueByKey(key string) (string, error) {
	if val, found := settingCache.Get(key); found {
		return val.(string), nil
	}

	var setting model.Setting
	if err := global.DB.Model(&model.Setting{}).Where("key = ?", key).First(&setting).Error; err != nil {
		return "", err
	}
	settingCache.Set(key, setting.Value, settingTTL)
	return setting.Value, nil
}

func (u *SettingRepo) Update(key, value string) error {
	if err := global.DB.Model(&model.Setting{}).Where("key = ?", key).Updates(map[string]interface{}{"value": value}).Error; err != nil {
		return err
	}
	settingCache.Set(key, value, settingTTL)
	return nil
}

func (u *SettingRepo) UpdateOrCreate(key, value string) error {
	var setting model.Setting
	result := global.DB.Where("key = ?", key).First(&setting)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if err := global.DB.Create(&model.Setting{Key: key, Value: value}).Error; err != nil {
				return err
			}
			settingCache.Set(key, value, settingTTL)
			return nil
		}
		return result.Error
	}
	if err := global.DB.Model(&setting).UpdateColumn("value", value).Error; err != nil {
		return err
	}
	settingCache.Set(key, value, settingTTL)
	return nil
}

func (u *SettingRepo) DefaultMenu() error {
	menus := helper.LoadMenus()
	if err := global.DB.Model(&model.Setting{}).
		Where("key = ?", "HideMenu").
		Update("value", menus).Error; err != nil {
		return err
	}
	settingCache.Set("HideMenu", menus, settingTTL)
	return nil
}
