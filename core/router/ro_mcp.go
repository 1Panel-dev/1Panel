package router

import (
	v2 "github.com/1Panel-dev/1Panel/core/app/api/v2"
	"github.com/1Panel-dev/1Panel/core/middleware"
	"github.com/gin-gonic/gin"
)

type PanelMCPRouter struct{}

func (*PanelMCPRouter) InitRouter(group *gin.RouterGroup) {
	b := v2.ApiGroupApp.BaseApi
	r := group.Group("mcp").Use(middleware.SessionAuth())
	r.GET("status", b.MCPServiceStatus)
	r.GET("settings", b.MCPSettings)
	r.POST("settings", b.UpdateMCPSettings)
	r.GET("catalog", b.MCPCatalog)
	r.GET("nodes", b.MCPNodes)
	r.POST("service/operate", b.OperateMCPService)
	r.POST("service/install", b.InstallMCPService)
	r.GET("service/logs", b.MCPServiceLogs)
	r.POST("credentials/search", b.SearchMCPCredentials)
	r.POST("credentials/create", b.CreateMCPCredential)
	r.POST("credentials/update", b.UpdateMCPCredential)
	r.POST("credentials/status", b.StatusMCPCredential)
	r.POST("credentials/revoke", b.RevokeMCPCredential)
	r.POST("credentials/reveal", b.RevealMCPCredential)
	r.POST("credentials/rotate", b.RotateMCPCredential)
	r.POST("credentials/export", b.ExportMCPCredential)
	r.POST("invocations/search", b.SearchMCPInvocations)
}
