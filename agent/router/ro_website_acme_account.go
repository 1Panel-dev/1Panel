package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type WebsiteAcmeAccountRouter struct {
}

func (a *WebsiteAcmeAccountRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("websites/acme")

	baseApi := v2.ApiGroupApp.BaseApi
	{
		groupRouter.POST("/search", baseApi.PageWebsiteAcmeAccount)
		groupRouter.POST("", baseApi.CreateWebsiteAcmeAccount)
		groupRouter.POST("/del", baseApi.DeleteWebsiteAcmeAccount)
		groupRouter.POST("/update", baseApi.UpdateWebsiteAcmeAccount)
	}
}
