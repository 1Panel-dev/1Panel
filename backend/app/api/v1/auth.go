package v1

import (
	"github.com/1Panel-dev/1Panel/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/utils/captcha"
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
	}
	helper.SuccessWithData(c, captcha)
}
