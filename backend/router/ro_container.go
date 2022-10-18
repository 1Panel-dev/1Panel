package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
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
		baRouter.GET("/exec", baseApi.ContainerExec)

		baRouter.POST("/search", baseApi.SearchContainer)
		baRouter.POST("/inspect", baseApi.Inspect)
		baRouter.POST("", baseApi.ContainerCreate)
		withRecordRouter.POST("operate", baseApi.ContainerOperation)
		withRecordRouter.POST("/log", baseApi.ContainerLogs)
		withRecordRouter.GET("/stats/:id", baseApi.ContainerStats)

		baRouter.POST("/repo/search", baseApi.SearchRepo)
		baRouter.PUT("/repo/:id", baseApi.UpdateRepo)
		baRouter.GET("/repo", baseApi.ListRepo)
		withRecordRouter.POST("/repo", baseApi.CreateRepo)
		withRecordRouter.POST("/repo/del", baseApi.DeleteRepo)

		baRouter.POST("/compose/search", baseApi.SearchCompose)
		baRouter.POST("/compose/up", baseApi.CreateCompose)
		baRouter.POST("/compose/operate", baseApi.OperatorCompose)

		baRouter.POST("/template/search", baseApi.SearchComposeTemplate)
		baRouter.PUT("/template/:id", baseApi.UpdateComposeTemplate)
		baRouter.GET("/template", baseApi.ListComposeTemplate)
		withRecordRouter.POST("/template", baseApi.CreateComposeTemplate)
		withRecordRouter.POST("/template/del", baseApi.DeleteComposeTemplate)

		baRouter.POST("/image/search", baseApi.SearchImage)
		baRouter.GET("/image", baseApi.ListImage)
		baRouter.POST("/image/pull", baseApi.ImagePull)
		baRouter.POST("/image/push", baseApi.ImagePush)
		baRouter.POST("/image/save", baseApi.ImageSave)
		baRouter.POST("/image/load", baseApi.ImageLoad)
		baRouter.POST("/image/remove", baseApi.ImageRemove)
		baRouter.POST("/image/tag", baseApi.ImageTag)
		baRouter.POST("/image/build", baseApi.ImageBuild)

		baRouter.POST("/network/del", baseApi.DeleteNetwork)
		baRouter.POST("/network/search", baseApi.SearchNetwork)
		baRouter.POST("/network", baseApi.CreateNetwork)
		baRouter.POST("/volume/del", baseApi.DeleteVolume)
		baRouter.POST("/volume/search", baseApi.SearchVolume)
		baRouter.GET("/volume", baseApi.ListVolume)
		baRouter.POST("/volume", baseApi.CreateVolume)
	}
}
