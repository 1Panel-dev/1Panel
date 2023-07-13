package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type SettingRouter struct{}

func (s *SettingRouter) InitSettingRouter(Router *gin.RouterGroup) {
	router := Router.Group("settings").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth())
	settingRouter := Router.Group("settings").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		router.POST("/search", baseApi.GetSettingInfo)
		router.POST("/expired/handle", baseApi.HandlePasswordExpired)
		settingRouter.GET("/search/available", baseApi.GetSystemAvailable)
		settingRouter.POST("/update", baseApi.UpdateSetting)
		settingRouter.POST("/port/update", baseApi.UpdatePort)
		settingRouter.POST("/ssl/update", baseApi.UpdateSSL)
		settingRouter.GET("/ssl/info", baseApi.LoadFromCert)
		settingRouter.POST("/password/update", baseApi.UpdatePassword)
		settingRouter.GET("/time/option", baseApi.LoadTimeZone)
		settingRouter.POST("/time/sync", baseApi.SyncTime)
		settingRouter.POST("/monitor/clean", baseApi.CleanMonitor)
		settingRouter.GET("/mfa/:interval", baseApi.GetMFA)
		settingRouter.POST("/mfa/bind", baseApi.MFABind)

		settingRouter.POST("/snapshot", baseApi.CreateSnapshot)
		settingRouter.POST("/snapshot/search", baseApi.SearchSnapshot)
		settingRouter.POST("/snapshot/import", baseApi.ImportSnapshot)
		settingRouter.POST("/snapshot/del", baseApi.DeleteSnapshot)
		settingRouter.POST("/snapshot/recover", baseApi.RecoverSnapshot)
		settingRouter.POST("/snapshot/rollback", baseApi.RollbackSnapshot)
		settingRouter.POST("/snapshot/description/update", baseApi.UpdateSnapDescription)

		settingRouter.GET("/backup/search", baseApi.ListBackup)
		settingRouter.GET("/backup/onedrive", baseApi.LoadOneDriveInfo)
		settingRouter.POST("/backup/backup", baseApi.Backup)
		settingRouter.POST("/backup/recover", baseApi.Recover)
		settingRouter.POST("/backup/recover/byupload", baseApi.RecoverByUpload)
		settingRouter.POST("/backup/search/files", baseApi.LoadFilesFromBackup)
		settingRouter.POST("/backup/buckets", baseApi.ListBuckets)
		settingRouter.POST("/backup", baseApi.CreateBackup)
		settingRouter.POST("/backup/del", baseApi.DeleteBackup)
		settingRouter.POST("/backup/update", baseApi.UpdateBackup)
		settingRouter.POST("/backup/record/search", baseApi.SearchBackupRecords)
		settingRouter.POST("/backup/record/download", baseApi.DownloadRecord)
		settingRouter.POST("/backup/record/del", baseApi.DeleteBackupRecord)

		settingRouter.POST("/upgrade", baseApi.Upgrade)
		settingRouter.POST("/upgrade/notes", baseApi.GetNotesByVersion)
		settingRouter.GET("/upgrade", baseApi.GetUpgradeInfo)
		settingRouter.GET("/basedir", baseApi.LoadBaseDir)
	}
}
