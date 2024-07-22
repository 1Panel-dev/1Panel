package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"

	"github.com/gin-gonic/gin"
)

type TerminalRouter struct{}

func (s *TerminalRouter) InitRouter(Router *gin.RouterGroup) {
	terminalRouter := Router.Group("terminals")
	baseApi := v1.ApiGroupApp.BaseApi
	{
		terminalRouter.GET("", baseApi.WsSsh)
	}
}
