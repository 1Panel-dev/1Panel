package router

import (
	"github.com/1Panel-dev/1Panel/backend/i18n"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	rou "github.com/1Panel-dev/1Panel/backend/router"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

var (
	Router *gin.Engine
)

func setWebStatic(rootRouter *gin.RouterGroup) {
	rootRouter.Static("/api/v1/images", "./uploads")
	rootRouter.Use(func(c *gin.Context) {
		c.Next()
	})
}

func Routers() *gin.Engine {
	Router = gin.Default()
	Router.Use(i18n.UseI18n())

	PublicGroup := Router.Group("")
	{
		PublicGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, "ok")
		})
		PublicGroup.Use(gzip.Gzip(gzip.DefaultCompression))
		setWebStatic(PublicGroup)
	}
	PrivateGroup := Router.Group("/api/v1")
	PrivateGroup.Use(middleware.GlobalLoading())
	for _, router := range rou.RouterGroupApp {
		router.InitRouter(PrivateGroup)
	}

	return Router
}
