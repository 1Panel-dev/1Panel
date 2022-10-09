package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/global"
	"gorm.io/gorm"
)

type TagRepo struct {
}

func (t TagRepo) BatchCreate(ctx context.Context, tags []*model.Tag) error {
	db := ctx.Value("db").(*gorm.DB)
	return db.Create(&tags).Error
}

func (t TagRepo) DeleteAll(ctx context.Context) error {
	db := ctx.Value("db").(*gorm.DB)
	return db.Where("1 = 1 ").Delete(&model.Tag{}).Error
}

func (t TagRepo) All() ([]model.Tag, error) {
	var tags []model.Tag
	if err := global.DB.Where("1 = 1 ").Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (t TagRepo) GetByIds(ids []uint) ([]model.Tag, error) {
	var tags []model.Tag
	if err := global.DB.Where("id in (?)", ids).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (t TagRepo) GetByKeys(keys []string) ([]model.Tag, error) {
	var tags []model.Tag
	if err := global.DB.Where("key in (?)", keys).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (t TagRepo) GetByAppId(appId uint) ([]model.Tag, error) {
	var tags []model.Tag
	if err := global.DB.Where("id in (select tag_id from app_tags where app_id = ?)", appId).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}
