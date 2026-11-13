package provider

import (
	"strings"
)

type AccountTypePreset struct {
	Key        string      `json:"key"`
	APIConfigs []APIConfig `json:"apiTypes"`
}

func FindAccountType(provider, accountType string) (AccountTypePreset, bool) {
	meta, ok := catalog[provider]
	if !ok || len(meta.AccountTypes) == 0 {
		return AccountTypePreset{}, false
	}
	if accountType == "" {
		switch provider {
		case "bailian", "ark", "xiaomi":
			accountType = "payAsYouGo"
		case "kimi", "zai":
			accountType = "chinaApi"
		default:
			return meta.AccountTypes[0], true
		}
	}
	for _, item := range meta.AccountTypes {
		if item.Key == accountType {
			return item, true
		}
	}
	return AccountTypePreset{}, false
}

func Purpose(apiType string) string {
	switch apiType {
	case "openai-completions", "openai-responses", "anthropic-messages", "gemini-generate-content":
		return "text"
	case "openai-embeddings":
		return "embedding"
	case "rerank":
		return "rerank"
	}
	if strings.HasSuffix(apiType, "-images") {
		return "image"
	}
	if strings.HasSuffix(apiType, "-transcriptions") {
		return "transcription"
	}
	if strings.HasSuffix(apiType, "-speech") {
		return "speech"
	}
	return ""
}
