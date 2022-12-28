package v1

import (
	"errors"
	"os/exec"
	"runtime"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/mfa"
	"github.com/beevik/ntp"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) GetSettingInfo(c *gin.Context) {
	setting, err := settingService.GetSettingInfo()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, setting)
}

func (b *BaseApi) GetDaemonjson(c *gin.Context) {
	value, err := settingService.GetDaemonjson()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, value)
}

func (b *BaseApi) UpdateSetting(c *gin.Context) {
	var req dto.SettingUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := settingService.Update(c, req.Key, req.Value); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

func (b *BaseApi) UpdatePassword(c *gin.Context) {
	var req dto.PasswordUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := settingService.UpdatePassword(c, req.OldPassword, req.NewPassword); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

func (b *BaseApi) HandlePasswordExpired(c *gin.Context) {
	var req dto.PasswordUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := settingService.HandlePasswordExpired(c, req.OldPassword, req.NewPassword); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

func (b *BaseApi) SyncTime(c *gin.Context) {
	ntime, err := ntp.Time("pool.ntp.org")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	system := runtime.GOOS
	if system == "linux" {
		cmd := exec.Command("date", "-s", ntime.Format("2006-01-02 15:04:05"))
		stdout, err := cmd.CombinedOutput()
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, errors.New(string(stdout)))
			return
		}
	}

	helper.SuccessWithData(c, ntime.Format("2006-01-02 15:04:05 MST -0700"))
}

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

func (b *BaseApi) GetMFA(c *gin.Context) {
	otp, err := mfa.GetOtp("admin")
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, otp)
}

func (b *BaseApi) MFABind(c *gin.Context) {
	var req dto.MfaCredential
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	success := mfa.ValidCode(req.Code, req.Secret)
	if !success {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, errors.New("code is not valid"))
		return
	}

	if err := settingService.Update(c, "MFAStatus", "enable"); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	if err := settingService.Update(c, "MFASecret", req.Secret); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}
