package provider

import (
	"fmt"
	"net/url"
	"strings"
)

type APIConfig struct {
	APIType         string
	BaseURL         string
	EditableBaseURL bool
	DiscoverModels  bool
	DefaultAuthMode string
	AuthModes       []string
	Models          []Model
}

const (
	AuthModeBearer  = "bearer"
	AuthModeXAPIKey = "x-api-key"
)

type Model struct {
	ID   string
	Name string
}

type Meta struct {
	Key            string
	DisplayName    string
	DisplayNameKey string
	Sort           uint
	DefaultAPIType string
	APIConfigs     []APIConfig
	EnvKey         string
	Models         []Model
}

var catalog = map[string]Meta{
	"custom": {
		Key: "custom", DisplayName: "Custom", Sort: 10, DefaultAPIType: "openai-completions", EnvKey: "CUSTOM_API_KEY",
		APIConfigs: editableAPIConfigs(true, "openai-completions", "openai-responses", "anthropic-messages", "openai-images", "openai-embeddings"),
	},
	"ollama": {
		Key: "ollama", DisplayName: "Ollama", Sort: 15, DefaultAPIType: "openai-responses",
		APIConfigs: editableAPIConfigs(false, "openai-responses", "openai-completions", "openai-embeddings"),
	},
	"vllm": {
		Key: "vllm", DisplayName: "vLLM", Sort: 20, DefaultAPIType: "openai-completions", EnvKey: "VLLM_API_KEY",
		APIConfigs: editableAPIConfigs(false, "openai-completions", "openai-responses", "anthropic-messages", "openai-images", "openai-embeddings"),
	},
	"deepseek": {
		Key: "deepseek", DisplayName: "DeepSeek", Sort: 25, DefaultAPIType: "openai-completions", EnvKey: "DEEPSEEK_API_KEY",
		APIConfigs: []APIConfig{
			{APIType: "openai-completions", BaseURL: "https://api.deepseek.com"},
			{APIType: "openai-responses", BaseURL: "https://api.deepseek.com"},
			anthropicAPIConfig("https://api.deepseek.com/anthropic", AuthModeXAPIKey),
		},
		Models: []Model{{ID: "deepseek-v4-flash", Name: "deepseek-v4-flash"}, {ID: "deepseek-v4-pro", Name: "deepseek-v4-pro"}},
	},
	"bailian-coding-plan": {
		Key: "bailian-coding-plan", DisplayNameKey: "AIProviderBailianCodingPlan", Sort: 30, DefaultAPIType: "openai-completions", EnvKey: "QWEN_API_KEY",
		APIConfigs: []APIConfig{
			{APIType: "openai-completions", BaseURL: "https://coding.dashscope.aliyuncs.com/v1"},
			anthropicAPIConfig("https://coding.dashscope.aliyuncs.com/apps/anthropic", AuthModeBearer),
		},
		Models: []Model{
			{ID: "qwen3-coder-plus", Name: "Qwen3-Coder-Plus"},
			{ID: "qwen3-max-2026-01-23", Name: "Qwen3-Max-2026-01-23"},
			{ID: "qwen3-coder-next", Name: "Qwen3-Coder-Next"},
			{ID: "glm-4.7", Name: "GLM-4.7"},
			{ID: "kimi-k2.5", Name: "Kimi K2.5"},
			{ID: "qwen3.5-plus", Name: "Qwen3.5-Plus"},
			{ID: "glm-5", Name: "GLM-5"},
			{ID: "MiniMax-M2.5", Name: "MiniMax M2.5"},
			{ID: "qwen3.6-plus", Name: "Qwen3.6-Plus"},
			{ID: "qwen3.7-plus", Name: "Qwen3.7-Plus"},
		},
	},
	"ark-coding-plan": {
		Key: "ark-coding-plan", DisplayNameKey: "AIProviderArkCodingPlan", Sort: 35, DefaultAPIType: "openai-completions", EnvKey: "ARK_API_KEY",
		APIConfigs: []APIConfig{
			{APIType: "openai-completions", BaseURL: "https://ark.cn-beijing.volces.com/api/coding/v3"},
			anthropicAPIConfig("https://ark.cn-beijing.volces.com/api/coding", AuthModeBearer),
		},
		Models: []Model{
			{ID: "ark-code-latest", Name: "Ark Coding Plan"}, {ID: "doubao-seed-code", Name: "Doubao Seed Code"},
			{ID: "glm-4.7", Name: "GLM 4.7 Coding"}, {ID: "kimi-k2-thinking", Name: "Kimi K2 Thinking"},
			{ID: "kimi-k2.5", Name: "Kimi K2.5 Coding"}, {ID: "doubao-seed-code-preview-251028", Name: "Doubao Seed Code Preview"},
		},
	},
	"zai": {
		Key: "zai", DisplayName: "Z.ai", Sort: 40, DefaultAPIType: "openai-completions", EnvKey: "ZAI_API_KEY",
		APIConfigs: []APIConfig{
			{APIType: "openai-completions", BaseURL: "https://open.bigmodel.cn/api/paas/v4", EditableBaseURL: true},
			{APIType: "openai-images", BaseURL: "https://open.bigmodel.cn/api/paas/v4", EditableBaseURL: true},
		},
		Models: []Model{{ID: "glm-5", Name: "GLM-5"}, {ID: "glm-4.7", Name: "GLM-4.7"}, {ID: "glm-4.7-flash", Name: "GLM-4.7-Flash"}, {ID: "glm-4.7-flashx", Name: "GLM-4.7-FlashX"}},
	},
	"minimax": {
		Key: "minimax", DisplayName: "MiniMax (CN)", Sort: 45, DefaultAPIType: "anthropic-messages", EnvKey: "MINIMAX_API_KEY",
		APIConfigs: []APIConfig{
			anthropicAPIConfig("https://api.minimaxi.com/anthropic", AuthModeXAPIKey, AuthModeBearer),
			{APIType: "openai-completions", BaseURL: "https://api.minimaxi.com/v1"},
			{APIType: "minimax-images", BaseURL: "https://api.minimaxi.com"},
		},
		Models: []Model{{ID: "MiniMax-M3", Name: "MiniMax M3"}, {ID: "MiniMax-M2.7", Name: "MiniMax M2.7"}, {ID: "MiniMax-M2.7-highspeed", Name: "MiniMax M2.7 highspeed"}},
	},
	"xiaomi": {
		Key: "xiaomi", DisplayName: "Xiaomi", Sort: 46, DefaultAPIType: "openai-completions", EnvKey: "XIAOMI_API_KEY",
		APIConfigs: []APIConfig{
			{APIType: "openai-completions", BaseURL: "https://api.xiaomimimo.com/v1"},
			{APIType: "openai-responses", BaseURL: "https://api.xiaomimimo.com/v1"},
			anthropicAPIConfig("https://api.xiaomimimo.com/anthropic", AuthModeBearer),
		},
		Models: []Model{{ID: "mimo-v2.5", Name: "Xiaomi MiMo V2.5"}, {ID: "mimo-v2.5-pro", Name: "Xiaomi MiMo V2.5 Pro"}},
	},
	"kimi": {
		Key: "kimi", DisplayName: "Kimi (CN)", Sort: 50, DefaultAPIType: "openai-completions", EnvKey: "KIMI_API_KEY",
		APIConfigs: []APIConfig{{APIType: "openai-completions", BaseURL: "https://api.moonshot.cn/v1"}},
		Models:     []Model{{ID: "kimi-k2.5", Name: "Kimi K2.5"}, {ID: "kimi-k2-0905-preview", Name: "Kimi K2 0905 Preview"}, {ID: "kimi-k2-thinking", Name: "Kimi K2 Thinking"}},
	},
	"kimi-coding": {
		Key: "kimi-coding", DisplayName: "Kimi Coding", Sort: 51, DefaultAPIType: "anthropic-messages", EnvKey: "KIMI_API_KEY",
		APIConfigs: []APIConfig{anthropicAPIConfig("https://api.kimi.com/coding/", AuthModeXAPIKey)},
		Models:     []Model{{ID: "kimi-code", Name: "Kimi Code"}, {ID: "k2p5", Name: "Kimi K2.5"}},
	},
	"openai": {
		Key: "openai", DisplayName: "OpenAI", Sort: 55, DefaultAPIType: "openai-responses", EnvKey: "OPENAI_API_KEY",
		APIConfigs: []APIConfig{
			{APIType: "openai-responses", BaseURL: "https://api.openai.com/v1"},
			{APIType: "openai-completions", BaseURL: "https://api.openai.com/v1"},
			{APIType: "openai-images", BaseURL: "https://api.openai.com/v1"},
			{APIType: "openai-embeddings", BaseURL: "https://api.openai.com/v1", Models: []Model{
				{ID: "text-embedding-3-small", Name: "text-embedding-3-small"},
				{ID: "text-embedding-3-large", Name: "text-embedding-3-large"},
			}},
		},
		Models: []Model{{ID: "gpt-5.4", Name: "gpt-5.4"}, {ID: "gpt-5.4-pro", Name: "gpt-5.4-pro"}, {ID: "gpt-5.4-mini", Name: "gpt-5.4-mini"}, {ID: "gpt-5.4-nano", Name: "gpt-5.4-nano"}},
	},
	"openrouter": {
		Key: "openrouter", DisplayName: "OpenRouter", Sort: 56, DefaultAPIType: "openai-completions", EnvKey: "OPENROUTER_API_KEY",
		APIConfigs: []APIConfig{
			{APIType: "openai-completions", BaseURL: "https://openrouter.ai/api/v1"},
			{APIType: "openrouter-images", BaseURL: "https://openrouter.ai"},
		},
		Models: []Model{{ID: "openrouter/free", Name: "openrouter/free"}, {ID: "openrouter/auto", Name: "openrouter/auto"}},
	},
	"anthropic": {
		Key: "anthropic", DisplayName: "Anthropic", Sort: 60, DefaultAPIType: "anthropic-messages", EnvKey: "ANTHROPIC_API_KEY",
		APIConfigs: []APIConfig{anthropicAPIConfig("https://api.anthropic.com", AuthModeXAPIKey)},
		Models:     []Model{{ID: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6"}, {ID: "claude-opus-4-6", Name: "Claude Opus 4.6"}, {ID: "claude-opus-4-5", Name: "Claude Opus 4.5"}, {ID: "claude-sonnet-4-5", Name: "Claude Sonnet 4.5"}, {ID: "claude-haiku-4-5", Name: "Claude Haiku 4.5"}},
	},
	"gemini": {
		Key: "gemini", DisplayName: "Gemini", Sort: 65, DefaultAPIType: "gemini-generate-content", EnvKey: "GEMINI_API_KEY",
		APIConfigs: []APIConfig{{APIType: "gemini-generate-content", BaseURL: "https://generativelanguage.googleapis.com"}},
		Models:     []Model{{ID: "gemini-3-flash-preview", Name: "Gemini 3 Flash Preview"}, {ID: "gemini-flash-latest", Name: "Gemini Flash Latest"}, {ID: "gemini-3-pro-preview", Name: "Gemini 3 Pro Preview"}},
	},
	"moonshot": {
		Key: "moonshot", DisplayName: "Moonshot (Global)", Sort: 70, DefaultAPIType: "openai-completions", EnvKey: "MOONSHOT_API_KEY",
		APIConfigs: []APIConfig{{APIType: "openai-completions", BaseURL: "https://api.moonshot.ai/v1"}},
		Models:     []Model{{ID: "kimi-k2.5", Name: "Kimi K2.5"}, {ID: "kimi-k2-0905-preview", Name: "Kimi K2 0905 Preview"}, {ID: "kimi-k2-thinking", Name: "Kimi K2 Thinking"}},
	},
	"bailian": {
		Key: "bailian", DisplayNameKey: "AIProviderBailian", Sort: 31, DefaultAPIType: "openai-completions", EnvKey: "DASHSCOPE_API_KEY",
		APIConfigs: []APIConfig{
			{
				APIType: "openai-completions", BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
				DiscoverModels: true,
				Models:         []Model{{ID: "qwen3.7-plus", Name: "qwen3.7-plus"}, {ID: "qwen3.6-plus", Name: "qwen3.6-plus"}, {ID: "qwen3.6-flash", Name: "qwen3.6-flash"}},
			},
			{
				APIType: "openai-responses", BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
				DiscoverModels: true,
				Models:         []Model{{ID: "qwen3.7-plus", Name: "qwen3.7-plus"}, {ID: "qwen3.6-plus", Name: "qwen3.6-plus"}, {ID: "qwen3.6-flash", Name: "qwen3.6-flash"}},
			},
			{
				APIType: "anthropic-messages", BaseURL: "https://dashscope.aliyuncs.com/apps/anthropic",
				DefaultAuthMode: AuthModeBearer,
				AuthModes:       []string{AuthModeBearer},
				Models:          []Model{{ID: "qwen3.7-plus", Name: "qwen3.7-plus"}, {ID: "qwen3.6-plus", Name: "qwen3.6-plus"}, {ID: "qwen3.6-flash", Name: "qwen3.6-flash"}},
			},
			{
				APIType: "dashscope-images", BaseURL: "https://dashscope.aliyuncs.com",
				Models: []Model{
					{ID: "qwen-image-2.0-pro", Name: "qwen-image-2.0-pro"},
					{ID: "qwen-image-2.0", Name: "qwen-image-2.0"},
					{ID: "wan2.7-image-pro", Name: "wan2.7-image-pro"},
					{ID: "wan2.7-image", Name: "wan2.7-image"},
				},
			},
		},
	},
	"ark": {
		Key: "ark", DisplayNameKey: "AIProviderArk", Sort: 36, DefaultAPIType: "openai-completions", EnvKey: "ARK_API_KEY",
		APIConfigs: []APIConfig{
			{
				APIType: "openai-completions", BaseURL: "https://ark.cn-beijing.volces.com/api/v3",
				DiscoverModels: true,
				Models:         []Model{{ID: "doubao-seed-2-0-pro-260215", Name: "doubao-seed-2-0-pro-260215"}, {ID: "doubao-seed-2-0-lite-260215", Name: "doubao-seed-2-0-lite-260215"}},
			},
			{
				APIType: "openai-responses", BaseURL: "https://ark.cn-beijing.volces.com/api/v3",
				DiscoverModels: true,
				Models:         []Model{{ID: "doubao-seed-2-0-pro-260215", Name: "doubao-seed-2-0-pro-260215"}, {ID: "doubao-seed-2-0-lite-260215", Name: "doubao-seed-2-0-lite-260215"}},
			},
			{
				APIType: "openai-images", BaseURL: "https://ark.cn-beijing.volces.com/api/v3",
				Models: []Model{
					{ID: "doubao-seedream-5-0-260128", Name: "doubao-seedream-5-0-260128"},
					{ID: "doubao-seedream-5-0-lite-260128", Name: "doubao-seedream-5-0-lite-260128"},
					{ID: "doubao-seedream-4-5-251128", Name: "doubao-seedream-4-5-251128"},
				},
			},
		},
	},
}

func editableAPIConfigs(discoverModels bool, apiTypes ...string) []APIConfig {
	configs := make([]APIConfig, 0, len(apiTypes))
	for _, apiType := range apiTypes {
		if apiType == "anthropic-messages" {
			config := anthropicAPIConfig("", AuthModeXAPIKey, AuthModeBearer)
			config.EditableBaseURL = true
			configs = append(configs, config)
			continue
		}
		configs = append(configs, APIConfig{
			APIType:         apiType,
			EditableBaseURL: true,
			DiscoverModels:  discoverModels && (apiType == "openai-completions" || apiType == "openai-responses"),
		})
	}
	return configs
}

func anthropicAPIConfig(baseURL, defaultAuthMode string, additionalAuthModes ...string) APIConfig {
	authModes := []string{defaultAuthMode}
	for _, authMode := range additionalAuthModes {
		if authMode != defaultAuthMode {
			authModes = append(authModes, authMode)
		}
	}
	return APIConfig{
		APIType:         "anthropic-messages",
		BaseURL:         baseURL,
		DefaultAuthMode: defaultAuthMode,
		AuthModes:       authModes,
	}
}

func Get(key string) (Meta, bool) {
	meta, ok := catalog[key]
	if !ok {
		return Meta{}, false
	}
	return cloneMeta(meta), true
}

func All() map[string]Meta {
	result := make(map[string]Meta, len(catalog))
	for key, meta := range catalog {
		result[key] = cloneMeta(meta)
	}
	return result
}

func FindAPIConfig(key, apiType string) (APIConfig, bool) {
	meta, ok := catalog[key]
	if !ok {
		return APIConfig{}, false
	}
	target := strings.TrimSpace(apiType)
	if target == "" {
		target = meta.DefaultAPIType
	}
	for _, config := range meta.APIConfigs {
		if config.APIType == target {
			config.AuthModes = append([]string(nil), config.AuthModes...)
			config.Models = append([]Model(nil), config.Models...)
			return config, true
		}
	}
	return APIConfig{}, false
}

func DefaultModels(key, apiType string) []Model {
	meta, ok := catalog[key]
	if !ok {
		return nil
	}
	target := strings.TrimSpace(apiType)
	if target == "" {
		target = meta.DefaultAPIType
	}
	for _, config := range meta.APIConfigs {
		if config.APIType != target {
			continue
		}
		if len(config.Models) > 0 {
			return append([]Model(nil), config.Models...)
		}
		if IsImageAPIType(config.APIType) || IsEmbeddingAPIType(config.APIType) {
			return nil
		}
		break
	}
	return append([]Model(nil), meta.Models...)
}

func ResolveAuthMode(provider, apiType, requested string) (string, error) {
	config, ok := FindAPIConfig(provider, apiType)
	if !ok {
		return "", fmt.Errorf("provider %s does not support api type %s", provider, apiType)
	}
	if config.APIType != "anthropic-messages" {
		return "", nil
	}
	authMode := strings.TrimSpace(requested)
	if authMode == "" {
		authMode = config.DefaultAuthMode
	}
	for _, allowed := range config.AuthModes {
		if authMode == allowed {
			return authMode, nil
		}
	}
	return "", fmt.Errorf("provider %s does not support auth mode %s", provider, authMode)
}

func DefaultBaseURL(key string) (string, bool) {
	config, ok := FindAPIConfig(key, "")
	if !ok || strings.TrimSpace(config.BaseURL) == "" {
		return "", false
	}
	return config.BaseURL, true
}

func DefaultAPIType(key string) string {
	meta, ok := catalog[key]
	if !ok || strings.TrimSpace(meta.DefaultAPIType) == "" {
		return "openai-completions"
	}
	return meta.DefaultAPIType
}

func ResolveBaseURL(key, apiType, requested string) (string, error) {
	config, ok := FindAPIConfig(key, apiType)
	if !ok {
		return "", fmt.Errorf("provider %s does not support api type %s", key, apiType)
	}
	if !config.EditableBaseURL {
		return strings.TrimRight(config.BaseURL, "/"), nil
	}
	baseURL := strings.TrimSpace(requested)
	if baseURL == "" {
		baseURL = config.BaseURL
	}
	if baseURL == "" {
		return "", fmt.Errorf("base url is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid base url")
	}
	if key == "custom" && (IsImageAPIType(config.APIType) || IsEmbeddingAPIType(config.APIType)) {
		return baseURL, nil
	}
	parsed.Path = normalizeEndpointPath(config.APIType, parsed.Path)
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/"), nil
}

func normalizeEndpointPath(apiType, value string) string {
	if IsImageAPIType(apiType) {
		return strings.TrimRight(value, "/")
	}
	path := strings.TrimRight(value, "/")
	suffixes := []string{}
	switch apiType {
	case "openai-completions":
		suffixes = []string{"/chat/completions"}
	case "openai-responses":
		suffixes = []string{"/responses"}
	case "anthropic-messages":
		suffixes = []string{"/v1/messages", "/messages"}
	case "openai-embeddings":
		suffixes = []string{"/v1/embeddings", "/embeddings"}
	}
	for _, suffix := range suffixes {
		if strings.HasSuffix(strings.ToLower(path), suffix) {
			return path[:len(path)-len(suffix)]
		}
	}
	return path
}

func IsEmbeddingAPIType(apiType string) bool {
	return apiType == "openai-embeddings"
}

func IsImageAPIType(apiType string) bool {
	switch apiType {
	case "openai-images", "dashscope-images", "minimax-images", "openrouter-images":
		return true
	default:
		return false
	}
}

func EnvKey(key string) string {
	meta, ok := catalog[key]
	if !ok {
		return ""
	}
	return meta.EnvKey
}

func DisplayName(key string) string {
	meta, ok := catalog[key]
	if !ok || strings.TrimSpace(meta.DisplayName) == "" {
		return key
	}
	return meta.DisplayName
}

func DisplayNameKey(key string) string {
	meta, ok := catalog[key]
	if !ok {
		return ""
	}
	return meta.DisplayNameKey
}

func NormalizeModelID(provider, modelID string) string {
	target := strings.TrimLeft(strings.TrimSpace(modelID), "/")
	for _, prefix := range legacyModelPrefixes[provider] {
		legacyPrefix := prefix + "/"
		if !strings.HasPrefix(target, legacyPrefix) {
			continue
		}
		candidate := strings.TrimSpace(strings.TrimPrefix(target, legacyPrefix))
		if candidate != "" {
			return candidate
		}
	}
	return target
}

var legacyModelPrefixes = map[string][]string{
	"custom":              {"custom"},
	"vllm":                {"custom"},
	"ollama":              {"ollama"},
	"deepseek":            {"deepseek"},
	"bailian-coding-plan": {"bailian-coding-plan"},
	"ark-coding-plan":     {"ark-coding-plan"},
	"zai":                 {"zai"},
	"minimax":             {"minimax"},
	"xiaomi":              {"xiaomi"},
	"kimi":                {"kimi", "moonshot"},
	"kimi-coding":         {"kimi-coding"},
	"openai":              {"openai"},
	"anthropic":           {"anthropic"},
	"gemini":              {"google", "gemini"},
	"moonshot":            {"moonshot"},
}

func cloneMeta(meta Meta) Meta {
	clone := meta
	clone.APIConfigs = make([]APIConfig, len(meta.APIConfigs))
	for index, config := range meta.APIConfigs {
		clone.APIConfigs[index] = config
		clone.APIConfigs[index].AuthModes = append([]string(nil), config.AuthModes...)
		clone.APIConfigs[index].Models = append([]Model(nil), config.Models...)
	}
	clone.Models = append([]Model(nil), meta.Models...)
	return clone
}
