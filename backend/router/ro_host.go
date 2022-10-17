package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"

	"github.com/gin-gonic/gin"
)

type HostRouter struct{}

func (s *HostRouter) InitHostRouter(Router *gin.RouterGroup) {
	hostRouter := Router.Group("hosts").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	withRecordRouter := Router.Group("hosts").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired()).
		Use(middleware.OperationRecord())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		withRecordRouter.POST("", baseApi.CreateHost)
		withRecordRouter.DELETE(":id", baseApi.DeleteHost)
		withRecordRouter.PUT(":id", baseApi.UpdateHost)
		hostRouter.POST("/search", baseApi.HostTree)
		hostRouter.POST("/testconn", baseApi.TestConn)
		hostRouter.GET(":id", baseApi.GetHostInfo)
	}
}
