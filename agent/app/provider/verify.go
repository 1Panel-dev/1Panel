package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/buserr"
)

type VerifyRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
}

func SkipVerification(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "custom", "vllm", "ollama", "kimi-coding":
		return true
	default:
		return false
	}
}

func VerifyAccount(provider, baseURL, apiKey string) error {
	req := BuildVerifyRequest(provider, baseURL, apiKey)
	var body *bytes.Buffer
	if len(req.Body) > 0 {
		body = bytes.NewBuffer(req.Body)
	} else {
		body = bytes.NewBuffer(nil)
	}
	httpReq, err := http.NewRequest(req.Method, req.URL, body)
	if err != nil {
		return err
	}
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(httpReq)
	if err != nil {
		return buserr.WithErr("ErrAgentAccountUnavailable", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return buserr.WithErr("ErrAgentAccountUnavailable", fmt.Errorf("verify failed: %s", resp.Status))
	}
	return nil
}

func BuildVerifyRequest(provider, baseURL, apiKey string) VerifyRequest {
	provider = strings.ToLower(strings.TrimSpace(provider))
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	headers := map[string]string{}
	request := VerifyRequest{Method: http.MethodGet, Headers: headers}

	switch provider {
	case "anthropic", "kimi-coding":
		headers["x-api-key"] = apiKey
		headers["anthropic-version"] = "2023-06-01"
		if strings.Contains(base, "/v1") {
			request.URL = base + "/models"
		} else {
			request.URL = base + "/v1/models"
		}
	case "gemini":
		if strings.Contains(base, "/v1beta") {
			request.URL = fmt.Sprintf("%s/models?key=%s", base, apiKey)
		} else {
			request.URL = fmt.Sprintf("%s/v1beta/models?key=%s", base, apiKey)
		}
	case "zai":
		headers["Authorization"] = fmt.Sprintf("Bearer %s", apiKey)
		request.URL = base + "/models"
	case "bailian-coding-plan":
		request.Method = http.MethodPost
		if !strings.Contains(base, "/v1") {
			base = base + "/v1"
		}
		request.URL = base + "/chat/completions"
		headers["Authorization"] = fmt.Sprintf("Bearer %s", apiKey)
		headers["Content-Type"] = "application/json"
		request.Body = mustJSON(map[string]interface{}{
			"model":      "qwen3.5-plus",
			"messages":   []map[string]string{{"role": "user", "content": "test"}},
			"max_tokens": 1,
		})
	case "ark-coding-plan":
		request.Method = http.MethodPost
		if !strings.Contains(base, "/api/coding/v3") {
			base = "https://ark.cn-beijing.volces.com/api/coding/v3"
		}
		request.URL = base + "/chat/completions"
		headers["Authorization"] = fmt.Sprintf("Bearer %s", apiKey)
		headers["Content-Type"] = "application/json"
		request.Body = mustJSON(map[string]interface{}{
			"model":      "doubao-seed-2.0-code",
			"messages":   []map[string]string{{"role": "user", "content": "test"}},
			"max_tokens": 1,
		})
	case "minimax":
		request.Method = http.MethodPost
		if !strings.Contains(base, "/v1") {
			base = base + "/v1"
		}
		request.URL = base + "/chat/completions"
		headers["Authorization"] = fmt.Sprintf("Bearer %s", apiKey)
		headers["Content-Type"] = "application/json"
		request.Body = mustJSON(map[string]interface{}{
			"model":      "MiniMax-M2.1",
			"messages":   []map[string]string{{"role": "user", "content": "test"}},
			"max_tokens": 1,
		})
	default:
		headers["Authorization"] = fmt.Sprintf("Bearer %s", apiKey)
		if strings.Contains(base, "/v1") {
			request.URL = base + "/models"
		} else {
			request.URL = base + "/v1/models"
		}
	}
	return request
}

func mustJSON(value interface{}) []byte {
	payload, err := json.Marshal(value)
	if err != nil {
		return []byte("{}")
	}
	return payload
}
