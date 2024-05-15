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
		cmdRouter.POST("/common/info", baseApi.LoadDBBaseInfo)
		cmdRouter.POST("/common/load/file", baseApi.LoadDBFile)
		cmdRouter.POST("/common/update/conf", baseApi.UpdateDBConfByFile)

		cmdRouter.POST("", baseApi.CreateMysql)
		cmdRouter.POST("/bind", baseApi.BindUser)
		cmdRouter.POST("load", baseApi.LoadDBFromRemote)
		cmdRouter.POST("/change/access", baseApi.ChangeMysqlAccess)
		cmdRouter.POST("/change/password", baseApi.ChangeMysqlPassword)
		cmdRouter.POST("/del/check", baseApi.DeleteCheckMysql)
		cmdRouter.POST("/del", baseApi.DeleteMysql)
		cmdRouter.POST("/description/update", baseApi.UpdateMysqlDescription)
		cmdRouter.POST("/variables/update", baseApi.UpdateMysqlVariables)
		cmdRouter.POST("/search", baseApi.SearchMysql)
		cmdRouter.POST("/variables", baseApi.LoadVariables)
		cmdRouter.POST("/status", baseApi.LoadStatus)
		cmdRouter.POST("/remote", baseApi.LoadRemoteAccess)
		cmdRouter.GET("/options", baseApi.ListDBName)

		cmdRouter.POST("/redis/persistence/conf", baseApi.LoadPersistenceConf)
		cmdRouter.POST("/redis/status", baseApi.LoadRedisStatus)
		cmdRouter.POST("/redis/conf", baseApi.LoadRedisConf)
		cmdRouter.GET("/redis/exec", baseApi.RedisWsSsh)
		cmdRouter.GET("/redis/check", baseApi.CheckHasCli)
		cmdRouter.POST("/redis/install/cli", baseApi.InstallCli)
		cmdRouter.POST("/redis/password", baseApi.ChangeRedisPassword)
		cmdRouter.POST("/redis/conf/update", baseApi.UpdateRedisConf)
		cmdRouter.POST("/redis/persistence/update", baseApi.UpdateRedisPersistenceConf)

		cmdRouter.POST("/db/check", baseApi.CheckDatabase)
		cmdRouter.POST("/db", baseApi.CreateDatabase)
		cmdRouter.GET("/db/:name", baseApi.GetDatabase)
		cmdRouter.GET("/db/list/:type", baseApi.ListDatabase)
		cmdRouter.GET("/db/item/:type", baseApi.LoadDatabaseItems)
		cmdRouter.POST("/db/update", baseApi.UpdateDatabase)
		cmdRouter.POST("/db/search", baseApi.SearchDatabase)
		cmdRouter.POST("/db/del/check", baseApi.DeleteCheckDatabase)
		cmdRouter.POST("/db/del", baseApi.DeleteDatabase)

		cmdRouter.POST("/pg", baseApi.CreatePostgresql)
		cmdRouter.POST("/pg/search", baseApi.SearchPostgresql)
		cmdRouter.POST("/pg/:database/load", baseApi.LoadPostgresqlDBFromRemote)
		cmdRouter.POST("/pg/bind", baseApi.BindPostgresqlUser)
		cmdRouter.POST("/pg/del/check", baseApi.DeleteCheckPostgresql)
		cmdRouter.POST("/pg/del", baseApi.DeletePostgresql)
		cmdRouter.POST("/pg/privileges", baseApi.ChangePostgresqlPrivileges)
		cmdRouter.POST("/pg/password", baseApi.ChangePostgresqlPassword)
		cmdRouter.POST("/pg/description", baseApi.UpdatePostgresqlDescription)
	}
}
