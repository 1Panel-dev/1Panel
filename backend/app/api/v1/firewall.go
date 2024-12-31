package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Firewall
// @Summary Load firewall base info
// @Success 200 {object} dto.FirewallBaseInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/base [get]
func (b *BaseApi) LoadFirewallBaseInfo(c *gin.Context) {
	data, err := firewallService.LoadBaseInfo()
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags Firewall
// @Summary Page firewall status
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

	if err := firewallService.OperateFirewall(req.Operation); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Firewall
// @Summary Create firewall port group
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// OperateForwardRule
// @Tags Firewall
// @Summary Update firewall port group
// @Accept json
// @Param request body dto.ForwardRuleOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/forward [post]
// @x-panel-log {"bodyKeys":["source_port"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新端口转发规则 [source_port]","formatEN":"update port forward rules [source_port]"}
func (b *BaseApi) OperateForwardRule(c *gin.Context) {
	var req dto.ForwardRuleOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := firewallService.OperateForwardRule(req); err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Firewall
// @Summary Create firewall group
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Firewall
// @Summary Batch create group
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Firewall
// @Summary Update firewall group
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

// @Tags Firewall
// @Summary Update address group
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
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
