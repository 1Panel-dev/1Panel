package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) LoadVLLMMonitor(c *gin.Context) {
	var req dto.MonitorVLLMSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := monitorService.LoadVLLMMonitorData(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

func (b *BaseApi) LoadVLLMCurrent(c *gin.Context) {
	var req dto.MonitorVLLMCurrent
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := monitorService.LoadVLLMCurrent(c.Request.Context(), req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

func (b *BaseApi) CleanVLLMMonitor(c *gin.Context) {
	var req dto.MonitorVLLMClean
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := monitorService.CleanVLLMMonitor(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}
