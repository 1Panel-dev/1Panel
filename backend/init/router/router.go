package router

import (
	"github.com/1Panel-dev/1Panel/docs"
	"github.com/1Panel-dev/1Panel/i18n"
	"github.com/1Panel-dev/1Panel/middleware"
	rou "github.com/1Panel-dev/1Panel/router"
	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"html/template"
)

func Routers() *gin.Engine {
	Router := gin.Default()

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
	{
		systemRouter.InitBaseRouter(PublicGroup) // 注册基础功能路由 不做鉴权
	}
	PrivateGroup := Router.Group("")
	{
		systemRouter.InitUserRouter(PrivateGroup) // 注册用户路由
	}

	return Router
}
