package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"

	"github.com/gin-gonic/gin"
)

type HostRouter struct{}

func (s *HostRouter) InitRouter(Router *gin.RouterGroup) {
	hostRouter := Router.Group("hosts").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		hostRouter.POST("", baseApi.CreateHost)
		hostRouter.POST("/del", baseApi.DeleteHost)
		hostRouter.POST("/update", baseApi.UpdateHost)
		hostRouter.POST("/update/group", baseApi.UpdateHostGroup)
		hostRouter.POST("/search", baseApi.SearchHost)
		hostRouter.POST("/tree", baseApi.HostTree)
		hostRouter.POST("/test/byinfo", baseApi.TestByInfo)
		hostRouter.POST("/test/byid/:id", baseApi.TestByID)

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

		hostRouter.GET("/ssh/conf", baseApi.LoadSSHConf)
		hostRouter.POST("/ssh/search", baseApi.GetSSHInfo)
		hostRouter.POST("/ssh/update", baseApi.UpdateSSH)
		hostRouter.POST("/ssh/generate", baseApi.GenerateSSH)
		hostRouter.POST("/ssh/secret", baseApi.LoadSSHSecret)
		hostRouter.POST("/ssh/log", baseApi.LoadSSHLogs)
		hostRouter.POST("/ssh/conffile/update", baseApi.UpdateSSHByfile)
		hostRouter.POST("/ssh/operate", baseApi.OperateSSH)

		hostRouter.GET("/command", baseApi.ListCommand)
		hostRouter.POST("/command", baseApi.CreateCommand)
		hostRouter.POST("/command/del", baseApi.DeleteCommand)
		hostRouter.POST("/command/search", baseApi.SearchCommand)
		hostRouter.GET("/command/tree", baseApi.SearchCommandTree)
		hostRouter.POST("/command/update", baseApi.UpdateCommand)

		hostRouter.GET("/command/redis", baseApi.ListRedisCommand)
		hostRouter.POST("/command/redis", baseApi.SaveRedisCommand)
		hostRouter.POST("/command/redis/search", baseApi.SearchRedisCommand)
		hostRouter.POST("/command/redis/del", baseApi.DeleteRedisCommand)

		hostRouter.POST("/tool", baseApi.GetToolStatus)
		hostRouter.POST("/tool/init", baseApi.InitToolConfig)
		hostRouter.POST("/tool/operate", baseApi.OperateTool)
		hostRouter.POST("/tool/config", baseApi.OperateToolConfig)
		hostRouter.POST("/tool/log", baseApi.GetToolLog)
		hostRouter.POST("/tool/supervisor/process", baseApi.OperateProcess)
		hostRouter.GET("/tool/supervisor/process", baseApi.GetProcess)
		hostRouter.POST("/tool/supervisor/process/file", baseApi.GetProcessFile)
	}
}
