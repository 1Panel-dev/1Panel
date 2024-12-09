package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Container Compose-template
// @Summary Create compose template
// @Accept json
// @Param request body dto.ComposeTemplateCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/template [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建 compose 模版 [name]","formatEN":"create compose template [name]"}
func (b *BaseApi) CreateComposeTemplate(c *gin.Context) {
	var req dto.ComposeTemplateCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := composeTemplateService.Create(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Container Compose-template
// @Summary Page compose templates
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Produce json
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/template/search [post]
func (b *BaseApi) SearchComposeTemplate(c *gin.Context) {
	var req dto.SearchWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := composeTemplateService.SearchWithPage(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Container Compose-template
// @Summary List compose templates
// @Produce json
// @Success 200 {array} dto.ComposeTemplateInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/template [get]
func (b *BaseApi) ListComposeTemplate(c *gin.Context) {
	list, err := composeTemplateService.List()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// @Tags Container Compose-template
// @Summary Delete compose template
// @Accept json
// @Param request body dto.BatchDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/template/del [post]
// @x-panel-log {"bodyKeys":["ids"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"ids","isList":true,"db":"compose_templates","output_column":"name","output_value":"names"}],"formatZH":"删除 compose 模版 [names]","formatEN":"delete compose template [names]"}
func (b *BaseApi) DeleteComposeTemplate(c *gin.Context) {
	var req dto.BatchDeleteReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := composeTemplateService.Delete(req.Ids); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Container Compose-template
// @Summary Update compose template
// @Accept json
// @Param request body dto.ComposeTemplateUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /containers/template/update [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"compose_templates","output_column":"name","output_value":"name"}],"formatZH":"更新 compose 模版 [name]","formatEN":"update compose template information [name]"}
func (b *BaseApi) UpdateComposeTemplate(c *gin.Context) {
	var req dto.ComposeTemplateUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	upMap := make(map[string]interface{})
	upMap["content"] = req.Content
	upMap["description"] = req.Description
	if err := composeTemplateService.Update(req.ID, upMap); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
