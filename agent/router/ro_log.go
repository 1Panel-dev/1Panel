package router

import (
	v1 "github.com/1Panel-dev/1Panel/agent/app/api/v1"

	"github.com/gin-gonic/gin"
)

type LogRouter struct{}

func (s *LogRouter) InitRouter(Router *gin.RouterGroup) {
	operationRouter := Router.Group("logs")
	baseApi := v1.ApiGroupApp.BaseApi
	{
		operationRouter.GET("/system/files", baseApi.GetSystemFiles)
		operationRouter.POST("/system", baseApi.GetSystemLogs)
	}
}
