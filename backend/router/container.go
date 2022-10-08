package router

import (
	v1 "github.com/1Panel-dev/1Panel/app/api/v1"
	"github.com/1Panel-dev/1Panel/middleware"
	"github.com/gin-gonic/gin"
)

type ContainerRouter struct{}

func (s *ContainerRouter) InitContainerRouter(Router *gin.RouterGroup) {
	baRouter := Router.Group("containers").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	withRecordRouter := Router.Group("containers").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired()).
		Use(middleware.OperationRecord())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		baRouter.POST("/search", baseApi.SearchContainer)
		baRouter.GET("/detail/:id", baseApi.ContainerDetail)
		withRecordRouter.POST("operate", baseApi.ContainerOperation)
		withRecordRouter.POST("/log", baseApi.ContainerLogs)
	}
}
