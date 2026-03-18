package provider

import (
	"strings"
)

type Model struct {
	ID            string
	Name          string
	ContextWindow int
	MaxTokens     int
	Reasoning     bool
	Input         []string
}

type Meta struct {
	Key            string
	DisplayName    string
	Sort           uint
	DefaultBaseURL string
	EnvKey         string
	Models         []Model
	Enabled        bool
}

var catalog = map[string]Meta{
	"custom": {
		Key:            "custom",
		DisplayName:    "Custom",
		Sort:           10,
		DefaultBaseURL: "",
		EnvKey:         "CUSTOM_API_KEY",
		Enabled:        true,
		Models:         []Model{},
	},
	"ollama": {
		Key:         "ollama",
		DisplayName: "Ollama",
		Sort:        15,
		Enabled:     true,
	},
	"vllm": {
		Key:            "vllm",
		DisplayName:    "vLLM",
		Sort:           20,
		DefaultBaseURL: "",
		EnvKey:         "VLLM_API_KEY",
		Enabled:        true,
		Models:         []Model{},
	},
	"deepseek": {
		Key:            "deepseek",
		DisplayName:    "DeepSeek",
		Sort:           25,
		DefaultBaseURL: "https://api.deepseek.com/v1",
		EnvKey:         "DEEPSEEK_API_KEY",
		Enabled:        true,
		Models: []Model{
			{ID: "deepseek/deepseek-chat", Name: "DeepSeek Chat"},
			{ID: "deepseek/deepseek-reasoner", Name: "DeepSeek Reasoner"},
			{ID: "deepseek/deepseek-r1:1.5b", Name: "DeepSeek R1 1.5B"},
		},
	},
	"bailian-coding-plan": {
		Key:            "bailian-coding-plan",
		DisplayName:    "阿里云百炼 Coding Plan",
		Sort:           30,
		DefaultBaseURL: "https://coding.dashscope.aliyuncs.com/v1",
		EnvKey:         "QWEN_API_KEY",
		Enabled:        true,
		Models: []Model{
			{ID: "bailian-coding-plan/qwen3.5-plus", Name: "Qwen3.5-Plus"},
			{ID: "bailian-coding-plan/qwen3-max", Name: "Qwen3-Max"},
			{ID: "bailian-coding-plan/qwen3-coder-next", Name: "Qwen3-Coder-Next"},
			{ID: "bailian-coding-plan/qwen3-coder-plus", Name: "Qwen3-Coder-Plus"},
			{ID: "bailian-coding-plan/minimax-m2.5", Name: "MiniMax M2.5"},
			{ID: "bailian-coding-plan/glm-5", Name: "GLM-5"},
			{ID: "bailian-coding-plan/kimi-k2.5", Name: "Kimi-k2.5"},
			{ID: "bailian-coding-plan/glm-4.7", Name: "GLM-4.7"},
		},
	},
	"ark-coding-plan": {
		Key:            "ark-coding-plan",
		DisplayName:    "方舟 Coding Plan",
		Sort:           35,
		DefaultBaseURL: "https://ark.cn-beijing.volces.com/api/coding/v3",
		EnvKey:         "ARK_API_KEY",
		Enabled:        true,
		Models: []Model{
			{ID: "ark-coding-plan/doubao-seed-2.0-code", Name: "Doubao-Seed-2.0-Code"},
			{ID: "ark-coding-plan/doubao-seed-code", Name: "Doubao-Seed-Code"},
			{ID: "ark-coding-plan/kimi-k2.5", Name: "Kimi-K2.5"},
			{ID: "ark-coding-plan/glm-4.7", Name: "GLM-4.7"},
			{ID: "ark-coding-plan/deepseek-v3.2", Name: "DeepSeek-V3.2"},
			{ID: "ark-coding-plan/kimi-k2-thinking", Name: "Kimi-K2-thinking"},
		},
	},
	"zai": {
		Key:            "zai",
		DisplayName:    "Z.ai",
		Sort:           40,
		DefaultBaseURL: "https://open.bigmodel.cn/api/paas/v4",
		EnvKey:         "ZAI_API_KEY",
		Enabled:        true,
		Models: []Model{
			{ID: "zai/glm-5", Name: "GLM-5"},
			{ID: "zai/glm-4.7", Name: "GLM-4.7"},
			{ID: "zai/glm-4.7-flash", Name: "GLM-4.7-Flash"},
			{ID: "zai/glm-4.7-flashx", Name: "GLM-4.7-FlashX"},
		},
	},
	"minimax": {
		Key:            "minimax",
		DisplayName:    "MiniMax (CN)",
		Sort:           45,
		DefaultBaseURL: "https://api.minimaxi.com/anthropic",
		EnvKey:         "MINIMAX_API_KEY",
		Enabled:        true,
		Models: []Model{
			{ID: "minimax/MiniMax-M2.5", Name: "MiniMax M2.5"},
			{ID: "minimax/MiniMax-M2.5-highspeed", Name: "MiniMax M2.5 highspeed"},
		},
	},
	"kimi": {
		Key:            "kimi",
		DisplayName:    "Kimi (CN)",
		Sort:           50,
		DefaultBaseURL: "https://api.moonshot.cn/v1",
		EnvKey:         "KIMI_API_KEY",
		Enabled:        true,
		Models: []Model{
			{ID: "kimi/kimi-k2.5", Name: "Kimi K2.5"},
			{ID: "kimi/kimi-k2-0905-preview", Name: "Kimi K2 0905 Preview"},
			{ID: "kimi/kimi-k2-thinking", Name: "Kimi K2 Thinking"},
		},
	},
	"kimi-coding": {
		Key:            "kimi-coding",
		DisplayName:    "Kimi Coding",
		Sort:           51,
		DefaultBaseURL: "https://api.kimi.com/coding/",
		EnvKey:         "KIMI_API_KEY",
		Enabled:        true,
		Models: []Model{
			{ID: "kimi-coding/k2p5", Name: "Kimi K2.5"},
		},
	},
	"openai": {
		Key:            "openai",
		DisplayName:    "OpenAI",
		Sort:           55,
		DefaultBaseURL: "https://api.openai.com/v1",
		EnvKey:         "OPENAI_API_KEY",
		Enabled:        true,
		Models: []Model{
			{ID: "openai/codex-mini-latest", Name: "Codex Mini"},
			{ID: "openai/gpt-4.1", Name: "GPT-4.1"},
			{ID: "openai/gpt-4o", Name: "GPT-4o"},
			{ID: "openai/gpt-4o-mini", Name: "GPT-4o Mini"},
			{ID: "openai/gpt-5", Name: "GPT-5"},
			{ID: "openai/gpt-5-mini", Name: "GPT-5 Mini"},
		},
	},
	"openrouter": {
		Key:            "openrouter",
		DisplayName:    "OpenRouter",
		Sort:           56,
		DefaultBaseURL: "https://openrouter.ai/api/v1",
		EnvKey:         "OPENROUTER_API_KEY",
		Enabled:        true,
		Models: []Model{
			{ID: "openrouter/free", Name: "openrouter/free"},
			{ID: "openrouter/auto", Name: "openrouter/auto"},
		},
	},
	"anthropic": {
		Key:            "anthropic",
		DisplayName:    "Anthropic",
		Sort:           60,
		DefaultBaseURL: "https://api.anthropic.com",
		EnvKey:         "ANTHROPIC_API_KEY",
		Enabled:        true,
		Models: []Model{
			{ID: "anthropic/claude-3-haiku-20240307", Name: "Claude 3 Haiku"},
			{ID: "anthropic/claude-3-5-haiku-latest", Name: "Claude 3.5 Haiku"},
			{ID: "anthropic/claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet"},
			{ID: "anthropic/claude-3-7-sonnet-20250219", Name: "Claude 3.7 Sonnet"},
			{ID: "anthropic/claude-opus-4-1", Name: "Claude Opus 4.1"},
		},
	},
	"gemini": {
		Key:            "gemini",
		DisplayName:    "Gemini",
		Sort:           65,
		DefaultBaseURL: "https://generativelanguage.googleapis.com",
		EnvKey:         "GEMINI_API_KEY",
		Enabled:        true,
		Models: []Model{
			{ID: "google/gemini-3-flash-preview", Name: "Gemini 3 Flash Preview"},
			{ID: "google/gemini-flash-latest", Name: "Gemini Flash Latest"},
			{ID: "google/gemini-3-pro-preview", Name: "Gemini 3 Pro Preview"},
		},
	},
	"moonshot": {
		Key:            "moonshot",
		DisplayName:    "Moonshot (Global)",
		Sort:           70,
		DefaultBaseURL: "https://api.moonshot.ai/v1",
		EnvKey:         "MOONSHOT_API_KEY",
		Enabled:        true,
		Models: []Model{
			{ID: "moonshot/kimi-k2.5", Name: "Kimi K2.5"},
			{ID: "moonshot/kimi-k2-0905-preview", Name: "Kimi K2 0905 Preview"},
			{ID: "moonshot/kimi-k2-thinking", Name: "Kimi K2 Thinking"},
		},
	},
}

func Get(key string) (Meta, bool) {
	meta, ok := catalog[strings.ToLower(strings.TrimSpace(key))]
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

func IsEnabled(key string) bool {
	meta, ok := catalog[strings.ToLower(strings.TrimSpace(key))]
	return ok && meta.Enabled
}

func DefaultBaseURL(key string) (string, bool) {
	meta, ok := catalog[strings.ToLower(strings.TrimSpace(key))]
	if !ok || strings.TrimSpace(meta.DefaultBaseURL) == "" {
		return "", false
	}
	return meta.DefaultBaseURL, true
}

func EnvKey(key string) string {
	meta, ok := catalog[strings.ToLower(strings.TrimSpace(key))]
	if !ok {
		return ""
	}
	return meta.EnvKey
}

func DisplayName(key string) string {
	meta, ok := catalog[strings.ToLower(strings.TrimSpace(key))]
	if !ok {
		return key
	}
	if strings.TrimSpace(meta.DisplayName) == "" {
		return key
	}
	return meta.DisplayName
}

func cloneMeta(meta Meta) Meta {
	clone := meta
	if len(meta.Models) > 0 {
		clone.Models = make([]Model, len(meta.Models))
		for i, item := range meta.Models {
			clone.Models[i] = normalizeModel(meta.Key, item)
		}
	}
	return clone
}

func normalizeModel(provider string, model Model) Model {
	clone := model
	clone.ID = strings.TrimSpace(clone.ID)
	clone.Name = strings.TrimSpace(clone.Name)
	if clone.Name == "" {
		clone.Name = clone.ID
	}
	if clone.MaxTokens <= 0 || clone.ContextWindow <= 0 {
		resolvedMaxTokens, resolvedContextWindow := catalogRuntimeDefaults(strings.ToLower(strings.TrimSpace(provider)))
		if clone.MaxTokens <= 0 {
			clone.MaxTokens = resolvedMaxTokens
		}
		if clone.ContextWindow <= 0 {
			clone.ContextWindow = resolvedContextWindow
		}
	}
	if len(clone.Input) == 0 {
		clone.Input = defaultModelInputs(provider)
	}
	if !clone.Reasoning {
		clone.Reasoning = isReasoningModel(clone.ID)
	}
	return clone
}

func defaultModelInputs(provider string) []string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "kimi-coding":
		return []string{"text", "image"}
	default:
		return []string{"text"}
	}
}

func catalogRuntimeDefaults(provider string) (int, int) {
	switch provider {
	case "deepseek":
		return 8192, 128000
	case "zai":
		return 131072, 204800
	case "openrouter":
		return 8192, 128000
	case "minimax", "kimi-coding":
		return 8192, 200000
	case "custom", "vllm":
		return 8192, 128000
	default:
		return 8192, 256000
	}
}
