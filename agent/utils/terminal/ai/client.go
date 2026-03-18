package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	providercatalog "github.com/1Panel-dev/1Panel/agent/app/provider"
)

const (
	defaultTimeout   = 30 * time.Second
	defaultUserAgent = "1panel-terminal-ai/1.0"
)

type Client interface {
	ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error)
}

type ClientConfig struct {
	Provider   string
	BaseURL    string
	APIKey     string
	Model      string
	Timeout    time.Duration
	HTTPClient *http.Client
}

type GeneratorConfig struct {
	Provider string
	BaseURL  string
	APIKey   string
	Model    string
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
	TopP        *float64      `json:"top_p,omitempty"`
}

type ChatCompletionResponse struct {
	ID      string        `json:"id"`
	Model   string        `json:"model"`
	Content string        `json:"content"`
	RawText string        `json:"rawText"`
	Usage   ResponseUsage `json:"usage"`
}

type ResponseUsage struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}

type openAICompatibleClient struct {
	config     ClientConfig
	httpClient *http.Client
}

func NewClient(cfg ClientConfig) (Client, error) {
	providerKey := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if providerKey == "" {
		providerKey = "custom"
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return nil, fmt.Errorf("model is required")
	}
	if strings.TrimSpace(cfg.APIKey) == "" && providerKey != "ollama" {
		return nil, fmt.Errorf("api key is required")
	}
	baseURL := normalizeBaseURL(providerKey, cfg.BaseURL)
	if baseURL == "" {
		return nil, fmt.Errorf("base url is required")
	}
	cfg.Provider = providerKey
	cfg.BaseURL = baseURL
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: cfg.Timeout}
	}
	return &openAICompatibleClient{
		config:     cfg,
		httpClient: client,
	}, nil
}

type ClientOption func(*ClientConfig)

func WithBaseURL(baseURL string) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.BaseURL = baseURL
	}
}

func WithTimeout(timeout time.Duration) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.Timeout = timeout
	}
}

func WithHTTPClient(client *http.Client) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.HTTPClient = client
	}
}

func (c *openAICompatibleClient) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("messages are required")
	}
	payload := openAIChatCompletionRequest{
		Model:       normalizeModelID(c.config.Model),
		Messages:    req.Messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, buildChatCompletionsURL(c.config.BaseURL), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+strings.TrimSpace(c.config.APIKey))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", defaultUserAgent)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, parseProviderError(resp.StatusCode, respBody)
	}

	var completionResp openAIChatCompletionResponse
	if err := json.Unmarshal(respBody, &completionResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	content := extractAssistantContent(completionResp)
	return &ChatCompletionResponse{
		ID:      completionResp.ID,
		Model:   completionResp.Model,
		Content: content,
		RawText: strings.TrimSpace(string(respBody)),
		Usage: ResponseUsage{
			PromptTokens:     completionResp.Usage.PromptTokens,
			CompletionTokens: completionResp.Usage.CompletionTokens,
			TotalTokens:      completionResp.Usage.TotalTokens,
		},
	}, nil
}

type openAIChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
	TopP        *float64      `json:"top_p,omitempty"`
}

type openAIChatCompletionResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message ChatMessage `json:"message"`
		Text    string      `json:"text"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func normalizeBaseURL(provider, rawBaseURL string) string {
	baseURL := strings.TrimSpace(rawBaseURL)
	if baseURL == "" {
		defaultBaseURL, ok := providercatalog.DefaultBaseURL(provider)
		if ok {
			baseURL = defaultBaseURL
		}
	}
	return strings.TrimRight(baseURL, "/")
}

func buildChatCompletionsURL(baseURL string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(baseURL, "/chat/completions") {
		return baseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return baseURL + "/chat/completions"
	}
	switch {
	case strings.HasSuffix(parsed.Path, "/v1"):
		parsed.Path += "/chat/completions"
	case strings.HasSuffix(parsed.Path, "/v1beta"):
		parsed.Path += "/chat/completions"
	default:
		parsed.Path = strings.TrimRight(parsed.Path, "/") + "/v1/chat/completions"
	}
	return strings.TrimRight(parsed.String(), "/")
}

func normalizeModelID(model string) string {
	model = strings.TrimSpace(model)
	if parts := strings.SplitN(model, "/", 2); len(parts) == 2 {
		return parts[1]
	}
	return model
}

func extractAssistantContent(resp openAIChatCompletionResponse) string {
	if len(resp.Choices) == 0 {
		return ""
	}
	if content := strings.TrimSpace(resp.Choices[0].Message.Content); content != "" {
		return content
	}
	return strings.TrimSpace(resp.Choices[0].Text)
}

func parseProviderError(statusCode int, body []byte) error {
	var errResp openAIErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && strings.TrimSpace(errResp.Error.Message) != "" {
		if strings.TrimSpace(errResp.Error.Code) != "" {
			return fmt.Errorf("provider returned %d: %s (%s)", statusCode, errResp.Error.Message, errResp.Error.Code)
		}
		return fmt.Errorf("provider returned %d: %s", statusCode, errResp.Error.Message)
	}
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = http.StatusText(statusCode)
	}
	return fmt.Errorf("provider returned %d: %s", statusCode, message)
}
