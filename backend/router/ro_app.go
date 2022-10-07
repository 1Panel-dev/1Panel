package router

import (
	v1 "github.com/1Panel-dev/1Panel/app/api/v1"
	"github.com/1Panel-dev/1Panel/middleware"
	"github.com/gin-gonic/gin"
)

type AppRouter struct {
}

func (a *AppRouter) InitAppRouter(Router *gin.RouterGroup) {
	appRouter := Router.Group("apps")
	appRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth())

	baseApi := v1.ApiGroupApp.BaseApi
	{
		appRouter.POST("/sync", baseApi.AppSync)
		appRouter.POST("/search", baseApi.AppSearch)
		appRouter.GET("/:id", baseApi.GetApp)
		appRouter.GET("/detail/:appid/:version", baseApi.GetAppDetail)
		appRouter.POST("/install", baseApi.InstallApp)
		appRouter.POST("/installed", baseApi.PageInstalled)
		appRouter.POST("/installed/op", baseApi.InstallOperate)
		appRouter.POST("/installed/sync", baseApi.InstalledSync)
		appRouter.GET("/services/:key", baseApi.GetServices)
	}
}
