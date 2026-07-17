package provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

type verifyErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
	Message string `json:"message"`
}

const defaultVerifyTimeout = 30 * time.Second

func SkipVerification(provider string) bool {
	switch provider {
	case "vllm", "ollama", "kimi-coding":
		return true
	default:
		return false
	}
}

func VerifyAccount(provider, apiType, authMode, baseURL, apiKey, model string) error {
	req := BuildVerifyRequest(provider, apiType, authMode, baseURL, apiKey, model)
	httpReq, err := http.NewRequest(req.Method, req.URL, bytes.NewReader(req.Body))
	if err != nil {
		return err
	}
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}
	httpReq.Header.Set("Accept", "application/json")
	resp, err := (&http.Client{Timeout: defaultVerifyTimeout}).Do(httpReq)
	if err != nil {
		return buserr.WithErr("ErrAgentAccountUnavailable", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
		if readErr != nil {
			return buserr.WithErr("ErrAgentAccountUnavailable", readErr)
		}
		return buserr.WithErr("ErrAgentAccountUnavailable", errors.New(verifyHTTPError(resp.StatusCode, body)))
	}
	return nil
}

func BuildVerifyRequest(provider, apiType, authMode, baseURL, apiKey, model string) VerifyRequest {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	headers := map[string]string{"Content-Type": "application/json"}
	request := VerifyRequest{Method: http.MethodPost, Headers: headers}

	if provider == "gemini" {
		request.URL = baseURL + "/v1beta/models/" + strings.TrimSpace(model) + ":generateContent"
		headers["x-goog-api-key"] = apiKey
		request.Body = mustJSON(map[string]interface{}{
			"contents": []map[string]interface{}{{"parts": []map[string]string{{"text": "test"}}}},
		})
		return request
	}

	switch apiType {
	case "anthropic-messages":
		request.URL = baseURL + "/v1/messages"
		if authMode == AuthModeBearer {
			headers["Authorization"] = "Bearer " + apiKey
		} else {
			headers["x-api-key"] = apiKey
		}
		headers["anthropic-version"] = "2023-06-01"
		request.Body = mustJSON(map[string]interface{}{
			"model": model, "max_tokens": 1, "stream": false,
			"messages": []map[string]interface{}{{"role": "user", "content": []map[string]string{{"type": "text", "text": "test"}}}},
		})
	case "openai-responses":
		request.URL = baseURL + "/responses"
		headers["Authorization"] = "Bearer " + apiKey
		request.Body = mustJSON(map[string]interface{}{"model": model, "input": "test", "max_output_tokens": 1, "stream": false})
	default:
		request.URL = baseURL + "/chat/completions"
		if provider != "ollama" || strings.TrimSpace(apiKey) != "" {
			headers["Authorization"] = "Bearer " + apiKey
		}
		request.Body = mustJSON(map[string]interface{}{
			"model": model, "messages": []map[string]string{{"role": "user", "content": "test"}}, "max_tokens": 1, "stream": false,
		})
	}
	return request
}

func verifyHTTPError(statusCode int, body []byte) string {
	message := strings.TrimSpace(string(body))
	var payload verifyErrorResponse
	if err := json.Unmarshal(body, &payload); err == nil {
		if value := strings.TrimSpace(payload.Error.Message); value != "" {
			message = value
		} else if value := strings.TrimSpace(payload.Message); value != "" {
			message = value
		}
	}
	if message == "" {
		return fmt.Sprintf("validation request returned status %d", statusCode)
	}
	return message
}

func mustJSON(value interface{}) []byte {
	payload, err := json.Marshal(value)
	if err != nil {
		return []byte("{}")
	}
	return payload
}
