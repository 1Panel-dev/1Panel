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
// @Success 200 {object} dto.DeviceBaseInfo
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Accept json
// @Success 200 {array} string
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Accept json
// @Param request body dto.OperationWithName true "request"
// @Success 200 {array} string
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Accept json
// @Param request body dto.UpdateByNameAndFile true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Accept json
// @Param request body dto.SettingUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Accept json
// @Param request body dto.ChangePasswd true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Accept json
// @Param request body dto.SwapHelper true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Accept json
// @Param request body dto.SettingUpdate true "request"
// @Success 200 {boolean} data
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Success 200 {object} dto.CleanData
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /toolbox/scan [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"扫描系统垃圾文件","formatEN":"scan System Junk Files"}
func (b *BaseApi) ScanSystem(c *gin.Context) {
	helper.SuccessWithData(c, deviceService.Scan())
}

// @Tags Device
// @Summary Clean system
// @Accept json
// @Param request body []dto.Clean true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
