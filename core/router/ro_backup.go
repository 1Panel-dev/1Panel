package router

import (
	v2 "github.com/1Panel-dev/1Panel/core/app/api/v2"
	"github.com/1Panel-dev/1Panel/core/middleware"
	"github.com/gin-gonic/gin"
)

type BackupRouter struct{}

func (s *BackupRouter) InitRouter(Router *gin.RouterGroup) {
	backupRouter := Router.Group("backup").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v2.ApiGroupApp.BaseApi
	{
		backupRouter.GET("/local", baseApi.GetLocalDir)
		backupRouter.GET("/onedrive", baseApi.LoadOneDriveInfo)
		backupRouter.GET("/options", baseApi.LoadBackupOptions)
		backupRouter.POST("/search", baseApi.SearchBackup)
		backupRouter.POST("/refresh/onedrive", baseApi.RefreshOneDriveToken)
		backupRouter.POST("/buckets", baseApi.ListBuckets)
		backupRouter.POST("", baseApi.CreateBackup)
		backupRouter.POST("/del", baseApi.DeleteBackup)
		backupRouter.POST("/update", baseApi.UpdateBackup)
	}
}
