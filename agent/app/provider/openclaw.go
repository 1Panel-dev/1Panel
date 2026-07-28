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

func BuildOpenClawProviderPatch(provider, modelName, apiType, authMode, baseURL, apiKey string) (*OpenClawProviderPatch, error) {
	if modelName == "" {
		return nil, fmt.Errorf("model is required")
	}
	resolvedAPIType := apiType
	if _, ok := FindAPIConfig(provider, resolvedAPIType); !ok {
		resolvedAPIType = DefaultAPIType(provider)
	}
	if IsImageAPIType(resolvedAPIType) {
		return nil, fmt.Errorf("api type %s does not support text generation", resolvedAPIType)
	}
	resolvedAuthMode, err := ResolveAuthMode(provider, resolvedAPIType, authMode)
	if err != nil {
		return nil, err
	}
	usesBearer := resolvedAuthMode == AuthModeBearer
	modelID := NormalizeModelID(provider, modelName)
	providerKey := provider
	preserveQualifiedModel := false
	switch provider {
	case "gemini":
		providerKey = "google"
		resolvedAPIType = "google-generative-ai"
		usesBearer = false
	case "moonshot", "kimi":
		providerKey = "moonshot"
		resolvedAPIType = "openai-completions"
		usesBearer = false
	case "ollama":
		apiKey = "ollama"
		usesBearer = false
	case "openai", "openrouter", "anthropic":
		preserveQualifiedModel = strings.Contains(modelName, "/")
	}

	primaryModel := providerKey + "/" + modelID
	if preserveQualifiedModel {
		primaryModel = modelName
	}
	return &OpenClawProviderPatch{
		PrimaryModel: primaryModel,
		ProviderKey:  providerKey,
		ModelID:      modelID,
		APIKey:       apiKey,
		BaseURL:      baseURL,
		APIType:      resolvedAPIType,
		AuthHeader:   usesBearer,
	}, nil
}
