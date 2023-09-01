package repo

import (
	"context"
	"fmt"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
	"gorm.io/gorm"
)

type MysqlRepo struct{}

type IMysqlRepo interface {
	Get(opts ...DBOption) (model.DatabaseMysql, error)
	WithByMysqlName(mysqlName string) DBOption
	WithByFrom(from string) DBOption
	List(opts ...DBOption) ([]model.DatabaseMysql, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.DatabaseMysql, error)
	Create(ctx context.Context, mysql *model.DatabaseMysql) error
	Delete(ctx context.Context, opts ...DBOption) error
	Update(id uint, vars map[string]interface{}) error
	DeleteLocal(ctx context.Context) error
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
	if err := db.First(&mysql).Error; err != nil {
		return mysql, err
	}

	pass, err := encrypt.StringDecrypt(mysql.Password)
	if err != nil {
		global.LOG.Errorf("decrypt database db %s password failed, err: %v", mysql.Name, err)
	}
	mysql.Password = pass
	return mysql, err
}

func (u *MysqlRepo) List(opts ...DBOption) ([]model.DatabaseMysql, error) {
	var mysqls []model.DatabaseMysql
	db := global.DB.Model(&model.DatabaseMysql{})
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.Find(&mysqls).Error; err != nil {
		return mysqls, err
	}
	for i := 0; i < len(mysqls); i++ {
		pass, err := encrypt.StringDecrypt(mysqls[i].Password)
		if err != nil {
			global.LOG.Errorf("decrypt database db %s password failed, err: %v", mysqls[i].Name, err)
		}
		mysqls[i].Password = pass
	}
	return mysqls, nil
}

func (u *MysqlRepo) Page(page, size int, opts ...DBOption) (int64, []model.DatabaseMysql, error) {
	var mysqls []model.DatabaseMysql
	db := global.DB.Model(&model.DatabaseMysql{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	if err := db.Limit(size).Offset(size * (page - 1)).Find(&mysqls).Error; err != nil {
		return count, mysqls, err
	}
	for i := 0; i < len(mysqls); i++ {
		pass, err := encrypt.StringDecrypt(mysqls[i].Password)
		if err != nil {
			global.LOG.Errorf("decrypt database db %s password failed, err: %v", mysqls[i].Name, err)
		}
		mysqls[i].Password = pass
	}
	return count, mysqls, nil
}

func (u *MysqlRepo) Create(ctx context.Context, mysql *model.DatabaseMysql) error {
	pass, err := encrypt.StringEncrypt(mysql.Password)
	if err != nil {
		return fmt.Errorf("decrypt database db %s password failed, err: %v", mysql.Name, err)
	}
	mysql.Password = pass
	return getTx(ctx).Create(mysql).Error
}

func (u *MysqlRepo) Delete(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.DatabaseMysql{}).Error
}

func (u *MysqlRepo) DeleteLocal(ctx context.Context) error {
	return getTx(ctx).Where("`from` = ?", "local").Delete(&model.DatabaseMysql{}).Error
}

func (u *MysqlRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.DatabaseMysql{}).Where("id = ?", id).Updates(vars).Error
}

func (u *MysqlRepo) WithByMysqlName(mysqlName string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("mysql_name = ?", mysqlName)
	}
}

func (u *MysqlRepo) WithByFrom(from string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		if len(from) != 0 {
			return g.Where("`from` = ?", from)
		}
		return g
	}
}
