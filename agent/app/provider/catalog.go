package provider

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/aihttp"
)

type APIConfig struct {
	Route             string   `json:"route"`
	DiscoveryRoute    string   `json:"discoveryRoute"`
	DiscoveryAuthMode string   `json:"discoveryAuthMode"`
	Purpose           string   `json:"purpose"`
	Discovery         string   `json:"discovery"`
	APIType           string   `json:"apiType"`
	BaseURL           string   `json:"baseUrl"`
	EditableBaseURL   bool     `json:"editableBaseUrl"`
	DiscoverModels    bool     `json:"supportsModelDiscovery"`
	DefaultAuthMode   string   `json:"defaultAuthMode"`
	AuthModes         []string `json:"authModes"`
}

const (
	AuthModeBearer  = "bearer"
	AuthModeXAPIKey = "x-api-key"
)

type Meta struct {
	AccountTypes   []AccountTypePreset
	Key            string
	DisplayName    string
	DisplayNameKey string
	Sort           uint
	DefaultAPIType string
	EnvKey         string
}

var catalog = map[string]Meta{
	"custom": {
		Key:            "custom",
		DisplayName:    "Custom",
		Sort:           10,
		DefaultAPIType: "openai-completions",
		EnvKey:         "CUSTOM_API_KEY",
		AccountTypes: []AccountTypePreset{
			{
				Key: "default",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-responses", Purpose: "text", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key", "bearer"}},
					{APIType: "openai-images", Purpose: "image", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "dashscope-images", Purpose: "image", Discovery: "DiscoveryNone", EditableBaseURL: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-embeddings", Purpose: "embedding", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "rerank", Purpose: "rerank", Discovery: "DiscoveryNone", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-transcriptions", Purpose: "transcription", Route: "/v1", Discovery: "DiscoveryNone", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-speech", Purpose: "speech", Route: "/v1", Discovery: "DiscoveryNone", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
		},
	},
	"ollama": {
		Key:            "ollama",
		DisplayName:    "Ollama",
		Sort:           40,
		DefaultAPIType: "openai-responses",
		AccountTypes: []AccountTypePreset{
			{
				Key: "default",
				APIConfigs: []APIConfig{
					{APIType: "openai-responses", Purpose: "text", Route: "/v1", Discovery: "DiscoveryOllamaModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-completions", Purpose: "text", Route: "/v1", Discovery: "DiscoveryOllamaModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", Discovery: "DiscoveryOllamaModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key"}},
					{APIType: "openai-images", Purpose: "image", Route: "/v1", Discovery: "DiscoveryOllamaModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-embeddings", Purpose: "embedding", Route: "/v1", Discovery: "DiscoveryOllamaModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
		},
	},
	"vllm": {
		Key:            "vllm",
		DisplayName:    "vLLM",
		Sort:           20,
		DefaultAPIType: "openai-completions",
		EnvKey:         "VLLM_API_KEY",
		AccountTypes: []AccountTypePreset{
			{
				Key: "default",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-responses", Purpose: "text", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer", "x-api-key"}},
					{APIType: "openai-images", Purpose: "image", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-embeddings", Purpose: "embedding", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "rerank", Purpose: "rerank", Route: "/v1", Discovery: "DiscoveryNone", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-transcriptions", Purpose: "transcription", Route: "/v1", Discovery: "DiscoveryNone", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
		},
	},
	"1panel-ai-gateway": {
		Key:            "1panel-ai-gateway",
		DisplayName:    "1Panel AI Gateway",
		DisplayNameKey: "AIProvider1PanelGateway",
		Sort:           30,
		DefaultAPIType: "openai-completions",
		EnvKey:         "ONEPANEL_AI_GATEWAY_API_KEY",
		AccountTypes: []AccountTypePreset{
			{
				Key: "default",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", EditableBaseURL: true, DiscoverModels: true},
					{APIType: "openai-responses", Purpose: "text", EditableBaseURL: true, DiscoverModels: true},
					{APIType: "anthropic-messages", Purpose: "text", EditableBaseURL: true, DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key", "bearer"}},
					{APIType: "openai-images", Purpose: "image", EditableBaseURL: true},
					{APIType: "openai-embeddings", Purpose: "embedding", EditableBaseURL: true},
				},
			},
		},
	},
	"deepseek": {
		Key:            "deepseek",
		DisplayName:    "DeepSeek",
		Sort:           60,
		DefaultAPIType: "openai-completions",
		EnvKey:         "DEEPSEEK_API_KEY",
		AccountTypes: []AccountTypePreset{
			{
				Key: "default",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://api.deepseek.com", Discovery: "DiscoveryOpenAIModels", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-responses", Purpose: "text", BaseURL: "https://api.deepseek.com", Discovery: "DiscoveryOpenAIModels", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://api.deepseek.com/anthropic", Route: "/anthropic", Discovery: "DiscoveryOpenAIModels", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key"}},
				},
			},
		},
	},
	"bailian-coding-plan": {
		Key:            "bailian-coding-plan",
		DisplayNameKey: "AIProviderBailianCodingPlan",
		Sort:           210,
		DefaultAPIType: "openai-completions",
		EnvKey:         "QWEN_API_KEY",
		AccountTypes: []AccountTypePreset{
			{
				Key: "codingPlan",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://coding.dashscope.aliyuncs.com/v1", Route: "/v1", Discovery: "DiscoveryNone", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://coding.dashscope.aliyuncs.com/apps/anthropic", Route: "/apps/anthropic", Discovery: "DiscoveryNone", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
		},
	},
	"bailian": {
		Key:            "bailian",
		DisplayNameKey: "AIProviderBailian",
		Sort:           80,
		DefaultAPIType: "openai-completions",
		EnvKey:         "DASHSCOPE_API_KEY",
		AccountTypes: []AccountTypePreset{
			{
				Key: "codingPlan",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://coding.dashscope.aliyuncs.com/v1", Route: "/v1", Discovery: "DiscoveryNone", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://coding.dashscope.aliyuncs.com/apps/anthropic", Route: "/apps/anthropic", Discovery: "DiscoveryNone", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
			{
				Key: "tokenPlan",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://token-plan.cn-beijing.maas.aliyuncs.com/compatible-mode/v1", Route: "/compatible-mode/v1", Discovery: "DiscoveryNone", DiscoveryRoute: "/compatible-mode/v1", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-responses", Purpose: "text", BaseURL: "https://token-plan.cn-beijing.maas.aliyuncs.com/compatible-mode/v1", Route: "/compatible-mode/v1", Discovery: "DiscoveryNone", DiscoveryRoute: "/compatible-mode/v1", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://token-plan.cn-beijing.maas.aliyuncs.com/apps/anthropic", Route: "/apps/anthropic", Discovery: "DiscoveryNone", DiscoveryRoute: "/compatible-mode/v1", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "dashscope-transcriptions", Purpose: "transcription", BaseURL: "https://token-plan.cn-beijing.maas.aliyuncs.com", Discovery: "DiscoveryNone", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "dashscope-speech", Purpose: "speech", BaseURL: "https://token-plan.cn-beijing.maas.aliyuncs.com", Discovery: "DiscoveryNone", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
			{
				Key: "payAsYouGo",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1", Route: "/compatible-mode/v1", Discovery: "DiscoveryDashScopeModels", DiscoveryRoute: "/compatible-mode/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-responses", Purpose: "text", BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1", Route: "/compatible-mode/v1", Discovery: "DiscoveryDashScopeModels", DiscoveryRoute: "/compatible-mode/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://dashscope.aliyuncs.com/apps/anthropic", Route: "/apps/anthropic", Discovery: "DiscoveryDashScopeModels", DiscoveryRoute: "/compatible-mode/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "dashscope-images", Purpose: "image", BaseURL: "https://dashscope.aliyuncs.com", Discovery: "DiscoveryDashScopeModels", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-embeddings", Purpose: "embedding", BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1", Route: "/compatible-mode/v1", Discovery: "DiscoveryDashScopeModels", DiscoveryRoute: "/compatible-mode/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "rerank", Purpose: "rerank", BaseURL: "https://dashscope.aliyuncs.com", Discovery: "DiscoveryNone", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "dashscope-transcriptions", Purpose: "transcription", BaseURL: "https://dashscope.aliyuncs.com", Discovery: "DiscoveryDashScopeModels", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "dashscope-speech", Purpose: "speech", BaseURL: "https://dashscope.aliyuncs.com", Discovery: "DiscoveryDashScopeModels", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
		},
	},
	"ark-coding-plan": {
		Key:            "ark-coding-plan",
		DisplayNameKey: "AIProviderArkCodingPlan",
		Sort:           220,
		DefaultAPIType: "openai-completions",
		EnvKey:         "ARK_API_KEY",
		AccountTypes: []AccountTypePreset{
			{
				Key: "codingPlan",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://ark.cn-beijing.volces.com/api/coding/v3", Route: "/api/coding/v3", Discovery: "DiscoveryNone", DiscoveryRoute: "/api/coding/v3", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-responses", Purpose: "text", BaseURL: "https://ark.cn-beijing.volces.com/api/coding/v3", Route: "/api/coding/v3", Discovery: "DiscoveryNone", DiscoveryRoute: "/api/coding/v3", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://ark.cn-beijing.volces.com/api/coding", Route: "/api/coding", Discovery: "DiscoveryNone", DiscoveryRoute: "/api/coding/v3", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-embeddings", Purpose: "embedding", BaseURL: "https://ark.cn-beijing.volces.com/api/coding/v3", Route: "/api/coding/v3", Discovery: "DiscoveryNone", DiscoveryRoute: "/api/coding/v3", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
		},
	},
	"ark": {
		Key:            "ark",
		DisplayNameKey: "AIProviderArk",
		Sort:           110,
		DefaultAPIType: "openai-completions",
		EnvKey:         "ARK_API_KEY",
		AccountTypes: []AccountTypePreset{
			{
				Key: "codingPlan",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://ark.cn-beijing.volces.com/api/coding/v3", Route: "/api/coding/v3", Discovery: "DiscoveryNone", DiscoveryRoute: "/api/coding/v3", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-responses", Purpose: "text", BaseURL: "https://ark.cn-beijing.volces.com/api/coding/v3", Route: "/api/coding/v3", Discovery: "DiscoveryNone", DiscoveryRoute: "/api/coding/v3", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://ark.cn-beijing.volces.com/api/coding", Route: "/api/coding", Discovery: "DiscoveryNone", DiscoveryRoute: "/api/coding/v3", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-embeddings", Purpose: "embedding", BaseURL: "https://ark.cn-beijing.volces.com/api/coding/v3", Route: "/api/coding/v3", Discovery: "DiscoveryNone", DiscoveryRoute: "/api/coding/v3", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
			{
				Key: "agentPlan",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://ark.cn-beijing.volces.com/api/plan/v3", Route: "/api/plan/v3", Discovery: "DiscoveryNone", DiscoveryRoute: "/api/plan/v3", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-responses", Purpose: "text", BaseURL: "https://ark.cn-beijing.volces.com/api/plan/v3", Route: "/api/plan/v3", Discovery: "DiscoveryNone", DiscoveryRoute: "/api/plan/v3", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://ark.cn-beijing.volces.com/api/plan", Route: "/api/plan", Discovery: "DiscoveryNone", DiscoveryRoute: "/api/plan/v3", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-images", Purpose: "image", BaseURL: "https://ark.cn-beijing.volces.com/api/plan/v3", Route: "/api/plan/v3", Discovery: "DiscoveryNone", DiscoveryRoute: "/api/plan/v3", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-embeddings", Purpose: "embedding", BaseURL: "https://ark.cn-beijing.volces.com/api/plan/v3", Route: "/api/plan/v3", Discovery: "DiscoveryNone", DiscoveryRoute: "/api/plan/v3", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
			{
				Key: "payAsYouGo",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://ark.cn-beijing.volces.com/api/v3", Route: "/api/v3", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/v3", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-responses", Purpose: "text", BaseURL: "https://ark.cn-beijing.volces.com/api/v3", Route: "/api/v3", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/v3", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-images", Purpose: "image", BaseURL: "https://ark.cn-beijing.volces.com/api/v3", Route: "/api/v3", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/v3", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-embeddings", Purpose: "embedding", BaseURL: "https://ark.cn-beijing.volces.com/api/v3", Route: "/api/v3", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/v3", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
			{
				Key: "speech",
				APIConfigs: []APIConfig{
					{APIType: "doubao-transcriptions", Purpose: "transcription", BaseURL: "https://openspeech.bytedance.com", Discovery: "DiscoveryNone", DiscoveryAuthMode: "x-api-key", DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key"}},
					{APIType: "doubao-speech", Purpose: "speech", BaseURL: "https://openspeech.bytedance.com", Discovery: "DiscoveryNone", DiscoveryAuthMode: "x-api-key", DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key"}},
				},
			},
		},
	},
	"zai": {
		Key:            "zai",
		DisplayName:    "Z.ai",
		Sort:           120,
		DefaultAPIType: "openai-completions",
		EnvKey:         "ZAI_API_KEY",
		AccountTypes: []AccountTypePreset{
			{
				Key: "chinaCodingPlan",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://open.bigmodel.cn/api/coding/paas/v4", Route: "/api/coding/paas/v4", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/coding/paas/v4", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://open.bigmodel.cn/api/anthropic", Route: "/api/anthropic", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/coding/paas/v4", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key"}},
				},
			},
			{
				Key: "globalCodingPlan",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://api.z.ai/api/coding/paas/v4", Route: "/api/coding/paas/v4", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/coding/paas/v4", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://api.z.ai/api/anthropic", Route: "/api/anthropic", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/coding/paas/v4", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key"}},
				},
			},
			{
				Key: "chinaApi",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://open.bigmodel.cn/api/paas/v4", Route: "/api/paas/v4", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/paas/v4", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://open.bigmodel.cn/api/anthropic", Route: "/api/anthropic", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/paas/v4", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key"}},
					{APIType: "openai-images", Purpose: "image", BaseURL: "https://open.bigmodel.cn/api/paas/v4", Route: "/api/paas/v4", Discovery: "DiscoveryNone", DiscoveryRoute: "/api/paas/v4", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "rerank", Purpose: "rerank", BaseURL: "https://open.bigmodel.cn/api/paas/v4", Route: "/api/paas/v4", Discovery: "DiscoveryNone", DiscoveryRoute: "/api/paas/v4", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
			{
				Key: "globalApi",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://api.z.ai/api/paas/v4", Route: "/api/paas/v4", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/paas/v4", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://api.z.ai/api/anthropic", Route: "/api/anthropic", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/paas/v4", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key"}},
					{APIType: "openai-images", Purpose: "image", BaseURL: "https://api.z.ai/api/paas/v4", Route: "/api/paas/v4", Discovery: "DiscoveryNone", DiscoveryRoute: "/api/paas/v4", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
		},
	},
	"minimax": {
		Key:            "minimax",
		DisplayName:    "MiniMax (CN)",
		Sort:           130,
		DefaultAPIType: "anthropic-messages",
		EnvKey:         "MINIMAX_API_KEY",
		AccountTypes: []AccountTypePreset{
			{
				Key: "default",
				APIConfigs: []APIConfig{
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://api.minimaxi.com/anthropic", Route: "/anthropic", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/anthropic", DiscoveryAuthMode: "x-api-key", DiscoverModels: true, DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key", "bearer"}},
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://api.minimaxi.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/anthropic", DiscoveryAuthMode: "x-api-key", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-responses", Purpose: "text", BaseURL: "https://api.minimaxi.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/anthropic", DiscoveryAuthMode: "x-api-key", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "minimax-images", Purpose: "image", BaseURL: "https://api.minimaxi.com", Discovery: "DiscoveryNone", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "minimax-speech", Purpose: "speech", BaseURL: "https://api.minimaxi.com", Discovery: "DiscoveryNone", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
		},
	},
	"xiaomi": {
		Key:            "xiaomi",
		DisplayName:    "Xiaomi",
		Sort:           140,
		DefaultAPIType: "openai-completions",
		EnvKey:         "XIAOMI_API_KEY",
		AccountTypes: []AccountTypePreset{
			{
				Key: "tokenPlan",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://token-plan-cn.xiaomimimo.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-responses", Purpose: "text", BaseURL: "https://token-plan-cn.xiaomimimo.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://token-plan-cn.xiaomimimo.com/anthropic", Route: "/anthropic", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", EditableBaseURL: true, DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
			{
				Key: "payAsYouGo",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://api.xiaomimimo.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-responses", Purpose: "text", BaseURL: "https://api.xiaomimimo.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://api.xiaomimimo.com/anthropic", Route: "/anthropic", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "mimo-transcriptions", Purpose: "transcription", BaseURL: "https://api.xiaomimimo.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "mimo-speech", Purpose: "speech", BaseURL: "https://api.xiaomimimo.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
		},
	},
	"kimi": {
		Key:            "kimi",
		DisplayName:    "Kimi (CN)",
		Sort:           150,
		DefaultAPIType: "openai-completions",
		EnvKey:         "KIMI_API_KEY",
		AccountTypes: []AccountTypePreset{
			{
				Key: "chinaTokenPlan",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://api.kimi.com/coding/v1", Route: "/coding/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/coding/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://api.kimi.com/coding", Route: "/coding", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/coding/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key", "bearer"}},
				},
			},
			{
				Key: "globalTokenPlan",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://api.kimi.com/coding/v1", Route: "/coding/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/coding/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://api.kimi.com/coding", Route: "/coding", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/coding/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key", "bearer"}},
				},
			},
			{
				Key: "chinaApi",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://api.moonshot.cn/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://api.moonshot.cn/anthropic", Route: "/anthropic", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
			{
				Key: "globalApi",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://api.moonshot.ai/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://api.moonshot.ai/anthropic", Route: "/anthropic", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
		},
	},
	"kimi-coding": {
		Key:            "kimi-coding",
		DisplayName:    "Kimi Coding",
		Sort:           230,
		DefaultAPIType: "anthropic-messages",
		EnvKey:         "KIMI_API_KEY",
		AccountTypes: []AccountTypePreset{
			{
				Key: "chinaTokenPlan",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://api.kimi.com/coding/v1", Route: "/coding/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/coding/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://api.kimi.com/coding", Route: "/coding", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/coding/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key", "bearer"}},
				},
			},
		},
	},
	"openai": {
		Key:            "openai",
		DisplayName:    "OpenAI",
		Sort:           170,
		DefaultAPIType: "openai-responses",
		EnvKey:         "OPENAI_API_KEY",
		AccountTypes: []AccountTypePreset{
			{
				Key: "default",
				APIConfigs: []APIConfig{
					{APIType: "openai-responses", Purpose: "text", BaseURL: "https://api.openai.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://api.openai.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-images", Purpose: "image", BaseURL: "https://api.openai.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-embeddings", Purpose: "embedding", BaseURL: "https://api.openai.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-transcriptions", Purpose: "transcription", BaseURL: "https://api.openai.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-speech", Purpose: "speech", BaseURL: "https://api.openai.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
		},
	},
	"openrouter": {
		Key:            "openrouter",
		DisplayName:    "OpenRouter",
		Sort:           180,
		DefaultAPIType: "openai-completions",
		EnvKey:         "OPENROUTER_API_KEY",
		AccountTypes: []AccountTypePreset{
			{
				Key: "default",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://openrouter.ai/api/v1", Route: "/api/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-responses", Purpose: "text", BaseURL: "https://openrouter.ai/api/v1", Route: "/api/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://openrouter.ai/api", Route: "/api", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openrouter-images", Purpose: "image", BaseURL: "https://openrouter.ai", Discovery: "DiscoveryOpenAIModels", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-embeddings", Purpose: "embedding", BaseURL: "https://openrouter.ai/api/v1", Route: "/api/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "rerank", Purpose: "rerank", BaseURL: "https://openrouter.ai/api/v1", Route: "/api/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-transcriptions", Purpose: "transcription", BaseURL: "https://openrouter.ai/api/v1", Route: "/api/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-speech", Purpose: "speech", BaseURL: "https://openrouter.ai/api/v1", Route: "/api/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/api/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
		},
	},
	"anthropic": {
		Key:            "anthropic",
		DisplayName:    "Anthropic",
		Sort:           190,
		DefaultAPIType: "anthropic-messages",
		EnvKey:         "ANTHROPIC_API_KEY",
		AccountTypes: []AccountTypePreset{
			{
				Key: "default",
				APIConfigs: []APIConfig{
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://api.anthropic.com", Discovery: "DiscoveryAnthropicModels", DiscoveryAuthMode: "x-api-key", DiscoverModels: true, DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key"}},
				},
			},
		},
	},
	"gemini": {
		Key:            "gemini",
		DisplayName:    "Gemini",
		Sort:           200,
		DefaultAPIType: "gemini-generate-content",
		EnvKey:         "GEMINI_API_KEY",
		AccountTypes: []AccountTypePreset{
			{
				Key: "default",
				APIConfigs: []APIConfig{
					{APIType: "gemini-generate-content", Purpose: "text", BaseURL: "https://generativelanguage.googleapis.com", Discovery: "DiscoveryGeminiModels", DiscoveryAuthMode: "x-goog-api-key", DiscoverModels: true, DefaultAuthMode: "x-goog-api-key", AuthModes: []string{"x-goog-api-key"}},
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai", Route: "/v1beta/openai", Discovery: "DiscoveryGeminiModels", DiscoveryAuthMode: "x-goog-api-key", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "gemini-speech", Purpose: "speech", BaseURL: "https://generativelanguage.googleapis.com", Discovery: "DiscoveryGeminiModels", DiscoveryAuthMode: "x-goog-api-key", DiscoverModels: true, DefaultAuthMode: "x-goog-api-key", AuthModes: []string{"x-goog-api-key"}},
				},
			},
		},
	},
	"moonshot": {
		Key:            "moonshot",
		DisplayName:    "Moonshot (Global)",
		Sort:           240,
		DefaultAPIType: "openai-completions",
		EnvKey:         "MOONSHOT_API_KEY",
		AccountTypes: []AccountTypePreset{
			{
				Key: "globalApi",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://api.moonshot.ai/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://api.moonshot.ai/anthropic", Route: "/anthropic", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
		},
	},
	"llmman": {
		Key:            "llmman",
		DisplayName:    "llmman",
		Sort:           250,
		DefaultAPIType: "openai-responses",
		AccountTypes: []AccountTypePreset{
			{
				Key: "default",
				APIConfigs: []APIConfig{
					{APIType: "openai-responses", Purpose: "text", BaseURL: "http://127.0.0.1:17434/v1", EditableBaseURL: true, DiscoverModels: true},
					{APIType: "openai-completions", Purpose: "text", BaseURL: "http://127.0.0.1:17434/v1", EditableBaseURL: true, DiscoverModels: true},
					{APIType: "openai-embeddings", Purpose: "embedding", BaseURL: "http://127.0.0.1:17434/v1", EditableBaseURL: true},
				},
			},
		},
	},
	"tokenhub": {
		Key:            "tokenhub",
		DisplayName:    "TokenHub",
		Sort:           70,
		DefaultAPIType: "openai-completions",
		AccountTypes: []AccountTypePreset{
			{
				Key: "chinaTokenPlan",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://tokenhub.tencentmaas.com/plan/v3", Route: "/v3", Discovery: "DiscoveryNone", DiscoveryRoute: "/v3", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://tokenhub.tencentmaas.com/plan/anthropic", Route: "/anthropic", Discovery: "DiscoveryNone", DiscoveryRoute: "/v3", DiscoveryAuthMode: "bearer", DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key"}},
				},
			},
			{
				Key: "globalTokenPlan",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://tokenhub-intl.tencentmaas.com/plan/v3", Route: "/v3", Discovery: "DiscoveryNone", DiscoveryRoute: "/v3", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://tokenhub-intl.tencentmaas.com/plan/anthropic", Route: "/anthropic", Discovery: "DiscoveryNone", DiscoveryRoute: "/v3", DiscoveryAuthMode: "bearer", DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key"}},
				},
			},
			{
				Key: "chinaApi",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://tokenhub.tencentmaas.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-responses", Purpose: "text", BaseURL: "https://tokenhub.tencentmaas.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://tokenhub.tencentmaas.com", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key"}},
					{APIType: "tokenhub-images", Purpose: "image", BaseURL: "https://tokenhub.tencentmaas.com", Discovery: "DiscoveryOpenAIModels", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-embeddings", Purpose: "embedding", BaseURL: "https://tokenhub.tencentmaas.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "tokenhub-transcriptions", Purpose: "transcription", BaseURL: "https://tokenhub.tencentmaas.com", Discovery: "DiscoveryOpenAIModels", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "tokenhub-speech", Purpose: "speech", BaseURL: "https://tokenhub.tencentmaas.com", Discovery: "DiscoveryOpenAIModels", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
			{
				Key: "globalApi",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://tokenhub-intl.tencentmaas.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-responses", Purpose: "text", BaseURL: "https://tokenhub-intl.tencentmaas.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://tokenhub-intl.tencentmaas.com", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key"}},
					{APIType: "openai-embeddings", Purpose: "embedding", BaseURL: "https://tokenhub-intl.tencentmaas.com/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
		},
	},
	"qianfan": {
		Key:            "qianfan",
		DisplayName:    "百度千帆",
		Sort:           90,
		DefaultAPIType: "openai-completions",
		AccountTypes: []AccountTypePreset{
			{
				Key: "codingPlan",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://qianfan.baidubce.com/v2/coding", Route: "/v2/coding", Discovery: "DiscoveryNone", DiscoveryRoute: "/v2/coding", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://qianfan.baidubce.com/anthropic/coding", Route: "/anthropic/coding", Discovery: "DiscoveryNone", DiscoveryRoute: "/v2/coding", DiscoveryAuthMode: "bearer", DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key"}},
				},
			},
			{
				Key: "tokenPlan",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://qianfan.baidubce.com/v2", Route: "/v2", Discovery: "DiscoveryNone", DiscoveryRoute: "/v2", DiscoveryAuthMode: "bearer", DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://qianfan.baidubce.com/anthropic", Route: "/anthropic", Discovery: "DiscoveryNone", DiscoveryRoute: "/v2", DiscoveryAuthMode: "bearer", DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key"}},
				},
			},
			{
				Key: "payAsYouGo",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://qianfan.baidubce.com/v2", Route: "/v2", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v2", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-responses", Purpose: "text", BaseURL: "https://qianfan.baidubce.com/v2", Route: "/v2", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v2", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://qianfan.baidubce.com/anthropic", Route: "/anthropic", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v2", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key"}},
					{APIType: "openai-images", Purpose: "image", BaseURL: "https://qianfan.baidubce.com/v2", Route: "/v2", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v2", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-embeddings", Purpose: "embedding", BaseURL: "https://qianfan.baidubce.com/v2", Route: "/v2", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v2", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "rerank", Purpose: "rerank", BaseURL: "https://qianfan.baidubce.com/v2", Route: "/v2", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v2", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
		},
	},
	"siliconflow": {
		Key:            "siliconflow",
		DisplayName:    "SiliconFlow",
		Sort:           100,
		DefaultAPIType: "openai-completions",
		AccountTypes: []AccountTypePreset{
			{
				Key: "default",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://api.siliconflow.cn/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://api.siliconflow.cn", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "siliconflow-images", Purpose: "image", BaseURL: "https://api.siliconflow.cn/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-embeddings", Purpose: "embedding", BaseURL: "https://api.siliconflow.cn/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "rerank", Purpose: "rerank", BaseURL: "https://api.siliconflow.cn/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-transcriptions", Purpose: "transcription", BaseURL: "https://api.siliconflow.cn/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-speech", Purpose: "speech", BaseURL: "https://api.siliconflow.cn/v1", Route: "/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
				},
			},
		},
	},
	"opencode": {
		Key:            "opencode",
		DisplayName:    "OpenCode",
		Sort:           160,
		DefaultAPIType: "openai-completions",
		AccountTypes: []AccountTypePreset{
			{
				Key: "go",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://opencode.ai/zen/go/v1", Route: "/zen/go/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/zen/go/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-responses", Purpose: "text", BaseURL: "https://opencode.ai/zen/go/v1", Route: "/zen/go/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/zen/go/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://opencode.ai/zen/go", Route: "/zen/go", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/zen/go/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key"}},
				},
			},
			{
				Key: "zen",
				APIConfigs: []APIConfig{
					{APIType: "openai-completions", Purpose: "text", BaseURL: "https://opencode.ai/zen/v1", Route: "/zen/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/zen/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "openai-responses", Purpose: "text", BaseURL: "https://opencode.ai/zen/v1", Route: "/zen/v1", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/zen/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "bearer", AuthModes: []string{"bearer"}},
					{APIType: "anthropic-messages", Purpose: "text", BaseURL: "https://opencode.ai/zen", Route: "/zen", Discovery: "DiscoveryOpenAIModels", DiscoveryRoute: "/zen/v1", DiscoveryAuthMode: "bearer", DiscoverModels: true, DefaultAuthMode: "x-api-key", AuthModes: []string{"x-api-key"}},
				},
			},
		},
	},
}

func Get(key string) (Meta, bool) {
	meta, ok := catalog[key]
	if !ok {
		return Meta{}, false
	}
	return meta, true
}

func All() map[string]Meta {
	result := make(map[string]Meta, len(catalog))
	for key, meta := range catalog {
		result[key] = meta
	}
	return result
}

func FindAPIConfig(key, apiType string, accountType ...string) (APIConfig, bool) {
	meta, ok := catalog[key]
	if !ok {
		return APIConfig{}, false
	}
	target := strings.TrimSpace(apiType)
	if target == "" {
		target = meta.DefaultAPIType
	}
	presets := meta.AccountTypes
	if len(accountType) > 0 {
		preset, ok := FindAccountType(key, accountType[0])
		if !ok {
			return APIConfig{}, false
		}
		presets = []AccountTypePreset{preset}
	}
	for _, preset := range presets {
		for _, config := range preset.APIConfigs {
			if config.APIType == target {
				config.AuthModes = append([]string(nil), config.AuthModes...)
				return config, true
			}
		}
	}
	return APIConfig{}, false
}

func ResolveAuthMode(provider, apiType, requested string, accountType ...string) (string, error) {
	config, ok := FindAPIConfig(provider, apiType, accountType...)
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

func ResolveBaseURL(key, apiType, requested string, accountType ...string) (string, error) {
	config, ok := FindAPIConfig(key, apiType, accountType...)
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
	if err != nil || aihttp.ValidateURL(baseURL) != nil {
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
	case "openai-images", "dashscope-images", "minimax-images", "openrouter-images", "tokenhub-images", "siliconflow-images":
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
	"llmman":              {"llmman"},
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
