package v2

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/auth"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/service"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
	"github.com/gin-gonic/gin"
)

func mcpRead(c *gin.Context, value any) bool {
	data, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, 64<<10))
	if err == nil {
		err = contract.DecodeObject(data, value)
	}
	if err == nil {
		switch value.(type) {
		case *dto.MCPCredentialCreate, *dto.MCPCredentialUpdate:
			var fields struct {
				ExpiresAt json.RawMessage `json:"expiresAt"`
			}
			if json.Unmarshal(data, &fields) != nil || len(fields.ExpiresAt) == 0 {
				err = contract.Invalid()
			}
		}
	}
	if err != nil {
		mcpManagementError(c, contract.Invalid())
		return false
	}
	return true
}
func mcpManagementError(c *gin.Context, err error) {
	code := http.StatusInternalServerError
	errorCode := "MCP_OPERATION_FAILED"
	invocationID := ""
	var known *contract.Error
	if errors.As(err, &known) {
		code = known.Status
		errorCode = known.Code
		invocationID = known.InvocationID
	} else if errors.Is(err, service.ErrMCPSettingsRevision) {
		code = 409
		errorCode = "MCP_CONFLICT"
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(200, struct {
		dto.Response
		ErrorCode    string `json:"errorCode"`
		InvocationID string `json:"invocationId,omitempty"`
	}{Response: dto.Response{Code: code, Message: mcpManagementMessage(errorCode)}, ErrorCode: errorCode, InvocationID: invocationID})
	c.Abort()
}
func mcpManagement(c *gin.Context, write bool, admin bool) (*service.MCPService, auth.APIKeyOwner, bool) {
	s := service.NewMCPService()
	owner, err := s.ManagementOwner(c, write)
	if err != nil {
		mcpManagementError(c, err)
		return nil, owner, false
	}
	if admin && !owner.IsSuperAdmin {
		mcpManagementError(c, contract.Forbidden())
		return nil, owner, false
	}
	c.Header("Cache-Control", "no-store")
	return s, owner, true
}
func (b *BaseApi) MCPCatalog(c *gin.Context) {
	s, owner, ok := mcpManagement(c, false, false)
	if ok {
		catalog, err := s.Catalog(owner)
		if err != nil {
			mcpManagementError(c, err)
			return
		}
		helper.SuccessWithData(c, catalog)
	}
}
func (b *BaseApi) MCPNodes(c *gin.Context) {
	s, owner, ok := mcpManagement(c, false, false)
	if !ok {
		return
	}
	if c.Request.URL.RawQuery != "" {
		mcpManagementError(c, contract.Invalid())
		return
	}
	nodes, err := s.Nodes(owner)
	if err != nil {
		mcpManagementError(c, err)
		return
	}
	helper.SuccessWithData(c, nodes)
}
func (b *BaseApi) SearchMCPCredentials(c *gin.Context) {
	s, owner, ok := mcpManagement(c, false, false)
	if !ok {
		return
	}
	var req dto.MCPCredentialSearch
	if !mcpRead(c, &req) {
		return
	}
	result, err := s.Search(owner, req)
	if err != nil {
		mcpManagementError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}
func (b *BaseApi) CreateMCPCredential(c *gin.Context) {
	s, owner, ok := mcpManagement(c, true, false)
	if !ok {
		return
	}
	var req dto.MCPCredentialCreate
	if !mcpRead(c, &req) {
		return
	}
	result, err := s.Create(owner, req)
	if err != nil {
		mcpManagementError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}
func (b *BaseApi) UpdateMCPCredential(c *gin.Context) {
	s, owner, ok := mcpManagement(c, true, false)
	if !ok {
		return
	}
	var req dto.MCPCredentialUpdate
	if !mcpRead(c, &req) {
		return
	}
	if err := s.Update(owner, req); err != nil {
		mcpManagementError(c, err)
		return
	}
	helper.Success(c)
}
func (b *BaseApi) StatusMCPCredential(c *gin.Context) {
	s, owner, ok := mcpManagement(c, true, false)
	if !ok {
		return
	}
	var req dto.MCPCredentialStatus
	if !mcpRead(c, &req) {
		return
	}
	if err := s.Status(owner, req); err != nil {
		mcpManagementError(c, err)
		return
	}
	helper.Success(c)
}
func (b *BaseApi) RevokeMCPCredential(c *gin.Context) {
	s, owner, ok := mcpManagement(c, true, false)
	if !ok {
		return
	}
	var req dto.MCPCredentialMutation
	if !mcpRead(c, &req) {
		return
	}
	if err := s.Revoke(owner, req); err != nil {
		mcpManagementError(c, err)
		return
	}
	helper.Success(c)
}
func (b *BaseApi) SearchMCPInvocations(c *gin.Context) {
	s, owner, ok := mcpManagement(c, false, false)
	if !ok {
		return
	}
	var req dto.MCPInvocationSearch
	if !mcpRead(c, &req) {
		return
	}
	result, err := s.Invocations(owner, req)
	if err != nil {
		mcpManagementError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}
func (b *BaseApi) MCPServiceStatus(c *gin.Context) {
	_, owner, ok := mcpManagement(c, false, false)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	result, err := service.NewMCPRuntimeService().Status(ctx)
	if err != nil {
		mcpManagementError(c, err)
		return
	}
	result.Endpoint = mcpPublicEndpoint(c, result.Endpoint)
	result.CanManageService = owner.IsSuperAdmin
	helper.SuccessWithData(c, result)
}
func (b *BaseApi) MCPSettings(c *gin.Context) {
	_, _, ok := mcpManagement(c, false, false)
	if !ok {
		return
	}
	result, err := service.NewMCPRuntimeService().Settings()
	if err != nil {
		mcpManagementError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}
func (b *BaseApi) UpdateMCPSettings(c *gin.Context) {
	_, _, ok := mcpManagement(c, true, true)
	if !ok {
		return
	}
	var req dto.MCPRuntimeSettingsUpdate
	if !mcpRead(c, &req) {
		return
	}
	result, err := service.NewMCPRuntimeService().Update(req)
	if err != nil {
		mcpServiceError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}
func (b *BaseApi) OperateMCPService(c *gin.Context) {
	_, _, ok := mcpManagement(c, true, true)
	if !ok {
		return
	}
	var req struct {
		Operation string `json:"operation"`
	}
	if !mcpRead(c, &req) {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 55*time.Second)
	defer cancel()
	result, err := service.NewMCPRuntimeService().Operate(ctx, req.Operation)
	if err != nil {
		mcpServiceError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}
func (b *BaseApi) InstallMCPService(c *gin.Context) {
	_, _, ok := mcpManagement(c, true, true)
	if !ok {
		return
	}
	var req dto.MCPRuntimeInstall
	if !mcpRead(c, &req) {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 6*time.Minute)
	defer cancel()
	result, err := service.NewMCPRuntimeService().Install(ctx, req)
	if err != nil {
		mcpServiceError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}
func (b *BaseApi) MCPServiceLogs(c *gin.Context) {
	_, _, ok := mcpManagement(c, false, true)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	result, err := service.NewMCPRuntimeService().Logs(ctx)
	if err != nil {
		mcpServiceError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}
func mcpServiceError(c *gin.Context, err error) {
	var known *contract.Error
	if errors.Is(err, service.ErrMCPSettingsRevision) || errors.As(err, &known) {
		mcpManagementError(c, err)
		return
	}
	helper.InternalServer(c, err)
}

func MCPRuntimeTools(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	token, err := service.MCPBearer(c)
	if err != nil {
		MCPRuntimeError(c, err)
		return
	}
	if !mcpRuntimeProof(c, token) {
		return
	}
	result, err := service.NewMCPService().RuntimeTools(c, token)
	if err != nil {
		MCPRuntimeError(c, err)
		return
	}
	c.JSON(200, result)
}
func MCPRuntimeInvoke(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	token, err := service.MCPBearer(c)
	if err != nil {
		MCPRuntimeError(c, err)
		return
	}
	if !mcpRuntimeProof(c, token) {
		return
	}
	data, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, 64<<10))
	var req contract.Invoke
	if err == nil {
		err = contract.DecodeObject(data, &req)
	}
	if err == nil {
		req.NodeID, err = contract.NormalizeNodeID(req.NodeID)
	}
	if err != nil || req.AdmissionID == "" {
		MCPRuntimeError(c, contract.Invalid())
		return
	}
	result, err := service.NewMCPService().Invoke(c, token, req)
	if err != nil {
		MCPRuntimeError(c, err)
		return
	}
	c.JSON(200, result)
}
func MCPRuntimeError(c *gin.Context, err error) {
	var known *contract.Error
	if !errors.As(err, &known) {
		known = contract.Fail("MCP_EXECUTION_FAILED", 503, "MCP service is unavailable")
	}
	c.Header("Cache-Control", "no-store")
	if known.RetryAfterSeconds > 0 {
		c.Header("Retry-After", strconv.Itoa(known.RetryAfterSeconds))
	}
	c.JSON(known.Status, known)
	c.Abort()
}

func mcpPublicEndpoint(c *gin.Context, configured string) string {
	if configured != "" {
		return configured
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if c.Request.Host == "" {
		return ""
	}
	return (&url.URL{Scheme: scheme, Host: c.Request.Host, Path: "/mcp"}).String()
}
