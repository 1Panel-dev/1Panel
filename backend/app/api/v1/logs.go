package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Logs
// @Summary Page login logs
// @Accept json
// @Param request body dto.SearchLgLogWithPage true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /logs/login [post]
func (b *BaseApi) GetLoginLogs(c *gin.Context) {
	var req dto.SearchLgLogWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := logService.PageLoginLog(c, req)
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
// @Accept json
// @Param request body dto.SearchOpLogWithPage true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /logs/operation [post]
func (b *BaseApi) GetOperationLogs(c *gin.Context) {
	var req dto.SearchOpLogWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := logService.PageOperationLog(req)
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
// @Summary Clean operation logs
// @Accept json
// @Param request body dto.CleanLog true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /logs/clean [post]
// @x-panel-log {"bodyKeys":["logType"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"清空 [logType] 日志信息","formatEN":"Clean the [logType] log information"}
func (b *BaseApi) CleanLogs(c *gin.Context) {
	var req dto.CleanLog
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := logService.CleanLogs(req.LogType); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Logs
// @Summary Load system log files
// @Success 200 {array} string
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /logs/system/files [get]
func (b *BaseApi) GetSystemFiles(c *gin.Context) {
	data, err := logService.ListSystemLogFile()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Logs
// @Summary Load system logs
// @Success 200 {string} data
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /logs/system [post]
func (b *BaseApi) GetSystemLogs(c *gin.Context) {
	var req dto.OperationWithName
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	data, err := logService.LoadSystemLog(req.Name)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}
