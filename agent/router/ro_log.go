package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type LogRouter struct{}

func (s *LogRouter) InitRouter(Router *gin.RouterGroup) {
	operationRouter := Router.Group("logs")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		operationRouter.GET("/system/files", baseApi.GetSystemFiles)
		operationRouter.GET("/system/status", baseApi.GetSystemLogStatus)
		operationRouter.POST("/system/read", baseApi.ReadSystemLog)
		operationRouter.GET("/system/services", baseApi.ListRunningServices)
		operationRouter.POST("/tasks/search", baseApi.PageTasks)
		operationRouter.POST("/tasks/read", baseApi.ReadTaskLogByLine)
		operationRouter.GET("/tasks/executing/count", baseApi.CountExecutingTasks)
	}
}
