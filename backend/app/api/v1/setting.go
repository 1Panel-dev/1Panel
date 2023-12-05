package v1

import (
	"errors"
	"os"
	"path"

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
// @x-panel-log {"bodyKeys":["key","value"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改系统配置 [key] => [value]","formatEN":"update system setting [key] => [value]"}
func (b *BaseApi) UpdateSetting(c *gin.Context) {
	var req dto.SettingUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
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
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改系统密码","formatEN":"update system password"}
func (b *BaseApi) UpdatePassword(c *gin.Context) {
	var req dto.PasswordUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
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
// @x-panel-log {"bodyKeys":["ssl"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改系统 ssl => [ssl]","formatEN":"update system ssl => [ssl]"}
func (b *BaseApi) UpdateSSL(c *gin.Context) {
	var req dto.SSLUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
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
// @Summary Download system cert
// @Description 下载证书
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/ssl/download [post]
func (b *BaseApi) DownloadSSL(c *gin.Context) {
	pathItem := path.Join(global.CONF.System.BaseDir, "1panel/secret/server.crt")
	if _, err := os.Stat(pathItem); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	c.File(pathItem)
}

// @Tags System Setting
// @Summary Load system address
// @Description 获取系统地址信息
// @Accept json
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/interface [get]
func (b *BaseApi) LoadInterfaceAddr(c *gin.Context) {
	data, err := settingService.LoadInterfaceAddr()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags System Setting
// @Summary Update system bind info
// @Description 更新系统监听信息
// @Accept json
// @Param request body dto.BindInfo true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/bind/update [post]
// @x-panel-log {"bodyKeys":["ipv6", "bindAddress"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改系统监听信息 => ipv6: [ipv6], 监听 IP: [bindAddress]","formatEN":"update system bind info => ipv6: [ipv6], 监听 IP: [bindAddress]"}
func (b *BaseApi) UpdateBindInfo(c *gin.Context) {
	var req dto.BindInfo
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := settingService.UpdateBindInfo(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags System Setting
// @Summary Update system port
// @Description 更新系统端口
// @Accept json
// @Param request body dto.PortUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /settings/port/update [post]
// @x-panel-log {"bodyKeys":["serverPort"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改系统端口 => [serverPort]","formatEN":"update system port => [serverPort]"}
func (b *BaseApi) UpdatePort(c *gin.Context) {
	var req dto.PortUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
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
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"重置过期密码","formatEN":"reset an expired Password"}
func (b *BaseApi) HandlePasswordExpired(c *gin.Context) {
	var req dto.PasswordUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := settingService.HandlePasswordExpired(c, req.OldPassword, req.NewPassword); err != nil {
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
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"清空监控数据","formatEN":"clean monitor datas"}
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
// @Accept json
// @Param request body dto.MfaCredential true "request"
// @Success 200 {object} mfa.Otp
// @Security ApiKeyAuth
// @Router /settings/mfa [post]
func (b *BaseApi) LoadMFA(c *gin.Context) {
	var req dto.MfaRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	otp, err := mfa.GetOtp("admin", req.Title, req.Interval)
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
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"mfa 绑定","formatEN":"bind mfa"}
func (b *BaseApi) MFABind(c *gin.Context) {
	var req dto.MfaCredential
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
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
