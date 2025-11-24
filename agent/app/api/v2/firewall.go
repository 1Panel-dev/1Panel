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
// @Summary Page firewall rules
// @Accept json
// @Param request body dto.RuleSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/search [post]
func (b *BaseApi) SearchFirewallRule(c *gin.Context) {
	var req dto.RuleSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := firewallService.SearchWithPage(req)
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
// @Summary Create group
// @Accept json
// @Param request body dto.PortRuleOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/port [post]
// @x-panel-log {"bodyKeys":["port","strategy"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"添加端口规则 [strategy] [port]","formatEN":"create port rules [strategy][port]"}
func (b *BaseApi) OperatePortRule(c *gin.Context) {
	var req dto.PortRuleOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := firewallService.OperatePortRule(req, true); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// OperateForwardRule
// @Tags Firewall
// @Summary Operate forward rule
// @Accept json
// @Param request body dto.ForwardRuleOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/forward [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新端口转发规则","formatEN":"update port forward rules"}
func (b *BaseApi) OperateForwardRule(c *gin.Context) {
	var req dto.ForwardRuleOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := firewallService.OperateForwardRule(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Firewall
// @Summary Operate Ip rule
// @Accept json
// @Param request body dto.AddrRuleOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/ip [post]
// @x-panel-log {"bodyKeys":["strategy","address"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"添加 ip 规则 [strategy] [address]","formatEN":"create address rules [strategy][address]"}
func (b *BaseApi) OperateIPRule(c *gin.Context) {
	var req dto.AddrRuleOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := firewallService.OperateAddressRule(req, true); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Firewall
// @Summary Batch operate rule
// @Accept json
// @Param request body dto.BatchRuleOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/batch [post]
func (b *BaseApi) BatchOperateRule(c *gin.Context) {
	var req dto.BatchRuleOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := firewallService.BatchOperateRule(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Firewall
// @Summary Update rule description
// @Accept json
// @Param request body dto.UpdateFirewallDescription true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/update/description [post]
func (b *BaseApi) UpdateFirewallDescription(c *gin.Context) {
	var req dto.UpdateFirewallDescription
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := firewallService.UpdateDescription(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Firewall
// @Summary Update port rule
// @Accept json
// @Param request body dto.PortRuleUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/update/port [post]
func (b *BaseApi) UpdatePortRule(c *gin.Context) {
	var req dto.PortRuleUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := firewallService.UpdatePortRule(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Firewall
// @Summary Update Ip rule
// @Accept json
// @Param request body dto.AddrRuleUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/update/addr [post]
func (b *BaseApi) UpdateAddrRule(c *gin.Context) {
	var req dto.AddrRuleUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := firewallService.UpdateAddrRule(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Firewall
// @Summary search iptables filter rules
// @Accept json
// @Param request body dto.SearchPageWithType true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/filter/search [post]
func (b *BaseApi) SearchFilterRules(c *gin.Context) {
	var req dto.SearchPageWithType
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	total, list, err := iptablesService.Search(req)
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
// @Summary Operate iptables filter rule
// @Accept json
// @Param request body dto.IptablesRuleOp true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/filter/rule/operate [post]
// @x-panel-log {"bodyKeys":["operation","chain"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operation] filter规则到 [chain]","formatEN":"[operation] filter rule to [chain]"}
func (b *BaseApi) OperateFilterRule(c *gin.Context) {
	var req dto.IptablesRuleOp
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := iptablesService.OperateRule(req, true); err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.Success(c)
}

// @Tags Firewall
// @Summary Batch operate iptables filter rules
// @Accept json
// @Param request body dto.IptablesBatchOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/filter/rule/batch [post]
func (b *BaseApi) BatchOperateFilterRule(c *gin.Context) {
	var req dto.IptablesBatchOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := iptablesService.BatchOperate(req); err != nil {
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
	if err := iptablesService.Operate(req); err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.Success(c)
}

// @Tags Firewall
// @Summary load chain status with name
// @Accept json
// @Param request body dto.OperationWithName true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/filter/chain/status [post]
func (b *BaseApi) LoadChainStatus(c *gin.Context) {
	var req dto.OperationWithName
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	helper.SuccessWithData(c, iptablesService.LoadChainStatus(req))
}
