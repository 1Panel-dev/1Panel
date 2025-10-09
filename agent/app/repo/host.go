package repo

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
)

type HostRepo struct{}

type IHostRepo interface {
	GetFirewallRecord(opts ...DBOption) (model.Firewall, error)
	ListFirewallRecord() ([]model.Firewall, error)
	SaveFirewallRecord(firewall *model.Firewall) error
	DeleteFirewallRecordByID(id uint) error
	DeleteFirewallRecord(fType, port, protocol, address, strategy string) error

	SyncCert(data []model.RootCert) error
	GetCert(opts ...DBOption) (model.RootCert, error)
	PageCert(limit, offset int, opts ...DBOption) (int64, []model.RootCert, error)
	ListCert(opts ...DBOption) ([]model.RootCert, error)
	SaveCert(cert *model.RootCert) error
	UpdateCert(id uint, vars map[string]interface{}) error
	DeleteCert(opts ...DBOption) error
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

func (u *HostRepo) GetCert(opts ...DBOption) (model.RootCert, error) {
	var cert model.RootCert
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&cert).Error
	return cert, err
}

func (u *HostRepo) PageCert(page, size int, opts ...DBOption) (int64, []model.RootCert, error) {
	var ops []model.RootCert
	db := global.DB.Model(&model.RootCert{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&ops).Error
	return count, ops, err
}

func (u *HostRepo) ListCert(opts ...DBOption) ([]model.RootCert, error) {
	var ops []model.RootCert
	db := global.DB.Model(&model.RootCert{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Find(&ops).Error
	return ops, err
}

func (u *HostRepo) SaveCert(cert *model.RootCert) error {
	return global.DB.Save(cert).Error
}

func (u *HostRepo) UpdateCert(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.RootCert{}).Where("id = ?", id).Updates(vars).Error
}

func (u *HostRepo) DeleteCert(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.RootCert{}).Error
}

func (u *HostRepo) SyncCert(data []model.RootCert) error {
	tx := global.DB.Begin()
	var oldCerts []model.RootCert
	_ = tx.Where("1 = ?", 1).Find(&oldCerts).Error
	oldCertsMap := make(map[string]uint)
	for _, item := range oldCerts {
		oldCertsMap[item.Name] = item.ID
	}
	for _, item := range data {
		if _, ok := oldCertsMap[item.Name]; ok {
			delete(oldCertsMap, item.Name)
			continue
		}
		item.PassPhrase, _ = encrypt.StringEncrypt("<UN-SET>")
		if err := tx.Model(model.RootCert{}).Create(&item).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	for _, val := range oldCertsMap {
		if err := tx.Where("id = ?", val).Delete(&model.RootCert{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}
