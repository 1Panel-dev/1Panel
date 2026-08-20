package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/accelerator"
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
	ok, client := accelerator.New()
	if ok {
		snapshot, err := client.Collect(c.Request.Context())
		if err != nil {
			helper.BadRequest(c, err)
			return
		}
		if warning := snapshot.Warning(); warning != nil {
			global.LOG.Warnf("load realtime accelerator data partially failed, err: %v", warning)
		}
		helper.SuccessWithData(c, &snapshot.Info)
		return
	}
	helper.SuccessWithData(c, &accelerator.Info{})
}

// @Tags AI
// @Summary Get CPU options
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
