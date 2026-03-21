package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type DnsRouter struct{}

func (a *DnsRouter) InitRouter(Router *gin.RouterGroup) {
	dnsRouter := Router.Group("dns")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		dnsRouter.POST("/zones/search", baseApi.PageDnsZone)
		dnsRouter.POST("/zones", baseApi.CreateDnsZone)
		dnsRouter.POST("/zones/update", baseApi.UpdateDnsZone)
		dnsRouter.GET("/zones/:id", baseApi.GetDnsZone)
		dnsRouter.POST("/zones/del", baseApi.DeleteDnsZone)
		dnsRouter.POST("/zones/sync", baseApi.SyncDnsZone)

		dnsRouter.POST("/records/search", baseApi.PageDnsRecord)
		dnsRouter.POST("/records", baseApi.CreateDnsRecord)
		dnsRouter.POST("/records/update", baseApi.UpdateDnsRecord)
		dnsRouter.POST("/records/del", baseApi.DeleteDnsRecord)

		dnsRouter.POST("/setup", baseApi.SetupDns)
		dnsRouter.GET("/status", baseApi.GetDnsStatus)
	}
}
