package service

import (
	"crypto/hmac"
	"encoding/base64"
	"strconv"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
	"github.com/1Panel-dev/1Panel/backend/utils/jwt"
	"github.com/1Panel-dev/1Panel/backend/utils/mfa"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type AuthService struct{}

type IAuthService interface {
	GetResponsePage() (string, error)
	VerifyCode(code string) (bool, error)
	Login(c *gin.Context, info dto.Login, entrance string) (*dto.UserLoginInfo, error)
	LogOut(c *gin.Context) error
	MFALogin(c *gin.Context, info dto.MFALogin, entrance string) (*dto.UserLoginInfo, error)
	GetSecurityEntrance() string
	IsLogin(c *gin.Context) bool
}

func NewIAuthService() IAuthService {
	return &AuthService{}
}

func (u *AuthService) Login(c *gin.Context, info dto.Login, entrance string) (*dto.UserLoginInfo, error) {
	nameSetting, err := settingRepo.Get(settingRepo.WithByKey("UserName"))
	if err != nil {
		return nil, errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}
	if nameSetting.Value != info.Name {
		return nil, constant.ErrAuth
	}
	if err = checkPassword(info.Password); err != nil {
		return nil, err
	}
	entranceSetting, err := settingRepo.Get(settingRepo.WithByKey("SecurityEntrance"))
	if err != nil {
		return nil, err
	}
	if len(entranceSetting.Value) != 0 && entranceSetting.Value != entrance {
		return nil, buserr.New(constant.ErrEntrance)
	}
	mfa, err := settingRepo.Get(settingRepo.WithByKey("MFAStatus"))
	if err != nil {
		return nil, err
	}
	if err = settingRepo.Update("Language", info.Language); err != nil {
		return nil, err
	}
	if mfa.Value == "enable" {
		return &dto.UserLoginInfo{Name: nameSetting.Value, MfaStatus: mfa.Value}, nil
	}

	loginUser, err := u.generateSession(c, info.Name, info.AuthMethod)
	if err != nil {
		return nil, err
	}
	if entrance != "" {
		entranceValue := base64.StdEncoding.EncodeToString([]byte(entrance))
		c.SetCookie("SecurityEntrance", entranceValue, 0, "", "", false, true)
	}
	return loginUser, nil
}

func (u *AuthService) MFALogin(c *gin.Context, info dto.MFALogin, entrance string) (*dto.UserLoginInfo, error) {
	nameSetting, err := settingRepo.Get(settingRepo.WithByKey("UserName"))
	if err != nil {
		return nil, errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}
	if nameSetting.Value != info.Name {
		return nil, constant.ErrAuth
	}
	if err = checkPassword(info.Password); err != nil {
		return nil, err
	}
	entranceSetting, err := settingRepo.Get(settingRepo.WithByKey("SecurityEntrance"))
	if err != nil {
		return nil, err
	}
	if len(entranceSetting.Value) != 0 && entranceSetting.Value != entrance {
		return nil, buserr.New(constant.ErrEntrance)
	}
	mfaSecret, err := settingRepo.Get(settingRepo.WithByKey("MFASecret"))
	if err != nil {
		return nil, err
	}
	mfaInterval, err := settingRepo.Get(settingRepo.WithByKey("MFAInterval"))
	if err != nil {
		return nil, err
	}
	success := mfa.ValidCode(info.Code, mfaInterval.Value, mfaSecret.Value)
	if !success {
		return nil, constant.ErrAuth
	}

	loginUser, err := u.generateSession(c, info.Name, info.AuthMethod)
	if err != nil {
		return nil, err
	}
	if entrance != "" {
		entranceValue := base64.StdEncoding.EncodeToString([]byte(entrance))
		c.SetCookie("SecurityEntrance", entranceValue, 0, "", "", false, true)
	}
	return loginUser, nil
}

func (u *AuthService) generateSession(c *gin.Context, name, authMethod string) (*dto.UserLoginInfo, error) {
	setting, err := settingRepo.Get(settingRepo.WithByKey("SessionTimeout"))
	if err != nil {
		return nil, err
	}
	httpsSetting, err := settingRepo.Get(settingRepo.WithByKey("SSL"))
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
		})
		token, err := j.CreateToken(claims)
		if err != nil {
			return nil, err
		}
		return &dto.UserLoginInfo{Name: name, Token: token}, nil
	}
	sID, _ := c.Cookie(constant.SessionName)
	sessionUser, err := global.SESSION.Get(sID)
	if err != nil {
		sID = uuid.New().String()
		c.SetCookie(constant.SessionName, sID, 0, "", "", httpsSetting.Value == "enable", true)
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
	httpsSetting, err := settingRepo.Get(settingRepo.WithByKey("SSL"))
	if err != nil {
		return err
	}
	sID, _ := c.Cookie(constant.SessionName)
	if sID != "" {
		c.SetCookie(constant.SessionName, sID, -1, "", "", httpsSetting.Value == "enable", true)
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
	return hmac.Equal([]byte(setting.Value), []byte(code)), nil
}

func (u *AuthService) GetResponsePage() (string, error) {
	pageCode, err := settingRepo.Get(settingRepo.WithByKey("NoAuthSetting"))
	if err != nil {
		return "", err
	}
	return pageCode.Value, nil
}

func (u *AuthService) GetSecurityEntrance() string {
	status, err := settingRepo.Get(settingRepo.WithByKey("SecurityEntrance"))
	if err != nil {
		return ""
	}
	if len(status.Value) == 0 {
		return ""
	}
	return status.Value
}

func (u *AuthService) IsLogin(c *gin.Context) bool {
	sID, _ := c.Cookie(constant.SessionName)
	_, err := global.SESSION.Get(sID)
	return err == nil
}

func checkPassword(password string) error {
	priKey, _ := settingRepo.Get(settingRepo.WithByKey("PASSWORD_PRIVATE_KEY"))

	privateKey, err := encrypt.ParseRSAPrivateKey(priKey.Value)
	if err != nil {
		return err
	}
	loginPassword, err := encrypt.DecryptPassword(password, privateKey)
	if err != nil {
		return err
	}
	passwordSetting, err := settingRepo.Get(settingRepo.WithByKey("Password"))
	if err != nil {
		return errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}
	existPassword, err := encrypt.StringDecrypt(passwordSetting.Value)
	if err != nil {
		return err
	}
	if !hmac.Equal([]byte(loginPassword), []byte(existPassword)) {
		return constant.ErrAuth
	}
	return nil
}
