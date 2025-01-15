package v2

import (
	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) SearchAppLauncher(c *gin.Context) {
	data, err := appLauncherService.Search()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags App Launcher
// @Summary Update app Launcher
// @Accept json
// @Param request body dto.ChangeShow true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /core/app/launcher/show [post]
// @x-panel-log {"bodyKeys":["key", "value"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"首页应用 [key] => 显示：[value]","formatEN":"app launcher [key] => show: [value]"}
func (b *BaseApi) UpdateAppLauncher(c *gin.Context) {
	var req dto.SettingUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := appLauncherService.ChangeShow(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
