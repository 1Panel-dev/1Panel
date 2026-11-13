package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/aihttp"
)

type VerifyRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
}

const (
	defaultVerifyTimeout   = 30 * time.Second
	defaultVerifyMaxTokens = 16
)

func VerifyAccount(provider, apiType, authMode, baseURL, apiKey, model string) error {
	if Purpose(apiType) == "speech" || Purpose(apiType) == "transcription" {
		return fmt.Errorf("%s verification requires audio input; disable automatic verification and test with an audio client", apiType)
	}
	switch apiType {
	case "openai-completions", "openai-responses", "anthropic-messages", "gemini-generate-content", "openai-embeddings", "openai-images", "dashscope-images", "minimax-images", "openrouter-images", "tokenhub-images", "siliconflow-images", "rerank":
	default:
		return fmt.Errorf("automatic verification is not supported for %s", apiType)
	}
	req := BuildVerifyRequest(provider, apiType, authMode, baseURL, apiKey, model)
	httpReq, err := http.NewRequest(req.Method, req.URL, bytes.NewReader(req.Body))
	if err != nil {
		return buserr.WithErr("ErrAgentAccountUnavailable", aihttp.ErrTarget)
	}
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}
	httpReq.Header.Set("Accept", "application/json")
	resp, err := aihttp.NewClient(defaultVerifyTimeout).Do(httpReq)
	if err != nil {
		return buserr.WithErr("ErrAgentAccountUnavailable", aihttp.SafeError(err, req.URL, apiKey))
	}
	defer resp.Body.Close()
	body, readErr := aihttp.ReadResponseBody(resp.Body)
	if readErr != nil {
		return buserr.WithErr("ErrAgentAccountUnavailable", readErr)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return buserr.WithErr("ErrAgentAccountUnavailable", aihttp.HTTPError(resp.StatusCode, body, req.URL, apiKey))
	}
	return nil
}

func BuildVerifyRequest(provider, apiType, authMode, baseURL, apiKey, model string) VerifyRequest {
	baseURL = strings.TrimSpace(baseURL)
	if provider != "custom" || !IsImageAPIType(apiType) {
		baseURL = strings.TrimRight(baseURL, "/")
	}
	headers := map[string]string{"Content-Type": "application/json"}
	request := VerifyRequest{Method: http.MethodPost, Headers: headers}

	if apiType == "gemini-generate-content" {
		request.URL = BuildTextRequestURL(apiType, baseURL, model)
		headers["x-goog-api-key"] = apiKey
		request.Body = mustJSON(map[string]interface{}{
			"contents": []map[string]interface{}{{"parts": []map[string]string{{"text": "test"}}}},
		})
		return request
	}

	switch apiType {
	case "rerank":
		request.URL = strings.TrimRight(baseURL, "/") + "/rerank"
		headers["Authorization"] = "Bearer " + apiKey
		request.Body = mustJSON(map[string]interface{}{"model": model, "query": "test", "documents": []string{"test"}})
		if provider == "bailian" {
			if model == "qwen3-rerank" {
				request.URL = baseURL + "/compatible-api/v1/reranks"
			} else {
				request.URL = baseURL + "/api/v1/services/rerank/text-rerank/text-rerank"
				request.Body = mustJSON(map[string]interface{}{"model": model, "input": map[string]interface{}{"query": "test", "documents": []string{"test"}}})
			}
		}
	case "tokenhub-images":
		request.URL = imageVerifyURL(provider, baseURL, "/v1/wand/hunyuan-image/v3-generation")
		headers["Authorization"] = "Bearer " + apiKey
		request.Body = mustJSON(map[string]interface{}{"model": model, "prompt": "test"})
	case "siliconflow-images":
		request.URL = imageVerifyURL(provider, baseURL, "/images/generations")
		headers["Authorization"] = "Bearer " + apiKey
		request.Body = mustJSON(map[string]interface{}{"model": model, "prompt": "test", "image_size": "1024x1024", "batch_size": 1})
	case "openai-embeddings":
		request.URL = embeddingVerifyURL(baseURL)
		headers["Authorization"] = "Bearer " + apiKey
		request.Body = mustJSON(map[string]interface{}{"model": model, "input": "ping"})
	case "openai-images":
		request.URL = imageVerifyURL(provider, baseURL, "/images/generations")
		headers["Authorization"] = "Bearer " + apiKey
		request.Body = mustJSON(map[string]interface{}{"model": model, "prompt": "test", "n": 1, "response_format": "url"})
	case "dashscope-images":
		request.URL = imageVerifyURL(provider, baseURL, "/api/v1/services/aigc/multimodal-generation/generation")
		headers["Authorization"] = "Bearer " + apiKey
		request.Body = mustJSON(map[string]interface{}{
			"model": model,
			"input": map[string]interface{}{"messages": []map[string]interface{}{
				{"role": "user", "content": []map[string]string{{"text": "test"}}},
			}},
			"parameters": map[string]interface{}{"n": 1},
		})
	case "minimax-images":
		request.URL = imageVerifyURL(provider, baseURL, "/v1/image_generation")
		headers["Authorization"] = "Bearer " + apiKey
		request.Body = mustJSON(map[string]interface{}{"model": model, "prompt": "test", "n": 1, "response_format": "url"})
	case "openrouter-images":
		request.URL = imageVerifyURL(provider, baseURL, "/api/v1/images")
		headers["Authorization"] = "Bearer " + apiKey
		request.Body = mustJSON(map[string]interface{}{"model": model, "prompt": "test", "n": 1, "response_format": "url"})
	case "anthropic-messages":
		request.URL = BuildTextRequestURL(apiType, baseURL, model)
		if authMode == AuthModeBearer {
			headers["Authorization"] = "Bearer " + apiKey
		} else {
			headers["x-api-key"] = apiKey
		}
		headers["anthropic-version"] = "2023-06-01"
		body := map[string]interface{}{
			"model": model, "max_tokens": defaultVerifyMaxTokens, "stream": false,
			"messages": []map[string]interface{}{{"role": "user", "content": []map[string]string{{"type": "text", "text": "test"}}}},
		}
		if provider == "minimax" && model == "MiniMax-M3.1-Flash" {
			body["thinking"] = map[string]string{"type": "adaptive"}
			body["output_config"] = map[string]string{"effort": "max"}
		}
		request.Body = mustJSON(body)
	case "openai-responses":
		request.URL = BuildTextRequestURL(apiType, baseURL, model)
		headers["Authorization"] = "Bearer " + apiKey
		request.Body = mustJSON(map[string]interface{}{"model": model, "input": "test", "max_output_tokens": defaultVerifyMaxTokens, "stream": false})
	default:
		request.URL = BuildTextRequestURL(apiType, baseURL, model)
		if (provider != "ollama" && provider != "llmman") || strings.TrimSpace(apiKey) != "" {
			headers["Authorization"] = "Bearer " + apiKey
		}
		body := map[string]interface{}{
			"model": model, "messages": []map[string]string{{"role": "user", "content": "test"}}, "max_tokens": defaultVerifyMaxTokens, "stream": false,
		}
		if provider == "minimax" && model == "MiniMax-M3.1-Flash" {
			body["reasoning_effort"] = "max"
		}
		request.Body = mustJSON(body)
	}
	if strings.TrimSpace(apiKey) == "" && (provider == "vllm" || provider == "ollama" || provider == "llmman") {
		delete(headers, "Authorization")
	}
	return request
}

func embeddingVerifyURL(baseURL string) string {
	lowerBaseURL := strings.ToLower(baseURL)
	if strings.HasSuffix(lowerBaseURL, "/embeddings") {
		return baseURL
	}
	if hasAPIVersionSuffix(lowerBaseURL) {
		return baseURL + "/embeddings"
	}
	return baseURL + "/v1/embeddings"
}

func imageVerifyURL(provider, baseURL, endpoint string) string {
	if provider == "custom" || strings.HasSuffix(strings.ToLower(baseURL), endpoint) {
		return baseURL
	}
	return baseURL + endpoint
}

func mustJSON(value interface{}) []byte {
	payload, err := json.Marshal(value)
	if err != nil {
		return []byte("{}")
	}
	return payload
}
