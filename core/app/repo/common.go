package repo

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"gorm.io/gorm"
)

func WithByID(id uint) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id = ?", id)
	}
}
func WithByGroupID(id uint) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("group_id = ?", id)
	}
}

func WithByIDs(ids []uint) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id in (?)", ids)
	}
}
func WithByName(name string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("`name` = ?", name)
	}
}
func WithoutByName(name string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("`name` != ?", name)
	}
}

func WithByType(ty string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("`type` = ?", ty)
	}
}
func WithByAddr(addr string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("addr = ?", addr)
	}
}
func WithByKey(key string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("key = ?", key)
	}
}
func WithByStatus(status string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("status = ?", status)
	}
}
func WithByNode(node string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("node = ?", node)
	}
}
func WithByGroupBelong(group string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("group_belong = ?", group)
	}
}

func WithOrderBy(orderStr string) global.DBOption {
	if orderStr == "createdAt" {
		orderStr = "created_at"
	}
	return func(g *gorm.DB) *gorm.DB {
		return g.Order(orderStr)
	}
}

func WithOrderRuleBy(orderBy, order string) global.DBOption {
	if orderBy == "createdAt" {
		orderBy = "created_at"
	}
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
