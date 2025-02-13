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

// @Tags AITools
// @Summary Create Ollama model
// @Accept json
// @Param request body dto.OllamaModelName true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /aitools/ollama/model [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"添加模型 [name]","formatEN":"add Ollama model [name]"}
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

// @Tags AITools
// @Summary Page Ollama models
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /aitools/ollama/model/search [post]
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

// @Tags AITools
// @Summary Page Ollama models
// @Accept json
// @Param request body dto.OllamaModelName true "request"
// @Success 200 {string} details
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /aitools/ollama/model/load [post]
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

// @Tags AITools
// @Summary Delete Ollama model
// @Accept json
// @Param request body dto.OllamaModelName true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /aitool/ollama/model/del [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"删除模型 [name]","formatEN":"remove Ollama model [name]"}
func (b *BaseApi) DeleteOllamaModel(c *gin.Context) {
	var req dto.OllamaModelName
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := AIToolService.Delete(req.Name); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithOutData(c)
}

// @Tags AITools
// @Summary Load gpu / xpu info
// @Accept json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /aitool/gpu/load [get]
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
