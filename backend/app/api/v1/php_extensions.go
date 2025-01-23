package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/constant"
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
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
		helper.SuccessWithData(c, dto.PageResult{
			Total: int64(len(list)),
			Items: list,
		})
	} else {
		total, list, err := phpExtensionsService.Page(req)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithOutData(c)
}
