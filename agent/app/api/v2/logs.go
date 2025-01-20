package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/gin-gonic/gin"
)

// @Tags Logs
// @Summary Load system log files
// @Success 200 {array} string
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /logs/system/files [get]
func (b *BaseApi) GetSystemFiles(c *gin.Context) {
	data, err := logService.ListSystemLogFile()
	if err != nil {
		helper.InternalServer(c, err)
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
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, data)
}
