package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) LoadFirewallSettings(c *gin.Context) {
	data, err := firewallSettingService.Load(c.Request.Context())
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

func (b *BaseApi) OperateFirewallBackend(c *gin.Context) {
	var request dto.FirewallBackendOperation
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	if err := firewallSettingService.Operate(c.Request.Context(), request); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}
