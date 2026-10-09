package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type DatabaseRouter struct{}

func (s *DatabaseRouter) InitRouter(Router *gin.RouterGroup) {
	cmdRouter := Router.Group("databases")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		cmdRouter.POST("/common/info", baseApi.LoadDBBaseInfo)
		cmdRouter.POST("/common/load/file", baseApi.LoadDBFile)
		cmdRouter.POST("/common/update/conf", baseApi.UpdateDBConfByFile)

		cmdRouter.POST("", baseApi.CreateMysql)
		cmdRouter.POST("/users/search", baseApi.ListMysqlUsers)
		cmdRouter.POST("/users", baseApi.CreateMysqlUser)
		cmdRouter.POST("/users/update", baseApi.UpdateMysqlUser)
		cmdRouter.POST("/users/password", baseApi.ChangeMysqlUserPassword)
		cmdRouter.POST("/users/password/save", baseApi.SaveMysqlUserPassword)
		cmdRouter.POST("/users/del", baseApi.DeleteMysqlUser)
		cmdRouter.POST("/grants/search", baseApi.ListMysqlGrants)
		cmdRouter.POST("/grants/summary", baseApi.ListMysqlGrantSummary)
		cmdRouter.POST("/grants", baseApi.GrantMysqlUser)
		cmdRouter.POST("/grants/del", baseApi.RevokeMysqlGrant)
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
		cmdRouter.POST("/format/options", baseApi.ListDBFormatCollationOptions)

		cmdRouter.POST("/redis/persistence/conf", baseApi.LoadPersistenceConf)
		cmdRouter.POST("/redis/status", baseApi.LoadRedisStatus)
		cmdRouter.POST("/redis/conf", baseApi.LoadRedisConf)
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

		cmdRouter.POST("/mongodb", baseApi.CreateMongodb)
		cmdRouter.POST("/mongodb/search", baseApi.SearchMongodb)
		cmdRouter.POST("/mongodb/description", baseApi.UpdateMongodbDescription)
		cmdRouter.POST("/mongodb/load", baseApi.LoadMongodbFromRemote)
		cmdRouter.POST("/mongodb/bind", baseApi.BindMongodbUser)
		cmdRouter.POST("/mongodb/password", baseApi.ChangeMongodbPassword)
		cmdRouter.POST("/mongodb/root/password", baseApi.ChangeMongodbRootPassword)
		cmdRouter.POST("/mongodb/privileges", baseApi.LoadMongodbPrivileges)
		cmdRouter.POST("/mongodb/privileges/change", baseApi.ChangeMongodbPrivileges)
		cmdRouter.POST("/mongodb/del/check", baseApi.DeleteCheckMongodb)
		cmdRouter.POST("/mongodb/del", baseApi.DeleteMongodb)
	}
}
