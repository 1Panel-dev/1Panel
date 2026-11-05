package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/gin-gonic/gin"
)

// @Tags PHP Extensions
// @Summary Page Extensions
// @Accept json
// @Param request body request.PHPExtensionsSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/php/extensions/search [post]
func (b *BaseApi) PagePHPExtensions(c *gin.Context) {
	var req request.PHPExtensionsSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if req.All {
		list, err := phpExtensionsService.List()
		if err != nil {
			helper.InternalServer(c, err)
			return
		}
		helper.SuccessWithData(c, dto.PageResult{
			Total: int64(len(list)),
			Items: list,
		})
	} else {
		total, list, err := phpExtensionsService.Page(req)
		if err != nil {
			helper.InternalServer(c, err)
			return
		}
		helper.SuccessWithData(c, dto.PageResult{
			Total: total,
			Items: list,
		})
	}

}

// @Tags PHP Extensions
// @Summary Create Extensions
// @Accept json
// @Param request body request.PHPExtensionsCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/php/extensions [post]
func (b *BaseApi) CreatePHPExtensions(c *gin.Context) {
	var req request.PHPExtensionsCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := phpExtensionsService.Create(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags PHP Extensions
// @Summary Update Extensions
// @Accept json
// @Param request body request.PHPExtensionsUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/php/extensions/update [post]
func (b *BaseApi) UpdatePHPExtensions(c *gin.Context) {
	var req request.PHPExtensionsUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := phpExtensionsService.Update(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags PHP Extensions
// @Summary Delete Extensions
// @Accept json
// @Param request body request.PHPExtensionsDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /runtimes/php/extensions/del [post]
func (b *BaseApi) DeletePHPExtensions(c *gin.Context) {
	var req request.PHPExtensionsDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := phpExtensionsService.Delete(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Router /runtimes/php/extensions/detail [post]
func (b *BaseApi) GetPHPExtensions(c *gin.Context) {
	var req request.PHPExtensionsDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	item, err := phpExtensionsService.Get(req.ID)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, item)
}

// @Router /runtimes/php/extensions/import [post]
// @x-panel-log {"bodyKeys":["name"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"同步 PHP 扩展模版 [name]","formatEN":"Sync PHP extension template [name]"}
func (b *BaseApi) ImportPHPExtensions(c *gin.Context) {
	var req request.PHPExtensionsCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	status, err := phpExtensionsService.Import(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, status)
}

// @Router /runtimes/php/extensions/push [post]
// @x-panel-log {"bodyKeys":["sourceNode","targetNodes"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"推送 PHP 扩展模版 [sourceNode] 至 [targetNodes]","formatEN":"Push PHP extension template from [sourceNode] to [targetNodes]"}
func (b *BaseApi) PushPHPExtensions(c *gin.Context) {
	var req dto.PHPExtensionsPush
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := phpExtensionsService.Push(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}
