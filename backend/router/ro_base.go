package router

import (
	v1 "github.com/1Panel-dev/1Panel/app/api/v1"
	"github.com/gin-gonic/gin"
)

type BaseRouter struct{}

func (s *BaseRouter) InitBaseRouter(Router *gin.RouterGroup) (R gin.IRoutes) {
	baseRouter := Router.Group("base")
	baseApi := v1.ApiGroupApp.BaseApi
	{
		baseRouter.POST("login", baseApi.Login)
	}
	return baseRouter
}
