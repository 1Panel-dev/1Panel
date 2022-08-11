package service

import (
	"github.com/pkg/errors"

	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/init/session"
	"github.com/1Panel-dev/1Panel/utils/encrypt"
	"github.com/1Panel-dev/1Panel/utils/jwt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
)

type UserService struct{}

type IUserService interface {
	Get(name string) (*dto.UserBack, error)
	Page(search dto.UserPage) (int64, interface{}, error)
	Register(userDto dto.UserCreate) error
	Login(c *gin.Context, info dto.Login) (*dto.UserLoginInfo, error)
	LogOut(c *gin.Context) error
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
		return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	return &dtoUser, err
}

func (u *UserService) Page(search dto.UserPage) (int64, interface{}, error) {
	total, users, err := userRepo.Page(search.Page, search.PageSize, commonRepo.WithLikeName(search.Name))
	var dtoUsers []dto.UserBack
	for _, user := range users {
		var item dto.UserBack
		if err := copier.Copy(&item, &user); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtoUsers = append(dtoUsers, item)
	}
	return total, dtoUsers, err
}

func (u *UserService) Register(userDto dto.UserCreate) error {
	user, _ := userRepo.Get(commonRepo.WithByName(userDto.Name))
	if user.ID != 0 {
		return errors.Wrap(constant.ErrRecordExist, "data exist")
	}
	if err := copier.Copy(&user, &userDto); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	return userRepo.Create(&user)
}

func (u *UserService) Login(c *gin.Context, info dto.Login) (*dto.UserLoginInfo, error) {
	user, err := userRepo.Get(commonRepo.WithByName(info.Name))
	if err != nil {
		return nil, errors.WithMessage(constant.ErrRecordNotFound, err.Error())
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
	sessionItem, err := global.SESSION.Get(c.Request, global.CONF.Session.SessionName)
	if err != nil {
		return nil, err
	}
	sessionItem.Values[global.CONF.Session.SessionUserKey] = session.SessionUser{
		ID:   user.ID,
		Name: user.Name,
	}
	if err := global.SESSION.Save(c.Request, c.Writer, sessionItem); err != nil {
		return nil, err
	}

	return &dto.UserLoginInfo{Name: user.Name}, err
}

func (u *UserService) LogOut(c *gin.Context) error {
	sID, _ := c.Cookie(global.CONF.Session.SessionName)
	if sID != "" {
		c.SetCookie(global.CONF.Session.SessionName, "", -1, "", "", false, false)
	}
	sessionItem, err := global.SESSION.Get(c.Request, global.CONF.Session.SessionName)
	if err != nil {
		return err
	}
	sessionItem.Options.MaxAge = -1
	if err := global.SESSION.Save(c.Request, c.Writer, sessionItem); err != nil {
		return err
	}

	return nil
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
