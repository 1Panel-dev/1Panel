package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type RuntimeRouter struct {
}

func (r *RuntimeRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("runtimes")

	baseApi := v2.ApiGroupApp.BaseApi
	{
		groupRouter.GET("/installed/delete/check/:id", baseApi.DeleteRuntimeCheck)
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

		groupRouter.GET("/php/:id/extensions", baseApi.GetRuntimeExtension)
		groupRouter.POST("/php/extensions/install", baseApi.InstallPHPExtension)
		groupRouter.POST("/php/extensions/uninstall", baseApi.UnInstallPHPExtension)

		groupRouter.GET("/php/config/:id", baseApi.GetPHPConfig)
		groupRouter.POST("/php/config", baseApi.UpdatePHPConfig)
		groupRouter.POST("/php/update", baseApi.UpdatePHPFile)
		groupRouter.POST("/php/file", baseApi.GetPHPConfigFile)
		groupRouter.POST("/php/fpm/config", baseApi.UpdateFPMConfig)
		groupRouter.GET("/php/fpm/config/:id", baseApi.GetFPMConfig)

		groupRouter.POST("/php/container/update", baseApi.UpdatePHPContainer)
		groupRouter.GET("/php/container/:id", baseApi.GetPHPContainerConfig)

		groupRouter.GET("/supervisor/process/:id", baseApi.GetSupervisorProcess)
		groupRouter.POST("/supervisor/process", baseApi.OperateSupervisorProcess)
		groupRouter.POST("/supervisor/process/file", baseApi.OperateSupervisorProcessFile)

		groupRouter.POST("/remark", baseApi.UpdateRuntimeRemark)
	}
}
