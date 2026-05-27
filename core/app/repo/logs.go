package repo

import (
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
	"gorm.io/gorm"
)

type LogRepo struct{}

type ILogRepo interface {
	CleanLogin() error
	CreateLoginLog(user *model.LoginLog) error
	PageLoginLog(limit, offset int, opts ...global.DBOption) (int64, []model.LoginLog, error)

	CleanOperation() error
	CreateOperationLog(user *model.OperationLog) error
	PageOperationLog(limit, offset int, opts ...global.DBOption) (int64, []model.OperationLog, error)

	WithByInfo(info string) global.DBOption
	WithBySource(source string) global.DBOption
	WithByLikeOperation(operation string) global.DBOption
}

func NewILogRepo() ILogRepo {
	return &LogRepo{}
}

func (u *LogRepo) CleanLogin() error {
	return global.DB.Exec("delete from login_logs;").Error
}

func (u *LogRepo) CreateLoginLog(log *model.LoginLog) error {
	return global.DB.Create(log).Error
}

func (u *LogRepo) PageLoginLog(page, size int, opts ...global.DBOption) (int64, []model.LoginLog, error) {
	var ops []model.LoginLog
	db := global.DB.Model(&model.LoginLog{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&ops).Error
	return count, ops, err
}

func (u *LogRepo) CleanOperation() error {
	return global.DB.Exec("delete from operation_logs").Error
}

func (u *LogRepo) CreateOperationLog(log *model.OperationLog) error {
	return global.DB.Create(log).Error
}

func (u *LogRepo) PageOperationLog(page, size int, opts ...global.DBOption) (int64, []model.OperationLog, error) {
	var ops []model.OperationLog
	db := global.DB.Model(&model.OperationLog{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&ops).Error
	return count, ops, err
}

func (c *LogRepo) WithByStatus(status string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(status) == 0 {
			return g
		}
		return g.Where("status = ?", status)
	}
}
func (c *LogRepo) WithBySource(source string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(source) == 0 {
			return g
		}
		return g.Where("source = ?", source)
	}
}
func (c *LogRepo) WithByInfo(info string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(info) == 0 {
			return g
		}
		infoStr := "%" + info + "%"
		return g.Where("ip LIKE ? OR user LIKE ?", infoStr, infoStr)
	}
}

func (c *LogRepo) WithByLikeOperation(operation string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(operation) == 0 {
			return g
		}
		infoStr := "%" + operation + "%"
		return g.Where("detail_zh LIKE ? OR detail_en LIKE ?", infoStr, infoStr)
	}
}
