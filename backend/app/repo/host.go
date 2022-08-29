package repo

import (
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/global"
	"gorm.io/gorm"
)

type HostRepo struct{}

type IHostRepo interface {
	Get(opts ...DBOption) (model.Host, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.Host, error)
	WithByInfo(info string) DBOption
	Create(host *model.Host) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error
}

func NewIHostService() IHostRepo {
	return &HostRepo{}
}

func (u *HostRepo) Get(opts ...DBOption) (model.Host, error) {
	var host model.Host
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&host).Error
	return host, err
}

func (u *HostRepo) Page(page, size int, opts ...DBOption) (int64, []model.Host, error) {
	var hosts []model.Host
	db := global.DB.Model(&model.Host{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&hosts).Error
	return count, hosts, err
}

func (c *HostRepo) WithByInfo(info string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		infoStr := "%" + info + "%"
		return g.Where("name LIKE ? OR addr LIKE ?", infoStr, infoStr)
	}
}

func (u *HostRepo) Create(host *model.Host) error {
	return global.DB.Create(host).Error
}

func (u *HostRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.Host{}).Where("id = ?", id).Updates(vars).Error
}

func (u *HostRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.Host{}).Error
}
