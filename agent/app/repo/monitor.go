package repo

import (
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
)

type MonitorRepo struct{}

type IMonitorRepo interface {
	GetBase(opts ...DBOption) ([]model.MonitorBase, error)
	GetIO(opts ...DBOption) ([]model.MonitorIO, error)
	GetNetwork(opts ...DBOption) ([]model.MonitorNetwork, error)

	CreateMonitorBase(model model.MonitorBase) error
	BatchCreateMonitorIO(ioList []model.MonitorIO) error
	BatchCreateMonitorNet(ioList []model.MonitorNetwork) error
	DelMonitorBase(timeForDelete time.Time) error
	DelMonitorIO(timeForDelete time.Time) error
	DelMonitorNet(timeForDelete time.Time) error
}

func NewIMonitorRepo() IMonitorRepo {
	return &MonitorRepo{}
}

func (u *MonitorRepo) GetBase(opts ...DBOption) ([]model.MonitorBase, error) {
	var data []model.MonitorBase
	db := global.MonitorDB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&data).Error
	return data, err
}
func (u *MonitorRepo) GetIO(opts ...DBOption) ([]model.MonitorIO, error) {
	var data []model.MonitorIO
	db := global.MonitorDB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&data).Error
	return data, err
}
func (u *MonitorRepo) GetNetwork(opts ...DBOption) ([]model.MonitorNetwork, error) {
	var data []model.MonitorNetwork
	db := global.MonitorDB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&data).Error
	return data, err
}

func (u *MonitorRepo) CreateMonitorBase(model model.MonitorBase) error {
	return global.MonitorDB.Create(&model).Error
}
func (u *MonitorRepo) BatchCreateMonitorIO(ioList []model.MonitorIO) error {
	return global.MonitorDB.CreateInBatches(ioList, len(ioList)).Error
}
func (u *MonitorRepo) BatchCreateMonitorNet(ioList []model.MonitorNetwork) error {
	return global.MonitorDB.CreateInBatches(ioList, len(ioList)).Error
}
func (u *MonitorRepo) DelMonitorBase(timeForDelete time.Time) error {
	return global.MonitorDB.Where("created_at < ?", timeForDelete).Delete(&model.MonitorBase{}).Error
}
func (u *MonitorRepo) DelMonitorIO(timeForDelete time.Time) error {
	return global.MonitorDB.Where("created_at < ?", timeForDelete).Delete(&model.MonitorIO{}).Error
}
func (u *MonitorRepo) DelMonitorNet(timeForDelete time.Time) error {
	return global.MonitorDB.Where("created_at < ?", timeForDelete).Delete(&model.MonitorNetwork{}).Error
}
