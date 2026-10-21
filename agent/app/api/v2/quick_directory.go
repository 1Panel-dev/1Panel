package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/gin-gonic/gin"
)

// @Tags Dashboard
// @Summary List quick directories
// @Success 200 {array} model.QuickDirectory
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dashboard/quick/directory [get]
func (b *BaseApi) ListQuickDirectories(c *gin.Context) {
	items, err := dashboardService.ListQuickDirectories()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, items)
}

// @Tags Dashboard
// @Summary Create quick directory
// @Accept json
// @Param request body dto.QuickDirectoryCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dashboard/quick/directory [post]
func (b *BaseApi) CreateQuickDirectory(c *gin.Context) {
	var req dto.QuickDirectoryCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	item, err := dashboardService.CreateQuickDirectory(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, item)
}

// @Tags Dashboard
// @Summary Update quick directory
// @Accept json
// @Param request body dto.QuickDirectoryUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dashboard/quick/directory/update [post]
func (b *BaseApi) UpdateQuickDirectory(c *gin.Context) {
	var req dto.QuickDirectoryUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	item, err := dashboardService.UpdateQuickDirectory(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, item)
}

// @Tags Dashboard
// @Summary Delete quick directory
// @Accept json
// @Param request body dto.QuickDirectoryDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dashboard/quick/directory/del [post]
func (b *BaseApi) DeleteQuickDirectory(c *gin.Context) {
	var req dto.QuickDirectoryDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := dashboardService.DeleteQuickDirectory(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}
