package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func DiscoverModels(baseURL, apiKey string) ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, buildModelDiscoveryURL(baseURL), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := (&http.Client{Timeout: defaultVerifyTimeout}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}
	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	models := make([]string, 0, len(payload.Data))
	seen := make(map[string]struct{}, len(payload.Data))
	for _, item := range payload.Data {
		id := strings.TrimSpace(item.ID)
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
	for _, apiType := range []string{"openai-completions", "openai-responses", "anthropic-messages"} {
		base = normalizeEndpointPath(apiType, base)
	}
	switch {
	case strings.HasSuffix(base, "/models"):
		return base
	case hasAPIVersionSuffix(base):
		return base + "/models"
	default:
		return base + "/v1/models"
	}
}

func hasAPIVersionSuffix(value string) bool {
	segment := value[strings.LastIndex(value, "/")+1:]
	if len(segment) < 2 || segment[0] != 'v' {
		return false
	}
	_, err := strconv.Atoi(segment[1:])
	return err == nil
}
