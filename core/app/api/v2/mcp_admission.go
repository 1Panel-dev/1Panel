package v2

import (
	"io"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/service"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
	"github.com/gin-gonic/gin"
	"net/http"
)

func mcpRuntimeProof(c *gin.Context, token string) bool {
	source, err := contract.VerifyAccessProof(c.GetHeader(contract.AccessProofHeader), token, time.Now())
	if err != nil {
		MCPRuntimeError(c, err)
		return false
	}
	c.Set("MCP_CLIENT_IP", source)
	return true
}

func MCPRuntimeAdmit(c *gin.Context) {
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
	var req contract.Admission
	if err != nil || contract.DecodeObject(data, &req) != nil {
		MCPRuntimeError(c, contract.Invalid())
		return
	}
	result, err := service.NewMCPService().Admit(c, token, req)
	if err != nil {
		MCPRuntimeError(c, err)
		return
	}
	c.JSON(200, result)
}

func MCPRuntimeComplete(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	token, err := service.MCPBearer(c)
	if err != nil {
		MCPRuntimeError(c, err)
		return
	}
	data, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, 1024))
	var req contract.Completion
	if err != nil || contract.DecodeObject(data, &req) != nil {
		MCPRuntimeError(c, contract.Invalid())
		return
	}
	result, err := service.NewMCPService().CompleteWithOutcome(token, req)
	if err != nil {
		MCPRuntimeError(c, err)
		return
	}
	if req.ErrorCode != "" {
		c.JSON(http.StatusOK, result)
		return
	}
	c.Status(204)
}
