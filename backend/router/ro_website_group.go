package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type WebsiteGroupRouter struct {
}

func (a *WebsiteGroupRouter) InitWebsiteGroupRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("websites/groups")
	groupRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.PasswordExpired())

	baseApi := v1.ApiGroupApp.BaseApi
	{
		groupRouter.GET("", baseApi.GetWebGroups)
		groupRouter.POST("", baseApi.CreateWebGroup)
		groupRouter.POST("/update", baseApi.UpdateWebGroup)
		groupRouter.POST("/del", baseApi.DeleteWebGroup)
	}
}
