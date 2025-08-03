package repo

import (
	"context"

	"github.com/1Panel-dev/1Panel/agent/app/model"
)

type TagRepo struct {
}

type ITagRepo interface {
	BatchCreate(ctx context.Context, tags []*model.Tag) error
	DeleteAll(ctx context.Context) error
	All() ([]model.Tag, error)
	GetByIds(ids []uint) ([]model.Tag, error)
	GetByKeys(keys []string) ([]model.Tag, error)
	GetByAppId(appId uint) ([]model.Tag, error)
	DeleteByID(ctx context.Context, id uint) error
	Create(ctx context.Context, tag *model.Tag) error
	Save(ctx context.Context, tag *model.Tag) error
	GetByKey(key string) (*model.Tag, error)
}

func NewITagRepo() ITagRepo {
	return &TagRepo{}
}

func (t TagRepo) BatchCreate(ctx context.Context, tags []*model.Tag) error {
	return getTx(ctx).Create(&tags).Error
}

func (t TagRepo) DeleteAll(ctx context.Context) error {
	return getTx(ctx).Where("1 = 1 ").Delete(&model.Tag{}).Error
}

func (t TagRepo) All() ([]model.Tag, error) {
	var tags []model.Tag
	if err := getDb().Where("1 = 1 ").Order("sort asc").Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (t TagRepo) GetByKey(key string) (*model.Tag, error) {
	var tag model.Tag
	if err := getDb().Where("key = ?", key).First(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (t TagRepo) GetByIds(ids []uint) ([]model.Tag, error) {
	var tags []model.Tag
	if err := getDb().Where("id in (?)", ids).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (t TagRepo) GetByKeys(keys []string) ([]model.Tag, error) {
	var tags []model.Tag
	if err := getDb().Where("key in (?)", keys).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (t TagRepo) GetByAppId(appId uint) ([]model.Tag, error) {
	var tags []model.Tag
	if err := getDb().Where("id in (select tag_id from app_tags where app_id = ?)", appId).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (t TagRepo) DeleteByID(ctx context.Context, id uint) error {
	return getTx(ctx).Where("id = ?", id).Delete(&model.Tag{}).Error
}

func (t TagRepo) Create(ctx context.Context, tag *model.Tag) error {
	return getTx(ctx).Create(tag).Error
}

func (t TagRepo) Save(ctx context.Context, tag *model.Tag) error {
	return getTx(ctx).Save(tag).Error
}
