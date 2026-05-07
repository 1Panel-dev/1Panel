package router

import (
	v2 "github.com/1Panel-dev/1Panel/core/app/api/v2"
	"github.com/1Panel-dev/1Panel/core/middleware"
	"github.com/gin-gonic/gin"
)

type BaseRouter struct{}

func (s *BaseRouter) InitRouter(Router *gin.RouterGroup) {
	baseRouter := Router.Group("auth")
	authRouter := Router.Group("auth").
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v2.ApiGroupApp.BaseApi
	{
		baseRouter.GET("/captcha", baseApi.Captcha)
		baseRouter.POST("/passkey/begin", baseApi.PasskeyBeginLogin)
		baseRouter.POST("/passkey/finish", baseApi.PasskeyFinishLogin)
		baseRouter.POST("/mfalogin", baseApi.MFALogin)
		baseRouter.POST("/login", baseApi.Login)
		baseRouter.POST("/logout", baseApi.LogOut)
		baseRouter.GET("/setting", baseApi.GetLoginSetting)
		baseRouter.GET("/welcome", baseApi.GetWelcomePage)

		authRouter.POST("/mfa", baseApi.LoadMFA)
		authRouter.POST("/mfa/bind", baseApi.MFABind)
		authRouter.POST("/mfa/close", baseApi.MFAClose)

		authRouter.POST("/passkey/register/begin", baseApi.PasskeyRegisterBegin)
		authRouter.POST("/passkey/register/finish", baseApi.PasskeyRegisterFinish)
		authRouter.GET("/passkey/list", baseApi.PasskeyList)
		authRouter.POST("/passkey/del", baseApi.PasskeyDelete)

		authRouter.POST("/api/generate", baseApi.GenerateApiKey)
		authRouter.POST("/api/update", baseApi.UpdateApiConfig)

		authRouter.GET("/current", baseApi.GetCurrentUser)
		authRouter.POST("/current/update", baseApi.UpdateCurrentUser)
		authRouter.POST("/expired/reset", baseApi.ResetPassword)
	}
}
