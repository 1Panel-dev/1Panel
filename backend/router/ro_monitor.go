package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"

	"github.com/gin-gonic/gin"
)

type MonitorRouter struct{}

func (s *MonitorRouter) InitRouter(Router *gin.RouterGroup) {
	monitorRouter := Router.Group("monitors").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		monitorRouter.POST("/search", baseApi.LoadMonitor)
		monitorRouter.GET("/netoptions", baseApi.GetNetworkOptions)
		monitorRouter.GET("/iooptions", baseApi.GetIOOptions)
	}
}
