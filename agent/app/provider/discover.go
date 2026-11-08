package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/aihttp"
)

func DiscoverModels(baseURL, apiKey string) ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, buildModelDiscoveryURL(baseURL), nil)
	if err != nil {
		return nil, aihttp.ErrTarget
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := aihttp.NewClient(defaultVerifyTimeout).Do(req)
	if err != nil {
		return nil, aihttp.SafeError(err)
	}
	defer resp.Body.Close()
	body, err := aihttp.ReadResponseBody(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed: status %d", resp.StatusCode)
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
