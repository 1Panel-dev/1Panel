package repo

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"gorm.io/gorm"
)

type DatabaseUserRepo struct{}

type IDatabaseUserRepo interface {
	Get(opts ...DBOption) (model.DatabaseUser, error)
	List(opts ...DBOption) ([]model.DatabaseUser, error)
	Save(user *model.DatabaseUser) error
	Delete(opts ...DBOption) error
	Update(vars map[string]interface{}, opts ...DBOption) error
	WithByDatabase(database string) DBOption
	WithByUser(username, host string) DBOption
}

func NewIDatabaseUserRepo() IDatabaseUserRepo {
	return &DatabaseUserRepo{}
}

func (u *DatabaseUserRepo) Get(opts ...DBOption) (model.DatabaseUser, error) {
	var user model.DatabaseUser
	db := global.DB.Model(&model.DatabaseUser{})
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.First(&user).Error; err != nil {
		return user, err
	}
	password, err := encrypt.StringDecrypt(user.Password)
	if err != nil {
		global.LOG.Errorf("decrypt database user %s password failed, err: %v", user.Username, err)
	}
	user.Password = password
	return user, nil
}

func (u *DatabaseUserRepo) List(opts ...DBOption) ([]model.DatabaseUser, error) {
	var users []model.DatabaseUser
	db := global.DB.Model(&model.DatabaseUser{})
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.Find(&users).Error; err != nil {
		return users, err
	}
	for i := 0; i < len(users); i++ {
		password, err := encrypt.StringDecrypt(users[i].Password)
		if err != nil {
			global.LOG.Errorf("decrypt database user %s password failed, err: %v", users[i].Username, err)
		}
		users[i].Password = password
	}
	return users, nil
}

func (u *DatabaseUserRepo) Save(user *model.DatabaseUser) error {
	if len(user.Password) != 0 {
		password, err := encrypt.StringEncrypt(user.Password)
		if err != nil {
			return fmt.Errorf("encrypt database user %s password failed, err: %v", user.Username, err)
		}
		user.Password = password
	}
	return global.DB.Save(user).Error
}

func (u *DatabaseUserRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.DatabaseUser{}).Error
}

func (u *DatabaseUserRepo) Update(vars map[string]interface{}, opts ...DBOption) error {
	db := global.DB.Model(&model.DatabaseUser{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Updates(vars).Error
}

func (u *DatabaseUserRepo) WithByDatabase(database string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("database = ?", database)
	}
}

func (u *DatabaseUserRepo) WithByUser(username, host string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("username = ? AND host = ?", username, host)
	}
}
