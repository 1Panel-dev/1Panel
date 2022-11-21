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
		cmdRouter.POST("/uplist", baseApi.MysqlUpList)
		cmdRouter.POST("/uplist/upload/:mysqlName", baseApi.UploadMysqlFiles)
		withRecordRouter.POST("/recover/byupload", baseApi.RecoverMysqlByUpload)
		withRecordRouter.POST("/recover", baseApi.RecoverMysql)
		withRecordRouter.POST("/backups/search", baseApi.SearchDBBackups)
		withRecordRouter.POST("/del", baseApi.DeleteMysql)
		withRecordRouter.POST("/variables/update", baseApi.UpdateMysqlVariables)
		withRecordRouter.POST("/conf/update/byfile", baseApi.UpdateMysqlConfByFile)
		cmdRouter.POST("/search", baseApi.SearchMysql)
		cmdRouter.GET("/variables", baseApi.LoadVariables)
		cmdRouter.GET("/status", baseApi.LoadStatus)
		cmdRouter.GET("/baseinfo", baseApi.LoadBaseinfo)
		cmdRouter.GET("/dbs", baseApi.ListDBName)

		cmdRouter.GET("/redis/persistence/conf", baseApi.LoadPersistenceConf)
		cmdRouter.GET("/redis/status", baseApi.LoadRedisStatus)
		cmdRouter.GET("/redis/conf", baseApi.LoadRedisConf)
		cmdRouter.GET("/redis/exec", baseApi.RedisExec)
		cmdRouter.POST("/redis/backup", baseApi.RedisBackup)
		cmdRouter.POST("/redis/recover", baseApi.RedisRecover)
		cmdRouter.POST("/redis/backup/records", baseApi.RedisBackupList)
		cmdRouter.POST("/redis/backup/del", baseApi.RedisBackupDelete)
		cmdRouter.POST("/redis/conf/update", baseApi.UpdateRedisConf)
		cmdRouter.POST("/redis/conf/update/byfile", baseApi.UpdateRedisConfByFile)
		cmdRouter.POST("/redis/conf/update/persistence", baseApi.UpdateRedisPersistenceConf)
	}
}
