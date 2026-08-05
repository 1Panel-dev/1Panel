package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
)

type qwenPawAuthStatus struct {
	Enabled  bool `json:"enabled"`
	HasUsers bool `json:"has_users"`
}

type qwenPawLoginResponse struct {
	Token string `json:"token"`
}

func updateQwenPawDashboardAuth(install *model.AppInstall, next agentDashboardAuth) error {
	if install == nil || install.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	current, err := readAgentDashboardAuthEnv(install.GetEnvPath(), constant.AppCopaw)
	if err != nil {
		return err
	}
	if current == next {
		return writeAgentDashboardAuthEnv(install.GetEnvPath(), constant.AppCopaw, next, true)
	}
	if err := ensureContainerRunning(install.ContainerName); err != nil {
		return err
	}

	baseURL := fmt.Sprintf("http://127.0.0.1:%d/api/auth", install.HttpPort)
	var status qwenPawAuthStatus
	if _, err := requestQwenPawAuth(http.MethodGet, baseURL+"/status", nil, "", &status); err != nil {
		return buserr.WithMap("ErrQwenPawAuthRequest", map[string]interface{}{"err": err.Error()}, err)
	}
	if !status.Enabled {
		return buserr.New("ErrQwenPawAuthDisabled")
	}

	if !status.HasUsers {
		payload := map[string]string{"username": next.Username, "password": next.Password}
		if _, err := requestQwenPawAuth(http.MethodPost, baseURL+"/register", payload, "", nil); err != nil {
			return buserr.WithMap("ErrQwenPawAuthRequest", map[string]interface{}{"err": err.Error()}, err)
		}
	} else {
		var login qwenPawLoginResponse
		payload := map[string]string{"username": current.Username, "password": current.Password}
		statusCode, err := requestQwenPawAuth(http.MethodPost, baseURL+"/login", payload, "", &login)
		if statusCode == http.StatusUnauthorized {
			return buserr.New("ErrQwenPawAuthOutOfSync")
		}
		if err != nil {
			return buserr.WithMap("ErrQwenPawAuthRequest", map[string]interface{}{"err": err.Error()}, err)
		}
		payload = map[string]string{"current_password": current.Password}
		if current.Username != next.Username {
			payload["new_username"] = next.Username
		}
		if current.Password != next.Password {
			payload["new_password"] = next.Password
		}
		if _, err := requestQwenPawAuth(http.MethodPost, baseURL+"/update-profile", payload, login.Token, nil); err != nil {
			return buserr.WithMap("ErrQwenPawAuthRequest", map[string]interface{}{"err": err.Error()}, err)
		}
	}
	return writeAgentDashboardAuthEnv(install.GetEnvPath(), constant.AppCopaw, next, true)
}

func requestQwenPawAuth(method, reqURL string, payload interface{}, token string, result interface{}) (int, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return 0, err
		}
		body = bytes.NewReader(data)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return resp.StatusCode, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		detail := strings.TrimSpace(string(data))
		var errorResponse struct {
			Detail string `json:"detail"`
		}
		if json.Unmarshal(data, &errorResponse) == nil && strings.TrimSpace(errorResponse.Detail) != "" {
			detail = strings.TrimSpace(errorResponse.Detail)
		}
		if detail == "" {
			detail = resp.Status
		}
		return resp.StatusCode, errors.New(detail)
	}
	if result != nil && len(data) > 0 {
		if err := json.Unmarshal(data, result); err != nil {
			return resp.StatusCode, err
		}
	}
	return resp.StatusCode, nil
}
