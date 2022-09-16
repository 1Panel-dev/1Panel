package router

import (
	v1 "github.com/1Panel-dev/1Panel/app/api/v1"
	"github.com/1Panel-dev/1Panel/middleware"

	"github.com/gin-gonic/gin"
)

type BackupRouter struct{}

func (s *BackupRouter) InitBackupRouter(Router *gin.RouterGroup) {
	baRouter := Router.Group("backups").Use(middleware.JwtAuth()).Use(middleware.SessionAuth())
	withRecordRouter := Router.Group("backups").Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.OperationRecord())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		baRouter.POST("/search", baseApi.PageBackup)
		baRouter.POST("/buckets", baseApi.ListBuckets)
		withRecordRouter.POST("", baseApi.CreateBackup)
		withRecordRouter.POST("/del", baseApi.DeleteBackup)
		withRecordRouter.PUT(":id", baseApi.UpdateBackup)
	}
}
