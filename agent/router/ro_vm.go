package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type VMRouter struct{}

func (s *VMRouter) InitRouter(Router *gin.RouterGroup) {
	vmRouter := Router.Group("vms")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		vmRouter.GET("/environment", baseApi.CheckVMEnvironment)
		vmRouter.POST("/environment/enable", baseApi.EnableVMEnvironment)
		vmRouter.POST("/search", baseApi.SearchVM)
		vmRouter.POST("/sync", baseApi.SyncVMs)
		vmRouter.POST("/info", baseApi.GetVM)
		vmRouter.POST("/health/check", baseApi.HealthCheckVM)
		vmRouter.POST("/stats", baseApi.GetVMStats)
		vmRouter.POST("", baseApi.CreateVM)
		vmRouter.POST("/templates/list", baseApi.ListVMTemplates)
		vmRouter.POST("/templates", baseApi.CreateVMTemplate)
		vmRouter.POST("/templates/del", baseApi.DeleteVMTemplate)
		vmRouter.POST("/update", baseApi.UpdateVM)
		vmRouter.POST("/disks/resize", baseApi.ResizeVMDisk)
		vmRouter.POST("/rename", baseApi.RenameVM)
		vmRouter.POST("/del", baseApi.DeleteVM)
		vmRouter.POST("/snapshots/list", baseApi.ListVMSnapshots)
		vmRouter.POST("/snapshots", baseApi.CreateVMSnapshot)
		vmRouter.POST("/snapshots/recover", baseApi.RecoverVMSnapshot)
		vmRouter.POST("/snapshots/del", baseApi.DeleteVMSnapshot)
		vmRouter.POST("/orphans/scan", baseApi.ScanVMOrphans)
		vmRouter.POST("/orphans/clean", baseApi.CleanVMOrphans)
		vmRouter.POST("/operate", baseApi.OperateVM)
		vmRouter.POST("/console", baseApi.GetVMConsole)
		vmRouter.GET("/console/ws", baseApi.ProxyVMConsole)

		vmRouter.POST("/iso/search", baseApi.SearchVMIso)
		vmRouter.POST("/iso/list", baseApi.ListVMIso)
		vmRouter.POST("/iso", baseApi.CreateVMIso)
		vmRouter.POST("/iso/import/start", baseApi.StartVMIsoImport)
		vmRouter.POST("/iso/import/update", baseApi.UpdateVMIsoImport)
		vmRouter.POST("/iso/update", baseApi.UpdateVMIso)
		vmRouter.POST("/iso/del", baseApi.DeleteVMIso)

		vmRouter.POST("/networks/list", baseApi.ListVMNetworks)
		vmRouter.GET("/networks/bridges", baseApi.ListVMNetworkBridges)
		vmRouter.GET("/networks/interfaces", baseApi.ListVMNetworkInterfaces)
		vmRouter.POST("/networks", baseApi.CreateVMNetwork)
		vmRouter.POST("/networks/update", baseApi.UpdateVMNetwork)
		vmRouter.POST("/networks/del", baseApi.DeleteVMNetwork)

		vmRouter.POST("/storages/list", baseApi.ListVMStorages)
		vmRouter.POST("/storages", baseApi.CreateVMStorage)
		vmRouter.POST("/storages/update", baseApi.UpdateVMStorage)
		vmRouter.POST("/storages/del", baseApi.DeleteVMStorage)
	}
}
