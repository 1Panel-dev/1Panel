package repo

import (
	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/internal/entity"
)

type UserRepo interface {
	CreateUser(user *entity.User) error
	ListUser() (entity.User, error)
}

func NewUserRepo() UserRepo {
	return &userRepo{}
}

type userRepo struct {
}

func (u *userRepo) CreateUser(user *entity.User) error {
	return global.DB.Create(user).Error
}

func (u *userRepo) ListUser() (entity.User, error) {
	var user entity.User
	if err := global.DB.Model(&entity.User{}).First(&user).Error; err != nil {
		return user, err
	}
	return user, nil
}
