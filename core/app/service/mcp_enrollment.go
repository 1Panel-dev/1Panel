package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/auth"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	paneli18n "github.com/1Panel-dev/1Panel/core/i18n"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
	"github.com/google/uuid"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"gorm.io/gorm"
)

const mcpSecretVersion = 1

func (s *MCPService) secretAEAD() (cipher.AEAD, error) {
	unavailable := contract.Fail("MCP_SECRET_KEY_UNAVAILABLE", 503, "MCP credential encryption is unavailable")
	if s.SecretKey == nil {
		return nil, unavailable
	}
	master, err := s.SecretKey()
	if err != nil || master == "" {
		return nil, unavailable
	}
	derived := sha256.Sum256([]byte("1panel:mcp-token:v1:" + master))
	block, err := aes.NewCipher(derived[:])
	if err != nil {
		return nil, unavailable
	}
	return cipher.NewGCM(block)
}

func (s *MCPService) sealCredentialToken(key *model.MCPCredential, token string) error {
	aead, err := s.secretAEAD()
	if err != nil {
		return err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return contract.Fail("MCP_SECRET_KEY_UNAVAILABLE", 503, "MCP credential encryption is unavailable")
	}
	key.SecretCiphertext = base64.RawStdEncoding.EncodeToString(aead.Seal(nonce, nonce, []byte(token), []byte("mcp-token:v1:"+key.ID)))
	key.SecretVersion = mcpSecretVersion
	return nil
}

func (s *MCPService) openCredentialToken(key model.MCPCredential) (string, error) {
	invalid := contract.Fail("MCP_SECRET_INVALID", 409, "The stored MCP token cannot be decrypted; rotate the credential")
	if key.SecretVersion != mcpSecretVersion {
		return "", invalid
	}
	aead, err := s.secretAEAD()
	if err != nil {
		return "", err
	}
	data, err := base64.RawStdEncoding.DecodeString(key.SecretCiphertext)
	if err != nil || len(data) < aead.NonceSize()+aead.Overhead() || len(data) > 256 {
		return "", invalid
	}
	plain, err := aead.Open(nil, data[:aead.NonceSize()], data[aead.NonceSize():], []byte("mcp-token:v1:"+key.ID))
	if err != nil {
		return "", invalid
	}
	token := string(plain)
	if len(token) != 68 || !strings.HasPrefix(token, "mcp_") || subtle.ConstantTimeCompare([]byte(tokenDigest(token)), []byte(key.TokenHash)) != 1 {
		return "", invalid
	}
	if _, err := hex.DecodeString(token[4:]); err != nil {
		return "", invalid
	}
	return token, nil
}

func (s *MCPService) enrollmentKey(owner auth.APIKeyOwner, req dto.MCPCredentialMutation) (model.MCPCredential, error) {
	var key model.MCPCredential
	if _, err := uuid.Parse(req.ID); err != nil || req.Revision == 0 {
		return key, contract.Invalid()
	}
	if err := s.authorize(owner, "mcp_access_manage"); err != nil {
		return key, err
	}
	var err error
	key, err = (repo.MCPRepo{DB: s.DB}).Credential(owner.Type, owner.ID, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return key, contract.Forbidden()
		}
		return key, err
	}
	if key.Revision != req.Revision {
		return key, contract.Fail("MCP_CONFLICT", 409, "Credential changed; refresh and retry")
	}
	if err := s.enrollmentActive(key, false); err != nil {
		return key, err
	}
	return key, nil
}

func (s *MCPService) enrollmentActive(key model.MCPCredential, allowDisabled bool) error {
	active := key.Status == "Enable" || allowDisabled && key.Status == "Disable"
	if !active || key.RevokedAt != nil || key.ExpiresAt != nil && !s.Now().Before(*key.ExpiresAt) {
		return contract.Fail("MCP_CREDENTIAL_INACTIVE", 409, "The credential is inactive or expired")
	}
	return nil
}

func (s *MCPService) Reveal(owner auth.APIKeyOwner, req dto.MCPCredentialMutation) (result *dto.MCPCredentialSecret, resultErr error) {
	defer func() {
		if err := s.auditEnrollment(s.DB, owner, "reveal", req, resultErr); err != nil {
			result, resultErr = nil, err
		}
	}()
	mcpMutationMu.Lock()
	defer mcpMutationMu.Unlock()
	key, err := s.enrollmentKey(owner, req)
	if err != nil {
		return nil, err
	}
	token, err := s.openCredentialToken(key)
	if err != nil {
		return nil, err
	}
	return &dto.MCPCredentialSecret{ID: key.ID, Token: token}, nil
}

func (s *MCPService) Rotate(owner auth.APIKeyOwner, req dto.MCPCredentialMutation) (result *dto.MCPCredentialCreated, resultErr error) {
	defer func() {
		if resultErr != nil {
			if err := s.auditEnrollment(s.DB, owner, "rotate", req, resultErr); err != nil {
				result, resultErr = nil, err
			}
		}
	}()
	if err := s.authorize(owner, "mcp_access_manage"); err != nil {
		return nil, err
	}
	var token string
	var item dto.MCPCredentialItem
	catalog, _ := s.catalog()
	err := s.mutate(owner, req, func(tx *gorm.DB, key *model.MCPCredential) error {
		if err := s.authorize(owner, "mcp_access_manage"); err != nil {
			return err
		}
		if err := s.enrollmentActive(*key, true); err != nil {
			return err
		}
		secret := make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return contract.Fail("MCP_SECRET_KEY_UNAVAILABLE", 503, "MCP credential encryption is unavailable")
		}
		token = "mcp_" + hex.EncodeToString(secret)
		key.TokenHash = tokenDigest(token)
		key.KeyHint = token[:8] + "…" + token[len(token)-4:]
		if err := s.sealCredentialToken(key, token); err != nil {
			return err
		}
		view := *s
		view.DB = tx
		var err error
		item, err = view.credentialItem(*key, catalogIndex(catalog))
		item.Revision = key.Revision + 1
		if err != nil {
			return err
		}
		return s.auditEnrollment(tx, owner, "rotate", dto.MCPCredentialMutation{ID: key.ID, Revision: item.Revision}, nil)
	})
	if err != nil {
		return nil, err
	}
	return &dto.MCPCredentialCreated{Item: item, Token: token}, nil
}

func mcpDocumentEndpoint(endpoint string) bool {
	u, err := url.Parse(endpoint)
	return err == nil && len(endpoint) <= 2048 && !strings.ContainsAny(endpoint, "\r\n\x00") &&
		(u.Scheme == "https" || u.Scheme == "http") && u.Host != "" && u.User == nil && u.Path == "/mcp" &&
		u.RawQuery == "" && u.Fragment == "" && u.Opaque == ""
}

func mcpDocumentJSON(value any) string {
	data, _ := json.MarshalIndent(value, "", "  ")
	return strings.ReplaceAll(string(data), "`", "\\u0060")
}

func mcpEnrollmentConfig(client, id, endpoint, token string) (string, string) {
	name := "1panel_" + strings.ReplaceAll(id, "-", "")
	if client == "codex" {
		urlJSON := mcpDocumentJSON(endpoint)
		authJSON := mcpDocumentJSON("Bearer " + token)
		return "toml", fmt.Sprintf("[mcp_servers.%s]\nurl = %s\nhttp_headers = { Authorization = %s }", name, urlJSON, authJSON)
	}
	return "json", mcpDocumentJSON(map[string]any{"mcpServers": map[string]any{name: map[string]any{
		"type": "http", "url": endpoint, "headers": map[string]string{"Authorization": "Bearer " + token},
	}}})
}

func mcpEnrollmentDocument(locale, client, syntax, config, endpoint string) (string, error) {
	localizer, err := paneli18n.NewMCPDocumentLocalizer(locale)
	if err != nil {
		return "", contract.Fail("MCP_DOCUMENT_UNAVAILABLE", 500, "MCP connection document is unavailable")
	}
	clientKey := map[string]string{
		"codex": "MCPDocumentClientCodex", "claude": "MCPDocumentClientClaude", "generic": "MCPDocumentClientGeneric",
	}[client]
	keys := []string{
		"MCPDocumentTitle", "MCPDocumentIntro", "MCPDocumentTaskTitle", "MCPDocumentTask", clientKey,
		"MCPDocumentSteps", "MCPDocumentConfigTitle", "MCPDocumentNotes",
	}
	if strings.HasPrefix(endpoint, "http://") {
		keys = append(keys, "MCPDocumentHTTPWarning")
	}
	texts := make([]string, len(keys))
	for index, key := range keys {
		texts[index], err = localizer.Localize(&goi18n.LocalizeConfig{MessageID: key})
		if err != nil || texts[index] == "" {
			return "", contract.Fail("MCP_DOCUMENT_UNAVAILABLE", 500, "MCP connection document is unavailable")
		}
	}
	content := fmt.Sprintf("# %s\n\n%s\n\n## %s\n\n%s\n\n%s\n\n%s\n\n## %s\n\n```%s\n%s\n```\n\n%s\n",
		texts[0], texts[1], texts[2], texts[3], texts[4], texts[5], texts[6], syntax, config, texts[7])
	if len(texts) > 8 {
		content += "\n" + texts[8] + "\n"
	}
	return content, nil
}

func (s *MCPService) ExportMarkdown(owner auth.APIKeyOwner, req dto.MCPEnrollmentExport, endpoint, locale string) (result *dto.MCPEnrollmentDocument, resultErr error) {
	defer func() {
		if err := s.auditEnrollment(s.DB, owner, "export", req.MCPCredentialMutation, resultErr); err != nil {
			result, resultErr = nil, err
		}
	}()
	if req.Client != "generic" && req.Client != "codex" && req.Client != "claude" {
		return nil, contract.Invalid()
	}
	if !mcpDocumentEndpoint(endpoint) {
		return nil, contract.Fail("MCP_ENDPOINT_UNAVAILABLE", 409, "Configure an MCP external endpoint before exporting")
	}
	mcpMutationMu.Lock()
	defer mcpMutationMu.Unlock()
	key, err := s.enrollmentKey(owner, req.MCPCredentialMutation)
	if err != nil {
		return nil, err
	}
	token, err := s.openCredentialToken(key)
	if err != nil {
		return nil, err
	}
	language, config := mcpEnrollmentConfig(req.Client, key.ID, endpoint, token)
	content, err := mcpEnrollmentDocument(locale, req.Client, language, config, endpoint)
	if err != nil {
		return nil, err
	}
	return &dto.MCPEnrollmentDocument{Filename: "1panel-mcp-" + key.ID + "-" + req.Client + ".md", Content: content}, nil
}
