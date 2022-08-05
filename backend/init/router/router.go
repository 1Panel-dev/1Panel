package router

import (
	rou "github.com/1Panel-dev/1Panel/router"
	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	Router := gin.Default()
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
