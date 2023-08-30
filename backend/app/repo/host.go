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
	Create(host *model.Host) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error

	GetFirewallRecord(opts ...DBOption) (model.Firewall, error)
	ListFirewallRecord() ([]model.Firewall, error)
	SaveFirewallRecord(firewall *model.Firewall) error
	DeleteFirewallRecordByID(id uint) error
	DeleteFirewallRecord(fType, port, protocol, address, strategy string) error
}

func NewIHostRepo() IHostRepo {
	return &HostRepo{}
}

func (h *HostRepo) Get(opts ...DBOption) (model.Host, error) {
	var host model.Host
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&host).Error
	return host, err
}

func (h *HostRepo) GetList(opts ...DBOption) ([]model.Host, error) {
	var hosts []model.Host
	db := global.DB.Model(&model.Host{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&hosts).Error
	return hosts, err
}

func (h *HostRepo) Page(page, size int, opts ...DBOption) (int64, []model.Host, error) {
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

func (h *HostRepo) WithByInfo(info string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(info) == 0 {
			return g
		}
		infoStr := "%" + info + "%"
		return g.Where("name LIKE ? OR addr LIKE ?", infoStr, infoStr)
	}
}

func (h *HostRepo) WithByPort(port uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("port = ?", port)
	}
}
func (h *HostRepo) WithByUser(user string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("user = ?", user)
	}
}
func (h *HostRepo) WithByAddr(addr string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("addr = ?", addr)
	}
}
func (h *HostRepo) WithByGroup(group string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(group) == 0 {
			return g
		}
		return g.Where("group_belong = ?", group)
	}
}

func (h *HostRepo) Create(host *model.Host) error {
	return global.DB.Create(host).Error
}

func (h *HostRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.Host{}).Where("id = ?", id).Updates(vars).Error
}

func (h *HostRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.Host{}).Error
}

func (h *HostRepo) GetFirewallRecord(opts ...DBOption) (model.Firewall, error) {
	var firewall model.Firewall
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&firewall).Error
	return firewall, err
}

func (h *HostRepo) ListFirewallRecord() ([]model.Firewall, error) {
	var datas []model.Firewall
	if err := global.DB.Find(&datas).Error; err != nil {
		return datas, nil
	}
	return datas, nil
}

func (h *HostRepo) SaveFirewallRecord(firewall *model.Firewall) error {
	if firewall.ID != 0 {
		return global.DB.Save(firewall).Error
	}
	var data model.Firewall
	if firewall.Type == "port" {
		_ = global.DB.Where("type = ? AND port = ? AND protocol = ? AND address = ? AND strategy = ?", "port", firewall.Port, firewall.Protocol, firewall.Address, firewall.Strategy).First(&data)
		if data.ID != 0 {
			firewall.ID = data.ID
		}
	} else {
		_ = global.DB.Where("type = ? AND address = ? AND strategy = ?", "address", firewall.Address, firewall.Strategy).First(&data)
		if data.ID != 0 {
			firewall.ID = data.ID
		}
	}
	return global.DB.Save(firewall).Error
}

func (h *HostRepo) DeleteFirewallRecordByID(id uint) error {
	return global.DB.Where("id = ?", id).Delete(&model.Firewall{}).Error
}

func (h *HostRepo) DeleteFirewallRecord(fType, port, protocol, address, strategy string) error {
	return global.DB.Where("type = ? AND port = ? AND protocol = ? AND address = ? AND strategy = ?", fType, port, protocol, address, strategy).Delete(&model.Firewall{}).Error
}
