package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type McpServerRouter struct {
}

func (m *McpServerRouter) InitRouter(Router *gin.RouterGroup) {
	mcpRouter := Router.Group("mcp")
	mcpRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.PasswordExpired())

	baseApi := v1.ApiGroupApp.BaseApi
	{
		mcpRouter.POST("/search", baseApi.PageMcpServers)
		mcpRouter.POST("/server", baseApi.CreateMcpServer)
	}
}
