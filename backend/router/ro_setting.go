package router

import (
	v1 "github.com/1Panel-dev/1Panel/app/api/v1"
	"github.com/1Panel-dev/1Panel/middleware"
	"github.com/gin-gonic/gin"
)

type SettingRouter struct{}

func (s *SettingRouter) InitSettingRouter(Router *gin.RouterGroup) {
	baseRouter := Router.Group("settings")
	settingRouter := Router.Group("settings").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	withRecordRouter := Router.Group("settings").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired()).
		Use(middleware.OperationRecord())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		baseRouter.POST("/search", baseApi.GetSettingInfo)
		baseRouter.PUT("/expired/handle", baseApi.HandlePasswordExpired)
		withRecordRouter.PUT("", baseApi.UpdateSetting)
		settingRouter.PUT("/password", baseApi.UpdatePassword)
		settingRouter.POST("/time/sync", baseApi.SyncTime)
		settingRouter.POST("/monitor/clean", baseApi.CleanMonitor)
		settingRouter.GET("/mfa", baseApi.GetMFA)
		settingRouter.POST("/mfa/bind", baseApi.MFABind)
	}
}
