package repo

import (
	"context"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

type BackupRepo struct{}

type IBackupRepo interface {
	Get(opts ...DBOption) (model.BackupAccount, error)
	List(opts ...DBOption) ([]model.BackupAccount, error)
	Create(backup []model.BackupAccount) error
	Save(backup *model.BackupAccount) error
	Delete(opts ...DBOption) error

	ListRecord(opts ...DBOption) ([]model.BackupRecord, error)
	PageRecord(page, size int, opts ...DBOption) (int64, []model.BackupRecord, error)
	CreateRecord(record *model.BackupRecord) error
	DeleteRecord(ctx context.Context, opts ...DBOption) error
	UpdateRecord(record *model.BackupRecord) error
	WithByDetailName(detailName string) DBOption
	WithByFileName(fileName string) DBOption
	WithByType(backupType string) DBOption
	WithByCronID(cronjobID uint) DBOption
}

func NewIBackupRepo() IBackupRepo {
	return &BackupRepo{}
}

func (u *BackupRepo) Get(opts ...DBOption) (model.BackupAccount, error) {
	var backup model.BackupAccount
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&backup).Error
	return backup, err
}

func (u *BackupRepo) Save(backup *model.BackupAccount) error {
	return global.DB.Save(backup).Error
}

func (u *BackupRepo) ListRecord(opts ...DBOption) ([]model.BackupRecord, error) {
	var users []model.BackupRecord
	db := global.DB.Model(&model.BackupRecord{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&users).Error
	return users, err
}

func (u *BackupRepo) PageRecord(page, size int, opts ...DBOption) (int64, []model.BackupRecord, error) {
	var users []model.BackupRecord
	db := global.DB.Model(&model.BackupRecord{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&users).Error
	return count, users, err
}

func (u *BackupRepo) WithByDetailName(detailName string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(detailName) == 0 {
			return g
		}
		return g.Where("detail_name = ?", detailName)
	}
}

func (u *BackupRepo) WithByFileName(fileName string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(fileName) == 0 {
			return g
		}
		return g.Where("file_name = ?", fileName)
	}
}

func (u *BackupRepo) WithByType(backupType string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(backupType) == 0 {
			return g
		}
		return g.Where("type = ?", backupType)
	}
}

func (u *BackupRepo) List(opts ...DBOption) ([]model.BackupAccount, error) {
	var ops []model.BackupAccount
	db := global.DB.Model(&model.BackupAccount{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&ops).Error
	return ops, err
}

func (u *BackupRepo) Create(backup []model.BackupAccount) error {
	return global.DB.Create(backup).Error
}

func (u *BackupRepo) CreateRecord(record *model.BackupRecord) error {
	return global.DB.Create(record).Error
}

func (u *BackupRepo) UpdateRecord(record *model.BackupRecord) error {
	return global.DB.Save(record).Error
}

func (u *BackupRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.BackupAccount{}).Error
}

func (u *BackupRepo) DeleteRecord(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.BackupRecord{}).Error
}

func (u *BackupRepo) WithByCronID(cronjobID uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("cronjob_id = ?", cronjobID)
	}
}
