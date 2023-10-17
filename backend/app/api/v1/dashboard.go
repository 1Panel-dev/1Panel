package v1

import (
	"errors"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Dashboard
// @Summary Load dashboard base info
// @Description 获取首页基础数据
// @Accept json
// @Param ioOption path string true "request"
// @Param netOption path string true "request"
// @Success 200 {object} dto.DashboardBase
// @Security ApiKeyAuth
// @Router /dashboard/base/:ioOption/:netOption [get]
func (b *BaseApi) LoadDashboardBaseInfo(c *gin.Context) {
	ioOption, ok := c.Params.Get("ioOption")
	if !ok {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, errors.New("error ioOption in path"))
		return
	}
	netOption, ok := c.Params.Get("netOption")
	if !ok {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, errors.New("error netOption in path"))
		return
	}
	data, err := dashboardService.LoadBaseInfo(ioOption, netOption)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Dashboard
// @Summary Load dashboard current info
// @Description 获取首页实时数据
// @Accept json
// @Param ioOption path string true "request"
// @Param netOption path string true "request"
// @Success 200 {object} dto.DashboardCurrent
// @Security ApiKeyAuth
// @Router /dashboard/current/:ioOption/:netOption [get]
func (b *BaseApi) LoadDashboardCurrentInfo(c *gin.Context) {
	ioOption, ok := c.Params.Get("ioOption")
	if !ok {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, errors.New("error ioOption in path"))
		return
	}
	netOption, ok := c.Params.Get("netOption")
	if !ok {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, errors.New("error netOption in path"))
		return
	}

	data := dashboardService.LoadCurrentInfo(ioOption, netOption)
	helper.SuccessWithData(c, data)
}

// @Tags Dashboard
// @Summary System restart
// @Description 重启服务器/面板
// @Accept json
// @Param operation path string true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /dashboard/system/restart/:operation [post]
func (b *BaseApi) SystemRestart(c *gin.Context) {
	operation, ok := c.Params.Get("operation")
	if !ok {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, errors.New("error operation in path"))
		return
	}
	if err := dashboardService.Restart(operation); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
