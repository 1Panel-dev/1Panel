package v2

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/mfa"
	"github.com/gin-gonic/gin"
)

// @Tags System Setting
// @Summary Load system setting info
// @Success 200 {object} dto.SettingInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/search [post]
func (b *BaseApi) GetSettingInfo(c *gin.Context) {
	setting, err := settingService.GetSettingInfo()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, setting)
}

// @Tags System Setting
// @Summary Load system terminal setting info
// @Success 200 {object} dto.TerminalInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/terminal/search [post]
func (b *BaseApi) GetTerminalSettingInfo(c *gin.Context) {
	setting, err := settingService.GetTerminalInfo()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, setting)
}

// @Tags System Setting
// @Summary Load system available status
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/search/available [get]
func (b *BaseApi) GetSystemAvailable(c *gin.Context) {
	helper.Success(c)
}

// @Tags System Setting
// @Summary Update system setting
// @Accept json
// @Param request body dto.SettingUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/update [post]
// @x-panel-log {"bodyKeys":["key","value"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改系统配置 [key] => [value]","formatEN":"update system setting [key] => [value]"}
func (b *BaseApi) UpdateSetting(c *gin.Context) {
	var req dto.SettingUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if req.Key == "SecurityEntrance" {
		if !checkEntrancePattern(req.Value) {
			helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrInvalidParams", fmt.Errorf("the format of the security entrance %s is incorrect.", req.Value))
			return
		}
	}

	if err := settingService.Update(req.Key, req.Value); err != nil {
		helper.InternalServer(c, err)
		return
	}
	if req.Key == "SecurityEntrance" {
		entranceValue := base64.StdEncoding.EncodeToString([]byte(req.Value))
		c.SetCookie("SecurityEntrance", entranceValue, 0, "", "", false, true)
	}
	helper.Success(c)
}

// @Tags System Setting
// @Summary Update system terminal setting
// @Accept json
// @Param request body dto.TerminalInfo true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/terminal/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改系统终端配置","formatEN":"update system terminal setting"}
func (b *BaseApi) UpdateTerminalSetting(c *gin.Context) {
	var req dto.TerminalInfo
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := settingService.UpdateTerminal(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags System Setting
// @Summary Update proxy setting
// @Accept json
// @Param request body dto.ProxyUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/proxy/update [post]
// @x-panel-log {"bodyKeys":["proxyUrl","proxyPort"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"服务器代理配置 [proxyPort]:[proxyPort]","formatEN":"set proxy [proxyPort]:[proxyPort]."}
func (b *BaseApi) UpdateProxy(c *gin.Context) {
	var req dto.ProxyUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if len(req.ProxyPasswd) != 0 && len(req.ProxyType) != 0 {
		pass, err := base64.StdEncoding.DecodeString(req.ProxyPasswd)
		if err != nil {
			helper.BadRequest(c, err)
			return
		}
		req.ProxyPasswd = string(pass)
	}

	if err := settingService.UpdateProxy(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags System Setting
// @Summary Update system setting
// @Accept json
// @Param request body dto.SettingUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/menu/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"隐藏高级功能菜单","formatEN":"Hide advanced feature menu."}
func (b *BaseApi) UpdateMenu(c *gin.Context) {
	var req dto.SettingUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := settingService.Update(req.Key, req.Value); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags System Setting
// @Summary Update system password
// @Accept json
// @Param request body dto.PasswordUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/password/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改系统密码","formatEN":"update system password"}
func (b *BaseApi) UpdatePassword(c *gin.Context) {
	var req dto.PasswordUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := settingService.UpdatePassword(c, req.OldPassword, req.NewPassword); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags System Setting
// @Summary Update system ssl
// @Accept json
// @Param request body dto.SSLUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/ssl/update [post]
// @x-panel-log {"bodyKeys":["ssl"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改系统 ssl => [ssl]","formatEN":"update system ssl => [ssl]"}
func (b *BaseApi) UpdateSSL(c *gin.Context) {
	var req dto.SSLUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := settingService.UpdateSSL(c, req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags System Setting
// @Summary Load system cert info
// @Success 200 {object} dto.SSLInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/ssl/info [get]
func (b *BaseApi) LoadFromCert(c *gin.Context) {
	info, err := settingService.LoadFromCert()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, info)
}

// @Tags System Setting
// @Summary Download system cert
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/ssl/download [post]
func (b *BaseApi) DownloadSSL(c *gin.Context) {
	pathItem := path.Join(global.CONF.Base.InstallDir, "1panel/secret/server.crt")
	if _, err := os.Stat(pathItem); err != nil {
		helper.InternalServer(c, err)
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
// @Router /core/settings/interface [get]
func (b *BaseApi) LoadInterfaceAddr(c *gin.Context) {
	data, err := settingService.LoadInterfaceAddr()
	if err != nil {
		helper.InternalServer(c, err)
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
// @Router /core/settings/bind/update [post]
// @x-panel-log {"bodyKeys":["ipv6", "bindAddress"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改系统监听信息 => ipv6: [ipv6], 监听 IP: [bindAddress]","formatEN":"update system bind info => ipv6: [ipv6], 监听 IP: [bindAddress]"}
func (b *BaseApi) UpdateBindInfo(c *gin.Context) {
	var req dto.BindInfo
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := settingService.UpdateBindInfo(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags System Setting
// @Summary Update system port
// @Accept json
// @Param request body dto.PortUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/port/update [post]
// @x-panel-log {"bodyKeys":["serverPort"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改系统端口 => [serverPort]","formatEN":"update system port => [serverPort]"}
func (b *BaseApi) UpdatePort(c *gin.Context) {
	var req dto.PortUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := settingService.UpdatePort(req.ServerPort); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags System Setting
// @Summary Reset system password expired
// @Accept json
// @Param request body dto.PasswordUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/expired/handle [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"重置过期密码","formatEN":"reset an expired Password"}
func (b *BaseApi) HandlePasswordExpired(c *gin.Context) {
	var req dto.PasswordUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := settingService.HandlePasswordExpired(c, req.OldPassword, req.NewPassword); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags System Setting
// @Summary Load mfa info
// @Accept json
// @Param request body dto.MfaCredential true "request"
// @Success 200 {object} mfa.Otp
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/mfa [post]
func (b *BaseApi) LoadMFA(c *gin.Context) {
	var req dto.MfaRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	otp, err := mfa.GetOtp("admin", req.Title, req.Interval)
	if err != nil {
		helper.InternalServer(c, err)
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
// @Router /core/settings/mfa/bind [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"mfa 绑定","formatEN":"bind mfa"}
func (b *BaseApi) MFABind(c *gin.Context) {
	var req dto.MfaCredential
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	success := mfa.ValidCode(req.Code, req.Interval, req.Secret)
	if !success {
		helper.InternalServer(c, errors.New("code is not valid"))
		return
	}

	if err := settingService.Update("MFAInterval", req.Interval); err != nil {
		helper.InternalServer(c, err)
		return
	}

	if err := settingService.Update("MFAStatus", constant.StatusEnable); err != nil {
		helper.InternalServer(c, err)
		return
	}

	if err := settingService.Update("MFASecret", req.Secret); err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.Success(c)
}

func (b *BaseApi) ReloadSSL(c *gin.Context) {
	clientIP := c.ClientIP()
	if clientIP != "127.0.0.1" {
		helper.InternalServer(c, errors.New("only localhost can reload ssl"))
		return
	}
	if err := settingService.UpdateSystemSSL(); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags System Setting
// @Summary generate api key
// @Accept json
// @Success 200 {string} key
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/api/config/generate/key [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"生成 API 接口密钥","formatEN":"generate api key"}
func (b *BaseApi) GenerateApiKey(c *gin.Context) {
	panelToken := c.GetHeader("1Panel-Token")
	if panelToken != "" {
		helper.BadAuth(c, "ErrApiConfigDisable", nil)
		return
	}
	apiKey, err := settingService.GenerateApiKey()
	if err != nil {
		helper.InternalServer(c, err)
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
		helper.BadAuth(c, "ErrApiConfigDisable", nil)
		return
	}
	var req dto.ApiInterfaceConfig
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := settingService.UpdateApiConfig(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags App
// @Summary Update appstore config
// @Accept json
// @Param request body dto.AppstoreUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/apps/store/update [post]
func (b *BaseApi) UpdateAppstoreConfig(c *gin.Context) {
	var req dto.AppstoreUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := settingService.UpdateAppstoreConfig(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags App
// @Summary Get appstore config
// @Success 200 {object} dto.AppstoreConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/apps/store/config [get]
func (b *BaseApi) GetAppstoreConfig(c *gin.Context) {
	res, err := settingService.GetAppstoreConfig()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

func checkEntrancePattern(val string) bool {
	if len(val) == 0 {
		return true
	}
	result, _ := regexp.MatchString("^[a-zA-Z0-9]{5,116}$", val)
	return result
}
