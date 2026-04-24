package auth

import (
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	initauth "github.com/1Panel-dev/1Panel/core/init/auth"
	"github.com/1Panel-dev/1Panel/core/init/session/psession"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	"github.com/1Panel-dev/1Panel/core/utils/mfa"
	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context, info dto.Login, entrance string) (*dto.UserLoginInfo, string, error) {
	settingRepo := repo.NewISettingRepo()
	nameSetting, err := settingRepo.Get(repo.WithByKey("UserName"))
	if err != nil {
		return nil, "", buserr.New("ErrRecordNotFound")
	}
	if info.Name != nameSetting.Value {
		return nil, "ErrAuth", buserr.New("ErrAuth")
	}
	priKey, _ := settingRepo.Get(repo.WithByKey("PASSWORD_PRIVATE_KEY"))
	passwordSetting, err := settingRepo.Get(repo.WithByKey("Password"))
	if err != nil {
		return nil, "", err
	}
	if err = CheckPassword(priKey.Value, info.Password, passwordSetting.Value); err != nil {
		return nil, "ErrAuth", err
	}
	entranceSetting, err := settingRepo.Get(repo.WithByKey("SecurityEntrance"))
	if err != nil {
		return nil, "", err
	}
	if len(entranceSetting.Value) != 0 && entranceSetting.Value != entrance {
		return nil, "ErrEntrance", buserr.New("ErrEntrance")
	}
	mfaSetting, err := settingRepo.Get(repo.WithByKey("MFAStatus"))
	if err != nil {
		return nil, "", err
	}
	if err = settingRepo.Update("Language", info.Language); err != nil {
		return nil, "", err
	}
	if mfaSetting.Value == constant.StatusEnable {
		return BeginMFALogin(c, nameSetting.Value, entrance, mfaSetting.Value), "", nil
	}

	sessionUser := psession.SessionUser{ID: psession.SuperAdminSessionUserID, Name: nameSetting.Value, Role: "ADMIN"}
	res, err := GenerateSession(c, sessionUser)
	if err != nil {
		return nil, "", err
	}
	if entrance != "" {
		SetSecurityEntranceCookie(c, entrance)
	}
	return res, "", nil
}

func MFALogin(c *gin.Context, info dto.MFALogin, entrance string) (*dto.UserLoginInfo, string, error) {
	name, errCode, err := VerifyMFALogin(c, info.SessionID, info.Code, entrance)
	if errCode != "" {
		return nil, errCode, err
	}

	sessionUser := psession.SessionUser{ID: psession.SuperAdminSessionUserID, Name: name, Role: "ADMIN"}
	res, err := GenerateSession(c, sessionUser)
	if err != nil {
		return nil, "", err
	}
	if entrance != "" {
		SetSecurityEntranceCookie(c, entrance)
	}
	return res, "", nil
}

func BeginMFALogin(c *gin.Context, name, entrance, mfaStatus string) *dto.UserLoginInfo {
	ip := common.GetRealClientIP(c)
	mfaSession := initauth.GetMFASessionStore().Set(name, entrance, ip)
	return &dto.UserLoginInfo{Name: name, MfaStatus: mfaStatus, MfaSession: mfaSession}
}

func VerifyMFALogin(c *gin.Context, sessionID, code, entrance string) (string, string, error) {
	settingRepo := repo.NewISettingRepo()
	mfaSessions := initauth.GetMFASessionStore()
	session, ok := mfaSessions.Get(sessionID)
	if !ok {
		return "", "ErrMFA", nil
	}
	if session.IP != common.GetRealClientIP(c) {
		return "", "ErrMFA", nil
	}
	if session.Entrance != entrance {
		return "", "", buserr.New("ErrEntrance")
	}
	mfaSecret, err := settingRepo.Get(repo.WithByKey("MFASecret"))
	if err != nil {
		return "", "", err
	}
	mfaInterval, err := settingRepo.Get(repo.WithByKey("MFAInterval"))
	if err != nil {
		return "", "", err
	}
	if !mfa.ValidCode(code, mfaInterval.Value, mfaSecret.Value) {
		return "", "ErrMFA", nil
	}
	mfaSessions.Delete(sessionID)
	return session.Name, "", nil
}

func GenerateSession(c *gin.Context, sessionUser psession.SessionUser) (*dto.UserLoginInfo, error) {
	settingRepo := repo.NewISettingRepo()
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

	if err := global.SESSION.SetFresh(c, sessionUser, httpsSetting.Value == constant.StatusEnable, lifeTime); err != nil {
		return nil, err
	}

	return &dto.UserLoginInfo{Name: sessionUser.Name, Role: sessionUser.Role}, nil
}

func SetSecurityEntranceCookie(c *gin.Context, entrance string) {
	settingRepo := repo.NewISettingRepo()
	entranceValue := base64.StdEncoding.EncodeToString([]byte(entrance))
	sslEnabled := false
	if setting, err := settingRepo.Get(repo.WithByKey("SSL")); err == nil {
		sslEnabled = setting.Value == constant.StatusEnable
	}
	c.SetCookie("SecurityEntrance", entranceValue, 0, "/", "", sslEnabled, true)
}

func CheckEntrance(entrance string) error {
	settingRepo := repo.NewISettingRepo()
	entranceSetting, err := settingRepo.Get(repo.WithByKey("SecurityEntrance"))
	if err != nil {
		return err
	}
	if len(entranceSetting.Value) != 0 && entranceSetting.Value != entrance {
		return buserr.New("ErrEntrance")
	}
	return nil
}

func CheckPassword(priKey, password, passwordFromDB string) error {
	privateKey, err := encrypt.ParseRSAPrivateKey(priKey)
	if err != nil {
		return err
	}
	loginPassword, err := encrypt.DecryptPassword(password, privateKey)
	if err != nil {
		return err
	}
	existPassword, err := encrypt.StringDecrypt(passwordFromDB)
	if err != nil {
		return err
	}
	if !hmac.Equal([]byte(loginPassword), []byte(existPassword)) {
		return buserr.New("ErrAuth")
	}
	return nil
}

func LoadMFA(req dto.MfaRequest) (mfa.Otp, error) {
	settingRepo := repo.NewISettingRepo()
	username, err := settingRepo.GetValueByKey("UserName")
	if err != nil {
		return mfa.Otp{}, err
	}
	otp, err := mfa.GetOtp(username, req.Title, req.Interval)
	if err != nil {
		return mfa.Otp{}, err
	}
	return otp, nil
}
func MFABind(req dto.MfaCredential) error {
	success := mfa.ValidCode(req.Code, req.Interval, req.Secret)
	if !success {
		return errors.New("code is not valid")
	}

	settingRepo := repo.NewISettingRepo()
	if err := settingRepo.Update("MFAInterval", req.Interval); err != nil {
		return err
	}
	if err := settingRepo.Update("MFAStatus", constant.StatusEnable); err != nil {
		return err
	}
	if err := settingRepo.Update("MFASecret", req.Secret); err != nil {
		return err
	}
	return nil
}
func GetCurrentUserInfo() (*dto.CurrentUserInfo, error) {
	setting, err := repo.NewISettingRepo().List()
	if err != nil {
		return nil, buserr.New("ErrRecordNotFound")
	}
	settingMap := make(map[string]string)
	for _, set := range setting {
		settingMap[set.Key] = set.Value
	}
	var info dto.CurrentUserInfo
	stringSettingMap := make(map[string]string, len(settingMap))
	for key, value := range settingMap {
		stringSettingMap[key] = value
	}
	delete(stringSettingMap, "SessionTimeout")
	delete(stringSettingMap, "ExpirationDays")
	arr, err := json.Marshal(stringSettingMap)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(arr, &info); err != nil {
		return nil, err
	}
	info.SessionTimeout, _ = strconv.Atoi(settingMap["SessionTimeout"])
	info.ExpirationDays, _ = strconv.Atoi(settingMap["ExpirationDays"])
	info.Name = settingMap["UserName"]
	return &info, nil
}
func UpdateCurrentUserInfo(c *gin.Context, req dto.CurrentUserUpdate) error {
	settingRepo := repo.NewISettingRepo()
	if len(req.Password) != 0 {
		if len(req.OldPassword) == 0 {
			return buserr.New("ErrInitialPassword")
		}
		oldPassword, err := base64.StdEncoding.DecodeString(req.OldPassword)
		if err != nil {
			return err
		}
		newPassword, err := base64.StdEncoding.DecodeString(req.Password)
		if err != nil {
			return err
		}
		if err := HandlePasswordExpired(c, string(oldPassword), string(newPassword)); err != nil {
			return err
		}
	}
	if err := settingRepo.Update("UserName", req.Name); err != nil {
		return err
	}
	if err := settingRepo.Update("SessionTimeout", strconv.Itoa(req.SessionTimeout)); err != nil {
		return err
	}
	if err := settingRepo.Update("ExpirationDays", strconv.Itoa(req.ExpirationDays)); err != nil {
		return err
	}

	expirationTime := ""
	if req.ExpirationDays != 0 {
		expirationTime = time.Now().AddDate(0, 0, req.ExpirationDays).Format(constant.DateTimeLayout)
	}
	if err := settingRepo.Update("ExpirationTime", expirationTime); err != nil {
		return err
	}
	deleteCurrentSession(c)
	return nil
}

func GenerateApiKey() (string, error) {
	apiKey := common.RandStr(32)
	if err := repo.NewISettingRepo().Update("ApiKey", apiKey); err != nil {
		return "", err
	}
	return apiKey, nil
}
func UpdateApiConfig(req dto.ApiInterfaceConfig) error {
	settingRepo := repo.NewISettingRepo()
	if err := settingRepo.UpdateOrCreate("ApiInterfaceStatus", req.ApiInterfaceStatus); err != nil {
		return err
	}
	if err := settingRepo.UpdateOrCreate("ApiKey", req.ApiKey); err != nil {
		return err
	}
	if err := settingRepo.UpdateOrCreate("IpWhiteList", req.IpWhiteList); err != nil {
		return err
	}
	if err := settingRepo.UpdateOrCreate("ApiKeyValidityTime", req.ApiKeyValidityTime); err != nil {
		return err
	}
	return nil
}

func HandlePasswordExpired(c *gin.Context, old, new string) error {
	settingRepo := repo.NewISettingRepo()
	setting, err := settingRepo.Get(repo.WithByKey("Password"))
	if err != nil {
		return err
	}
	passwordFromDB, err := encrypt.StringDecrypt(setting.Value)
	if err != nil {
		return err
	}
	if passwordFromDB == old {
		newPassword, err := encrypt.StringEncrypt(new)
		if err != nil {
			return err
		}
		if err := settingRepo.Update("Password", newPassword); err != nil {
			return err
		}

		expiredSetting, err := settingRepo.Get(repo.WithByKey("ExpirationDays"))
		if err != nil {
			return err
		}
		timeout, _ := strconv.Atoi(expiredSetting.Value)
		if err := settingRepo.Update("ExpirationTime", time.Now().AddDate(0, 0, timeout).Format(constant.DateTimeLayout)); err != nil {
			return err
		}
		return nil
	}
	return buserr.New("ErrInitialPassword")
}

func LoadSessionTimeout(sessionUser psession.SessionUser) (int, error) {
	settingRepo := repo.NewISettingRepo()
	sessionTimeout, err := settingRepo.GetValueByKey("SessionTimeout")
	if err != nil {
		return 0, err
	}
	lifeTime, _ := strconv.Atoi(sessionTimeout)
	return lifeTime, nil
}
func LoadExpired(sessionUser psession.SessionUser) (bool, time.Time, error) {
	settingRepo := repo.NewISettingRepo()
	expirationDays, err := settingRepo.GetValueByKey("ExpirationDays")
	if err != nil {
		return true, time.Time{}, err
	}
	expiredDays, _ := strconv.Atoi(expirationDays)
	if expiredDays == 0 {
		return false, time.Time{}, nil
	}

	expirationTime, err := settingRepo.GetValueByKey("ExpirationTime")
	if err != nil {
		return true, time.Time{}, err
	}
	expiredTime, err := time.ParseInLocation(constant.DateTimeLayout, expirationTime, common.LoadExpiredLocation())
	if err != nil {
		return true, time.Time{}, err
	}
	return true, expiredTime, nil
}

func deleteCurrentSession(c *gin.Context) {
	if c == nil {
		return
	}
	sessionUser, err := global.SESSION.Get(c)
	if err != nil || sessionUser.ID == "" {
		return
	}
	_ = global.SESSION.DeleteByID(sessionUser.ID)
}
