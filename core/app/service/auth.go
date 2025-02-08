package service

import (
	"crypto/hmac"
	"strconv"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	"github.com/1Panel-dev/1Panel/core/utils/jwt"
	"github.com/1Panel-dev/1Panel/core/utils/mfa"
	"github.com/gin-gonic/gin"
)

type AuthService struct{}

type IAuthService interface {
	CheckIsSafety(code string) (string, error)
	GetResponsePage() (string, error)
	VerifyCode(code string) (bool, error)
	Login(c *gin.Context, info dto.Login, entrance string) (*dto.UserLoginInfo, string, error)
	LogOut(c *gin.Context) error
	MFALogin(c *gin.Context, info dto.MFALogin, entrance string) (*dto.UserLoginInfo, string, error)
	GetSecurityEntrance() string
	IsLogin(c *gin.Context) bool
}

func NewIAuthService() IAuthService {
	return &AuthService{}
}

func (u *AuthService) Login(c *gin.Context, info dto.Login, entrance string) (*dto.UserLoginInfo, string, error) {
	nameSetting, err := settingRepo.Get(repo.WithByKey("UserName"))
	if err != nil {
		return nil, "", buserr.New("ErrRecordNotFound")
	}
	if nameSetting.Value != info.Name {
		return nil, "ErrAuth", nil
	}
	if err = checkPassword(info.Password); err != nil {
		return nil, "ErrAuth", err
	}
	entranceSetting, err := settingRepo.Get(repo.WithByKey("SecurityEntrance"))
	if err != nil {
		return nil, "", err
	}
	if len(entranceSetting.Value) != 0 && entranceSetting.Value != entrance {
		return nil, "ErrEntrance", nil
	}
	mfa, err := settingRepo.Get(repo.WithByKey("MFAStatus"))
	if err != nil {
		return nil, "", err
	}
	if err = settingRepo.Update("Language", info.Language); err != nil {
		return nil, "", err
	}
	if mfa.Value == constant.StatusEnable {
		return &dto.UserLoginInfo{Name: nameSetting.Value, MfaStatus: mfa.Value}, "", nil
	}
	res, err := u.generateSession(c, info.Name, info.AuthMethod)
	if err != nil {
		return nil, "", err
	}
	return res, "", nil
}

func (u *AuthService) MFALogin(c *gin.Context, info dto.MFALogin, entrance string) (*dto.UserLoginInfo, string, error) {
	nameSetting, err := settingRepo.Get(repo.WithByKey("UserName"))
	if err != nil {
		return nil, "", buserr.New("ErrRecordNotFound")
	}
	if nameSetting.Value != info.Name {
		return nil, "ErrAuth", nil
	}
	if err = checkPassword(info.Password); err != nil {
		return nil, "ErrAuth", err
	}
	entranceSetting, err := settingRepo.Get(repo.WithByKey("SecurityEntrance"))
	if err != nil {
		return nil, "", err
	}
	if len(entranceSetting.Value) != 0 && entranceSetting.Value != entrance {
		return nil, "", buserr.New("ErrEntrance")
	}
	mfaSecret, err := settingRepo.Get(repo.WithByKey("MFASecret"))
	if err != nil {
		return nil, "", err
	}
	mfaInterval, err := settingRepo.Get(repo.WithByKey("MFAInterval"))
	if err != nil {
		return nil, "", err
	}
	success := mfa.ValidCode(info.Code, mfaInterval.Value, mfaSecret.Value)
	if !success {
		return nil, "ErrAuth", nil
	}
	res, err := u.generateSession(c, info.Name, info.AuthMethod)
	if err != nil {
		return nil, "", err
	}
	return res, "", nil
}

func (u *AuthService) generateSession(c *gin.Context, name, authMethod string) (*dto.UserLoginInfo, error) {
	setting, err := settingRepo.Get(repo.WithByKey("SessionTimeout"))
	if err != nil {
		return nil, err
	}
	httpsSetting, err := settingRepo.Get(repo.WithByKey("SSL"))
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
			Name:    name,
			IsAgent: false,
		})
		token, err := j.CreateToken(claims)
		if err != nil {
			return nil, err
		}
		return &dto.UserLoginInfo{Name: name, Token: token}, nil
	}
	sessionUser, err := global.SESSION.Get(c)
	if err != nil {
		err := global.SESSION.Set(c, sessionUser, httpsSetting.Value == constant.StatusEnable, lifeTime)
		if err != nil {
			return nil, err
		}
		return &dto.UserLoginInfo{Name: name}, nil
	}
	if err := global.SESSION.Set(c, sessionUser, httpsSetting.Value == constant.StatusEnable, lifeTime); err != nil {
		return nil, err
	}

	return &dto.UserLoginInfo{Name: name}, nil
}

func (u *AuthService) LogOut(c *gin.Context) error {
	httpsSetting, err := settingRepo.Get(repo.WithByKey("SSL"))
	if err != nil {
		return err
	}
	sID, _ := c.Cookie(constant.SessionName)
	if sID != "" {
		c.SetCookie(constant.SessionName, sID, -1, "", "", httpsSetting.Value == constant.StatusEnable, true)
		err := global.SESSION.Delete(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *AuthService) VerifyCode(code string) (bool, error) {
	setting, err := settingRepo.Get(repo.WithByKey("SecurityEntrance"))
	if err != nil {
		return false, err
	}
	return hmac.Equal([]byte(setting.Value), []byte(code)), nil
}

func (u *AuthService) CheckIsSafety(code string) (string, error) {
	status, err := settingRepo.Get(repo.WithByKey("SecurityEntrance"))
	if err != nil {
		return "", err
	}
	if len(status.Value) == 0 {
		return "disable", nil
	}
	if status.Value == code {
		return "pass", nil
	}
	return "unpass", nil
}

func (u *AuthService) GetResponsePage() (string, error) {
	pageCode, err := settingRepo.Get(repo.WithByKey("NoAuthSetting"))
	if err != nil {
		return "", err
	}
	return pageCode.Value, nil
}

func (u *AuthService) GetSecurityEntrance() string {
	status, err := settingRepo.Get(repo.WithByKey("SecurityEntrance"))
	if err != nil {
		return ""
	}
	if len(status.Value) == 0 {
		return ""
	}
	return status.Value
}

func (u *AuthService) IsLogin(c *gin.Context) bool {
	_, err := global.SESSION.Get(c)
	return err == nil
}

func checkPassword(password string) error {
	priKey, _ := settingRepo.Get(repo.WithByKey("PASSWORD_PRIVATE_KEY"))

	privateKey, err := encrypt.ParseRSAPrivateKey(priKey.Value)
	if err != nil {
		return err
	}
	loginPassword, err := encrypt.DecryptPassword(password, privateKey)
	if err != nil {
		return err
	}
	passwordSetting, err := settingRepo.Get(repo.WithByKey("Password"))
	if err != nil {
		return err
	}
	existPassword, err := encrypt.StringDecrypt(passwordSetting.Value)
	if err != nil {
		return err
	}
	if !hmac.Equal([]byte(loginPassword), []byte(existPassword)) {
		return buserr.New("ErrAuth")
	}
	return nil
}
