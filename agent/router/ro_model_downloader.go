package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type ModelDownloaderRouter struct{}

func (m *ModelDownloaderRouter) InitRouter(Router *gin.RouterGroup) {
	baseApi := v2.ApiGroupApp.BaseApi
	for _, prefix := range []string{"ai/model/downloader", "xpack/model/downloader"} {
		router := Router.Group(prefix)
		router.POST("/featured/search", baseApi.PageFeaturedModels)
		router.POST("/featured/sync", baseApi.SyncFeaturedModels)
		router.POST("/featured/sync/status", baseApi.UpdateFeaturedModelsSync)
		router.POST("/hf/search", baseApi.SearchHFModels)
		router.POST("/hf/info", baseApi.GetHFModelInfo)
		router.POST("/hf/download", baseApi.StartHFModelDownload)
		router.POST("/modelscope/search", baseApi.SearchModelScopeModels)
		router.POST("/modelscope/info", baseApi.GetModelScopeModelInfo)
		router.POST("/modelscope/download", baseApi.StartModelScopeModelDownload)
		router.GET("/settings", baseApi.GetModelDownloaderSettings)
		router.POST("/settings", baseApi.UpdateModelDownloaderSettings)
		router.POST("/tasks/search", baseApi.PageModelDownloadTasks)
		router.POST("/tasks/cancel", baseApi.CancelModelDownloadTask)
		router.POST("/tasks/retry", baseApi.RetryModelDownloadTask)
		router.POST("/tasks/remove", baseApi.RemoveModelDownloadTask)
		router.POST("/local/search", baseApi.PageLocalDownloadedModels)
		router.POST("/local/delete", baseApi.DeleteLocalDownloadedModel)
	}
}
