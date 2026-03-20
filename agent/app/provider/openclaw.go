package provider

import (
	"fmt"
	"strings"
)

type OpenClawPatch struct {
	PrimaryModel string
	Models       map[string]interface{}
}

type openClawModelSpec struct {
	ID            string
	Name          string
	Reasoning     bool
	Input         []string
	ContextWindow int
	MaxTokens     int
}

type openClawPatchSpec struct {
	PrimaryModel string
	Provider     string
	APIKey       string
	BaseURL      string
	APIType      string
	AuthHeader   bool
	Model        openClawModelSpec
}

func BuildOpenClawPatch(provider, modelName, apiType string, reasoning bool, maxTokens, contextWindow int, baseURL, apiKey string) (*OpenClawPatch, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	modelName = strings.TrimSpace(modelName)
	if modelName == "" {
		return nil, fmt.Errorf("model is required")
	}
	apiType, maxTokens, contextWindow = ResolveRuntimeParams(provider, apiType, maxTokens, contextWindow)
	modelID := modelName
	if parts := strings.SplitN(modelName, "/", 2); len(parts) == 2 {
		modelID = parts[1]
	}

	var spec openClawPatchSpec
	switch provider {
	case "deepseek":
		spec = buildDeepseekPatchSpec(modelName, reasoning, maxTokens, contextWindow, baseURL, apiKey)
	case "gemini":
		spec = buildGenericPatchSpec(provider, modelName, modelID, apiType, reasoning, maxTokens, contextWindow, baseURL, apiKey)
	case "moonshot", "kimi":
		spec = buildMoonshotPatchSpec(provider, modelName, modelID, reasoning, maxTokens, contextWindow, baseURL, apiKey)
	case "bailian-coding-plan":
		spec = buildBailianPatchSpec(modelID, reasoning, maxTokens, contextWindow, baseURL, apiKey)
	case "ark-coding-plan":
		spec = buildArkPatchSpec(modelID, reasoning, maxTokens, contextWindow, baseURL, apiKey)
	case "minimax":
		spec = buildMiniMaxPatchSpec(modelID, reasoning, maxTokens, contextWindow, baseURL, apiKey)
	case "xiaomi":
		spec = buildXiaomiPatchSpec(modelID, reasoning, maxTokens, contextWindow, baseURL, apiKey)
	case "custom", "vllm":
		spec = buildCustomPatchSpec(provider, modelName, apiType, reasoning, maxTokens, contextWindow, baseURL, apiKey)
	case "ollama":
		spec = buildOllamaPatchSpec(modelName, modelID, apiType, reasoning, maxTokens, contextWindow, baseURL)
	case "kimi-coding":
		spec = buildKimiCodingPatchSpec(modelName, modelID, reasoning, maxTokens, contextWindow, baseURL, apiKey)
	case "zai":
		spec = buildZaiPatchSpec(modelID, reasoning, maxTokens, contextWindow, baseURL, apiKey)
	default:
		spec = buildGenericPatchSpec(provider, modelName, modelID, apiType, reasoning, maxTokens, contextWindow, baseURL, apiKey)
	}
	return buildOpenClawPatch(spec), nil
}

func buildOpenClawPatch(spec openClawPatchSpec) *OpenClawPatch {
	return &OpenClawPatch{
		PrimaryModel: spec.PrimaryModel,
		Models: providerModels(
			spec.Provider,
			spec.APIKey,
			spec.BaseURL,
			spec.APIType,
			spec.AuthHeader,
			buildOpenClawModel(spec.Model),
		),
	}
}

func buildOpenClawModel(spec openClawModelSpec) map[string]interface{} {
	return map[string]interface{}{
		"id":            spec.ID,
		"name":          spec.Name,
		"reasoning":     spec.Reasoning,
		"input":         spec.Input,
		"contextWindow": spec.ContextWindow,
		"maxTokens":     spec.MaxTokens,
		"cost":          map[string]interface{}{},
	}
}

func buildDeepseekPatchSpec(modelName string, reasoning bool, maxTokens, contextWindow int, baseURL, apiKey string) openClawPatchSpec {
	return openClawPatchSpec{
		PrimaryModel: modelName,
		Provider:     "deepseek",
		APIKey:       strings.TrimSpace(apiKey),
		BaseURL:      baseURL,
		APIType:      "openai-completions",
		Model: openClawModelSpec{
			ID:            "deepseek-chat",
			Name:          "DeepSeek Chat",
			Reasoning:     reasoning,
			Input:         []string{"text"},
			ContextWindow: contextWindow,
			MaxTokens:     maxTokens,
		},
	}
}

func buildMoonshotPatchSpec(provider, modelName, modelID string, reasoning bool, maxTokens, contextWindow int, baseURL, apiKey string) openClawPatchSpec {
	configProvider := provider
	primaryModel := modelName
	if provider == "kimi" {
		configProvider = "moonshot"
		primaryModel = "moonshot/" + modelID
	}
	return openClawPatchSpec{
		PrimaryModel: primaryModel,
		Provider:     configProvider,
		APIKey:       strings.TrimSpace(apiKey),
		BaseURL:      baseURL,
		APIType:      "openai-completions",
		Model: openClawModelSpec{
			ID:            modelID,
			Name:          modelID,
			Reasoning:     reasoning,
			Input:         []string{"text"},
			ContextWindow: contextWindow,
			MaxTokens:     maxTokens,
		},
	}
}

func buildBailianPatchSpec(modelID string, reasoning bool, maxTokens, contextWindow int, baseURL, apiKey string) openClawPatchSpec {
	return openClawPatchSpec{
		PrimaryModel: "bailian-coding-plan/" + modelID,
		Provider:     "bailian-coding-plan",
		APIKey:       strings.TrimSpace(apiKey),
		BaseURL:      baseURL,
		APIType:      "openai-completions",
		Model: openClawModelSpec{
			ID:            modelID,
			Name:          modelID,
			Reasoning:     reasoning,
			Input:         []string{"text"},
			ContextWindow: contextWindow,
			MaxTokens:     maxTokens,
		},
	}
}

func buildArkPatchSpec(modelID string, reasoning bool, maxTokens, contextWindow int, baseURL, apiKey string) openClawPatchSpec {
	return openClawPatchSpec{
		PrimaryModel: "ark-coding-plan/" + modelID,
		Provider:     "ark-coding-plan",
		APIKey:       strings.TrimSpace(apiKey),
		BaseURL:      baseURL,
		APIType:      "openai-completions",
		Model: openClawModelSpec{
			ID:            modelID,
			Name:          modelID,
			Reasoning:     reasoning,
			Input:         []string{"text"},
			ContextWindow: contextWindow,
			MaxTokens:     maxTokens,
		},
	}
}

func buildMiniMaxPatchSpec(modelID string, reasoning bool, maxTokens, contextWindow int, baseURL, apiKey string) openClawPatchSpec {
	return openClawPatchSpec{
		PrimaryModel: "minimax/" + modelID,
		Provider:     "minimax",
		APIKey:       strings.TrimSpace(apiKey),
		BaseURL:      baseURL,
		APIType:      "anthropic-messages",
		AuthHeader:   true,
		Model: openClawModelSpec{
			ID:            modelID,
			Name:          modelID,
			Reasoning:     reasoning,
			Input:         []string{"text"},
			ContextWindow: contextWindow,
			MaxTokens:     maxTokens,
		},
	}
}

func buildXiaomiPatchSpec(modelID string, reasoning bool, maxTokens, contextWindow int, baseURL, apiKey string) openClawPatchSpec {
	normalizedID := strings.TrimSpace(modelID)
	return openClawPatchSpec{
		PrimaryModel: "xiaomi/" + normalizedID,
		Provider:     "xiaomi",
		APIKey:       strings.TrimSpace(apiKey),
		BaseURL:      baseURL,
		APIType:      "anthropic-messages",
		Model: openClawModelSpec{
			ID:            normalizedID,
			Name:          normalizedID,
			Reasoning:     reasoning,
			Input:         []string{"text"},
			ContextWindow: contextWindow,
			MaxTokens:     maxTokens,
		},
	}
}

func buildCustomPatchSpec(provider, modelName, apiType string, reasoning bool, maxTokens, contextWindow int, baseURL, apiKey string) openClawPatchSpec {
	customModelID := normalizeCustomModel(modelName)
	return openClawPatchSpec{
		PrimaryModel: provider + "/" + customModelID,
		Provider:     provider,
		APIKey:       strings.TrimSpace(apiKey),
		BaseURL:      strings.TrimSpace(baseURL),
		APIType:      apiType,
		Model: openClawModelSpec{
			ID:            customModelID,
			Name:          customModelID,
			Reasoning:     reasoning,
			Input:         []string{"text"},
			ContextWindow: contextWindow,
			MaxTokens:     maxTokens,
		},
	}
}

func buildOllamaPatchSpec(modelName, modelID, apiType string, reasoning bool, maxTokens, contextWindow int, baseURL string) openClawPatchSpec {
	return openClawPatchSpec{
		PrimaryModel: modelName,
		Provider:     "ollama",
		APIKey:       "ollama",
		BaseURL:      strings.TrimSpace(baseURL),
		APIType:      apiType,
		Model: openClawModelSpec{
			ID:            modelID,
			Name:          modelID,
			Reasoning:     reasoning,
			Input:         []string{"text"},
			ContextWindow: contextWindow,
			MaxTokens:     maxTokens,
		},
	}
}

func buildKimiCodingPatchSpec(modelName, modelID string, reasoning bool, maxTokens, contextWindow int, baseURL, apiKey string) openClawPatchSpec {
	return openClawPatchSpec{
		PrimaryModel: modelName,
		Provider:     "kimi-coding",
		APIKey:       strings.TrimSpace(apiKey),
		BaseURL:      baseURL,
		APIType:      "anthropic-messages",
		Model: openClawModelSpec{
			ID:            modelID,
			Name:          "Kimi for Coding",
			Reasoning:     reasoning,
			Input:         []string{"text", "image"},
			ContextWindow: contextWindow,
			MaxTokens:     maxTokens,
		},
	}
}

func buildZaiPatchSpec(modelID string, reasoning bool, maxTokens, contextWindow int, baseURL, apiKey string) openClawPatchSpec {
	return openClawPatchSpec{
		PrimaryModel: "zai/" + modelID,
		Provider:     "zai",
		APIKey:       strings.TrimSpace(apiKey),
		BaseURL:      baseURL,
		APIType:      "openai-completions",
		Model: openClawModelSpec{
			ID:            modelID,
			Name:          modelID,
			Reasoning:     reasoning,
			Input:         []string{"text"},
			ContextWindow: contextWindow,
			MaxTokens:     maxTokens,
		},
	}
}

func buildGenericPatchSpec(provider, modelName, modelID, apiType string, reasoning bool, maxTokens, contextWindow int, baseURL, apiKey string) openClawPatchSpec {
	providerName := provider
	primaryModel := modelName
	if provider == "gemini" {
		providerName = "google"
		primaryModel = "google/" + modelID
	}
	return openClawPatchSpec{
		PrimaryModel: primaryModel,
		Provider:     providerName,
		APIKey:       strings.TrimSpace(apiKey),
		BaseURL:      baseURL,
		APIType:      apiType,
		Model: openClawModelSpec{
			ID:            modelID,
			Name:          modelID,
			Reasoning:     reasoning,
			Input:         []string{"text"},
			ContextWindow: contextWindow,
			MaxTokens:     maxTokens,
		},
	}
}

func providerModels(provider, apiKey, baseURL, api string, authHeader bool, model map[string]interface{}) map[string]interface{} {
	providerConfig := map[string]interface{}{
		"apiKey":  apiKey,
		"baseUrl": baseURL,
		"api":     api,
		"models":  []map[string]interface{}{model},
	}
	if authHeader {
		providerConfig["authHeader"] = true
	}
	return map[string]interface{}{
		"mode": "merge",
		"providers": map[string]interface{}{
			provider: providerConfig,
		},
	}
}

func normalizeCustomModel(modelName string) string {
	trim := strings.TrimSpace(modelName)
	trim = strings.TrimLeft(trim, "/")
	if parts := strings.SplitN(trim, "/", 2); len(parts) == 2 && strings.EqualFold(parts[0], "custom") {
		return strings.TrimLeft(strings.TrimSpace(parts[1]), "/")
	}
	return trim
}
