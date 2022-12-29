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
	baseApi := v1.ApiGroupApp.BaseApi
	{
		cmdRouter.POST("", baseApi.CreateMysql)
		cmdRouter.POST("/change/access", baseApi.ChangeMysqlAccess)
		cmdRouter.POST("/change/password", baseApi.ChangeMysqlPassword)
		cmdRouter.POST("/backup", baseApi.BackupMysql)
		cmdRouter.POST("/recover/byupload", baseApi.RecoverMysqlByUpload)
		cmdRouter.POST("/recover", baseApi.RecoverMysql)
		cmdRouter.POST("/del/check", baseApi.DeleteCheckMysql)
		cmdRouter.POST("/del", baseApi.DeleteMysql)
		cmdRouter.POST("/description/update", baseApi.UpdateMysqlDescription)
		cmdRouter.POST("/variables/update", baseApi.UpdateMysqlVariables)
		cmdRouter.POST("/conffile/update", baseApi.UpdateMysqlConfByFile)
		cmdRouter.POST("/search", baseApi.SearchMysql)
		cmdRouter.GET("/variables", baseApi.LoadVariables)
		cmdRouter.GET("/status", baseApi.LoadStatus)
		cmdRouter.GET("/baseinfo", baseApi.LoadBaseinfo)
		cmdRouter.GET("/remote", baseApi.LoadRemoteAccess)
		cmdRouter.GET("/options", baseApi.ListDBName)

		cmdRouter.GET("/redis/persistence/conf", baseApi.LoadPersistenceConf)
		cmdRouter.GET("/redis/status", baseApi.LoadRedisStatus)
		cmdRouter.GET("/redis/conf", baseApi.LoadRedisConf)
		cmdRouter.GET("/redis/exec", baseApi.RedisWsSsh)
		cmdRouter.POST("/redis/password", baseApi.ChangeRedisPassword)
		cmdRouter.POST("/redis/backup", baseApi.RedisBackup)
		cmdRouter.POST("/redis/recover", baseApi.RedisRecover)
		cmdRouter.POST("/redis/backup/search", baseApi.RedisBackupList)
		cmdRouter.POST("/redis/conf/update", baseApi.UpdateRedisConf)
		cmdRouter.POST("/redis/conffile/update", baseApi.UpdateRedisConfByFile)
		cmdRouter.POST("/redis/persistence/update", baseApi.UpdateRedisPersistenceConf)
	}
}
