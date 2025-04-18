package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/gin-gonic/gin"
)

// @Tags App
// @Summary List Upgrade Ignored App
// @Accept json
// @Success 200 {array} model.AppIgnoreUpgrade
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/ignored/detail [get]
func (b *BaseApi) ListAppIgnored(c *gin.Context) {
	res, err := appIgnoreUpgradeService.List()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags App
// @Summary Ignore Upgrade App
// @Accept json
// @Param request body request.AppIgnoreUpgradeReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/installed/ignore [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"忽略应用升级","formatEN":"Ignore application upgrade"}
func (b *BaseApi) IgnoreAppUpgrade(c *gin.Context) {
	var req request.AppIgnoreUpgradeReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := appIgnoreUpgradeService.CreateAppIgnore(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// 写一个去掉忽略的接口
// @Tags App
// @Summary Cancel Ignore Upgrade App
// @Accept json
// @Param request body request.ReqWithID true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /apps/ignored/cancel [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"取消忽略应用升级","formatEN":"Cancel ignore application upgrade"}
func (b *BaseApi) CancelIgnoreAppUpgrade(c *gin.Context) {
	var req request.ReqWithID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := appIgnoreUpgradeService.Delete(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}
