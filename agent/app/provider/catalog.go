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

type RuntimeDefault struct {
	APIType       string
	ContextWindow int
	MaxTokens     int
	Input         []string
}

type Meta struct {
	Key            string
	DisplayName    string
	Sort           uint
	DefaultBaseURL string
	EnvKey         string
	Default        RuntimeDefault
	Models         []Model
}

var catalog = map[string]Meta{
	"custom": {
		Key:            "custom",
		DisplayName:    "Custom",
		Sort:           10,
		DefaultBaseURL: "",
		EnvKey:         "CUSTOM_API_KEY",
		Default: RuntimeDefault{
			APIType:       "openai-completions",
			ContextWindow: 128000,
			MaxTokens:     8192,
			Input:         []string{"text"},
		},
		Models: []Model{},
	},
	"ollama": {
		Key:         "ollama",
		DisplayName: "Ollama",
		Sort:        15,
		Default: RuntimeDefault{
			APIType:       "openai-responses",
			ContextWindow: 160000,
			MaxTokens:     8192,
			Input:         []string{"text"},
		},
	},
	"vllm": {
		Key:            "vllm",
		DisplayName:    "vLLM",
		Sort:           20,
		DefaultBaseURL: "",
		EnvKey:         "VLLM_API_KEY",
		Default: RuntimeDefault{
			APIType:       "openai-completions",
			ContextWindow: 128000,
			MaxTokens:     8192,
			Input:         []string{"text"},
		},
		Models: []Model{},
	},
	"deepseek": {
		Key:            "deepseek",
		DisplayName:    "DeepSeek",
		Sort:           25,
		DefaultBaseURL: "https://api.deepseek.com/v1",
		EnvKey:         "DEEPSEEK_API_KEY",
		Default: RuntimeDefault{
			APIType:       "openai-completions",
			ContextWindow: 128000,
			MaxTokens:     8192,
			Input:         []string{"text"},
		},
		Models: []Model{
			{ID: "deepseek/deepseek-chat", Name: "DeepSeek Chat"},
			{ID: "deepseek/deepseek-reasoner", Name: "DeepSeek Reasoner", Reasoning: true},
			{ID: "deepseek/deepseek-r1:1.5b", Name: "DeepSeek R1 1.5B", Reasoning: true},
		},
	},
	"bailian-coding-plan": {
		Key:            "bailian-coding-plan",
		DisplayName:    "阿里云百炼 Coding Plan",
		Sort:           30,
		DefaultBaseURL: "https://coding.dashscope.aliyuncs.com/v1",
		EnvKey:         "QWEN_API_KEY",
		Default: RuntimeDefault{
			APIType:       "openai-completions",
			ContextWindow: 256000,
			MaxTokens:     8192,
			Input:         []string{"text"},
		},
		Models: []Model{
			{ID: "bailian-coding-plan/qwen3.5-plus", Name: "Qwen3.5-Plus", Reasoning: true},
			{ID: "bailian-coding-plan/qwen3-max", Name: "Qwen3-Max", Reasoning: true},
			{ID: "bailian-coding-plan/qwen3-coder-next", Name: "Qwen3-Coder-Next", Reasoning: true},
			{ID: "bailian-coding-plan/qwen3-coder-plus", Name: "Qwen3-Coder-Plus", Reasoning: true},
			{ID: "bailian-coding-plan/minimax-m2.5", Name: "MiniMax M2.5", Reasoning: true},
			{ID: "bailian-coding-plan/glm-5", Name: "GLM-5", Reasoning: true},
			{ID: "bailian-coding-plan/kimi-k2.5", Name: "Kimi-k2.5", Reasoning: true},
			{ID: "bailian-coding-plan/glm-4.7", Name: "GLM-4.7", Reasoning: true},
		},
	},
	"ark-coding-plan": {
		Key:            "ark-coding-plan",
		DisplayName:    "方舟 Coding Plan",
		Sort:           35,
		DefaultBaseURL: "https://ark.cn-beijing.volces.com/api/coding/v3",
		EnvKey:         "ARK_API_KEY",
		Default: RuntimeDefault{
			APIType:       "openai-completions",
			ContextWindow: 256000,
			MaxTokens:     8192,
			Input:         []string{"text"},
		},
		Models: []Model{
			{ID: "ark-coding-plan/doubao-seed-2.0-code", Name: "Doubao-Seed-2.0-Code"},
			{ID: "ark-coding-plan/doubao-seed-code", Name: "Doubao-Seed-Code"},
			{ID: "ark-coding-plan/kimi-k2.5", Name: "Kimi-K2.5", Reasoning: true},
			{ID: "ark-coding-plan/glm-4.7", Name: "GLM-4.7", Reasoning: true},
			{ID: "ark-coding-plan/deepseek-v3.2", Name: "DeepSeek-V3.2", Reasoning: true},
			{ID: "ark-coding-plan/kimi-k2-thinking", Name: "Kimi-K2-thinking", Reasoning: true},
		},
	},
	"zai": {
		Key:            "zai",
		DisplayName:    "Z.ai",
		Sort:           40,
		DefaultBaseURL: "https://open.bigmodel.cn/api/paas/v4",
		EnvKey:         "ZAI_API_KEY",
		Default: RuntimeDefault{
			APIType:       "openai-completions",
			ContextWindow: 204800,
			MaxTokens:     131072,
			Input:         []string{"text"},
		},
		Models: []Model{
			{ID: "zai/glm-5", Name: "GLM-5", Reasoning: true},
			{ID: "zai/glm-4.7", Name: "GLM-4.7", Reasoning: true},
			{ID: "zai/glm-4.7-flash", Name: "GLM-4.7-Flash", Reasoning: true},
			{ID: "zai/glm-4.7-flashx", Name: "GLM-4.7-FlashX", Reasoning: true},
		},
	},
	"minimax": {
		Key:            "minimax",
		DisplayName:    "MiniMax (CN)",
		Sort:           45,
		DefaultBaseURL: "https://api.minimaxi.com/anthropic",
		EnvKey:         "MINIMAX_API_KEY",
		Default: RuntimeDefault{
			APIType:       "anthropic-messages",
			ContextWindow: 200000,
			MaxTokens:     8192,
			Input:         []string{"text"},
		},
		Models: []Model{
			{ID: "minimax/MiniMax-M2.7", Name: "MiniMax M2.7"},
			{ID: "minimax/MiniMax-M2.7-highspeed", Name: "MiniMax M2.7 highspeed"},
			{ID: "minimax/MiniMax-M2.5", Name: "MiniMax M2.5", Reasoning: true},
			{ID: "minimax/MiniMax-M2.5-highspeed", Name: "MiniMax M2.5 highspeed"},
		},
	},
	"xiaomi": {
		Key:            "xiaomi",
		DisplayName:    "Xiaomi",
		Sort:           46,
		DefaultBaseURL: "https://api.xiaomimimo.com/anthropic",
		EnvKey:         "XIAOMI_API_KEY",
		Default: RuntimeDefault{
			APIType:       "anthropic-messages",
			ContextWindow: 262144,
			MaxTokens:     8192,
			Input:         []string{"text"},
		},
		Models: []Model{
			{ID: "xiaomi/mimo-v2-pro", Name: "Xiaomi MiMo V2 Pro"},
			{ID: "xiaomi/mimo-v2-omni", Name: "Xiaomi MiMo V2 Omni"},
			{ID: "xiaomi/mimo-v2-tts", Name: "Xiaomi MiMo V2 TTS"},
			{ID: "xiaomi/mimo-v2-flash", Name: "Xiaomi MiMo V2 Flash"},
		},
	},
	"kimi": {
		Key:            "kimi",
		DisplayName:    "Kimi (CN)",
		Sort:           50,
		DefaultBaseURL: "https://api.moonshot.cn/v1",
		EnvKey:         "KIMI_API_KEY",
		Default: RuntimeDefault{
			APIType:       "openai-completions",
			ContextWindow: 256000,
			MaxTokens:     8192,
			Input:         []string{"text", "image"},
		},
		Models: []Model{
			{ID: "kimi/kimi-k2.5", Name: "Kimi K2.5", Reasoning: true},
			{ID: "kimi/kimi-k2-0905-preview", Name: "Kimi K2 0905 Preview"},
			{ID: "kimi/kimi-k2-thinking", Name: "Kimi K2 Thinking", Reasoning: true},
		},
	},
	"kimi-coding": {
		Key:            "kimi-coding",
		DisplayName:    "Kimi Coding",
		Sort:           51,
		DefaultBaseURL: "https://api.kimi.com/coding/",
		EnvKey:         "KIMI_API_KEY",
		Default: RuntimeDefault{
			APIType:       "openai-completions",
			ContextWindow: 262144,
			MaxTokens:     32768,
			Input:         []string{"text", "image"},
		},
		Models: []Model{
			{ID: "kimi-coding/k2p5", Name: "Kimi K2.5", Reasoning: true},
		},
	},
	"openai": {
		Key:            "openai",
		DisplayName:    "OpenAI",
		Sort:           55,
		DefaultBaseURL: "https://api.openai.com/v1",
		EnvKey:         "OPENAI_API_KEY",
		Default: RuntimeDefault{
			APIType:       "openai-completions",
			ContextWindow: 256000,
			MaxTokens:     8192,
			Input:         []string{"text", "image"},
		},
		Models: []Model{
			{ID: "openai/codex-mini-latest", Name: "Codex Mini", Reasoning: true},
			{ID: "openai/gpt-5", Name: "GPT-5", Reasoning: true},
			{ID: "openai/gpt-5-mini", Name: "GPT-5 Mini", Reasoning: true},
			{ID: "openai/gpt-5.4", Name: "GPT-5.4", Reasoning: true},
			{ID: "openai/gpt-5.3-codex", Name: "GPT-5.3-Codex", Reasoning: true},
		},
	},
	"openrouter": {
		Key:            "openrouter",
		DisplayName:    "OpenRouter",
		Sort:           56,
		DefaultBaseURL: "https://openrouter.ai/api/v1",
		EnvKey:         "OPENROUTER_API_KEY",
		Default: RuntimeDefault{
			APIType:       "openai-completions",
			ContextWindow: 128000,
			MaxTokens:     8192,
			Input:         []string{"text"},
		},
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
		Default: RuntimeDefault{
			APIType:       "openai-completions",
			ContextWindow: 256000,
			MaxTokens:     8192,
			Input:         []string{"text", "image"},
		},
		Models: []Model{
			{ID: "anthropic/claude-3-haiku-20240307", Name: "Claude 3 Haiku"},
			{ID: "anthropic/claude-3-5-haiku-latest", Name: "Claude 3.5 Haiku"},
			{ID: "anthropic/claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet"},
			{ID: "anthropic/claude-3-7-sonnet-20250219", Name: "Claude 3.7 Sonnet", Reasoning: true},
			{ID: "anthropic/claude-opus-4-1", Name: "Claude Opus 4.1", Reasoning: true},
		},
	},
	"gemini": {
		Key:            "gemini",
		DisplayName:    "Gemini",
		Sort:           65,
		DefaultBaseURL: "https://generativelanguage.googleapis.com",
		EnvKey:         "GEMINI_API_KEY",
		Default: RuntimeDefault{
			APIType:       "openai-completions",
			ContextWindow: 256000,
			MaxTokens:     8192,
			Input:         []string{"text", "image"},
		},
		Models: []Model{
			{ID: "google/gemini-3-flash-preview", Name: "Gemini 3 Flash Preview", Reasoning: true},
			{ID: "google/gemini-flash-latest", Name: "Gemini Flash Latest"},
			{ID: "google/gemini-3-pro-preview", Name: "Gemini 3 Pro Preview", Reasoning: true},
		},
	},
	"moonshot": {
		Key:            "moonshot",
		DisplayName:    "Moonshot (Global)",
		Sort:           70,
		DefaultBaseURL: "https://api.moonshot.ai/v1",
		EnvKey:         "MOONSHOT_API_KEY",
		Default: RuntimeDefault{
			APIType:       "openai-completions",
			ContextWindow: 256000,
			MaxTokens:     8192,
			Input:         []string{"text"},
		},
		Models: []Model{
			{ID: "moonshot/kimi-k2.5", Name: "Kimi K2.5", Reasoning: true},
			{ID: "moonshot/kimi-k2-0905-preview", Name: "Kimi K2 0905 Preview"},
			{ID: "moonshot/kimi-k2-thinking", Name: "Kimi K2 Thinking", Reasoning: true},
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

func FindModel(key, modelID string) (Model, bool) {
	meta, ok := Get(key)
	if !ok {
		return Model{}, false
	}
	for _, item := range meta.Models {
		if item.ID == modelID {
			return item, true
		}
	}
	return Model{}, false
}

func cloneMeta(meta Meta) Meta {
	clone := meta
	if len(meta.Default.Input) > 0 {
		clone.Default.Input = make([]string, len(meta.Default.Input))
		copy(clone.Default.Input, meta.Default.Input)
	}
	if len(meta.Models) > 0 {
		clone.Models = make([]Model, len(meta.Models))
		for i, item := range meta.Models {
			clone.Models[i] = normalizeModel(meta, item)
		}
	}
	return clone
}

func normalizeModel(meta Meta, model Model) Model {
	clone := model
	clone.ID = strings.TrimSpace(clone.ID)
	clone.Name = strings.TrimSpace(clone.Name)
	if clone.Name == "" {
		clone.Name = clone.ID
	}
	if clone.MaxTokens <= 0 {
		clone.MaxTokens = meta.Default.MaxTokens
	}
	if clone.ContextWindow <= 0 {
		clone.ContextWindow = meta.Default.ContextWindow
	}
	if len(clone.Input) == 0 && len(meta.Default.Input) > 0 {
		clone.Input = make([]string, len(meta.Default.Input))
		copy(clone.Input, meta.Default.Input)
	}
	return clone
}

func ResolveRuntimeParams(provider, apiType string, maxTokens, contextWindow int) (string, int, int) {
	defaultAPIType := "openai-completions"
	defaultMaxTokens := 8192
	defaultContextWindow := 256000
	if meta, ok := Get(provider); ok {
		if meta.Default.APIType != "" {
			defaultAPIType = meta.Default.APIType
		}
		if meta.Default.MaxTokens > 0 {
			defaultMaxTokens = meta.Default.MaxTokens
		}
		if meta.Default.ContextWindow > 0 {
			defaultContextWindow = meta.Default.ContextWindow
		}
	}
	resolvedAPI := apiType
	if strings.TrimSpace(apiType) == "" {
		resolvedAPI = defaultAPIType
	}
	resolvedMaxTokens := defaultMaxTokens
	resolvedContextWindow := defaultContextWindow
	if maxTokens > 0 {
		resolvedMaxTokens = maxTokens
	}
	if contextWindow > 0 {
		resolvedContextWindow = contextWindow
	}
	return resolvedAPI, resolvedMaxTokens, resolvedContextWindow
}
