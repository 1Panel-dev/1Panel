package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type WebsiteDnsAccountRouter struct {
}

func (a *WebsiteDnsAccountRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("websites/dns")
	groupRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.PasswordExpired())

	baseApi := v1.ApiGroupApp.BaseApi
	{
		groupRouter.POST("/search", baseApi.PageWebsiteDnsAccount)
		groupRouter.POST("", baseApi.CreateWebsiteDnsAccount)
		groupRouter.POST("/update", baseApi.UpdateWebsiteDnsAccount)
		groupRouter.POST("/del", baseApi.DeleteWebsiteDnsAccount)
	}
}
