package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/app/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AppInstallBackupRepo struct {
}

func (a AppInstallBackupRepo) WithAppInstallID(appInstallID uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("app_install_id = ?", appInstallID)
	}
}

func (a AppInstallBackupRepo) Create(ctx context.Context, backup model.AppInstallBackup) error {
	return getTx(ctx).Omit(clause.Associations).Create(&backup).Error
}

func (a AppInstallBackupRepo) Delete(opts ...DBOption) error {
	return getDb(opts...).Omit(clause.Associations).Delete(&model.AppInstallBackup{}).Error
}
func (a AppInstallBackupRepo) GetBy(opts ...DBOption) ([]model.AppInstallBackup, error) {
	var backups []model.AppInstallBackup
	if err := getDb(opts...).Find(&backups); err != nil {
		return backups, nil
	}
	return backups, nil
}

func (a AppInstallBackupRepo) GetFirst(opts ...DBOption) (model.AppInstallBackup, error) {
	var backup model.AppInstallBackup
	db := getDb(opts...).Model(&model.AppInstallBackup{})
	err := db.First(&backup).Error
	return backup, err
}

func (a AppInstallBackupRepo) Page(page, size int, opts ...DBOption) (int64, []model.AppInstallBackup, error) {
	var backups []model.AppInstallBackup
	db := getDb(opts...).Model(&model.AppInstallBackup{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Debug().Limit(size).Offset(size * (page - 1)).Find(&backups).Error
	return count, backups, err
}
