package v2

import (
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	appauth "github.com/1Panel-dev/1Panel/core/app/auth"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/common"
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
// @Summary Load base system setting info
// @Success 200 {object} dto.SettingBaseInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/search/base [post]
func (b *BaseApi) GetSettingBaseInfo(c *gin.Context) {
	setting, err := settingService.GetSettingBaseInfo()
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
			helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrInvalidParams", buserr.WithName("ErrEntranceFormat", req.Value))
			return
		}
	}
	if !checkSettingValueRange(req.Key, req.Value) {
		helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrInvalidParams", buserr.WithName("ErrInvalidParams", req.Value))
		return
	}
	if req.Key == "PasskeyTrustedProxies" {
		value, err := normalizePasskeyTrustedProxies(req.Value)
		if err != nil {
			helper.BadRequest(c, err)
			return
		}
		req.Value = value
	}

	if err := settingService.Update(c, req.Key, req.Value); err != nil {
		helper.InternalServer(c, err)
		return
	}
	if req.Key == "SecurityEntrance" {
		appauth.SetSecurityEntranceCookie(c, req.Value)
	}
	helper.Success(c)
}

func checkSettingValueRange(key, value string) bool {
	switch key {
	case "SessionTimeout":
		valueNum, err := strconv.Atoi(value)
		if err != nil {
			return false
		}
		return valueNum >= 300 && valueNum <= 864000
	case "ExpirationDays":
		valueNum, err := strconv.Atoi(value)
		if err != nil {
			return false
		}
		return valueNum >= 0 && valueNum <= 60
	default:
		return true
	}
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

	if err := settingService.Update(c, req.Key, req.Value); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Menu Setting
// @Summary Default menu
// @Accept json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/menu/default [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"初始化菜单","formatEN":"Init menu."}
func (b *BaseApi) DefaultMenu(c *gin.Context) {
	if err := settingService.DefaultMenu(); err != nil {
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

// @Tags App
// @Summary Update appstore config
// @Accept json
// @Param request body dto.AppstoreUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/apps/store/update [post]
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
// @Router /core/settings/apps/store/config [get]
func (b *BaseApi) GetAppstoreConfig(c *gin.Context) {
	res, err := settingService.GetAppstoreConfig()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags System Setting
// @Summary Load dashboard memo
// @Success 200 {string} memo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/memo [get]
func (b *BaseApi) GetMemo(c *gin.Context) {
	memo, err := settingService.GetMemo()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, memo)
}

// @Tags System Setting
// @Summary Update dashboard memo
// @Accept json
// @Param request body dto.MemoUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/settings/memo [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新仪表盘备忘录","formatEN":"update dashboard memo"}
func (b *BaseApi) UpdateMemo(c *gin.Context) {
	var req dto.MemoUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := settingService.UpdateMemo(req.Content); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func checkEntrancePattern(val string) bool {
	if len(val) == 0 {
		return true
	}
	result, _ := regexp.MatchString("^[a-zA-Z0-9]{5,116}$", val)
	if !result {
		return false
	}
	lowerVal := strings.ToLower(val)
	for key := range constant.WebUrlMap {
		if key == "/" || !strings.HasPrefix(key, "/") {
			continue
		}
		if strings.Count(key, "/") != 1 {
			continue
		}
		segment := strings.ToLower(strings.TrimPrefix(key, "/"))
		if len(segment) < 5 {
			continue
		}
		if lowerVal == segment {
			return false
		}
	}
	assetsList := [2]string{"public", "assets"}
	for _, item := range assetsList {
		if lowerVal == item {
			return false
		}
	}
	return true
}

func normalizePasskeyTrustedProxies(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	lines := strings.Split(value, "\n")
	validLines := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		validLines = append(validLines, line)
	}
	if len(validLines) == 0 {
		return "", nil
	}
	normalized := strings.Join(validLines, "\n")
	if _, err := common.HandleIPList(normalized); err != nil {
		return "", err
	}
	return normalized, nil
}
