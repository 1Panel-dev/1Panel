package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type WebsiteRouter struct {
}

func (a *WebsiteRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("websites")
	groupRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.PasswordExpired())

	baseApi := v1.ApiGroupApp.BaseApi
	{
		groupRouter.POST("/search", baseApi.PageWebsite)
		groupRouter.GET("/list", baseApi.GetWebsites)
		groupRouter.POST("", baseApi.CreateWebsite)
		groupRouter.POST("/operate", baseApi.OpWebsite)
		groupRouter.POST("/log", baseApi.OpWebsiteLog)
		groupRouter.POST("/check", baseApi.CreateWebsiteCheck)
		groupRouter.GET("/options", baseApi.GetWebsiteOptions)
		groupRouter.POST("/update", baseApi.UpdateWebsite)
		groupRouter.GET("/:id", baseApi.GetWebsite)
		groupRouter.POST("/del", baseApi.DeleteWebsite)
		groupRouter.POST("/default/server", baseApi.ChangeDefaultServer)

		groupRouter.GET("/domains/:websiteId", baseApi.GetWebDomains)
		groupRouter.POST("/domains/del", baseApi.DeleteWebDomain)
		groupRouter.POST("/domains", baseApi.CreateWebDomain)

		groupRouter.GET("/:id/config/:type", baseApi.GetWebsiteNginx)
		groupRouter.POST("/config", baseApi.GetNginxConfig)
		groupRouter.POST("/config/update", baseApi.UpdateNginxConfig)
		groupRouter.POST("/nginx/update", baseApi.UpdateWebsiteNginxConfig)

		groupRouter.GET("/:id/https", baseApi.GetHTTPSConfig)
		groupRouter.POST("/:id/https", baseApi.UpdateHTTPSConfig)

		groupRouter.POST("/waf/config", baseApi.GetWebsiteWafConfig)
		groupRouter.POST("/waf/update", baseApi.UpdateWebsiteWafConfig)
		groupRouter.POST("/waf/file/update", baseApi.UpdateWebsiteWafFile)

		groupRouter.GET("/php/config/:id", baseApi.GetWebsitePHPConfig)
		groupRouter.POST("/php/config", baseApi.UpdateWebsitePHPConfig)
		groupRouter.POST("/php/update", baseApi.UpdatePHPFile)
		groupRouter.POST("/php/version", baseApi.ChangePHPVersion)

		groupRouter.POST("/rewrite", baseApi.GetRewriteConfig)
		groupRouter.POST("/rewrite/update", baseApi.UpdateRewriteConfig)

		groupRouter.POST("/dir/update", baseApi.UpdateSiteDir)
		groupRouter.POST("/dir/permission", baseApi.UpdateSiteDirPermission)
		groupRouter.POST("/dir", baseApi.GetDirConfig)

		groupRouter.POST("/proxies", baseApi.GetProxyConfig)
		groupRouter.POST("/proxies/update", baseApi.UpdateProxyConfig)
		groupRouter.POST("/proxies/file", baseApi.UpdateProxyConfigFile)

		groupRouter.POST("/auths", baseApi.GetAuthConfig)
		groupRouter.POST("/auths/update", baseApi.UpdateAuthConfig)

		groupRouter.POST("/leech", baseApi.GetAntiLeech)
		groupRouter.POST("/leech/update", baseApi.UpdateAntiLeech)

		groupRouter.POST("/redirect/update", baseApi.UpdateRedirectConfig)
		groupRouter.POST("/redirect", baseApi.GetRedirectConfig)
		groupRouter.POST("/redirect/file", baseApi.UpdateRedirectConfigFile)
	}
}
