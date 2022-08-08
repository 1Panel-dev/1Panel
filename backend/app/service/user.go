package service

import (
	"errors"

	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/utils/encrypt"
	"github.com/1Panel-dev/1Panel/utils/jwt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"

	"gorm.io/gorm"
)

type UserService struct{}

type IUserService interface {
	Get(name string) (*dto.UserBack, error)
	Page(page, size int) (int64, interface{}, error)
	Register(userDto dto.UserCreate) error
	Login(c *gin.Context, info dto.Login) (*dto.UserLoginInfo, error)
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
	var dtoUser dto.UserBack
	if err := copier.Copy(&dtoUser, &user); err != nil {
		return nil, constant.ErrCopyTransform
	}
	return &dtoUser, err
}

func (u *UserService) Page(page, size int) (int64, interface{}, error) {
	total, users, err := userRepo.Page(page, size)
	var dtoUsers []dto.UserBack
	for _, user := range users {
		var item dto.UserBack
		if err := copier.Copy(&item, &user); err != nil {
			return 0, nil, constant.ErrCopyTransform
		}
		dtoUsers = append(dtoUsers, item)
	}
	return total, dtoUsers, err
}

func (u *UserService) Register(userDto dto.UserCreate) error {
	var user model.User
	if err := copier.Copy(&user, &userDto); err != nil {
		return constant.ErrCopyTransform
	}
	if !errors.Is(global.DB.Where("name = ?", user.Name).First(&user).Error, gorm.ErrRecordNotFound) {
		return constant.ErrRecordExist
	}
	return userRepo.Create(&user)
}

func (u *UserService) Login(c *gin.Context, info dto.Login) (*dto.UserLoginInfo, error) {
	user, err := userRepo.Get(commonRepo.WithByName(info.Name))
	if err != nil {
		return nil, err
	}
	pass, err := encrypt.StringDecrypt(user.Password)
	if err != nil {
		return nil, err
	}
	if info.Password != pass {
		return nil, errors.New("login failed")
	}
	if info.AuthMethod == constant.AuthMethodJWT {
		j := jwt.NewJWT()
		claims := j.CreateClaims(jwt.BaseClaims{
			ID:   user.ID,
			Name: user.Name,
		})
		token, err := j.CreateToken(claims)
		if err != nil {
			return nil, err
		}
		return &dto.UserLoginInfo{Name: user.Name, Token: token}, err
	}

	sID, _ := c.Cookie(global.CONF.Session.SessionName)
	if sID != "" {
		c.SetCookie(global.CONF.Session.SessionName, "", -1, "", "", false, false)
	}
	session, err := global.SESSION.New(c.Request, global.CONF.Session.SessionName)
	if err != nil {
		return nil, err
	}
	session.Values[global.CONF.Session.SessionUserKey] = user
	if err := global.SESSION.Save(c.Request, c.Writer, session); err != nil {
		return nil, err
	}

	return &dto.UserLoginInfo{Name: user.Name}, err
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
