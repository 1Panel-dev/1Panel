package v2

import (
	"sort"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
)

// @Tags Monitor
// @Summary Load monitor data
// @Param request body dto.MonitorSearch true "request"
// @Success 200 {array} dto.MonitorData
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/monitor/search [post]
func (b *BaseApi) LoadMonitor(c *gin.Context) {
	var req dto.MonitorSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	data, err := monitorService.LoadMonitorData(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Monitor
// @Summary Clean monitor data
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/monitor/clean [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"清空监控数据","formatEN":"clean monitor datas"}
func (b *BaseApi) CleanMonitor(c *gin.Context) {
	if err := monitorService.CleanData(); err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithOutData(c)
}

// @Tags Monitor
// @Summary Load monitor setting
// @Success 200 {object} dto.MonitorSetting
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/monitor/setting [get]
func (b *BaseApi) LoadMonitorSetting(c *gin.Context) {
	setting, err := monitorService.LoadSetting()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, setting)
}

// @Tags Monitor
// @Summary Update monitor setting
// @Param request body dto.MonitorSettingUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/monitor/setting/update [post]
// @x-panel-log {"bodyKeys":["key", "value"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改默认监控网卡 [name]-[value]","formatEN":"update default monitor [name]-[value]"}
func (b *BaseApi) UpdateMonitorSetting(c *gin.Context) {
	var req dto.MonitorSettingUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := monitorService.UpdateSetting(req.Key, req.Value); err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithOutData(c)
}

func (b *BaseApi) GetNetworkOptions(c *gin.Context) {
	netStat, _ := net.IOCounters(true)
	var options []string
	options = append(options, "all")
	for _, net := range netStat {
		options = append(options, net.Name)
	}
	sort.Strings(options)
	helper.SuccessWithData(c, options)
}

func (b *BaseApi) GetIOOptions(c *gin.Context) {
	diskStat, _ := disk.IOCounters()
	var options []string
	options = append(options, "all")
	for _, net := range diskStat {
		options = append(options, net.Name)
	}
	sort.Strings(options)
	helper.SuccessWithData(c, options)
}
