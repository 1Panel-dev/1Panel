package v1

import (
	"encoding/base64"
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/captcha"
	"github.com/gin-gonic/gin"
)

type BaseApi struct{}

// @Tags Auth
// @Summary User login
// @Accept json
// @Param EntranceCode header string true "Secure entrance base64 encrypted string"
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
// @Accept json
// @Param request body dto.MFALogin true "request"
// @Success 200 {object} dto.UserLoginInfo
// @Router /auth/mfalogin [post]
// @Header 200 {string} EntranceCode
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
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
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
// @Success 200 {boolean} isDemo
// @Router /auth/demo [get]
func (b *BaseApi) CheckIsDemo(c *gin.Context) {
	helper.SuccessWithData(c, global.CONF.System.IsDemo)
}

// @Tags Auth
// @Summary Check System isIntl
// @Success 200 {boolean} isIntl
// @Router /auth/intl [get]
func (b *BaseApi) CheckIsIntl(c *gin.Context) {
	helper.SuccessWithData(c, global.CONF.System.IsIntl)
}

// @Tags Auth
// @Summary Load System Language
// @Success 200 {string} language
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
	logs.Agent = c.GetHeader("User-Agent")
	_ = logService.CreateLoginLog(logs)
}
