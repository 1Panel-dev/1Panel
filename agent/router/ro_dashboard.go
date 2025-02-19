package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type DashboardRouter struct{}

func (s *DashboardRouter) InitRouter(Router *gin.RouterGroup) {
	cmdRouter := Router.Group("dashboard")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		cmdRouter.GET("/base/os", baseApi.LoadDashboardOsInfo)
		cmdRouter.GET("/app/launcher", baseApi.LoadAppLauncher)
		cmdRouter.POST("/app/launcher/sync", baseApi.SyncAppLauncher)
		cmdRouter.POST("/app/launcher/option", baseApi.LoadAppLauncherOption)
		cmdRouter.GET("/base/:ioOption/:netOption", baseApi.LoadDashboardBaseInfo)
		cmdRouter.GET("/current/node", baseApi.LoadCurrentInfoForNode)
		cmdRouter.GET("/current/:ioOption/:netOption", baseApi.LoadDashboardCurrentInfo)
		cmdRouter.POST("/system/restart/:operation", baseApi.SystemRestart)
	}
}
