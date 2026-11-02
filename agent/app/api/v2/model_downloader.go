package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/gin-gonic/gin"
)

func handleModelDownloaderError(c *gin.Context, err error) {
	if service.IsModelDownloaderInputError(err) {
		helper.BadRequest(c, err)
		return
	}
	helper.InternalServer(c, err)
}

// @Tags AI
// @Summary Search Hugging Face models
// @Accept json
// @Param request body dto.ModelSearch true "request"
// @Produce json
// @Success 200 {object} dto.PageResult{items=[]dto.ModelItem}
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/model/downloader/hf/search [post]
func (b *BaseApi) SearchHFModels(c *gin.Context) {
	var req dto.ModelSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, items, err := modelDownloaderService.SearchHFModels(req)
	if err != nil {
		handleModelDownloaderError(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{Total: total, Items: items})
}

// @Tags AI
// @Summary Get Hugging Face model details
// @Accept json
// @Param request body dto.ModelInfoReq true "request"
// @Produce json
// @Success 200 {object} dto.ModelInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/model/downloader/hf/info [post]
func (b *BaseApi) GetHFModelInfo(c *gin.Context) {
	var req dto.ModelInfoReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	item, err := modelDownloaderService.GetHFModelInfo(req)
	if err != nil {
		handleModelDownloaderError(c, err)
		return
	}
	helper.SuccessWithData(c, item)
}

// @Tags AI
// @Summary Download Hugging Face model
// @Description Creates an asynchronous download task. Query /ai/model/downloader/tasks/search for progress.
// @Accept json
// @Param request body dto.ModelDownload true "request"
// @Produce json
// @Success 200 {object} dto.ModelDownloadTaskItem
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/model/downloader/hf/download [post]
// @x-panel-log {"bodyKeys":["repoID","revision"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"下载 Hugging Face 模型 [repoID]","formatEN":"download Hugging Face model [repoID]"}
func (b *BaseApi) StartHFModelDownload(c *gin.Context) {
	var req dto.ModelDownload
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	task, err := modelDownloaderService.StartDownload(req)
	if err != nil {
		handleModelDownloaderError(c, err)
		return
	}
	helper.SuccessWithData(c, task)
}

// @Tags AI
// @Summary Search ModelScope models
// @Accept json
// @Param request body dto.ModelSearch true "request"
// @Produce json
// @Success 200 {object} dto.PageResult{items=[]dto.ModelItem}
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/model/downloader/modelscope/search [post]
func (b *BaseApi) SearchModelScopeModels(c *gin.Context) {
	var req dto.ModelSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, items, err := modelDownloaderService.SearchModelScopeModels(req)
	if err != nil {
		handleModelDownloaderError(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{Total: total, Items: items})
}

// @Tags AI
// @Summary Get ModelScope model details
// @Accept json
// @Param request body dto.ModelInfoReq true "request"
// @Produce json
// @Success 200 {object} dto.ModelInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/model/downloader/modelscope/info [post]
func (b *BaseApi) GetModelScopeModelInfo(c *gin.Context) {
	var req dto.ModelInfoReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	item, err := modelDownloaderService.GetModelScopeModelInfo(req)
	if err != nil {
		handleModelDownloaderError(c, err)
		return
	}
	helper.SuccessWithData(c, item)
}

// @Tags AI
// @Summary Download ModelScope model
// @Description Creates an asynchronous download task. Query /ai/model/downloader/tasks/search for progress.
// @Accept json
// @Param request body dto.ModelDownload true "request"
// @Produce json
// @Success 200 {object} dto.ModelDownloadTaskItem
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/model/downloader/modelscope/download [post]
// @x-panel-log {"bodyKeys":["repoID","revision"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"下载 ModelScope 模型 [repoID]","formatEN":"download ModelScope model [repoID]"}
func (b *BaseApi) StartModelScopeModelDownload(c *gin.Context) {
	var req dto.ModelDownload
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	task, err := modelDownloaderService.StartModelScopeDownload(req)
	if err != nil {
		handleModelDownloaderError(c, err)
		return
	}
	helper.SuccessWithData(c, task)
}

// @Tags AI
// @Summary Get model downloader settings
// @Produce json
// @Success 200 {object} dto.ModelDownloaderSetting
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/model/downloader/settings [get]
func (b *BaseApi) GetModelDownloaderSettings(c *gin.Context) {
	settings, err := modelDownloaderService.GetSettings()
	if err != nil {
		handleModelDownloaderError(c, err)
		return
	}
	helper.SuccessWithData(c, settings)
}

// @Tags AI
// @Summary Update model downloader settings
// @Description Updates the download directory, source endpoints and tokens. The featuredModelsSync field is ignored; use /ai/model/downloader/featured/sync/status to update automatic synchronization.
// @Accept json
// @Param request body dto.ModelDownloaderSetting true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/model/downloader/settings [post]
// @x-panel-log {"bodyKeys":["modelDir"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新模型下载设置 [modelDir]","formatEN":"update model downloader settings [modelDir]"}
func (b *BaseApi) UpdateModelDownloaderSettings(c *gin.Context) {
	var req dto.ModelDownloaderSetting
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := modelDownloaderService.UpdateSettings(req); err != nil {
		handleModelDownloaderError(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Page model download tasks
// @Description Set includeLocal to true to include locally downloaded models.
// @Accept json
// @Param request body dto.ModelDownloadTaskSearch true "request"
// @Produce json
// @Success 200 {object} dto.PageResult{items=[]dto.ModelDownloadTaskItem}
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/model/downloader/tasks/search [post]
func (b *BaseApi) PageModelDownloadTasks(c *gin.Context) {
	var req dto.ModelDownloadTaskSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, items, err := modelDownloaderService.PageTasks(req)
	if err != nil {
		handleModelDownloaderError(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{Items: items, Total: total})
}

// @Tags AI
// @Summary Cancel model download task
// @Description Only waiting or downloading tasks can be canceled. The withFile field is ignored.
// @Accept json
// @Param request body dto.ModelDownloadTaskOperate true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/model/downloader/tasks/cancel [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"model_download_tasks","output_column":"repo_id","output_value":"repoID"}],"formatZH":"取消模型下载任务 [repoID]","formatEN":"cancel model download task [repoID]"}
func (b *BaseApi) CancelModelDownloadTask(c *gin.Context) {
	var req dto.ModelDownloadTaskOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := modelDownloaderService.CancelTask(req); err != nil {
		handleModelDownloaderError(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Retry model download task
// @Description Only failed or canceled tasks can be retried. Returns the task while the download runs asynchronously.
// @Accept json
// @Param request body dto.ModelDownloadTaskRetry true "request"
// @Produce json
// @Success 200 {object} dto.ModelDownloadTaskItem
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/model/downloader/tasks/retry [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"model_download_tasks","output_column":"repo_id","output_value":"repoID"}],"formatZH":"重试模型下载任务 [repoID]","formatEN":"retry model download task [repoID]"}
func (b *BaseApi) RetryModelDownloadTask(c *gin.Context) {
	var req dto.ModelDownloadTaskRetry
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	task, err := modelDownloaderService.RetryTask(req)
	if err != nil {
		handleModelDownloaderError(c, err)
		return
	}
	helper.SuccessWithData(c, task)
}

// @Tags AI
// @Summary Remove model download task
// @Description Waiting or downloading tasks cannot be removed. Set withFile to true to also delete downloaded files.
// @Accept json
// @Param request body dto.ModelDownloadTaskOperate true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/model/downloader/tasks/remove [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"model_download_tasks","output_column":"repo_id","output_value":"repoID"}],"formatZH":"移除模型下载任务 [repoID]","formatEN":"remove model download task [repoID]"}
func (b *BaseApi) RemoveModelDownloadTask(c *gin.Context) {
	var req dto.ModelDownloadTaskOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := modelDownloaderService.RemoveTask(req); err != nil {
		handleModelDownloaderError(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Page locally downloaded models
// @Accept json
// @Param request body dto.LocalModelSearch true "request"
// @Produce json
// @Success 200 {object} dto.PageResult{items=[]dto.LocalModelItem}
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/model/downloader/local/search [post]
func (b *BaseApi) PageLocalDownloadedModels(c *gin.Context) {
	var req dto.LocalModelSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, items, err := modelDownloaderService.ListLocalModels(req)
	if err != nil {
		handleModelDownloaderError(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{Items: items, Total: total})
}

// @Tags AI
// @Summary Delete locally downloaded model
// @Accept json
// @Param request body dto.LocalModelDelete true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/model/downloader/local/delete [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"删除本地模型 [name]","formatEN":"delete local model [name]"}
func (b *BaseApi) DeleteLocalDownloadedModel(c *gin.Context) {
	var req dto.LocalModelDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := modelDownloaderService.DeleteLocalModel(req); err != nil {
		handleModelDownloaderError(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Synchronize featured models
// @Description Synchronizes the featured model catalog immediately, independently of the automatic synchronization setting. Does nothing in offline mode.
// @Accept json
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/model/downloader/featured/sync [post]
func (b *BaseApi) SyncFeaturedModels(c *gin.Context) {
	if err := modelDownloaderService.SyncFeaturedModels(); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Page featured models
// @Accept json
// @Param request body dto.FeaturedModelSearch true "request"
// @Produce json
// @Success 200 {object} dto.FeaturedModelPage{items=[]dto.FeaturedModel}
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/model/downloader/featured/search [post]
func (b *BaseApi) PageFeaturedModels(c *gin.Context) {
	var req dto.FeaturedModelSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	result, err := modelDownloaderService.PageFeaturedModels(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

// @Tags AI
// @Summary Update featured model automatic synchronization
// @Description Set value to Enable or Disable. This updates the automatic synchronization setting without triggering an immediate synchronization.
// @Accept json
// @Param request body dto.ModelDownloaderSyncSetting true "request"
// @Produce json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/model/downloader/featured/sync/status [post]
// @x-panel-log {"bodyKeys":["value"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新推荐模型自动同步 [value]","formatEN":"update featured model automatic synchronization [value]"}
func (b *BaseApi) UpdateFeaturedModelsSync(c *gin.Context) {
	var req dto.ModelDownloaderSyncSetting
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := modelDownloaderService.UpdateFeaturedModelsSync(req); err != nil {
		handleModelDownloaderError(c, err)
		return
	}
	helper.Success(c)
}
