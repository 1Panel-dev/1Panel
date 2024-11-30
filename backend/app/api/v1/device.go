package v1

import (
	"encoding/base64"

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
// @Router /toolbox/device/base [post]
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
// @Success 200 {Array} string
// @Security ApiKeyAuth
// @Router /toolbox/device/zone/options [get]
func (b *BaseApi) LoadTimeOption(c *gin.Context) {
	list, err := deviceService.LoadTimeZone()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// @Tags Device
// @Summary load conf
// @Description 获取系统配置文件
// @Accept json
// @Param request body dto.OperationWithName true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /toolbox/device/conf [post]
func (b *BaseApi) LoadDeviceConf(c *gin.Context) {
	var req dto.OperationWithName
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	list, err := deviceService.LoadConf(req.Name)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, list)
}

// @Tags Device
// @Summary Update device conf by file
// @Description 通过文件修改配置
// @Accept json
// @Param request body dto.UpdateByNameAndFile true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /toolbox/device/update/byconf [post]
func (b *BaseApi) UpdateDeviceByFile(c *gin.Context) {
	var req dto.UpdateByNameAndFile
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := deviceService.UpdateByConf(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Device
// @Summary Update device
// @Description 修改系统参数
// @Accept json
// @Param request body dto.SettingUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /toolbox/device/update/conf [post]
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
// @Success 200
// @Security ApiKeyAuth
// @Router /toolbox/device/update/host [post]
// @x-panel-log {"bodyKeys":["key","value"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改主机 Host [key] => [value]","formatEN":"update device host [key] => [value]"}
func (b *BaseApi) UpdateDeviceHost(c *gin.Context) {
	var req []dto.HostHelper
	if err := helper.CheckBind(&req, c); err != nil {
		return
	}

	if err := deviceService.UpdateHosts(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Device
// @Summary Update device passwd
// @Description 修改系统密码
// @Accept json
// @Param request body dto.ChangePasswd true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /toolbox/device/update/passwd [post]
func (b *BaseApi) UpdateDevicePasswd(c *gin.Context) {
	var req dto.ChangePasswd
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if len(req.Passwd) != 0 {
		password, err := base64.StdEncoding.DecodeString(req.Passwd)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		req.Passwd = string(password)
	}
	if err := deviceService.UpdatePasswd(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Device
// @Summary Update device swap
// @Description 修改系统 Swap
// @Accept json
// @Param request body dto.SwapHelper true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /toolbox/device/update/swap [post]
// @x-panel-log {"bodyKeys":["operate","path"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operate] 主机 swap [path]","formatEN":"[operate] device swap [path]"}
func (b *BaseApi) UpdateDeviceSwap(c *gin.Context) {
	var req dto.SwapHelper
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := deviceService.UpdateSwap(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Device
// @Summary Check device DNS conf
// @Description 检查系统 DNS 配置可用性
// @Accept json
// @Param request body dto.SettingUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /toolbox/device/check/dns [post]
func (b *BaseApi) CheckDNS(c *gin.Context) {
	var req dto.SettingUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	data, err := deviceService.CheckDNS(req.Key, req.Value)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Device
// @Summary Scan system
// @Description 扫描系统垃圾文件
// @Success 200
// @Security ApiKeyAuth
// @Router /toolbox/scan [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"扫描系统垃圾文件","formatEN":"scan System Junk Files"}
func (b *BaseApi) ScanSystem(c *gin.Context) {
	helper.SuccessWithData(c, deviceService.Scan())
}

// @Tags Device
// @Summary Clean system
// @Description 清理系统垃圾文件
// @Accept json
// @Param request body []dto.Clean true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /toolbox/clean [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"清理系统垃圾文件","formatEN":"Clean system junk files"}
func (b *BaseApi) SystemClean(c *gin.Context) {
	var req []dto.Clean
	if err := helper.CheckBind(&req, c); err != nil {
		return
	}

	deviceService.Clean(req)

	helper.SuccessWithData(c, nil)
}
