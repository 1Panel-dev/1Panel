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
	}
}
