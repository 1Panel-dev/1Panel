package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type NginxRouter struct {
}

func (a *NginxRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("openresty")
	groupRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.PasswordExpired())

	baseApi := v1.ApiGroupApp.BaseApi
	{
		groupRouter.GET("", baseApi.GetNginx)
		groupRouter.POST("/scope", baseApi.GetNginxConfigByScope)
		groupRouter.POST("/update", baseApi.UpdateNginxConfigByScope)
		groupRouter.GET("/status", baseApi.GetNginxStatus)
		groupRouter.POST("/file", baseApi.UpdateNginxFile)
		groupRouter.POST("/clear", baseApi.ClearNginxProxyCache)
	}
}
