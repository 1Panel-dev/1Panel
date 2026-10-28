package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
	"github.com/google/uuid"
	"golang.org/x/net/idna"
	"gorm.io/gorm"
)

type mcpWebsiteAgentRequest struct {
	Operation string `json:"operation"`
	WebsiteID uint   `json:"websiteId"`
	BindingID string `json:"bindingId"`
	RequestID string `json:"requestId"`
	Domain    string `json:"domain"`
	EnableSSL bool   `json:"enableSSL"`
	SSLID     uint   `json:"sslID"`
	Upstream  string `json:"upstream"`
	CoreHost  string `json:"coreHost"`
	ProxyKey  string `json:"proxyKey"`
}

type mcpWebsiteAgentResponse struct {
	State         string `json:"state"`
	WebsiteID     uint   `json:"websiteId"`
	BindingID     string `json:"bindingId"`
	RequestID     string `json:"requestId"`
	Domain        string `json:"domain"`
	Endpoint      string `json:"endpoint"`
	EnableSSL     bool   `json:"enableSSL"`
	SSLID         uint   `json:"sslID"`
	WebsiteStatus string `json:"websiteStatus"`
	Reason        string `json:"reason,omitempty"`
	Error         string `json:"error,omitempty"`
	Code          string `json:"code,omitempty"`
	Definitive    bool   `json:"definitive,omitempty"`
}

type mcpWebsiteTransport interface {
	Call(context.Context, mcpWebsiteAgentRequest) (mcpWebsiteAgentResponse, error)
}

type mcpWebsiteHTTPTransport struct{}

func (mcpWebsiteHTTPTransport) Call(ctx context.Context, request mcpWebsiteAgentRequest) (mcpWebsiteAgentResponse, error) {
	var result mcpWebsiteAgentResponse
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	transport := &http.Transport{Proxy: nil, DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "unix", "/etc/1panel/agent.sock")
	}}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	data, err := json.Marshal(request)
	if err != nil {
		return result, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://agent/internal/mcp/website", bytes.NewReader(data))
	if err != nil {
		return result, err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return result, errors.New("MCP website operation could not be confirmed")
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, (64<<10)+1))
	if err != nil || len(body) > 64<<10 || json.Unmarshal(body, &result) != nil {
		return mcpWebsiteAgentResponse{}, errors.New("MCP website operation could not be confirmed")
	}
	if response.StatusCode != http.StatusOK && result.Error == "" {
		return mcpWebsiteAgentResponse{}, errors.New("MCP website operation could not be confirmed")
	}
	return result, nil
}

type mcpWebsiteOperation struct {
	dto.MCPWebsitePendingRequest
	Hash      string `json:"hash"`
	BindingID string `json:"bindingId"`
	State     string `json:"state"`
	ErrorCode string `json:"errorCode,omitempty"`
}

func (s *MCPRuntimeService) websiteAgent() mcpWebsiteTransport {
	if s.websiteTransport != nil {
		return s.websiteTransport
	}
	return mcpWebsiteHTTPTransport{}
}

func (s *MCPRuntimeService) websiteSetting(ctx context.Context) (model.MCPSetting, error) {
	row, err := (repo.MCPRepo{DB: global.DB.WithContext(ctx)}).Setting()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.MCPSetting{Revision: 1}, nil
	}
	return row, err
}

func decodeMCPWebsiteOperation(value string) (*mcpWebsiteOperation, error) {
	if value == "" {
		return nil, nil
	}
	var operation mcpWebsiteOperation
	if json.Unmarshal([]byte(value), &operation) != nil || !validMCPWebsiteRequestID(operation.RequestID) ||
		!validMCPWebsiteRequestID(operation.BindingID) || len(operation.Hash) != 64 ||
		(operation.State != "pending" && operation.State != "succeeded" && operation.State != "failed") ||
		(operation.Operation != "bind" && operation.Operation != "update" && operation.Operation != "unbind") {
		return nil, contract.Fail("MCP_WEBSITE_UNAVAILABLE", 503, "MCP website state is unavailable")
	}
	return &operation, nil
}

func validMCPWebsiteRequestID(id string) bool {
	parsed, err := uuid.Parse(id)
	return err == nil && parsed != uuid.Nil && parsed.String() == id
}

func normalizeMCPWebsiteDomain(domain string) (string, error) {
	if domain == "" || domain != strings.TrimSpace(domain) || len(domain) > 320 || strings.ContainsAny(domain, "\\\r\n\t /?#@\x00") {
		return "", contract.Invalid()
	}
	u, err := url.Parse("http://" + domain)
	if err != nil || u.Host == "" || u.Hostname() == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return "", contract.Invalid()
	}
	host := strings.ToLower(u.Hostname())
	if ip := net.ParseIP(host); ip == nil {
		host, err = idna.Lookup.ToASCII(host)
		if err != nil || len(host) > 253 {
			return "", contract.Invalid()
		}
		for _, label := range strings.Split(host, ".") {
			if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
				return "", contract.Invalid()
			}
			for _, ch := range label {
				if !(ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' || ch == '-') {
					return "", contract.Invalid()
				}
			}
		}
	} else {
		host = ip.String()
	}
	if port := u.Port(); port != "" {
		number, err := strconv.Atoi(port)
		if err != nil || number < 1 || number > 65535 {
			return "", contract.Invalid()
		}
		return net.JoinHostPort(host, strconv.Itoa(number)), nil
	}
	if strings.HasSuffix(domain, ":") {
		return "", contract.Invalid()
	}
	if strings.Contains(host, ":") {
		return "[" + host + "]", nil
	}
	return host, nil
}

func validMCPWebsiteMetadata(response mcpWebsiteAgentResponse) bool {
	if !mcpDocumentEndpoint(response.Endpoint) || response.EnableSSL && response.SSLID == 0 || !response.EnableSSL && response.SSLID != 0 {
		return false
	}
	domain, err := normalizeMCPWebsiteDomain(response.Domain)
	if err != nil {
		return false
	}
	domainURL, _ := url.Parse("http://" + domain)
	endpoint, _ := url.Parse(response.Endpoint)
	return strings.EqualFold(domainURL.Hostname(), endpoint.Hostname()) && (endpoint.Scheme == "https") == response.EnableSSL
}

func (s *MCPRuntimeService) websiteAgentRequest(ctx context.Context, row model.MCPSetting, operation string, operationID string) (mcpWebsiteAgentRequest, error) {
	masterKey, err := encrypt.APIKeyEncryptionKey()
	if err != nil {
		return mcpWebsiteAgentRequest{}, contract.Fail("MCP_WEBSITE_UNAVAILABLE", 503, "MCP website state is unavailable")
	}
	key, err := contract.WebsiteProxyKey(masterKey, row.WebsiteBindingID)
	if err != nil {
		return mcpWebsiteAgentRequest{}, contract.Fail("MCP_WEBSITE_UNAVAILABLE", 503, "MCP website state is unavailable")
	}
	host := strings.Trim(global.CONF.Conn.BindAddress, "[]")
	ip := net.ParseIP(host)
	if host == "" || host == "localhost" || host == "0.0.0.0" {
		host = "127.0.0.1"
	} else if ip != nil && ip.IsUnspecified() {
		host = "::1"
	} else if ip == nil {
		return mcpWebsiteAgentRequest{}, contract.Fail("MCP_WEBSITE_UNAVAILABLE", 503, "MCP website upstream is unavailable")
	}
	port, err := strconv.Atoi(global.CONF.Conn.Port)
	if err != nil || port < 1 || port > 65535 {
		return mcpWebsiteAgentRequest{}, contract.Fail("MCP_WEBSITE_UNAVAILABLE", 503, "MCP website upstream is unavailable")
	}
	scheme := "http"
	if global.CONF.Conn.SSL == constant.StatusEnable || global.CONF.Conn.SSL == constant.StatusMux {
		scheme = "https"
	}
	address := net.JoinHostPort(host, strconv.Itoa(port))
	coreHost := address
	if binding, err := (repo.MCPRepo{DB: global.DB.WithContext(ctx)}).PanelSetting("BindDomain"); err == nil && binding.Value != "" {
		coreHost, err = normalizeMCPWebsiteDomain(binding.Value)
		if err != nil {
			return mcpWebsiteAgentRequest{}, contract.Fail("MCP_WEBSITE_UNAVAILABLE", 503, "MCP website upstream is unavailable")
		}
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return mcpWebsiteAgentRequest{}, err
	}
	return mcpWebsiteAgentRequest{Operation: operation, WebsiteID: row.WebsiteID, BindingID: row.WebsiteBindingID,
		RequestID: operationID, Upstream: (&url.URL{Scheme: scheme, Host: address, Path: "/mcp"}).String(), CoreHost: coreHost, ProxyKey: key}, nil
}

func (s *MCPRuntimeService) WebsiteStatus(ctx context.Context) (dto.MCPWebsiteStatus, error) {
	row, err := s.websiteSetting(ctx)
	if err != nil {
		return dto.MCPWebsiteStatus{}, err
	}
	return s.websiteStatus(ctx, row)
}

func (s *MCPRuntimeService) websiteStatus(ctx context.Context, row model.MCPSetting) (dto.MCPWebsiteStatus, error) {
	result := dto.MCPWebsiteStatus{State: "unbound", Revision: row.Revision, WebsiteID: row.WebsiteID}
	operation, err := decodeMCPWebsiteOperation(row.WebsiteOperation)
	if err != nil {
		return result, err
	}
	if operation != nil {
		result.RequestID = operation.RequestID
	}
	if row.WebsiteID == 0 && row.WebsiteBindingID == "" && (operation == nil || operation.State != "pending") {
		return result, nil
	}
	result.State = "unavailable"
	if operation != nil && operation.State == "pending" {
		result.State = "pending"
		pending := operation.MCPWebsitePendingRequest
		result.PendingRequest = &pending
		result.Domain, result.EnableSSL, result.SSLID = pending.Domain, pending.EnableSSL, pending.SSLID
	}
	request, err := s.websiteAgentRequest(ctx, row, "status", result.RequestID)
	if err != nil {
		result.Reason = "MCP_WEBSITE_UNAVAILABLE"
		return result, nil
	}
	response, err := s.websiteAgent().Call(ctx, request)
	if err != nil || response.Error != "" || response.BindingID != row.WebsiteBindingID ||
		(row.WebsiteID != 0 && response.WebsiteID != row.WebsiteID) {
		result.Reason = "MCP_WEBSITE_UNAVAILABLE"
		return result, nil
	}
	result.WebsiteID, result.Domain = response.WebsiteID, response.Domain
	result.EnableSSL, result.SSLID, result.WebsiteStatus = response.EnableSSL, response.SSLID, response.WebsiteStatus
	if result.State == "pending" {
		return result, nil
	}
	if row.WebsiteID != 0 && response.State == "bound" && response.WebsiteID != 0 && validMCPWebsiteMetadata(response) {
		result.State, result.Endpoint = "bound", response.Endpoint
	} else {
		result.Reason = "MCP_WEBSITE_UNAVAILABLE"
	}
	return result, nil
}

func (s *MCPRuntimeService) BindWebsite(ctx context.Context, req dto.MCPWebsiteBind) (dto.MCPWebsiteStatus, error) {
	domain, err := normalizeMCPWebsiteDomain(req.Domain)
	if err != nil {
		return dto.MCPWebsiteStatus{}, err
	}
	parsed, _ := url.Parse("http://" + domain)
	if net.ParseIP(parsed.Hostname()) != nil {
		return dto.MCPWebsiteStatus{}, contract.Invalid()
	}
	return s.mutateWebsite(ctx, dto.MCPWebsitePendingRequest{Operation: "bind", RequestID: req.RequestID, Domain: domain, EnableSSL: req.EnableSSL, SSLID: req.SSLID}, req.Revision)
}

func (s *MCPRuntimeService) UpdateWebsite(ctx context.Context, req dto.MCPWebsiteUpdate) (dto.MCPWebsiteStatus, error) {
	return s.mutateWebsite(ctx, dto.MCPWebsitePendingRequest{Operation: "update", RequestID: req.RequestID, EnableSSL: req.EnableSSL, SSLID: req.SSLID}, req.Revision)
}

func (s *MCPRuntimeService) UnbindWebsite(ctx context.Context, req dto.MCPWebsiteUnbind) (dto.MCPWebsiteStatus, error) {
	return s.mutateWebsite(ctx, dto.MCPWebsitePendingRequest{Operation: "unbind", RequestID: req.RequestID}, req.Revision)
}

func mcpWebsiteOperationError(code string, uncertain bool) error {
	if uncertain {
		return contract.Fail("MCP_WEBSITE_PENDING", 503, "MCP website operation could not be confirmed")
	}
	switch code {
	case "MCP_INVALID_ARGUMENTS", "MCP_WEBSITE_INVALID", "MCP_WEBSITE_INVALID_ARGUMENTS":
		return contract.Invalid()
	case "MCP_CONFLICT", "MCP_WEBSITE_CONFLICT":
		return contract.Fail("MCP_CONFLICT", 409, "MCP website configuration changed")
	case "MCP_FORBIDDEN":
		return contract.Forbidden()
	default:
		return contract.Fail("MCP_WEBSITE_UNAVAILABLE", 503, "MCP website operation failed")
	}
}

func (s *MCPRuntimeService) mutateWebsite(ctx context.Context, pending dto.MCPWebsitePendingRequest, revision uint64) (dto.MCPWebsiteStatus, error) {
	if err := lockMCPRuntimeMutation(); err != nil {
		return dto.MCPWebsiteStatus{}, err
	}
	defer mcpRuntimeMutation.Unlock()
	return s.mutateWebsiteLocked(ctx, pending, revision)
}

func (s *MCPRuntimeService) mutateWebsiteLocked(ctx context.Context, pending dto.MCPWebsitePendingRequest, revision uint64) (dto.MCPWebsiteStatus, error) {
	if !validMCPWebsiteRequestID(pending.RequestID) || revision == 0 ||
		(pending.Operation != "unbind" && (pending.EnableSSL && pending.SSLID == 0 || !pending.EnableSSL && pending.SSLID != 0)) {
		return dto.MCPWebsiteStatus{}, contract.Invalid()
	}
	row, err := s.websiteSetting(ctx)
	if err != nil {
		return dto.MCPWebsiteStatus{}, err
	}
	previous, err := decodeMCPWebsiteOperation(row.WebsiteOperation)
	if err != nil {
		return dto.MCPWebsiteStatus{}, err
	}
	encoded, _ := json.Marshal(pending)
	digest := sha256.Sum256(encoded)
	hash := hex.EncodeToString(digest[:])
	operation := mcpWebsiteOperation{MCPWebsitePendingRequest: pending, Hash: hash, State: "pending", BindingID: row.WebsiteBindingID}
	if previous != nil && previous.RequestID == pending.RequestID {
		if previous.Hash != hash {
			return dto.MCPWebsiteStatus{}, contract.Fail("MCP_CONFLICT", 409, "MCP website request was already used")
		}
		if previous.State == "succeeded" {
			return s.websiteStatus(ctx, row)
		}
		if previous.State == "failed" {
			return dto.MCPWebsiteStatus{}, mcpWebsiteOperationError(previous.ErrorCode, false)
		}
		operation = *previous
	} else {
		cancelPendingChange := previous != nil && previous.State == "pending" &&
			(previous.Operation == "bind" || previous.Operation == "update") && pending.Operation == "unbind"
		if previous != nil && previous.State == "pending" && !cancelPendingChange || row.Revision != revision {
			return dto.MCPWebsiteStatus{}, ErrMCPSettingsRevision
		}
		if pending.Operation == "bind" {
			if row.WebsiteID != 0 {
				return dto.MCPWebsiteStatus{}, contract.Fail("MCP_CONFLICT", 409, "MCP already has a bound website")
			}
			settings, err := s.settings(ctx)
			if err != nil {
				return dto.MCPWebsiteStatus{}, err
			}
			if !s.manager(settings).Status(ctx, settings.Enabled).Installed {
				return dto.MCPWebsiteStatus{}, contract.Fail("MCP_WEBSITE_SERVICE_REQUIRED", 409, "Install the MCP service before binding a website")
			}
			operation.BindingID = pending.RequestID
			row.WebsiteBindingID = operation.BindingID
		} else if row.WebsiteID == 0 && !cancelPendingChange || !validMCPWebsiteRequestID(row.WebsiteBindingID) {
			return dto.MCPWebsiteStatus{}, contract.Fail("MCP_CONFLICT", 409, "MCP has no bound website")
		}
		if _, err := s.websiteAgentRequest(ctx, row, pending.Operation, pending.RequestID); err != nil {
			return dto.MCPWebsiteStatus{}, err
		}
		if err := s.saveWebsiteOperation(ctx, &row, operation, row.WebsiteID, row.WebsiteBindingID); err != nil {
			return dto.MCPWebsiteStatus{}, err
		}
	}
	request, err := s.websiteAgentRequest(ctx, row, pending.Operation, pending.RequestID)
	if err != nil {
		return dto.MCPWebsiteStatus{}, mcpWebsiteOperationError("", true)
	}
	request.Domain, request.EnableSSL, request.SSLID = pending.Domain, pending.EnableSSL, pending.SSLID
	response, callErr := s.websiteAgent().Call(ctx, request)
	if callErr != nil {
		return dto.MCPWebsiteStatus{}, mcpWebsiteOperationError("", true)
	}
	if response.Error != "" {
		if !response.Definitive || pending.Operation == "unbind" && row.WebsiteID == 0 && row.WebsiteBindingID != "" {
			return dto.MCPWebsiteStatus{}, mcpWebsiteOperationError("", true)
		}
		operation.State, operation.ErrorCode = "failed", response.Code
		bindingID := row.WebsiteBindingID
		if pending.Operation == "bind" && row.WebsiteID == 0 {
			bindingID = ""
		}
		if err := s.saveWebsiteOperation(ctx, &row, operation, row.WebsiteID, bindingID); err != nil {
			return dto.MCPWebsiteStatus{}, mcpWebsiteOperationError("", true)
		}
		return dto.MCPWebsiteStatus{}, mcpWebsiteOperationError(response.Code, false)
	}
	if response.BindingID != operation.BindingID || response.RequestID != pending.RequestID ||
		(pending.Operation != "unbind" && (response.State != "bound" || response.WebsiteID == 0 || !validMCPWebsiteMetadata(response))) ||
		(pending.Operation != "unbind" && (response.EnableSSL != pending.EnableSSL || response.SSLID != pending.SSLID)) ||
		(pending.Operation == "update" && response.WebsiteID != row.WebsiteID) ||
		(pending.Operation == "unbind" && response.State != "unbound") {
		return dto.MCPWebsiteStatus{}, mcpWebsiteOperationError("", true)
	}
	websiteID, bindingID := response.WebsiteID, operation.BindingID
	if pending.Operation == "unbind" {
		websiteID, bindingID = 0, ""
	}
	operation.State = "succeeded"
	if err := s.saveWebsiteOperation(ctx, &row, operation, websiteID, bindingID); err != nil {
		return dto.MCPWebsiteStatus{}, mcpWebsiteOperationError("", true)
	}
	if pending.Operation == "unbind" {
		return dto.MCPWebsiteStatus{State: "unbound", Revision: row.Revision, RequestID: pending.RequestID}, nil
	}
	return dto.MCPWebsiteStatus{State: "bound", Revision: row.Revision, WebsiteID: response.WebsiteID, Domain: response.Domain,
		Endpoint: response.Endpoint, EnableSSL: response.EnableSSL, SSLID: response.SSLID, WebsiteStatus: response.WebsiteStatus, RequestID: pending.RequestID}, nil
}

func (s *MCPRuntimeService) saveWebsiteOperation(ctx context.Context, row *model.MCPSetting, operation mcpWebsiteOperation, websiteID uint, bindingID string) error {
	data, err := json.Marshal(operation)
	if err != nil {
		return err
	}
	err = global.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		store := repo.MCPRepo{DB: tx}
		if err := store.CreateSettingIfAbsent(&model.MCPSetting{ID: 1}); err != nil {
			return err
		}
		rowsAffected, err := store.UpdateWebsiteSetting(model.MCPSetting{
			WebsiteID: websiteID, WebsiteBindingID: bindingID, WebsiteOperation: string(data), Revision: row.Revision + 1,
		}, row.Revision)
		if err != nil {
			return err
		}
		if rowsAffected != 1 {
			return ErrMCPSettingsRevision
		}
		return nil
	})
	if err == nil {
		row.ID, row.WebsiteID, row.WebsiteBindingID, row.WebsiteOperation, row.Revision = 1, websiteID, bindingID, string(data), row.Revision+1
	}
	return err
}

func (s *MCPRuntimeService) SyncWebsite(ctx context.Context) error {
	mcpRuntimeMutation.Lock()
	defer mcpRuntimeMutation.Unlock()
	row, err := s.websiteSetting(ctx)
	if err != nil {
		return err
	}
	operation, err := decodeMCPWebsiteOperation(row.WebsiteOperation)
	if err != nil {
		return err
	}
	if row.WebsiteID == 0 || operation != nil && operation.State == "pending" {
		return nil
	}
	request, err := s.websiteAgentRequest(ctx, row, "sync", "")
	if err != nil {
		return err
	}
	response, err := s.websiteAgent().Call(ctx, request)
	if err != nil || response.Error != "" || response.State != "bound" || response.WebsiteID != row.WebsiteID || response.BindingID != row.WebsiteBindingID {
		return contract.Fail("MCP_WEBSITE_UNAVAILABLE", 503, "MCP website synchronization failed")
	}
	return nil
}
