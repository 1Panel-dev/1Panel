package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"

	"github.com/gin-gonic/gin"
)

type ToolboxRouter struct{}

func (s *ToolboxRouter) InitRouter(Router *gin.RouterGroup) {
	toolboxRouter := Router.Group("toolbox").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		toolboxRouter.POST("/device/base", baseApi.LoadDeviceBaseInfo)
		toolboxRouter.GET("/device/zone/options", baseApi.LoadTimeOption)
		toolboxRouter.POST("/device/update/conf", baseApi.UpdateDeviceConf)
		toolboxRouter.POST("/device/update/host", baseApi.UpdateDeviceHost)
		toolboxRouter.POST("/device/update/passwd", baseApi.UpdateDevicePasswd)
		toolboxRouter.POST("/device/update/swap", baseApi.UpdateDeviceSwap)
		toolboxRouter.POST("/device/update/byconf", baseApi.UpdateDeviceByFile)
		toolboxRouter.POST("/device/check/dns", baseApi.CheckDNS)
		toolboxRouter.POST("/device/conf", baseApi.LoadDeviceConf)

		toolboxRouter.POST("/scan", baseApi.ScanSystem)
		toolboxRouter.POST("/clean", baseApi.SystemClean)

		toolboxRouter.GET("/fail2ban/base", baseApi.LoadFail2BanBaseInfo)
		toolboxRouter.GET("/fail2ban/load/conf", baseApi.LoadFail2BanConf)
		toolboxRouter.POST("/fail2ban/search", baseApi.SearchFail2Ban)
		toolboxRouter.POST("/fail2ban/operate", baseApi.OperateFail2Ban)
		toolboxRouter.POST("/fail2ban/operate/sshd", baseApi.OperateSSHD)
		toolboxRouter.POST("/fail2ban/update", baseApi.UpdateFail2BanConf)
		toolboxRouter.POST("/fail2ban/update/byconf", baseApi.UpdateFail2BanConfByFile)
	}
}
