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

		settingRouter.POST("/snapshot", baseApi.CreateSnapshot)
		settingRouter.POST("/snapshot/status", baseApi.LoadSnapShotStatus)
		settingRouter.POST("/snapshot/search", baseApi.SearchSnapshot)
		settingRouter.POST("/snapshot/import", baseApi.ImportSnapshot)
		settingRouter.POST("/snapshot/del", baseApi.DeleteSnapshot)
		settingRouter.POST("/snapshot/recover", baseApi.RecoverSnapshot)
		settingRouter.POST("/snapshot/rollback", baseApi.RollbackSnapshot)
		settingRouter.POST("/snapshot/description/update", baseApi.UpdateSnapDescription)

		settingRouter.GET("/backup/search", baseApi.ListBackup)
		settingRouter.GET("/backup/onedrive", baseApi.LoadOneDriveInfo)
		settingRouter.POST("/backup/backup", baseApi.Backup)
		settingRouter.POST("/backup/refresh/onedrive", baseApi.RefreshOneDriveToken)
		settingRouter.POST("/backup/recover", baseApi.Recover)
		settingRouter.POST("/backup/recover/byupload", baseApi.RecoverByUpload)
		settingRouter.POST("/backup/search/files", baseApi.LoadFilesFromBackup)
		settingRouter.POST("/backup/buckets", baseApi.ListBuckets)
		settingRouter.POST("/backup", baseApi.CreateBackup)
		settingRouter.POST("/backup/del", baseApi.DeleteBackup)
		settingRouter.POST("/backup/update", baseApi.UpdateBackup)
		settingRouter.POST("/backup/record/search", baseApi.SearchBackupRecords)
		settingRouter.POST("/backup/record/search/bycronjob", baseApi.SearchBackupRecordsByCronjob)
		settingRouter.POST("/backup/record/download", baseApi.DownloadRecord)
		settingRouter.POST("/backup/record/del", baseApi.DeleteBackupRecord)

		settingRouter.GET("/basedir", baseApi.LoadBaseDir)
	}
}
