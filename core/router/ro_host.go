package router

import (
	v2 "github.com/1Panel-dev/1Panel/core/app/api/v2"
	"github.com/1Panel-dev/1Panel/core/middleware"
	"github.com/gin-gonic/gin"
)

type HostRouter struct{}

func (s *HostRouter) InitRouter(Router *gin.RouterGroup) {
	hostRouter := Router.Group("hosts").
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v2.ApiGroupApp.BaseApi
	{
		hostRouter.POST("", baseApi.CreateHost)
		hostRouter.POST("/info", baseApi.GetHostByID)
		hostRouter.POST("/del", baseApi.DeleteHost)
		hostRouter.POST("/update", baseApi.UpdateHost)
		hostRouter.POST("/update/group", baseApi.UpdateHostGroup)
		hostRouter.POST("/search", baseApi.SearchHost)
		hostRouter.POST("/tree", baseApi.HostTree)
		hostRouter.POST("/test/byinfo", baseApi.TestByInfo)
		hostRouter.POST("/test/byid/:id", baseApi.TestByID)

		hostRouter.GET("/terminal", baseApi.WsSsh)
	}
}
