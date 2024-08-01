package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"gorm.io/gorm"
)

type TaskRepo struct {
}

type ITaskRepo interface {
	Create(ctx context.Context, task *model.Task) error
	GetFirst(opts ...DBOption) (model.Task, error)
	Page(page, size int, opts ...DBOption) (int64, []model.Task, error)
	Update(ctx context.Context, task *model.Task) error

	WithByID(id string) DBOption
	WithType(taskType string) DBOption
	WithResourceID(id uint) DBOption
	WithStatus(status string) DBOption
}

func NewITaskRepo() ITaskRepo {
	return &TaskRepo{}
}

func (t TaskRepo) WithByID(id string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("id = ?", id)
	}
}

func (t TaskRepo) WithType(taskType string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("type = ?", taskType)
	}
}

func (t TaskRepo) WithStatus(status string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("status = ?", status)
	}
}

func (t TaskRepo) WithResourceID(id uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("resource_id = ?", id)
	}
}

func (t TaskRepo) Create(ctx context.Context, task *model.Task) error {
	return getTx(ctx).Create(&task).Error
}

func (t TaskRepo) GetFirst(opts ...DBOption) (model.Task, error) {
	var task model.Task
	db := getDb(opts...).Model(&model.Task{})
	if err := db.First(&task).Error; err != nil {
		return task, err
	}
	return task, nil
}

func (t TaskRepo) Page(page, size int, opts ...DBOption) (int64, []model.Task, error) {
	var tasks []model.Task
	db := getDb(opts...).Model(&model.Task{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&tasks).Error
	return count, tasks, err
}

func (t TaskRepo) Update(ctx context.Context, task *model.Task) error {
	return getTx(ctx).Save(&task).Error
}
