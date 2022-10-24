package repo

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
)

type MysqlRepo struct{}

type IMysqlRepo interface {
	Get(opts ...DBOption) (model.DatabaseMysql, error)
	List(opts ...DBOption) ([]model.DatabaseMysql, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.DatabaseMysql, error)
	Create(mysql *model.DatabaseMysql) error
	Delete(opts ...DBOption) error
	Update(id uint, vars map[string]interface{}) error
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

func (u *MysqlRepo) Create(mysql *model.DatabaseMysql) error {
	return global.DB.Create(mysql).Error
}

func (u *MysqlRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.DatabaseMysql{}).Error
}

func (u *MysqlRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.DatabaseMysql{}).Where("id = ?", id).Updates(vars).Error
}
