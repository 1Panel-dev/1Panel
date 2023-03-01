package repo

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
)

type HostRepo struct{}

type IHostRepo interface {
	Get(opts ...DBOption) (model.Host, error)
	GetList(opts ...DBOption) ([]model.Host, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.Host, error)
	WithByInfo(info string) DBOption
	WithByPort(port uint) DBOption
	WithByUser(user string) DBOption
	WithByAddr(addr string) DBOption
	WithByGroup(group string) DBOption
	Create(host *model.Host) error
	ChangeGroup(oldGroup, newGroup string) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error
}

func NewIHostRepo() IHostRepo {
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

func (u *HostRepo) GetList(opts ...DBOption) ([]model.Host, error) {
	var hosts []model.Host
	db := global.DB.Model(&model.Host{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&hosts).Error
	return hosts, err
}

func (u *HostRepo) Page(page, size int, opts ...DBOption) (int64, []model.Host, error) {
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

func (c *HostRepo) WithByInfo(info string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(info) == 0 {
			return g
		}
		infoStr := "%" + info + "%"
		return g.Where("name LIKE ? OR addr LIKE ?", infoStr, infoStr)
	}
}

func (u *HostRepo) WithByPort(port uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("port = ?", port)
	}
}
func (u *HostRepo) WithByUser(user string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("user = ?", user)
	}
}
func (u *HostRepo) WithByAddr(addr string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("addr = ?", addr)
	}
}
func (u *HostRepo) WithByGroup(group string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(group) == 0 {
			return g
		}
		return g.Where("group_belong = ?", group)
	}
}

func (u *HostRepo) Create(host *model.Host) error {
	return global.DB.Create(host).Error
}

func (u *HostRepo) ChangeGroup(oldGroup, newGroup string) error {
	return global.DB.Model(&model.Host{}).Where("group_belong = ?", oldGroup).Updates(map[string]interface{}{"group_belong": newGroup}).Error
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
