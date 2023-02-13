package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"

	"github.com/gin-gonic/gin"
)

type BackupRouter struct{}

func (s *BackupRouter) InitBackupRouter(Router *gin.RouterGroup) {
	baRouter := Router.Group("backups").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		baRouter.GET("/search", baseApi.ListBackup)
		baRouter.POST("/search/files", baseApi.LoadFilesFromBackup)
		baRouter.POST("/buckets", baseApi.ListBuckets)
		baRouter.POST("", baseApi.CreateBackup)
		baRouter.POST("/del", baseApi.DeleteBackup)
		baRouter.POST("/update", baseApi.UpdateBackup)
		baRouter.POST("/record/search", baseApi.SearchBackupRecords)
		baRouter.POST("/record/download", baseApi.DownloadRecord)
		baRouter.POST("/record/del", baseApi.DeleteBackupRecord)
	}
}
