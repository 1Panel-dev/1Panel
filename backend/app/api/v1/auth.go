package v1

import (
	"errors"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/captcha"
	"github.com/1Panel-dev/1Panel/backend/utils/qqwry"
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
	go saveLoginLogs(c, err)
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
	if err := authService.SafeEntrance(c, code); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrUnSafety, constant.ErrTypeNotSafety, errors.New("missing code"))
		return
	}

	helper.SuccessWithData(c, nil)
}

func (b *BaseApi) CheckIsFirstLogin(c *gin.Context) {
	helper.SuccessWithData(c, authService.CheckIsFirst())
}

func (b *BaseApi) InitUserInfo(c *gin.Context) {
	var req dto.InitUser
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := authService.InitUser(c, req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
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
	qqWry, err := qqwry.NewQQwry()
	if err != nil {
		global.LOG.Errorf("load qqwry datas failed: %s", err)
	}
	res := qqWry.Find(logs.IP)
	logs.Agent = c.GetHeader("User-Agent")
	logs.Address = res.Area
	_ = logService.CreateLoginLog(logs)
}
