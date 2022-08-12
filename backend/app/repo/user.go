package repo

import (
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/global"
)

type UserRepo struct{}

type IUserRepo interface {
	Get(opts ...DBOption) (model.User, error)
	GetList(opts ...DBOption) ([]model.User, error)
	Page(limit, offset int, opts ...DBOption) (int64, []model.User, error)
	Create(user *model.User) error
	Update(id uint, vars map[string]interface{}) error
	Save(user model.User) error
	Delete(opts ...DBOption) error
}

func NewIUserService() IUserRepo {
	return &UserRepo{}
}

func (u *UserRepo) Get(opts ...DBOption) (model.User, error) {
	var user model.User
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&user).Error
	return user, err
}

func (u *UserRepo) GetList(opts ...DBOption) ([]model.User, error) {
	var users []model.User
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&users).Error
	return users, err
}

func (u *UserRepo) Page(page, size int, opts ...DBOption) (int64, []model.User, error) {
	var users []model.User
	db := global.DB.Model(&model.User{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&users).Error
	return count, users, err
}

func (u *UserRepo) Create(user *model.User) error {
	return global.DB.Create(user).Error
}

func (u *UserRepo) Update(id uint, vars map[string]interface{}) error {
	return global.DB.Model(&model.User{}).Where("id = ?", id).Updates(vars).Error
}

func (u *UserRepo) Save(user model.User) error {
	return global.DB.Save(user).Error
}

func (u *UserRepo) Delete(opts ...DBOption) error {
	db := global.DB
	for _, opt := range opts {
		db = opt(db)
	}

	return db.Delete(&model.User{}).Error
}
