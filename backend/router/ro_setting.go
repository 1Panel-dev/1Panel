package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type SettingRouter struct{}

func (s *SettingRouter) InitRouter(Router *gin.RouterGroup) {
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
		settingRouter.GET("/interface", baseApi.LoadInterfaceAddr)
		settingRouter.POST("/menu/update", baseApi.UpdateMenu)
		settingRouter.POST("/proxy/update", baseApi.UpdateProxy)
		settingRouter.POST("/bind/update", baseApi.UpdateBindInfo)
		settingRouter.POST("/port/update", baseApi.UpdatePort)
		settingRouter.POST("/ssl/update", baseApi.UpdateSSL)
		settingRouter.GET("/ssl/info", baseApi.LoadFromCert)
		settingRouter.POST("/ssl/download", baseApi.DownloadSSL)
		settingRouter.POST("/password/update", baseApi.UpdatePassword)
		settingRouter.POST("/mfa", baseApi.LoadMFA)
		settingRouter.POST("/mfa/bind", baseApi.MFABind)

		settingRouter.POST("/snapshot", baseApi.CreateSnapshot)
		settingRouter.POST("/snapshot/status", baseApi.LoadSnapShotStatus)
		settingRouter.POST("/snapshot/search", baseApi.SearchSnapshot)
		settingRouter.POST("/snapshot/size", baseApi.LoadSnapshotSize)
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
		settingRouter.POST("/backup/record/size", baseApi.LoadBackupSize)
		settingRouter.POST("/backup/record/search/bycronjob", baseApi.SearchBackupRecordsByCronjob)
		settingRouter.POST("/backup/record/size/bycronjob", baseApi.LoadBackupSizeByCronjob)
		settingRouter.POST("/backup/record/download", baseApi.DownloadRecord)
		settingRouter.POST("/backup/record/del", baseApi.DeleteBackupRecord)

		settingRouter.POST("/upgrade", baseApi.Upgrade)
		settingRouter.POST("/upgrade/notes", baseApi.GetNotesByVersion)
		settingRouter.GET("/upgrade", baseApi.GetUpgradeInfo)
		settingRouter.GET("/basedir", baseApi.LoadBaseDir)
		settingRouter.POST("/api/config/generate/key", baseApi.GenerateApiKey)
		settingRouter.POST("/api/config/update", baseApi.UpdateApiConfig)
	}
}
