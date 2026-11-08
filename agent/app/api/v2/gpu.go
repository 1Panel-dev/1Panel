package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) LoadModelGPUDevices(c *gin.Context) {
	helper.SuccessWithData(c, service.LoadModelGPUDevices())
}

// @Tags AI
// @Summary Load gpu / xpu info
// @Accept json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/gpu/load [get]
func (b *BaseApi) LoadGpuInfo(c *gin.Context) {
	info, err := service.LoadGPUInfo(c.Request.Context(), c.Query("refresh") == "true")
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, info)
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

func (b *BaseApi) ListGPUGroups(c *gin.Context) {
	result, err := gpuGroupService.List()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

func (b *BaseApi) SaveGPUGroup(c *gin.Context) {
	var req dto.GPUGroupSave
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := gpuGroupService.Save(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) DeleteGPUGroup(c *gin.Context) {
	var req dto.OperateByID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := gpuGroupService.Delete(req.ID); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) LoadModelGPUUsage(c *gin.Context) {
	var req dto.ModelGPUUsageRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	result, err := service.LoadModelGPUUsage(c.Request.Context(), req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}
