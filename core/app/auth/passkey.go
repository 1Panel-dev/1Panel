package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/init/session/psession"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	"github.com/1Panel-dev/1Panel/core/utils/passkey"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

func EvaluatePasskeyStatus(c *gin.Context, configured func() (bool, error)) bool {
	enabled, err := PasskeyEnabled(c)
	if err != nil {
		global.LOG.Errorf("passkey enabled check failed, err: %v", err)
		enabled = false
	}
	configuredOK, err := configured()
	if err != nil {
		global.LOG.Errorf("passkey config check failed, err: %v", err)
		configuredOK = false
	}
	return enabled && configuredOK
}

func PasskeyStatus(c *gin.Context) bool {
	return EvaluatePasskeyStatus(c, communityPasskeyConfigured)
}

func PasskeyBeginLogin(c *gin.Context, entrance string) (*dto.PasskeyBeginResponse, string, error) {
	if err := CheckEntrance(entrance); err != nil {
		return nil, "ErrEntrance", err
	}
	config, msgKey, err := PasskeyConfig(c)
	if err != nil {
		return nil, msgKey, err
	}
	records, err := loadCommunityPasskeyCredentialRecords()
	if err != nil {
		return nil, "", err
	}
	if len(records) == 0 {
		return nil, "ErrPasskeyNotConfigured", buserr.New("ErrPasskeyNotConfigured")
	}
	user, err := communityPasskeyUser(records, true)
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
	sessionID := passkey.GetPasskeySessionStore().Set(passkey.PasskeySessionKindLogin, "", *sessionData)
	return &dto.PasskeyBeginResponse{SessionID: sessionID, PublicKey: assertion.Response}, "", nil
}

func PasskeyFinishLogin(c *gin.Context, sessionID, entrance string) (*dto.UserLoginInfo, string, error) {
	if sessionID == "" {
		return nil, "ErrPasskeySession", buserr.New("ErrPasskeySession")
	}
	if err := CheckEntrance(entrance); err != nil {
		return nil, "ErrEntrance", err
	}
	config, msgKey, err := PasskeyConfig(c)
	if err != nil {
		return nil, msgKey, err
	}
	sessionStore := passkey.GetPasskeySessionStore()
	session, ok := sessionStore.Get(sessionID)
	if !ok || session.Kind != passkey.PasskeySessionKindLogin {
		return nil, "ErrPasskeySession", buserr.New("ErrPasskeySession")
	}
	sessionStore.Delete(sessionID)
	records, err := loadCommunityPasskeyCredentialRecords()
	if err != nil {
		return nil, "", err
	}
	if len(records) == 0 {
		return nil, "ErrPasskeyNotConfigured", buserr.New("ErrPasskeyNotConfigured")
	}
	user, err := communityPasskeyUser(records, true)
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
	if err := UpdatePasskeyCredentialRecord(records, credential); err != nil {
		return nil, "ErrAuth", err
	}
	if err := saveCommunityPasskeyCredentialRecords(records); err != nil {
		return nil, "", err
	}
	userSetting, err := repo.NewISettingRepo().Get(repo.WithByKey("UserName"))
	if err != nil {
		return nil, "", err
	}
	res, err := GenerateSession(c, psession.SessionUser{ID: psession.SuperAdminSessionUserID, Name: userSetting.Value, Role: "ADMIN"})
	if err != nil {
		return nil, "", err
	}
	if entrance != "" {
		SetSecurityEntranceCookie(c, entrance)
	}
	return res, "", nil
}

func PasskeyBeginRegister(c *gin.Context, name string) (*dto.PasskeyBeginResponse, string, error) {
	config, msgKey, err := PasskeyConfig(c)
	if err != nil {
		return nil, msgKey, err
	}
	records, err := loadCommunityPasskeyCredentialRecords()
	if err != nil {
		return nil, "", err
	}
	if len(records) >= passkey.PasskeyMaxCredentials {
		return nil, "ErrPasskeyLimit", buserr.New("ErrPasskeyLimit")
	}
	user, err := communityPasskeyUser(records, true)
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
	sessionID := passkey.GetPasskeySessionStore().Set(passkey.PasskeySessionKindRegister, strings.TrimSpace(name), *sessionData)
	return &dto.PasskeyBeginResponse{SessionID: sessionID, PublicKey: creation.Response}, "", nil
}

func PasskeyFinishRegister(c *gin.Context, sessionID string) (string, error) {
	if sessionID == "" {
		return "ErrPasskeySession", buserr.New("ErrPasskeySession")
	}
	config, msgKey, err := PasskeyConfig(c)
	if err != nil {
		return msgKey, err
	}
	sessionStore := passkey.GetPasskeySessionStore()
	session, ok := sessionStore.Get(sessionID)
	if !ok || session.Kind != passkey.PasskeySessionKindRegister {
		return "ErrPasskeySession", buserr.New("ErrPasskeySession")
	}
	sessionStore.Delete(sessionID)
	records, err := loadCommunityPasskeyCredentialRecords()
	if err != nil {
		return "", err
	}
	if len(records) >= passkey.PasskeyMaxCredentials {
		return "ErrPasskeyLimit", buserr.New("ErrPasskeyLimit")
	}
	user, err := communityPasskeyUser(records, true)
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
	if PasskeyCredentialExists(records, credential.ID) {
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
		FlagsValue: CredentialFlagsValue(credential.Flags),
		Credential: *credential,
	})
	if err := saveCommunityPasskeyCredentialRecords(records); err != nil {
		return "", err
	}
	return "", nil
}

func PasskeyList() ([]dto.PasskeyInfo, error) {
	records, err := loadCommunityPasskeyCredentialRecords()
	if err != nil {
		return nil, err
	}
	list := make([]dto.PasskeyInfo, 0, len(records))
	for _, record := range records {
		list = append(list, dto.PasskeyInfo{ID: record.ID, Name: record.Name, CreatedAt: record.CreatedAt, LastUsedAt: record.LastUsedAt})
	}
	return list, nil
}

func PasskeyDelete(id string) error {
	records, err := loadCommunityPasskeyCredentialRecords()
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
	return saveCommunityPasskeyCredentialRecords(records)
}

func ClearPasskeys() error {
	settingRepo := repo.NewISettingRepo()
	if err := settingRepo.Update(passkey.PasskeyUserIDSettingKey, ""); err != nil {
		return err
	}
	return settingRepo.Update(passkey.PasskeyCredentialSettingKey, "")
}

func communityPasskeyConfigured() (bool, error) {
	bindDomain, err := repo.NewISettingRepo().Get(repo.WithByKey("BindDomain"))
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(bindDomain.Value) == "" {
		return false, nil
	}
	records, err := loadCommunityPasskeyCredentialRecords()
	if err != nil {
		return false, err
	}
	return len(records) > 0, nil
}

func communityPasskeyUser(records []passkey.PasskeyCredentialRecord, allowCreate bool) (*passkey.PasskeyUser, error) {
	settingRepo := repo.NewISettingRepo()
	storedUserID, err := settingRepo.Get(repo.WithByKey(passkey.PasskeyUserIDSettingKey))
	if err != nil {
		return nil, err
	}
	rawUserID, encodedUserID, err := GeneratePasskeyUserID(storedUserID.Value, allowCreate)
	if err != nil {
		return nil, err
	}
	if storedUserID.Value == "" && encodedUserID != "" {
		if err := settingRepo.Update(passkey.PasskeyUserIDSettingKey, encodedUserID); err != nil {
			return nil, err
		}
	}
	nameSetting, err := settingRepo.Get(repo.WithByKey("UserName"))
	if err != nil {
		return nil, err
	}
	return NewPasskeyUser(rawUserID, nameSetting.Value, records), nil
}

func loadCommunityPasskeyCredentialRecords() ([]passkey.PasskeyCredentialRecord, error) {
	setting, err := repo.NewISettingRepo().Get(repo.WithByKey(passkey.PasskeyCredentialSettingKey))
	if err != nil {
		return nil, err
	}
	return LoadPasskeyCredentialRecords(setting.Value)
}

func saveCommunityPasskeyCredentialRecords(records []passkey.PasskeyCredentialRecord) error {
	encoded, err := SavePasskeyCredentialRecords(records)
	if err != nil {
		return err
	}
	return repo.NewISettingRepo().Update(passkey.PasskeyCredentialSettingKey, encoded)
}

func PasskeyEnabled(c *gin.Context) (bool, error) {
	return strings.EqualFold(PasskeyRequestScheme(c), "https"), nil
}

func PasskeyConfig(c *gin.Context) (*webauthn.Config, string, error) {
	enabled, err := PasskeyEnabled(c)
	if err != nil {
		return nil, "", err
	}
	if !enabled {
		return nil, "ErrPasskeyDisabled", buserr.New("ErrPasskeyDisabled")
	}
	origin, rpID, err := PasskeyOriginAndRPID(c)
	if err != nil {
		return nil, "", err
	}
	panelName, err := repo.NewISettingRepo().Get(repo.WithByKey("PanelName"))
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

func PasskeyOriginAndRPID(c *gin.Context) (string, string, error) {
	host := passkeyRequestHost(c)
	if host == "" {
		return "", "", fmt.Errorf("missing request host")
	}
	scheme := PasskeyRequestScheme(c)
	origin := fmt.Sprintf("%s://%s", scheme, host)

	bindDomain, err := repo.NewISettingRepo().Get(repo.WithByKey("BindDomain"))
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

func NewPasskeyUser(userID []byte, name string, records []passkey.PasskeyCredentialRecord) *passkey.PasskeyUser {
	credentials := make([]webauthn.Credential, len(records))
	for i, record := range records {
		credentials[i] = record.Credential
	}
	return &passkey.PasskeyUser{
		ID:          userID,
		Name:        name,
		DisplayName: name,
		Credentials: credentials,
	}
}

func GeneratePasskeyUserID(encoded string, allowCreate bool) ([]byte, string, error) {
	if encoded == "" {
		if !allowCreate {
			return nil, "", buserr.New("ErrPasskeyNotConfigured")
		}
		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			return nil, "", err
		}
		return raw, base64.RawURLEncoding.EncodeToString(raw), nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, "", err
	}
	return raw, encoded, nil
}

func LoadPasskeyCredentialRecords(encryptedValue string) ([]passkey.PasskeyCredentialRecord, error) {
	if encryptedValue == "" {
		return []passkey.PasskeyCredentialRecord{}, nil
	}
	decrypted, err := encrypt.StringDecrypt(encryptedValue)
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

func SavePasskeyCredentialRecords(records []passkey.PasskeyCredentialRecord) (string, error) {
	if len(records) == 0 {
		return "", nil
	}
	copyRecords := make([]passkey.PasskeyCredentialRecord, len(records))
	copy(copyRecords, records)
	for i := range copyRecords {
		copyRecords[i].FlagsValue = CredentialFlagsValue(copyRecords[i].Credential.Flags)
	}
	raw, err := json.Marshal(copyRecords)
	if err != nil {
		return "", err
	}
	return encrypt.StringEncrypt(string(raw))
}

func PasskeyCredentialExists(records []passkey.PasskeyCredentialRecord, credentialID []byte) bool {
	encoded := base64.RawURLEncoding.EncodeToString(credentialID)
	for _, record := range records {
		if record.ID == encoded {
			return true
		}
	}
	return false
}

func UpdatePasskeyCredentialRecord(records []passkey.PasskeyCredentialRecord, credential *webauthn.Credential) error {
	encoded := base64.RawURLEncoding.EncodeToString(credential.ID)
	for i := range records {
		if records[i].ID == encoded {
			records[i].Credential = *credential
			records[i].FlagsValue = CredentialFlagsValue(credential.Flags)
			records[i].LastUsedAt = time.Now().Format(constant.DateTimeLayout)
			return nil
		}
	}
	return buserr.New("ErrPasskeyNotConfigured")
}

func CredentialFlagsValue(flags webauthn.CredentialFlags) uint8 {
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

func PasskeyRequestScheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}
	if !passkeyIsFromTrustedProxy(c) {
		return "http"
	}
	if proto := passkeyForwardedProto(c.GetHeader("Forwarded")); proto != "" {
		return proto
	}
	if proto := passkeyXForwardedProto(c.GetHeader("X-Forwarded-Proto")); proto != "" {
		return proto
	}
	return "http"
}

func passkeyRequestHost(c *gin.Context) string {
	host := c.Request.Host
	if strings.Contains(host, ",") {
		host = strings.TrimSpace(strings.Split(host, ",")[0])
	}
	return strings.TrimSpace(host)
}

func passkeyIsFromTrustedProxy(c *gin.Context) bool {
	remoteIP := passkeyRemoteIP(c.Request.RemoteAddr)
	if remoteIP == nil {
		return false
	}
	proxies, err := loadPasskeyTrustedProxies()
	if err != nil {
		global.LOG.Errorf("load passkey trusted proxies failed, err: %v", err)
		return false
	}
	for _, cidr := range proxies {
		if cidr.Contains(remoteIP) {
			return true
		}
	}
	return false
}

func passkeyRemoteIP(remoteAddr string) net.IP {
	if host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr)); err == nil {
		return net.ParseIP(host)
	}
	return net.ParseIP(strings.TrimSpace(remoteAddr))
}

func loadPasskeyTrustedProxies() ([]*net.IPNet, error) {
	setting, err := repo.NewISettingRepo().Get(repo.WithByKey("PasskeyTrustedProxies"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return parsePasskeyTrustedProxies("127.0.0.1\n::1")
		}
		return nil, err
	}
	return parsePasskeyTrustedProxies(setting.Value)
}

func parsePasskeyTrustedProxies(value string) ([]*net.IPNet, error) {
	lines := strings.Split(value, "\n")
	proxies := make([]*net.IPNet, 0, len(lines))
	for _, line := range lines {
		entry := strings.TrimSpace(line)
		if entry == "" {
			continue
		}
		if !strings.Contains(entry, "/") {
			if strings.Contains(entry, ":") {
				entry += "/128"
			} else {
				entry += "/32"
			}
		}
		_, cidr, err := net.ParseCIDR(entry)
		if err != nil {
			return nil, err
		}
		proxies = append(proxies, cidr)
	}
	return proxies, nil
}

func passkeyForwardedProto(forwarded string) string {
	for _, part := range strings.Split(forwarded, ";") {
		item := strings.TrimSpace(part)
		if len(item) < 6 || !strings.EqualFold(item[:6], "proto=") {
			continue
		}
		proto := strings.Trim(strings.TrimSpace(item[6:]), `"`)
		if strings.EqualFold(proto, "https") {
			return "https"
		}
		if strings.EqualFold(proto, "http") {
			return "http"
		}
	}
	return ""
}

func passkeyXForwardedProto(forwarded string) string {
	if forwarded == "" {
		return ""
	}
	proto := strings.TrimSpace(strings.Split(forwarded, ",")[0])
	if strings.EqualFold(proto, "https") {
		return "https"
	}
	if strings.EqualFold(proto, "http") {
		return "http"
	}
	return ""
}

func stripHostPort(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	if strings.HasPrefix(host, "[") {
		if parsedHost, _, err := net.SplitHostPort(host); err == nil {
			return strings.Trim(parsedHost, "[]")
		}
	}
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		return parsedHost
	}
	return strings.Trim(host, "[]")
}
