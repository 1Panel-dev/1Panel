package router

import (
	"html/template"

	v1 "github.com/1Panel-dev/1Panel/app/api/v1"
	"github.com/1Panel-dev/1Panel/docs"
	"github.com/1Panel-dev/1Panel/i18n"
	"github.com/1Panel-dev/1Panel/middleware"
	rou "github.com/1Panel-dev/1Panel/router"
	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Routers() *gin.Engine {
	Router := gin.Default()
	Router.Use(middleware.CSRF())
	Router.Use(middleware.LoadCsrfToken())

	docs.SwaggerInfo.BasePath = "/api/v1"

	Router.Use(i18n.GinI18nLocalize())
	Router.GET("/api/v1/info", v1.ApiGroupApp.BaseApi.GetSafetyStatus)
	Router.GET("/api/v1/:code", v1.ApiGroupApp.BaseApi.SafeEntrance)
	Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	Router.SetFuncMap(template.FuncMap{
		"Localize": ginI18n.GetMessage,
	})
	Router.Use(middleware.JwtAuth())

	systemRouter := rou.RouterGroupApp

	PublicGroup := Router.Group("")
	{
		PublicGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, "ok")
		})
	}

	PrivateGroup := Router.Group("/api/v1")
	PrivateGroup.Use(middleware.SafetyAuth())
	{
		systemRouter.InitBaseRouter(PrivateGroup)
		systemRouter.InitHostRouter(PrivateGroup)
		systemRouter.InitBackupRouter(PrivateGroup)
		systemRouter.InitGroupRouter(PrivateGroup)
		systemRouter.InitCommandRouter(PrivateGroup)
		systemRouter.InitTerminalRouter(PrivateGroup)
		systemRouter.InitMonitorRouter(PrivateGroup)
		systemRouter.InitOperationLogRouter(PrivateGroup)
		systemRouter.InitFileRouter(PrivateGroup)
		systemRouter.InitSettingRouter(PrivateGroup)
	}

	return Router
}
