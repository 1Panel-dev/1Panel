package repo

import (
	"context"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
)

type MysqlRepo struct{}

type IMysqlRepo interface {
	Get(opts ...DBOption) (model.DatabaseMysql, error)
	WithByMysqlName(mysqlName string) DBOption
	List(opts ...DBOption) ([]model.DatabaseMysql, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.DatabaseMysql, error)
	Create(ctx context.Context, mysql *model.DatabaseMysql) error
	Delete(ctx context.Context, opts ...DBOption) error
	Update(id uint, vars map[string]interface{}) error
	UpdateDatabaseInfo(id uint, vars map[string]interface{}) error
}

func NewIMysqlRepo() IMysqlRepo {
	return &MysqlRepo{}
}

func (u *MysqlRepo) Get(opts ...DBOption) (model.DatabaseMysql, error) {
	var mysql model.DatabaseMysql
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&mysql).Error
	return mysql, err
}

func (u *MysqlRepo) List(opts ...DBOption) ([]model.DatabaseMysql, error) {
	var users []model.DatabaseMysql
	db := global.DB.Model(&model.DatabaseMysql{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&users).Error
	return users, err
}

func (u *MysqlRepo) Page(page, size int, opts ...DBOption) (int64, []model.DatabaseMysql, error) {
	var users []model.DatabaseMysql
	db := global.DB.Model(&model.DatabaseMysql{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&users).Error
	return count, users, err
}

func (u *MysqlRepo) Create(ctx context.Context, mysql *model.DatabaseMysql) error {
	return getTx(ctx).Create(mysql).Error
}

func (u *MysqlRepo) Delete(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.DatabaseMysql{}).Error
}

func (u *MysqlRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.DatabaseMysql{}).Where("id = ?", id).Updates(vars).Error
}

func (u *MysqlRepo) UpdateDatabaseInfo(id uint, vars map[string]interface{}) error {
	if err := global.DB.Model(&model.AppInstall{}).Where("id = ?", id).Updates(vars).Error; err != nil {
		return err
	}
	return nil
}

func (u *MysqlRepo) WithByMysqlName(mysqlName string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("mysql_name = ?", mysqlName)
	}
}
