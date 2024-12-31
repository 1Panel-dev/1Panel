package v1

import (
	"encoding/base64"
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
// @Success 200 {object} dto.SettingInfo
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/search/available [get]
func (b *BaseApi) GetSystemAvailable(c *gin.Context) {
	helper.SuccessWithData(c, nil)
}

// @Tags System Setting
// @Summary Update system setting
// @Accept json
// @Param request body dto.SettingUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Summary Update proxy setting
// @Accept json
// @Param request body dto.ProxyUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/proxy/update [post]
// @x-panel-log {"bodyKeys":["proxyUrl","proxyPort"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"服务器代理配置 [proxyPort]:[proxyPort]","formatEN":"set proxy [proxyPort]:[proxyPort]."}
func (b *BaseApi) UpdateProxy(c *gin.Context) {
	var req dto.ProxyUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if len(req.ProxyPasswd) != 0 && len(req.ProxyType) != 0 {
		pass, err := base64.StdEncoding.DecodeString(req.ProxyPasswd)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
			return
		}
		req.ProxyPasswd = string(pass)
	}

	if err := settingService.UpdateProxy(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags System Setting
// @Summary Update system setting
// @Accept json
// @Param request body dto.SettingUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/menu/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"隐藏高级功能菜单","formatEN":"Hide advanced feature menu."}
func (b *BaseApi) UpdateMenu(c *gin.Context) {
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
// @Accept json
// @Param request body dto.PasswordUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Accept json
// @Param request body dto.SSLUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Success 200 {object} dto.SSLInfo
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Accept json
// @Success 200 {array} string
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Accept json
// @Param request body dto.BindInfo true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Accept json
// @Param request body dto.PortUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Accept json
// @Param request body dto.PasswordUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Summary Load local base dir
// @Success 200 {string} path
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/basedir [get]
func (b *BaseApi) LoadBaseDir(c *gin.Context) {
	helper.SuccessWithData(c, global.CONF.System.DataDir)
}

// @Tags System Setting
// @Summary Load mfa info
// @Accept json
// @Param request body dto.MfaCredential true "request"
// @Success 200 {object} mfa.Otp
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Accept json
// @Param request body dto.MfaCredential true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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

// @Tags System Setting
// @Summary Generate api key
// @Accept json
// @Success 200 {string} apiKey
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/api/config/generate/key [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"生成 API 接口密钥","formatEN":"generate api key"}
func (b *BaseApi) GenerateApiKey(c *gin.Context) {
	panelToken := c.GetHeader("1Panel-Token")
	if panelToken != "" {
		helper.ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrApiConfigDisable, nil)
		return
	}
	apiKey, err := settingService.GenerateApiKey()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, apiKey)
}

// @Tags System Setting
// @Summary Update api config
// @Accept json
// @Param request body dto.ApiInterfaceConfig true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/api/config/update [post]
// @x-panel-log {"bodyKeys":["ipWhiteList"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新 API 接口配置 => IP 白名单: [ipWhiteList]","formatEN":"update api config => IP White List: [ipWhiteList]"}
func (b *BaseApi) UpdateApiConfig(c *gin.Context) {
	panelToken := c.GetHeader("1Panel-Token")
	if panelToken != "" {
		helper.ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrApiConfigDisable, nil)
		return
	}
	var req dto.ApiInterfaceConfig
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := settingService.UpdateApiConfig(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
