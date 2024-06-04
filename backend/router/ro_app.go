package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type AppRouter struct {
}

func (a *AppRouter) InitRouter(Router *gin.RouterGroup) {
	appRouter := Router.Group("apps")
	appRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.PasswordExpired())

	baseApi := v1.ApiGroupApp.BaseApi
	{
		appRouter.POST("/sync", baseApi.SyncApp)
		appRouter.GET("/checkupdate", baseApi.GetAppListUpdate)
		appRouter.POST("/search", baseApi.SearchApp)
		appRouter.GET("/:key", baseApi.GetApp)
		appRouter.GET("/detail/:appId/:version/:type", baseApi.GetAppDetail)
		appRouter.GET("/details/:id", baseApi.GetAppDetailByID)
		appRouter.POST("/install", baseApi.InstallApp)
		appRouter.GET("/tags", baseApi.GetAppTags)
		appRouter.POST("/installed/check", baseApi.CheckAppInstalled)
		appRouter.POST("/installed/loadport", baseApi.LoadPort)
		appRouter.POST("/installed/conninfo", baseApi.LoadConnInfo)
		appRouter.GET("/installed/delete/check/:appInstallId", baseApi.DeleteCheck)
		appRouter.POST("/installed/search", baseApi.SearchAppInstalled)
		appRouter.GET("/installed/list", baseApi.ListAppInstalled)
		appRouter.POST("/installed/op", baseApi.OperateInstalled)
		appRouter.POST("/installed/sync", baseApi.SyncInstalled)
		appRouter.POST("/installed/port/change", baseApi.ChangeAppPort)
		appRouter.GET("/services/:key", baseApi.GetServices)
		appRouter.POST("/installed/conf", baseApi.GetDefaultConfig)
		appRouter.GET("/installed/params/:appInstallId", baseApi.GetParams)
		appRouter.POST("/installed/params/update", baseApi.UpdateInstalled)
		appRouter.POST("/installed/ignore", baseApi.IgnoreUpgrade)
		appRouter.GET("/ignored/detail", baseApi.GetIgnoredApp)
		appRouter.POST("/installed/update/versions", baseApi.GetUpdateVersions)
	}
}
