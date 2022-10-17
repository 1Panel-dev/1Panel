package v1

import (
	"errors"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/captcha"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
	"github.com/gin-gonic/gin"
)

type BaseApi struct{}

func (b *BaseApi) Login(c *gin.Context) {
	var req dto.Login
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := captcha.VerifyCode(req.CaptchaID, req.Captcha); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	user, err := authService.Login(c, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, user)
}

func (b *BaseApi) MFALogin(c *gin.Context) {
	var req dto.MFALogin
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	user, err := authService.MFALogin(c, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, user)
}

func (b *BaseApi) LogOut(c *gin.Context) {
	if err := authService.LogOut(c); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

func (b *BaseApi) Captcha(c *gin.Context) {
	captcha, err := captcha.CreateCaptcha()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, captcha)
}

func (b *BaseApi) GetSafetyStatus(c *gin.Context) {
	if err := authService.SafetyStatus(c); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrUnSafety, constant.ErrTypeNotSafety, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

func (b *BaseApi) SafeEntrance(c *gin.Context) {
	code, exist := c.Params.Get("code")
	if !exist {
		helper.ErrorWithDetail(c, constant.CodeErrUnSafety, constant.ErrTypeNotSafety, errors.New("missing code"))
		return
	}
	ok, err := authService.VerifyCode(code)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrUnSafety, constant.ErrTypeNotSafety, errors.New("missing code"))
		return
	}
	if !ok {
		helper.ErrorWithDetail(c, constant.CodeErrUnSafety, constant.ErrTypeNotSafety, errors.New("missing code"))
		return
	}
	codeWithMD5 := encrypt.Md5(code)
	cookieValue, _ := encrypt.StringEncrypt(codeWithMD5)
	c.SetCookie(codeWithMD5, cookieValue, 604800, "", "", false, false)

	helper.SuccessWithData(c, nil)
}
