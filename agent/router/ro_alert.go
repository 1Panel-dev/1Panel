package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type AlertRouter struct {
}

func (a *AlertRouter) InitRouter(Router *gin.RouterGroup) {
	alertRouter := Router.Group("alert")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		alertRouter.POST("", baseApi.CreateAlert)
		alertRouter.POST("/update", baseApi.UpdateAlert)
		alertRouter.POST("/search", baseApi.PageAlert)
		alertRouter.POST("/status", baseApi.UpdateAlertStatus)
		alertRouter.POST("/del", baseApi.DeleteAlert)
		alertRouter.GET("/disks/list", baseApi.GetDisks)
		alertRouter.POST("/logs/search", baseApi.PageAlertLogs)
		alertRouter.POST("/logs/clean", baseApi.CleanAlertLogs)
		alertRouter.GET("/clams/list", baseApi.GetClams)
		alertRouter.POST("/cronjob/list", baseApi.GetCronJobs)

		alertRouter.POST("/config/update", baseApi.UpdateAlertConfig)
		alertRouter.POST("/config/info", baseApi.GetAlertConfig)
		alertRouter.POST("/config/del", baseApi.DeleteAlertConfig)
		alertRouter.POST("/config/test", baseApi.TestAlertConfig)
	}
}
