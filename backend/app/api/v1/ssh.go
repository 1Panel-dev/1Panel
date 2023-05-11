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
