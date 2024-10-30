package router

import (
	v2 "github.com/1Panel-dev/1Panel/core/app/api/v2"
	"github.com/gin-gonic/gin"
)

type AppLauncherRouter struct{}

func (s *AppLauncherRouter) InitRouter(Router *gin.RouterGroup) {
	launcherRouter := Router.Group("launcher")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		launcherRouter.GET("/search", baseApi.SearchAppLauncher)
		launcherRouter.POST("/change/show", baseApi.UpdateAppLauncher)
	}
}
