package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type WebsiteDnsRecordRouter struct {
}

func (a *WebsiteDnsRecordRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("websites/dns/record")

	baseApi := v2.ApiGroupApp.BaseApi
	{
		groupRouter.POST("/zones", baseApi.ListWebsiteDnsZones)
		groupRouter.POST("/search", baseApi.ListWebsiteDnsRecords)
		groupRouter.POST("", baseApi.CreateWebsiteDnsRecord)
		groupRouter.POST("/update", baseApi.UpdateWebsiteDnsRecord)
		groupRouter.POST("/del", baseApi.DeleteWebsiteDnsRecord)
	}
}
