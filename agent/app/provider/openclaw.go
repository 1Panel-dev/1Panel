package provider

import (
	"fmt"
	"strings"
)

type OpenClawPatch struct {
	PrimaryModel string
	Models       map[string]interface{}
}

func BuildOpenClawPatch(provider, modelName, apiType string, maxTokens, contextWindow int, baseURL, apiKey string) (*OpenClawPatch, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	modelName = strings.TrimSpace(modelName)
	if modelName == "" {
		return nil, fmt.Errorf("model is required")
	}
	modelID := modelName
	if parts := strings.SplitN(modelName, "/", 2); len(parts) == 2 {
		modelID = parts[1]
	}

	switch provider {
	case "deepseek":
		return buildDeepseekPatch(modelName, baseURL, apiKey), nil
	case "gemini":
		return buildGenericPatch(provider, modelName, modelID, apiType, maxTokens, contextWindow, baseURL, apiKey), nil
	case "moonshot", "kimi":
		return buildMoonshotPatch(provider, modelName, modelID, baseURL, apiKey), nil
	case "bailian-coding-plan":
		return buildBailianPatch(modelID, maxTokens, contextWindow, baseURL, apiKey), nil
	case "ark-coding-plan":
		return buildArkPatch(modelID, maxTokens, contextWindow, baseURL, apiKey), nil
	case "minimax":
		return buildMiniMaxPatch(modelID, baseURL, apiKey), nil
	case "custom", "vllm":
		return buildCustomPatch(provider, modelName, apiType, maxTokens, contextWindow, baseURL, apiKey), nil
	case "ollama":
		return buildOllamaPatch(modelName, modelID, apiType, baseURL), nil
	case "kimi-coding":
		return buildKimiCodingPatch(modelName, modelID, baseURL, apiKey), nil
	case "zai":
		return buildZaiPatch(modelID, maxTokens, contextWindow, baseURL, apiKey), nil
	default:
		return buildGenericPatch(provider, modelName, modelID, apiType, maxTokens, contextWindow, baseURL, apiKey), nil
	}
}

func buildDeepseekPatch(modelName, baseURL, apiKey string) *OpenClawPatch {
	return &OpenClawPatch{
		PrimaryModel: modelName,
		Models: providerModels("deepseek", strings.TrimSpace(apiKey), firstNonEmpty(strings.TrimSpace(baseURL), "https://api.deepseek.com/v1"), "openai-completions", map[string]interface{}{
			"id":            "deepseek-chat",
			"name":          "DeepSeek Chat",
			"reasoning":     false,
			"input":         []string{"text"},
			"contextWindow": 128000,
			"maxTokens":     8192,
			"cost":          map[string]interface{}{},
		}),
	}
}

func buildMoonshotPatch(provider, modelName, modelID, baseURL, apiKey string) *OpenClawPatch {
	configProvider := provider
	primaryModel := modelName
	if provider == "kimi" {
		configProvider = "moonshot"
		primaryModel = "moonshot/" + modelID
	}
	return &OpenClawPatch{
		PrimaryModel: primaryModel,
		Models: providerModels(configProvider, strings.TrimSpace(apiKey), withCatalogDefault(provider, baseURL), "openai-completions", map[string]interface{}{
			"id":            modelID,
			"name":          modelID,
			"reasoning":     strings.Contains(strings.ToLower(modelID), "thinking"),
			"input":         []string{"text"},
			"contextWindow": 256000,
			"maxTokens":     8192,
			"cost":          map[string]interface{}{},
		}),
	}
}

func buildBailianPatch(modelID string, maxTokens, contextWindow int, baseURL, apiKey string) *OpenClawPatch {
	normalizedID := normalizeBailianCodingPlanModelID(modelID)
	return &OpenClawPatch{
		PrimaryModel: "bailian-coding-plan/" + bailianPrimaryModelID(normalizedID),
		Models: providerModels("bailian-coding-plan", strings.TrimSpace(apiKey), withCatalogDefault("bailian-coding-plan", baseURL), "openai-completions", map[string]interface{}{
			"id":            normalizedID,
			"name":          normalizedID,
			"reasoning":     isReasoningModel(normalizedID),
			"input":         []string{"text"},
			"contextWindow": fallbackInt(contextWindow, 256000),
			"maxTokens":     fallbackInt(maxTokens, 8192),
			"cost":          map[string]interface{}{},
		}),
	}
}

func buildArkPatch(modelID string, maxTokens, contextWindow int, baseURL, apiKey string) *OpenClawPatch {
	normalizedID := normalizeArkCodingPlanModelID(modelID)
	return &OpenClawPatch{
		PrimaryModel: "ark-coding-plan/" + normalizedID,
		Models: providerModels("ark-coding-plan", strings.TrimSpace(apiKey), withCatalogDefault("ark-coding-plan", baseURL), "openai-completions", map[string]interface{}{
			"id":            normalizedID,
			"name":          normalizedID,
			"reasoning":     isReasoningModel(normalizedID),
			"input":         []string{"text"},
			"contextWindow": fallbackInt(contextWindow, 256000),
			"maxTokens":     fallbackInt(maxTokens, 8192),
			"cost":          map[string]interface{}{},
		}),
	}
}

func buildMiniMaxPatch(modelID, baseURL, apiKey string) *OpenClawPatch {
	normalizedID := normalizeMiniMaxModelID(modelID)
	return &OpenClawPatch{
		PrimaryModel: "minimax/" + normalizedID,
		Models: map[string]interface{}{
			"mode": "merge",
			"providers": map[string]interface{}{
				"minimax": map[string]interface{}{
					"apiKey":     strings.TrimSpace(apiKey),
					"baseUrl":    firstNonEmpty(strings.TrimSpace(baseURL), "https://api.minimaxi.com/anthropic"),
					"api":        "anthropic-messages",
					"authHeader": true,
					"models": []map[string]interface{}{{
						"id":            normalizedID,
						"name":          strings.ReplaceAll(normalizedID, "-", " "),
						"reasoning":     false,
						"input":         []string{"text"},
						"contextWindow": 200000,
						"maxTokens":     8192,
						"cost":          map[string]interface{}{},
					}},
				},
			},
		},
	}
}

func buildCustomPatch(provider, modelName, apiType string, maxTokens, contextWindow int, baseURL, apiKey string) *OpenClawPatch {
	customModelID := normalizeCustomModel(modelName)
	return &OpenClawPatch{
		PrimaryModel: provider + "/" + customModelID,
		Models: providerModels(provider, strings.TrimSpace(apiKey), strings.TrimSpace(baseURL), apiType, map[string]interface{}{
			"id":            customModelID,
			"name":          customModelID,
			"reasoning":     isReasoningModel(customModelID),
			"input":         []string{"text"},
			"contextWindow": fallbackInt(contextWindow, 128000),
			"maxTokens":     fallbackInt(maxTokens, 8192),
			"cost":          map[string]interface{}{},
		}),
	}
}

func buildOllamaPatch(modelName, modelID, apiType, baseURL string) *OpenClawPatch {
	api := normalizeAPIType(apiType)
	if api != "openai-completions" && api != "openai-responses" {
		api = "openai-responses"
	}
	return &OpenClawPatch{
		PrimaryModel: modelName,
		Models: providerModels("ollama", "ollama", strings.TrimSpace(baseURL), api, map[string]interface{}{
			"id":            modelID,
			"name":          modelID,
			"reasoning":     api != "openai-completions",
			"input":         []string{"text"},
			"contextWindow": 160000,
			"maxTokens":     8192,
			"cost":          map[string]interface{}{},
		}),
	}
}

func buildKimiCodingPatch(modelName, modelID, baseURL, apiKey string) *OpenClawPatch {
	return &OpenClawPatch{
		PrimaryModel: modelName,
		Models: providerModels("kimi-coding", strings.TrimSpace(apiKey), withCatalogDefault("kimi-coding", baseURL), "anthropic-messages", map[string]interface{}{
			"id":            modelID,
			"name":          "Kimi for Coding",
			"reasoning":     true,
			"input":         []string{"text", "image"},
			"contextWindow": 262144,
			"maxTokens":     32768,
			"cost":          map[string]interface{}{},
		}),
	}
}

func buildZaiPatch(modelID string, maxTokens, contextWindow int, baseURL, apiKey string) *OpenClawPatch {
	return &OpenClawPatch{
		PrimaryModel: "zai/" + modelID,
		Models: providerModels("zai", strings.TrimSpace(apiKey), withCatalogDefault("zai", baseURL), "openai-completions", map[string]interface{}{
			"id":            modelID,
			"name":          zaiModelDisplayName(modelID),
			"reasoning":     modelID == "glm-5",
			"input":         []string{"text"},
			"contextWindow": fallbackInt(contextWindow, 204800),
			"maxTokens":     fallbackInt(maxTokens, 131072),
			"cost":          map[string]interface{}{},
		}),
	}
}

func buildGenericPatch(provider, modelName, modelID, apiType string, maxTokens, contextWindow int, baseURL, apiKey string) *OpenClawPatch {
	providerName := provider
	primaryModel := modelName
	if provider == "gemini" {
		providerName = "google"
		primaryModel = "google/" + modelID
	}
	return &OpenClawPatch{
		PrimaryModel: primaryModel,
		Models: providerModels(providerName, strings.TrimSpace(apiKey), withCatalogDefault(provider, baseURL), normalizeAPIType(apiType), map[string]interface{}{
			"id":            modelID,
			"name":          modelID,
			"reasoning":     isReasoningModel(modelID),
			"input":         []string{"text"},
			"contextWindow": fallbackInt(contextWindow, 256000),
			"maxTokens":     fallbackInt(maxTokens, 8192),
			"cost":          map[string]interface{}{},
		}),
	}
}

func providerModels(provider, apiKey, baseURL, api string, model map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"mode": "merge",
		"providers": map[string]interface{}{
			provider: map[string]interface{}{
				"apiKey":  apiKey,
				"baseUrl": baseURL,
				"api":     api,
				"models":  []map[string]interface{}{model},
			},
		},
	}
}

func withCatalogDefault(provider, baseURL string) string {
	if strings.TrimSpace(baseURL) != "" {
		return strings.TrimSpace(baseURL)
	}
	if defaultURL, ok := DefaultBaseURL(provider); ok {
		return defaultURL
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func fallbackInt(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func normalizeAPIType(apiType string) string {
	trim := strings.ToLower(strings.TrimSpace(apiType))
	if trim == "" {
		return "openai-completions"
	}
	return trim
}

func normalizeCustomModel(modelName string) string {
	trim := strings.TrimSpace(modelName)
	trim = strings.TrimLeft(trim, "/")
	if parts := strings.SplitN(trim, "/", 2); len(parts) == 2 && strings.EqualFold(parts[0], "custom") {
		return strings.TrimLeft(strings.TrimSpace(parts[1]), "/")
	}
	return trim
}

func normalizeBailianCodingPlanModelID(modelID string) string {
	trim := strings.TrimSpace(modelID)
	switch strings.ToLower(trim) {
	case "minimax-m2.5", "minimax m2.5", "minimax/minimax-m2.5", "minimax/minimax m2.5":
		return "MiniMax/MiniMax-M2.5"
	default:
		return trim
	}
}

func normalizeArkCodingPlanModelID(modelID string) string {
	return strings.ToLower(strings.TrimSpace(modelID))
}

func normalizeMiniMaxModelID(modelID string) string {
	switch strings.ToLower(strings.TrimSpace(modelID)) {
	case "minimax-m2.1", "minimax m2.1", "minimax-m2.1-preview", "minimax-m2.1-latest":
		return "MiniMax-M2.1"
	case "minimax-m2.1-lightning", "minimax m2.1 lightning":
		return "MiniMax-M2.1-lightning"
	default:
		return modelID
	}
}

func zaiModelDisplayName(modelID string) string {
	switch strings.ToLower(strings.TrimSpace(modelID)) {
	case "glm-5":
		return "GLM-5"
	case "glm-4.7":
		return "GLM-4.7"
	case "glm-4.7-flash":
		return "GLM-4.7-Flash"
	case "glm-4.7-flashx":
		return "GLM-4.7-FlashX"
	default:
		return strings.TrimSpace(modelID)
	}
}

func bailianPrimaryModelID(modelID string) string {
	trim := strings.TrimSpace(modelID)
	if trim == "" {
		return ""
	}
	parts := strings.Split(trim, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		part := strings.TrimSpace(parts[i])
		if part != "" {
			return part
		}
	}
	return trim
}

func isReasoningModel(modelID string) bool {
	trim := strings.ToLower(strings.TrimSpace(modelID))
	return strings.Contains(trim, "reason") || strings.Contains(trim, "thinking")
}
