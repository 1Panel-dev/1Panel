package alert_webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	"github.com/google/uuid"
)

const (
	coreKeyPrefix      = "core:v1:"
	secretSchemaV1     = 1
	maxSecretLength    = 256 * 1024
	maxSecretHeaders   = 64
	maxSecretValueSize = 16 * 1024
)

type secretConfig struct {
	SchemaVersion int               `json:"schemaVersion"`
	URL           string            `json:"url"`
	Headers       map[string]string `json:"headers,omitempty"`
}

func ExportSecretPlain(config model.AlertConfig) (string, error) {
	if config.Type != constant.Custom {
		return "", fmt.Errorf("alert config %d is not a custom webhook", config.ID)
	}
	if !strings.HasPrefix(config.SecretConfig, coreKeyPrefix) {
		return "", fmt.Errorf("custom webhook secret is not encrypted with the core key")
	}
	payload := strings.TrimPrefix(config.SecretConfig, coreKeyPrefix)
	if payload == "" {
		return "", fmt.Errorf("custom webhook secret ciphertext is empty")
	}
	key, err := loadEncryptKey()
	if err != nil {
		return "", err
	}
	plainText, err := encrypt.StringDecryptWithKey(payload, key)
	if err != nil {
		return "", fmt.Errorf("decrypt custom webhook secret: %w", err)
	}
	secret, err := decodeAndValidateSecret(plainText)
	if err != nil {
		return "", err
	}
	canonical, err := json.Marshal(secret)
	if err != nil {
		return "", fmt.Errorf("encode custom webhook secret: %w", err)
	}
	return string(canonical), nil
}

func loadEncryptKey() (string, error) {
	if key := strings.TrimSpace(global.CONF.Base.EncryptKey); key != "" {
		return key, nil
	}
	if global.DB != nil {
		var setting model.Setting
		if err := global.DB.Where("key = ?", "EncryptKey").First(&setting).Error; err == nil {
			if key := strings.TrimSpace(setting.Value); key != "" {
				return key, nil
			}
		}
	}
	return "", fmt.Errorf("custom webhook core encrypt key is empty")
}

func decodeAndValidateSecret(plainText string) (secretConfig, error) {
	if len(plainText) == 0 || len(plainText) > maxSecretLength {
		return secretConfig{}, fmt.Errorf("custom webhook secret has an invalid size")
	}
	var secret secretConfig
	decoder := json.NewDecoder(strings.NewReader(plainText))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&secret); err != nil {
		return secretConfig{}, fmt.Errorf("decode custom webhook secret: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return secretConfig{}, fmt.Errorf("decode custom webhook secret: multiple JSON values are not allowed")
	}
	if secret.SchemaVersion != secretSchemaV1 {
		return secretConfig{}, fmt.Errorf("unsupported custom webhook secret schemaVersion: %d", secret.SchemaVersion)
	}
	if err := validateSecretURL(secret.URL); err != nil {
		return secretConfig{}, err
	}
	if len(secret.Headers) > maxSecretHeaders {
		return secretConfig{}, fmt.Errorf("custom webhook secret headers exceed the limit")
	}
	if secret.Headers == nil {
		secret.Headers = map[string]string{}
	}
	for uid, value := range secret.Headers {
		if _, err := uuid.Parse(strings.TrimSpace(uid)); err != nil {
			return secretConfig{}, fmt.Errorf("custom webhook secret header uid must be a UUID")
		}
		if value == "" || len(value) > maxSecretValueSize || strings.ContainsAny(value, "\r\n") {
			return secretConfig{}, fmt.Errorf("custom webhook secret header is invalid")
		}
	}
	return secret, nil
}

func validateSecretURL(rawURL string) error {
	if rawURL == "" {
		return nil
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid custom webhook URL")
	}
	hostname := strings.TrimSuffix(strings.TrimSpace(parsed.Hostname()), ".")
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || hostname == "" || strings.Contains(hostname, "%") || parsed.User != nil || parsed.Fragment != "" {
		return fmt.Errorf("invalid custom webhook URL")
	}
	if strings.HasSuffix(parsed.Host, ":") {
		return fmt.Errorf("invalid custom webhook URL")
	}
	if port := parsed.Port(); port != "" {
		value, err := strconv.Atoi(port)
		if err != nil || value < 1 || value > 65535 {
			return fmt.Errorf("invalid custom webhook URL")
		}
	}
	return nil
}
