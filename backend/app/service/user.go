package service

import (
	"errors"

	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/global"

	"gorm.io/gorm"
)

type UserService struct{}

type IUserService interface {
	Get(name string) (*dto.UserBack, error)
	Page(page, size int) (int64, interface{}, error)
	Register(userDto dto.UserCreate) error
	Login(info *model.User) (*dto.UserBack, error)
	Delete(name string) error
	Save(req model.User) error
	Update(upMap map[string]interface{}) error
}

func NewIUserService() IUserService {
	return &UserService{}
}

func (u *UserService) Get(name string) (*dto.UserBack, error) {
	user, err := userRepo.Get(commonRepo.WithByName(name))
	if err != nil {
		return nil, err
	}
	dtoUser := &dto.UserBack{
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
	return dtoUser, nil
}

func (u *UserService) Page(page, size int) (int64, interface{}, error) {
	total, users, err := userRepo.Page(page, size)
	var dtoUsers []dto.UserBack
	for _, user := range users {
		dtoUsers = append(dtoUsers, dto.UserBack{
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		})
	}
	return total, dtoUsers, err
}

func (u *UserService) Register(userDto dto.UserCreate) error {
	user := userDto.UserCreateToMo()
	if !errors.Is(global.DB.Where("name = ?", user.Name).First(&user).Error, gorm.ErrRecordNotFound) {
		return errors.New("用户名已注册")
	}
	return userRepo.Create(&user)
}

func (u *UserService) Login(info *model.User) (*dto.UserBack, error) {
	user, err := userRepo.Get(commonRepo.WithByName(info.Name))
	if err != nil {
		return nil, err
	}
	if user.Password != info.Password {
		return nil, errors.New("登录失败")
	}
	dtoUser := &dto.UserBack{
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
	return dtoUser, err
}

func (u *UserService) Delete(name string) error {
	return userRepo.Delete(commonRepo.WithByName(name))
}

func (u *UserService) Save(req model.User) error {
	return userRepo.Save(req)
}

func (u *UserService) Update(upMap map[string]interface{}) error {
	return userRepo.Update(upMap)
}
