package repo

import (
	"gorm.io/gorm"
)

type DBOption func(*gorm.DB) *gorm.DB

type ICommonRepo interface {
	WithOrderBy(orderStr string) DBOption
}

type CommonRepo struct{}

func NewCommonRepo() ICommonRepo {
	return &CommonRepo{}
}
func (c *CommonRepo) WithOrderBy(orderStr string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Order(orderStr)
	}
}
