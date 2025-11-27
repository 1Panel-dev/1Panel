package repo

import (
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
	"gorm.io/gorm"
)

type HostRepo struct{}

type IHostRepo interface {
	Get(opts ...global.DBOption) (model.Host, error)
	GetList(opts ...global.DBOption) ([]model.Host, error)
	Page(limit, offset int, opts ...global.DBOption) (int64, []model.Host, error)
	Create(host *model.Host) error
	Update(id uint, vars map[string]interface{}) error
	UpdateGroup(group, newGroup uint) error
	Delete(opts ...global.DBOption) error

	WithByInfo(info string) global.DBOption
	WithByPort(port uint) global.DBOption
	WithByUser(user string) global.DBOption
}

func NewIHostRepo() IHostRepo {
	return &HostRepo{}
}

func (h *HostRepo) Get(opts ...global.DBOption) (model.Host, error) {
	var host model.Host
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&host).Error
	return host, err
}

func (h *HostRepo) GetList(opts ...global.DBOption) ([]model.Host, error) {
	var hosts []model.Host
	db := global.DB.Model(&model.Host{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&hosts).Error
	return hosts, err
}

func (h *HostRepo) Page(page, size int, opts ...global.DBOption) (int64, []model.Host, error) {
	var users []model.Host
	db := global.DB.Model(&model.Host{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&users).Error
	return count, users, err
}

func (h *HostRepo) WithByInfo(info string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(info) == 0 {
			return g
		}
		infoStr := "%" + info + "%"
		return g.Where("name LIKE ? OR addr LIKE ?", infoStr, infoStr)
	}
}

func (h *HostRepo) WithByPort(port uint) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("port = ?", port)
	}
}
func (h *HostRepo) WithByUser(user string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("user = ?", user)
	}
}

func (h *HostRepo) Create(host *model.Host) error {
	return global.DB.Create(host).Error
}

func (h *HostRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.Host{}).Where("id = ?", id).Updates(vars).Error
}

func (h *HostRepo) UpdateGroup(group, newGroup uint) error {
	return global.DB.Model(&model.Host{}).Where("group_id = ?", group).Updates(map[string]interface{}{"group_id": newGroup}).Error
}

func (h *HostRepo) Delete(opts ...global.DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.Host{}).Error
}
