package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type ContainerRouter struct{}

func (s *ContainerRouter) InitRouter(Router *gin.RouterGroup) {
	baRouter := Router.Group("containers").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		baRouter.GET("/exec", baseApi.ContainerWsSsh)
		baRouter.GET("/stats/:id", baseApi.ContainerStats)

		baRouter.POST("", baseApi.ContainerCreate)
		baRouter.POST("/update", baseApi.ContainerUpdate)
		baRouter.POST("/upgrade", baseApi.ContainerUpgrade)
		baRouter.POST("/info", baseApi.ContainerInfo)
		baRouter.POST("/search", baseApi.SearchContainer)
		baRouter.POST("/list", baseApi.ListContainer)
		baRouter.GET("/list/stats", baseApi.ContainerListStats)
		baRouter.GET("/search/log", baseApi.ContainerLogs)
		baRouter.POST("/download/log", baseApi.DownloadContainerLogs)
		baRouter.GET("/limit", baseApi.LoadResourceLimit)
		baRouter.POST("/clean/log", baseApi.CleanContainerLog)
		baRouter.POST("/load/log", baseApi.LoadContainerLog)
		baRouter.POST("/inspect", baseApi.Inspect)
		baRouter.POST("/rename", baseApi.ContainerRename)
		baRouter.POST("/commit", baseApi.ContainerCommit)
		baRouter.POST("/operate", baseApi.ContainerOperation)
		baRouter.POST("/prune", baseApi.ContainerPrune)

		baRouter.GET("/repo", baseApi.ListRepo)
		baRouter.POST("/repo/status", baseApi.CheckRepoStatus)
		baRouter.POST("/repo/search", baseApi.SearchRepo)
		baRouter.POST("/repo/update", baseApi.UpdateRepo)
		baRouter.POST("/repo", baseApi.CreateRepo)
		baRouter.POST("/repo/del", baseApi.DeleteRepo)

		baRouter.POST("/compose/search", baseApi.SearchCompose)
		baRouter.POST("/compose", baseApi.CreateCompose)
		baRouter.POST("/compose/test", baseApi.TestCompose)
		baRouter.POST("/compose/operate", baseApi.OperatorCompose)
		baRouter.POST("/compose/update", baseApi.ComposeUpdate)
		baRouter.GET("/compose/search/log", baseApi.ComposeLogs)

		baRouter.GET("/template", baseApi.ListComposeTemplate)
		baRouter.POST("/template/search", baseApi.SearchComposeTemplate)
		baRouter.POST("/template/update", baseApi.UpdateComposeTemplate)
		baRouter.POST("/template", baseApi.CreateComposeTemplate)
		baRouter.POST("/template/del", baseApi.DeleteComposeTemplate)

		baRouter.GET("/image", baseApi.ListImage)
		baRouter.GET("/image/all", baseApi.ListAllImage)
		baRouter.POST("/image/search", baseApi.SearchImage)
		baRouter.POST("/image/pull", baseApi.ImagePull)
		baRouter.POST("/image/push", baseApi.ImagePush)
		baRouter.POST("/image/save", baseApi.ImageSave)
		baRouter.POST("/image/load", baseApi.ImageLoad)
		baRouter.POST("/image/remove", baseApi.ImageRemove)
		baRouter.POST("/image/tag", baseApi.ImageTag)
		baRouter.POST("/image/build", baseApi.ImageBuild)

		baRouter.GET("/network", baseApi.ListNetwork)
		baRouter.POST("/network/del", baseApi.DeleteNetwork)
		baRouter.POST("/network/search", baseApi.SearchNetwork)
		baRouter.POST("/network", baseApi.CreateNetwork)
		baRouter.GET("/volume", baseApi.ListVolume)
		baRouter.POST("/volume/del", baseApi.DeleteVolume)
		baRouter.POST("/volume/search", baseApi.SearchVolume)
		baRouter.POST("/volume", baseApi.CreateVolume)

		baRouter.GET("/daemonjson", baseApi.LoadDaemonJson)
		baRouter.GET("/daemonjson/file", baseApi.LoadDaemonJsonFile)
		baRouter.GET("/docker/status", baseApi.LoadDockerStatus)
		baRouter.POST("/docker/operate", baseApi.OperateDocker)
		baRouter.POST("/daemonjson/update", baseApi.UpdateDaemonJson)
		baRouter.POST("/logoption/update", baseApi.UpdateLogOption)
		baRouter.POST("/ipv6option/update", baseApi.UpdateIpv6Option)
		baRouter.POST("/daemonjson/update/byfile", baseApi.UpdateDaemonJsonByFile)
	}
}
