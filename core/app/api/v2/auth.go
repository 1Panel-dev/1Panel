package v2

import (
	"encoding/base64"
	"net/http"
	"os"
	"path"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	initauth "github.com/1Panel-dev/1Panel/core/init/auth"
	"github.com/1Panel-dev/1Panel/core/utils/captcha"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
	"github.com/gin-gonic/gin"
)

type BaseApi struct{}

// @Tags Auth
// @Summary User login
// @Accept json
// @Param EntranceCode header string true "安全入口 base64 加密串"
// @Param request body dto.Login true "request"
// @Success 200 {object} dto.UserLoginInfo
// @Router /core/auth/login [post]
func (b *BaseApi) Login(c *gin.Context) {
	var req dto.Login
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	ip := common.GetRealClientIP(c)
	if global.IPTracker.IsLocked(ip) {
		helper.BadAuth(c, "ErrLoginLocked", nil)
		return
	}
	needCaptcha := global.IPTracker.NeedCaptcha(ip)
	if needCaptcha {
		if errMsg := captcha.VerifyCode(req.CaptchaID, req.Captcha); errMsg != "" {
			helper.BadAuth(c, errMsg, nil)
			return
		}
	}

	entranceItem := c.Request.Header.Get("EntranceCode")
	var entrance []byte
	if len(entranceItem) != 0 {
		entrance, _ = base64.StdEncoding.DecodeString(entranceItem)
	}
	if len(entrance) == 0 {
		cookieValue, err := c.Cookie("SecurityEntrance")
		if err == nil {
			entrance, _ = base64.StdEncoding.DecodeString(cookieValue)
		}
	}

	user, msgKey, err := xpack.AuthProvider.Login(c, req, string(entrance))
	if user == nil || user.MfaStatus != constant.StatusEnable {
		go saveLoginLogs(c, wrapLoginErr(msgKey, err))
	}
	if msgKey == "ErrAuth" || msgKey == "ErrEntrance" {
		if msgKey == "ErrAuth" {
			global.IPTracker.RecordFailure(ip)
			global.IPTracker.SetNeedCaptcha(ip)
		}
		helper.BadAuth(c, msgKey, err)
		return
	}
	if err != nil {
		global.IPTracker.RecordFailure(ip)
		global.IPTracker.SetNeedCaptcha(ip)
		helper.InternalServer(c, err)
		return
	}
	if user == nil || user.MfaStatus != constant.StatusEnable {
		global.IPTracker.Clear(ip)
	}
	helper.SuccessWithData(c, user)
}

// @Tags Auth
// @Summary User login with mfa
// @Accept json
// @Param request body dto.MFALogin true "request"
// @Success 200 {object} dto.UserLoginInfo
// @Router /core/auth/mfalogin [post]
// @Header 200 {string} EntranceCode "安全入口"
func (b *BaseApi) MFALogin(c *gin.Context) {
	var req dto.MFALogin
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	ip := common.GetRealClientIP(c)
	if global.IPTracker.IsLocked(ip) {
		helper.BadAuth(c, "ErrLoginLocked", nil)
		return
	}

	entranceItem := c.Request.Header.Get("EntranceCode")
	var entrance []byte
	if len(entranceItem) != 0 {
		entrance, _ = base64.StdEncoding.DecodeString(entranceItem)
	}

	user, msgKey, err := xpack.AuthProvider.MFALogin(c, req, string(entrance))
	go saveLoginLogs(c, wrapLoginErr(msgKey, err))
	if msgKey == "ErrMFA" {
		global.IPTracker.RecordFailure(ip)
		failures := initauth.GetMFASessionStore().RecordFailure(req.SessionID)
		if failures >= initauth.MFASessionMaxFailures {
			global.IPTracker.SetNeedCaptcha(ip)
			helper.BadAuth(c, "ErrCaptchaCode", nil)
			return
		}
		helper.BadAuth(c, msgKey, err)
		return
	}
	if err != nil {
		global.IPTracker.RecordFailure(ip)
		helper.InternalServer(c, err)
		return
	}
	global.IPTracker.Clear(ip)
	helper.SuccessWithData(c, user)
}

// @Tags Auth
// @Summary User login with passkey
// @Success 200 {object} dto.PasskeyBeginResponse
// @Router /core/auth/passkey/begin [post]
func (b *BaseApi) PasskeyBeginLogin(c *gin.Context) {
	entrance := loadEntranceFromRequest(c)
	res, msgKey, err := xpack.AuthProvider.PasskeyBeginLogin(c, entrance)
	if msgKey != "" {
		if msgKey == "ErrEntrance" {
			helper.BadAuth(c, msgKey, err)
			return
		}
		if msgKey == "ErrPasskeyNotConfigured" {
			helper.ErrorWithDetail(c, http.StatusNotFound, msgKey, err)
			return
		}
		helper.ErrorWithDetail(c, http.StatusBadRequest, msgKey, err)
		return
	}
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Auth
// @Summary User login with passkey
// @Success 200 {object} dto.UserLoginInfo
// @Router /core/auth/passkey/finish [post]
func (b *BaseApi) PasskeyFinishLogin(c *gin.Context) {
	sessionID := c.GetHeader("Passkey-Session")
	entrance := loadEntranceFromRequest(c)
	user, msgKey, err := xpack.AuthProvider.PasskeyFinishLogin(c, sessionID, entrance)
	go saveLoginLogs(c, wrapLoginErr(msgKey, err))
	if msgKey == "ErrAuth" || msgKey == "ErrEntrance" {
		if msgKey == "ErrAuth" {
			global.IPTracker.SetNeedCaptcha(common.GetRealClientIP(c))
		}
		helper.BadAuth(c, msgKey, err)
		return
	}
	if msgKey != "" {
		helper.ErrorWithDetail(c, http.StatusBadRequest, msgKey, err)
		return
	}
	if err != nil {
		global.IPTracker.SetNeedCaptcha(common.GetRealClientIP(c))
		helper.InternalServer(c, err)
		return
	}
	global.IPTracker.Clear(common.GetRealClientIP(c))
	helper.SuccessWithData(c, user)
}

// @Tags Auth
// @Summary User logout
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/auth/logout [post]
func (b *BaseApi) LogOut(c *gin.Context) {
	if err := authService.LogOut(c); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Auth
// @Summary Load captcha
// @Success 200 {object} dto.CaptchaResponse
// @Router /core/auth/captcha [get]
func (b *BaseApi) Captcha(c *gin.Context) {
	captcha, err := captcha.CreateCaptcha()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, captcha)
}

func (b *BaseApi) GetWelcomePage(c *gin.Context) {
	count, _, _ := logService.PageLoginLog(c, dto.SearchLgLogWithPage{PageInfo: dto.PageInfo{Page: 1, PageSize: 10}})
	if count != 1 {
		helper.Success(c)
		return
	}
	file, err := os.ReadFile(path.Join(global.CONF.Base.InstallDir, "1panel/welcome/index.html"))
	if err != nil {
		helper.Success(c)
		return
	}
	helper.SuccessWithData(c, string(file))
}

// @Tags Auth
// @Summary Get Setting For Login
// @Success 200 {object} dto.LoginSetting
// @Router /core/auth/setting [get]
func (b *BaseApi) GetLoginSetting(c *gin.Context) {
	settingInfo, err := settingService.GetSettingInfo()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	ip := common.GetRealClientIP(c)
	needCaptcha := global.IPTracker.NeedCaptcha(ip)
	res := &dto.LoginSetting{
		IsDemo:       global.CONF.Base.IsDemo,
		IsIntl:       global.CONF.Base.Edition == "intl",
		IsFxplay:     global.CONF.Base.IsFxplay,
		IsOffLine:    global.CONF.Base.IsOffLine,
		IsEnterprise: global.CONF.Base.IsEnterprise,
		Language:     settingInfo.Language,
		MenuTabs:     settingInfo.MenuTabs,
		PanelName:    settingInfo.PanelName,
		Theme:        settingInfo.Theme,
		NeedCaptcha:  needCaptcha,
	}
	res.PasskeySetting = xpack.AuthProvider.PasskeyStatus(c)
	helper.SuccessWithData(c, res)
}

// @Tags Auth
// @Summary Begin passkey registration
// @Accept json
// @Param request body dto.PasskeyRegisterRequest true "request"
// @Success 200 {object} dto.PasskeyBeginResponse
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/auth/passkey/register/begin [post]
func (b *BaseApi) PasskeyRegisterBegin(c *gin.Context) {
	var req dto.PasskeyRegisterRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, msgKey, err := xpack.AuthProvider.PasskeyBeginRegister(c, req.Name)
	if msgKey != "" {
		helper.ErrorWithDetail(c, http.StatusBadRequest, msgKey, err)
		return
	}
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags Auth
// @Summary Finish passkey registration
// @Accept json
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/auth/passkey/register/finish [post]
func (b *BaseApi) PasskeyRegisterFinish(c *gin.Context) {
	sessionID := c.GetHeader("Passkey-Session")
	msgKey, err := xpack.AuthProvider.PasskeyFinishRegister(c, sessionID)
	if msgKey != "" {
		helper.ErrorWithDetail(c, http.StatusBadRequest, msgKey, err)
		return
	}
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Auth
// @Summary List passkeys
// @Success 200 {array} dto.PasskeyInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/auth/passkey/list [get]
func (b *BaseApi) PasskeyList(c *gin.Context) {
	list, err := xpack.AuthProvider.PasskeyList(c)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, list)
}

// @Tags Auth
// @Summary Delete passkey
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/auth/passkey/del [post]
func (b *BaseApi) PasskeyDelete(c *gin.Context) {
	var req dto.PasskeyID
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := xpack.AuthProvider.PasskeyDelete(c, req.ID); err != nil {
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
// @Router /core/auth/mfa [post]
func (b *BaseApi) LoadMFA(c *gin.Context) {
	var req dto.MfaRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	otp, err := xpack.AuthProvider.LoadMFA(c, req)
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
// @Router /core/auth/mfa/bind [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"mfa 绑定","formatEN":"bind mfa"}
func (b *BaseApi) MFABind(c *gin.Context) {
	var req dto.MfaCredential
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := xpack.AuthProvider.MFABind(c, req); err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.Success(c)
}

// @Tags Auth
// @Summary generate api key
// @Accept json
// @Success 200 {string} key
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/auth/api/generate [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"生成 API 接口密钥","formatEN":"generate api key"}
func (b *BaseApi) GenerateApiKey(c *gin.Context) {
	panelToken := c.GetHeader("1Panel-Token")
	if panelToken != "" {
		helper.BadAuth(c, "ErrApiConfigDisable", nil)
		return
	}
	apiKey, err := xpack.AuthProvider.GenerateApiKey(c)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, apiKey)
}

// @Tags Auth
// @Summary Update api config
// @Accept json
// @Param request body dto.ApiInterfaceConfig true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/auth/api/update [post]
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

	if err := xpack.AuthProvider.UpdateApiConfig(c, req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Auth
// @Summary Load current user info
// @Success 200 {object} dto.CurrentUserInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/auth/current [get]
func (b *BaseApi) GetCurrentUser(c *gin.Context) {
	userInfo, err := xpack.AuthProvider.GetCurrentUserInfo(c)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, userInfo)
}

// @Tags Auth
// @Summary Update current user info
// @Accept json
// @Param request body dto.CurrentUserUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/auth/current/update [post]
func (b *BaseApi) UpdateCurrentUser(c *gin.Context) {
	var req dto.CurrentUserUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := xpack.AuthProvider.UpdateCurrentUserInfo(c, req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Auth
// @Summary Reset system password expired
// @Accept json
// @Param request body dto.PasswordUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /core/auth/expired/reset [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"重置过期密码","formatEN":"reset an expired Password"}
func (b *BaseApi) ResetPassword(c *gin.Context) {
	var req dto.PasswordUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := xpack.AuthProvider.HandlePasswordExpired(c, req.OldPassword, req.NewPassword); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func saveLoginLogs(c *gin.Context, err error) {
	var logs model.LoginLog
	if err != nil {
		logs.Status = constant.StatusFailed
		logs.Message = err.Error()
	} else {
		logs.Status = constant.StatusSuccess
	}
	logs.IP = c.ClientIP()
	logs.Agent = c.GetHeader("User-Agent")
	_ = logService.CreateLoginLog(logs)
}

func wrapLoginErr(msgKey string, err error) error {
	if err != nil {
		return err
	}
	if msgKey == "ErrAuth" || msgKey == "ErrEntrance" {
		return buserr.New(msgKey)
	}
	return nil
}

func loadEntranceFromRequest(c *gin.Context) string {
	entranceItem := c.Request.Header.Get("EntranceCode")
	var entrance []byte
	if len(entranceItem) != 0 {
		entrance, _ = base64.StdEncoding.DecodeString(entranceItem)
	}
	if len(entrance) == 0 {
		cookieValue, err := c.Cookie("SecurityEntrance")
		if err == nil {
			entrance, _ = base64.StdEncoding.DecodeString(cookieValue)
		}
	}
	return string(entrance)
}
