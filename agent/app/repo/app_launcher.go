package repo

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
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
	ListQuickDirectories() ([]model.QuickDirectory, error)
	CountQuickDirectories() (int64, error)
	CreateQuickDirectory(item *model.QuickDirectory) error
	UpdateQuickDirectory(item *model.QuickDirectory) error
	DeleteQuickDirectory(id uint, legacy bool) error
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
			{Name: "Agent", Title: "aiTools.agents.agent", Recommend: 1, IsShow: true, Router: "/ai/agents/agent"},
			{Name: "Website", Title: "menu.website", Recommend: 10, IsShow: true, Router: "/websites"},
			{Name: "Database", Title: "menu.database", Recommend: 30, IsShow: true, Router: "/databases"},
			{Name: "Cronjob", Title: "menu.cronjob", Recommend: 50, IsShow: false, Router: "/cronjobs"},
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
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// Legacy entries remain in the overview configuration until explicitly deleted.
func legacyQuickDirectories(db *gorm.DB) *gorm.DB {
	return db.Model(&model.QuickJump{}).Where("name = ? AND is_show = ? AND detail <> ?", "File", true, "")
}

func (u *LauncherRepo) ListQuickDirectories() ([]model.QuickDirectory, error) {
	items := make([]model.QuickDirectory, 0)
	if err := global.DB.Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	var quicks []model.QuickJump
	if err := legacyQuickDirectories(global.DB).Order("id ASC").Find(&quicks).Error; err != nil {
		return nil, err
	}
	for _, quick := range quicks {
		items = append(items, model.QuickDirectory{BaseModel: quick.BaseModel, Path: quick.Detail, Alias: quick.Alias, Legacy: true})
	}
	return items, nil
}
func (u *LauncherRepo) CountQuickDirectories() (int64, error) {
	var count, legacyCount int64
	if err := global.DB.Model(&model.QuickDirectory{}).Count(&count).Error; err != nil {
		return 0, err
	}
	if err := legacyQuickDirectories(global.DB).Count(&legacyCount).Error; err != nil {
		return 0, err
	}
	return count + legacyCount, nil
}
func (u *LauncherRepo) CreateQuickDirectory(item *model.QuickDirectory) error {
	return global.DB.Create(item).Error
}
func (u *LauncherRepo) UpdateQuickDirectory(item *model.QuickDirectory) error {
	var result *gorm.DB
	if item.Legacy {
		result = legacyQuickDirectories(global.DB).Where("id = ?", item.ID).Updates(map[string]interface{}{"detail": item.Path, "alias": item.Alias})
	} else {
		result = global.DB.Model(&model.QuickDirectory{}).Where("id = ?", item.ID).Updates(map[string]interface{}{"path": item.Path, "alias": item.Alias})
	}
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (u *LauncherRepo) DeleteQuickDirectory(id uint, legacy bool) error {
	var result *gorm.DB
	if legacy {
		result = legacyQuickDirectories(global.DB).Where("id = ?", id).Updates(map[string]interface{}{"detail": nil, "alias": nil})
	} else {
		result = global.DB.Delete(&model.QuickDirectory{}, id)
	}
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
