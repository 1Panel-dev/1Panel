package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type MailRouter struct{}

func (a *MailRouter) InitRouter(Router *gin.RouterGroup) {
	mailRouter := Router.Group("mail")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		mailRouter.GET("/domains", baseApi.ListMailDomains)
		mailRouter.POST("/domains", baseApi.CreateMailDomain)
		mailRouter.POST("/domains/del", baseApi.DeleteMailDomain)

		mailRouter.POST("/accounts/search", baseApi.SearchMailAccounts)
		mailRouter.POST("/accounts", baseApi.CreateMailAccount)
		mailRouter.POST("/accounts/update", baseApi.UpdateMailAccount)
		mailRouter.POST("/accounts/del", baseApi.DeleteMailAccount)

		mailRouter.POST("/aliases/search", baseApi.SearchMailAliases)
		mailRouter.POST("/aliases", baseApi.CreateMailAlias)
		mailRouter.POST("/aliases/del", baseApi.DeleteMailAlias)

		mailRouter.POST("/setup", baseApi.SetupMail)
		mailRouter.GET("/status", baseApi.GetMailStatus)
	}
}
