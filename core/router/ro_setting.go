package router

import (
	v2 "github.com/1Panel-dev/1Panel/core/app/api/v2"
	"github.com/1Panel-dev/1Panel/core/middleware"
	"github.com/gin-gonic/gin"
)

type SettingRouter struct{}

func (s *SettingRouter) InitRouter(Router *gin.RouterGroup) {
	baseApi := v2.ApiGroupApp.BaseApi

	authRouter := Router.Group("settings").
		Use(middleware.SessionAuth())
	{
		authRouter.POST("/search/base", baseApi.GetSettingBaseInfo)
	}

	settingRouter := Router.Group("settings").
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	{
		settingRouter.POST("/search", baseApi.GetSettingInfo)
		settingRouter.POST("/terminal/search", baseApi.GetTerminalSettingInfo)
		settingRouter.GET("/search/available", baseApi.GetSystemAvailable)
		settingRouter.POST("/update", baseApi.UpdateSetting)
		settingRouter.POST("/terminal/update", baseApi.UpdateTerminalSetting)
		settingRouter.GET("/interface", baseApi.LoadInterfaceAddr)
		settingRouter.POST("/menu/update", baseApi.UpdateMenu)
		settingRouter.POST("/menu/default", baseApi.DefaultMenu)
		settingRouter.POST("/proxy/update", baseApi.UpdateProxy)
		settingRouter.POST("/bind/update", baseApi.UpdateBindInfo)
		settingRouter.POST("/port/update", baseApi.UpdatePort)
		settingRouter.POST("/ssl/update", baseApi.UpdateSSL)
		settingRouter.GET("/ssl/info", baseApi.LoadFromCert)
		settingRouter.POST("/ssl/download", baseApi.DownloadSSL)
		settingRouter.POST("/upgrade", baseApi.Upgrade)
		settingRouter.POST("/upgrade/notes", baseApi.GetNotesByVersion)
		settingRouter.GET("/upgrade/releases", baseApi.LoadRelease)
		settingRouter.GET("/upgrade", baseApi.GetUpgradeInfo)
		settingRouter.POST("/apps/store/update", baseApi.UpdateAppstoreConfig)
		settingRouter.GET("/apps/store/config", baseApi.GetAppstoreConfig)
		settingRouter.GET("/memo", baseApi.GetMemo)
		settingRouter.POST("/memo", baseApi.UpdateMemo)
	}

	internalRouter := Router.Group("settings")
	{
		internalRouter.POST("/ssl/reload", baseApi.ReloadSSL)
	}
}
