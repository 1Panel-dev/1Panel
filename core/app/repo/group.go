package repo

import (
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
	"gorm.io/gorm"
)

type GroupRepo struct{}

type IGroupRepo interface {
	Get(opts ...DBOption) (model.Group, error)
	GetList(opts ...DBOption) ([]model.Group, error)
	Create(group *model.Group) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error

	WithByID(id uint) DBOption
	WithByGroupID(id uint) DBOption
	WithByGroupType(ty string) DBOption
	WithByDefault(isDefalut bool) DBOption
	CancelDefault(groupType string) error
}

func NewIGroupRepo() IGroupRepo {
	return &GroupRepo{}
}

func (c *GroupRepo) WithByID(id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id = ?", id)
	}
}

func (c *GroupRepo) WithByGroupID(id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("group_id = ?", id)
	}
}

func (c *GroupRepo) WithByGroupType(ty string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("`type` = ?", ty)
	}
}

func (c *GroupRepo) WithByDefault(isDefalut bool) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("is_default = ?", isDefalut)
	}
}

func (u *GroupRepo) Get(opts ...DBOption) (model.Group, error) {
	var group model.Group
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&group).Error
	return group, err
}

func (u *GroupRepo) GetList(opts ...DBOption) ([]model.Group, error) {
	var groups []model.Group
	db := global.DB.Model(&model.Group{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&groups).Error
	return groups, err
}

func (u *GroupRepo) Create(group *model.Group) error {
	return global.DB.Create(group).Error
}

func (u *GroupRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.Group{}).Where("id = ?", id).Updates(vars).Error
}

func (u *GroupRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.Group{}).Error
}

func (u *GroupRepo) CancelDefault(groupType string) error {
	return global.DB.Model(&model.Group{}).
		Where("is_default = ? AND type = ?", 1, groupType).
		Updates(map[string]interface{}{"is_default": 0}).Error
}
