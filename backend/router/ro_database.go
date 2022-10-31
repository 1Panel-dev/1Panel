package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"

	"github.com/gin-gonic/gin"
)

type DatabaseRouter struct{}

func (s *DatabaseRouter) InitDatabaseRouter(Router *gin.RouterGroup) {
	cmdRouter := Router.Group("databases").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	withRecordRouter := Router.Group("databases").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired()).
		Use(middleware.OperationRecord())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		withRecordRouter.POST("", baseApi.CreateMysql)
		withRecordRouter.PUT("/:id", baseApi.UpdateMysql)
		withRecordRouter.POST("/backup", baseApi.BackupMysql)
		withRecordRouter.POST("/recover", baseApi.RecoverMysql)
		withRecordRouter.POST("/backups/search", baseApi.SearchDBBackups)
		withRecordRouter.POST("/del", baseApi.DeleteMysql)
		withRecordRouter.POST("/variables/update", baseApi.UpdateMysqlVariables)
		cmdRouter.POST("/search", baseApi.SearchMysql)
		cmdRouter.GET("/variables/:name", baseApi.LoadVariables)
		cmdRouter.GET("/status/:name", baseApi.LoadStatus)
		cmdRouter.GET("/baseinfo/:name", baseApi.LoadBaseinfo)
		cmdRouter.GET("/versions", baseApi.LoadVersions)
		cmdRouter.GET("/dbs/:name", baseApi.ListDBNameByVersion)

		cmdRouter.POST("/redis/search", baseApi.SearchRedis)
		withRecordRouter.POST("/redis", baseApi.SetRedis)
		withRecordRouter.POST("/redis/del", baseApi.DeleteRedis)
		withRecordRouter.POST("/redis/cleanall", baseApi.CleanRedis)
		withRecordRouter.POST("/redis/status", baseApi.LoadRedisStatus)
		withRecordRouter.POST("/redis/conf/update", baseApi.UpdateRedisConf)
		withRecordRouter.POST("/redis/conf", baseApi.LoadRedisConf)
		cmdRouter.GET("/redis/versions", baseApi.LoadRedisRunningVersion)
	}
}
