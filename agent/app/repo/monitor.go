package repo

import (
	"fmt"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

type MonitorRepo struct{}

type GPUHistoryPoint struct {
	model.MonitorGPU
	Bucket        int64
	PowerPercent  *float64
	MemoryPercent *float64
	ProcessCount  *float64
}

type IMonitorRepo interface {
	GetBase(opts ...DBOption) ([]model.MonitorBase, error)
	GetGPU(opts ...DBOption) ([]model.MonitorGPU, error)
	CountGPU(opts ...DBOption) (int64, error)
	GetGPUHistory(start time.Time, bucketSeconds int64, aggregation string, opts ...DBOption) ([]GPUHistoryPoint, error)
	GetGPUDevices() ([]model.MonitorGPU, error)
	GetIO(opts ...DBOption) ([]model.MonitorIO, error)
	GetNetwork(opts ...DBOption) ([]model.MonitorNetwork, error)

	CreateMonitorBase(model model.MonitorBase) error
	BatchCreateMonitorGPU(list []model.MonitorGPU) error
	BatchCreateMonitorIO(ioList []model.MonitorIO) error
	BatchCreateMonitorNet(ioList []model.MonitorNetwork) error
	DelMonitorBase(timeForDelete time.Time) error
	DelMonitorGPU(timeForDelete time.Time) error
	DelMonitorIO(timeForDelete time.Time) error
	DelMonitorNet(timeForDelete time.Time) error

	WithByProductName(name string) DBOption
	WithByGPUDevice(deviceID, name string, legacy bool) DBOption
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
func (u *MonitorRepo) GetGPU(opts ...DBOption) ([]model.MonitorGPU, error) {
	var data []model.MonitorGPU
	db := global.GPUMonitorDB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&data).Error
	return data, err
}

func (u *MonitorRepo) CreateMonitorBase(model model.MonitorBase) error {
	return global.MonitorDB.Create(&model).Error
}
func (s *MonitorRepo) BatchCreateMonitorGPU(list []model.MonitorGPU) error {
	if len(list) == 0 {
		return nil
	}
	return global.GPUMonitorDB.CreateInBatches(&list, len(list)).Error
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
func (s *MonitorRepo) DelMonitorGPU(timeForDelete time.Time) error {
	return global.GPUMonitorDB.Where("created_at < ?", timeForDelete).Delete(&model.MonitorGPU{}).Error
}

func (s *MonitorRepo) WithByProductName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("product_name = ?", name)
	}
}

func (u *MonitorRepo) GetGPUDevices() ([]model.MonitorGPU, error) {
	var data []model.MonitorGPU
	err := global.GPUMonitorDB.Model(&model.MonitorGPU{}).Select("device_id, product_name, device_type").Group("device_id, product_name, device_type").Order("product_name, device_id").Find(&data).Error
	return data, err
}

func (u *MonitorRepo) WithByGPUDevice(deviceID, name string, legacy bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if deviceID != "" {
			return db.Where("device_id = ?", deviceID)
		}
		db = db.Where("product_name = ?", name)
		if legacy {
			db = db.Where("device_id IS NULL OR device_id = ''")
		}
		return db
	}
}

func (u *MonitorRepo) CountGPU(opts ...DBOption) (int64, error) {
	db := global.GPUMonitorDB.Model(&model.MonitorGPU{})
	for _, opt := range opts {
		db = opt(db)
	}
	var count int64
	err := db.Count(&count).Error
	return count, err
}

func (u *MonitorRepo) GetGPUHistory(start time.Time, bucketSeconds int64, aggregation string, opts ...DBOption) ([]GPUHistoryPoint, error) {
	db := global.GPUMonitorDB.Model(&model.MonitorGPU{})
	for _, opt := range opts {
		db = opt(db)
	}
	expressions := []string{
		"CASE WHEN max_power_limit > 0 THEN 100.0 * power_draw / max_power_limit END",
		"CASE WHEN mem_total > 0 AND mem_used IS NOT NULL THEN 100.0 * mem_used / mem_total ELSE memory_util END",
		"CASE WHEN (process_status = 'ok' OR process_status IS NULL OR process_status = '') AND json_valid(processes) THEN CASE WHEN json_type(processes) = 'array' THEN json_array_length(processes) END END",
	}
	aliases := []string{"power_percent", "memory_percent", "process_count"}
	columns := []string{"*"}
	if bucketSeconds > 0 {
		operation := "AVG"
		if aggregation == "max" {
			operation = "MAX"
		}
		columns = []string{fmt.Sprintf("(CAST(strftime('%%s', created_at) AS INTEGER) - %d) / %d AS bucket", start.Unix(), bucketSeconds)}
		for _, column := range []string{"memory_activity", "encoder_util", "decoder_util", "jpeg_util", "ofa_util", "media_util", "compute_util", "copy_util", "hotspot_temperature", "fan_rpm", "ai_cpu_util", "ctrl_cpu_util", "ddr_used", "ddr_total", "hbm_used", "hbm_total", "ddr_bandwidth", "hbm_bandwidth", "memory_bandwidth", "media_frequency", "hugepages_used", "hugepages_total", "gpu_util", "temperature", "memory_temperature", "power_draw", "max_power_limit", "mem_used", "mem_total", "frequency", "memory_frequency", "fan_speed"} {
			columns = append(columns, operation+"("+column+") AS "+column)
		}
		for i := range expressions {
			expressions[i] = operation + "(" + expressions[i] + ")"
		}
		db = db.Group("bucket").Order("bucket ASC")
	} else {
		db = db.Order("created_at ASC, id ASC")
	}
	for i, expression := range expressions {
		columns = append(columns, expression+" AS "+aliases[i])
	}
	var data []GPUHistoryPoint
	err := db.Select(strings.Join(columns, ", ")).Scan(&data).Error
	return data, err
}
