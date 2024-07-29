package router

import (
	"github.com/1Panel-dev/1Panel/agent/i18n"
	rou "github.com/1Panel-dev/1Panel/agent/router"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

var (
	Router *gin.Engine
)

func Routers() *gin.Engine {
	Router = gin.Default()
	Router.Use(i18n.UseI18n())

	PublicGroup := Router.Group("")
	{
		PublicGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, "ok")
		})
		PublicGroup.Use(gzip.Gzip(gzip.DefaultCompression))
		PublicGroup.Static("/api/v1/images", "./uploads")
	}
	PrivateGroup := Router.Group("/api/v2")
	for _, router := range rou.RouterGroupApp {
		router.InitRouter(PrivateGroup)
	}

	return Router
}
