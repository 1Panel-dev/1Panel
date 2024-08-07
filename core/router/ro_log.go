package router

import (
	v2 "github.com/1Panel-dev/1Panel/core/app/api/v2"

	"github.com/gin-gonic/gin"
)

type LogRouter struct{}

func (s *LogRouter) InitRouter(Router *gin.RouterGroup) {
	operationRouter := Router.Group("logs")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		operationRouter.POST("/login", baseApi.GetLoginLogs)
		operationRouter.POST("/operation", baseApi.GetOperationLogs)
		operationRouter.POST("/clean", baseApi.CleanLogs)
	}
}
