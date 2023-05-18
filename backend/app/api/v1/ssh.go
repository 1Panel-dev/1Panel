package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/gin-gonic/gin"
)

// @Tags SSH
// @Summary Load host ssh setting info
// @Description 加载 SSH 配置信息
// @Success 200 {object} dto.SSHInfo
// @Security ApiKeyAuth
// @Router /host/ssh/search [post]
func (b *BaseApi) GetSSHInfo(c *gin.Context) {
	info, err := sshService.GetSSHInfo()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, info)
}

// @Tags SSH
// @Summary Update host ssh setting
// @Description 更新 SSH 配置
// @Accept json
// @Param request body dto.SettingUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /host/ssh/update [post]
// @x-panel-log {"bodyKeys":["key","value"],"paramKeys":[],"BeforeFuntions":[],"formatZH":"修改 SSH 配置 [key] => [value]","formatEN":"update SSH setting [key] => [value]"}
func (b *BaseApi) UpdateSSH(c *gin.Context) {
	var req dto.SettingUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := sshService.Update(req.Key, req.Value); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Update host ssh setting by file
// @Description 上传文件更新 SSH 配置
// @Accept json
// @Param request body dto.SSHConf true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /host/conffile/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFuntions":[],"formatZH":"修改 SSH 配置文件","formatEN":"update SSH conf"}
func (b *BaseApi) UpdateSSHByfile(c *gin.Context) {
	var req dto.SSHConf
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := sshService.UpdateByFile(req.File); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Generate host ssh secret
// @Description 生成 ssh 密钥
// @Accept json
// @Param request body dto.GenerateSSH true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /host/ssh/generate [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFuntions":[],"formatZH":"生成 SSH 密钥 ","formatEN":"generate SSH secret"}
func (b *BaseApi) GenerateSSH(c *gin.Context) {
	var req dto.GenerateSSH
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	if err := sshService.GenerateSSH(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Load host ssh secret
// @Description 获取 ssh 密钥
// @Accept json
// @Param request body dto.GenerateLoad true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /host/ssh/secret [post]
func (b *BaseApi) LoadSSHSecret(c *gin.Context) {
	var req dto.GenerateLoad
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	data, err := sshService.LoadSSHSecret(req.EncryptionMode)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags SSH
// @Summary Load host ssh logs
// @Description 获取 ssh 登录日志
// @Accept json
// @Param request body dto.SearchSSHLog true "request"
// @Success 200 {object} dto.SSHLog
// @Security ApiKeyAuth
// @Router /host/ssh/logs [post]
func (b *BaseApi) LoadSSHLogs(c *gin.Context) {
	var req dto.SearchSSHLog
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}
	if err := global.VALID.Struct(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, err)
		return
	}

	data, err := sshService.LoadLog(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}
