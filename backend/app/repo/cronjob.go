package repo

import (
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
)

type CronjobRepo struct{}

type ICronjobRepo interface {
	Get(opts ...DBOption) (model.Cronjob, error)
	GetRecord(opts ...DBOption) (model.JobRecords, error)
	RecordFirst(id uint) (model.JobRecords, error)
	ListRecord(opts ...DBOption) ([]model.JobRecords, error)
	List(opts ...DBOption) ([]model.Cronjob, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.Cronjob, error)
	Create(cronjob *model.Cronjob) error
	WithByJobID(id int) DBOption
	WithByDbName(name string) DBOption
	WithByDefaultDownload(account string) DBOption
	WithByRecordDropID(id int) DBOption
	WithByRecordFile(file string) DBOption
	Save(id uint, cronjob model.Cronjob) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error
	DeleteRecord(opts ...DBOption) error
	StartRecords(cronjobID uint, fromLocal bool, targetPath string) model.JobRecords
	UpdateRecords(id uint, vars map[string]interface{}) error
	EndRecords(record model.JobRecords, status, message, records string)
	PageRecords(page, size int, opts ...DBOption) (int64, []model.JobRecords, error)
}

func NewICronjobRepo() ICronjobRepo {
	return &CronjobRepo{}
}

func (u *CronjobRepo) Get(opts ...DBOption) (model.Cronjob, error) {
	var cronjob model.Cronjob
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&cronjob).Error
	return cronjob, err
}

func (u *CronjobRepo) GetRecord(opts ...DBOption) (model.JobRecords, error) {
	var record model.JobRecords
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&record).Error
	return record, err
}

func (u *CronjobRepo) List(opts ...DBOption) ([]model.Cronjob, error) {
	var cronjobs []model.Cronjob
	db := global.DB.Model(&model.Cronjob{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&cronjobs).Error
	return cronjobs, err
}

func (u *CronjobRepo) ListRecord(opts ...DBOption) ([]model.JobRecords, error) {
	var cronjobs []model.JobRecords
	db := global.DB.Model(&model.JobRecords{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&cronjobs).Error
	return cronjobs, err
}

func (u *CronjobRepo) Page(page, size int, opts ...DBOption) (int64, []model.Cronjob, error) {
	var cronjobs []model.Cronjob
	db := global.DB.Model(&model.Cronjob{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&cronjobs).Error
	return count, cronjobs, err
}

func (u *CronjobRepo) RecordFirst(id uint) (model.JobRecords, error) {
	var record model.JobRecords
	err := global.DB.Where("cronjob_id = ?", id).Order("created_at desc").First(&record).Error
	return record, err
}

func (u *CronjobRepo) PageRecords(page, size int, opts ...DBOption) (int64, []model.JobRecords, error) {
	var cronjobs []model.JobRecords
	db := global.DB.Model(&model.JobRecords{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Order("created_at desc").Limit(size).Offset(size * (page - 1)).Find(&cronjobs).Error
	return count, cronjobs, err
}

func (u *CronjobRepo) Create(cronjob *model.Cronjob) error {
	return global.DB.Create(cronjob).Error
}

func (c *CronjobRepo) WithByJobID(id int) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("cronjob_id = ?", id)
	}
}

func (c *CronjobRepo) WithByDbName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("db_name = ?", name)
	}
}

func (c *CronjobRepo) WithByDefaultDownload(account string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("default_download = ?", account)
	}
}

func (c *CronjobRepo) WithByRecordFile(file string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("records = ?", file)
	}
}

func (c *CronjobRepo) WithByRecordDropID(id int) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id < ?", id)
	}
}

func (u *CronjobRepo) StartRecords(cronjobID uint, fromLocal bool, targetPath string) model.JobRecords {
	var record model.JobRecords
	record.StartTime = time.Now()
	record.CronjobID = cronjobID
	record.FromLocal = fromLocal
	record.Status = constant.StatusWaiting
	if err := global.DB.Create(&record).Error; err != nil {
		global.LOG.Errorf("create record status failed, err: %v", err)
	}
	return record
}
func (u *CronjobRepo) EndRecords(record model.JobRecords, status, message, records string) {
	errMap := make(map[string]interface{})
	errMap["records"] = records
	errMap["status"] = status
	errMap["file"] = record.File
	errMap["message"] = message
	errMap["interval"] = time.Since(record.StartTime).Milliseconds()
	if err := global.DB.Model(&model.JobRecords{}).Where("id = ?", record.ID).Updates(errMap).Error; err != nil {
		global.LOG.Errorf("update record status failed, err: %v", err)
	}
}

func (u *CronjobRepo) Save(id uint, cronjob model.Cronjob) error {
	return global.DB.Model(&model.Cronjob{}).Where("id = ?", id).Save(&cronjob).Error
}
func (u *CronjobRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.Cronjob{}).Where("id = ?", id).Updates(vars).Error
}

func (u *CronjobRepo) UpdateRecords(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.JobRecords{}).Where("id = ?", id).Updates(vars).Error
}

func (u *CronjobRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.Cronjob{}).Error
}
func (u *CronjobRepo) DeleteRecord(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.JobRecords{}).Error
}
