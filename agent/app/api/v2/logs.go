package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Logs
// @Summary Load system log files
// @Description 获取系统日志文件列表
// @Success 200
// @Security ApiKeyAuth
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
// @Description 获取系统日志
// @Success 200
// @Security ApiKeyAuth
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
