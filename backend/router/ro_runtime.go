package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type RuntimeRouter struct {
}

func (r *RuntimeRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("runtimes")
	groupRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.PasswordExpired())

	baseApi := v1.ApiGroupApp.BaseApi
	{
		groupRouter.GET("/installed/delete/check/:runTimeId", baseApi.DeleteRuntimeCheck)
		groupRouter.POST("/search", baseApi.SearchRuntimes)
		groupRouter.POST("", baseApi.CreateRuntime)
		groupRouter.POST("/del", baseApi.DeleteRuntime)
		groupRouter.POST("/update", baseApi.UpdateRuntime)
		groupRouter.GET("/:id", baseApi.GetRuntime)
		groupRouter.POST("/sync", baseApi.SyncStatus)

		groupRouter.POST("/node/package", baseApi.GetNodePackageRunScript)
		groupRouter.POST("/operate", baseApi.OperateRuntime)
		groupRouter.POST("/node/modules", baseApi.GetNodeModules)
		groupRouter.POST("/node/modules/operate", baseApi.OperateNodeModules)

		groupRouter.POST("/php/extensions/search", baseApi.PagePHPExtensions)
		groupRouter.POST("/php/extensions", baseApi.CreatePHPExtensions)
		groupRouter.POST("/php/extensions/update", baseApi.UpdatePHPExtensions)
		groupRouter.POST("/php/extensions/del", baseApi.DeletePHPExtensions)
	}

}
