package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/backend/app/model"
)

type AppTagRepo struct {
}

type IAppTagRepo interface {
	BatchCreate(ctx context.Context, tags []*model.AppTag) error
	DeleteByAppIds(ctx context.Context, appIds []uint) error
	DeleteAll(ctx context.Context) error
	GetByAppId(appId uint) ([]model.AppTag, error)
	GetByTagIds(tagIds []uint) ([]model.AppTag, error)
}

func NewIAppTagRepo() IAppTagRepo {
	return &AppTagRepo{}
}

func (a AppTagRepo) BatchCreate(ctx context.Context, tags []*model.AppTag) error {
	return getTx(ctx).Create(&tags).Error
}

func (a AppTagRepo) DeleteByAppIds(ctx context.Context, appIds []uint) error {
	return getTx(ctx).Where("app_id in (?)", appIds).Delete(&model.AppTag{}).Error
}

func (a AppTagRepo) DeleteAll(ctx context.Context) error {
	return getTx(ctx).Where("1 = 1").Delete(&model.AppTag{}).Error
}

func (a AppTagRepo) GetByAppId(appId uint) ([]model.AppTag, error) {
	var appTags []model.AppTag
	if err := getDb().Where("app_id = ?", appId).Find(&appTags).Error; err != nil {
		return nil, err
	}
	return appTags, nil
}

func (a AppTagRepo) GetByTagIds(tagIds []uint) ([]model.AppTag, error) {
	var appTags []model.AppTag
	if err := getDb().Where("tag_id in (?)", tagIds).Find(&appTags).Error; err != nil {
		return nil, err
	}
	return appTags, nil
}
