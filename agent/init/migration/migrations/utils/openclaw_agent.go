package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	providercatalog "github.com/1Panel-dev/1Panel/agent/app/provider"
)

type OpenclawMeta struct {
	Provider string
	Model    string
	BaseURL  string
	APIKey   string
	Token    string
}

func GetEnvStr(envMap map[string]interface{}, key string) string {
	if envMap == nil {
		return ""
	}
	if value, ok := envMap[key]; ok {
		switch v := value.(type) {
		case string:
			return strings.TrimSpace(v)
		default:
			return strings.TrimSpace(fmt.Sprintf("%v", v))
		}
	}
	return ""
}

func ParseOpenclawMeta(fileData []byte) OpenclawMeta {
	meta := OpenclawMeta{}
	content := map[string]interface{}{}
	if err := json.Unmarshal(fileData, &content); err != nil {
		return meta
	}

	meta.Token = getNestedString(content, "gateway", "auth", "token")
	meta.Model = getNestedString(content, "agents", "defaults", "model", "primary")

	providerKey := ""
	if parts := strings.SplitN(meta.Model, "/", 2); len(parts) == 2 {
		providerKey = strings.TrimSpace(parts[0])
	}

	providers := getNestedMap(content, "models", "providers")
	providerConfig := map[string]interface{}{}
	if providerKey != "" {
		if cfg, ok := providers[providerKey].(map[string]interface{}); ok {
			providerConfig = cfg
		}
	}
	if len(providerConfig) == 0 && len(providers) == 1 {
		for key, value := range providers {
			if cfg, ok := value.(map[string]interface{}); ok {
				providerKey = key
				providerConfig = cfg
				break
			}
		}
	}

	meta.Provider = providerKey
	meta.BaseURL = getString(providerConfig, "baseUrl")
	meta.APIKey = getString(providerConfig, "apiKey")

	if meta.Model == "" && providerKey != "" {
		if modelID := getProviderFirstModelID(providerConfig); modelID != "" {
			meta.Model = providerKey + "/" + modelID
		}
	}

	return meta
}

func NormalizeOpenclawProvider(provider, baseURL string) string {
	p := strings.ToLower(strings.TrimSpace(provider))
	base := strings.ToLower(strings.TrimSpace(baseURL))
	switch p {
	case "minimax-portal":
		return "minimax"
	case "moonshot":
		if strings.Contains(base, "moonshot.cn") {
			return "kimi"
		}
		return "moonshot"
	default:
		return p
	}
}

func DefaultBaseURL(provider string) (string, bool) {
	return providercatalog.DefaultBaseURL(provider)
}

func getNestedMap(data map[string]interface{}, keys ...string) map[string]interface{} {
	current := data
	for _, key := range keys {
		next, ok := current[key].(map[string]interface{})
		if !ok {
			return map[string]interface{}{}
		}
		current = next
	}
	return current
}

func getNestedString(data map[string]interface{}, keys ...string) string {
	current := data
	for i, key := range keys {
		value, ok := current[key]
		if !ok {
			return ""
		}
		if i == len(keys)-1 {
			switch v := value.(type) {
			case string:
				return strings.TrimSpace(v)
			default:
				return strings.TrimSpace(fmt.Sprintf("%v", v))
			}
		}
		next, ok := value.(map[string]interface{})
		if !ok {
			return ""
		}
		current = next
	}
	return ""
}

func getString(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}
	value, ok := data[key]
	if !ok {
		return ""
	}
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func getProviderFirstModelID(data map[string]interface{}) string {
	if data == nil {
		return ""
	}
	rawModels, ok := data["models"].([]interface{})
	if !ok || len(rawModels) == 0 {
		return ""
	}
	first, ok := rawModels[0].(map[string]interface{})
	if !ok {
		return ""
	}
	return getString(first, "id")
}
