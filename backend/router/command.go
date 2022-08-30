package router

import (
	v1 "github.com/1Panel-dev/1Panel/app/api/v1"
	"github.com/1Panel-dev/1Panel/middleware"

	"github.com/gin-gonic/gin"
)

type CommandRouter struct{}

func (s *CommandRouter) InitCommandRouter(Router *gin.RouterGroup) {
	userRouter := Router.Group("commands")
	userRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth())
	withRecordRouter := userRouter.Use(middleware.OperationRecord())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		withRecordRouter.POST("", baseApi.CreateCommand)
		withRecordRouter.POST("/del", baseApi.DeleteCommand)
		userRouter.POST("/search", baseApi.SearchCommand)
		userRouter.GET("", baseApi.ListCommand)
		userRouter.PUT(":id", baseApi.UpdateCommand)
	}
}
