package service

import "github.com/1Panel-dev/1Panel/internal/repo"

type UserService interface {
}

func NewUserService() UserService {
	return &userService{}
}

type userService struct {
	userRepo repo.UserRepo
}

func (u *userService) Get() {
}
