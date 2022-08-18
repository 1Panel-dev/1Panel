package router

import (
	v1 "github.com/1Panel-dev/1Panel/app/api/v1"
	"github.com/1Panel-dev/1Panel/middleware"

	"github.com/gin-gonic/gin"
)

type HostRouter struct{}

func (s *HostRouter) InitHostRouter(Router *gin.RouterGroup) {
	userRouter := Router.Group("hosts")
	userRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth())
	withRecordRouter := userRouter.Use(middleware.OperationRecord())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		withRecordRouter.POST("", baseApi.Create)
		withRecordRouter.POST("/del", baseApi.DeleteHost)
		userRouter.POST("/search", baseApi.PageHosts)
		userRouter.PUT(":id", baseApi.UpdateHost)
	}
}
