package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
)

type AppTagRepo struct {
}

func (a AppTagRepo) BatchCreate(ctx context.Context, tags []*model.AppTag) error {
	db := ctx.Value("db").(*gorm.DB)
	return db.Create(&tags).Error
}

func (a AppTagRepo) DeleteByAppIds(ctx context.Context, appIds []uint) error {
	db := ctx.Value("db").(*gorm.DB)
	return db.Where("app_id in (?)", appIds).Delete(&model.AppTag{}).Error
}

func (a AppTagRepo) DeleteAll(ctx context.Context) error {
	db := ctx.Value("db").(*gorm.DB)
	return db.Where("1 = 1").Delete(&model.AppTag{}).Error
}

func (a AppTagRepo) GetByAppId(appId uint) ([]model.AppTag, error) {
	var appTags []model.AppTag
	if err := global.DB.Where("app_id = ?", appId).Find(&appTags).Error; err != nil {
		return nil, err
	}
	return appTags, nil
}

func (a AppTagRepo) GetByTagIds(tagIds []uint) ([]model.AppTag, error) {
	var appTags []model.AppTag
	if err := global.DB.Where("tag_id in (?)", tagIds).Find(&appTags).Error; err != nil {
		return nil, err
	}
	return appTags, nil
}
