package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
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
	WithAppInstallID(appInstallID uint) DBOption
	WithTypeList(dbType string) DBOption
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
	if err := db.First(&database).Error; err != nil {
		return database, err
	}
	pass, err := encrypt.StringDecrypt(database.Password)
	if err != nil {
		global.LOG.Errorf("decrypt database %s password failed, err: %v", database.Name, err)
	}
	database.Password = pass
	return database, nil
}

func (d *DatabaseRepo) Page(page, size int, opts ...DBOption) (int64, []model.Database, error) {
	var databases []model.Database
	db := global.DB.Model(&model.Database{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	if err := db.Limit(size).Offset(size * (page - 1)).Find(&databases).Error; err != nil {
		return count, databases, err
	}
	for i := 0; i < len(databases); i++ {
		pass, err := encrypt.StringDecrypt(databases[i].Password)
		if err != nil {
			global.LOG.Errorf("decrypt database db %s password failed, err: %v", databases[i].Name, err)
		}
		databases[i].Password = pass
	}
	return count, databases, nil
}

func (d *DatabaseRepo) GetList(opts ...DBOption) ([]model.Database, error) {
	var databases []model.Database
	db := global.DB.Model(&model.Database{})
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.Find(&databases).Error; err != nil {
		return databases, err
	}
	for i := 0; i < len(databases); i++ {
		pass, err := encrypt.StringDecrypt(databases[i].Password)
		if err != nil {
			global.LOG.Errorf("decrypt database db %s password failed, err: %v", databases[i].Name, err)
		}
		databases[i].Password = pass
	}
	return databases, nil
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

func (d *DatabaseRepo) WithTypeList(dbType string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if !strings.Contains(dbType, ",") {
			return g.Where("`type` = ?", dbType)
		}
		types := strings.Split(dbType, ",")
		var (
			rules  []string
			values []interface{}
		)
		for _, ty := range types {
			if len(ty) != 0 {
				rules = append(rules, "`type` = ?")
				values = append(values, ty)
			}
		}
		return g.Where(strings.Join(rules, " OR "), values...)
	}
}

func (d *DatabaseRepo) WithAppInstallID(appInstallID uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("app_install_id = ?", appInstallID)
	}
}

func (d *DatabaseRepo) Create(ctx context.Context, database *model.Database) error {
	pass, err := encrypt.StringEncrypt(database.Password)
	if err != nil {
		return fmt.Errorf("decrypt database db %s password failed, err: %v", database.Name, err)
	}
	database.Password = pass
	return getTx(ctx).Create(database).Error
}

func (d *DatabaseRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.Database{}).Where("id = ?", id).Updates(vars).Error
}

func (d *DatabaseRepo) Delete(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.Database{}).Error
}
