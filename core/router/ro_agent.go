package router

import (
	v2 "github.com/1Panel-dev/1Panel/core/app/api/v2"
	"github.com/gin-gonic/gin"
)

type AgentRouter struct{}

func (s *AgentRouter) InitRouter(Router *gin.RouterGroup) {
	baseApi := v2.ApiGroupApp.BaseApi
	{
		Router.POST("/backup", baseApi.GetBackup)
		Router.POST("/backup/list", baseApi.ListBackup)
	}
}
