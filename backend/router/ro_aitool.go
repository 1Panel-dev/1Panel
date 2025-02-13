package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type AIToolsRouter struct {
}

func (a *AIToolsRouter) InitRouter(Router *gin.RouterGroup) {
	aiToolsRouter := Router.Group("aitools")
	aiToolsRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.PasswordExpired())

	baseApi := v1.ApiGroupApp.BaseApi
	{
		aiToolsRouter.POST("/ollama/model", baseApi.CreateOllamaModel)
		aiToolsRouter.POST("/ollama/model/search", baseApi.SearchOllamaModel)
		aiToolsRouter.POST("/ollama/model/load", baseApi.LoadOllamaModelDetail)
		aiToolsRouter.POST("/ollama/model/del", baseApi.DeleteOllamaModel)
		aiToolsRouter.GET("/gpu/load", baseApi.LoadGpuInfo)
	}
}
