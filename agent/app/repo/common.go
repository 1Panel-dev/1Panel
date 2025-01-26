package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

type DBOption func(*gorm.DB) *gorm.DB

func WithByID(id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id = ?", id)
	}
}

func WithByNOTID(id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id != ?", id)
	}
}

func WithByIDs(ids []uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id in (?)", ids)
	}
}

func WithByIDNotIn(ids []uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id not in (?)", ids)
	}
}

func WithByName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("name = ?", name)
	}
}

func WithByLikeName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(name) == 0 {
			return g
		}
		return g.Where("name like ?", "%"+name+"%")
	}
}

func WithByDetailName(detailName string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(detailName) == 0 {
			return g
		}
		return g.Where("detail_name = ?", detailName)
	}
}

func WithByType(tp string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("type = ?", tp)
	}
}

func WithTypes(types []string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("type in (?)", types)
	}
}

func WithByStatus(status string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(status) == 0 {
			return g
		}
		return g.Where("status = ?", status)
	}
}
func WithByFrom(from string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("`from` = ?", from)
	}
}

func WithByDate(startTime, endTime time.Time) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("start_time > ? AND start_time < ?", startTime, endTime)
	}
}

func WithByCreatedAt(startTime, endTime time.Time) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("created_at > ? AND created_at < ?", startTime, endTime)
	}
}

func WithOrderBy(orderStr string) DBOption {
	if orderStr == "createdAt" {
		orderStr = "created_at"
	}
	return func(g *gorm.DB) *gorm.DB {
		return g.Order(orderStr)
	}
}
func WithOrderRuleBy(orderBy, order string) DBOption {
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

func getTx(ctx context.Context, opts ...DBOption) *gorm.DB {
	tx, ok := ctx.Value(constant.DB).(*gorm.DB)
	if ok {
		for _, opt := range opts {
			tx = opt(tx)
		}
		return tx
	}
	return getDb(opts...)
}

func getDb(opts ...DBOption) *gorm.DB {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db
}
