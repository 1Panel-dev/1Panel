package router

import (
	v1 "github.com/1Panel-dev/1Panel/app/api/v1"
	"github.com/1Panel-dev/1Panel/middleware"

	"github.com/gin-gonic/gin"
)

type CommandRouter struct{}

func (s *CommandRouter) InitCommandRouter(Router *gin.RouterGroup) {
	cmdRouter := Router.Group("commands").Use(middleware.JwtAuth()).Use(middleware.SessionAuth())
	withRecordRouter := Router.Group("commands").Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.OperationRecord())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		withRecordRouter.POST("", baseApi.CreateCommand)
		withRecordRouter.POST("/del", baseApi.DeleteCommand)
		withRecordRouter.PUT(":id", baseApi.UpdateCommand)
		cmdRouter.POST("/search", baseApi.SearchCommand)
		cmdRouter.GET("", baseApi.ListCommand)
	}
}
