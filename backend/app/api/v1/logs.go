package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/gin-gonic/gin"
)

// Page login logs
// @Tags Logs
// @Summary Page login logs
// @Description 获取系统登录日志列表分页
// @Accept json
// @Param request body request.PageInfo true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /logs/login [post]
func (b *BaseApi) GetLoginLogs(c *gin.Context) {
	var req dto.PageInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
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

// Page operation logs
// @Tags Logs
// @Summary Page operation logs
// @Description 获取系统操作日志列表分页
// @Accept json
// @Param request body request.PageInfo true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /logs/operation [post]
func (b *BaseApi) GetOperationLogs(c *gin.Context) {
	var req dto.PageInfo
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
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

// Clean operation logs
// @Tags Logs
// @Summary Clean operation logs
// @Description 清空操作日志
// @Accept json
// @Param request body dto.CleanLog true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /logs/clean [post]
// @x-panel-log {"bodyKeys":["logType"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"清空 [logType] 日志信息","formatEN":"Clean the [logType] log information"}
func (b *BaseApi) CleanLogs(c *gin.Context) {
	var req dto.CleanLog
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := logService.CleanLogs(req.LogType); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}
