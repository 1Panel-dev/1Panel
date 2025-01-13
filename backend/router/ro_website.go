package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type WebsiteRouter struct {
}

func (a *WebsiteRouter) InitRouter(Router *gin.RouterGroup) {
	websiteRouter := Router.Group("websites")
	websiteRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.PasswordExpired())

	baseApi := v1.ApiGroupApp.BaseApi
	{
		websiteRouter.POST("/search", baseApi.PageWebsite)
		websiteRouter.GET("/list", baseApi.GetWebsites)
		websiteRouter.POST("", baseApi.CreateWebsite)
		websiteRouter.POST("/operate", baseApi.OpWebsite)
		websiteRouter.POST("/log", baseApi.OpWebsiteLog)
		websiteRouter.POST("/check", baseApi.CreateWebsiteCheck)
		websiteRouter.GET("/options", baseApi.GetWebsiteOptions)
		websiteRouter.POST("/update", baseApi.UpdateWebsite)
		websiteRouter.GET("/:id", baseApi.GetWebsite)
		websiteRouter.POST("/del", baseApi.DeleteWebsite)
		websiteRouter.POST("/default/server", baseApi.ChangeDefaultServer)

		websiteRouter.GET("/domains/:websiteId", baseApi.GetWebDomains)
		websiteRouter.POST("/domains/del", baseApi.DeleteWebDomain)
		websiteRouter.POST("/domains", baseApi.CreateWebDomain)

		websiteRouter.GET("/:id/config/:type", baseApi.GetWebsiteNginx)
		websiteRouter.POST("/config", baseApi.GetNginxConfig)
		websiteRouter.POST("/config/update", baseApi.UpdateNginxConfig)
		websiteRouter.POST("/nginx/update", baseApi.UpdateWebsiteNginxConfig)

		websiteRouter.GET("/:id/https", baseApi.GetHTTPSConfig)
		websiteRouter.POST("/:id/https", baseApi.UpdateHTTPSConfig)

		websiteRouter.GET("/php/config/:id", baseApi.GetWebsitePHPConfig)
		websiteRouter.POST("/php/config", baseApi.UpdateWebsitePHPConfig)
		websiteRouter.POST("/php/update", baseApi.UpdatePHPFile)
		websiteRouter.POST("/php/version", baseApi.ChangePHPVersion)

		websiteRouter.POST("/rewrite", baseApi.GetRewriteConfig)
		websiteRouter.POST("/rewrite/update", baseApi.UpdateRewriteConfig)

		websiteRouter.POST("/dir/update", baseApi.UpdateSiteDir)
		websiteRouter.POST("/dir/permission", baseApi.UpdateSiteDirPermission)
		websiteRouter.POST("/dir", baseApi.GetDirConfig)

		websiteRouter.POST("/proxies", baseApi.GetProxyConfig)
		websiteRouter.POST("/proxies/update", baseApi.UpdateProxyConfig)
		websiteRouter.POST("/proxies/file", baseApi.UpdateProxyConfigFile)
		websiteRouter.POST("/proxies/del", baseApi.DeleteProxyConfig)

		websiteRouter.POST("/auths", baseApi.GetAuthConfig)
		websiteRouter.POST("/auths/update", baseApi.UpdateAuthConfig)

		websiteRouter.POST("/leech", baseApi.GetAntiLeech)
		websiteRouter.POST("/leech/update", baseApi.UpdateAntiLeech)

		websiteRouter.POST("/redirect/update", baseApi.UpdateRedirectConfig)
		websiteRouter.POST("/redirect", baseApi.GetRedirectConfig)
		websiteRouter.POST("/redirect/file", baseApi.UpdateRedirectConfigFile)

		websiteRouter.GET("/default/html/:type", baseApi.GetDefaultHtml)
		websiteRouter.POST("/default/html/update", baseApi.UpdateDefaultHtml)
	}
}
