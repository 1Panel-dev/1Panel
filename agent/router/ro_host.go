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
		hostRouter.POST("", baseApi.CreateHost)
		hostRouter.POST("/info", baseApi.GetHostByID)
		hostRouter.POST("/del", baseApi.DeleteHost)
		hostRouter.POST("/update", baseApi.UpdateHost)
		hostRouter.POST("/update/group", baseApi.UpdateHostGroup)
		hostRouter.POST("/search", baseApi.SearchHost)
		hostRouter.POST("/tree", baseApi.HostTree)
		hostRouter.POST("/test/byinfo", baseApi.TestByInfo)
		hostRouter.POST("/test/byid", baseApi.TestByID)

		hostRouter.POST("/firewall/base", baseApi.LoadFirewallBaseInfo)
		hostRouter.POST("/firewall/operate", baseApi.OperateFirewall)
		hostRouter.GET("/firewall/settings", baseApi.LoadFirewallSettings)
		hostRouter.POST("/firewall/settings/operate", baseApi.OperateFirewallBackend)
		hostRouter.POST("/firewall/forward/base", baseApi.LoadForwardingBaseInfo)
		hostRouter.POST("/firewall/forward/search", baseApi.SearchForwardingRules)
		hostRouter.POST("/firewall/forward/operate", baseApi.OperateForwardingRules)
		hostRouter.POST("/firewall/forward/enable", baseApi.EnableForwarding)
		hostRouter.POST("/firewall/rules/search", baseApi.SearchFirewallRules)
		hostRouter.POST("/firewall/rules/reset", baseApi.ResetFirewallRules)
		hostRouter.POST("/firewall/rules/native/detail", baseApi.LoadFirewallNativeDetail)
		hostRouter.POST("/firewall/rules/check", baseApi.CheckFirewallRules)
		hostRouter.POST("/firewall/rules", baseApi.CreateFirewallRules)
		hostRouter.POST("/firewall/rules/sync/preview", baseApi.PreviewFirewallRuleSync)
		hostRouter.GET("/firewall/rules/sync/task", baseApi.LoadFirewallRuleSyncTask)
		hostRouter.POST("/firewall/rules/sync", baseApi.SyncFirewallRules)
		hostRouter.POST("/firewall/rules/update", baseApi.UpdateFirewallRule)
		hostRouter.POST("/firewall/rules/delete", baseApi.DeleteFirewallRules)
		hostRouter.POST("/firewall/rules/reorder", baseApi.ReorderFirewallRule)

		hostRouter.POST("/firewall/filter/operate", baseApi.OperateFilterChain)
		hostRouter.GET("/firewall/docker/ports", baseApi.ListDockerPortGuard)
		hostRouter.POST("/firewall/docker/sync", baseApi.SyncDockerPortGuard)
		hostRouter.POST("/firewall/docker/operate", baseApi.OperateDockerPortGuard)
		hostRouter.POST("/firewall/docker/policies/batch", baseApi.UpsertDockerPortGuardPolicies)
		hostRouter.POST("/firewall/docker/policies/delete/batch", baseApi.DeleteDockerPortGuardPolicies)

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
		hostRouter.POST("/ssh/log/clean", baseApi.CleanSSHLogs)
		hostRouter.POST("/ssh/operate", baseApi.OperateSSH)
		hostRouter.POST("/ssh/file", baseApi.LoadSSHFile)
		hostRouter.POST("/ssh/file/update", baseApi.UpdateSSHByFile)

		hostRouter.POST("/ssh/cert", baseApi.CreateRootCert)
		hostRouter.POST("/ssh/cert/update", baseApi.EditRootCert)
		hostRouter.POST("/ssh/cert/sync", baseApi.SyncRootCert)
		hostRouter.POST("/ssh/cert/search", baseApi.SearchRootCert)
		hostRouter.POST("/ssh/cert/delete", baseApi.DeleteRootCert)

		hostRouter.POST("/tool/status", baseApi.GetToolStatus)
		hostRouter.POST("/tool/init", baseApi.InitToolConfig)
		hostRouter.POST("/tool/operate", baseApi.OperateTool)
		hostRouter.POST("/tool/config/get", baseApi.GetToolConfig)
		hostRouter.POST("/tool/config/set", baseApi.UpdateToolConfig)
		hostRouter.POST("/tool/supervisor/process", baseApi.OperateProcess)
		hostRouter.GET("/tool/supervisor/process", baseApi.GetProcess)
		hostRouter.POST("/tool/supervisor/process/file/get", baseApi.GetProcessFile)
		hostRouter.POST("/tool/supervisor/process/file", baseApi.OperateProcessFile)

		hostRouter.GET("/terminal/local", baseApi.WsLocalTerminal)
		hostRouter.GET("/terminal/ssh", baseApi.WsHostSSH)
		hostRouter.GET("/terminal/container", baseApi.WsContainerTerminal)

		hostRouter.GET("/disks", baseApi.GetCompleteDiskInfo)
		hostRouter.POST("/disks/partition", baseApi.PartitionDisk)
		hostRouter.POST("/disks/mount", baseApi.MountDisk)
		hostRouter.POST("/disks/unmount", baseApi.UnmountDisk)

		hostRouter.GET("/components/:name", baseApi.CheckComponentExistence)
	}
}
