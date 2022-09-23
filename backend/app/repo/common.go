package repo

import "gorm.io/gorm"

type DBOption func(*gorm.DB) *gorm.DB

type ICommonRepo interface {
	WithByID(id uint) DBOption
	WithByName(name string) DBOption
	WithOrderBy(orderStr string) DBOption
	WithLikeName(name string) DBOption
	WithIdsIn(ids []uint) DBOption
}

type CommonRepo struct{}

func (c *CommonRepo) WithByID(id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id = ?", id)
	}
}

func (c *CommonRepo) WithByName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("name = ?", name)
	}
}

func (c *CommonRepo) WithByType(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("type = ?", name)
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

func (c *CommonRepo) WithLikeName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("name like ?", "%"+name+"%")
	}
}

func (c *CommonRepo) WithOrderBy(orderStr string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Order(orderStr)
	}
}

func (c *CommonRepo) WithIdsIn(ids []uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id in (?)", ids)
	}
}
