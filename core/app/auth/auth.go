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
	if !global.CONF.Base.IsDemo {
		if err = settingRepo.Update("Language", info.Language); err != nil {
			return nil, "", err
		}
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

func BeginAuthSourceMFALogin(
	c *gin.Context,
	name, entrance, mfaStatus, authSource string,
	authSourceID uint,
	authSourceConfigVersion uint64,
) *dto.UserLoginInfo {
	ip := common.GetRealClientIP(c)
	mfaSession := initauth.GetMFASessionStore().SetWithAuthSource(
		name,
		entrance,
		ip,
		authSource,
		authSourceID,
		authSourceConfigVersion,
	)
	return &dto.UserLoginInfo{Name: name, MfaStatus: mfaStatus, MfaSession: mfaSession}
}

func BeginAuthSourceMFALoginWithSession(
	c *gin.Context,
	name, entrance, mfaStatus, authSource string,
	authSourceID uint,
	authSourceConfigVersion uint64,
	externalIssuer, externalNameID, externalNameIDFormat, externalSessionIndex string,
	externalSessionExpiresAt time.Time,
	externalSessionRequired bool,
) *dto.UserLoginInfo {
	ip := common.GetRealClientIP(c)
	mfaSession := initauth.GetMFASessionStore().SetWithAuthSourceSession(
		name, entrance, ip, authSource, authSourceID, authSourceConfigVersion,
		externalIssuer, externalNameID, externalNameIDFormat, externalSessionIndex,
		externalSessionExpiresAt,
		externalSessionRequired,
	)
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
	mfaSecret, err := settingRepo.GetValueByKey("MFASecret")
	if err != nil {
		return "", "", err
	}
	mfaInterval, err := settingRepo.GetValueByKey("MFAInterval")
	if err != nil {
		return "", "", err
	}
	interval, err := strconv.Atoi(mfaInterval)
	if err != nil {
		return "", "", err
	}
	if !mfa.ValidCode(interval, code, mfaSecret) {
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
	loginPassword, err := DecryptLoginPassword(priKey, password)
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

func DecryptLoginPassword(priKey, password string) (string, error) {
	privateKey, err := encrypt.ParseRSAPrivateKey(priKey)
	if err != nil {
		return "", err
	}
	loginPassword, err := encrypt.DecryptPassword(password, privateKey)
	if err != nil {
		return "", err
	}
	return loginPassword, nil
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
	success := mfa.ValidCode(req.Interval, req.Code, req.Secret)
	if !success {
		return errors.New("code is not valid")
	}

	settingRepo := repo.NewISettingRepo()
	if err := settingRepo.Update("MFAInterval", strconv.Itoa(req.Interval)); err != nil {
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

func MFAClose() error {
	return repo.NewISettingRepo().Update("MFAStatus", constant.StatusDisable)
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
	delete(stringSettingMap, "MFAInterval")
	delete(stringSettingMap, "ApiKeyValidityTime")
	arr, err := json.Marshal(stringSettingMap)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(arr, &info); err != nil {
		return nil, err
	}
	info.MFAInterval, _ = strconv.Atoi(settingMap["MFAInterval"])
	info.ApiKeyValidityTime, _ = strconv.Atoi(settingMap["ApiKeyValidityTime"])
	info.Name = settingMap["UserName"]
	info.AuthSource = "local"
	info.AuthSourceStatus = "active"
	info.Role = "ADMIN"
	info.Permissions = []string{}
	info.NodeRoles = []dto.CurrentUserNodeRole{}
	return &info, nil
}
func LoadPasswordExpirationTime(_ *gin.Context) (string, error) {
	return repo.NewISettingRepo().GetValueByKey("ExpirationTime")
}
func SyncPasswordExpirationTime(expirationDays string) error {
	expiredDays, _ := strconv.Atoi(expirationDays)
	return repo.NewISettingRepo().Update("ExpirationTime", buildPasswordExpirationTime(expiredDays))
}
func UpdateCurrentUserInfo(c *gin.Context, req dto.CurrentUserUpdate) error {
	settingRepo := repo.NewISettingRepo()
	currentName, err := settingRepo.GetValueByKey("UserName")
	if err != nil {
		return err
	}
	shouldDeleteSession := len(req.Password) != 0 || req.Name != currentName
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
	if shouldDeleteSession {
		deleteCurrentSession(c)
	}
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
	trustedProxies, err := NormalizeAPITrustedProxies(req.ApiTrustedProxies)
	if err != nil {
		return err
	}
	if err := settingRepo.UpdateOrCreate("ApiInterfaceStatus", req.ApiInterfaceStatus); err != nil {
		return err
	}
	if err := settingRepo.UpdateOrCreate("ApiKey", req.ApiKey); err != nil {
		return err
	}
	if err := settingRepo.UpdateOrCreate("IpWhiteList", req.IpWhiteList); err != nil {
		return err
	}
	if err := settingRepo.UpdateOrCreate("ApiTrustedProxies", trustedProxies); err != nil {
		return err
	}
	if err := settingRepo.UpdateOrCreate("ApiKeyValidityTime", strconv.Itoa(req.ApiKeyValidityTime)); err != nil {
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
		if err := settingRepo.Update("ExpirationTime", buildPasswordExpirationTime(timeout)); err != nil {
			return err
		}
		return nil
	}
	return buserr.New("ErrInitialPassword")
}

func buildPasswordExpirationTime(expirationDays int) string {
	if expirationDays == 0 {
		return ""
	}
	return time.Now().AddDate(0, 0, expirationDays).Format(constant.DateTimeLayout)
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
