package provider

import "strings"

type Model struct {
	ID   string
	Name string
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
	"ollama": {
		Key:         "ollama",
		DisplayName: "Ollama",
		Sort:        1,
		Enabled:     true,
	},
	"deepseek": {
		Key:            "deepseek",
		DisplayName:    "DeepSeek",
		Sort:           2,
		DefaultBaseURL: "https://api.deepseek.com/v1",
		EnvKey:         "DEEPSEEK_API_KEY",
		Enabled:        true,
		Models: []Model{
			{ID: "deepseek/deepseek-chat", Name: "DeepSeek Chat"},
			{ID: "deepseek/deepseek-reasoner", Name: "DeepSeek Reasoner"},
			{ID: "deepseek/deepseek-r1:1.5b", Name: "DeepSeek R1 1.5B"},
		},
	},
	"openai": {
		Key:            "openai",
		DisplayName:    "OpenAI",
		Sort:           3,
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
	"anthropic": {
		Key:            "anthropic",
		DisplayName:    "Anthropic",
		Sort:           4,
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
		Sort:           5,
		DefaultBaseURL: "https://generativelanguage.googleapis.com",
		EnvKey:         "GEMINI_API_KEY",
		Enabled:        true,
		Models: []Model{
			{ID: "google/gemini-1.5-flash", Name: "Gemini 1.5 Flash"},
			{ID: "google/gemini-1.5-pro", Name: "Gemini 1.5 Pro"},
			{ID: "google/gemini-2.0-flash", Name: "Gemini 2.0 Flash"},
			{ID: "google/gemini-2.5-flash", Name: "Gemini 2.5 Flash"},
			{ID: "google/gemini-2.5-pro", Name: "Gemini 2.5 Pro"},
			{ID: "google/gemini-3-flash-preview", Name: "Gemini 3 Flash Preview"},
		},
	},
	"minimax": {
		Key:            "minimax",
		DisplayName:    "MiniMax (CN)",
		Sort:           6,
		DefaultBaseURL: "https://api.minimaxi.com/anthropic",
		EnvKey:         "MINIMAX_API_KEY",
		Enabled:        true,
		Models: []Model{
			{ID: "minimax/MiniMax-M2.1", Name: "MiniMax M2.1"},
			{ID: "minimax/MiniMax-M2.1-lightning", Name: "MiniMax M2.1 Lightning"},
		},
	},
	"moonshot": {
		Key:            "moonshot",
		DisplayName:    "Moonshot (Global)",
		Sort:           7,
		DefaultBaseURL: "https://api.moonshot.ai/v1",
		EnvKey:         "MOONSHOT_API_KEY",
		Enabled:        true,
		Models: []Model{
			{ID: "moonshot/kimi-k2.5", Name: "Kimi K2.5"},
			{ID: "moonshot/kimi-k2-0905-preview", Name: "Kimi K2 0905 Preview"},
			{ID: "moonshot/kimi-k2-thinking", Name: "Kimi K2 Thinking"},
		},
	},
	"kimi": {
		Key:            "kimi",
		DisplayName:    "Kimi (CN)",
		Sort:           8,
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
		Sort:           9,
		DefaultBaseURL: "https://api.moonshot.cn/anthropic/v1",
		EnvKey:         "KIMI_API_KEY",
		Enabled:        true,
		Models: []Model{
			{ID: "kimi-coding/k2p5", Name: "Kimi K2.5"},
		},
	},
	"qwen": {
		Key:            "qwen",
		DisplayName:    "Qwen",
		Sort:           10,
		DefaultBaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
		EnvKey:         "QWEN_API_KEY",
		Enabled:        false,
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
		copy(clone.Models, meta.Models)
	}
	return clone
}
