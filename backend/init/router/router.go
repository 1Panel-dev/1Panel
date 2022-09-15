package router

import (
	"html/template"

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
	Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	Router.Use(i18n.GinI18nLocalize())

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
	{
		systemRouter.InitBaseRouter(PrivateGroup)
		systemRouter.InitUserRouter(PrivateGroup)
		systemRouter.InitHostRouter(PrivateGroup)
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
