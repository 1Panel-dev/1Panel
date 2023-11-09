package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"

	"github.com/gin-gonic/gin"
)

type ToolboxRouter struct{}

func (s *ToolboxRouter) InitToolboxRouter(Router *gin.RouterGroup) {
	toolboxRouter := Router.Group("toolbox").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		toolboxRouter.GET("/fail2ban/base", baseApi.LoadFail2banBaseInfo)
		toolboxRouter.POST("/fail2ban/search", baseApi.SearchFail2ban)
		toolboxRouter.POST("/fail2ban/operate", baseApi.OperateFail2ban)
		toolboxRouter.POST("/fail2ban/operate/sshd", baseApi.OperateSSHD)
		toolboxRouter.POST("/fail2ban/update", baseApi.UpdateFail2banConf)
		toolboxRouter.POST("/fail2ban/update/byconf", baseApi.UpdateFail2banConfByFile)
	}
}
