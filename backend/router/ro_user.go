package router

import (
	v1 "1Panel/app/api/v1"

	"github.com/gin-gonic/gin"
)

type UserRouter struct{}

func (s *UserRouter) InitUserRouter(Router *gin.RouterGroup) {
	userRouter := Router.Group("users")
	baseApi := v1.ApiGroupApp.BaseApi
	{
		userRouter.POST("", baseApi.Register)
		userRouter.DELETE("", baseApi.DeleteUser)
		userRouter.GET("", baseApi.GetUserList)
		userRouter.GET(":name", baseApi.GetUserInfo)
	}
}
