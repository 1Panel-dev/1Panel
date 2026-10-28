package mcpruntime

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"time"

	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
)

var pullLayerID = regexp.MustCompile(`^[a-f0-9]{1,64}$`)

func (m *DockerManager) pull(ctx context.Context) error {
	if err := ValidateDockerImage(m.image); err != nil {
		return err
	}
	if m.port < 1024 || m.port > 65535 {
		return errors.New("MCP listen port must be between 1024 and 65535")
	}
	ctx, cancel := context.WithTimeout(ctx, 4*time.Minute+5*time.Second)
	defer cancel()
	data, err := json.Marshal(dockerRequest{Operation: "pull", Image: m.image, Port: m.port, StreamProgress: true})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, dockerAgentEndpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := m.client.Do(req)
	defer m.client.CloseIdleConnections()
	if err != nil {
		return operationError("MCP Docker runtime agent is unavailable or timed out")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || response.Header.Get("Content-Type") != "application/x-ndjson" {
		return operationError("MCP Docker image download was rejected by the local agent")
	}
	reader := bufio.NewScanner(io.LimitReader(response.Body, (16<<20)+1))
	reader.Buffer(make([]byte, 4096), 4096)
	var completed bool
	var operationErr error
	var size int
	for reader.Scan() {
		line := reader.Bytes()
		size += len(line) + 1
		if size > 16<<20 || completed {
			return operationError("MCP Docker image progress response is invalid or too large")
		}
		var event struct {
			Progress
			Type  string `json:"type"`
			Error string `json:"error,omitempty"`
		}
		if contract.DecodeObjectLimit(line, &event, 4096) != nil {
			return operationError("MCP Docker image progress response is invalid")
		}
		switch event.Type {
		case "progress":
			if event.Error != "" || !validPullProgress(event.Progress) {
				return operationError("MCP Docker image progress response is invalid")
			}
			ReportProgress(ctx, event.Progress)
		case "result":
			if event.Status != "" || event.ID != "" || event.Current != 0 || event.Total != 0 || len(event.Error) > 1024 {
				return operationError("MCP Docker image progress result is invalid")
			}
			completed = true
			if event.Error != "" {
				operationErr = operationError(event.Error)
			}
		default:
			return operationError("MCP Docker image progress response is invalid")
		}
	}
	if reader.Err() != nil || !completed {
		return operationError("MCP Docker image download was interrupted")
	}
	return operationErr
}

func validPullProgress(progress Progress) bool {
	switch progress.Status {
	case "checking_image", "local_image", "pulling", "downloading", "extracting", "waiting", "verifying_checksum", "download_complete", "pull_complete", "already_exists", "verifying_image", "verified_image":
	default:
		return false
	}
	return (progress.ID == "" || pullLayerID.MatchString(progress.ID)) && progress.Current >= 0 && progress.Total >= 0 &&
		(progress.Total == 0 || progress.Current <= progress.Total)
}
