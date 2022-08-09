package repo

import (
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/global"
)

type OperationRepo struct{}

type IOperationRepo interface {
	Create(user *model.OperationLog) error
}

func NewIOperationService() IOperationRepo {
	return &OperationRepo{}
}

func (u *OperationRepo) Create(user *model.OperationLog) error {
	return global.DB.Create(user).Error
}
