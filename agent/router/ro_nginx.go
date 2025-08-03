package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type NginxRouter struct {
}

func (a *NginxRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("openresty")

	baseApi := v2.ApiGroupApp.BaseApi
	{
		groupRouter.GET("", baseApi.GetNginx)
		groupRouter.POST("/scope", baseApi.GetNginxConfigByScope)
		groupRouter.POST("/update", baseApi.UpdateNginxConfigByScope)
		groupRouter.GET("/status", baseApi.GetNginxStatus)
		groupRouter.POST("/file", baseApi.UpdateNginxFile)
		groupRouter.POST("/build", baseApi.BuildNginx)
		groupRouter.POST("/modules/update", baseApi.UpdateNginxModule)
		groupRouter.GET("/modules", baseApi.GetNginxModules)
		groupRouter.POST("/https", baseApi.OperateDefaultHTTPs)
		groupRouter.GET("/https", baseApi.GetDefaultHTTPsStatus)
	}
}
