package provider

import (
	"fmt"
	"strings"
)

type OpenClawProviderPatch struct {
	PrimaryModel string
	ProviderKey  string
	ModelID      string
	APIKey       string
	BaseURL      string
	APIType      string
	AuthHeader   bool
}

func BuildOpenClawProviderPatch(provider, modelName, apiType, baseURL, apiKey string) (*OpenClawProviderPatch, error) {
	if modelName == "" {
		return nil, fmt.Errorf("model is required")
	}
	resolvedAPIType, _, _ := ResolveRuntimeParams(provider, apiType, 0, 0)
	modelID := resolveOpenClawModelID(provider, modelName)
	switch provider {
	case "deepseek":
		return newOpenClawProviderPatch(modelName, "deepseek", modelID, apiKey, baseURL, "openai-completions", false), nil
	case "gemini":
		return newOpenClawProviderPatch("google/"+modelID, "google", modelID, apiKey, baseURL, resolvedAPIType, false), nil
	case "moonshot", "kimi":
		return buildMoonshotProviderPatch(provider, modelName, modelID, baseURL, apiKey), nil
	case "bailian-coding-plan":
		return newOpenClawProviderPatch("bailian-coding-plan/"+modelID, "bailian-coding-plan", modelID, apiKey, baseURL, "openai-completions", false), nil
	case "ark-coding-plan":
		return newOpenClawProviderPatch("ark-coding-plan/"+modelID, "ark-coding-plan", modelID, apiKey, baseURL, "openai-completions", false), nil
	case "minimax":
		return newOpenClawProviderPatch("minimax/"+modelID, "minimax", modelID, apiKey, baseURL, "anthropic-messages", true), nil
	case "xiaomi":
		return newOpenClawProviderPatch("xiaomi/"+modelID, "xiaomi", modelID, apiKey, baseURL, "openai-completions", false), nil
	case "custom", "vllm":
		return newOpenClawProviderPatch(provider+"/"+modelID, provider, modelID, apiKey, baseURL, resolvedAPIType, false), nil
	case "ollama":
		return newOpenClawProviderPatch(modelName, "ollama", modelID, "ollama", baseURL, resolvedAPIType, false), nil
	case "kimi-coding":
		return newOpenClawProviderPatch(modelName, "kimi-coding", modelID, apiKey, baseURL, "anthropic-messages", false), nil
	case "zai":
		return newOpenClawProviderPatch("zai/"+modelID, "zai", modelID, apiKey, baseURL, "openai-completions", false), nil
	default:
		return newOpenClawProviderPatch(modelName, provider, modelID, apiKey, baseURL, resolvedAPIType, false), nil
	}
}

func buildMoonshotProviderPatch(provider, modelName, modelID, baseURL, apiKey string) *OpenClawProviderPatch {
	providerKey := provider
	primaryModel := modelName
	if provider == "kimi" {
		providerKey = "moonshot"
		primaryModel = "moonshot/" + modelID
	}
	return newOpenClawProviderPatch(primaryModel, providerKey, modelID, apiKey, baseURL, "openai-completions", false)
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

func resolveOpenClawModelID(provider, modelName string) string {
	if provider == "custom" || provider == "vllm" {
		target := strings.TrimLeft(modelName, "/")
		if parts := strings.SplitN(target, "/", 2); len(parts) == 2 && parts[0] == "custom" {
			return strings.TrimLeft(parts[1], "/")
		}
		return target
	}
	if parts := strings.SplitN(modelName, "/", 2); len(parts) == 2 {
		return parts[1]
	}
	return modelName
}
