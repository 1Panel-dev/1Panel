package router

import (
	v1 "github.com/1Panel-dev/1Panel/app/api/v1"
	"github.com/1Panel-dev/1Panel/middleware"

	"github.com/gin-gonic/gin"
)

type UserRouter struct{}

func (s *UserRouter) InitUserRouter(Router *gin.RouterGroup) {
	userRouter := Router.Group("users")
	userRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth())
	withRecordRouter := userRouter.Use(middleware.OperationRecord())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		withRecordRouter.POST("", baseApi.Register)
		withRecordRouter.POST("/del", baseApi.DeleteUser)
		userRouter.POST("/search", baseApi.PageUsers)
		userRouter.GET(":id", baseApi.GetUserInfo)
		userRouter.POST(":id", baseApi.UpdateUser)
	}
}
