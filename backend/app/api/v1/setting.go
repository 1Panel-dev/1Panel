package v1

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/mfa"
	"github.com/gin-gonic/gin"
)

// @Tags System Setting
// @Summary Load system setting info
// @Description 加载系统配置信息
// @Success 200 {object} dto.SettingInfo
// @Security ApiKeyAuth
// @Router /settings/search [post]
func (b *BaseApi) GetSettingInfo(c *gin.Context) {
	setting, err := settingService.GetSettingInfo()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, setting)
}

// @Tags System Setting
// @Summary Load system available status
// @Description 获取系统可用状态
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/search/available [get]
func (b *BaseApi) GetSystemAvailable(c *gin.Context) {
	helper.SuccessWithData(c, nil)
}

// @Tags System Setting
// @Summary Update system setting
// @Description 更新系统配置
// @Accept json
// @Param request body dto.SettingUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/update [post]
// @x-panel-log {"bodyKeys":["key","value"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"修改系统配置 [key] => [value]","formatEN":"update system setting [key] => [value]"}
func (b *BaseApi) UpdateSetting(c *gin.Context) {
	var req dto.SettingUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := settingService.Update(req.Key, req.Value); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags System Setting
// @Summary Update system password
// @Description 更新系统登录密码
// @Accept json
// @Param request body dto.PasswordUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/password/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFuntions":[],"formatZH":"修改系统密码","formatEN":"update system password"}
func (b *BaseApi) UpdatePassword(c *gin.Context) {
	var req dto.PasswordUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := settingService.UpdatePassword(c, req.OldPassword, req.NewPassword); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags System Setting
// @Summary Update system ssl
// @Description 修改系统 ssl 登录
// @Accept json
// @Param request body dto.SSLUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/ssl/update [post]
// @x-panel-log {"bodyKeys":["ssl"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"修改系统 ssl => [ssl]","formatEN":"update system ssl => [ssl]"}
func (b *BaseApi) UpdateSSL(c *gin.Context) {
	var req dto.SSLUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := settingService.UpdateSSL(c, req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags System Setting
// @Summary Load system cert info
// @Description 获取证书信息
// @Success 200 {object} dto.SettingInfo
// @Security ApiKeyAuth
// @Router /settings/ssl/info [get]
func (b *BaseApi) LoadFromCert(c *gin.Context) {
	info, err := settingService.LoadFromCert()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, info)
}

// @Tags System Setting
// @Summary Update system port
// @Description 更新系统端口
// @Accept json
// @Param request body dto.PortUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/port/update [post]
// @x-panel-log {"bodyKeys":["serverPort"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"修改系统端口 => [serverPort]","formatEN":"update system port => [serverPort]"}
func (b *BaseApi) UpdatePort(c *gin.Context) {
	var req dto.PortUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := settingService.UpdatePort(req.ServerPort); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags System Setting
// @Summary Reset system password expired
// @Description 重置过期系统登录密码
// @Accept json
// @Param request body dto.PasswordUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/expired/handle [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFuntions":[],"formatZH":"重置过期密码","formatEN":"reset an expired Password"}
func (b *BaseApi) HandlePasswordExpired(c *gin.Context) {
	var req dto.PasswordUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := settingService.HandlePasswordExpired(c, req.OldPassword, req.NewPassword); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags System Setting
// @Summary Load time zone options
// @Description 加载系统可用时区
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/time/option [get]
func (b *BaseApi) LoadTimeZone(c *gin.Context) {
	zones, err := settingService.LoadTimeZone()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, zones)
}

// @Tags System Setting
// @Summary Sync system time
// @Description 系统时间同步
// @Accept json
// @Param request body dto.SyncTime true "request"
// @Success 200 {string} ntime
// @Security ApiKeyAuth
// @Router /settings/time/sync [post]
// @x-panel-log {"bodyKeys":["ntpSite"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"系统时间同步[ntpSite]","formatEN":"sync system time [ntpSite]"}
func (b *BaseApi) SyncTime(c *gin.Context) {
	var req dto.SyncTime
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := settingService.SyncTime(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags System Setting
// @Summary Load local backup dir
// @Description 获取安装根目录
// @Success 200 {string} path
// @Security ApiKeyAuth
// @Router /settings/basedir [get]
func (b *BaseApi) LoadBaseDir(c *gin.Context) {
	helper.SuccessWithData(c, global.CONF.System.DataDir)
}

// @Tags System Setting
// @Summary Clean monitor datas
// @Description 清空监控数据
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/monitor/clean [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFuntions":[],"formatZH":"清空监控数据","formatEN":"clean monitor datas"}
func (b *BaseApi) CleanMonitor(c *gin.Context) {
	if err := global.DB.Exec("DELETE FROM monitor_bases").Error; err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	if err := global.DB.Exec("DELETE FROM monitor_ios").Error; err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	if err := global.DB.Exec("DELETE FROM monitor_networks").Error; err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags System Setting
// @Summary Load mfa info
// @Description 获取 mfa 信息
// @Param interval path string true "request"
// @Success 200 {object} mfa.Otp
// @Security ApiKeyAuth
// @Router /settings/mfa/:interval [get]
func (b *BaseApi) GetMFA(c *gin.Context) {
	intervalStr, ok := c.Params.Get("interval")
	if !ok {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, errors.New("error interval in path"))
		return
	}
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, fmt.Errorf("type conversion failed, err: %v", err))
		return
	}

	otp, err := mfa.GetOtp("admin", interval)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, otp)
}

// @Tags System Setting
// @Summary Bind mfa
// @Description Mfa 绑定
// @Accept json
// @Param request body dto.MfaCredential true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/mfa/bind [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFuntions":[],"formatZH":"mfa 绑定","formatEN":"bind mfa"}
func (b *BaseApi) MFABind(c *gin.Context) {
	var req dto.MfaCredential
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	success := mfa.ValidCode(req.Code, req.Interval, req.Secret)
	if !success {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, errors.New("code is not valid"))
		return
	}

	if err := settingService.Update("MFAInterval", req.Interval); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	if err := settingService.Update("MFAStatus", "enable"); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	if err := settingService.Update("MFASecret", req.Secret); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}
