package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/ai_tools/gpu"
	"github.com/1Panel-dev/1Panel/backend/utils/ai_tools/gpu/common"
	"github.com/1Panel-dev/1Panel/backend/utils/ai_tools/xpu"
	"github.com/gin-gonic/gin"
)

// @Tags AI
// @Summary Create Ollama model
// @Accept json
// @Param request body dto.OllamaModelName true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/ollama/model [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"添加 Ollama 模型 [name]","formatEN":"add Ollama model [name]"}
func (b *BaseApi) CreateOllamaModel(c *gin.Context) {
	var req dto.OllamaModelName
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := AIToolService.Create(req.Name); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags AI
// @Summary Rereate Ollama model
// @Accept json
// @Param request body dto.OllamaModelName true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/ollama/model/recreate [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"添加 Ollama 模型重试 [name]","formatEN":"re-add Ollama model [name]"}
func (b *BaseApi) RecreateOllamaModel(c *gin.Context) {
	var req dto.OllamaModelName
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := AIToolService.Recreate(req.Name); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags AI
// @Summary Close Ollama model conn
// @Accept json
// @Param request body dto.OllamaModelName true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/ollama/model/close [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"关闭 Ollama 模型连接 [name]","formatEN":"close conn for Ollama model [name]"}
func (b *BaseApi) CloseOllamaModel(c *gin.Context) {
	var req dto.OllamaModelName
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := AIToolService.Close(req.Name); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags AI
// @Summary Sync Ollama model list
// @Success 200 {array} dto.OllamaModelDropList
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/ollama/model/sync [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"同步 Ollama 模型列表","formatEN":"sync Ollama model list"}
func (b *BaseApi) SyncOllamaModel(c *gin.Context) {
	list, err := AIToolService.Sync()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, list)
}

// @Tags AI
// @Summary Page Ollama models
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/ollama/model/search [post]
func (b *BaseApi) SearchOllamaModel(c *gin.Context) {
	var req dto.SearchWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := AIToolService.Search(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags AI
// @Summary Page Ollama models
// @Accept json
// @Param request body dto.OllamaModelName true "request"
// @Success 200 {string} details
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/ollama/model/load [post]
func (b *BaseApi) LoadOllamaModelDetail(c *gin.Context) {
	var req dto.OllamaModelName
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	detail, err := AIToolService.LoadDetail(req.Name)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, detail)
}

// @Tags AI
// @Summary Delete Ollama model
// @Accept json
// @Param request body dto.ForceDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/ollama/model/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"ids","isList":true,"db":"ollama_models","output_column":"name","output_value":"names"}],"formatZH":"删除 Ollama 模型 [names]","formatEN":"remove Ollama model [names]"}
func (b *BaseApi) DeleteOllamaModel(c *gin.Context) {
	var req dto.ForceDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := AIToolService.Delete(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithOutData(c)
}

// @Tags AI
// @Summary Load gpu / xpu info
// @Accept json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/gpu/load [get]
func (b *BaseApi) LoadGpuInfo(c *gin.Context) {
	ok, client := gpu.New()
	if ok {
		info, err := client.LoadGpuInfo()
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
		helper.SuccessWithData(c, info)
		return
	}
	xpuOK, xpuClient := xpu.New()
	if xpuOK {
		info, err := xpuClient.LoadGpuInfo()
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
		helper.SuccessWithData(c, info)
		return
	}
	helper.SuccessWithData(c, &common.GpuInfo{})
}

// @Tags AI
// @Summary Bind domain
// @Accept json
// @Param request body dto.OllamaBindDomain true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/domain/bind [post]
func (b *BaseApi) BindDomain(c *gin.Context) {
	var req dto.OllamaBindDomain
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := AIToolService.BindDomain(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags AI
// @Summary Get bind domain
// @Accept json
// @Param request body dto.OllamaBindDomainReq true "request"
// @Success 200 {object} dto.OllamaBindDomainRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/domain/get [post]
func (b *BaseApi) GetBindDomain(c *gin.Context) {
	var req dto.OllamaBindDomainReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := AIToolService.GetBindDomain(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// Tags AI
// Summary Update bind domain
// Accept json
// Param request body dto.OllamaBindDomain true "request"
// Success 200
// Security ApiKeyAuth
// Security Timestamp
// Router /ai/domain/update [post]
func (b *BaseApi) UpdateBindDomain(c *gin.Context) {
	var req dto.OllamaBindDomain
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := AIToolService.UpdateBindDomain(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}
