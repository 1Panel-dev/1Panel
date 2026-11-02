package middleware

import (
	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	mcpmeta "github.com/1Panel-dev/1Panel/core/utils/mcpapi/metadata"
	"github.com/gin-gonic/gin"
)

func MCPErrorMetadata() gin.HandlerFunc {
	return helper.WithErrorMetadata(func(c *gin.Context, _ int, messageKey string, _ error) {
		if messageKey == "ErrProxy" && c.GetBool("MCP_API_AUTH") && c.Request != nil && mcpmeta.TrustedRequest(c.Request) {
			c.Header(mcpmeta.BusinessCodeHeader, "ErrProxy")
		}
	})
}
