package v2

import (
	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/gin-gonic/gin"
)

// @Tags System Setting
// @Summary Load upgrade info
// @Success 200 {object} dto.UpgradeInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/upgrade [get]
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
// @Accept json
// @Param request body dto.Upgrade true "request"
// @Success 200 {string} notes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/upgrade [get]
func (b *BaseApi) GetNotesByVersion(c *gin.Context) {
	var req dto.Upgrade
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	notes, err := upgradeService.LoadNotes(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, notes)
}

// @Tags System Setting
// @Summary Upgrade
// @Accept json
// @Param request body dto.Upgrade true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/upgrade [post]
// @x-panel-log {"bodyKeys":["version"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新系统 => [version]","formatEN":"upgrade system => [version]"}
func (b *BaseApi) Upgrade(c *gin.Context) {
	var req dto.Upgrade
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := upgradeService.Upgrade(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}

// @Tags System Setting
// @Summary Upgrade
// @Accept json
// @Param request body dto.OperateByID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/rollback [post]
// @x-panel-log {"bodyKeys":["id"],"paramKeys":[],"BeforeFunctions":[{"input_column":"id","input_value":"id","isList":false,"db":"upgrade_logs","output_column":"old_version","output_value":"version"}],"formatZH":"回滚系统 => [version]","formatEN":"rollback system => [version]"}
func (b *BaseApi) Rollback(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := upgradeService.Rollback(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithOutData(c)
}
