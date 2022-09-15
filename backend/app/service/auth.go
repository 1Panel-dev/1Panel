package service

import (
	"strconv"

	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/utils/encrypt"
	"github.com/1Panel-dev/1Panel/utils/jwt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type AuthService struct{}

type IAuthService interface {
	SafetyStatus(c *gin.Context) error
	VerifyCode(code string) (bool, error)
	Login(c *gin.Context, info dto.Login) (*dto.UserLoginInfo, error)
	LogOut(c *gin.Context) error
}

func NewIAuthService() IAuthService {
	return &AuthService{}
}

func (u *AuthService) Login(c *gin.Context, info dto.Login) (*dto.UserLoginInfo, error) {
	nameSetting, err := settingRepo.Get(settingRepo.WithByKey("UserName"))
	if err != nil {
		return nil, errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}
	passwrodSetting, err := settingRepo.Get(settingRepo.WithByKey("Password"))
	if err != nil {
		return nil, errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}
	pass, err := encrypt.StringDecrypt(passwrodSetting.Value)
	if err != nil {
		return nil, err
	}
	if info.Password != pass {
		return nil, errors.New("login failed")
	}
	setting, err := settingRepo.Get(settingRepo.WithByKey("SessionTimeout"))
	if err != nil {
		return nil, err
	}
	lifeTime, err := strconv.Atoi(setting.Value)
	if err != nil {
		return nil, err
	}

	if info.AuthMethod == constant.AuthMethodJWT {
		j := jwt.NewJWT()
		claims := j.CreateClaims(jwt.BaseClaims{
			Name: nameSetting.Value,
		}, lifeTime)
		token, err := j.CreateToken(claims)
		if err != nil {
			return nil, err
		}
		return &dto.UserLoginInfo{Name: nameSetting.Value, Token: token}, err
	}
	sID, _ := c.Cookie(constant.SessionName)
	sessionUser, err := global.SESSION.Get(sID)
	if err != nil {
		sID = uuid.NewV4().String()
		c.SetCookie(constant.SessionName, sID, lifeTime, "", "", false, false)
		err := global.SESSION.Set(sID, sessionUser, lifeTime)
		if err != nil {
			return nil, err
		}
		return &dto.UserLoginInfo{Name: nameSetting.Value}, nil
	}
	if err := global.SESSION.Set(sID, sessionUser, lifeTime); err != nil {
		return nil, err
	}

	return &dto.UserLoginInfo{Name: nameSetting.Value}, nil
}

func (u *AuthService) LogOut(c *gin.Context) error {
	sID, _ := c.Cookie(constant.SessionName)
	if sID != "" {
		c.SetCookie(constant.SessionName, sID, -1, "", "", false, false)
		err := global.SESSION.Delete(sID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *AuthService) VerifyCode(code string) (bool, error) {
	setting, err := settingRepo.Get(settingRepo.WithByKey("SecurityEntrance"))
	if err != nil {
		return false, err
	}
	return setting.Value == code, nil
}

func (u *AuthService) SafetyStatus(c *gin.Context) error {
	setting, err := settingRepo.Get(settingRepo.WithByKey("SecurityEntrance"))
	if err != nil {
		return err
	}
	codeWithEcrypt, err := c.Cookie(encrypt.Md5(setting.Value))
	if err != nil {
		return err
	}
	code, err := encrypt.StringDecrypt(codeWithEcrypt)
	if err != nil {
		return err
	}
	if code != encrypt.Md5(setting.Value) {
		return errors.New("code not match")
	}
	return nil
}
