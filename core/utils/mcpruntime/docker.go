package mcpruntime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
)

const (
	DockerImageRepository = "1panel/1panel-mcp-server"
	DefaultDockerImage    = DockerImageRepository + ":v0.0.1"
	DockerContainerName   = "1panel-mcp"
	dockerAgentEndpoint   = "http://agent/internal/mcp/docker"
)

var dockerDigestPattern = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)

func ValidateDockerImage(value string) error {
	if len(value) > 180 {
		return errors.New("invalid MCP Docker image")
	}
	if strings.HasPrefix(value, DockerImageRepository+"@") && dockerDigestPattern.MatchString(strings.TrimPrefix(value, DockerImageRepository+"@")) {
		return nil
	}
	if strings.HasPrefix(value, DockerImageRepository+":") && versionPattern.MatchString(strings.TrimPrefix(value, DockerImageRepository+":")) {
		return nil
	}
	return errors.New("MCP Docker image must use the official repository with a fixed semantic version or SHA256 digest")
}

type DockerManager struct {
	port           int
	image          string
	client         *http.Client
	installMu      sync.Mutex
	installImageID string
}

type dockerRequest struct {
	Operation string `json:"operation"`
	Image     string `json:"image"`
	Port      int    `json:"port"`
	Enabled   bool   `json:"enabled"`
	Restart   bool   `json:"restart"`
	ImageID   string `json:"imageID,omitempty"`
}

type dockerResponse struct {
	State *struct {
		Installed      bool   `json:"installed"`
		Status         string `json:"status"`
		Ready          bool   `json:"ready"`
		Reason         string `json:"reason"`
		AdapterVersion string `json:"adapterVersion"`
		Image          string `json:"image"`
		ImageID        string `json:"imageID"`
	} `json:"state"`
	Catalog json.RawMessage `json:"catalog"`
	Logs    string          `json:"logs"`
	Error   string          `json:"error"`
	ImageID string          `json:"imageID"`
}

func NewDocker(port int, image string) *DockerManager {
	transport := &http.Transport{Proxy: nil, DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "unix", "/etc/1panel/agent.sock")
	}, MaxIdleConns: 1, IdleConnTimeout: 30 * time.Second}
	return &DockerManager{port: port, image: image, client: &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

func (m *DockerManager) call(ctx context.Context, operation string, enabled, restart bool, imageID string) (dockerResponse, error) {
	if err := ValidateDockerImage(m.image); err != nil {
		return dockerResponse{}, err
	}
	if m.port < 1024 || m.port > 65535 {
		return dockerResponse{}, errors.New("MCP listen port must be between 1024 and 65535")
	}
	timeout := 65 * time.Second
	if operation == "catalog" || operation == "candidate_catalog" || operation == "logs" || operation == "prepare" {
		timeout = 5 * time.Second
	}
	if operation == "status" {
		timeout = 8 * time.Second
	}
	if operation == "pull" {
		timeout = 4*time.Minute + 5*time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	request := dockerRequest{Operation: operation, Image: m.image, Port: m.port, Enabled: enabled, Restart: restart, ImageID: imageID}
	data, err := json.Marshal(request)
	if err != nil {
		return dockerResponse{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, dockerAgentEndpoint, bytes.NewReader(data))
	if err != nil {
		return dockerResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := m.client.Do(req)
	defer m.client.CloseIdleConnections()
	if err != nil {
		return dockerResponse{}, errors.New("MCP Docker runtime agent is unavailable or timed out")
	}
	defer response.Body.Close()
	limit := maxCatalogBytes + 64<<10
	body, err := io.ReadAll(io.LimitReader(response.Body, int64(limit+1)))
	if err != nil || len(body) > limit {
		return dockerResponse{}, errors.New("MCP Docker agent response is incomplete or too large")
	}
	var result dockerResponse
	if contract.DecodeObjectLimit(body, &result, limit) != nil {
		return dockerResponse{}, errors.New("MCP Docker agent returned an invalid response")
	}
	if result.Error != "" {
		if len(result.Error) > 1024 {
			return dockerResponse{}, errors.New("MCP Docker operation failed")
		}
		return dockerResponse{}, errors.New(result.Error)
	}
	if response.StatusCode != http.StatusOK {
		return dockerResponse{}, errors.New("MCP Docker operation was rejected by the local agent")
	}
	return result, nil
}

func (m *DockerManager) Prepare(ctx context.Context) error {
	_, err := m.call(ctx, "prepare", false, false, "")
	return err
}

func (m *DockerManager) Catalog(ctx context.Context) ([]contract.Tool, error) {
	tools, _, err := m.readCatalog(ctx, "catalog")
	return tools, err
}

func (m *DockerManager) readCatalog(ctx context.Context, operation string) ([]contract.Tool, string, error) {
	select {
	case catalogSlots <- struct{}{}:
		defer func() { <-catalogSlots }()
	default:
		return nil, "", errors.New("MCP tool catalog is busy; retry shortly")
	}
	response, err := m.call(ctx, operation, false, false, "")
	if err != nil {
		return nil, "", err
	}
	var manifest struct {
		ContractVersion int             `json:"contractVersion"`
		Tools           []contract.Tool `json:"tools"`
	}
	if !dockerDigestPattern.MatchString(response.ImageID) || contract.DecodeObjectLimit(response.Catalog, &manifest, maxCatalogBytes) != nil || manifest.ContractVersion != ContractVersion {
		return nil, "", errors.New("MCP Docker catalog is invalid or incompatible; update the panel and adapter together")
	}
	tools, err := contract.ValidateCatalog(manifest.Tools)
	return tools, response.ImageID, err
}

func (m *DockerManager) Status(ctx context.Context, enabled bool) State {
	response, err := m.call(ctx, "status", enabled, false, "")
	if err != nil {
		return State{Status: "unsupported", Reason: err.Error(), Image: m.image}
	}
	if response.State == nil {
		return State{Status: "unsupported", Reason: "MCP Docker agent returned no state", Image: m.image}
	}
	s := response.State
	return State{Installed: s.Installed, Status: s.Status, Ready: s.Ready, Reason: s.Reason, AdapterVersion: s.AdapterVersion, Image: s.Image, ImageID: s.ImageID}
}

func (m *DockerManager) Apply(ctx context.Context, enabled, restart bool) error {
	var imageID string
	if enabled {
		var err error
		if _, imageID, err = m.readCatalog(ctx, "candidate_catalog"); err != nil {
			return err
		}
	}
	_, err := m.call(ctx, "apply", enabled, restart, imageID)
	return err
}

func (m *DockerManager) InstallImage(ctx context.Context, enabled bool) error {
	m.installMu.Lock()
	m.installImageID = ""
	m.installMu.Unlock()
	if _, err := m.call(ctx, "pull", false, false, ""); err != nil {
		return err
	}
	_, imageID, err := m.readCatalog(ctx, "candidate_catalog")
	if err != nil {
		return err
	}
	m.installMu.Lock()
	m.installImageID = imageID
	m.installMu.Unlock()
	_, err = m.call(ctx, "install", enabled, true, imageID)
	return err
}

func (m *DockerManager) RestoreAbsent(ctx context.Context) error {
	m.installMu.Lock()
	imageID := m.installImageID
	m.installMu.Unlock()
	if imageID == "" {
		return nil
	}
	_, err := m.call(ctx, "rollback_absent", false, false, imageID)
	return err
}

func (m *DockerManager) Restore(ctx context.Context, imageID string, enabled bool) error {
	if imageID != "" && !dockerDigestPattern.MatchString(imageID) {
		return errors.New("invalid MCP rollback image identity")
	}
	if enabled && imageID == "" {
		return errors.New("MCP rollback requires the previous image identity")
	}
	_, err := m.call(ctx, "rollback", enabled, false, imageID)
	return err
}

func (m *DockerManager) Logs(ctx context.Context) (string, error) {
	response, err := m.call(ctx, "logs", false, false, "")
	if err != nil {
		return "", err
	}
	if len(response.Logs) > 64<<10 {
		return "", errors.New("MCP Docker logs exceed permitted size")
	}
	return redactLog(response.Logs), nil
}
