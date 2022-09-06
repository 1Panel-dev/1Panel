package router

import (
	v1 "github.com/1Panel-dev/1Panel/app/api/v1"
	"github.com/1Panel-dev/1Panel/middleware"

	"github.com/gin-gonic/gin"
)

type MonitorRouter struct{}

func (s *MonitorRouter) InitMonitorRouter(Router *gin.RouterGroup) {
	monitorRouter := Router.Group("monitors").Use(middleware.JwtAuth()).Use(middleware.SessionAuth())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		monitorRouter.POST("/search", baseApi.LoadMonitor)
		monitorRouter.GET("/netoptions", baseApi.GetNetworkOptions)
	}
}
