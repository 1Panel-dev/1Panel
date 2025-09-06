package repo

import (
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClamRepo struct{}

type IClamRepo interface {
	Page(limit, offset int, opts ...DBOption) (int64, []model.Clam, error)
	Create(clam *model.Clam) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error
	Get(opts ...DBOption) (model.Clam, error)
	List(opts ...DBOption) ([]model.Clam, error)

	WithByClamID(id uint) DBOption
	ListRecord(opts ...DBOption) ([]model.ClamRecord, error)
	RecordFirst(id uint) (model.ClamRecord, error)
	DeleteRecord(opts ...DBOption) error
	StartRecords(clamID uint) model.ClamRecord
	EndRecords(record model.ClamRecord, status, message string)
	PageRecords(page, size int, opts ...DBOption) (int64, []model.ClamRecord, error)
}

func NewIClamRepo() IClamRepo {
	return &ClamRepo{}
}

func (u *ClamRepo) Get(opts ...DBOption) (model.Clam, error) {
	var clam model.Clam
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&clam).Error
	return clam, err
}

func (u *ClamRepo) List(opts ...DBOption) ([]model.Clam, error) {
	var clam []model.Clam
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&clam).Error
	return clam, err
}

func (u *ClamRepo) Page(page, size int, opts ...DBOption) (int64, []model.Clam, error) {
	var users []model.Clam
	db := global.DB.Model(&model.Clam{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&users).Error
	return count, users, err
}

func (u *ClamRepo) Create(clam *model.Clam) error {
	return global.DB.Create(clam).Error
}

func (u *ClamRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.Clam{}).Where("id = ?", id).Updates(vars).Error
}

func (u *ClamRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.Clam{}).Error
}

func (c *ClamRepo) WithByClamID(id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("clam_id = ?", id)
	}
}

func (u *ClamRepo) ListRecord(opts ...DBOption) ([]model.ClamRecord, error) {
	var record []model.ClamRecord
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&record).Error
	return record, err
}

func (u *ClamRepo) RecordFirst(id uint) (model.ClamRecord, error) {
	var record model.ClamRecord
	err := global.DB.Where("clam_id = ?", id).Order("created_at desc").First(&record).Error
	return record, err
}

func (u *ClamRepo) PageRecords(page, size int, opts ...DBOption) (int64, []model.ClamRecord, error) {
	var records []model.ClamRecord
	db := global.DB.Model(&model.ClamRecord{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Order("created_at desc").Limit(size).Offset(size * (page - 1)).Find(&records).Error
	return count, records, err
}
func (u *ClamRepo) StartRecords(clamID uint) model.ClamRecord {
	var record model.ClamRecord
	record.StartTime = time.Now()
	record.ClamID = clamID
	record.TaskID = uuid.New().String()
	record.Status = constant.StatusWaiting
	if err := global.DB.Create(&record).Error; err != nil {
		global.LOG.Errorf("create record status failed, err: %v", err)
	}
	_ = u.Update(clamID, map[string]interface{}{"is_executing": true})
	return record
}
func (u *ClamRepo) EndRecords(record model.ClamRecord, status, message string) {
	upMap := make(map[string]interface{})
	upMap["status"] = status
	upMap["message"] = message
	upMap["task_id"] = record.TaskID
	upMap["scan_time"] = record.ScanTime
	upMap["infected_files"] = record.InfectedFiles
	upMap["total_error"] = record.TotalError
	if err := global.DB.Model(&model.ClamRecord{}).Where("id = ?", record.ID).Updates(upMap).Error; err != nil {
		global.LOG.Errorf("update record status failed, err: %v", err)
	}
	_ = u.Update(record.ClamID, map[string]interface{}{"is_executing": false})
}
func (u *ClamRepo) DeleteRecord(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.ClamRecord{}).Error
}
