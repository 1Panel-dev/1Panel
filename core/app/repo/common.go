package repo

import (
	"gorm.io/gorm"
)

type DBOption func(*gorm.DB) *gorm.DB

type ICommonRepo interface {
	WithByID(id uint) DBOption
	WithByName(name string) DBOption
	WithByIDs(ids []uint) DBOption
	WithByType(ty string) DBOption
	WithOrderBy(orderStr string) DBOption
}

type CommonRepo struct{}

func NewICommonRepo() ICommonRepo {
	return &CommonRepo{}
}

func (c *CommonRepo) WithByID(id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id = ?", id)
	}
}
func (c *CommonRepo) WithByName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(name) == 0 {
			return g
		}
		return g.Where("`name` = ?", name)
	}
}
func (c *CommonRepo) WithByIDs(ids []uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id in (?)", ids)
	}
}
func (c *CommonRepo) WithByType(ty string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(ty) == 0 {
			return g
		}
		return g.Where("`type` = ?", ty)
	}
}
func (c *CommonRepo) WithOrderBy(orderStr string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Order(orderStr)
	}
}
