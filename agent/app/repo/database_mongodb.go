package repo

import (
	"context"
	"fmt"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"gorm.io/gorm"
)

type MongodbRepo struct{}

type IMongodbRepo interface {
	Get(opts ...DBOption) (model.DatabaseMongodb, error)
	WithByMongodbName(mongodbName string) DBOption
	List(opts ...DBOption) ([]model.DatabaseMongodb, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.DatabaseMongodb, error)
	Create(ctx context.Context, mongodb *model.DatabaseMongodb) error
	Delete(ctx context.Context, opts ...DBOption) error
	Update(id uint, vars map[string]interface{}) error
	DeleteLocal(ctx context.Context) error
}

func NewIMongodbRepo() IMongodbRepo {
	return &MongodbRepo{}
}

func (u *MongodbRepo) Get(opts ...DBOption) (model.DatabaseMongodb, error) {
	var mongodb model.DatabaseMongodb
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.First(&mongodb).Error; err != nil {
		return mongodb, err
	}

	pass, err := encrypt.StringDecrypt(mongodb.Password)
	if err != nil {
		global.LOG.Errorf("decrypt mongodb db %s password failed, err: %v", mongodb.Name, err)
	}
	mongodb.Password = pass
	return mongodb, err
}

func (u *MongodbRepo) List(opts ...DBOption) ([]model.DatabaseMongodb, error) {
	var mongodbs []model.DatabaseMongodb
	db := global.DB.Model(&model.DatabaseMongodb{})
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.Find(&mongodbs).Error; err != nil {
		return mongodbs, err
	}
	for i := 0; i < len(mongodbs); i++ {
		pass, err := encrypt.StringDecrypt(mongodbs[i].Password)
		if err != nil {
			global.LOG.Errorf("decrypt mongodb db %s password failed, err: %v", mongodbs[i].Name, err)
		}
		mongodbs[i].Password = pass
	}
	return mongodbs, nil
}

func (u *MongodbRepo) Page(page, size int, opts ...DBOption) (int64, []model.DatabaseMongodb, error) {
	var mongodbs []model.DatabaseMongodb
	db := global.DB.Model(&model.DatabaseMongodb{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	if err := db.Limit(size).Offset(size * (page - 1)).Find(&mongodbs).Error; err != nil {
		return count, mongodbs, err
	}
	for i := 0; i < len(mongodbs); i++ {
		pass, err := encrypt.StringDecrypt(mongodbs[i].Password)
		if err != nil {
			global.LOG.Errorf("decrypt mongodb db %s password failed, err: %v", mongodbs[i].Name, err)
		}
		mongodbs[i].Password = pass
	}
	return count, mongodbs, nil
}

func (u *MongodbRepo) Create(ctx context.Context, mongodb *model.DatabaseMongodb) error {
	pass, err := encrypt.StringEncrypt(mongodb.Password)
	if err != nil {
		return fmt.Errorf("encrypt mongodb db %s password failed, err: %v", mongodb.Name, err)
	}
	mongodb.Password = pass
	return getTx(ctx).Create(mongodb).Error
}

func (u *MongodbRepo) Delete(ctx context.Context, opts ...DBOption) error {
	return getTx(ctx, opts...).Delete(&model.DatabaseMongodb{}).Error
}

func (u *MongodbRepo) DeleteLocal(ctx context.Context) error {
	return getTx(ctx).Where("`from` = ?", "local").Delete(&model.DatabaseMongodb{}).Error
}

func (u *MongodbRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.DatabaseMongodb{}).Where("id = ?", id).Updates(vars).Error
}

func (u *MongodbRepo) WithByMongodbName(mongodbName string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("mongodb_name = ?", mongodbName)
	}
}
