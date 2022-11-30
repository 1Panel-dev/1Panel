package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type WebsiteRouter struct {
}

func (a *WebsiteRouter) InitWebsiteRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("websites")
	groupRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth())

	baseApi := v1.ApiGroupApp.BaseApi
	{
		groupRouter.POST("", baseApi.CreateWebsite)
		groupRouter.GET("/options", baseApi.GetWebsiteOptions)
		groupRouter.POST("/backup/:domain", baseApi.BackupWebsite)
		groupRouter.POST("/recover", baseApi.RecoverWebsite)
		groupRouter.POST("/update", baseApi.UpdateWebSite)
		groupRouter.GET("/:id", baseApi.GetWebSite)
		groupRouter.GET("/:id/nginx", baseApi.GetWebSiteNginx)
		groupRouter.POST("/search", baseApi.PageWebsite)
		groupRouter.POST("/del", baseApi.DeleteWebSite)
		groupRouter.GET("/domains/:websiteId", baseApi.GetWebDomains)
		groupRouter.DELETE("/domains/:id", baseApi.DeleteWebDomain)
		groupRouter.POST("/domains", baseApi.CreateWebDomain)
		groupRouter.POST("/config", baseApi.GetNginxConfig)
		groupRouter.POST("/config/update", baseApi.UpdateNginxConfig)
		groupRouter.GET("/:id/https", baseApi.GetHTTPSConfig)
		groupRouter.POST("/:id/https", baseApi.UpdateHTTPSConfig)
	}
}
