package router

import (
	v1 "github.com/1Panel-dev/1Panel/app/api/v1"
	"github.com/1Panel-dev/1Panel/middleware"
	"github.com/gin-gonic/gin"
)

type SettingRouter struct{}

func (s *SettingRouter) InitSettingRouter(Router *gin.RouterGroup) {
	settingRouter := Router.Group("settings").Use(middleware.JwtAuth()).Use(middleware.SessionAuth())
	withRecordRouter := Router.Group("settings").Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.OperationRecord())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		settingRouter.POST("/search", baseApi.GetSettingInfo)
		withRecordRouter.PUT("", baseApi.UpdateSetting)
		settingRouter.PUT("/password", baseApi.UpdatePassword)
		settingRouter.POST("/time/sync", baseApi.SyncTime)
		settingRouter.POST("/monitor/clean", baseApi.CleanMonitor)
		settingRouter.GET("/mfa", baseApi.GetMFA)
		settingRouter.POST("/mfa/bind", baseApi.MFABind)
	}
}
