package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"

	"github.com/gin-gonic/gin"
)

type HostRouter struct{}

func (s *HostRouter) InitHostRouter(Router *gin.RouterGroup) {
	hostRouter := Router.Group("hosts").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		hostRouter.POST("", baseApi.CreateHost)
		hostRouter.POST("/del", baseApi.DeleteHost)
		hostRouter.POST("/update", baseApi.UpdateHost)
		hostRouter.POST("/update/group", baseApi.UpdateHostGroup)
		hostRouter.POST("/search", baseApi.SearchHost)
		hostRouter.POST("/tree", baseApi.HostTree)
		hostRouter.POST("/test/byinfo", baseApi.TestByInfo)
		hostRouter.POST("/test/byid/:id", baseApi.TestByID)
		hostRouter.GET(":id", baseApi.GetHostInfo)

		hostRouter.GET("/command", baseApi.ListCommand)
		hostRouter.POST("/command", baseApi.CreateCommand)
		hostRouter.POST("/command/del", baseApi.DeleteCommand)
		hostRouter.POST("/command/search", baseApi.SearchCommand)
		hostRouter.POST("/command/update", baseApi.UpdateCommand)
	}
}
