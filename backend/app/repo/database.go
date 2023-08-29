package repo

import (
	"context"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
)

type DatabaseRepo struct{}

type IDatabaseRepo interface {
	GetList(opts ...DBOption) ([]model.Database, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.Database, error)
	Create(ctx context.Context, database *model.Database) error
	Update(id uint, vars map[string]interface{}) error
	Delete(ctx context.Context, opts ...DBOption) error
	Get(opts ...DBOption) (model.Database, error)
	WithByFrom(from string) DBOption
	WithoutByFrom(from string) DBOption
	WithByMysqlList() DBOption
	WithAppInstallID(appInstallID uint) DBOption
}

func NewIDatabaseRepo() IDatabaseRepo {
	return &DatabaseRepo{}
}

func (d *DatabaseRepo) Get(opts ...DBOption) (model.Database, error) {
	var database model.Database
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&database).Error
	return database, err
}

func (d *DatabaseRepo) Page(page, size int, opts ...DBOption) (int64, []model.Database, error) {
	var users []model.Database
	db := global.DB.Model(&model.Database{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&users).Error
	return count, users, err
}

func (d *DatabaseRepo) GetList(opts ...DBOption) ([]model.Database, error) {
	var databases []model.Database
	db := global.DB.Model(&model.Database{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&databases).Error
	return databases, err
}

func (d *DatabaseRepo) WithByMysqlList() DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("type = ? OR type = ?", "mysql", "mariadb")
	}
}

func (d *DatabaseRepo) WithByFrom(from string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("`from` = ?", from)
	}
}

func (d *DatabaseRepo) WithoutByFrom(from string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("`from` != ?", from)
	}
}
func (d *DatabaseRepo) WithAppInstallID(appInstallID uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("app_install_id = ?", appInstallID)
	}
}

func (d *DatabaseRepo) Create(ctx context.Context, database *model.Database) error {
	return getTx(ctx).Create(database).Error
}

func (d *DatabaseRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.Database{}).Where("id = ?", id).Updates(vars).Error
}

func (d *DatabaseRepo) Delete(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.Database{}).Error
}
