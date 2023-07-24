package repo

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
)

type RemoteDBRepo struct{}

type IRemoteDBRepo interface {
	GetList(opts ...DBOption) ([]model.RemoteDB, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.RemoteDB, error)
	Create(database *model.RemoteDB) error
	Update(id uint, vars map[string]interface{}) error
	Delete(opts ...DBOption) error
	Get(opts ...DBOption) (model.RemoteDB, error)
	WithByFrom(from string) DBOption
	WithoutByFrom(from string) DBOption
}

func NewIRemoteDBRepo() IRemoteDBRepo {
	return &RemoteDBRepo{}
}

func (u *RemoteDBRepo) Get(opts ...DBOption) (model.RemoteDB, error) {
	var database model.RemoteDB
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&database).Error
	return database, err
}

func (u *RemoteDBRepo) Page(page, size int, opts ...DBOption) (int64, []model.RemoteDB, error) {
	var users []model.RemoteDB
	db := global.DB.Model(&model.RemoteDB{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&users).Error
	return count, users, err
}

func (u *RemoteDBRepo) GetList(opts ...DBOption) ([]model.RemoteDB, error) {
	var databases []model.RemoteDB
	db := global.DB.Model(&model.RemoteDB{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&databases).Error
	return databases, err
}

func (c *RemoteDBRepo) WithByFrom(from string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("`from` != ?", from)
	}
}

func (c *RemoteDBRepo) WithoutByFrom(from string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("`from` != ?", from)
	}
}

func (u *RemoteDBRepo) Create(database *model.RemoteDB) error {
	return global.DB.Create(database).Error
}

func (u *RemoteDBRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.RemoteDB{}).Where("id = ?", id).Updates(vars).Error
}

func (u *RemoteDBRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.RemoteDB{}).Error
}
