package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type WebsiteDnsAccountRouter struct {
}

func (a *WebsiteDnsAccountRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("websites/dns")

	baseApi := v2.ApiGroupApp.BaseApi
	{
		groupRouter.POST("/search", baseApi.PageWebsiteDnsAccount)
		groupRouter.POST("", baseApi.CreateWebsiteDnsAccount)
		groupRouter.POST("/update", baseApi.UpdateWebsiteDnsAccount)
		groupRouter.POST("/del", baseApi.DeleteWebsiteDnsAccount)
	}
}
