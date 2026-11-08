package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/aihttp"
)

const (
	mcpOutputTransportSSE            = "sse"
	mcpOutputTransportStreamableHTTP = "streamableHttp"
)

func probeMcpConnection(endpoint, outputTransport, protocolVersion string) (bool, string) {
	client := aihttp.NewClient(10 * time.Second)
	if outputTransport == mcpOutputTransportSSE {
		request, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return false, aihttp.ErrTarget.Error()
		}
		response, err := client.Do(request)
		if err != nil {
			return false, aihttp.SafeError(err).Error()
		}
		defer response.Body.Close()
		if response.StatusCode >= 200 && response.StatusCode < 300 && strings.Contains(response.Header.Get("Content-Type"), "text/event-stream") {
			return true, "ok"
		}
		if _, err := aihttp.ReadResponseBody(response.Body); err != nil {
			return false, aihttp.SafeError(err).Error()
		}
		return false, fmt.Sprintf("unexpected response: status=%d", response.StatusCode)
	}
	payload := map[string]interface{}{
		"jsonrpc": "2.0", "id": 1, "method": "initialize",
		"params": map[string]interface{}{
			"protocolVersion": protocolVersion,
			"capabilities":    map[string]interface{}{},
			"clientInfo":      map[string]string{"name": "1Panel", "version": "1.0.0"},
		},
	}
	body, _ := json.Marshal(payload)
	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return false, aihttp.ErrTarget.Error()
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	response, err := client.Do(request)
	if err != nil {
		return false, aihttp.SafeError(err).Error()
	}
	defer response.Body.Close()
	return handleMcpConnectionHTTPResponse(response)
}

func handleMcpConnectionHTTPResponse(response *http.Response) (bool, string) {
	if response.StatusCode >= 200 && response.StatusCode < 300 && strings.Contains(response.Header.Get("Content-Type"), "text/event-stream") {
		return true, "ok"
	}
	if _, err := aihttp.ReadResponseBody(response.Body); err != nil {
		return false, aihttp.SafeError(err).Error()
	}
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return true, "ok"
	}
	return false, fmt.Sprintf("endpoint returned status %d", response.StatusCode)
}
