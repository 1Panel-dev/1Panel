package middleware

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/utils/mcpapi"
	"github.com/gin-gonic/gin"
)

func MCPErrorMetadata() gin.HandlerFunc {
	return helper.WithErrorMetadata(func(c *gin.Context, _ int, _ string, err error) {
		mcpapi.SetBusinessErrorHeader(c.Request, c.Writer.Header(), err)
	})
}
