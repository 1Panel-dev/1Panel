package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/gin-gonic/gin"
)

// @Tags SSH
// @Summary Load host SSH setting info
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
// @Summary Operate SSH
// @Description 修改 SSH 服务状态
// @Accept json
// @Param request body dto.Operate true "request"
// @Security ApiKeyAuth
// @Router /host/ssh/operate [post]
// @x-panel-log {"bodyKeys":["operation"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operation] SSH ","formatEN":"[operation] SSH"}
func (b *BaseApi) OperateSSH(c *gin.Context) {
	var req dto.Operate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := sshService.OperateSSH(req.Operation); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Update host SSH setting
// @Description 更新 SSH 配置
// @Accept json
// @Param request body dto.SSHUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /host/ssh/update [post]
// @x-panel-log {"bodyKeys":["key","value"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改 SSH 配置 [key] => [value]","formatEN":"update SSH setting [key] => [value]"}
func (b *BaseApi) UpdateSSH(c *gin.Context) {
	var req dto.SSHUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := sshService.Update(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Update host SSH setting by file
// @Description 上传文件更新 SSH 配置
// @Accept json
// @Param request body dto.SSHConf true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /host/conffile/update [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改 SSH 配置文件","formatEN":"update SSH conf"}
func (b *BaseApi) UpdateSSHByfile(c *gin.Context) {
	var req dto.SSHConf
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := sshService.UpdateByFile(req.File); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Generate host SSH secret
// @Description 生成 SSH 密钥
// @Accept json
// @Param request body dto.GenerateSSH true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /host/ssh/generate [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"生成 SSH 密钥 ","formatEN":"generate SSH secret"}
func (b *BaseApi) GenerateSSH(c *gin.Context) {
	var req dto.GenerateSSH
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := sshService.GenerateSSH(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags SSH
// @Summary Load host SSH secret
// @Description 获取 SSH 密钥
// @Accept json
// @Param request body dto.GenerateLoad true "request"
// @Success 200
// @Security ApiKeyAuth
// @Router /host/ssh/secret [post]
func (b *BaseApi) LoadSSHSecret(c *gin.Context) {
	var req dto.GenerateLoad
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
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
// @Summary Load host SSH logs
// @Description 获取 SSH 登录日志
// @Accept json
// @Param request body dto.SearchSSHLog true "request"
// @Success 200 {object} dto.SSHLog
// @Security ApiKeyAuth
// @Router /host/ssh/log [post]
func (b *BaseApi) LoadSSHLogs(c *gin.Context) {
	var req dto.SearchSSHLog
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	data, err := sshService.LoadLog(req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags SSH
// @Summary Load host SSH conf
// @Description 获取 SSH 配置文件
// @Success 200
// @Security ApiKeyAuth
// @Router /host/ssh/conf [get]
func (b *BaseApi) LoadSSHConf(c *gin.Context) {
	data, err := sshService.LoadSSHConf()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}
