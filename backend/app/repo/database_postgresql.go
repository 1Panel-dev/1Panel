package repo

import (
	"context"
	"fmt"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
	"gorm.io/gorm"
)

type PostgresqlRepo struct{}

type IPostgresqlRepo interface {
	Get(opts ...DBOption) (model.DatabasePostgresql, error)
	WithByPostgresqlName(postgresqlName string) DBOption
	WithByFrom(from string) DBOption
	List(opts ...DBOption) ([]model.DatabasePostgresql, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.DatabasePostgresql, error)
	Create(ctx context.Context, postgresql *model.DatabasePostgresql) error
	Delete(ctx context.Context, opts ...DBOption) error
	Update(id uint, vars map[string]interface{}) error
	DeleteLocal(ctx context.Context) error
}

func NewIPostgresqlRepo() IPostgresqlRepo {
	return &PostgresqlRepo{}
}

func (u *PostgresqlRepo) Get(opts ...DBOption) (model.DatabasePostgresql, error) {
	var postgresql model.DatabasePostgresql
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.First(&postgresql).Error; err != nil {
		return postgresql, err
	}

	pass, err := encrypt.StringDecrypt(postgresql.Password)
	if err != nil {
		global.LOG.Errorf("decrypt database db %s password failed, err: %v", postgresql.Name, err)
	}
	postgresql.Password = pass
	return postgresql, err
}

func (u *PostgresqlRepo) List(opts ...DBOption) ([]model.DatabasePostgresql, error) {
	var postgresqls []model.DatabasePostgresql
	db := global.DB.Model(&model.DatabasePostgresql{})
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.Find(&postgresqls).Error; err != nil {
		return postgresqls, err
	}
	for i := 0; i < len(postgresqls); i++ {
		pass, err := encrypt.StringDecrypt(postgresqls[i].Password)
		if err != nil {
			global.LOG.Errorf("decrypt database db %s password failed, err: %v", postgresqls[i].Name, err)
		}
		postgresqls[i].Password = pass
	}
	return postgresqls, nil
}

func (u *PostgresqlRepo) Page(page, size int, opts ...DBOption) (int64, []model.DatabasePostgresql, error) {
	var postgresqls []model.DatabasePostgresql
	db := global.DB.Model(&model.DatabasePostgresql{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	if err := db.Limit(size).Offset(size * (page - 1)).Find(&postgresqls).Error; err != nil {
		return count, postgresqls, err
	}
	for i := 0; i < len(postgresqls); i++ {
		pass, err := encrypt.StringDecrypt(postgresqls[i].Password)
		if err != nil {
			global.LOG.Errorf("decrypt database db %s password failed, err: %v", postgresqls[i].Name, err)
		}
		postgresqls[i].Password = pass
	}
	return count, postgresqls, nil
}

func (u *PostgresqlRepo) Create(ctx context.Context, postgresql *model.DatabasePostgresql) error {
	pass, err := encrypt.StringEncrypt(postgresql.Password)
	if err != nil {
		return fmt.Errorf("decrypt database db %s password failed, err: %v", postgresql.Name, err)
	}
	postgresql.Password = pass
	return getTx(ctx).Create(postgresql).Error
}

func (u *PostgresqlRepo) Delete(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.DatabasePostgresql{}).Error
}

func (u *PostgresqlRepo) DeleteLocal(ctx context.Context) error {
	return getTx(ctx).Where("`from` = ?", "local").Delete(&model.DatabasePostgresql{}).Error
}

func (u *PostgresqlRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.DatabasePostgresql{}).Where("id = ?", id).Updates(vars).Error
}

func (u *PostgresqlRepo) WithByPostgresqlName(postgresqlName string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("postgresql_name = ?", postgresqlName)
	}
}

func (u *PostgresqlRepo) WithByFrom(from string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(from) != 0 {
			return g.Where("`from` = ?", from)
		}
		return g
	}
}
