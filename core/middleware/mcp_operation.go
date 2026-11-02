package middleware

import (
	"github.com/1Panel-dev/1Panel/core/utils/mcpapi"
	"github.com/gin-gonic/gin"
)

func MCPLogPolicy(c *gin.Context) bool {
	return mcpapi.IsRequest(c.Request) && c.GetBool("MCP_API_AUTH")
}
