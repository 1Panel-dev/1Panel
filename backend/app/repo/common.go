package repo

import "gorm.io/gorm"

type DBOption func(*gorm.DB) *gorm.DB

type ICommonRepo interface {
	WithByID(id uint) DBOption
	WithByName(name string) DBOption
	WithOrderBy(orderStr string) DBOption
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

func (c *CommonRepo) WithOrderBy(orderStr string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Order(orderStr)
	}
}
