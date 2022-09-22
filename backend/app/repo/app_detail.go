package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/app/model"
	"gorm.io/gorm"
)

type AppDetailRepo struct {
}

func (a AppDetailRepo) BatchCreate(ctx context.Context, details []*model.AppDetail) error {
	db := ctx.Value("db").(*gorm.DB)
	return db.Model(&model.AppDetail{}).Create(&details).Error
}

func (a AppDetailRepo) DeleteByAppIds(ctx context.Context, appIds []uint) error {
	db := ctx.Value("db").(*gorm.DB)
	return db.Where("app_id in (?)", appIds).Delete(&model.AppDetail{}).Error
}

func (a AppDetailRepo) GetByAppId(ctx context.Context, appId string) ([]model.AppDetail, error) {
	db := ctx.Value("db").(*gorm.DB)
	var details []model.AppDetail
	if err := db.Where("app_id = ?", appId).Find(&details).Error; err != nil {
		return nil, err
	}
	return details, nil
}
