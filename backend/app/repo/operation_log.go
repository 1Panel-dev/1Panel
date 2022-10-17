package repo

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
)

type OperationRepo struct{}

type IOperationRepo interface {
	Create(user *model.OperationLog) error
	Page(limit, offset int, opts ...DBOption) (int64, []model.OperationLog, error)
	Delete(opts ...DBOption) error
}

func NewIOperationRepo() IOperationRepo {
	return &OperationRepo{}
}

func (u *OperationRepo) Create(user *model.OperationLog) error {
	return global.DB.Create(user).Error
}

func (u *OperationRepo) Page(page, size int, opts ...DBOption) (int64, []model.OperationLog, error) {
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

func (u *OperationRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}

	return db.Delete(&model.OperationLog{}).Error
}
