package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type HostRouter struct{}

func (s *HostRouter) InitRouter(Router *gin.RouterGroup) {
	hostRouter := Router.Group("hosts")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		hostRouter.GET("/firewall/base", baseApi.LoadFirewallBaseInfo)
		hostRouter.POST("/firewall/search", baseApi.SearchFirewallRule)
		hostRouter.POST("/firewall/operate", baseApi.OperateFirewall)
		hostRouter.POST("/firewall/port", baseApi.OperatePortRule)
		hostRouter.POST("/firewall/forward", baseApi.OperateForwardRule)
		hostRouter.POST("/firewall/ip", baseApi.OperateIPRule)
		hostRouter.POST("/firewall/batch", baseApi.BatchOperateRule)
		hostRouter.POST("/firewall/update/port", baseApi.UpdatePortRule)
		hostRouter.POST("/firewall/update/addr", baseApi.UpdateAddrRule)
		hostRouter.POST("/firewall/update/description", baseApi.UpdateFirewallDescription)

		hostRouter.POST("/monitor/search", baseApi.LoadMonitor)
		hostRouter.POST("/monitor/clean", baseApi.CleanMonitor)
		hostRouter.GET("/monitor/netoptions", baseApi.GetNetworkOptions)
		hostRouter.GET("/monitor/iooptions", baseApi.GetIOOptions)
		hostRouter.GET("/monitor/setting", baseApi.LoadMonitorSetting)
		hostRouter.POST("/monitor/setting/update", baseApi.UpdateMonitorSetting)

		hostRouter.POST("/ssh/search", baseApi.GetSSHInfo)
		hostRouter.POST("/ssh/update", baseApi.UpdateSSH)
		hostRouter.POST("/ssh/log", baseApi.LoadSSHLogs)
		hostRouter.POST("/ssh/log/export", baseApi.ExportSSHLogs)
		hostRouter.POST("/ssh/operate", baseApi.OperateSSH)
		hostRouter.POST("/ssh/file", baseApi.LoadSSHFile)
		hostRouter.POST("/ssh/file/update", baseApi.UpdateSSHByFile)

		hostRouter.POST("/ssh/cert", baseApi.CreateRootCert)
		hostRouter.POST("/ssh/cert/update", baseApi.EditRootCert)
		hostRouter.POST("/ssh/cert/sync", baseApi.SyncRootCert)
		hostRouter.POST("/ssh/cert/search", baseApi.SearchRootCert)
		hostRouter.POST("/ssh/cert/delete", baseApi.DeleteRootCert)

		hostRouter.POST("/tool", baseApi.GetToolStatus)
		hostRouter.POST("/tool/init", baseApi.InitToolConfig)
		hostRouter.POST("/tool/operate", baseApi.OperateTool)
		hostRouter.POST("/tool/config", baseApi.OperateToolConfig)
		hostRouter.POST("/tool/supervisor/process", baseApi.OperateProcess)
		hostRouter.GET("/tool/supervisor/process", baseApi.GetProcess)
		hostRouter.POST("/tool/supervisor/process/file", baseApi.GetProcessFile)

		hostRouter.GET("/terminal", baseApi.WsSSH)

		hostRouter.GET("/disks", baseApi.GetCompleteDiskInfo)
		hostRouter.POST("/disks/partition", baseApi.PartitionDisk)
		hostRouter.POST("/disks/mount", baseApi.MountDisk)
		hostRouter.POST("/disks/unmount", baseApi.UnmountDisk)

		hostRouter.GET("/components/:name", baseApi.CheckComponentExistence)
	}
}
