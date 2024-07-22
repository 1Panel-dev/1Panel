package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"

	"github.com/gin-gonic/gin"
)

type DashboardRouter struct{}

func (s *DashboardRouter) InitRouter(Router *gin.RouterGroup) {
	cmdRouter := Router.Group("dashboard")
	baseApi := v1.ApiGroupApp.BaseApi
	{
		cmdRouter.GET("/base/os", baseApi.LoadDashboardOsInfo)
		cmdRouter.GET("/base/:ioOption/:netOption", baseApi.LoadDashboardBaseInfo)
		cmdRouter.GET("/current/:ioOption/:netOption", baseApi.LoadDashboardCurrentInfo)
		cmdRouter.POST("/system/restart/:operation", baseApi.SystemRestart)
	}
}
