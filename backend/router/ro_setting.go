package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type SettingRouter struct{}

func (s *SettingRouter) InitSettingRouter(Router *gin.RouterGroup) {
	baseRouter := Router.Group("settings")
	settingRouter := Router.Group("settings").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		baseRouter.POST("/search", baseApi.GetSettingInfo)
		baseRouter.POST("/expired/handle", baseApi.HandlePasswordExpired)
		baseRouter.POST("/update", baseApi.UpdateSetting)
		settingRouter.POST("/password/update", baseApi.UpdatePassword)
		settingRouter.POST("/time/sync", baseApi.SyncTime)
		settingRouter.POST("/monitor/clean", baseApi.CleanMonitor)
		settingRouter.GET("/mfa", baseApi.GetMFA)
		settingRouter.POST("/mfa/bind", baseApi.MFABind)
		settingRouter.POST("/snapshot", baseApi.CreateSnapshot)
		settingRouter.POST("/snapshot/search", baseApi.SearchSnapshot)
	}
}
