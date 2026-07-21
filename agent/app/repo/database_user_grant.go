package repo

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

type DatabaseUserGrantRepo struct{}

type IDatabaseUserGrantRepo interface {
	Get(opts ...DBOption) (model.DatabaseUserGrant, error)
	List(opts ...DBOption) ([]model.DatabaseUserGrant, error)
	Save(grant *model.DatabaseUserGrant) error
	Replace(dbType, database string, grants []model.DatabaseUserGrant) error
	Delete(opts ...DBOption) error
	Update(vars map[string]interface{}, opts ...DBOption) error
	WithByDatabase(database string) DBOption
	WithByDBName(dbName string) DBOption
	WithByDBNames(dbNames []string) DBOption
	WithByUser(username, host string) DBOption
}

func (u *DatabaseUserGrantRepo) Get(opts ...DBOption) (model.DatabaseUserGrant, error) {
	var grant model.DatabaseUserGrant
	db := global.DB.Model(&model.DatabaseUserGrant{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&grant).Error
	return grant, err
}

func NewIDatabaseUserGrantRepo() IDatabaseUserGrantRepo {
	return &DatabaseUserGrantRepo{}
}

func (u *DatabaseUserGrantRepo) List(opts ...DBOption) ([]model.DatabaseUserGrant, error) {
	var grants []model.DatabaseUserGrant
	db := global.DB.Model(&model.DatabaseUserGrant{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&grants).Error
	return grants, err
}

func (u *DatabaseUserGrantRepo) Save(grant *model.DatabaseUserGrant) error {
	return global.DB.Save(grant).Error
}

func (u *DatabaseUserGrantRepo) Replace(dbType, database string, grants []model.DatabaseUserGrant) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("`type` = ? AND database = ?", dbType, database).Delete(&model.DatabaseUserGrant{}).Error; err != nil {
			return err
		}
		if len(grants) != 0 {
			return tx.Create(&grants).Error
		}
		return nil
	})
}

func (u *DatabaseUserGrantRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.DatabaseUserGrant{}).Error
}

func (u *DatabaseUserGrantRepo) Update(vars map[string]interface{}, opts ...DBOption) error {
	db := global.DB.Model(&model.DatabaseUserGrant{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Updates(vars).Error
}

func (u *DatabaseUserGrantRepo) WithByDatabase(database string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("database = ?", database)
	}
}

func (u *DatabaseUserGrantRepo) WithByDBName(dbName string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("db_name = ?", dbName)
	}
}

func (u *DatabaseUserGrantRepo) WithByDBNames(dbNames []string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("db_name IN ?", dbNames)
	}
}

func (u *DatabaseUserGrantRepo) WithByUser(username, host string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("username = ? AND host = ?", username, host)
	}
}
