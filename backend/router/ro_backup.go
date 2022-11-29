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
	withRecordRouter := Router.Group("backups").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired()).
		Use(middleware.OperationRecord())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		baRouter.GET("/search", baseApi.ListBackup)
		baRouter.POST("/buckets", baseApi.ListBuckets)
		withRecordRouter.POST("", baseApi.CreateBackup)
		withRecordRouter.POST("/del", baseApi.DeleteBackup)
		withRecordRouter.POST("/record/search", baseApi.SearchBackupRecords)
		withRecordRouter.POST("/record/download", baseApi.DownloadRecord)
		withRecordRouter.POST("/record/del", baseApi.DeleteBackupRecord)
		withRecordRouter.PUT(":id", baseApi.UpdateBackup)
	}
}
