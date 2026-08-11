package v2

import (
	"errors"
	"net/http"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/gin-gonic/gin"
)

// @Tags Firewall
// @Summary List unified firewall v2 rules
// @Accept json
// @Param request body dto.FirewallRuleInventory true "request"
// @Success 200 {object} dto.FirewallRuleInventoryResponse
// @Failure 400 {object} dto.Response
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules/search [post]
func (b *BaseApi) SearchFirewallRules(c *gin.Context) {
	var request dto.FirewallRuleInventory
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	inventory, err := firewallService.Inventory(c.Request.Context(), request)
	if err != nil {
		handleFirewallRuleError(c, err)
		return
	}
	helper.SuccessWithData(c, inventory)
}

// @Tags Firewall
// @Summary Check a unified firewall v2 rule for duplicates and conflicts
// @Accept json
// @Param request body dto.FirewallRuleCheck true "request"
// @Success 200 {object} dto.FirewallRuleCheckResponse
// @Failure 400 {object} dto.Response
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules/check [post]
func (b *BaseApi) CheckFirewallRule(c *gin.Context) {
	var request dto.FirewallRuleCheck
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	result, err := firewallService.Check(c.Request.Context(), c.ClientIP(), request)
	if err != nil {
		handleFirewallRuleError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

// @Tags Firewall
// @Summary Check multiple unified firewall v2 rules
// @Accept json
// @Param request body dto.FirewallRuleBatchCheck true "request"
// @Success 200 {object} dto.FirewallRuleBatchCheckResponse
// @Failure 400 {object} dto.Response
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules/check/batch [post]
func (b *BaseApi) CheckFirewallRulesBatch(c *gin.Context) {
	var request dto.FirewallRuleBatchCheck
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	result, err := firewallService.CheckBatch(c.Request.Context(), c.ClientIP(), request)
	if err != nil {
		handleFirewallRuleError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

// @Tags Firewall
// @Summary Create a unified firewall v2 rule
// @Accept json
// @Param request body dto.FirewallRuleCreate true "request"
// @Success 200
// @Failure 400 {object} dto.Response
// @Failure 409 {object} dto.Response
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"添加防火墙规则","formatEN":"create firewall rule"}
func (b *BaseApi) CreateFirewallRule(c *gin.Context) {
	var request dto.FirewallRuleCreate
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	if err := firewallService.Create(c.Request.Context(), request); err != nil {
		handleFirewallRuleError(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Firewall
// @Summary Create multiple unified firewall v2 rules
// @Accept json
// @Param request body dto.FirewallRuleBatchCreate true "request"
// @Success 200 {object} dto.FirewallRuleBatchCreateResponse
// @Failure 400 {object} dto.Response
// @Failure 409 {object} dto.Response
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules/batch [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"批量添加防火墙规则","formatEN":"batch create firewall rules"}
func (b *BaseApi) CreateFirewallRulesBatch(c *gin.Context) {
	var request dto.FirewallRuleBatchCreate
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	result, err := firewallService.CreateBatch(c.Request.Context(), request)
	if err != nil {
		handleFirewallRuleError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

// @Tags Firewall
// @Summary Delete a managed unified firewall v2 rule
// @Param uuid path string true "managed rule UUID"
// @Success 200
// @Failure 400 {object} dto.Response
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules/{uuid} [delete]
// @x-panel-log {"bodyKeys":[],"paramKeys":["uuid"],"BeforeFunctions":[],"formatZH":"删除防火墙规则 [uuid]","formatEN":"delete firewall rule [uuid]"}
func (b *BaseApi) DeleteFirewallRule(c *gin.Context) {
	request := dto.FirewallRuleDelete{
		UUID: strings.TrimSpace(c.Param("uuid")),
	}
	if request.UUID == "" {
		helper.BadRequest(c, repo.ErrFirewallPersistenceInvalid)
		return
	}
	if err := firewallService.Delete(c.Request.Context(), request); err != nil {
		handleFirewallRuleError(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Firewall
// @Summary Update a managed unified firewall v2 rule
// @Accept json
// @Param uuid path string true "managed rule UUID"
// @Param request body dto.FirewallRuleUpdate true "request"
// @Success 200
// @Failure 400 {object} dto.Response
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules/{uuid} [put]
// @x-panel-log {"bodyKeys":[],"paramKeys":["uuid"],"BeforeFunctions":[],"formatZH":"更新防火墙规则 [uuid]","formatEN":"update firewall rule [uuid]"}
func (b *BaseApi) UpdateFirewallRule(c *gin.Context) {
	var request dto.FirewallRuleUpdate
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	request.UUID = strings.TrimSpace(c.Param("uuid"))
	if request.UUID == "" {
		helper.BadRequest(c, repo.ErrFirewallPersistenceInvalid)
		return
	}
	if err := firewallService.Update(c.Request.Context(), c.ClientIP(), request); err != nil {
		handleFirewallRuleError(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Firewall
// @Summary Reorder a managed unified firewall v2 rule
// @Accept json
// @Param uuid path string true "managed rule UUID"
// @Param request body dto.FirewallRuleReorder true "request"
// @Success 200
// @Failure 400 {object} dto.Response
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules/{uuid}/reorder [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":["uuid"],"BeforeFunctions":[],"formatZH":"调整防火墙规则顺序 [uuid]","formatEN":"reorder firewall rule [uuid]"}
func (b *BaseApi) ReorderFirewallRule(c *gin.Context) {
	var request dto.FirewallRuleReorder
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	request.UUID = strings.TrimSpace(c.Param("uuid"))
	if request.UUID == "" {
		helper.BadRequest(c, repo.ErrFirewallPersistenceInvalid)
		return
	}
	if err := firewallService.Reorder(c.Request.Context(), c.ClientIP(), request); err != nil {
		handleFirewallRuleError(c, err)
		return
	}
	helper.Success(c)
}

func handleFirewallRuleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, filter.ErrLockoutRisk), errors.Is(err, filter.ErrProtectedRule):
		helper.ErrorWithBusinessCode(c, http.StatusBadRequest, "FW_LOCKOUT_RISK", "ErrInvalidParams", err)
	case errors.Is(err, filter.ErrRuleStale):
		helper.ErrorWithBusinessCode(c, http.StatusConflict, "FW_RULE_STALE", "ErrInvalidParams", err)
	case errors.Is(err, filter.ErrRuleCheckRequired):
		helper.ErrorWithBusinessCode(c, http.StatusConflict, "FW_RULE_CHECK_REQUIRED", "ErrInvalidParams", err)
	case errors.Is(err, filter.ErrUnsupportedScope), errors.Is(err, filter.ErrInvalidScope),
		errors.Is(err, filter.ErrProviderUnavailable), errors.Is(err, filter.ErrAdapterUnavailable):
		helper.ErrorWithBusinessCode(c, http.StatusBadRequest, "FW_SCOPE_UNSUPPORTED", "ErrInvalidParams", err)
	case errors.Is(err, filter.ErrInvalidRule), errors.Is(err, filter.ErrRuleOperation),
		errors.Is(err, repo.ErrFirewallPersistenceInvalid):
		helper.ErrorWithBusinessCode(c, http.StatusBadRequest, "FW_RULE_UNSUPPORTED", "ErrInvalidParams", err)
	case errors.Is(err, filter.ErrVerificationFailed):
		helper.ErrorWithBusinessCode(c, http.StatusInternalServerError, "FW_VERIFY_FAILED", "ErrInternalServer", err)
	default:
		helper.ErrorWithBusinessCode(c, http.StatusInternalServerError, "FW_APPLY_FAILED", "ErrInternalServer", err)
	}
}
