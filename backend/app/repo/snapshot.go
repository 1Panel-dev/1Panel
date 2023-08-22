package repo

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
)

type ISnapshotRepo interface {
	Get(opts ...DBOption) (model.Snapshot, error)
	GetList(opts ...DBOption) ([]model.Snapshot, error)
	Create(snap *model.Snapshot) error
	Update(id uint, vars map[string]interface{}) error
	Page(limit, offset int, opts ...DBOption) (int64, []model.Snapshot, error)
	Delete(opts ...DBOption) error

	GetStatus(snapID uint) (model.SnapshotStatus, error)
	GetStatusList(opts ...DBOption) ([]model.SnapshotStatus, error)
	CreateStatus(snap *model.SnapshotStatus) error
	DeleteStatus(snapID uint) error
	UpdateStatus(id uint, vars map[string]interface{}) error
}

func NewISnapshotRepo() ISnapshotRepo {
	return &SnapshotRepo{}
}

type SnapshotRepo struct{}

func (u *SnapshotRepo) Get(opts ...DBOption) (model.Snapshot, error) {
	var Snapshot model.Snapshot
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&Snapshot).Error
	return Snapshot, err
}

func (u *SnapshotRepo) GetList(opts ...DBOption) ([]model.Snapshot, error) {
	var snaps []model.Snapshot
	db := global.DB.Model(&model.Snapshot{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&snaps).Error
	return snaps, err
}

func (u *SnapshotRepo) Page(page, size int, opts ...DBOption) (int64, []model.Snapshot, error) {
	var users []model.Snapshot
	db := global.DB.Model(&model.Snapshot{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&users).Error
	return count, users, err
}

func (u *SnapshotRepo) Create(Snapshot *model.Snapshot) error {
	return global.DB.Create(Snapshot).Error
}

func (u *SnapshotRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.Snapshot{}).Where("id = ?", id).Updates(vars).Error
}

func (u *SnapshotRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.Snapshot{}).Error
}

func (u *SnapshotRepo) GetStatus(snapID uint) (model.SnapshotStatus, error) {
	var data model.SnapshotStatus
	if err := global.DB.Where("snap_id = ?", snapID).First(&data).Error; err != nil {
		return data, err
	}
	return data, nil
}

func (u *SnapshotRepo) GetStatusList(opts ...DBOption) ([]model.SnapshotStatus, error) {
	var status []model.SnapshotStatus
	db := global.DB.Model(&model.SnapshotStatus{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&status).Error
	return status, err
}

func (u *SnapshotRepo) CreateStatus(snap *model.SnapshotStatus) error {
	return global.DB.Create(snap).Error
}

func (u *SnapshotRepo) DeleteStatus(snapID uint) error {
	return global.DB.Where("snap_id = ?", snapID).Delete(&model.SnapshotStatus{}).Error
}

func (u *SnapshotRepo) UpdateStatus(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.SnapshotStatus{}).Where("id = ?", id).Updates(vars).Error
}
