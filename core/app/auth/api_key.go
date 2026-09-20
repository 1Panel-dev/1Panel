package auth

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/init/session/psession"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const APIKeyOwnerPanelAdmin = "panel_admin"
const APIKeyOwnerEnterpriseUser = "enterprise_user"

func IsAPICredentialSetting(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "apikey", "apiinterfacestatus", "ipwhitelist", "apitrustedproxies", "apikeyvaliditytime", "encryptkey":
		return true
	default:
		return false
	}
}

type APIKeyOwner struct {
	Type         string
	ID           string
	Name         string
	IsSuperAdmin bool
}

type APIKeyOwnerProvider interface {
	CurrentAPIKeyOwner(c *gin.Context) (APIKeyOwner, error)
	ResolveAPIKeyOwner(c *gin.Context, ownerType, ownerID string) (APIKeyOwner, error)
	LoadLegacyAPIKey(owner APIKeyOwner) (APIAuthConfig, error)
	SaveLegacyAPIKey(owner APIKeyOwner, config APIAuthConfig) error
}

type APIKeyBinding struct {
	KeyID       string
	KeyName     string
	Owner       APIKeyOwner
	Revision    uint64
	Fingerprint string
}

func RequireAPIKeySession(c *gin.Context) (psession.SessionUser, error) {
	if c == nil || c.GetBool("API_AUTH") || c.GetBool("LOCAL_REQUEST") || HasAPICredentials(c) || global.SESSION == nil {
		return psession.SessionUser{}, buserr.New("ErrNotLogin")
	}
	u, err := global.SESSION.Get(c)
	if err != nil || u.ID == "" || u.Name == "" {
		return psession.SessionUser{}, buserr.New("ErrNotLogin")
	}
	c.Set(psession.GinContextSessionUserKey, u)
	return u, nil
}

func HasAPICredentials(c *gin.Context) bool {
	return c.GetHeader("1Panel-Token") != "" || c.GetHeader("1Panel-Timestamp") != "" || c.GetHeader("1Panel-Key-ID") != "" || c.GetHeader("1Panel-Signature-Version") != ""
}

func LoadLegacyAPIKeyPolicy(owner APIKeyOwner) (model.LegacyAPIKeyPolicy, error) {
	policy := model.LegacyAPIKeyPolicy{OwnerType: owner.Type, OwnerID: owner.ID, AllowAppBinding: true, Revision: 1}
	var saved model.LegacyAPIKeyPolicy
	err := repo.APIKeyOwnerQuery(global.DB, owner.Type, owner.ID).First(&saved).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return policy, nil
	}
	return saved, err
}

func LegacyAPIKeyFingerprint(owner APIKeyOwner, config APIAuthConfig, policy model.LegacyAPIKeyPolicy) (string, uint64) {
	data, _ := json.Marshal([]interface{}{owner.Type, owner.ID, config.ApiInterfaceStatus, config.ApiKey, config.IpWhiteList, config.ApiTrustedProxies, config.ApiKeyValidityTime, policy.AllowAppBinding, policy.Revision})
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), (binary.BigEndian.Uint64(hash[:8]) & ((1 << 53) - 1)) + 1
}

func APIKeySecret(key model.APIKey) (string, error) {
	if key.SecretVersion != 1 || key.SecretCiphertext == "" {
		return "", buserr.New("ErrApiConfigKeyInvalid")
	}
	master, err := encrypt.APIKeyEncryptionKey()
	if err != nil {
		return "", err
	}
	return encrypt.DecryptAPIKey(key.SecretCiphertext, key.ID, master)
}

func ValidateAPIKeyState(key model.APIKey, now time.Time) error {
	if key.Status != constant.StatusEnable {
		return buserr.New("ErrApiConfigStatusInvalid")
	}
	if key.ExpiresAt != nil && !now.Before(*key.ExpiresAt) {
		return buserr.New("ErrApiConfigKeyExpired")
	}
	return nil
}

func configForAPIKey(key model.APIKey, secret string, owner APIKeyOwner) APIAuthConfig {
	return APIAuthConfig{ApiInterfaceStatus: key.Status, ApiKey: secret, IpWhiteList: key.IPWhiteList, ApiTrustedProxies: key.APITrustedProxies, ApiKeyValidityTime: key.APIKeyValidityTime, KeyID: key.ID, KeyName: key.Name, KeyRevision: key.Revision, KeyExpiresAt: key.ExpiresAt, Owner: owner}
}

func loadMultiAPIKeyConfig(c *gin.Context, legacy APIAuthConfigLoader, provider APIKeyOwnerProvider) (APIAuthConfig, error) {
	keyID := c.GetHeader("1Panel-Key-ID")
	q := global.DB.Model(&model.APIKey{}).Select("id", "secret_ciphertext", "secret_version", "revision")
	if keyID != "" {
		q = q.Where("id = ?", keyID)
	} else {
		q = q.Where("status <> ?", "Revoked")
	}
	master, err := encrypt.APIKeyEncryptionKey()
	if err != nil {
		if keyID == "" {
			return legacy(c)
		}
		return APIAuthConfig{}, err
	}
	var candidates []model.APIKey
	var matched *model.APIKey
	err = q.Order("id").FindInBatches(&candidates, 128, func(_ *gorm.DB, _ int) error {
		for _, key := range candidates {
			if key.SecretVersion != 1 {
				continue
			}
			secret, err := encrypt.DecryptAPIKey(key.SecretCiphertext, key.ID, master)
			if err != nil {
				continue
			}
			if IsValid1PanelTokenWithVersion(c.GetHeader("1Panel-Token"), c.GetHeader("1Panel-Timestamp"), secret, c.GetHeader("1Panel-Signature-Version")) {
				copy := key
				matched = &copy
				return errAPIKeyMatched
			}
		}
		return nil
	}).Error
	if err != nil && !errors.Is(err, errAPIKeyMatched) {
		return APIAuthConfig{}, err
	}
	if matched == nil {
		if keyID != "" {
			return APIAuthConfig{}, buserr.New("ErrApiConfigKeyInvalid")
		}
		return legacy(c)
	}
	key, err := repo.NewAPIKeyRepo().Get(matched.ID)
	if err != nil || key.Revision != matched.Revision {
		return APIAuthConfig{}, buserr.New("ErrApiConfigStatusInvalid")
	}
	if err = ValidateAPIKeyState(key, time.Now()); err != nil {
		return APIAuthConfig{}, err
	}
	owner, err := provider.ResolveAPIKeyOwner(c, key.OwnerType, key.OwnerID)
	if err != nil {
		return APIAuthConfig{}, buserr.New("ErrApiConfigStatusInvalid")
	}
	secret, err := APIKeySecret(key)
	if err != nil {
		return APIAuthConfig{}, buserr.New("ErrApiConfigKeyInvalid")
	}
	return configForAPIKey(key, secret, owner), nil
}

var errAPIKeyMatched = errors.New("API credential matched")

func SetAPIKeyContext(c *gin.Context, config APIAuthConfig, provider APIKeyOwnerProvider) {
	if config.KeyID != "" {
		c.Set("API_AUTH_KEY_KIND", "apiKey")
		c.Set("API_AUTH_KEY_ID", config.KeyID)
		c.Set("API_AUTH_KEY_NAME", config.KeyName)
		c.Set("API_AUTH_KEY_REVISION", config.KeyRevision)
		c.Set("API_AUTH_OWNER_TYPE", config.Owner.Type)
		c.Set("API_AUTH_OWNER_ID", config.Owner.ID)
		if config.KeyExpiresAt != nil {
			c.Set("API_AUTH_KEY_EXPIRES_AT", *config.KeyExpiresAt)
		}
		role := "COMMON_USER"
		if config.Owner.IsSuperAdmin {
			role = "ADMIN"
		}
		c.Set(psession.GinContextSessionUserKey, psession.SessionUser{ID: config.Owner.ID, Name: config.Owner.Name, Role: role})
		c.Set("API_AUTH_USERNAME", config.Owner.Name)
		return
	}
	c.Set("API_AUTH_KEY_KIND", "legacy")
	c.Set("API_AUTH_KEY_ID", "legacy")
	c.Set("API_AUTH_KEY_NAME", "Legacy API Key")
	if provider != nil {
		if owner, err := provider.CurrentAPIKeyOwner(c); err == nil {
			c.Set("API_AUTH_OWNER_TYPE", owner.Type)
			c.Set("API_AUTH_OWNER_ID", owner.ID)
		}
	}
}
