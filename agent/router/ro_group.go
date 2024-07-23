package router

import (
	v1 "github.com/1Panel-dev/1Panel/agent/app/api/v1"
	"github.com/gin-gonic/gin"
)

type WebsiteGroupRouter struct {
}

func (a *WebsiteGroupRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("groups")

	baseApi := v1.ApiGroupApp.BaseApi
	{
		groupRouter.POST("", baseApi.CreateGroup)
		groupRouter.POST("/del", baseApi.DeleteGroup)
		groupRouter.POST("/update", baseApi.UpdateGroup)
		groupRouter.POST("/search", baseApi.ListGroup)
	}
}
