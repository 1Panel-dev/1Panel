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

		toolboxRouter.GET("/ftp/base", baseApi.LoadFtpBaseInfo)
		toolboxRouter.POST("/ftp/log/search", baseApi.LoadFtpLogInfo)
		toolboxRouter.POST("/ftp/operate", baseApi.OperateFtp)
		toolboxRouter.POST("/ftp/search", baseApi.SearchFtp)
		toolboxRouter.POST("/ftp", baseApi.CreateFtp)
		toolboxRouter.POST("/ftp/update", baseApi.UpdateFtp)
		toolboxRouter.POST("/ftp/del", baseApi.DeleteFtp)
		toolboxRouter.POST("/ftp/sync", baseApi.SyncFtp)

		toolboxRouter.POST("/clam/search", baseApi.SearchClam)
		toolboxRouter.POST("/clam/record/search", baseApi.SearchClamRecord)
		toolboxRouter.POST("/clam/record/clean", baseApi.CleanClamRecord)
		toolboxRouter.POST("/clam/record/log", baseApi.LoadClamRecordLog)
		toolboxRouter.POST("/clam/file/search", baseApi.SearchClamFile)
		toolboxRouter.POST("/clam/file/update", baseApi.UpdateFile)
		toolboxRouter.POST("/clam", baseApi.CreateClam)
		toolboxRouter.POST("/clam/base", baseApi.LoadClamBaseInfo)
		toolboxRouter.POST("/clam/operate", baseApi.OperateClam)
		toolboxRouter.POST("/clam/update", baseApi.UpdateClam)
		toolboxRouter.POST("/clam/status/update", baseApi.UpdateClamStatus)
		toolboxRouter.POST("/clam/del", baseApi.DeleteClam)
		toolboxRouter.POST("/clam/handle", baseApi.HandleClamScan)
	}
}
