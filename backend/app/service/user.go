package service

import (
	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/utils/encrypt"
	"github.com/1Panel-dev/1Panel/utils/jwt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type UserService struct{}

type IUserService interface {
	Get(id uint) (*dto.UserInfo, error)
	Page(search dto.SearchWithPage) (int64, interface{}, error)
	Register(userDto dto.UserCreate) error
	Login(c *gin.Context, info dto.Login) (*dto.UserLoginInfo, error)
	LogOut(c *gin.Context) error
	Delete(name string) error
	Save(req model.User) error
	Update(id uint, upMap map[string]interface{}) error
	BatchDelete(ids []uint) error
}

func NewIUserService() IUserService {
	return &UserService{}
}

func (u *UserService) Get(id uint) (*dto.UserInfo, error) {
	user, err := userRepo.Get(commonRepo.WithByID(id))
	if err != nil {
		return nil, constant.ErrRecordNotFound
	}
	var dtoUser dto.UserInfo
	if err := copier.Copy(&dtoUser, &user); err != nil {
		return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	return &dtoUser, err
}

func (u *UserService) Page(search dto.SearchWithPage) (int64, interface{}, error) {
	total, users, err := userRepo.Page(search.Page, search.PageSize, commonRepo.WithLikeName(search.Name))
	var dtoUsers []dto.UserInfo
	for _, user := range users {
		var item dto.UserInfo
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
		return constant.ErrRecordExist
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
	lifeTime := global.CONF.Session.ExpiresTime
	sID, _ := c.Cookie(global.CONF.Session.SessionName)
	sessionUser, err := global.SESSION.Get(sID)
	if err != nil {
		sID = uuid.NewV4().String()
		c.SetCookie(global.CONF.Session.SessionName, sID, lifeTime, "", "", false, false)
		err := global.SESSION.Set(sID, sessionUser, lifeTime)
		if err != nil {
			return nil, err
		}
		return &dto.UserLoginInfo{Name: user.Name}, nil
	}
	if err := global.SESSION.Set(sID, sessionUser, lifeTime); err != nil {
		return nil, err
	}

	return &dto.UserLoginInfo{Name: user.Name}, nil
}

func (u *UserService) LogOut(c *gin.Context) error {
	sID, _ := c.Cookie(global.CONF.Session.SessionName)
	if sID != "" {
		c.SetCookie(global.CONF.Session.SessionName, sID, -1, "", "", false, false)
		err := global.SESSION.Delete(sID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *UserService) Delete(name string) error {
	return userRepo.Delete(commonRepo.WithByName(name))
}

func (u *UserService) BatchDelete(ids []uint) error {
	return userRepo.Delete(commonRepo.WithIdsIn(ids))
}

func (u *UserService) Save(req model.User) error {
	return userRepo.Save(req)
}

func (u *UserService) Update(id uint, upMap map[string]interface{}) error {
	return userRepo.Update(id, upMap)
}
