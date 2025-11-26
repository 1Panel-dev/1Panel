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
		cmdRouter.GET("/quick/option", baseApi.LoadQuickOption)
		cmdRouter.POST("/quick/change", baseApi.UpdateQuickJump)
		cmdRouter.GET("/app/launcher", baseApi.LoadAppLauncher)
		cmdRouter.POST("/app/launcher/show", baseApi.UpdateAppLauncher)
		cmdRouter.POST("/app/launcher/option", baseApi.LoadAppLauncherOption)
		cmdRouter.GET("/base/:ioOption/:netOption", baseApi.LoadDashboardBaseInfo)
		cmdRouter.GET("/current/node", baseApi.LoadCurrentInfoForNode)
		cmdRouter.GET("/current/:ioOption/:netOption", baseApi.LoadDashboardCurrentInfo)
		cmdRouter.GET("/current/top/cpu", baseApi.LoadDashboardTopCPU)
		cmdRouter.GET("/current/top/mem", baseApi.LoadDashboardTopMem)
		cmdRouter.POST("/system/restart/:operation", baseApi.SystemRestart)
	}
}
