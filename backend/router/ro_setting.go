package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type SettingRouter struct{}

func (s *SettingRouter) InitSettingRouter(Router *gin.RouterGroup) {
	settingRouter := Router.Group("settings").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		settingRouter.POST("/search", baseApi.GetSettingInfo)
		settingRouter.GET("/search/available", baseApi.GetSystemAvailable)
		settingRouter.POST("/expired/handle", baseApi.HandlePasswordExpired)
		settingRouter.POST("/update", baseApi.UpdateSetting)
		settingRouter.POST("/port/update", baseApi.UpdatePort)
		settingRouter.POST("/password/update", baseApi.UpdatePassword)
		settingRouter.POST("/time/sync", baseApi.SyncTime)
		settingRouter.POST("/monitor/clean", baseApi.CleanMonitor)
		settingRouter.GET("/mfa", baseApi.GetMFA)
		settingRouter.POST("/mfa/bind", baseApi.MFABind)
		settingRouter.POST("/snapshot", baseApi.CreateSnapshot)
		settingRouter.POST("/snapshot/search", baseApi.SearchSnapshot)
		settingRouter.POST("/snapshot/del", baseApi.DeleteSnapshot)
		settingRouter.POST("/snapshot/recover", baseApi.RecoverSnapshot)
		settingRouter.POST("/snapshot/rollback", baseApi.RollbackSnapshot)
		settingRouter.POST("/upgrade", baseApi.Upgrade)
		settingRouter.GET("/upgrade", baseApi.GetUpgradeInfo)
		settingRouter.GET("/basedir", baseApi.LoadBaseDir)
	}
}
