package repo

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
)

type LauncherRepo struct{}

type ILauncherRepo interface {
	Get(opts ...DBOption) (model.AppLauncher, error)
	List(opts ...DBOption) ([]model.AppLauncher, error)
	Create(launcher *model.AppLauncher) error
	Save(launcher *model.AppLauncher) error
	Delete(opts ...DBOption) error

	SyncAll(data []model.AppLauncher) error
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

func (u *LauncherRepo) SyncAll(data []model.AppLauncher) error {
	tx := global.DB.Begin()
	if err := tx.Where("1 = 1").Delete(&model.AppLauncher{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if len(data) == 0 {
		tx.Commit()
		return nil
	}
	if err := tx.Model(model.AppLauncher{}).Save(&data).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
