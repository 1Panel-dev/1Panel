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

type ICommonRepo interface {
	WithByID(id uint) DBOption
	WithByIDs(ids []uint) DBOption
	WithByName(name string) DBOption
	WithByNames(names []string) DBOption
	WithByLikeName(name string) DBOption
	WithByDetailName(detailName string) DBOption

	WithByFrom(from string) DBOption
	WithByType(tp string) DBOption
	WithTypes(types []string) DBOption
	WithByStatus(status string) DBOption

	WithOrderBy(orderStr string) DBOption
	WithOrderRuleBy(orderBy, order string) DBOption

	WithByDate(startTime, endTime time.Time) DBOption
	WithByCreatedAt(startTime, endTime time.Time) DBOption
}

type CommonRepo struct{}

func NewCommonRepo() ICommonRepo {
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
		return g.Where("name = ?", name)
	}
}
func (c *CommonRepo) WithByNames(names []string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("name in (?)", names)
	}
}
func (c *CommonRepo) WithByLikeName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(name) == 0 {
			return g
		}
		return g.Where("name like ?", "%"+name+"%")
	}
}
func (c *CommonRepo) WithByDetailName(detailName string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(detailName) == 0 {
			return g
		}
		return g.Where("detail_name = ?", detailName)
	}
}

func (c *CommonRepo) WithByType(tp string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("type = ?", tp)
	}
}
func (c *CommonRepo) WithTypes(types []string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("type in (?)", types)
	}
}
func (c *CommonRepo) WithByStatus(status string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(status) == 0 {
			return g
		}
		return g.Where("status = ?", status)
	}
}
func (c *CommonRepo) WithByFrom(from string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("`from` = ?", from)
	}
}

func (c *CommonRepo) WithByDate(startTime, endTime time.Time) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("start_time > ? AND start_time < ?", startTime, endTime)
	}
}

func (c *CommonRepo) WithByCreatedAt(startTime, endTime time.Time) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("created_at > ? AND created_at < ?", startTime, endTime)
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
