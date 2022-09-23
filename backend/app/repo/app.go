package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/global"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AppRepo struct {
}

func (a AppRepo) Page(page, size int, opts ...DBOption) (int64, []model.App, error) {
	var apps []model.App
	db := global.DB.Model(&model.App{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Debug().Limit(size).Offset(size * (page - 1)).Preload("AppTags").Find(&apps).Error
	return count, apps, err
}

func (a AppRepo) GetFirst(opts ...DBOption) (model.App, error) {
	var app model.App
	db := global.DB.Model(&model.App{})
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.First(&app).Error; err != nil {
		return app, err
	}
	return app, nil
}

func (a AppRepo) BatchCreate(ctx context.Context, apps []*model.App) error {
	db := ctx.Value("db").(*gorm.DB)
	return db.Omit(clause.Associations).Create(apps).Error
}

func (a AppRepo) GetByKey(ctx context.Context, key string) (model.App, error) {
	db := ctx.Value("db").(*gorm.DB)
	var app model.App
	if err := db.Where("key = ?", key).First(&app).Error; err != nil {
		return app, err
	}
	return app, nil
}

func (a AppRepo) Create(ctx context.Context, app *model.App) error {
	db := ctx.Value("db").(*gorm.DB)
	return db.Omit(clause.Associations).Create(app).Error
}

func (a AppRepo) Save(ctx context.Context, app *model.App) error {
	db := ctx.Value("db").(*gorm.DB)
	return db.Omit(clause.Associations).Save(app).Error
}

func (a AppRepo) UpdateAppConfig(ctx context.Context, app *model.AppConfig) error {
	db := ctx.Value("db").(*gorm.DB)
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
	}).Create(app).Error
}
