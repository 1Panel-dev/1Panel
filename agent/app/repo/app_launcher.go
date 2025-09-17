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

	GetQuickJump(opts ...DBOption) (model.QuickJump, error)
	ListQuickJump(withAll bool) []model.QuickJump
	UpdateQuicks(quicks []model.QuickJump) error
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

func (u *LauncherRepo) GetQuickJump(opts ...DBOption) (model.QuickJump, error) {
	var launcher model.QuickJump
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&launcher).Error
	return launcher, err
}
func (u *LauncherRepo) ListQuickJump(withAll bool) []model.QuickJump {
	var quicks []model.QuickJump
	if withAll {
		_ = global.DB.Find(&quicks).Error
	} else {
		_ = global.DB.Where("is_show = ?", true).Find(&quicks).Error
	}
	if !withAll && len(quicks) == 0 {
		return []model.QuickJump{
			{Name: "Website", Title: "menu.website", Recommend: 10, IsShow: true, Router: "/websites"},
			{Name: "Database", Title: "menu.database", Recommend: 30, IsShow: true, Router: "/databases"},
			{Name: "Cronjob", Title: "menu.cronjob", Recommend: 50, IsShow: true, Router: "/cronjobs"},
			{Name: "AppInstalled", Title: "home.appInstalled", Recommend: 70, IsShow: true, Router: "/apps/installed"},
		}
	}

	return quicks
}
func (u *LauncherRepo) UpdateQuicks(quicks []model.QuickJump) error {
	tx := global.DB.Begin()
	for _, item := range quicks {
		if err := tx.Model(&model.QuickJump{}).Where("id = ?", item.ID).Updates(map[string]interface{}{
			"is_show": item.IsShow,
			"detail":  item.Detail,
			"alias":   item.Alias,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}
