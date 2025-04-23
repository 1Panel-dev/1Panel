package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type BackupRouter struct{}

func (s *BackupRouter) InitRouter(Router *gin.RouterGroup) {
	backupRouter := Router.Group("backups")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		backupRouter.GET("/check/:name", baseApi.CheckBackupUsed)
		backupRouter.GET("/options", baseApi.LoadBackupOptions)
		backupRouter.POST("/search", baseApi.SearchBackup)

		backupRouter.GET("/local", baseApi.GetLocalDir)
		backupRouter.POST("/refresh/token", baseApi.RefreshToken)
		backupRouter.POST("/buckets", baseApi.ListBuckets)
		backupRouter.POST("", baseApi.CreateBackup)
		backupRouter.POST("/del", baseApi.DeleteBackup)
		backupRouter.POST("/update", baseApi.UpdateBackup)

		backupRouter.POST("/backup", baseApi.Backup)
		backupRouter.POST("/recover", baseApi.Recover)
		backupRouter.POST("/recover/byupload", baseApi.RecoverByUpload)
		backupRouter.POST("/search/files", baseApi.LoadFilesFromBackup)
		backupRouter.POST("/record/search", baseApi.SearchBackupRecords)
		backupRouter.POST("/record/size", baseApi.LoadBackupRecordSize)
		backupRouter.POST("/record/search/bycronjob", baseApi.SearchBackupRecordsByCronjob)
		backupRouter.POST("/record/download", baseApi.DownloadRecord)
		backupRouter.POST("/record/del", baseApi.DeleteBackupRecord)
		backupRouter.POST("/record/description/update", baseApi.UpdateRecordDescription)
	}
}
