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
		backupRouter.GET("/check/:id", baseApi.CheckBackupUsed)
		backupRouter.POST("/backup", baseApi.Backup)
		backupRouter.POST("/recover", baseApi.Recover)
		backupRouter.POST("/recover/byupload", baseApi.RecoverByUpload)
		backupRouter.POST("/search/files", baseApi.LoadFilesFromBackup)
		backupRouter.POST("/record/search", baseApi.SearchBackupRecords)
		backupRouter.POST("/record/search/bycronjob", baseApi.SearchBackupRecordsByCronjob)
		backupRouter.POST("/record/download", baseApi.DownloadRecord)
		backupRouter.POST("/record/del", baseApi.DeleteBackupRecord)
	}
}
