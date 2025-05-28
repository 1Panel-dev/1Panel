package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type WebsiteRouter struct {
}

func (a *WebsiteRouter) InitRouter(Router *gin.RouterGroup) {
	websiteRouter := Router.Group("websites")

	baseApi := v2.ApiGroupApp.BaseApi
	{
		websiteRouter.POST("/search", baseApi.PageWebsite)
		websiteRouter.GET("/list", baseApi.GetWebsites)
		websiteRouter.POST("", baseApi.CreateWebsite)
		websiteRouter.POST("/operate", baseApi.OpWebsite)
		websiteRouter.POST("/log", baseApi.OpWebsiteLog)
		websiteRouter.POST("/check", baseApi.CreateWebsiteCheck)
		websiteRouter.POST("/options", baseApi.GetWebsiteOptions)
		websiteRouter.POST("/update", baseApi.UpdateWebsite)
		websiteRouter.GET("/:id", baseApi.GetWebsite)
		websiteRouter.POST("/del", baseApi.DeleteWebsite)
		websiteRouter.POST("/default/server", baseApi.ChangeDefaultServer)
		websiteRouter.POST("/group/change", baseApi.ChangeWebsiteGroup)

		websiteRouter.GET("/domains/:websiteId", baseApi.GetWebDomains)
		websiteRouter.POST("/domains/del", baseApi.DeleteWebDomain)
		websiteRouter.POST("/domains", baseApi.CreateWebDomain)
		websiteRouter.POST("/domains/update", baseApi.UpdateWebDomain)

		websiteRouter.GET("/:id/config/:type", baseApi.GetWebsiteNginx)
		websiteRouter.POST("/config", baseApi.GetNginxConfig)
		websiteRouter.POST("/config/update", baseApi.UpdateNginxConfig)
		websiteRouter.POST("/nginx/update", baseApi.UpdateWebsiteNginxConfig)

		websiteRouter.GET("/:id/https", baseApi.GetHTTPSConfig)
		websiteRouter.POST("/:id/https", baseApi.UpdateHTTPSConfig)

		websiteRouter.POST("/rewrite", baseApi.GetRewriteConfig)
		websiteRouter.POST("/rewrite/update", baseApi.UpdateRewriteConfig)
		websiteRouter.POST("/rewrite/custom", baseApi.OperateCustomRewrite)
		websiteRouter.GET("/rewrite/custom", baseApi.ListCustomRewrite)

		websiteRouter.POST("/dir/update", baseApi.UpdateSiteDir)
		websiteRouter.POST("/dir/permission", baseApi.UpdateSiteDirPermission)
		websiteRouter.POST("/dir", baseApi.GetDirConfig)

		websiteRouter.POST("/proxies", baseApi.GetProxyConfig)
		websiteRouter.POST("/proxies/update", baseApi.UpdateProxyConfig)
		websiteRouter.POST("/proxies/file", baseApi.UpdateProxyConfigFile)
		websiteRouter.POST("/proxy/config", baseApi.UpdateProxyCache)
		websiteRouter.GET("/proxy/config/:id", baseApi.GetProxyCache)
		websiteRouter.POST("/proxy/clear", baseApi.ClearProxyCache)

		websiteRouter.POST("/auths", baseApi.GetAuthConfig)
		websiteRouter.POST("/auths/update", baseApi.UpdateAuthConfig)
		websiteRouter.POST("/auths/path", baseApi.GetPathAuthConfig)
		websiteRouter.POST("/auths/path/update", baseApi.UpdatePathAuthConfig)

		websiteRouter.POST("/leech", baseApi.GetAntiLeech)
		websiteRouter.POST("/leech/update", baseApi.UpdateAntiLeech)

		websiteRouter.POST("/redirect/update", baseApi.UpdateRedirectConfig)
		websiteRouter.POST("/redirect", baseApi.GetRedirectConfig)
		websiteRouter.POST("/redirect/file", baseApi.UpdateRedirectConfigFile)

		websiteRouter.GET("/default/html/:type", baseApi.GetDefaultHtml)
		websiteRouter.POST("/default/html/update", baseApi.UpdateDefaultHtml)

		websiteRouter.GET("/:id/lbs", baseApi.GetLoadBalances)
		websiteRouter.POST("/lbs/create", baseApi.CreateLoadBalance)
		websiteRouter.POST("/lbs/del", baseApi.DeleteLoadBalance)
		websiteRouter.POST("/lbs/update", baseApi.UpdateLoadBalance)
		websiteRouter.POST("/lbs/file", baseApi.UpdateLoadBalanceFile)

		websiteRouter.POST("/php/version", baseApi.ChangePHPVersion)

		websiteRouter.POST("/realip/config", baseApi.SetRealIPConfig)
		websiteRouter.GET("/realip/config/:id", baseApi.GetRealIPConfig)

		websiteRouter.GET("/resource/:id", baseApi.GetWebsiteResource)
		websiteRouter.GET("/databases", baseApi.GetWebsiteDatabase)
		websiteRouter.POST("/databases", baseApi.ChangeWebsiteDatabase)

		websiteRouter.POST("/crosssite", baseApi.OperateCrossSiteAccess)
	}
}
