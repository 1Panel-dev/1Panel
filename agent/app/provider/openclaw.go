package provider

import (
	"fmt"
	"strings"
)

type OpenClawPatch struct {
	PrimaryModel string
	Models       map[string]interface{}
}

type OpenClawProviderPatch struct {
	PrimaryModel string
	ProviderKey  string
	ModelID      string
	APIKey       string
	BaseURL      string
	APIType      string
	AuthHeader   bool
}

func BuildOpenClawPatch(provider, modelName, apiType string, reasoning bool, maxTokens, contextWindow int, baseURL, apiKey string) (*OpenClawPatch, error) {
	providerPatch, err := BuildOpenClawProviderPatch(provider, modelName, apiType, baseURL, apiKey)
	if err != nil {
		return nil, err
	}
	_, maxTokens, contextWindow = ResolveRuntimeParams(provider, apiType, maxTokens, contextWindow)
	return newOpenClawPatch(
		providerPatch,
		resolveOpenClawModelName(provider, providerPatch.ModelID),
		resolveOpenClawModelInput(provider, providerPatch.ModelID),
		reasoning,
		contextWindow,
		maxTokens,
	), nil
}

func BuildOpenClawProviderPatch(provider, modelName, apiType, baseURL, apiKey string) (*OpenClawProviderPatch, error) {
	modelName = strings.TrimSpace(modelName)
	if modelName == "" {
		return nil, fmt.Errorf("model is required")
	}
	resolvedAPIType, _, _ := ResolveRuntimeParams(provider, apiType, 0, 0)
	modelID := resolveOpenClawModelID(provider, modelName)
	switch provider {
	case "deepseek":
		return newOpenClawProviderPatch(modelName, "deepseek", modelID, strings.TrimSpace(apiKey), baseURL, "openai-completions", false), nil
	case "gemini":
		return newOpenClawProviderPatch("google/"+modelID, "google", modelID, strings.TrimSpace(apiKey), baseURL, resolvedAPIType, false), nil
	case "moonshot", "kimi":
		return buildMoonshotProviderPatch(provider, modelName, modelID, baseURL, apiKey), nil
	case "bailian-coding-plan":
		return newOpenClawProviderPatch("bailian-coding-plan/"+modelID, "bailian-coding-plan", modelID, strings.TrimSpace(apiKey), baseURL, "openai-completions", false), nil
	case "ark-coding-plan":
		return newOpenClawProviderPatch("ark-coding-plan/"+modelID, "ark-coding-plan", modelID, strings.TrimSpace(apiKey), baseURL, "openai-completions", false), nil
	case "minimax":
		return newOpenClawProviderPatch("minimax/"+modelID, "minimax", modelID, strings.TrimSpace(apiKey), baseURL, "anthropic-messages", true), nil
	case "xiaomi":
		return newOpenClawProviderPatch("xiaomi/"+modelID, "xiaomi", modelID, strings.TrimSpace(apiKey), baseURL, "openai-completions", false), nil
	case "custom", "vllm":
		return newOpenClawProviderPatch(provider+"/"+modelID, provider, modelID, strings.TrimSpace(apiKey), strings.TrimSpace(baseURL), resolvedAPIType, false), nil
	case "ollama":
		return newOpenClawProviderPatch(modelName, "ollama", modelID, "ollama", strings.TrimSpace(baseURL), resolvedAPIType, false), nil
	case "kimi-coding":
		return newOpenClawProviderPatch(modelName, "kimi-coding", modelID, strings.TrimSpace(apiKey), baseURL, "anthropic-messages", false), nil
	case "zai":
		return newOpenClawProviderPatch("zai/"+modelID, "zai", modelID, strings.TrimSpace(apiKey), baseURL, "openai-completions", false), nil
	default:
		return newOpenClawProviderPatch(modelName, provider, modelID, strings.TrimSpace(apiKey), baseURL, resolvedAPIType, false), nil
	}
}

func buildMoonshotProviderPatch(provider, modelName, modelID, baseURL, apiKey string) *OpenClawProviderPatch {
	providerKey := provider
	primaryModel := modelName
	if provider == "kimi" {
		providerKey = "moonshot"
		primaryModel = "moonshot/" + modelID
	}
	return newOpenClawProviderPatch(primaryModel, providerKey, modelID, strings.TrimSpace(apiKey), baseURL, "openai-completions", false)
}

func newOpenClawProviderPatch(primaryModel, providerKey, modelID, apiKey, baseURL, apiType string, authHeader bool) *OpenClawProviderPatch {
	return &OpenClawProviderPatch{
		PrimaryModel: primaryModel,
		ProviderKey:  providerKey,
		ModelID:      modelID,
		APIKey:       apiKey,
		BaseURL:      baseURL,
		APIType:      apiType,
		AuthHeader:   authHeader,
	}
}

func newOpenClawPatch(providerPatch *OpenClawProviderPatch, modelName string, input []string, reasoning bool, contextWindow, maxTokens int) *OpenClawPatch {
	model := map[string]interface{}{
		"id":            providerPatch.ModelID,
		"name":          modelName,
		"reasoning":     reasoning,
		"input":         input,
		"contextWindow": contextWindow,
		"maxTokens":     maxTokens,
		"cost":          map[string]interface{}{},
	}
	providerConfig := map[string]interface{}{
		"apiKey":  providerPatch.APIKey,
		"baseUrl": providerPatch.BaseURL,
		"api":     providerPatch.APIType,
		"models":  []map[string]interface{}{model},
	}
	if providerPatch.AuthHeader {
		providerConfig["authHeader"] = true
	}
	return &OpenClawPatch{
		PrimaryModel: providerPatch.PrimaryModel,
		Models: map[string]interface{}{
			"mode": "merge",
			"providers": map[string]interface{}{
				providerPatch.ProviderKey: providerConfig,
			},
		},
	}
}

func resolveOpenClawModelID(provider, modelName string) string {
	if provider == "custom" || provider == "vllm" {
		return normalizeCustomModel(modelName)
	}
	if parts := strings.SplitN(modelName, "/", 2); len(parts) == 2 {
		return parts[1]
	}
	return modelName
}

func resolveOpenClawModelName(provider, modelID string) string {
	if item, ok := resolveOpenClawCatalogModel(provider, modelID); ok {
		return item.Name
	}
	if provider == "kimi-coding" {
		return "Kimi for Coding"
	}
	return modelID
}

func resolveOpenClawModelInput(provider, modelID string) []string {
	if item, ok := resolveOpenClawCatalogModel(provider, modelID); ok && len(item.Input) > 0 {
		return item.Input
	}
	if provider == "kimi-coding" {
		return []string{"text", "image"}
	}
	return []string{"text"}
}

func resolveOpenClawCatalogModel(provider, modelID string) (Model, bool) {
	candidates := []string{provider + "/" + modelID}
	if provider == "gemini" {
		candidates = append(candidates, "google/"+modelID)
	}
	for _, candidate := range candidates {
		if item, ok := FindModel(provider, candidate); ok {
			return item, true
		}
	}
	return Model{}, false
}

func normalizeCustomModel(modelName string) string {
	trim := strings.TrimSpace(modelName)
	trim = strings.TrimLeft(trim, "/")
	if parts := strings.SplitN(trim, "/", 2); len(parts) == 2 && strings.EqualFold(parts[0], "custom") {
		return strings.TrimLeft(strings.TrimSpace(parts[1]), "/")
	}
	return trim
}
