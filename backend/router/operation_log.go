package router

import (
	v1 "github.com/1Panel-dev/1Panel/app/api/v1"
	"github.com/1Panel-dev/1Panel/middleware"

	"github.com/gin-gonic/gin"
)

type OperationLogRouter struct{}

func (s *OperationLogRouter) InitOperationLogRouter(Router *gin.RouterGroup) {
	operationRouter := Router.Group("operations")
	operationRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		operationRouter.GET("", baseApi.GetOperationList)
	}
}
