package router

import (
	v1 "github.com/1Panel-dev/1Panel/app/api/v1"
	"github.com/1Panel-dev/1Panel/middleware"

	"github.com/gin-gonic/gin"
)

type TerminalRouter struct{}

func (s *UserRouter) InitTerminalRouter(Router *gin.RouterGroup) {
	terminalRouter := Router.Group("terminals")
	withRecordRouter := terminalRouter.Use(middleware.OperationRecord())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		withRecordRouter.GET("", baseApi.WsSsh)
	}
}
