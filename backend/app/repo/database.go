package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"gorm.io/gorm"
)

type DatabaseRepo struct {
}

func (d DatabaseRepo) ByAppInstallId(installId uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("app_install_id = ?", installId)
	}
}

func (d DatabaseRepo) Create(ctx context.Context, database *model.AppDatabase) error {
	return getTx(ctx).Model(&model.AppDatabase{}).Create(&database).Error
}

func (d DatabaseRepo) DeleteBy(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Model(&model.AppDatabase{}).Delete(&model.AppDatabase{}).Error
}

func (d DatabaseRepo) GetBy(opts ...DBOption) ([]model.AppDatabase, error) {
	db := getDb(opts...)
	var databases []model.AppDatabase
	err := db.Find(&databases).Error
	if err != nil {
		return nil, err
	}
	return databases, nil
}

func (d DatabaseRepo) GetFirst(opts ...DBOption) (model.AppDatabase, error) {
	db := getDb(opts...)
	var database model.AppDatabase
	err := db.Find(&database).Error
	if err != nil {
		return database, err
	}
	return database, nil
}
