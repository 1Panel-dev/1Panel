package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type UpgradeRouter struct{}

func (s *UpgradeRouter) InitRouter(Router *gin.RouterGroup) {
	upgradeRouter := Router.Group("upgrades")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		upgradeRouter.POST("/upgrade", baseApi.Upgrade)
		upgradeRouter.POST("/rollback", baseApi.Rollback)
	}
}
