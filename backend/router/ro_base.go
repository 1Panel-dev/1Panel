package router

import (
	v1 "github.com/1Panel-dev/1Panel/app/api/v1"
	"github.com/1Panel-dev/1Panel/middleware"
	"github.com/gin-gonic/gin"
)

type BaseRouter struct{}

func (s *BaseRouter) InitBaseRouter(Router *gin.RouterGroup) {
	baseRouter := Router.Group("auth")
	withRecordRouter := Router.Group("auth").Use(middleware.OperationRecord())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		baseRouter.GET("captcha", baseApi.Captcha)
		withRecordRouter.POST("login", baseApi.Login)
		withRecordRouter.POST("logout", baseApi.LogOut)
	}
}
