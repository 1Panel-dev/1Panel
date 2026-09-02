package alert_config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
)

const MaskedSecret = "******"

type secretMutation struct {
	Action string `json:"action"`
	Value  string `json:"value,omitempty"`
}

func IsLegacySecretType(configType string) bool {
	return secretField(configType) != ""
}

func UsesMutation(configType, rawMutation string) (bool, error) {
	field := secretField(configType)
	if field == "" {
		return false, nil
	}
	root, err := decodeObject(rawMutation)
	if err != nil {
		return false, fmt.Errorf("decode alert config mutation: %w", err)
	}
	if raw, ok := root[field]; ok && rawIsObject(raw) {
		return true, nil
	}
	if !isWebhookType(configType) {
		return false, nil
	}
	raw, ok := root["webhooks"]
	if !ok {
		return false, nil
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return false, fmt.Errorf("webhooks mutation must be an array")
	}
	for _, item := range items {
		if rawURL, ok := item["url"]; ok && rawIsObject(rawURL) {
			return true, nil
		}
	}
	return false, nil
}

func Prepare(configType, rawMutation, status string, existing *model.AlertConfig) (string, error) {
	field := secretField(configType)
	if field == "" {
		return rawMutation, nil
	}
	root, err := decodeObject(rawMutation)
	if err != nil {
		return "", fmt.Errorf("decode alert config mutation: %w", err)
	}
	var existingRoot map[string]json.RawMessage
	if existing != nil {
		if existing.Type != configType {
			return "", fmt.Errorf("alert config %d has type %s, not %s", existing.ID, existing.Type, configType)
		}
		existingRoot, err = decodeObject(existing.Config)
		if err != nil {
			return "", fmt.Errorf("decode stored alert config: %w", err)
		}
	}

	existingSecret, err := storedSecret(existingRoot, field)
	if err != nil {
		return "", err
	}
	if raw, ok := root[field]; ok {
		value, err := mergeSecret(field, raw, existingSecret, existing != nil)
		if err != nil {
			return "", fmt.Errorf("merge alert config %s: %w", field, err)
		}
		root[field] = mustJSON(value)
	} else if existing != nil {
		root[field] = mustJSON(existingSecret)
	}

	if isWebhookType(configType) {
		if err := mergeWebhookArray(root, existingRoot, existing != nil); err != nil {
			return "", err
		}
	}
	if status == constant.AlertEnable {
		if configType == constant.SMSConfig {
			phone, err := storedSecret(root, "phone")
			if err != nil || phone == "" {
				return "", fmt.Errorf("SMS phone is required while the config is enabled")
			}
		}
		if isWebhookType(configType) && !hasWebhookURL(root) {
			return "", fmt.Errorf("webhook URL is required while the config is enabled")
		}
	}
	return encodeObject(root)
}

func secretField(configType string) string {
	switch configType {
	case constant.EmailConfig:
		return "password"
	case constant.SMSConfig:
		return "phone"
	case constant.WeCom, constant.DingTalk, constant.FeiShu, constant.Bark:
		return "url"
	default:
		return ""
	}
}

func isWebhookType(configType string) bool {
	switch configType {
	case constant.WeCom, constant.DingTalk, constant.FeiShu, constant.Bark:
		return true
	default:
		return false
	}
}

func decodeObject(raw string) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	var result map[string]json.RawMessage
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("alert config must be a JSON object")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("multiple JSON values are not allowed")
		}
		return nil, err
	}
	return result, nil
}

func encodeObject(value map[string]json.RawMessage) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("encode alert config: %w", err)
	}
	return string(encoded), nil
}

func mustJSON(value any) json.RawMessage {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return encoded
}

func decodeStoredSecret(raw json.RawMessage) (string, error) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("stored secret is not a string")
	}
	return value, nil
}

func storedSecret(root map[string]json.RawMessage, field string) (string, error) {
	if root == nil {
		return "", nil
	}
	raw, ok := root[field]
	if !ok {
		return "", nil
	}
	value, err := decodeStoredSecret(raw)
	if err != nil {
		return "", fmt.Errorf("stored alert config %s is invalid", field)
	}
	return value, nil
}

func mergeSecret(field string, raw json.RawMessage, existing string, hasExisting bool) (string, error) {
	if value, err := decodeStoredSecret(raw); err == nil {
		if hasExisting && isLegacyMaskValue(field, value) {
			return existing, nil
		}
		return value, nil
	}
	var mutation secretMutation
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&mutation); err != nil {
		return "", fmt.Errorf("secret mutation must be a string or keep/replace/clear object")
	}
	switch mutation.Action {
	case "keep":
		if !hasExisting {
			return "", fmt.Errorf("secret cannot be kept because the config does not exist")
		}
		return existing, nil
	case "replace":
		return mutation.Value, nil
	case "clear":
		return "", nil
	default:
		return "", fmt.Errorf("secret action must be keep, replace, or clear")
	}
}

func isLegacyMaskValue(field, value string) bool {
	if field == "phone" {
		return isLegacyMaskedPhone(value)
	}
	return value == MaskedSecret
}

func isLegacyMaskedPhone(value string) bool {
	value = strings.TrimSpace(value)
	if !strings.Contains(value, "*") {
		return false
	}
	return maskPhone(strings.ReplaceAll(value, "*", "0")) == value
}

func rawIsObject(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) > 0 && trimmed[0] == '{'
}

func mergeWebhookArray(root, existingRoot map[string]json.RawMessage, hasExisting bool) error {
	raw, ok := root["webhooks"]
	if !ok {
		return nil
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return fmt.Errorf("webhooks mutation must be an array")
	}
	var existingItems []map[string]json.RawMessage
	if existingRoot != nil {
		if existingRaw, ok := existingRoot["webhooks"]; ok {
			if err := json.Unmarshal(existingRaw, &existingItems); err != nil {
				return fmt.Errorf("stored webhooks must be an array")
			}
		}
	}
	for index := range items {
		rawURL, ok := items[index]["url"]
		if !ok {
			continue
		}
		var existingURL string
		itemExists := hasExisting && index < len(existingItems)
		if itemExists {
			var err error
			existingURL, err = storedSecret(existingItems[index], "url")
			if err != nil {
				return fmt.Errorf("stored webhook URL %d is invalid", index)
			}
		}
		value, err := mergeSecret("url", rawURL, existingURL, itemExists)
		if err != nil {
			return fmt.Errorf("merge webhook URL %d: %w", index, err)
		}
		items[index]["url"] = mustJSON(value)
	}
	root["webhooks"] = mustJSON(items)
	return nil
}

func hasWebhookURL(root map[string]json.RawMessage) bool {
	if value, err := storedSecret(root, "url"); err == nil && value != "" {
		return true
	}
	raw, ok := root["webhooks"]
	if !ok {
		return false
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return false
	}
	for _, item := range items {
		if value, err := storedSecret(item, "url"); err == nil && value != "" {
			return true
		}
	}
	return false
}

func maskPhone(value string) string {
	runes := []rune(strings.TrimSpace(value))
	switch {
	case len(runes) >= 11:
		return string(runes[:3]) + strings.Repeat("*", len(runes)-7) + string(runes[len(runes)-4:])
	case len(runes) >= 8:
		return string(runes[:2]) + strings.Repeat("*", len(runes)-4) + string(runes[len(runes)-2:])
	case utf8.RuneCountInString(value) == 0:
		return ""
	default:
		return MaskedSecret
	}
}
