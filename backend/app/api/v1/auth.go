package v1

import (
	"encoding/base64"
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/captcha"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/geo"
	"github.com/gin-gonic/gin"
)

type BaseApi struct{}

// @Tags Auth
// @Summary User login
// @Description 用户登录
// @Accept json
// @Param EntranceCode header string true "安全入口 base64 加密串"
// @Param request body dto.Login true "request"
// @Success 200 {object} dto.UserLoginInfo
// @Router /auth/login [post]
func (b *BaseApi) Login(c *gin.Context) {
	var req dto.Login
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if req.AuthMethod != "jwt" && !req.IgnoreCaptcha {
		if err := captcha.VerifyCode(req.CaptchaID, req.Captcha); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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

	user, err := authService.Login(c, req, string(entrance))
	go saveLoginLogs(c, err)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, user)
}

// @Tags Auth
// @Summary User login with mfa
// @Description 用户 mfa 登录
// @Accept json
// @Param request body dto.MFALogin true "request"
// @Success 200 {object} dto.UserLoginInfo
// @Router /auth/mfalogin [post]
// @Header 200 {string} EntranceCode "安全入口"
func (b *BaseApi) MFALogin(c *gin.Context) {
	var req dto.MFALogin
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	entranceItem := c.Request.Header.Get("EntranceCode")
	var entrance []byte
	if len(entranceItem) != 0 {
		entrance, _ = base64.StdEncoding.DecodeString(entranceItem)
	}

	user, err := authService.MFALogin(c, req, string(entrance))
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, user)
}

// @Tags Auth
// @Summary User logout
// @Description 用户登出
// @Success 200
// @Security ApiKeyAuth
// @Router /auth/logout [post]
func (b *BaseApi) LogOut(c *gin.Context) {
	if err := authService.LogOut(c); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Auth
// @Summary Load captcha
// @Description 加载验证码
// @Success 200 {object} dto.CaptchaResponse
// @Router /auth/captcha [get]
func (b *BaseApi) Captcha(c *gin.Context) {
	captcha, err := captcha.CreateCaptcha()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, captcha)
}

func (b *BaseApi) GetResponsePage(c *gin.Context) {
	pageCode, err := authService.GetResponsePage()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, pageCode)
}

// @Tags Auth
// @Summary Check System isDemo
// @Description 判断是否为demo环境
// @Success 200
// @Router /auth/demo [get]
func (b *BaseApi) CheckIsDemo(c *gin.Context) {
	helper.SuccessWithData(c, global.CONF.System.IsDemo)
}

// @Tags Auth
// @Summary Check System isDemo
// @Description 判断是否为国际版
// @Success 200
// @Router /auth/intl [get]
func (b *BaseApi) CheckIsIntl(c *gin.Context) {
	helper.SuccessWithData(c, global.CONF.System.IsIntl)
}

// @Tags Auth
// @Summary Load System Language
// @Description 获取系统语言设置
// @Success 200
// @Router /auth/language [get]
func (b *BaseApi) GetLanguage(c *gin.Context) {
	settingInfo, err := settingService.GetSettingInfo()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, settingInfo.Language)
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
	address, err := geo.GetIPLocation(logs.IP, common.GetLang(c))
	if err != nil {
		global.LOG.Errorf("get ip location failed: %s", err)
	}
	logs.Agent = c.GetHeader("User-Agent")
	logs.Address = address
	_ = logService.CreateLoginLog(logs)
}
