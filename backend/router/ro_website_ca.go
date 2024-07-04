package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type WebsiteCARouter struct {
}

func (a *WebsiteCARouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("websites/ca")
	groupRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.PasswordExpired())

	baseApi := v1.ApiGroupApp.BaseApi
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
