package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type SettingRouter struct{}

func (s *SettingRouter) InitRouter(Router *gin.RouterGroup) {
	settingRouter := Router.Group("settings")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		settingRouter.POST("/search", baseApi.GetSettingInfo)
		settingRouter.GET("/search/available", baseApi.GetSystemAvailable)
		settingRouter.POST("/update", baseApi.UpdateSetting)
		settingRouter.GET("/get/:key", baseApi.GetSettingByKey)

		settingRouter.GET("/snapshot/load", baseApi.LoadSnapshotData)
		settingRouter.POST("/snapshot", baseApi.CreateSnapshot)
		settingRouter.POST("/snapshot/recreate", baseApi.RecreateSnapshot)
		settingRouter.POST("/snapshot/search", baseApi.SearchSnapshot)
		settingRouter.POST("/snapshot/import", baseApi.ImportSnapshot)
		settingRouter.POST("/snapshot/del", baseApi.DeleteSnapshot)
		settingRouter.POST("/snapshot/recover", baseApi.RecoverSnapshot)
		settingRouter.POST("/snapshot/rollback", baseApi.RollbackSnapshot)
		settingRouter.POST("/snapshot/description/update", baseApi.UpdateSnapDescription)

		settingRouter.GET("/basedir", baseApi.LoadBaseDir)

		settingRouter.POST("/ssh/check", baseApi.CheckLocalConn)
		settingRouter.POST("/ssh", baseApi.SaveLocalConnInfo)
		settingRouter.POST("/ssh/check/info", baseApi.CheckLocalConnByInfo)
	}
}
