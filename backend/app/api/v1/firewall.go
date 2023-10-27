package v1

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

// @Tags Firewall
// @Summary Load firewall base info
// @Description 获取防火墙基础信息
// @Success 200 {object} dto.FirewallBaseInfo
// @Security ApiKeyAuth
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
// @Description 获取防火墙规则列表分页
// @Accept json
// @Param request body dto.RuleSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
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
// @Description 修改防火墙状态
// @Accept json
// @Param request body dto.FirewallOperation true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
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
// @Summary Create group
// @Description 创建防火墙端口规则
// @Accept json
// @Param request body dto.PortRuleOperate true "request"
// @Success 200
// @Security ApiKeyAuth
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

// @Tags Firewall
// @Summary Create group
// @Description 创建防火墙 IP 规则
// @Accept json
// @Param request body dto.AddrRuleOperate true "request"
// @Success 200
// @Security ApiKeyAuth
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
// @Summary Create group
// @Description 批量删除防火墙规则
// @Accept json
// @Param request body dto.BatchRuleOperate true "request"
// @Success 200
// @Security ApiKeyAuth
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
// @Description 更新防火墙描述
// @Accept json
// @Param request body dto.UpdateFirewallDescription true "request"
// @Success 200
// @Security ApiKeyAuth
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
// @Summary Create group
// @Description 更新端口防火墙规则
// @Accept json
// @Param request body dto.PortRuleUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
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
// @Summary Create group
// @Description 更新 ip 防火墙规则
// @Accept json
// @Param request body dto.AddrRuleUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
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
