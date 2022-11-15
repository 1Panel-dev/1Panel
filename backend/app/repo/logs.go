package repo

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
)

type LogRepo struct{}

type ILogRepo interface {
	CreateLoginLog(user *model.LoginLog) error
	PageLoginLog(limit, offset int, opts ...DBOption) (int64, []model.LoginLog, error)

	CreateOperationLog(user *model.OperationLog) error
	PageOperationLog(limit, offset int, opts ...DBOption) (int64, []model.OperationLog, error)
}

func NewILogRepo() ILogRepo {
	return &LogRepo{}
}

func (u *LogRepo) CreateLoginLog(log *model.LoginLog) error {
	return global.DB.Create(log).Error
}

func (u *LogRepo) PageLoginLog(page, size int, opts ...DBOption) (int64, []model.LoginLog, error) {
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

func (u *LogRepo) CreateOperationLog(log *model.OperationLog) error {
	return global.DB.Create(log).Error
}

func (u *LogRepo) PageOperationLog(page, size int, opts ...DBOption) (int64, []model.OperationLog, error) {
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
