package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type WebsiteSSLRouter struct {
}

func (a *WebsiteSSLRouter) InitWebsiteSSLRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("websites/ssl")
	groupRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth())

	baseApi := v1.ApiGroupApp.BaseApi
	{
		groupRouter.POST("/search", baseApi.PageWebsiteSSL)
		groupRouter.POST("", baseApi.CreateWebsiteSSL)
		groupRouter.POST("/apply", baseApi.ApplyWebsiteSSL)
		groupRouter.POST("/resolve", baseApi.GetDNSResolve)
		groupRouter.POST("/:id", baseApi.DeleteWebsiteSSL)
	}
}
