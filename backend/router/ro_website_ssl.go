package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type WebsiteSSLRouter struct {
}

func (a *WebsiteSSLRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("websites/ssl")
	groupRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.PasswordExpired())

	baseApi := v1.ApiGroupApp.BaseApi
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
	}
}
