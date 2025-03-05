package repo

import (
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
	"gorm.io/gorm"
)

type IScriptRepo interface {
	Get(opts ...global.DBOption) (model.ScriptLibrary, error)
	GetList(opts ...global.DBOption) ([]model.ScriptLibrary, error)
	Create(snap *model.ScriptLibrary) error
	Update(id uint, vars map[string]interface{}) error
	Page(limit, offset int, opts ...global.DBOption) (int64, []model.ScriptLibrary, error)
	Delete(opts ...global.DBOption) error

	WithByInfo(info string) global.DBOption
}

func NewIScriptRepo() IScriptRepo {
	return &ScriptRepo{}
}

type ScriptRepo struct{}

func (u *ScriptRepo) Get(opts ...global.DBOption) (model.ScriptLibrary, error) {
	var ScriptLibrary model.ScriptLibrary
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&ScriptLibrary).Error
	return ScriptLibrary, err
}

func (u *ScriptRepo) GetList(opts ...global.DBOption) ([]model.ScriptLibrary, error) {
	var snaps []model.ScriptLibrary
	db := global.DB.Model(&model.ScriptLibrary{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&snaps).Error
	return snaps, err
}

func (u *ScriptRepo) Page(page, size int, opts ...global.DBOption) (int64, []model.ScriptLibrary, error) {
	var users []model.ScriptLibrary
	db := global.DB.Model(&model.ScriptLibrary{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&users).Error
	return count, users, err
}

func (u *ScriptRepo) Create(ScriptLibrary *model.ScriptLibrary) error {
	return global.DB.Create(ScriptLibrary).Error
}

func (u *ScriptRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.ScriptLibrary{}).Where("id = ?", id).Updates(vars).Error
}

func (u *ScriptRepo) Delete(opts ...global.DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.ScriptLibrary{}).Error
}

func (u *ScriptRepo) WithByInfo(info string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("name LIKE ? OR description LIKE ?", "%"+info+"%", "%"+info+"%")
	}
}
