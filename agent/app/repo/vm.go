package repo

import (
	"encoding/json"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

type VMRepo struct{}

type IVMRepo interface {
	Get(opts ...DBOption) (model.VirtualMachine, error)
	List() ([]model.VirtualMachine, error)
	Count() (int64, error)
	Page(page, pageSize int, opts ...DBOption) (int64, []model.VirtualMachine, error)
	Create(vm *model.VirtualMachine) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error
	WithByNetwork(network string) DBOption
	WithByStorage(storage string) DBOption

	GetTemplate(opts ...DBOption) (model.VMTemplate, error)
	ListTemplates(opts ...DBOption) ([]model.VMTemplate, error)
	CreateTemplate(item *model.VMTemplate) error
	UpdateTemplateSource(oldName string, newName string) error
	DeleteTemplate(opts ...DBOption) error

	GetIso(opts ...DBOption) (model.VMIso, error)
	PageIso(page, pageSize int, opts ...DBOption) (int64, []model.VMIso, error)
	ListIsos(opts ...DBOption) ([]model.VMIso, error)
	CreateIso(iso *model.VMIso) error
	UpdateIso(id uint, vars map[string]interface{}) error
	DeleteIso(opts ...DBOption) error

	GetNetwork(opts ...DBOption) (model.VMNetwork, error)
	PageNetwork(page, pageSize int, opts ...DBOption) (int64, []model.VMNetwork, error)
	ListNetworks(opts ...DBOption) ([]model.VMNetwork, error)
	CreateNetwork(network *model.VMNetwork) error
	UpdateNetwork(id uint, vars map[string]interface{}) error
	DeleteNetwork(opts ...DBOption) error

	GetStorage(opts ...DBOption) (model.VMStorage, error)
	PageStorage(page, pageSize int, opts ...DBOption) (int64, []model.VMStorage, error)
	ListStorages(opts ...DBOption) ([]model.VMStorage, error)
	CreateStorage(storage *model.VMStorage) error
	UpdateStorage(id uint, vars map[string]interface{}) error
	DeleteStorage(opts ...DBOption) error
}

func NewIVMRepo() IVMRepo {
	return &VMRepo{}
}

func (v *VMRepo) Get(opts ...DBOption) (model.VirtualMachine, error) {
	var vm model.VirtualMachine
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&vm).Error
	return vm, err
}

func (v *VMRepo) List() ([]model.VirtualMachine, error) {
	var vms []model.VirtualMachine
	err := global.DB.Order("created_at desc").Find(&vms).Error
	return vms, err
}

func (v *VMRepo) Count() (int64, error) {
	var count int64
	err := global.DB.Model(&model.VirtualMachine{}).Count(&count).Error
	return count, err
}

func (v *VMRepo) Page(page, pageSize int, opts ...DBOption) (int64, []model.VirtualMachine, error) {
	var vms []model.VirtualMachine
	var total int64
	db := global.DB.Model(&model.VirtualMachine{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&total).Error
	if err != nil {
		return 0, nil, err
	}
	err = db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&vms).Error
	return total, vms, err
}

func (v *VMRepo) Create(vm *model.VirtualMachine) error {
	return global.DB.Create(vm).Error
}

func (v *VMRepo) Update(id uint, vars map[string]interface{}) error {
	if paths, ok := vars["iso_paths"].([]string); ok {
		data, err := json.Marshal(paths)
		if err != nil {
			return err
		}
		vars["iso_paths"] = string(data)
	}
	return global.DB.Model(&model.VirtualMachine{}).Where("id = ?", id).Updates(vars).Error
}

func (v *VMRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.VirtualMachine{}).Error
}

func (v *VMRepo) WithByNetwork(network string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("network = ?", network)
	}
}

func (v *VMRepo) WithByStorage(storage string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("storage_pool = ?", storage)
	}
}

func (v *VMRepo) GetTemplate(opts ...DBOption) (model.VMTemplate, error) {
	var item model.VMTemplate
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&item).Error
	return item, err
}

func (v *VMRepo) ListTemplates(opts ...DBOption) ([]model.VMTemplate, error) {
	var items []model.VMTemplate
	db := global.DB.Model(&model.VMTemplate{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Order("created_at desc").Order("id desc").Find(&items).Error
	return items, err
}

func (v *VMRepo) CreateTemplate(item *model.VMTemplate) error {
	return global.DB.Create(item).Error
}

func (v *VMRepo) UpdateTemplateSource(oldName string, newName string) error {
	return global.DB.Model(&model.VMTemplate{}).Where("source_vm = ?", oldName).Update("source_vm", newName).Error
}

func (v *VMRepo) DeleteTemplate(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.VMTemplate{}).Error
}

func (v *VMRepo) GetIso(opts ...DBOption) (model.VMIso, error) {
	var iso model.VMIso
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&iso).Error
	return iso, err
}

func (v *VMRepo) PageIso(page, pageSize int, opts ...DBOption) (int64, []model.VMIso, error) {
	var isos []model.VMIso
	var total int64
	db := global.DB.Model(&model.VMIso{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&total).Error
	if err != nil {
		return 0, nil, err
	}
	err = db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&isos).Error
	return total, isos, err
}

func (v *VMRepo) ListIsos(opts ...DBOption) ([]model.VMIso, error) {
	var isos []model.VMIso
	db := global.DB.Model(&model.VMIso{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Order("created_at desc").Order("id desc").Find(&isos).Error
	return isos, err
}

func (v *VMRepo) CreateIso(iso *model.VMIso) error {
	return global.DB.Create(iso).Error
}

func (v *VMRepo) UpdateIso(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.VMIso{}).Where("id = ?", id).Updates(vars).Error
}

func (v *VMRepo) DeleteIso(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.VMIso{}).Error
}

func (v *VMRepo) GetNetwork(opts ...DBOption) (model.VMNetwork, error) {
	var network model.VMNetwork
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&network).Error
	return network, err
}

func (v *VMRepo) PageNetwork(page, pageSize int, opts ...DBOption) (int64, []model.VMNetwork, error) {
	var networks []model.VMNetwork
	var total int64
	db := global.DB.Model(&model.VMNetwork{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&total).Error
	if err != nil {
		return 0, nil, err
	}
	err = db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&networks).Error
	return total, networks, err
}

func (v *VMRepo) ListNetworks(opts ...DBOption) ([]model.VMNetwork, error) {
	var networks []model.VMNetwork
	db := global.DB.Model(&model.VMNetwork{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Order("created_at desc").Order("id desc").Find(&networks).Error
	return networks, err
}

func (v *VMRepo) CreateNetwork(network *model.VMNetwork) error {
	return global.DB.Create(network).Error
}

func (v *VMRepo) UpdateNetwork(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.VMNetwork{}).Where("id = ?", id).Updates(vars).Error
}

func (v *VMRepo) DeleteNetwork(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.VMNetwork{}).Error
}

func (v *VMRepo) GetStorage(opts ...DBOption) (model.VMStorage, error) {
	var storage model.VMStorage
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&storage).Error
	return storage, err
}

func (v *VMRepo) PageStorage(page, pageSize int, opts ...DBOption) (int64, []model.VMStorage, error) {
	var storages []model.VMStorage
	var total int64
	db := global.DB.Model(&model.VMStorage{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&total).Error
	if err != nil {
		return 0, nil, err
	}
	err = db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&storages).Error
	return total, storages, err
}

func (v *VMRepo) ListStorages(opts ...DBOption) ([]model.VMStorage, error) {
	var storages []model.VMStorage
	db := global.DB.Model(&model.VMStorage{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Order("created_at desc").Order("id desc").Find(&storages).Error
	return storages, err
}

func (v *VMRepo) CreateStorage(storage *model.VMStorage) error {
	return global.DB.Create(storage).Error
}

func (v *VMRepo) UpdateStorage(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.VMStorage{}).Where("id = ?", id).Updates(vars).Error
}

func (v *VMRepo) DeleteStorage(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.VMStorage{}).Error
}
