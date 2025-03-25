package router

import (
	v2 "github.com/1Panel-dev/1Panel/core/app/api/v2"
	"github.com/1Panel-dev/1Panel/core/middleware"
	"github.com/gin-gonic/gin"
)

type CommandRouter struct{}

func (s *CommandRouter) InitRouter(Router *gin.RouterGroup) {
	commandRouter := Router.Group("commands").
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v2.ApiGroupApp.BaseApi
	{
		commandRouter.POST("/list", baseApi.ListCommand)
		commandRouter.POST("", baseApi.CreateCommand)
		commandRouter.POST("/del", baseApi.DeleteCommand)
		commandRouter.POST("/search", baseApi.SearchCommand)
		commandRouter.POST("/tree", baseApi.SearchCommandTree)
		commandRouter.POST("/update", baseApi.UpdateCommand)
	}
}
