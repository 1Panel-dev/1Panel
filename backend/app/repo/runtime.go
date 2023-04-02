package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"gorm.io/gorm"
)

type RuntimeRepo struct {
}

type IRuntimeRepo interface {
	WithNameOrImage(name string, image string) DBOption
	WithOtherNameOrImage(name string, image string, id uint) DBOption
	Page(page, size int, opts ...DBOption) (int64, []model.Runtime, error)
	Create(ctx context.Context, runtime *model.Runtime) error
	Save(runtime *model.Runtime) error
	DeleteBy(opts ...DBOption) error
	GetFirst(opts ...DBOption) (*model.Runtime, error)
}

func NewIRunTimeRepo() IRuntimeRepo {
	return &RuntimeRepo{}
}

func (r *RuntimeRepo) WithNameOrImage(name string, image string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("name = ? or image = ?", name, image)
	}
}

func (r *RuntimeRepo) WithOtherNameOrImage(name string, image string, id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("name = ? or image = ? and id != ?", name, image, id)
	}
}

func (r *RuntimeRepo) Page(page, size int, opts ...DBOption) (int64, []model.Runtime, error) {
	var runtimes []model.Runtime
	db := getDb(opts...).Model(&model.Runtime{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&runtimes).Error
	return count, runtimes, err
}

func (r *RuntimeRepo) Create(ctx context.Context, runtime *model.Runtime) error {
	db := getTx(ctx).Model(&model.Runtime{})
	return db.Create(&runtime).Error
}

func (r *RuntimeRepo) Save(runtime *model.Runtime) error {
	return getDb().Save(&runtime).Error
}

func (r *RuntimeRepo) DeleteBy(opts ...DBOption) error {
	return getDb(opts...).Delete(&model.Runtime{}).Error
}

func (r *RuntimeRepo) GetFirst(opts ...DBOption) (*model.Runtime, error) {
	var runtime model.Runtime
	if err := getDb(opts...).First(&runtime).Error; err != nil {
		return nil, err
	}
	return &runtime, nil
}
