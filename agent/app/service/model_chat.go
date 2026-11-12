package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/provider"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/utils/aihttp"
)

func (a AgentService) ChatAccount(ctx context.Context, input dto.ModelChat) (*http.Response, error) {
	if len(input.Messages) == 0 || input.Messages[len(input.Messages)-1].Role != "user" {
		return nil, fmt.Errorf("the last chat message must be from the user")
	}
	account, err := agentAccountRepo.GetFirst(repo.WithByID(input.ID))
	if err != nil {
		return nil, err
	}
	models, err := loadAgentAccountModels(account)
	if err != nil {
		return nil, err
	}
	var selected *dto.AgentAccountModel
	for i := range models {
		if models[i].ID == input.Model {
			selected = &models[i]
			break
		}
	}
	if selected == nil {
		return nil, fmt.Errorf("model is not configured for this service")
	}
	request, err := buildModelChatRequest(ctx, account, *selected, input.Messages)
	if err != nil {
		return nil, err
	}
	response, err := aihttp.NewClient(10 * time.Minute).Do(request)
	if err != nil {
		return nil, aihttp.SafeError(err, request.URL.String(), account.APIKey)
	}
	if response.StatusCode != http.StatusOK || !strings.HasPrefix(response.Header.Get("Content-Type"), "text/event-stream") {
		defer response.Body.Close()
		body, err := aihttp.ReadResponseBody(response.Body)
		if err != nil {
			return nil, err
		}
		if response.StatusCode != http.StatusOK {
			return nil, aihttp.HTTPError(response.StatusCode, body, request.URL.String(), account.APIKey)
		}
		return nil, aihttp.ErrResponse
	}
	return response, nil
}

func buildModelChatRequest(ctx context.Context, account *model.AgentAccount, selected dto.AgentAccountModel, messages []dto.ModelChatMessage) (*http.Request, error) {
	apiType := account.APIType
	if apiType == "" {
		apiType = provider.DefaultAPIType(account.Provider)
	}
	baseURL := strings.TrimRight(account.BaseURL, "/")
	endpoint := baseURL + "/chat/completions"
	body := map[string]interface{}{"model": selected.ID, "messages": messages, "stream": true}
	headers := map[string]string{}
	if account.APIKey != "" {
		headers["Authorization"] = "Bearer " + account.APIKey
	}
	switch apiType {
	case "openai-completions":
		if account.Provider == "minimax" && selected.ID == "MiniMax-M3.1-Flash" {
			body["reasoning_effort"] = "max"
		}
	case "openai-responses":
		endpoint = baseURL + "/responses"
		delete(body, "messages")
		body["input"] = messages
		body["store"] = false
	case "anthropic-messages":
		endpoint = baseURL + "/v1/messages"
		body["max_tokens"] = 4096
		headers["anthropic-version"] = "2023-06-01"
		if account.AuthMode != provider.AuthModeBearer {
			delete(headers, "Authorization")
			headers["x-api-key"] = account.APIKey
		}
		if account.Provider == "minimax" && selected.ID == "MiniMax-M3.1-Flash" {
			body["thinking"] = map[string]string{"type": "adaptive"}
			body["output_config"] = map[string]string{"effort": "max"}
		}
	case "gemini-generate-content":
		endpoint = baseURL + "/v1beta/models/" + url.PathEscape(selected.ID) + ":streamGenerateContent?alt=sse"
		contents := make([]map[string]interface{}, 0, len(messages))
		for _, message := range messages {
			role := message.Role
			if role == "assistant" {
				role = "model"
			}
			contents = append(contents, map[string]interface{}{"role": role, "parts": []map[string]string{{"text": message.Content}}})
		}
		body = map[string]interface{}{"contents": contents}
		delete(headers, "Authorization")
		headers["x-goog-api-key"] = account.APIKey
	default:
		return nil, fmt.Errorf("unsupported chat protocol: %s", apiType)
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, aihttp.ErrTarget
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "text/event-stream")
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	return request, nil
}
