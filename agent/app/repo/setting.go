package repo

import (
	"errors"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

type SettingRepo struct{}

type ISettingRepo interface {
	GetList(opts ...DBOption) ([]model.Setting, error)
	Get(opts ...DBOption) (model.Setting, error)
	GetValueByKey(key string) (string, error)
	Create(key, value string) error
	Update(key, value string) error
	WithByKey(key string) DBOption

	UpdateOrCreate(key, value string) error

	GetDescription(opts ...DBOption) (model.CommonDescription, error)
	GetDescriptionList(opts ...DBOption) ([]model.CommonDescription, error)
	CreateDescription(data *model.CommonDescription) error
	UpdateDescription(id string, val map[string]interface{}) error
	DelDescription(id string) error
	WithByDescriptionID(id string) DBOption
}

func NewISettingRepo() ISettingRepo {
	return &SettingRepo{}
}

func (s *SettingRepo) GetList(opts ...DBOption) ([]model.Setting, error) {
	var settings []model.Setting
	db := global.DB.Model(&model.Setting{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&settings).Error
	return settings, err
}

func (s *SettingRepo) Create(key, value string) error {
	setting := &model.Setting{
		Key:   key,
		Value: value,
	}
	return global.DB.Create(setting).Error
}

func (s *SettingRepo) Get(opts ...DBOption) (model.Setting, error) {
	var settings model.Setting
	db := global.DB.Model(&model.Setting{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&settings).Error
	return settings, err
}

func (s *SettingRepo) GetValueByKey(key string) (string, error) {
	var setting model.Setting
	if err := global.DB.Model(&model.Setting{}).Where("key = ?", key).First(&setting).Error; err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (s *SettingRepo) WithByKey(key string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("key = ?", key)
	}
}

func (s *SettingRepo) Update(key, value string) error {
	return global.DB.Model(&model.Setting{}).Where("key = ?", key).Updates(map[string]interface{}{"value": value}).Error
}

func (s *SettingRepo) UpdateOrCreate(key, value string) error {
	var setting model.Setting
	result := global.DB.Where("key = ?", key).First(&setting)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return global.DB.Create(&model.Setting{Key: key, Value: value}).Error
		}
		return result.Error
	}
	return global.DB.Model(&setting).UpdateColumn("value", value).Error
}

func (s *SettingRepo) GetDescriptionList(opts ...DBOption) ([]model.CommonDescription, error) {
	var lists []model.CommonDescription
	db := global.DB.Model(&model.CommonDescription{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&lists).Error
	return lists, err
}
func (s *SettingRepo) GetDescription(opts ...DBOption) (model.CommonDescription, error) {
	var data model.CommonDescription
	db := global.DB.Model(&model.CommonDescription{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&data).Error
	return data, err
}
func (s *SettingRepo) CreateDescription(data *model.CommonDescription) error {
	return global.DB.Create(data).Error
}
func (s *SettingRepo) UpdateDescription(id string, val map[string]interface{}) error {
	return global.DB.Model(&model.CommonDescription{}).Where("id = ?", id).Updates(val).Error
}
func (s *SettingRepo) DelDescription(id string) error {
	return global.DB.Where("id = ?", id).Delete(&model.CommonDescription{}).Error
}
func (s *SettingRepo) WithByDescriptionID(id string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id = ?", id)
	}
}
