package router

import (
	"net/http"

	"github.com/1Panel-dev/1Panel/core/app/service"
	"github.com/gin-gonic/gin"
)

const panelMCPTaskAccessHeader = "X-Panel-MCP-Task-Access"

func panelMCPTaskProxyPolicy() agentProxyPolicy {
	return agentProxyPolicy{
		clear: func(req *http.Request) { req.Header.Del(panelMCPTaskAccessHeader) },
		grant: setPanelMCPTaskAccess,
	}
}

func setPanelMCPTaskAccess(c *gin.Context, local bool) {
	if !local || !isPanelMCPTaskEndpoint(c.Request.Method, c.Request.URL.Path) {
		return
	}
	owner, err := service.NewMCPService().ManagementOwner(c, false)
	if err == nil && owner.IsSuperAdmin {
		c.Request.Header.Set(panelMCPTaskAccessHeader, "1")
	}
}

func isPanelMCPTaskEndpoint(method, reqPath string) bool {
	return method == http.MethodPost && (reqPath == "/api/v2/logs/tasks/search" || reqPath == "/api/v2/logs/tasks/read") ||
		method == http.MethodGet && reqPath == "/api/v2/logs/tasks/executing/count"
}
