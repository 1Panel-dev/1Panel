package router

import (
	v1 "github.com/1Panel-dev/1Panel/agent/app/api/v1"
	"github.com/gin-gonic/gin"
)

type NginxRouter struct {
}

func (a *NginxRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("openresty")

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
