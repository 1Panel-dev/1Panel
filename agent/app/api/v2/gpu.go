package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/gpu"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/gpu/common"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/xpu"
	"github.com/gin-gonic/gin"
)

// @Tags AI
// @Summary Load gpu / xpu info
// @Accept json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/gpu/load [get]
func (b *BaseApi) LoadGpuInfo(c *gin.Context) {
	ok, client := gpu.New()
	if ok {
		info, err := client.LoadGpuInfo()
		if err != nil {
			helper.BadRequest(c, err)
			return
		}
		helper.SuccessWithData(c, info)
		return
	}
	xpuOK, xpuClient := xpu.New()
	if xpuOK {
		info, err := xpuClient.LoadGpuInfo()
		if err != nil {
			helper.BadRequest(c, err)
			return
		}
		helper.SuccessWithData(c, info)
		return
	}
	helper.SuccessWithData(c, &common.GpuInfo{})
}

// @Tags AI
// @Summary Get CPU options
// @Router /ai/gpu/options [get]
func (b *BaseApi) GetCPUOptions(c *gin.Context) {
	helper.SuccessWithData(c, monitorService.LoadGPUOptions())
}

// @Tags Monitor
// @Summary Load monitor data
// @Param request body dto.MonitorGPUSearch true "request"
// @Success 200 {object} dto.MonitorGPUData
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/gpu/search [post]
func (b *BaseApi) LoadGPUMonitor(c *gin.Context) {
	var req dto.MonitorGPUSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	data, err := monitorService.LoadGPUMonitorData(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}
