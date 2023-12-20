package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"

	"github.com/gin-gonic/gin"
)

type DashboardRouter struct{}

func (s *DashboardRouter) InitRouter(Router *gin.RouterGroup) {
	cmdRouter := Router.Group("dashboard").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		cmdRouter.GET("/base/:ioOption/:netOption", baseApi.LoadDashboardBaseInfo)
		cmdRouter.GET("/current/:ioOption/:netOption", baseApi.LoadDashboardCurrentInfo)
		cmdRouter.POST("/system/restart/:operation", baseApi.SystemRestart)
	}
}
