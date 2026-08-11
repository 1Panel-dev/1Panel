package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/gin-gonic/gin"
)

// @Tags Firewall
// @Summary Load firewall base info
// @Accept json
// @Param request body dto.OperationWithName true "request"
// @Success 200 {object} dto.FirewallBaseInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/base [post]
func (b *BaseApi) LoadFirewallBaseInfo(c *gin.Context) {
	var req dto.OperationWithName
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	data, err := firewallService.LoadBaseInfo(req.Name)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Firewall
// @Summary Load forwarding base info
// @Accept json
// @Success 200 {object} dto.FirewallBaseInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/forward/base [post]
func (b *BaseApi) LoadForwardingBaseInfo(c *gin.Context) {
	data, err := forwardingService.LoadBaseInfo()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Firewall
// @Summary Page forwarding rules
// @Accept json
// @Param request body dto.ForwardRuleSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/forward/search [post]
func (b *BaseApi) SearchForwardingRules(c *gin.Context) {
	var req dto.ForwardRuleSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, list, err := forwardingService.SearchWithPage(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Firewall
// @Summary Operate firewall
// @Accept json
// @Param request body dto.FirewallOperation true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/operate [post]
// @x-panel-log {"bodyKeys":["operation"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operation] 防火墙","formatEN":"[operation] firewall"}
func (b *BaseApi) OperateFirewall(c *gin.Context) {
	var req dto.FirewallOperation
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := firewallService.OperateFirewall(req); err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.Success(c)
}

// @Tags Firewall
// OperateForwardRule
// @Tags Firewall
// @Summary Operate forward rule
// @Accept json
// @Param request body dto.ForwardRuleOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/forward/operate [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新端口转发规则","formatEN":"update port forward rules"}
func (b *BaseApi) OperateForwardRule(c *gin.Context) {
	var req dto.ForwardRuleOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := forwardingService.Operate(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Firewall
// @Summary Apply/Unload/Init iptables filter
// @Accept json
// @Param request body dto.IptablesOp true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/filter/operate [post]
// @x-panel-log {"bodyKeys":["operate"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operate] iptables filter 防火墙","formatEN":"[operate] iptables filter firewall"}
func (b *BaseApi) OperateFilterChain(c *gin.Context) {
	var req dto.IptablesOp
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := firewallService.OperateFilterChain(req); err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.Success(c)
}

// @Tags Firewall
// @Summary Enable forwarding
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/forward/enable [post]
func (b *BaseApi) EnableForwarding(c *gin.Context) {
	if err := forwardingService.Enable(); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}
