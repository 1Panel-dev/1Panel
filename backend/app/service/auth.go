package service

import (
	"strconv"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
	"github.com/1Panel-dev/1Panel/backend/utils/jwt"
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
	if info.Password != pass && nameSetting.Value == info.Name {
		return nil, errors.New("login failed")
	}
	mfa, err := settingRepo.Get(settingRepo.WithByKey("MFAStatus"))
	if err != nil {
		return nil, err
	}
	if mfa.Value == "enable" {
		mfaSecret, err := settingRepo.Get(settingRepo.WithByKey("MFASecret"))
		if err != nil {
			return nil, err
		}
		return &dto.UserLoginInfo{Name: nameSetting.Value, MfaStatus: mfa.Value, MfaSecret: mfaSecret.Value}, nil
	}

	return u.generateSession(c, info.Name, info.AuthMethod)
}

func (u *AuthService) MFALogin(c *gin.Context, info dto.MFALogin) (*dto.UserLoginInfo, error) {
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
	if info.Password != pass && nameSetting.Value == info.Name {
		return nil, errors.New("login failed")
	}

	return u.generateSession(c, info.Name, info.AuthMethod)
}

func (u *AuthService) generateSession(c *gin.Context, name, authMethod string) (*dto.UserLoginInfo, error) {
	setting, err := settingRepo.Get(settingRepo.WithByKey("SessionTimeout"))
	if err != nil {
		return nil, err
	}
	lifeTime, err := strconv.Atoi(setting.Value)
	if err != nil {
		return nil, err
	}

	if authMethod == constant.AuthMethodJWT {
		j := jwt.NewJWT()
		claims := j.CreateClaims(jwt.BaseClaims{
			Name: name,
		}, lifeTime)
		token, err := j.CreateToken(claims)
		if err != nil {
			return nil, err
		}
		return &dto.UserLoginInfo{Name: name, Token: token}, nil
	}
	sID, _ := c.Cookie(constant.SessionName)
	sessionUser, err := global.SESSION.Get(sID)
	if err != nil {
		sID = uuid.NewV4().String()
		c.SetCookie(constant.SessionName, sID, 604800, "", "", false, false)
		err := global.SESSION.Set(sID, sessionUser, lifeTime)
		if err != nil {
			return nil, err
		}
		return &dto.UserLoginInfo{Name: name}, nil
	}
	if err := global.SESSION.Set(sID, sessionUser, lifeTime); err != nil {
		return nil, err
	}

	return &dto.UserLoginInfo{Name: name}, nil
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
