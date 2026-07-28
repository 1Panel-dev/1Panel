package v2

import (
	"io"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/gin-gonic/gin"
)

// @Tags Website Template
// @Summary Page website templates
// @Accept json
// @Param request body request.WebsiteTemplateSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/templates/search [post]
func (b *BaseApi) PageWebsiteTemplate(c *gin.Context) {
	var req request.WebsiteTemplateSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, templates, err := websiteTemplateService.PageTemplate(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Total: total,
		Items: templates,
	})
}

// @Tags Website Template
// @Summary Create website template
// @Accept json
// @Param request body request.WebsiteTemplateCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/templates [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建网站模板 [name]","formatEN":"Create website template [name]"}
func (b *BaseApi) CreateWebsiteTemplate(c *gin.Context) {
	var req request.WebsiteTemplateCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteTemplateService.CreateTemplate(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Website Template
// @Summary Update website template
// @Accept json
// @Param request body request.WebsiteTemplateUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/templates/update [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新网站模板 [name]","formatEN":"Update website template [name]"}
func (b *BaseApi) UpdateWebsiteTemplate(c *gin.Context) {
	var req request.WebsiteTemplateUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteTemplateService.UpdateTemplate(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Website Template
// @Summary Delete website template
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/templates/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"website_templates","output_column":"name","output_value":"name"}],"formatZH":"删除网站模板 [name]","formatEN":"Delete website template [name]"}
func (b *BaseApi) DeleteWebsiteTemplate(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteTemplateService.DeleteTemplate(req.ID); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Website Template
// @Summary Get website template
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200 {object} response.WebsiteTemplateDTO
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/templates/get [post]
func (b *BaseApi) GetWebsiteTemplate(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	template, err := websiteTemplateService.GetTemplate(req.ID)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, template)
}

// @Tags Website Template
// @Summary Upload website template zip
// @Accept multipart/form-data
// @Param file formData file true "file"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/templates/upload [post]
func (b *BaseApi) UploadTemplateZip(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	filePath, variables, err := websiteTemplateService.SaveUploadZip(fileHeader.Filename, content)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, gin.H{"filePath": filePath, "variables": variables})
}

// @Tags Website Template
// @Summary Preview website template
// @Accept json
// @Param request body request.WebsitePreviewReq true "request"
// @Success 200 {object} response.WebsitePreviewDTO
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/templates/preview [post]
func (b *BaseApi) PreviewWebsiteTemplate(c *gin.Context) {
	var req request.WebsitePreviewReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	preview, err := websiteTemplateService.Preview(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, preview)
}

// @Tags Website Template
// @Summary Page website template outputs
// @Accept json
// @Param request body request.WebsiteTemplateOutputSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/templates/outputs/search [post]
func (b *BaseApi) PageWebsiteTemplateOutput(c *gin.Context) {
	var req request.WebsiteTemplateOutputSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, outputs, err := websiteTemplateService.PageOutput(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Total: total,
		Items: outputs,
	})
}

// @Tags Website Template
// @Summary Create website template output
// @Accept json
// @Param request body request.WebsiteTemplateOutputCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/templates/outputs [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"生成模板产物 [name]","formatEN":"Generate template output [name]"}
func (b *BaseApi) CreateWebsiteTemplateOutput(c *gin.Context) {
	var req request.WebsiteTemplateOutputCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteTemplateService.CreateOutput(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Website Template
// @Summary Delete website template output
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/templates/outputs/del [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"website_template_outputs","output_column":"name","output_value":"name"}],"formatZH":"删除模板产物 [name]","formatEN":"Delete template output [name]"}
func (b *BaseApi) DeleteWebsiteTemplateOutput(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteTemplateService.DeleteOutput(req.ID); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Website Template
// @Summary Get website template output
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200 {object} response.WebsiteTemplateOutputDTO
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/templates/outputs/get [post]
func (b *BaseApi) GetWebsiteTemplateOutput(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	output, err := websiteTemplateService.GetOutput(req.ID)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, output)
}

// @Tags Website Template
// @Summary Import website template output to site dir
// @Accept json
// @Param request body request.WebsiteTemplateImportReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /websites/templates/outputs/import [post]
// @x-panel-log {"bodyKeys":["outputID"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"导入模板产物 [outputID]","formatEN":"Import template output [outputID]"}
func (b *BaseApi) ImportTemplateOutput(c *gin.Context) {
	var req request.WebsiteTemplateImportReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := websiteTemplateService.ImportToSite(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}
