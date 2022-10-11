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

		baRouter.POST("/repo/search", baseApi.SearchRepo)
		baRouter.PUT("/repo/:id", baseApi.UpdateRepo)
		baRouter.GET("/repo", baseApi.ListRepo)
		withRecordRouter.POST("/repo", baseApi.CreateRepo)
		withRecordRouter.POST("/repo/del", baseApi.DeleteRepo)

		baRouter.POST("/image/search", baseApi.SearchImage)
		baRouter.POST("/image/pull", baseApi.ImagePull)
		baRouter.POST("/image/push", baseApi.ImagePush)
		baRouter.POST("/image/save", baseApi.ImageSave)
		baRouter.POST("/image/load", baseApi.ImageLoad)
		baRouter.POST("/image/remove", baseApi.ImageRemove)

		baRouter.POST("/network/del", baseApi.DeleteNetwork)
		baRouter.POST("/network/search", baseApi.SearchNetwork)
		baRouter.POST("/network", baseApi.CreateNetwork)
		baRouter.POST("/volume/del", baseApi.DeleteVolume)
		baRouter.POST("/volume/search", baseApi.SearchVolume)
		baRouter.POST("/volume", baseApi.CreateVolume)
	}
}
