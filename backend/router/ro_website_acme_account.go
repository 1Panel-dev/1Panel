package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type WebsiteAcmeAccountRouter struct {
}

func (a *WebsiteAcmeAccountRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("websites/acme")
	groupRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.PasswordExpired())

	baseApi := v1.ApiGroupApp.BaseApi
	{
		groupRouter.POST("/search", baseApi.PageWebsiteAcmeAccount)
		groupRouter.POST("", baseApi.CreateWebsiteAcmeAccount)
		groupRouter.POST("/del", baseApi.DeleteWebsiteAcmeAccount)
	}
}
