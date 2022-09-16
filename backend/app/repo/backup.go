package repo

import (
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/global"
)

type BackupRepo struct{}

type IBackupRepo interface {
	Page(page, size int, opts ...DBOption) (int64, []model.BackupAccount, error)
	Create(backup *model.BackupAccount) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error
}

func NewIBackupService() IBackupRepo {
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

func (u *BackupRepo) Page(page, size int, opts ...DBOption) (int64, []model.BackupAccount, error) {
	var ops []model.BackupAccount
	db := global.DB.Model(&model.BackupAccount{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&ops).Error
	return count, ops, err
}

func (u *BackupRepo) Create(backup *model.BackupAccount) error {
	return global.DB.Create(backup).Error
}

func (u *BackupRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.BackupAccount{}).Where("id = ?", id).Updates(vars).Error
}

func (u *BackupRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.BackupAccount{}).Error
}
