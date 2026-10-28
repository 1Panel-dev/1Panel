package v2

import (
	"net"
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/taskscope"
	"github.com/gin-gonic/gin"
)

func PanelMCPTaskAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		if hasPanelMCPTaskAccess(c.Request) {
			c.Request = c.Request.WithContext(taskscope.WithGrants(c.Request.Context(), constant.TaskScopePanelMCP))
		}
		c.Next()
	}
}

func hasPanelMCPTaskAccess(req *http.Request) bool {
	if !global.IsMaster {
		return false
	}
	local, ok := req.Context().Value(http.LocalAddrContextKey).(net.Addr)
	values := req.Header.Values("X-Panel-MCP-Task-Access")
	return ok && local.Network() == "unix" && len(values) == 1 && values[0] == "1"
}
