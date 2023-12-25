package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"

	"github.com/gin-gonic/gin"
)

type DatabaseRouter struct{}

func (s *DatabaseRouter) InitRouter(Router *gin.RouterGroup) {
	cmdRouter := Router.Group("databases").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		cmdRouter.POST("", baseApi.CreateMysql)
		cmdRouter.POST("/bind", baseApi.BindUser)
		cmdRouter.POST("load", baseApi.LoadDBFromRemote)
		cmdRouter.POST("/change/access", baseApi.ChangeMysqlAccess)
		cmdRouter.POST("/change/password", baseApi.ChangeMysqlPassword)
		cmdRouter.POST("/del/check", baseApi.DeleteCheckMysql)
		cmdRouter.POST("/del", baseApi.DeleteMysql)
		cmdRouter.POST("/description/update", baseApi.UpdateMysqlDescription)
		cmdRouter.POST("/variables/update", baseApi.UpdateMysqlVariables)
		cmdRouter.POST("/conffile/update", baseApi.UpdateMysqlConfByFile)
		cmdRouter.POST("/search", baseApi.SearchMysql)
		cmdRouter.POST("/load/file", baseApi.LoadDatabaseFile)
		cmdRouter.POST("/variables", baseApi.LoadVariables)
		cmdRouter.POST("/status", baseApi.LoadStatus)
		cmdRouter.POST("/baseinfo", baseApi.LoadBaseinfo)
		cmdRouter.POST("/remote", baseApi.LoadRemoteAccess)
		cmdRouter.GET("/options", baseApi.ListDBName)

		cmdRouter.GET("/redis/persistence/conf", baseApi.LoadPersistenceConf)
		cmdRouter.GET("/redis/status", baseApi.LoadRedisStatus)
		cmdRouter.GET("/redis/conf", baseApi.LoadRedisConf)
		cmdRouter.GET("/redis/exec", baseApi.RedisWsSsh)
		cmdRouter.POST("/redis/password", baseApi.ChangeRedisPassword)
		cmdRouter.POST("/redis/backup/search", baseApi.RedisBackupList)
		cmdRouter.POST("/redis/conf/update", baseApi.UpdateRedisConf)
		cmdRouter.POST("/redis/conffile/update", baseApi.UpdateRedisConfByFile)
		cmdRouter.POST("/redis/persistence/update", baseApi.UpdateRedisPersistenceConf)

		cmdRouter.POST("/db/check", baseApi.CheckDatabase)
		cmdRouter.POST("/db", baseApi.CreateDatabase)
		cmdRouter.GET("/db/:name", baseApi.GetDatabase)
		cmdRouter.GET("/db/list/:type", baseApi.ListDatabase)
		cmdRouter.POST("/db/update", baseApi.UpdateDatabase)
		cmdRouter.POST("/db/search", baseApi.SearchDatabase)
		cmdRouter.POST("/db/del/check", baseApi.DeleteCheckDatabase)
		cmdRouter.POST("/db/del", baseApi.DeleteDatabase)
	}
}
