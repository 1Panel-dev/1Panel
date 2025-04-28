package repo

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

type GroupRepo struct{}

type IGroupRepo interface {
	Get(opts ...DBOption) (model.Group, error)
	GetList(opts ...DBOption) ([]model.Group, error)
	Create(group *model.Group) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error
	WithByDefault(isDefault bool) DBOption
	WithByWebsiteDefault() DBOption
}

func NewIGroupRepo() IGroupRepo {
	return &GroupRepo{}
}

func (g *GroupRepo) WithByDefault(isDefault bool) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("is_default = ?", isDefault)
	}
}

func (g *GroupRepo) Get(opts ...DBOption) (model.Group, error) {
	var group model.Group
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&group).Error
	return group, err
}

func (g *GroupRepo) GetList(opts ...DBOption) ([]model.Group, error) {
	var groups []model.Group
	db := global.DB.Model(&model.Group{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&groups).Error
	return groups, err
}

func (g *GroupRepo) Create(group *model.Group) error {
	return global.DB.Create(group).Error
}

func (g *GroupRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.Group{}).Where("id = ?", id).Updates(vars).Error
}

func (g *GroupRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.Group{}).Error
}

func (g *GroupRepo) WithByWebsiteDefault() DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("is_default = ? AND type = ?", 1, "website")
	}
}
