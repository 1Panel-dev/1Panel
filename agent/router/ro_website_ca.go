package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type WebsiteCARouter struct {
}

func (a *WebsiteCARouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("websites/ca")

	baseApi := v2.ApiGroupApp.BaseApi
	{
		groupRouter.POST("/search", baseApi.PageWebsiteCA)
		groupRouter.POST("", baseApi.CreateWebsiteCA)
		groupRouter.POST("/del", baseApi.DeleteWebsiteCA)
		groupRouter.POST("/obtain", baseApi.ObtainWebsiteCA)
		groupRouter.POST("/renew", baseApi.RenewWebsiteCA)
		groupRouter.GET("/:id", baseApi.GetWebsiteCA)
		groupRouter.POST("/download", baseApi.DownloadCAFile)
	}
}
