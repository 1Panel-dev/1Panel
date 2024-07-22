package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/gin-gonic/gin"
)

type ProcessRouter struct {
}

func (f *ProcessRouter) InitRouter(Router *gin.RouterGroup) {
	processRouter := Router.Group("process")
	baseApi := v1.ApiGroupApp.BaseApi
	{
		processRouter.GET("/ws", baseApi.ProcessWs)
		processRouter.POST("/stop", baseApi.StopProcess)
	}
}
