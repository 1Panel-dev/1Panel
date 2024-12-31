package v1

import (
	"sort"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
)

// @Tags Monitor
// @Summary Load monitor datas
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

	loc, _ := time.LoadLocation(common.LoadTimeZoneByCmd())
	req.StartTime = req.StartTime.In(loc)
	req.EndTime = req.EndTime.In(loc)

	var backdatas []dto.MonitorData
	if req.Param == "all" || req.Param == "cpu" || req.Param == "memory" || req.Param == "load" {
		var bases []model.MonitorBase
		if err := global.MonitorDB.
			Where("created_at > ? AND created_at < ?", req.StartTime, req.EndTime).
			Find(&bases).Error; err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}

		var itemData dto.MonitorData
		itemData.Param = "base"
		for _, base := range bases {
			itemData.Date = append(itemData.Date, base.CreatedAt)
			itemData.Value = append(itemData.Value, base)
		}
		backdatas = append(backdatas, itemData)
	}
	if req.Param == "all" || req.Param == "io" {
		var bases []model.MonitorIO
		if err := global.MonitorDB.
			Where("created_at > ? AND created_at < ?", req.StartTime, req.EndTime).
			Find(&bases).Error; err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}

		var itemData dto.MonitorData
		itemData.Param = "io"
		for _, base := range bases {
			itemData.Date = append(itemData.Date, base.CreatedAt)
			itemData.Value = append(itemData.Value, base)
		}
		backdatas = append(backdatas, itemData)
	}
	if req.Param == "all" || req.Param == "network" {
		var bases []model.MonitorNetwork
		if err := global.MonitorDB.
			Where("name = ? AND created_at > ? AND created_at < ?", req.Info, req.StartTime, req.EndTime).
			Find(&bases).Error; err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}

		var itemData dto.MonitorData
		itemData.Param = "network"
		for _, base := range bases {
			itemData.Date = append(itemData.Date, base.CreatedAt)
			itemData.Value = append(itemData.Value, base)
		}
		backdatas = append(backdatas, itemData)
	}
	helper.SuccessWithData(c, backdatas)
}

// @Tags Monitor
// @Summary Clean monitor datas
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/monitor/clean [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"清空监控数据","formatEN":"clean monitor datas"}
func (b *BaseApi) CleanMonitor(c *gin.Context) {
	if err := global.MonitorDB.Exec("DELETE FROM monitor_bases").Error; err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	if err := global.MonitorDB.Exec("DELETE FROM monitor_ios").Error; err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	if err := global.MonitorDB.Exec("DELETE FROM monitor_networks").Error; err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
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
