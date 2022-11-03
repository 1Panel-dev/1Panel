package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type AppRouter struct {
}

func (a *AppRouter) InitAppRouter(Router *gin.RouterGroup) {
	appRouter := Router.Group("apps")
	appRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth())

	baseApi := v1.ApiGroupApp.BaseApi
	{
		appRouter.POST("/sync", baseApi.SyncApp)
		appRouter.POST("/search", baseApi.SearchApp)
		appRouter.GET("/:id", baseApi.GetApp)
		appRouter.GET("/detail/:appId/:version", baseApi.GetAppDetail)
		appRouter.POST("/install", baseApi.InstallApp)
		appRouter.GET("/installed/:appInstallId/versions", baseApi.GetUpdateVersions)
		appRouter.POST("/installed", baseApi.SearchInstalled)
		appRouter.POST("/installed/op", baseApi.OperateInstalled)
		appRouter.POST("/installed/sync", baseApi.SyncInstalled)
		appRouter.POST("/installed/backups", baseApi.SearchInstalledBackup)
		appRouter.POST("/installed/backups/del", baseApi.DeleteAppBackup)
		appRouter.POST("/installed/port/change", baseApi.ChangeAppPort)
		appRouter.GET("/services/:key", baseApi.GetServices)
	}
}
