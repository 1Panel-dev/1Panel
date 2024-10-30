package repo

import (
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
)

type LauncherRepo struct{}

type ILauncherRepo interface {
	Get(opts ...DBOption) (model.AppLauncher, error)
	List(opts ...DBOption) ([]model.AppLauncher, error)
	Create(launcher *model.AppLauncher) error
	Delete(opts ...DBOption) error
}

func NewILauncherRepo() ILauncherRepo {
	return &LauncherRepo{}
}

func (u *LauncherRepo) Get(opts ...DBOption) (model.AppLauncher, error) {
	var launcher model.AppLauncher
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&launcher).Error
	return launcher, err
}
func (u *LauncherRepo) List(opts ...DBOption) ([]model.AppLauncher, error) {
	var ops []model.AppLauncher
	db := global.DB.Model(&model.AppLauncher{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&ops).Error
	return ops, err
}

func (u *LauncherRepo) Create(launcher *model.AppLauncher) error {
	return global.DB.Create(launcher).Error
}

func (u *LauncherRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.AppLauncher{}).Error
}
