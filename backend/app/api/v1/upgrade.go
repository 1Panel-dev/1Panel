package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags System Setting
// @Summary Load upgrade info
// @Description 系统更新信息
// @Success 200 {object} dto.UpgradeInfo
// @Security ApiKeyAuth
// @Router /settings/upgrade [get]
func (b *BaseApi) GetUpgradeInfo(c *gin.Context) {
	info, err := upgradeService.SearchUpgrade()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, info)
}

// @Tags System Setting
// @Summary Load release notes by version
// @Description 获取版本 release notes
// @Accept json
// @Param request body dto.Upgrade true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/upgrade [get]
func (b *BaseApi) GetNotesByVersion(c *gin.Context) {
	var req dto.Upgrade
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	notes, err := upgradeService.LoadNotes(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, notes)
}

// @Tags System Setting
// @Summary Upgrade
// @Description 系统更新
// @Accept json
// @Param request body dto.Upgrade true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/upgrade [post]
// @x-panel-log {"bodyKeys":["version"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新系统 => [version]","formatEN":"upgrade system => [version]"}
func (b *BaseApi) Upgrade(c *gin.Context) {
	var req dto.Upgrade
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := upgradeService.Upgrade(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
