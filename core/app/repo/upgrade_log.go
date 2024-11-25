package repo

import (
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
	"gorm.io/gorm"
)

type UpgradeLogRepo struct{}

type IUpgradeLogRepo interface {
	Get(opts ...DBOption) (model.UpgradeLog, error)
	List(opts ...DBOption) ([]model.UpgradeLog, error)
	Create(log *model.UpgradeLog) error
	Page(limit, offset int, opts ...DBOption) (int64, []model.UpgradeLog, error)
	Delete(opts ...DBOption) error

	WithByNodeID(nodeID uint) DBOption
	WithByIDs(ids []uint) DBOption
	WithByID(id uint) DBOption
}

func NewIUpgradeLogRepo() IUpgradeLogRepo {
	return &UpgradeLogRepo{}
}

func (u *UpgradeLogRepo) Get(opts ...DBOption) (model.UpgradeLog, error) {
	var log model.UpgradeLog
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&log).Error
	return log, err
}

func (u *UpgradeLogRepo) List(opts ...DBOption) ([]model.UpgradeLog, error) {
	var logs []model.UpgradeLog
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&logs).Error
	return logs, err
}

func (u *UpgradeLogRepo) Clean() error {
	return global.DB.Exec("delete from upgrade_logs;").Error
}

func (u *UpgradeLogRepo) Create(log *model.UpgradeLog) error {
	return global.DB.Create(log).Error
}

func (u *UpgradeLogRepo) Save(log *model.UpgradeLog) error {
	return global.DB.Save(log).Error
}

func (u *UpgradeLogRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.UpgradeLog{}).Error
}

func (u *UpgradeLogRepo) Page(page, size int, opts ...DBOption) (int64, []model.UpgradeLog, error) {
	var ops []model.UpgradeLog
	db := global.DB.Model(&model.UpgradeLog{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&ops).Error
	return count, ops, err
}

func (c *UpgradeLogRepo) WithByNodeID(nodeID uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("node_id = ?", nodeID)
	}
}
func (c *UpgradeLogRepo) WithByID(id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id = ?", id)
	}
}
func (c *UpgradeLogRepo) WithByIDs(ids []uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id in (?)", ids)
	}
}
