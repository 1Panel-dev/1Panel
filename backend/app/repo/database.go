package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
)

type DatabaseRepo struct {
}

func (d DatabaseRepo) ByAppInstallId(installId uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("app_install_id = ?", installId)
	}
}

func (d DatabaseRepo) Create(ctx context.Context, database *model.Database) error {
	db := ctx.Value("db").(*gorm.DB).Model(&model.Database{})
	return db.Create(&database).Error
}

func (d DatabaseRepo) DeleteBy(ctx context.Context, opts ...DBOption) error {
	db := ctx.Value("db").(*gorm.DB).Model(&model.Database{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.Database{}).Error
}

func (d DatabaseRepo) GetBy(opts ...DBOption) ([]model.Database, error) {
	db := global.DB.Model(model.Database{})
	var databases []model.Database
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&databases).Error
	if err != nil {
		return nil, err
	}
	return databases, nil
}

func (d DatabaseRepo) GetFirst(opts ...DBOption) (model.Database, error) {
	db := global.DB.Model(model.Database{})
	var database model.Database
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&database).Error
	if err != nil {
		return database, err
	}
	return database, nil
}
