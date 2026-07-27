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
// @Summary Get host system log status
// @Produce json
// @Success 200 {object} dto.SystemLogStatus
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /logs/system/status [get]
func (b *BaseApi) GetSystemLogStatus(c *gin.Context) {
	data, err := logService.GetSystemLogStatus()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Logs
// @Summary Read host logs
// @Accept json
// @Param request body dto.SystemLogReq true "request"
// @Produce json
// @Success 200 {object} dto.SystemLogRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /logs/system/read [post]
func (b *BaseApi) ReadSystemLog(c *gin.Context) {
	var req dto.SystemLogReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := logService.ReadSystemLog(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Logs
// @Summary List running host services
// @Produce json
// @Success 200 {array} string
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /logs/system/services [get]
func (b *BaseApi) ListRunningServices(c *gin.Context) {
	data, err := logService.ListRunningServices()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}
