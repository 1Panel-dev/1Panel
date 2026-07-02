package repo

import (
	"fmt"
	"time"

	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/re"
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
func WithByStringIDs(ids []string) global.DBOption {
	var idItems []uint
	for _, id := range ids {
		var idItem uint
		if _, err := fmt.Sscanf(id, "%d", &idItem); err == nil && idItem != 0 {
			idItems = append(idItems, idItem)
		}
	}
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id IN (?)", idItems)
	}
}
func WithByName(name string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("`name` = ?", name)
	}
}
func WithByLikeName(name string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(name) == 0 {
			return g
		}
		return g.Where("name like ?", "%"+name+"%")
	}
}
func WithByUserID(userID string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("user_id = ?", userID)
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

func WithByCreatedAt(startTime, endTime time.Time) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("created_at > ? AND created_at < ?", startTime, endTime)
	}
}

func WithByNode(node string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("node = ?", node)
	}
}

func WithOrderDesc(orderBy string) global.DBOption {
	return WithOrderRuleBy(orderBy, constant.OrderDesc)
}

func WithOrderAsc(orderBy string) global.DBOption {
	return WithOrderRuleBy(orderBy, constant.OrderAsc)
}

func WithOrderRuleBy(orderBy, order string) global.DBOption {
	if orderBy == "createdAt" {
		orderBy = "created_at"
	}
	if !re.GetRegex(re.OrderByValidationPattern).MatchString(orderBy) {
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
