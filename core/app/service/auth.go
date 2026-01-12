package service

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	"github.com/1Panel-dev/1Panel/core/utils/mfa"
	"github.com/1Panel-dev/1Panel/core/utils/passkey"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

type AuthService struct{}

type IAuthService interface {
	GetResponsePage() (string, error)
	VerifyCode(code string) (bool, error)
	Login(c *gin.Context, info dto.Login, entrance string) (*dto.UserLoginInfo, string, error)
	LogOut(c *gin.Context) error
	MFALogin(c *gin.Context, info dto.MFALogin, entrance string) (*dto.UserLoginInfo, string, error)
	PasskeyBeginLogin(c *gin.Context, entrance string) (*dto.PasskeyBeginResponse, string, error)
	PasskeyFinishLogin(c *gin.Context, sessionID, entrance string) (*dto.UserLoginInfo, string, error)
	PasskeyBeginRegister(c *gin.Context, name string) (*dto.PasskeyBeginResponse, string, error)
	PasskeyFinishRegister(c *gin.Context, sessionID string) (string, error)
	PasskeyList() ([]dto.PasskeyInfo, error)
	PasskeyDelete(id string) error
	PasskeyStatus(c *gin.Context) bool
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
		return nil, "ErrAuth", buserr.New("ErrAuth")
	}
	if err = checkPassword(info.Password); err != nil {
		return nil, "ErrAuth", err
	}
	entranceSetting, err := settingRepo.Get(repo.WithByKey("SecurityEntrance"))
	if err != nil {
		return nil, "", err
	}
	if len(entranceSetting.Value) != 0 && entranceSetting.Value != entrance {
		return nil, "ErrEntrance", buserr.New("ErrEntrance")
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
	res, err := u.generateSession(c, info.Name)
	if err != nil {
		return nil, "", err
	}
	if entrance != "" {
		entranceValue := base64.StdEncoding.EncodeToString([]byte(entrance))
		c.SetCookie("SecurityEntrance", entranceValue, 0, "", "", false, true)
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
	res, err := u.generateSession(c, info.Name)
	if err != nil {
		return nil, "", err
	}
	if entrance != "" {
		entranceValue := base64.StdEncoding.EncodeToString([]byte(entrance))
		c.SetCookie("SecurityEntrance", entranceValue, 0, "", "", false, true)
	}
	return res, "", nil
}

func (u *AuthService) generateSession(c *gin.Context, name string) (*dto.UserLoginInfo, error) {
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

func (u *AuthService) PasskeyStatus(c *gin.Context) bool {
	enabled, err := u.passkeyEnabled(c)
	if err != nil {
		global.LOG.Errorf("passkey enabled check failed, err: %v", err)
		enabled = false
	}
	configured, err := u.passkeyConfigured()
	if err != nil {
		global.LOG.Errorf("passkey config check failed, err: %v", err)
		configured = false
	}
	return enabled && configured
}

func (u *AuthService) PasskeyBeginLogin(c *gin.Context, entrance string) (*dto.PasskeyBeginResponse, string, error) {
	if err := u.checkEntrance(entrance); err != nil {
		return nil, "ErrEntrance", err
	}

	config, msgKey, err := u.passkeyConfig(c)
	if err != nil {
		return nil, msgKey, err
	}

	records, err := loadPasskeyCredentialRecords()
	if err != nil {
		return nil, "", err
	}
	if len(records) == 0 {
		return nil, "ErrPasskeyNotConfigured", buserr.New("ErrPasskeyNotConfigured")
	}

	user, err := u.passkeyUser(records, true)
	if err != nil {
		return nil, "", err
	}

	wa, err := webauthn.New(config)
	if err != nil {
		return nil, "", err
	}
	assertion, sessionData, err := wa.BeginLogin(user)
	if err != nil {
		return nil, "", err
	}
	passkeySessions := passkey.GetPasskeySessionStore()
	sessionID := passkeySessions.Set(passkey.PasskeySessionKindLogin, "", *sessionData)
	return &dto.PasskeyBeginResponse{SessionID: sessionID, PublicKey: assertion.Response}, "", nil
}

func (u *AuthService) PasskeyFinishLogin(c *gin.Context, sessionID, entrance string) (*dto.UserLoginInfo, string, error) {
	if sessionID == "" {
		return nil, "ErrPasskeySession", buserr.New("ErrPasskeySession")
	}

	if err := u.checkEntrance(entrance); err != nil {
		return nil, "ErrEntrance", err
	}

	config, msgKey, err := u.passkeyConfig(c)
	if err != nil {
		return nil, msgKey, err
	}

	passkeySessions := passkey.GetPasskeySessionStore()
	session, ok := passkeySessions.Get(sessionID)
	if !ok || session.Kind != passkey.PasskeySessionKindLogin {
		return nil, "ErrPasskeySession", buserr.New("ErrPasskeySession")
	}
	passkeySessions.Delete(sessionID)

	records, err := loadPasskeyCredentialRecords()
	if err != nil {
		return nil, "", err
	}
	if len(records) == 0 {
		return nil, "ErrPasskeyNotConfigured", buserr.New("ErrPasskeyNotConfigured")
	}

	user, err := u.passkeyUser(records, true)
	if err != nil {
		return nil, "", err
	}

	wa, err := webauthn.New(config)
	if err != nil {
		return nil, "", err
	}
	credential, err := wa.FinishLogin(user, session.Session, c.Request)
	if err != nil {
		return nil, "ErrAuth", err
	}

	if err := updatePasskeyCredentialRecord(records, credential); err != nil {
		return nil, "ErrAuth", err
	}
	if err := savePasskeyCredentialRecords(records); err != nil {
		return nil, "", err
	}

	userSetting, err := settingRepo.Get(repo.WithByKey("UserName"))
	if err != nil {
		return nil, "", err
	}
	res, err := u.generateSession(c, userSetting.Value)
	if err != nil {
		return nil, "", err
	}
	if entrance != "" {
		entranceValue := base64.StdEncoding.EncodeToString([]byte(entrance))
		c.SetCookie("SecurityEntrance", entranceValue, 0, "", "", false, true)
	}
	return res, "", nil
}

func (u *AuthService) PasskeyBeginRegister(c *gin.Context, name string) (*dto.PasskeyBeginResponse, string, error) {
	config, msgKey, err := u.passkeyConfig(c)
	if err != nil {
		return nil, msgKey, err
	}
	records, err := loadPasskeyCredentialRecords()
	if err != nil {
		return nil, "", err
	}
	if len(records) >= passkey.PasskeyMaxCredentials {
		return nil, "ErrPasskeyLimit", buserr.New("ErrPasskeyLimit")
	}
	user, err := u.passkeyUser(records, true)
	if err != nil {
		return nil, "", err
	}

	wa, err := webauthn.New(config)
	if err != nil {
		return nil, "", err
	}
	exclusions := make([]protocol.CredentialDescriptor, len(user.Credentials))
	for i, credential := range user.Credentials {
		exclusions[i] = credential.Descriptor()
	}
	creation, sessionData, err := wa.BeginRegistration(user, webauthn.WithExclusions(exclusions))
	if err != nil {
		return nil, "", err
	}

	passkeySessions := passkey.GetPasskeySessionStore()
	sessionID := passkeySessions.Set(passkey.PasskeySessionKindRegister, strings.TrimSpace(name), *sessionData)
	return &dto.PasskeyBeginResponse{SessionID: sessionID, PublicKey: creation.Response}, "", nil
}

func (u *AuthService) PasskeyFinishRegister(c *gin.Context, sessionID string) (string, error) {
	if sessionID == "" {
		return "ErrPasskeySession", buserr.New("ErrPasskeySession")
	}
	config, msgKey, err := u.passkeyConfig(c)
	if err != nil {
		return msgKey, err
	}

	passkeySessions := passkey.GetPasskeySessionStore()
	session, ok := passkeySessions.Get(sessionID)
	if !ok || session.Kind != passkey.PasskeySessionKindRegister {
		return "ErrPasskeySession", buserr.New("ErrPasskeySession")
	}

	passkeySessions.Delete(sessionID)

	records, err := loadPasskeyCredentialRecords()
	if err != nil {
		return "", err
	}
	if len(records) >= passkey.PasskeyMaxCredentials {
		return "ErrPasskeyLimit", buserr.New("ErrPasskeyLimit")
	}

	user, err := u.passkeyUser(records, true)
	if err != nil {
		return "", err
	}

	wa, err := webauthn.New(config)
	if err != nil {
		return "", err
	}
	credential, err := wa.FinishRegistration(user, session.Session, c.Request)
	if err != nil {
		return "ErrPasskeyVerify", err
	}

	if passkeyCredentialExists(records, credential.ID) {
		return "ErrPasskeyDuplicate", buserr.New("ErrPasskeyDuplicate")
	}

	displayName := strings.TrimSpace(session.Name)
	if displayName == "" {
		displayName = fmt.Sprintf("%s-%s", passkey.PasskeyCredentialNameDefault, time.Now().Format("20060102150405"))
	}

	records = append(records, passkey.PasskeyCredentialRecord{
		ID:         base64.RawURLEncoding.EncodeToString(credential.ID),
		Name:       displayName,
		CreatedAt:  time.Now().Format(constant.DateTimeLayout),
		LastUsedAt: "",
		FlagsValue: credentialFlagsValue(credential.Flags),
		Credential: *credential,
	})

	if err := savePasskeyCredentialRecords(records); err != nil {
		return "", err
	}
	return "", nil
}

func (u *AuthService) PasskeyList() ([]dto.PasskeyInfo, error) {
	records, err := loadPasskeyCredentialRecords()
	if err != nil {
		return nil, err
	}
	list := make([]dto.PasskeyInfo, 0, len(records))
	for _, record := range records {
		list = append(list, dto.PasskeyInfo{
			ID:         record.ID,
			Name:       record.Name,
			CreatedAt:  record.CreatedAt,
			LastUsedAt: record.LastUsedAt,
		})
	}
	return list, nil
}

func (u *AuthService) PasskeyDelete(id string) error {
	records, err := loadPasskeyCredentialRecords()
	if err != nil {
		return err
	}
	index := -1
	for i, record := range records {
		if record.ID == id {
			index = i
			break
		}
	}
	if index == -1 {
		return buserr.New("ErrRecordNotFound")
	}
	records = append(records[:index], records[index+1:]...)
	return savePasskeyCredentialRecords(records)
}

func (u *AuthService) passkeyEnabled(c *gin.Context) (bool, error) {
	sslSetting, err := settingRepo.Get(repo.WithByKey("SSL"))
	if err != nil {
		return false, err
	}
	if sslSetting.Value == constant.StatusDisable {
		return false, nil
	}
	return strings.EqualFold(passkeyRequestScheme(c), "https"), nil
}

func (u *AuthService) passkeyConfigured() (bool, error) {
	bindDomain, err := settingRepo.Get(repo.WithByKey("BindDomain"))
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(bindDomain.Value) == "" {
		return false, nil
	}
	records, err := loadPasskeyCredentialRecords()
	if err != nil {
		return false, err
	}
	return len(records) > 0, nil
}

func (u *AuthService) passkeyUser(records []passkey.PasskeyCredentialRecord, allowCreate bool) (*passkey.PasskeyUser, error) {
	userID, err := u.passkeyUserID(allowCreate)
	if err != nil {
		return nil, err
	}
	nameSetting, err := settingRepo.Get(repo.WithByKey("UserName"))
	if err != nil {
		return nil, err
	}
	credentials := make([]webauthn.Credential, len(records))
	for i, record := range records {
		credentials[i] = record.Credential
	}
	return &passkey.PasskeyUser{
		ID:          userID,
		Name:        nameSetting.Value,
		DisplayName: nameSetting.Value,
		Credentials: credentials,
	}, nil
}

func (u *AuthService) passkeyUserID(allowCreate bool) ([]byte, error) {
	setting, err := settingRepo.Get(repo.WithByKey(passkey.PasskeyUserIDSettingKey))
	if err != nil {
		return nil, err
	}
	if setting.Value == "" {
		if !allowCreate {
			return nil, buserr.New("ErrPasskeyNotConfigured")
		}
		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			return nil, err
		}
		encoded := base64.RawURLEncoding.EncodeToString(raw)
		if err := settingRepo.Update(passkey.PasskeyUserIDSettingKey, encoded); err != nil {
			return nil, err
		}
		return raw, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(setting.Value)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (u *AuthService) passkeyConfig(c *gin.Context) (*webauthn.Config, string, error) {
	enabled, err := u.passkeyEnabled(c)
	if err != nil {
		return nil, "", err
	}
	if !enabled {
		return nil, "ErrPasskeyDisabled", buserr.New("ErrPasskeyDisabled")
	}
	origin, rpID, err := passkeyOriginAndRPID(c)
	if err != nil {
		return nil, "", err
	}
	panelName, err := settingRepo.Get(repo.WithByKey("PanelName"))
	if err != nil {
		return nil, "", err
	}
	return &webauthn.Config{
		RPID:          rpID,
		RPDisplayName: panelName.Value,
		RPOrigins:     []string{origin},
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			UserVerification: protocol.VerificationRequired,
		},
	}, "", nil
}

func (u *AuthService) checkEntrance(entrance string) error {
	entranceSetting, err := settingRepo.Get(repo.WithByKey("SecurityEntrance"))
	if err != nil {
		return err
	}
	if len(entranceSetting.Value) != 0 && entranceSetting.Value != entrance {
		return buserr.New("ErrEntrance")
	}
	return nil
}

func loadPasskeyCredentialRecords() ([]passkey.PasskeyCredentialRecord, error) {
	setting, err := settingRepo.Get(repo.WithByKey(passkey.PasskeyCredentialSettingKey))
	if err != nil {
		return nil, err
	}
	if setting.Value == "" {
		return []passkey.PasskeyCredentialRecord{}, nil
	}
	decrypted, err := encrypt.StringDecrypt(setting.Value)
	if err != nil {
		return nil, err
	}
	var records []passkey.PasskeyCredentialRecord
	if err := json.Unmarshal([]byte(decrypted), &records); err != nil {
		return nil, err
	}
	for i := range records {
		records[i].Credential.Flags = webauthn.NewCredentialFlags(protocol.AuthenticatorFlags(records[i].FlagsValue))
	}
	return records, nil
}

func savePasskeyCredentialRecords(records []passkey.PasskeyCredentialRecord) error {
	if len(records) == 0 {
		return settingRepo.Update(passkey.PasskeyCredentialSettingKey, "")
	}
	for i := range records {
		records[i].FlagsValue = credentialFlagsValue(records[i].Credential.Flags)
	}
	raw, err := json.Marshal(records)
	if err != nil {
		return err
	}
	encrypted, err := encrypt.StringEncrypt(string(raw))
	if err != nil {
		return err
	}
	return settingRepo.Update(passkey.PasskeyCredentialSettingKey, encrypted)
}

func passkeyCredentialExists(records []passkey.PasskeyCredentialRecord, credentialID []byte) bool {
	encoded := base64.RawURLEncoding.EncodeToString(credentialID)
	for _, record := range records {
		if record.ID == encoded {
			return true
		}
	}
	return false
}

func updatePasskeyCredentialRecord(records []passkey.PasskeyCredentialRecord, credential *webauthn.Credential) error {
	encoded := base64.RawURLEncoding.EncodeToString(credential.ID)
	for i := range records {
		if records[i].ID == encoded {
			records[i].Credential = *credential
			records[i].FlagsValue = credentialFlagsValue(credential.Flags)
			records[i].LastUsedAt = time.Now().Format(constant.DateTimeLayout)
			return nil
		}
	}
	return buserr.New("ErrPasskeyNotConfigured")
}

func credentialFlagsValue(flags webauthn.CredentialFlags) uint8 {
	var value protocol.AuthenticatorFlags
	if flags.UserPresent {
		value |= protocol.FlagUserPresent
	}
	if flags.UserVerified {
		value |= protocol.FlagUserVerified
	}
	if flags.BackupEligible {
		value |= protocol.FlagBackupEligible
	}
	if flags.BackupState {
		value |= protocol.FlagBackupState
	}
	return uint8(value)
}

func passkeyOriginAndRPID(c *gin.Context) (string, string, error) {
	host := passkeyRequestHost(c)
	if host == "" {
		return "", "", fmt.Errorf("missing request host")
	}
	scheme := passkeyRequestScheme(c)
	origin := fmt.Sprintf("%s://%s", scheme, host)

	bindDomain, err := settingRepo.Get(repo.WithByKey("BindDomain"))
	if err != nil {
		return "", "", err
	}
	bindDomainValue := strings.TrimSpace(bindDomain.Value)
	if bindDomainValue == "" {
		return "", "", buserr.New("ErrPasskeyNotConfigured")
	}
	hostDomain := stripHostPort(host)
	bindDomainValue = stripHostPort(bindDomainValue)
	if hostDomain == "" || !strings.EqualFold(hostDomain, bindDomainValue) {
		return "", "", buserr.New("ErrPasskeyDisabled")
	}
	return origin, bindDomainValue, nil
}

func passkeyRequestHost(c *gin.Context) string {
	host := c.Request.Host
	if strings.Contains(host, ",") {
		host = strings.TrimSpace(strings.Split(host, ",")[0])
	}
	return strings.TrimSpace(host)
}

func passkeyRequestScheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}
	return "http"
}

func stripHostPort(hostport string) string {
	if hostport == "" {
		return hostport
	}
	hostport = strings.TrimSpace(hostport)
	if host, _, err := net.SplitHostPort(hostport); err == nil {
		return strings.Trim(host, "[]")
	}
	return strings.Trim(hostport, "[]")
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
