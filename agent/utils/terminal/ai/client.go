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
	"unicode/utf8"

	providercatalog "github.com/1Panel-dev/1Panel/agent/app/provider"
	"github.com/1Panel-dev/1Panel/agent/global"
)

const (
	defaultTimeout        = 30 * time.Second
	defaultUserAgent      = "1panel-terminal-ai/1.0"
	defaultAnthropicToken = 1024
)

type Client interface {
	ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error)
}

type ClientConfig struct {
	Provider   string
	BaseURL    string
	APIKey     string
	Model      string
	APIType    string
	MaxTokens  int
	Timeout    time.Duration
	HTTPClient *http.Client
}

type GeneratorConfig struct {
	Provider  string
	BaseURL   string
	APIKey    string
	Model     string
	APIType   string
	MaxTokens int
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

type ThinkingMode struct {
	Type string `json:"type"`
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

type terminalAIClient struct {
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
	cfg.APIType = normalizeClientAPIType(providerKey, cfg.APIType)
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: cfg.Timeout}
	}
	return &terminalAIClient{
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

func (c *terminalAIClient) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("messages are required")
	}
	switch c.config.APIType {
	case "anthropic-messages":
		return c.chatCompletionAnthropic(ctx, req)
	case "gemini-generate-content":
		return c.chatCompletionGemini(ctx, req)
	case "openai-responses":
		return c.chatCompletionOpenAIResponses(ctx, req)
	default:
		return c.chatCompletionOpenAI(ctx, req)
	}
}

func (c *terminalAIClient) chatCompletionOpenAI(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	payload := openAIChatCompletionRequest{
		Model:       normalizeModelID(c.config.Model),
		Messages:    req.Messages,
		MaxTokens:   firstPositive(req.MaxTokens, c.config.MaxTokens),
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Thinking:    thinkingModeForModel(c.config.Model),
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

	respBody, err := c.do(httpReq)
	if err != nil {
		return nil, err
	}

	var completionResp openAIChatCompletionResponse
	if err := json.Unmarshal(respBody, &completionResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	content := extractAssistantContent(completionResp)
	return &ChatCompletionResponse{
		ID:      completionResp.ID,
		Model:   firstNonEmptyString(c.config.Model, completionResp.Model),
		Content: content,
		RawText: strings.TrimSpace(string(respBody)),
		Usage: ResponseUsage{
			PromptTokens:     completionResp.Usage.PromptTokens,
			CompletionTokens: completionResp.Usage.CompletionTokens,
			TotalTokens:      completionResp.Usage.TotalTokens,
		},
	}, nil
}

func (c *terminalAIClient) chatCompletionOpenAIResponses(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	payload := openAIResponsesRequest{
		Model:           normalizeModelID(c.config.Model),
		Input:           toResponsesInput(req.Messages),
		MaxOutputTokens: firstPositive(req.MaxTokens, c.config.MaxTokens),
		Temperature:     req.Temperature,
		TopP:            req.TopP,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, buildResponsesURL(c.config.BaseURL), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if strings.TrimSpace(c.config.APIKey) != "" {
		httpReq.Header.Set("Authorization", "Bearer "+strings.TrimSpace(c.config.APIKey))
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", defaultUserAgent)

	respBody, err := c.do(httpReq)
	if err != nil {
		return nil, err
	}

	var completionResp openAIResponsesResponse
	if err := json.Unmarshal(respBody, &completionResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	content := extractResponsesContent(completionResp)
	return &ChatCompletionResponse{
		ID:      completionResp.ID,
		Model:   firstNonEmptyString(c.config.Model, completionResp.Model),
		Content: content,
		RawText: strings.TrimSpace(string(respBody)),
		Usage: ResponseUsage{
			PromptTokens:     completionResp.Usage.InputTokens,
			CompletionTokens: completionResp.Usage.OutputTokens,
			TotalTokens:      completionResp.Usage.TotalTokens,
		},
	}, nil
}

func (c *terminalAIClient) chatCompletionAnthropic(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	system, messages := toAnthropicMessages(req.Messages)
	payload := anthropicMessagesRequest{
		Model:       normalizeModelID(c.config.Model),
		System:      system,
		Messages:    messages,
		MaxTokens:   firstPositive(req.MaxTokens, c.config.MaxTokens, defaultAnthropicToken),
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, buildAnthropicMessagesURL(c.config.BaseURL), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("x-api-key", strings.TrimSpace(c.config.APIKey))
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", defaultUserAgent)

	respBody, err := c.do(httpReq)
	if err != nil {
		return nil, err
	}

	var completionResp anthropicMessagesResponse
	if err := json.Unmarshal(respBody, &completionResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	content := extractAnthropicContent(completionResp)
	return &ChatCompletionResponse{
		ID:      completionResp.ID,
		Model:   firstNonEmptyString(c.config.Model, completionResp.Model),
		Content: content,
		RawText: strings.TrimSpace(string(respBody)),
		Usage: ResponseUsage{
			PromptTokens:     completionResp.Usage.InputTokens,
			CompletionTokens: completionResp.Usage.OutputTokens,
			TotalTokens:      completionResp.Usage.InputTokens + completionResp.Usage.OutputTokens,
		},
	}, nil
}

func (c *terminalAIClient) chatCompletionGemini(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	payload := geminiGenerateContentRequest{
		Contents: toGeminiContents(req.Messages),
		GenerationConfig: &geminiGenerationConfig{
			MaxOutputTokens: firstPositive(req.MaxTokens, c.config.MaxTokens),
			Temperature:     req.Temperature,
			TopP:            req.TopP,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, buildGeminiGenerateContentURL(c.config.BaseURL, normalizeModelID(c.config.Model)), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("x-goog-api-key", strings.TrimSpace(c.config.APIKey))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", defaultUserAgent)

	respBody, err := c.do(httpReq)
	if err != nil {
		return nil, err
	}

	var completionResp geminiGenerateContentResponse
	if err := json.Unmarshal(respBody, &completionResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	content := extractGeminiContent(completionResp)
	return &ChatCompletionResponse{
		Model:   firstNonEmptyString(c.config.Model, completionResp.ModelVersion),
		Content: content,
		RawText: strings.TrimSpace(string(respBody)),
		Usage: ResponseUsage{
			PromptTokens:     completionResp.Usage.PromptTokenCount,
			CompletionTokens: completionResp.Usage.CandidatesTokenCount,
			TotalTokens:      completionResp.Usage.TotalTokenCount,
		},
	}, nil
}

func (c *terminalAIClient) do(httpReq *http.Request) ([]byte, error) {
	start := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logRequestCost(httpReq, 0, time.Since(start), err)
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logRequestCost(httpReq, resp.StatusCode, time.Since(start), err)
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		providerErr := parseProviderError(resp.StatusCode, respBody)
		c.logRequestCost(httpReq, resp.StatusCode, time.Since(start), providerErr)
		return nil, providerErr
	}
	c.logRequestCost(httpReq, resp.StatusCode, time.Since(start), nil)
	return respBody, nil
}

func (c *terminalAIClient) logRequestCost(httpReq *http.Request, statusCode int, duration time.Duration, err error) {
	if global.LOG == nil || httpReq == nil {
		return
	}
	if err != nil {
		global.LOG.Warnf("[terminal-ai] request finished, provider=%s model=%s apiType=%s method=%s url=%s status=%d duration=%s err=%v",
			c.config.Provider,
			c.config.Model,
			c.config.APIType,
			httpReq.Method,
			httpReq.URL.String(),
			statusCode,
			duration,
			err,
		)
		return
	}
	global.LOG.Infof("[terminal-ai] request finished, provider=%s model=%s apiType=%s method=%s url=%s status=%d duration=%s",
		c.config.Provider,
		c.config.Model,
		c.config.APIType,
		httpReq.Method,
		httpReq.URL.String(),
		statusCode,
		duration,
	)
}

type openAIChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
	TopP        *float64      `json:"top_p,omitempty"`
	Thinking    *ThinkingMode `json:"thinking,omitempty"`
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

type openAIResponsesRequest struct {
	Model           string                 `json:"model"`
	Input           []openAIResponsesInput `json:"input"`
	MaxOutputTokens int                    `json:"max_output_tokens,omitempty"`
	Temperature     *float64               `json:"temperature,omitempty"`
	TopP            *float64               `json:"top_p,omitempty"`
}

type openAIResponsesInput struct {
	Role    string                     `json:"role"`
	Content []openAIResponsesInputPart `json:"content"`
}

type openAIResponsesInputPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type openAIResponsesResponse struct {
	ID     string `json:"id"`
	Model  string `json:"model"`
	Output []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	OutputText string `json:"output_text"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

type anthropicMessagesRequest struct {
	Model       string             `json:"model"`
	System      string             `json:"system,omitempty"`
	Messages    []anthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature *float64           `json:"temperature,omitempty"`
	TopP        *float64           `json:"top_p,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicMessagesResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type geminiGenerateContentRequest struct {
	Contents         []geminiContent         `json:"contents"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string              `json:"role,omitempty"`
	Parts []geminiContentPart `json:"parts"`
}

type geminiContentPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	Temperature     *float64 `json:"temperature,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
}

type geminiGenerateContentResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	ModelVersion string `json:"modelVersion"`
	Usage        struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
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

func buildResponsesURL(baseURL string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(baseURL, "/responses") {
		return baseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return baseURL + "/responses"
	}
	if strings.HasSuffix(parsed.Path, "/v1") {
		parsed.Path += "/responses"
	} else {
		parsed.Path = strings.TrimRight(parsed.Path, "/") + "/v1/responses"
	}
	return strings.TrimRight(parsed.String(), "/")
}

func buildAnthropicMessagesURL(baseURL string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(baseURL, "/messages") {
		return baseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return baseURL + "/messages"
	}
	if strings.HasSuffix(parsed.Path, "/v1") {
		parsed.Path += "/messages"
	} else {
		parsed.Path = strings.TrimRight(parsed.Path, "/") + "/v1/messages"
	}
	return strings.TrimRight(parsed.String(), "/")
}

func buildGeminiGenerateContentURL(baseURL, model string) string {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	model = strings.TrimSpace(model)
	if model == "" {
		model = "gemini-3-flash-preview"
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return strings.TrimRight(baseURL, "/") + "/v1beta/models/" + model + ":generateContent"
	}
	if strings.Contains(parsed.Path, "/models/") && strings.HasSuffix(parsed.Path, ":generateContent") {
		return parsed.String()
	}
	if strings.Contains(parsed.Path, "/v1beta") {
		parsed.Path = strings.TrimRight(parsed.Path, "/") + "/models/" + model + ":generateContent"
	} else {
		parsed.Path = strings.TrimRight(parsed.Path, "/") + "/v1beta/models/" + model + ":generateContent"
	}
	return parsed.String()
}

func normalizeModelID(model string) string {
	model = strings.TrimSpace(model)
	if parts := strings.SplitN(model, "/", 2); len(parts) == 2 {
		return parts[1]
	}
	return model
}

func thinkingModeForModel(model string) *ThinkingMode {
	if strings.EqualFold(normalizeModelID(model), "kimi-k2.5") {
		return &ThinkingMode{Type: "disabled"}
	}
	return nil
}

func normalizeClientAPIType(provider, apiType string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "anthropic", "kimi-coding", "minimax":
		return "anthropic-messages"
	case "gemini":
		return "gemini-generate-content"
	case "ollama":
		trim := strings.ToLower(strings.TrimSpace(apiType))
		if trim == "openai-completions" {
			return trim
		}
		return "openai-responses"
	default:
		trim := strings.ToLower(strings.TrimSpace(apiType))
		if trim == "" {
			return "openai-completions"
		}
		return trim
	}
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

func extractResponsesContent(resp openAIResponsesResponse) string {
	if text := strings.TrimSpace(resp.OutputText); text != "" {
		return text
	}
	for _, item := range resp.Output {
		for _, content := range item.Content {
			if strings.TrimSpace(content.Text) != "" {
				return strings.TrimSpace(content.Text)
			}
		}
	}
	return ""
}

func extractAnthropicContent(resp anthropicMessagesResponse) string {
	var parts []string
	for _, item := range resp.Content {
		if strings.TrimSpace(item.Text) == "" {
			continue
		}
		parts = append(parts, strings.TrimSpace(item.Text))
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func extractGeminiContent(resp geminiGenerateContentResponse) string {
	if len(resp.Candidates) == 0 {
		return ""
	}
	var parts []string
	for _, part := range resp.Candidates[0].Content.Parts {
		if strings.TrimSpace(part.Text) == "" {
			continue
		}
		parts = append(parts, strings.TrimSpace(part.Text))
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func toResponsesInput(messages []ChatMessage) []openAIResponsesInput {
	result := make([]openAIResponsesInput, 0, len(messages))
	for _, message := range messages {
		if strings.TrimSpace(message.Content) == "" {
			continue
		}
		result = append(result, openAIResponsesInput{
			Role: normalizeRole(message.Role),
			Content: []openAIResponsesInputPart{{
				Type: "input_text",
				Text: message.Content,
			}},
		})
	}
	return result
}

func toAnthropicMessages(messages []ChatMessage) (string, []anthropicMessage) {
	var systemParts []string
	result := make([]anthropicMessage, 0, len(messages))
	for _, message := range messages {
		content := strings.TrimSpace(message.Content)
		if content == "" {
			continue
		}
		role := normalizeRole(message.Role)
		if role == "system" {
			systemParts = append(systemParts, content)
			continue
		}
		if role != "assistant" {
			role = "user"
		}
		result = append(result, anthropicMessage{
			Role:    role,
			Content: content,
		})
	}
	if len(result) == 0 && len(systemParts) > 0 {
		result = append(result, anthropicMessage{
			Role:    "user",
			Content: "Generate one shell command only.",
		})
	}
	return strings.Join(systemParts, "\n\n"), result
}

func toGeminiContents(messages []ChatMessage) []geminiContent {
	result := make([]geminiContent, 0, len(messages))
	for _, message := range messages {
		content := strings.TrimSpace(message.Content)
		if content == "" {
			continue
		}
		role := normalizeRole(message.Role)
		if role == "assistant" {
			role = "model"
		} else {
			role = "user"
		}
		result = append(result, geminiContent{
			Role: role,
			Parts: []geminiContentPart{{
				Text: content,
			}},
		})
	}
	return result
}

func normalizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "assistant", "model":
		return "assistant"
	case "system":
		return "system"
	default:
		return "user"
	}
}

func parseProviderError(statusCode int, body []byte) error {
	var errResp openAIErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && strings.TrimSpace(errResp.Error.Message) != "" {
		if strings.TrimSpace(errResp.Error.Code) != "" {
			return fmt.Errorf("provider returned %d: %s (%s)", statusCode, errResp.Error.Message, errResp.Error.Code)
		}
		return fmt.Errorf("provider returned %d: %s", statusCode, errResp.Error.Message)
	}

	var generic map[string]interface{}
	if err := json.Unmarshal(body, &generic); err == nil {
		if message := extractErrorMessage(generic); message != "" {
			return fmt.Errorf("provider returned %d: %s", statusCode, message)
		}
	}

	message := strings.TrimSpace(string(body))
	if message == "" {
		message = http.StatusText(statusCode)
	}
	if !utf8.ValidString(message) {
		message = http.StatusText(statusCode)
	}
	return fmt.Errorf("provider returned %d: %s", statusCode, message)
}

func extractErrorMessage(value map[string]interface{}) string {
	if errorValue, ok := value["error"]; ok {
		switch typed := errorValue.(type) {
		case string:
			return strings.TrimSpace(typed)
		case map[string]interface{}:
			if msg, ok := typed["message"].(string); ok {
				return strings.TrimSpace(msg)
			}
			if msg, ok := typed["status"].(string); ok {
				return strings.TrimSpace(msg)
			}
		}
	}
	if msg, ok := value["message"].(string); ok {
		return strings.TrimSpace(msg)
	}
	return ""
}

func firstPositive(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
