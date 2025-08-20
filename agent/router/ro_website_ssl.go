package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type WebsiteSSLRouter struct {
}

func (a *WebsiteSSLRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("websites/ssl")

	baseApi := v2.ApiGroupApp.BaseApi
	{
		groupRouter.POST("/search", baseApi.PageWebsiteSSL)
		groupRouter.POST("", baseApi.CreateWebsiteSSL)
		groupRouter.POST("/resolve", baseApi.GetDNSResolve)
		groupRouter.POST("/del", baseApi.DeleteWebsiteSSL)
		groupRouter.GET("/website/:websiteId", baseApi.GetWebsiteSSLByWebsiteId)
		groupRouter.GET("/:id", baseApi.GetWebsiteSSLById)
		groupRouter.POST("/update", baseApi.UpdateWebsiteSSL)
		groupRouter.POST("/upload", baseApi.UploadWebsiteSSL)
		groupRouter.POST("/obtain", baseApi.ApplyWebsiteSSL)
		groupRouter.POST("/download", baseApi.DownloadWebsiteSSL)
		groupRouter.POST("/import", baseApi.ImportMasterSSL)
	}
}
