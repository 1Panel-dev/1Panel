package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type TerminalRouter struct{}

func (s *TerminalRouter) InitRouter(Router *gin.RouterGroup) {
	terminalRouter := Router.Group("terminals")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		terminalRouter.GET("", baseApi.WsSsh)
	}
}
