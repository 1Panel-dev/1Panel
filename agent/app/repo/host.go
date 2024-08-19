package repo

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
)

type HostRepo struct{}

type IHostRepo interface {
	GetFirewallRecord(opts ...DBOption) (model.Firewall, error)
	ListFirewallRecord() ([]model.Firewall, error)
	SaveFirewallRecord(firewall *model.Firewall) error
	DeleteFirewallRecordByID(id uint) error
	DeleteFirewallRecord(fType, port, protocol, address, strategy string) error
}

func NewIHostRepo() IHostRepo {
	return &HostRepo{}
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
