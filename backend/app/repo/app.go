package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AppRepo struct {
}

func (a AppRepo) WithKey(key string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("key = ?", key)
	}
}

func (a AppRepo) WithType(typeStr string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("type = ?", typeStr)
	}
}

func (a AppRepo) Page(page, size int, opts ...DBOption) (int64, []model.App, error) {
	var apps []model.App
	db := getDb(opts...).Model(&model.App{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Debug().Limit(size).Offset(size * (page - 1)).Preload("AppTags").Find(&apps).Error
	return count, apps, err
}

func (a AppRepo) GetFirst(opts ...DBOption) (model.App, error) {
	var app model.App
	db := getDb(opts...).Model(&model.App{})
	if err := db.Preload("AppTags").First(&app).Error; err != nil {
		return app, err
	}
	return app, nil
}

func (a AppRepo) GetBy(opts ...DBOption) ([]model.App, error) {
	var apps []model.App
	db := getDb(opts...).Model(&model.App{})
	if err := db.Preload("Details").Preload("AppTags").Find(&apps).Error; err != nil {
		return apps, err
	}
	return apps, nil
}

func (a AppRepo) BatchCreate(ctx context.Context, apps []model.App) error {
	return getTx(ctx).Omit(clause.Associations).Create(&apps).Error
}

func (a AppRepo) GetByKey(ctx context.Context, key string) (model.App, error) {
	var app model.App
	if err := getTx(ctx).Where("key = ?", key).First(&app).Error; err != nil {
		return app, err
	}
	return app, nil
}

func (a AppRepo) Create(ctx context.Context, app *model.App) error {
	return getTx(ctx).Omit(clause.Associations).Create(app).Error
}

func (a AppRepo) Save(ctx context.Context, app *model.App) error {
	return getTx(ctx).Omit(clause.Associations).Save(app).Error
}
