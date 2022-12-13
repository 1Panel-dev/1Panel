package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"

	"github.com/gin-gonic/gin"
)

type CommandRouter struct{}

func (s *CommandRouter) InitCommandRouter(Router *gin.RouterGroup) {
	cmdRouter := Router.Group("commands").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		cmdRouter.GET("", baseApi.ListCommand)
		cmdRouter.POST("", baseApi.CreateCommand)
		cmdRouter.POST("/del", baseApi.DeleteCommand)
		cmdRouter.POST("/search", baseApi.SearchCommand)
		cmdRouter.POST("/update", baseApi.UpdateCommand)
	}
}
