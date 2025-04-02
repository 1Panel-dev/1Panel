package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type AIToolsRouter struct {
}

func (a *AIToolsRouter) InitRouter(Router *gin.RouterGroup) {
	aiToolsRouter := Router.Group("ai")
	aiToolsRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.PasswordExpired())

	baseApi := v1.ApiGroupApp.BaseApi
	{
		aiToolsRouter.POST("/ollama/close", baseApi.CloseOllamaModel)
		aiToolsRouter.POST("/ollama/model", baseApi.CreateOllamaModel)
		aiToolsRouter.POST("/ollama/model/recreate", baseApi.RecreateOllamaModel)
		aiToolsRouter.POST("/ollama/model/search", baseApi.SearchOllamaModel)
		aiToolsRouter.POST("/ollama/model/sync", baseApi.SyncOllamaModel)
		aiToolsRouter.POST("/ollama/model/load", baseApi.LoadOllamaModelDetail)
		aiToolsRouter.POST("/ollama/model/del", baseApi.DeleteOllamaModel)
		aiToolsRouter.GET("/gpu/load", baseApi.LoadGpuInfo)
		aiToolsRouter.POST("/domain/bind", baseApi.BindDomain)
		aiToolsRouter.POST("/domain/get", baseApi.GetBindDomain)
		aiToolsRouter.POST("/domain/update", baseApi.UpdateBindDomain)

		aiToolsRouter.POST("/mcp/search", baseApi.PageMcpServers)
		aiToolsRouter.POST("/mcp/server", baseApi.CreateMcpServer)
		aiToolsRouter.POST("/mcp/server/update", baseApi.UpdateMcpServer)
		aiToolsRouter.POST("/mcp/server/del", baseApi.DeleteMcpServer)
		aiToolsRouter.POST("/mcp/server/op", baseApi.OperateMcpServer)
		aiToolsRouter.POST("/mcp/domain/bind", baseApi.BindMcpDomain)
		aiToolsRouter.GET("/mcp/domain/get", baseApi.GetMcpBindDomain)
		aiToolsRouter.POST("/mcp/domain/update", baseApi.UpdateMcpBindDomain)
	}
}
