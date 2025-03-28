package repo

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
)

type LauncherRepo struct{}

type ILauncherRepo interface {
	Get(opts ...DBOption) (model.AppLauncher, error)
	ListName(opts ...DBOption) ([]string, error)
	Create(launcher *model.AppLauncher) error
	Save(launcher *model.AppLauncher) error
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
func (u *LauncherRepo) ListName(opts ...DBOption) ([]string, error) {
	var ops []model.AppLauncher
	db := global.DB.Model(&model.AppLauncher{})
	for _, opt := range opts {
		db = opt(db)
	}
	_ = db.Find(&ops).Error
	var names []string
	for i := 0; i < len(ops); i++ {
		names = append(names, ops[i].Key)
	}
	return names, nil
}

func (u *LauncherRepo) Create(launcher *model.AppLauncher) error {
	return global.DB.Create(launcher).Error
}

func (u *LauncherRepo) Save(launcher *model.AppLauncher) error {
	return global.DB.Save(launcher).Error
}

func (u *LauncherRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.AppLauncher{}).Error
}
