package repo

import (
	"errors"
	"time"

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

	CreateMonitorBase(model model.MonitorBase) error
	BatchCreateMonitorIO(ioList []model.MonitorIO) error
	BatchCreateMonitorNet(ioList []model.MonitorNetwork) error
	DelMonitorBase(timeForDelete time.Time) error
	DelMonitorIO(timeForDelete time.Time) error
	DelMonitorNet(timeForDelete time.Time) error
	UpdateOrCreate(key, value string) error
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

func (s *SettingRepo) CreateMonitorBase(model model.MonitorBase) error {
	return global.MonitorDB.Create(&model).Error
}
func (s *SettingRepo) BatchCreateMonitorIO(ioList []model.MonitorIO) error {
	return global.MonitorDB.CreateInBatches(ioList, len(ioList)).Error
}
func (s *SettingRepo) BatchCreateMonitorNet(ioList []model.MonitorNetwork) error {
	return global.MonitorDB.CreateInBatches(ioList, len(ioList)).Error
}
func (s *SettingRepo) DelMonitorBase(timeForDelete time.Time) error {
	return global.MonitorDB.Where("created_at < ?", timeForDelete).Delete(&model.MonitorBase{}).Error
}
func (s *SettingRepo) DelMonitorIO(timeForDelete time.Time) error {
	return global.MonitorDB.Where("created_at < ?", timeForDelete).Delete(&model.MonitorIO{}).Error
}
func (s *SettingRepo) DelMonitorNet(timeForDelete time.Time) error {
	return global.MonitorDB.Where("created_at < ?", timeForDelete).Delete(&model.MonitorNetwork{}).Error
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
