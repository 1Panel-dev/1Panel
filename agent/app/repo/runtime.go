package repo

import (
	"context"
	"fmt"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

type RuntimeRepo struct {
}

type IRuntimeRepo interface {
	WithImage(image string) DBOption
	WithNotId(id uint) DBOption
	WithStatus(status string) DBOption
	WithDetailId(id uint) DBOption
	WithDetailIdsIn(ids []uint) DBOption
	WithPort(port int) DBOption
	WithNormalStatus(status string) DBOption
	Page(page, size int, opts ...DBOption) (int64, []model.Runtime, error)
	Create(ctx context.Context, runtime *model.Runtime) error
	Save(runtime *model.Runtime) error
	DeleteBy(opts ...DBOption) error
	GetFirst(ctx context.Context, opts ...DBOption) (*model.Runtime, error)
	List(opts ...DBOption) ([]model.Runtime, error)
}

func NewIRunTimeRepo() IRuntimeRepo {
	return &RuntimeRepo{}
}

func (r *RuntimeRepo) WithStatus(status string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("status = ?", status)
	}
}

func (r *RuntimeRepo) WithNormalStatus(status string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("status = ? or status = 'Normal'", status)
	}
}

func (r *RuntimeRepo) WithImage(image string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("image = ?", image)
	}
}

func (r *RuntimeRepo) WithDetailId(id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("app_detail_id = ?", id)
	}
}

func (r *RuntimeRepo) WithDetailIdsIn(ids []uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("app_detail_id in(?) ", ids)
	}
}

func (r *RuntimeRepo) WithNotId(id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id != ?", id)
	}
}

func (r *RuntimeRepo) WithPort(port int) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		portStr := fmt.Sprintf("%d", port)
		return g.Debug().Where(
			"port = ? OR port LIKE ? OR port LIKE ? OR port LIKE ?",
			portStr,
			portStr+",%",
			"%,"+portStr,
			"%,"+portStr+",%",
		)
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

func (r *RuntimeRepo) List(opts ...DBOption) ([]model.Runtime, error) {
	var runtimes []model.Runtime
	db := global.DB.Model(&model.Runtime{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&runtimes).Error
	return runtimes, err
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

func (r *RuntimeRepo) GetFirst(ctx context.Context, opts ...DBOption) (*model.Runtime, error) {
	var runtime model.Runtime
	if err := getTx(ctx, opts...).First(&runtime).Error; err != nil {
		return nil, err
	}
	return &runtime, nil
}
