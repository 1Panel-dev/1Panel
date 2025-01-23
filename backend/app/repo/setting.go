package repo

import (
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
)

type SettingRepo struct{}

type ISettingRepo interface {
	GetList(opts ...DBOption) ([]model.Setting, error)
	Get(opts ...DBOption) (model.Setting, error)
	Create(key, value string) error
	Update(key, value string) error
	WithByKey(key string) DBOption
	UpdateOrCreate(key, value string) error

	CreateMonitorBase(model model.MonitorBase) error
	BatchCreateMonitorIO(ioList []model.MonitorIO) error
	BatchCreateMonitorNet(ioList []model.MonitorNetwork) error
	DelMonitorBase(timeForDelete time.Time) error
	DelMonitorIO(timeForDelete time.Time) error
	DelMonitorNet(timeForDelete time.Time) error
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

func (u *SettingRepo) Create(key, value string) error {
	setting := &model.Setting{
		Key:   key,
		Value: value,
	}
	return global.DB.Create(setting).Error
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

func (u *SettingRepo) CreateMonitorBase(model model.MonitorBase) error {
	return global.MonitorDB.Create(&model).Error
}
func (u *SettingRepo) BatchCreateMonitorIO(ioList []model.MonitorIO) error {
	return global.MonitorDB.CreateInBatches(ioList, len(ioList)).Error
}
func (u *SettingRepo) BatchCreateMonitorNet(ioList []model.MonitorNetwork) error {
	return global.MonitorDB.CreateInBatches(ioList, len(ioList)).Error
}
func (u *SettingRepo) DelMonitorBase(timeForDelete time.Time) error {
	return global.MonitorDB.Where("created_at < ?", timeForDelete).Delete(&model.MonitorBase{}).Error
}
func (u *SettingRepo) DelMonitorIO(timeForDelete time.Time) error {
	return global.MonitorDB.Where("created_at < ?", timeForDelete).Delete(&model.MonitorIO{}).Error
}
func (u *SettingRepo) DelMonitorNet(timeForDelete time.Time) error {
	return global.MonitorDB.Where("created_at < ?", timeForDelete).Delete(&model.MonitorNetwork{}).Error
}

func (u *SettingRepo) UpdateOrCreate(key, value string) error {
	return global.DB.Model(&model.Setting{}).Where("key = ?", key).Assign(model.Setting{Key: key, Value: value}).FirstOrCreate(&model.Setting{}).Error
}
