package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/aihttp"
)

func DiscoverModels(baseURL, apiKey string) ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, buildModelDiscoveryURL(baseURL), nil)
	if err != nil {
		return nil, aihttp.ErrTarget
	}
	if strings.TrimSpace(apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := aihttp.NewClient(defaultVerifyTimeout).Do(req)
	if err != nil {
		return nil, aihttp.SafeError(err, req.URL.String(), apiKey)
	}
	defer resp.Body.Close()
	body, err := aihttp.ReadResponseBody(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, aihttp.HTTPError(resp.StatusCode, body, req.URL.String(), apiKey)
	}
	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, aihttp.ErrResponse
	}
	if len(payload.Data) > aihttp.MaxDiscoveredModels {
		return nil, fmt.Errorf("model discovery exceeded the model limit")
	}
	models := make([]string, 0, len(payload.Data))
	seen := make(map[string]struct{}, len(payload.Data))
	for _, item := range payload.Data {
		id := strings.TrimSpace(item.ID)
		if len(id) > 512 {
			return nil, aihttp.ErrResponse
		}
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		models = append(models, id)
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("no models found")
	}
	return models, nil
}

func buildModelDiscoveryURL(baseURL string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	parsed, err := url.Parse(base)
	if err != nil {
		return base
	}
	path := strings.TrimRight(parsed.Path, "/")
	for _, suffix := range []string{"/chat/completions", "/responses", "/messages", "/images/generations", "/embeddings"} {
		if strings.HasSuffix(strings.ToLower(path), suffix) {
			path = path[:len(path)-len(suffix)]
			break
		}
	}
	if !strings.HasSuffix(path, "/models") {
		if !hasAPIVersionSuffix(path) {
			path += "/v1"
		}
		path += "/models"
	}
	parsed.Path = path
	parsed.RawPath = ""
	return parsed.String()
}

func hasAPIVersionSuffix(value string) bool {
	segment := value[strings.LastIndex(value, "/")+1:]
	if len(segment) < 2 || segment[0] != 'v' {
		return false
	}
	_, err := strconv.Atoi(segment[1:])
	return err == nil
}

func DiscoverProtocolModels(provider string, config APIConfig, baseURL, apiKey, authMode string) ([]string, error) {
	if baseURL == "" {
		baseURL = config.BaseURL
	}
	if config.DiscoveryAuthMode != "" && provider != "custom" && provider != "vllm" {
		authMode = config.DiscoveryAuthMode
	}
	if config.Route != config.DiscoveryRoute && provider != "custom" && provider != "vllm" && config.Discovery != "DiscoveryOllamaModels" {
		baseURL = strings.TrimSuffix(strings.TrimRight(baseURL, "/"), config.Route) + config.DiscoveryRoute
	}
	endpoint := buildModelDiscoveryURL(baseURL)
	strategy := config.Discovery
	root, err := url.Parse(baseURL)
	if err != nil {
		return nil, aihttp.ErrTarget
	}
	root.Path, root.RawQuery = "", ""
	switch strategy {
	case "DiscoveryGeminiModels":
		endpoint = root.String() + "/v1beta/models?pageSize=100"
		authMode = "x-goog-api-key"
	case "DiscoveryAnthropicModels":
		endpoint = strings.TrimRight(baseURL, "/") + "/v1/models?limit=100"
	case "DiscoveryOllamaModels":
		endpoint = strings.TrimSuffix(strings.TrimRight(baseURL, "/"), "/v1") + "/api/tags"
	case "DiscoveryDashScopeModels":
		if config.Purpose == "text" {
			endpoint = root.String() + "/compatible-mode/v1/models"
			strategy = "DiscoveryOpenAIModels"
		} else {
			endpoint = root.String() + "/api/v1/models?page_size=100&page_no=1"
		}
	}
	if provider == "deepseek" {
		endpoint = root.String() + "/models"
	}
	if provider == "minimax" {
		endpoint = root.String() + "/v1/models"
		authMode = "bearer"
	}
	if provider == "openrouter" {
		endpoint = root.String() + "/api/v1/models?output_modalities=" + url.QueryEscape(config.Purpose)
		if config.Purpose == "embedding" {
			endpoint = root.String() + "/api/v1/models?output_modalities=embeddings"
		}
	}
	result := []string{}
	seen := map[string]bool{}
	seenPages := map[string]bool{}
	ctx, cancel := context.WithTimeout(context.Background(), defaultVerifyTimeout)
	defer cancel()
	for page := 0; page < 50; page++ {
		if seenPages[endpoint] {
			return nil, aihttp.ErrResponse
		}
		seenPages[endpoint] = true
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, aihttp.ErrTarget
		}
		if apiKey != "" {
			switch authMode {
			case "x-api-key":
				req.Header.Set("x-api-key", apiKey)
			case "x-goog-api-key":
				req.Header.Set("x-goog-api-key", apiKey)
			default:
				req.Header.Set("Authorization", "Bearer "+apiKey)
			}
		}
		req.Header.Set("anthropic-version", "2023-06-01")
		response, err := aihttp.NewClient(defaultVerifyTimeout).Do(req)
		if err != nil {
			return nil, aihttp.SafeError(err, endpoint, apiKey)
		}
		body, readErr := aihttp.ReadResponseBody(response.Body)
		response.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		if response.StatusCode != http.StatusOK {
			return nil, aihttp.HTTPError(response.StatusCode, body, endpoint, apiKey)
		}
		var payload struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
			Models []struct {
				Name  string `json:"name"`
				Model string `json:"model"`
			} `json:"models"`
			HasMore       bool   `json:"has_more"`
			LastID        string `json:"last_id"`
			NextPageToken string `json:"nextPageToken"`
			Output        struct {
				Models []struct {
					Model        string   `json:"model"`
					Capabilities []string `json:"capabilities"`
				} `json:"models"`
				Total int `json:"total"`
				Page  int `json:"page_no"`
			} `json:"output"`
		}
		if json.Unmarshal(body, &payload) != nil {
			return nil, aihttp.ErrResponse
		}
		ids := []string{}
		for _, row := range payload.Output.Models {
			for _, capability := range row.Capabilities {
				if map[string]string{"TG": "text", "IG": "image", "TR": "embedding", "ME": "embedding", "ASR": "transcription", "TTS": "speech"}[capability] == config.Purpose {
					ids = append(ids, row.Model)
					break
				}
			}
		}
		for _, row := range payload.Data {
			ids = append(ids, row.ID)
		}
		for _, row := range payload.Models {
			id := row.Name
			if row.Model != "" {
				id = row.Model
			}
			ids = append(ids, strings.TrimPrefix(id, "models/"))
		}
		for _, id := range ids {
			id = strings.TrimSpace(id)
			if id == "" || len(id) > 512 {
				continue
			}
			if !seen[id] {
				result = append(result, id)
				seen[id] = true
			}
			if len(result) > aihttp.MaxDiscoveredModels {
				return nil, fmt.Errorf("model discovery exceeded the model limit")
			}
		}
		next, _ := url.Parse(endpoint)
		query := next.Query()
		if strategy == "DiscoveryAnthropicModels" && payload.HasMore && payload.LastID != "" {
			query.Set("after_id", payload.LastID)
		} else if strategy == "DiscoveryGeminiModels" && payload.NextPageToken != "" {
			query.Set("pageToken", payload.NextPageToken)
		} else if strategy == "DiscoveryDashScopeModels" && payload.Output.Page*100 < payload.Output.Total {
			if payload.Output.Page < 1 || len(payload.Output.Models) == 0 {
				return nil, aihttp.ErrResponse
			}
			query.Set("page_no", strconv.Itoa(payload.Output.Page+1))
		} else {
			return result, nil
		}
		next.RawQuery = query.Encode()
		endpoint = next.String()
	}
	return nil, fmt.Errorf("model discovery exceeded the page limit")
}
