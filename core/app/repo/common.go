package repo

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/core/constant"
	"gorm.io/gorm"
)

type DBOption func(*gorm.DB) *gorm.DB

type ICommonRepo interface {
	WithByID(id uint) DBOption
	WithByIDs(ids []uint) DBOption
	WithByName(name string) DBOption
	WithLikeName(name string) DBOption
	WithByType(ty string) DBOption
	WithByKey(key string) DBOption
	WithOrderBy(orderStr string) DBOption
	WithByStatus(status string) DBOption

	WithOrderRuleBy(orderBy, order string) DBOption
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
func (c *CommonRepo) WithByIDs(ids []uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id in (?)", ids)
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
func (c *CommonRepo) WithLikeName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(name) == 0 {
			return g
		}
		return g.Where("name like ? or command like ?", "%"+name+"%", "%"+name+"%")
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
func (c *CommonRepo) WithByKey(key string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("key = ?", key)
	}
}
func (c *CommonRepo) WithByStatus(status string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("status = ?", status)
	}
}
func (c *CommonRepo) WithOrderBy(orderStr string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Order(orderStr)
	}
}

func (c *CommonRepo) WithOrderRuleBy(orderBy, order string) DBOption {
	switch order {
	case constant.OrderDesc:
		order = "desc"
	case constant.OrderAsc:
		order = "asc"
	default:
		orderBy = "created_at"
		order = "desc"
	}
	return func(g *gorm.DB) *gorm.DB {
		return g.Order(fmt.Sprintf("%s %s", orderBy, order))
	}
}
