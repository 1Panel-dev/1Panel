package v2

import (
	"errors"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Dashboard
// @Summary Load os info
// @Description 获取服务器基础数据
// @Accept json
// @Success 200 {object} dto.OsInfo
// @Security ApiKeyAuth
// @Router /dashboard/base/os [get]
func (b *BaseApi) LoadDashboardOsInfo(c *gin.Context) {
	data, err := dashboardService.LoadOsInfo()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Dashboard
// @Summary Load app launcher
// @Description 获取应用展示列表
// @Accept json
// @Success 200 {Array} dto.dto.AppLauncher
// @Security ApiKeyAuth
// @Router /dashboard/app/launcher [get]
func (b *BaseApi) LoadAppLauncher(c *gin.Context) {
	data, err := dashboardService.LoadAppLauncher()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Dashboard
// @Summary Load app launcher options
// @Description 获取应用展示选项
// @Accept json
// @Param request body dto.SearchByFilter true "request"
// @Success 200 {Array} dto.LauncherOption
// @Security ApiKeyAuth
// @Router /dashboard/app/launcher/option [post]
func (b *BaseApi) LoadAppLauncherOption(c *gin.Context) {
	var req dto.SearchByFilter
	if err := helper.CheckBind(&req, c); err != nil {
		return
	}
	data, err := dashboardService.ListLauncherOption(req.Filter)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}

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
		helper.BadRequest(c, errors.New("error ioOption in path"))
		return
	}
	netOption, ok := c.Params.Get("netOption")
	if !ok {
		helper.BadRequest(c, errors.New("error ioOption in path"))
		return
	}
	data, err := dashboardService.LoadBaseInfo(ioOption, netOption)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Dashboard
// @Summary Load dashboard current info for node
// @Description 获取节点实时数据
// @Success 200 {object} dto.NodeCurrent
// @Security ApiKeyAuth
// @Router /dashboard/current/node [get]
func (b *BaseApi) LoadCurrentInfoForNode(c *gin.Context) {
	data := dashboardService.LoadCurrentInfoForNode()
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
		helper.BadRequest(c, errors.New("error ioOption in path"))
		return
	}
	netOption, ok := c.Params.Get("netOption")
	if !ok {
		helper.BadRequest(c, errors.New("error netOption in path"))
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
		helper.BadRequest(c, errors.New("error operation in path"))
		return
	}
	if err := dashboardService.Restart(operation); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
