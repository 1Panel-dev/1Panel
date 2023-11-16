package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Device
// @Summary Load device base info
// @Description 获取设备基础信息
// @Success 200 {object} dto.DeviceBaseInfo
// @Security ApiKeyAuth
// @Router /toolbox/device/base [get]
func (b *BaseApi) LoadDeviceBaseInfo(c *gin.Context) {
	data, err := deviceService.LoadBaseInfo()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Device
// @Summary list time zone options
// @Description 获取系统可用时区选项
// @Accept json
// @Success 200 {Array} dto.TimeZoneOptions
// @Security ApiKeyAuth
// @Router /toolbox/device/zone/options [get]
func (b *BaseApi) SearchTimeOption(c *gin.Context) {
	list, err := deviceService.LoadTimeZone()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// @Tags Device
// @Summary Update device
// @Description 修改系统参数
// @Accept json
// @Param request body dto.SettingUpdate true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /toolbox/device/conf [post]
// @x-panel-log {"bodyKeys":["key","value"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改主机参数 [key] => [value]","formatEN":"update device conf [key] => [value]"}
func (b *BaseApi) UpdateDeviceConf(c *gin.Context) {
	var req dto.SettingUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := deviceService.Update(req.Key, req.Value); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Device
// @Summary Update device hosts
// @Description 修改系统 hosts
// @Accept json
// @Param request body {Array} true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Router /toolbox/device/host [post]
// @x-panel-log {"bodyKeys":["key","value"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改主机 Host [key] => [value]","formatEN":"update device host [key] => [value]"}
func (b *BaseApi) UpdateDeviceHost(c *gin.Context) {
	var req []dto.HostHelper
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := deviceService.UpdateHosts(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}
