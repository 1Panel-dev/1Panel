package router

import (
	v1 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type AIToolsRouter struct {
}

func (a *AIToolsRouter) InitRouter(Router *gin.RouterGroup) {
	aiToolsRouter := Router.Group("ai")

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

		aiToolsRouter.POST("/tensorrt/search", baseApi.PageTensorRTLLMs)
		aiToolsRouter.POST("/tensorrt/create", baseApi.CreateTensorRTLLM)
		aiToolsRouter.POST("/tensorrt/update", baseApi.UpdateTensorRTLLM)
		aiToolsRouter.POST("/tensorrt/delete", baseApi.DeleteTensorRTLLM)
		aiToolsRouter.POST("/tensorrt/operate", baseApi.OperateTensorRTLLM)

		aiToolsRouter.POST("/agents", baseApi.CreateAgent)
		aiToolsRouter.POST("/agents/search", baseApi.PageAgents)
		aiToolsRouter.POST("/agents/delete", baseApi.DeleteAgent)
		aiToolsRouter.POST("/agents/token/reset", baseApi.ResetAgentToken)
		aiToolsRouter.POST("/agents/model/update", baseApi.UpdateAgentModelConfig)
		aiToolsRouter.GET("/agents/providers", baseApi.GetAgentProviders)
		aiToolsRouter.POST("/agents/accounts", baseApi.CreateAgentAccount)
		aiToolsRouter.POST("/agents/accounts/update", baseApi.UpdateAgentAccount)
		aiToolsRouter.POST("/agents/accounts/search", baseApi.PageAgentAccounts)
		aiToolsRouter.POST("/agents/accounts/verify", baseApi.VerifyAgentAccount)
		aiToolsRouter.POST("/agents/accounts/delete", baseApi.DeleteAgentAccount)
		aiToolsRouter.POST("/agents/channel/feishu/get", baseApi.GetAgentFeishuConfig)
		aiToolsRouter.POST("/agents/channel/feishu/update", baseApi.UpdateAgentFeishuConfig)
		aiToolsRouter.POST("/agents/channel/feishu/approve", baseApi.ApproveAgentFeishuPairing)
	}
}
