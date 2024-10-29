package v2

import (
	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Logs
// @Summary Page login logs
// @Description 获取系统登录日志列表分页
// @Accept json
// @Param request body dto.SearchLgLogWithPage true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /core/logs/login [post]
func (b *BaseApi) GetLoginLogs(c *gin.Context) {
	var req dto.SearchLgLogWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := logService.PageLoginLog(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Logs
// @Summary Page operation logs
// @Description 获取系统操作日志列表分页
// @Accept json
// @Param request body dto.SearchOpLogWithPage true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /core/logs/operation [post]
func (b *BaseApi) GetOperationLogs(c *gin.Context) {
	var req dto.SearchOpLogWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := logService.PageOperationLog(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Logs
// @Summary Clean operation logs
// @Description 清空操作日志
// @Accept json
// @Param request body dto.CleanLog true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /core/logs/clean [post]
// @x-panel-log {"bodyKeys":["logType"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"清空 [logType] 日志信息","formatEN":"Clean the [logType] log information"}
func (b *BaseApi) CleanLogs(c *gin.Context) {
	var req dto.CleanLog
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := logService.CleanLogs(req.LogType); err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, nil)
}
