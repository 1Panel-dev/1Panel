package service

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/1Panel-dev/1Panel/core/app/auth"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/apikey_audit"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	terminalsession "github.com/1Panel-dev/1Panel/core/utils/terminal_session"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const APIKeyLimit = 20

var apiKeyMutationMu sync.Mutex

type APIKeyService struct {
	Provider        auth.APIKeyOwnerProvider
	RevokeTerminals terminalsession.RevokeFunc
}

func NewAPIKeyService() *APIKeyService {
	provider, _ := xpack.AuthProvider.(auth.APIKeyOwnerProvider)
	return &APIKeyService{Provider: provider, RevokeTerminals: xpack.AuthProvider.RevokeTerminalSessions}
}

func (s *APIKeyService) owner(c *gin.Context) (auth.APIKeyOwner, error) {
	if _, err := auth.RequireAPIKeySession(c); err != nil {
		return auth.APIKeyOwner{}, err
	}
	if s.Provider == nil {
		return auth.APIKeyOwner{}, buserr.New("ErrApiConfigStatusInvalid")
	}
	return s.Provider.CurrentAPIKeyOwner(c)
}

func apiKeyItem(key model.APIKey) dto.APIKeyItem {
	status := key.Status
	if status != "Revoked" && key.ExpiresAt != nil && !time.Now().Before(*key.ExpiresAt) {
		status = "Expired"
	}
	return dto.APIKeyItem{ID: key.ID, Kind: "apiKey", Name: key.Name, Description: key.Description, KeyHint: key.KeyHint, Status: status, IPWhiteList: key.IPWhiteList, APITrustedProxies: key.APITrustedProxies, APIKeyValidityTime: key.APIKeyValidityTime, ExpiresAt: key.ExpiresAt, AllowAppBinding: key.AllowAppBinding, Revision: key.Revision, CreatedAt: &key.CreatedAt}
}

func (s *APIKeyService) legacy(owner auth.APIKeyOwner) (dto.APIKeyItem, auth.APIAuthConfig, model.LegacyAPIKeyPolicy, string, error) {
	config, err := s.Provider.LoadLegacyAPIKey(owner)
	if err != nil {
		return dto.APIKeyItem{}, config, model.LegacyAPIKeyPolicy{}, "", err
	}
	policy, err := auth.LoadLegacyAPIKeyPolicy(owner)
	if err != nil {
		return dto.APIKeyItem{}, config, policy, "", err
	}
	fingerprint, revision := auth.LegacyAPIKeyFingerprint(owner, config, policy)
	item := dto.APIKeyItem{ID: "legacy", Kind: "legacy", Name: "Legacy API Key", KeyHint: keyHint(config.ApiKey), Status: config.ApiInterfaceStatus, IPWhiteList: config.IpWhiteList, APITrustedProxies: config.ApiTrustedProxies, APIKeyValidityTime: config.ApiKeyValidityTime, AllowAppBinding: policy.AllowAppBinding, Revision: revision}
	return item, config, policy, fingerprint, nil
}

func (s *APIKeyService) Search(c *gin.Context, req dto.APIKeySearch) (*dto.APIKeyPage, error) {
	owner, err := s.owner(c)
	if err != nil {
		return nil, err
	}
	if req.Page < 1 || req.PageSize < 1 || req.PageSize > 100 {
		return nil, buserr.New("ErrInvalidParams")
	}
	keys, err := repo.NewAPIKeyRepo().List(owner.Type, owner.ID, req.ExcludeRevoked)
	if err != nil {
		return nil, err
	}
	items := make([]dto.APIKeyItem, 0, len(keys)+1)
	legacy, config, _, _, err := s.legacy(owner)
	if err != nil {
		return nil, err
	}
	if config.ApiKey != "" {
		items = append(items, legacy)
	}
	used := 0
	for _, key := range keys {
		items = append(items, apiKeyItem(key))
		if key.Status != "Revoked" {
			used++
		}
	}
	result := &dto.APIKeyPage{Items: []dto.APIKeyItem{}, Total: len(items), Used: used, Limit: APIKeyLimit}
	start := (req.Page - 1) * req.PageSize
	if start >= 0 && start < len(items) {
		result.Items = items[start:min(start+req.PageSize, len(items))]
	}
	return result, nil
}

func validateAPIKeyFields(fields *dto.APIKeyFields, create bool) (*time.Time, error) {
	fields.Name = strings.TrimSpace(fields.Name)
	fields.Description = strings.TrimSpace(fields.Description)
	if fields.Name == "" || utf8.RuneCountInString(fields.Name) > 64 || utf8.RuneCountInString(fields.Description) > 256 || fields.APIKeyValidityTime < 0 || fields.APIKeyValidityTime > 1440 || len(fields.IPWhiteList) > 4096 || len(fields.APITrustedProxies) > 4096 {
		return nil, buserr.New("ErrInvalidParams")
	}
	ips, err := common.HandleIPList(fields.IPWhiteList)
	if err != nil || len(ips) == 0 {
		return nil, buserr.New("ErrInvalidParams")
	}

	fields.IPWhiteList = strings.Join(ips, "\n")
	fields.APITrustedProxies, err = auth.NormalizeAPITrustedProxies(fields.APITrustedProxies)
	if err != nil {
		return nil, err
	}
	if len(fields.ExpiresAt) == 0 {
		if !create {
			return nil, buserr.New("ErrInvalidParams")
		}
		expires := time.Now().UTC().Add(90 * 24 * time.Hour)
		return &expires, nil
	}
	if bytes.Equal(bytes.TrimSpace(fields.ExpiresAt), []byte("null")) {
		return nil, nil
	}
	var expires time.Time
	if err = json.Unmarshal(fields.ExpiresAt, &expires); err != nil || !expires.After(time.Now()) {
		return nil, buserr.New("ErrInvalidParams")
	}
	return &expires, nil
}

func keyHint(secret string) string {
	if len(secret) < 8 {
		return ""
	}
	return secret[:4] + "••••" + secret[len(secret)-4:]
}

func (s *APIKeyService) Create(c *gin.Context, req dto.APIKeyCreate) (*dto.APIKeyCreated, error) {
	owner, err := s.owner(c)
	if err != nil {
		return nil, err
	}
	if _, err = uuid.Parse(req.RequestID); err != nil {
		return nil, buserr.New("ErrInvalidParams")
	}
	expires, err := validateAPIKeyFields(&req.APIKeyFields, true)
	if err != nil {
		return nil, err
	}
	encoded, _ := json.Marshal(req.APIKeyFields)
	digest := sha256.Sum256(encoded)
	requestHash := hex.EncodeToString(digest[:])
	master, err := encrypt.APIKeyEncryptionKey()
	if err != nil {
		return nil, err
	}
	apiKeyMutationMu.Lock()
	var key model.APIKey
	var secret string
	repeated := false
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		q := repo.APIKeyOwnerQuery(tx, owner.Type, owner.ID)
		err := q.Where("request_id = ?", req.RequestID).First(&key).Error
		if err == nil {
			if key.RequestHash != requestHash {
				return buserr.New("ErrAPIKeyConflict")
			}
			repeated = true
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var count int64
		if err = repo.APIKeyOwnerQuery(tx.Model(&model.APIKey{}), owner.Type, owner.ID).Where("status <> ?", "Revoked").Count(&count).Error; err != nil {
			return err
		}
		if count >= APIKeyLimit {
			return buserr.New("ErrAPIKeyLimit")
		}
		if err = ensureAPIKeyName(tx, owner, req.Name, ""); err != nil {
			return err
		}
		bytes := make([]byte, 16)
		if _, err = rand.Read(bytes); err != nil {
			return err
		}
		secret = hex.EncodeToString(bytes)
		key = model.APIKey{ID: uuid.NewString(), OwnerType: owner.Type, OwnerID: owner.ID, Name: req.Name, ActiveName: &req.Name, Description: req.Description, RequestID: req.RequestID, RequestHash: requestHash, SecretVersion: 1, KeyHint: keyHint(secret), Status: constant.StatusEnable, IPWhiteList: req.IPWhiteList, APITrustedProxies: req.APITrustedProxies, APIKeyValidityTime: req.APIKeyValidityTime, ExpiresAt: expires, AllowAppBinding: req.AllowAppBinding, Revision: 1}
		key.SecretCiphertext, err = encrypt.EncryptAPIKey(secret, key.ID, master)
		if err != nil {
			return err
		}
		return tx.Create(&key).Error
	})
	apiKeyMutationMu.Unlock()
	if err != nil {
		return nil, err
	}
	if !repeated {
		recordAPIKeyEvent(c, owner, key.ID, key.Name, "create")
	}
	return &dto.APIKeyCreated{Item: apiKeyItem(key), APIKey: secret, AlreadyCreated: repeated}, nil
}

func ensureAPIKeyName(tx *gorm.DB, owner auth.APIKeyOwner, name, id string) error {
	var count int64
	err := repo.APIKeyOwnerQuery(tx.Model(&model.APIKey{}), owner.Type, owner.ID).Where("active_name = ? AND id <> ?", name, id).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return buserr.New("ErrAPIKeyNameExists")
	}
	return nil
}

func (s *APIKeyService) Update(c *gin.Context, req dto.APIKeyUpdate) error {
	owner, err := s.owner(c)
	if err != nil {
		return err
	}
	if req.ID == "legacy" {
		return s.updateLegacy(c, owner, req.APIKeyMutation, &req.APIKeyFields, "update", "")
	}
	expires, err := validateAPIKeyFields(&req.APIKeyFields, false)
	if err != nil {
		return err
	}
	return s.mutate(c, owner, req.APIKeyMutation, "update", func(tx *gorm.DB, key *model.APIKey) error {
		if err := ensureAPIKeyName(tx, owner, req.Name, key.ID); err != nil {
			return err
		}
		key.Name = req.Name
		key.ActiveName = &req.Name
		key.Description = req.Description
		key.IPWhiteList = req.IPWhiteList
		key.APITrustedProxies = req.APITrustedProxies
		key.APIKeyValidityTime = req.APIKeyValidityTime
		key.ExpiresAt = expires
		key.AllowAppBinding = req.AllowAppBinding
		return nil
	})
}

func (s *APIKeyService) Status(c *gin.Context, req dto.APIKeyStatus) error {
	owner, err := s.owner(c)
	if err != nil {
		return err
	}
	if req.Status != constant.StatusEnable && req.Status != constant.StatusDisable {
		return buserr.New("ErrInvalidParams")
	}
	action := "disable"
	if req.Status == constant.StatusEnable {
		action = "enable"
	}
	if req.ID == "legacy" {
		return s.updateLegacy(c, owner, req.APIKeyMutation, nil, action, req.Status)
	}
	return s.mutate(c, owner, req.APIKeyMutation, action, func(_ *gorm.DB, key *model.APIKey) error {
		if req.Status == constant.StatusEnable && key.ExpiresAt != nil && !time.Now().Before(*key.ExpiresAt) {
			return buserr.New("ErrApiConfigKeyExpired")
		}
		key.Status = req.Status
		return nil
	})
}

func (s *APIKeyService) Revoke(c *gin.Context, req dto.APIKeyMutation) error {
	owner, err := s.owner(c)
	if err != nil {
		return err
	}
	if req.ID == "legacy" {
		return s.updateLegacy(c, owner, req, nil, "revoke", constant.StatusDisable)
	}
	return s.mutate(c, owner, req, "revoke", func(_ *gorm.DB, key *model.APIKey) error {
		now := time.Now()
		key.Status = "Revoked"
		key.RevokedAt = &now
		key.ActiveName = nil
		key.SecretCiphertext = ""
		key.AllowAppBinding = false
		return nil
	})
}

func (s *APIKeyService) mutate(c *gin.Context, owner auth.APIKeyOwner, req dto.APIKeyMutation, action string, change func(*gorm.DB, *model.APIKey) error) error {
	if req.ID == "" || req.Revision == 0 {
		return buserr.New("ErrInvalidParams")
	}
	apiKeyMutationMu.Lock()
	var key model.APIKey
	closeTerminals := action == "disable" || action == "revoke"
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := repo.APIKeyOwnerQuery(tx, owner.Type, owner.ID).Where("id = ?", req.ID).First(&key).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return buserr.New("ErrAPIKeyNotFound")
			}
			return err
		}
		if key.Revision != req.Revision || key.Status == "Revoked" {
			return buserr.New("ErrAPIKeyConflict")
		}
		before := key
		if err := change(tx, &key); err != nil {
			return err
		}
		closeTerminals = closeTerminals || before.IPWhiteList != key.IPWhiteList || before.APITrustedProxies != key.APITrustedProxies || before.APIKeyValidityTime != key.APIKeyValidityTime || (key.ExpiresAt != nil && (before.ExpiresAt == nil || key.ExpiresAt.Before(*before.ExpiresAt)))
		key.Revision++
		result := tx.Model(&model.APIKey{}).Where("id = ? AND revision = ?", key.ID, req.Revision).Select("*").Updates(&key)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return buserr.New("ErrAPIKeyConflict")
		}
		return nil
	})
	apiKeyMutationMu.Unlock()
	if err != nil {
		return err
	}
	recordAPIKeyEvent(c, owner, key.ID, key.Name, action)
	if closeTerminals {
		s.revokeAPIKeyTerminals(c, owner, key.ID)
	}
	return nil
}

func (s *APIKeyService) updateLegacy(c *gin.Context, owner auth.APIKeyOwner, req dto.APIKeyMutation, fields *dto.APIKeyFields, action, status string) error {
	return auth.WithLegacyAPIKeyMutation(func() error {
		return s.updateLegacyLocked(c, owner, req, fields, action, status)
	})
}

func (s *APIKeyService) updateLegacyLocked(c *gin.Context, owner auth.APIKeyOwner, req dto.APIKeyMutation, fields *dto.APIKeyFields, action, status string) error {
	item, config, policy, _, err := s.legacy(owner)
	if err != nil {
		return err
	}
	if config.ApiKey == "" {
		return buserr.New("ErrAPIKeyNotFound")
	}
	if req.Revision == 0 || req.Revision != item.Revision {
		return buserr.New("ErrAPIKeyConflict")
	}
	previousConfig := config
	if fields != nil {
		fields.Name = "Legacy API Key"
		legacyWindow := fields.APIKeyValidityTime
		if legacyWindow > 1440 && legacyWindow == config.ApiKeyValidityTime {
			fields.APIKeyValidityTime = 1440
		}
		expires, err := validateAPIKeyFields(fields, false)
		fields.APIKeyValidityTime = legacyWindow
		if err != nil {
			return err
		}
		if expires != nil {
			return buserr.New("ErrInvalidParams")
		}
		config.IpWhiteList = fields.IPWhiteList
		config.ApiTrustedProxies = fields.APITrustedProxies
		config.ApiKeyValidityTime = fields.APIKeyValidityTime
		policy.AllowAppBinding = fields.AllowAppBinding
	}
	if status != "" {
		config.ApiInterfaceStatus = status
	}
	if action == "revoke" {
		config.ApiKey = ""
		policy.AllowAppBinding = false
	}
	allowAppBinding := policy.AllowAppBinding
	policy.AllowAppBinding = false
	policy.Revision++
	if err = global.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(&policy).Error; err != nil {
		return err
	}
	if err = s.Provider.SaveLegacyAPIKey(owner, config); err != nil {
		return err
	}
	if allowAppBinding {
		policy.AllowAppBinding = true
		policy.Revision++
		if err = global.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(&policy).Error; err != nil {
			return err
		}
	}
	recordAPIKeyEvent(c, owner, "legacy", item.Name, action)
	if action == "disable" || action == "revoke" || previousConfig.IpWhiteList != config.IpWhiteList || previousConfig.ApiTrustedProxies != config.ApiTrustedProxies || previousConfig.ApiKeyValidityTime != config.ApiKeyValidityTime {
		s.revokeAPIKeyTerminals(c, owner, "legacy")
	}
	return nil
}

func recordAPIKeyEvent(c *gin.Context, owner auth.APIKeyOwner, id, name, action string) {
	if err := apikey_audit.Record(c, apikey_audit.Event{OwnerType: owner.Type, OwnerID: owner.ID, KeyID: id, KeyName: name, Action: action, Status: constant.StatusSuccess}); err != nil && global.LOG != nil {
		global.LOG.Errorf("API key audit failed: %v", err)
	}
}

func (s *APIKeyService) AuditLegacyChange(c *gin.Context) {
	owner, err := s.owner(c)
	if err != nil {
		if global.LOG != nil {
			global.LOG.Errorf("legacy API key audit owner unavailable: %v", err)
		}
		return
	}
	recordAPIKeyEvent(c, owner, "legacy", "Legacy API Key", "update")
}

func (s *APIKeyService) revokeAPIKeyTerminals(c *gin.Context, owner auth.APIKeyOwner, id string) {
	sessionID := "api-key:" + id
	if id == "legacy" {
		sessionID = terminalsession.APIAuthSessionID(owner.ID)
	}
	if err := terminalsession.RevokeWithRetry("auth_session", owner.ID, sessionID, s.RevokeTerminals); err != nil {
		c.Set("API_KEY_TERMINAL_CLOSE_PENDING", true)
	}
}

func (s *APIKeyService) PrepareAppBinding(c *gin.Context, id string) (*auth.APIKeyBinding, error) {
	owner, err := s.owner(c)
	if err != nil {
		return nil, err
	}
	if id == "" {
		id = "legacy"
	}
	if id == "legacy" {
		item, config, _, fingerprint, err := s.legacy(owner)
		if err != nil {
			return nil, err
		}
		if config.ApiKey == "" || config.ApiInterfaceStatus != constant.StatusEnable {
			return nil, buserr.New("ErrApiConfigStatusInvalid")
		}
		if !item.AllowAppBinding {
			return nil, buserr.New("ErrAPIKeyAppBindingDisabled")
		}
		return &auth.APIKeyBinding{KeyID: id, KeyName: item.Name, Owner: owner, Revision: item.Revision, Fingerprint: fingerprint}, nil
	}
	key, err := repo.NewAPIKeyRepo().Get(id)
	if err != nil || key.OwnerID != owner.ID || key.OwnerType != owner.Type {
		return nil, buserr.New("ErrAPIKeyNotFound")
	}
	if err = auth.ValidateAPIKeyState(key, time.Now()); err != nil {
		return nil, err
	}
	if !key.AllowAppBinding {
		return nil, buserr.New("ErrAPIKeyAppBindingDisabled")
	}
	return &auth.APIKeyBinding{KeyID: key.ID, KeyName: key.Name, Owner: owner, Revision: key.Revision}, nil
}

func (s *APIKeyService) ResolveAppBinding(c *gin.Context, binding auth.APIKeyBinding) (string, error) {
	if s.Provider == nil {
		return "", buserr.New("ErrApiConfigStatusInvalid")
	}
	owner, err := s.Provider.ResolveAPIKeyOwner(c, binding.Owner.Type, binding.Owner.ID)
	if err != nil {
		return "", buserr.New("ErrApiConfigStatusInvalid")
	}
	if binding.KeyID == "legacy" {
		item, config, _, fingerprint, err := s.legacy(owner)
		if err != nil {
			return "", err
		}
		if binding.Fingerprint == "" || fingerprint != binding.Fingerprint || binding.Revision != item.Revision {
			return "", buserr.New("ErrAPIKeyConflict")
		}
		if config.ApiKey == "" || config.ApiInterfaceStatus != constant.StatusEnable {
			return "", buserr.New("ErrApiConfigStatusInvalid")
		}
		if !item.AllowAppBinding {
			return "", buserr.New("ErrAPIKeyAppBindingDisabled")
		}
		if !auth.IsIPInWhiteList(auth.GetAPIClientIP(c, config.ApiTrustedProxies), config.IpWhiteList) {
			return "", buserr.New("ErrApiConfigIPInvalid")
		}
		c.Set("API_AUTH_CLIENT_IP", auth.GetAPIClientIP(c, config.ApiTrustedProxies))
		return config.ApiKey, nil
	}
	key, err := repo.NewAPIKeyRepo().Get(binding.KeyID)
	if err != nil || key.OwnerID != owner.ID || key.OwnerType != owner.Type {
		return "", buserr.New("ErrAPIKeyNotFound")
	}
	if key.Revision != binding.Revision {
		return "", buserr.New("ErrAPIKeyConflict")
	}
	if err = auth.ValidateAPIKeyState(key, time.Now()); err != nil {
		return "", err
	}
	if !key.AllowAppBinding {
		return "", buserr.New("ErrAPIKeyAppBindingDisabled")
	}
	if !auth.IsIPInWhiteList(auth.GetAPIClientIP(c, key.APITrustedProxies), key.IPWhiteList) {
		return "", buserr.New("ErrApiConfigIPInvalid")
	}
	c.Set("API_AUTH_CLIENT_IP", auth.GetAPIClientIP(c, key.APITrustedProxies))
	return auth.APIKeySecret(key)
}
