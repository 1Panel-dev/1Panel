package repo

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"gorm.io/gorm"
)

type HostRepo struct{}

type IHostRepo interface {
	Get(opts ...DBOption) (model.Host, error)
	GetList(opts ...DBOption) ([]model.Host, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.Host, error)
	Create(host *model.Host) error
	Update(id uint, vars map[string]interface{}) error
	UpdateGroup(group, newGroup uint) error
	Delete(opts ...DBOption) error

	WithByInfo(info string) DBOption
	WithByPort(port uint) DBOption
	WithByUser(user string) DBOption

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

func (h *HostRepo) Create(host *model.Host) error {
	return global.DB.Create(host).Error
}

func (h *HostRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.Host{}).Where("id = ?", id).Updates(vars).Error
}

func (h *HostRepo) UpdateGroup(group, newGroup uint) error {
	return global.DB.Model(&model.Host{}).Where("group_id = ?", group).Updates(map[string]interface{}{"group_id": newGroup}).Error
}

func (h *HostRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.Host{}).Error
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
