package router

import (
	v1 "github.com/1Panel-dev/1Panel/app/api/v1"
	"github.com/1Panel-dev/1Panel/middleware"

	"github.com/gin-gonic/gin"
)

type GroupRouter struct{}

func (s *GroupRouter) InitGroupRouter(Router *gin.RouterGroup) {
	userRouter := Router.Group("group")
	userRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth())
	withRecordRouter := userRouter.Use(middleware.OperationRecord())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		withRecordRouter.POST("", baseApi.CreateGroup)
		withRecordRouter.POST("/del", baseApi.DeleteGroup)
		userRouter.GET("", baseApi.ListGroup)
		userRouter.PUT(":id", baseApi.UpdateGroup)
	}
}
