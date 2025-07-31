package repo

import (
	"context"
	"gorm.io/gorm"

	"github.com/1Panel-dev/1Panel/agent/app/model"
)

type AppTagRepo struct {
}

type IAppTagRepo interface {
	BatchCreate(ctx context.Context, tags []*model.AppTag) error
	DeleteByAppIds(ctx context.Context, appIds []uint) error
	DeleteAll(ctx context.Context) error
	GetByAppId(appId uint) ([]model.AppTag, error)
	GetByTagIds(tagIds []uint) ([]model.AppTag, error)
	DeleteBy(ctx context.Context, opts ...DBOption) error
	GetFirst(opts ...DBOption) (*model.AppTag, error)

	WithByTagID(tagID uint) DBOption
	WithByAppID(appId uint) DBOption
}

func NewIAppTagRepo() IAppTagRepo {
	return &AppTagRepo{}
}

func (a AppTagRepo) WithByTagID(tagID uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("tag_id = ?", tagID)
	}
}

func (a AppTagRepo) WithByAppID(appId uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("app_id = ?", appId)
	}
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

func (a AppTagRepo) DeleteBy(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.AppTag{}).Error
}

func (a AppTagRepo) GetFirst(opts ...DBOption) (*model.AppTag, error) {
	var appTag model.AppTag
	if err := getDb(opts...).First(&appTag).Error; err != nil {
		return nil, err
	}
	return &appTag, nil

}
