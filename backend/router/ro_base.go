package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/gin-gonic/gin"
)

type BaseRouter struct{}

func (s *BaseRouter) InitRouter(Router *gin.RouterGroup) {
	baseRouter := Router.Group("auth")
	baseApi := v1.ApiGroupApp.BaseApi
	{
		baseRouter.GET("/captcha", baseApi.Captcha)
		baseRouter.POST("/mfalogin", baseApi.MFALogin)
		baseRouter.POST("/login", baseApi.Login)
		baseRouter.POST("/logout", baseApi.LogOut)
		baseRouter.GET("/demo", baseApi.CheckIsDemo)
		baseRouter.GET("/language", baseApi.GetLanguage)
		baseRouter.GET("/intl", baseApi.CheckIsIntl)
	}
}
