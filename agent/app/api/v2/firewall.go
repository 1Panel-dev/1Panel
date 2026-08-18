package v2

import (
	"errors"
	"net/http"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/gin-gonic/gin"
)

// @Tags Firewall
// @Summary Load firewall base info
// @Accept json
// @Param request body dto.OperationWithName true "request"
// @Success 200 {object} dto.FirewallSubsystemStatus
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/base [post]
func (b *BaseApi) LoadFirewallBaseInfo(c *gin.Context) {
	var request dto.OperationWithName
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}

	data, err := firewallService.LoadBaseInfo(request.Name)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, data)
}

// @Tags Firewall
// @Summary Operate firewall
// @Accept json
// @Param request body dto.FirewallLifecycleOperation true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/operate [post]
// @x-panel-log {"bodyKeys":["operation"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operation] 防火墙","formatEN":"[operation] firewall"}
func (b *BaseApi) OperateFirewall(c *gin.Context) {
	var request dto.FirewallLifecycleOperation
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}

	if err := firewallService.OperateFirewall(request); err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.Success(c)
}

// @Tags Firewall
// @Summary Load forwarding base info
// @Accept json
// @Success 200 {object} dto.FirewallSubsystemStatus
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
	var request dto.ForwardRuleSearch
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	total, items, err := forwardingService.SearchRules(request)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, dto.PageResult{Items: items, Total: total})
}

// @Tags Firewall
// @Summary Operate forwarding rules
// @Accept json
// @Param request body dto.ForwardRuleOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/forward/operate [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新端口转发规则","formatEN":"update port forward rules"}
func (b *BaseApi) OperateForwardingRules(c *gin.Context) {
	var request dto.ForwardRuleOperate
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}

	if err := forwardingService.OperateRules(request); err != nil {
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

// @Tags Firewall
// @Summary Apply/Unload/Init firewall filter chain
// @Accept json
// @Param request body dto.FilterChainOperation true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/filter/operate [post]
// @x-panel-log {"bodyKeys":["operate"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operate] 防火墙过滤链","formatEN":"[operate] firewall filter chain"}
func (b *BaseApi) OperateFilterChain(c *gin.Context) {
	var request dto.FilterChainOperation
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	if err := firewallService.OperateFilterChain(request); err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.Success(c)
}

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
// @Summary Load one provider-native firewall object definition
// @Accept json
// @Param request body dto.FirewallNativeDetail true "request"
// @Success 200 {string} string
// @Failure 400 {object} dto.Response
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules/native/detail [post]
func (b *BaseApi) LoadFirewallNativeDetail(c *gin.Context) {
	var request dto.FirewallNativeDetail
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	info, err := firewallService.LoadFirewallNativeDetail(c.Request.Context(), request)
	if err != nil {
		handleFirewallRuleError(c, err)
		return
	}
	helper.SuccessWithData(c, info)
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
	result, err := firewallService.CheckRulesBatch(c.Request.Context(), c.ClientIP(), request)
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
	result, err := firewallService.CreateRulesBatch(c.Request.Context(), request)
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
// @Summary Delete multiple managed unified firewall v2 rules
// @Accept json
// @Param request body dto.FirewallRuleBatchDelete true "request"
// @Success 200 {object} dto.FirewallRuleBatchDeleteResponse
// @Failure 400 {object} dto.Response
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules/delete/batch [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"批量删除防火墙规则","formatEN":"batch delete firewall rules"}
func (b *BaseApi) DeleteFirewallRulesBatch(c *gin.Context) {
	var request dto.FirewallRuleBatchDelete
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	result, err := firewallService.DeleteRulesBatch(c.Request.Context(), request)
	if err != nil {
		handleFirewallRuleError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
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
	case errors.Is(err, repo.ErrFirewallRuleRevisionConflict):
		helper.ErrorWithBusinessCode(c, http.StatusConflict, "FW_RULE_REVISION_CONFLICT", "ErrInvalidParams", err)
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

func (b *BaseApi) ListDockerPortGuard(c *gin.Context) {
	data, err := dockerPortGuardService.LoadOverview(c.Request.Context())
	if err != nil {
		handleDockerPortGuardError(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

func (b *BaseApi) SyncDockerPortGuard(c *gin.Context) {
	if err := dockerPortGuardService.Reconcile(c.Request.Context()); err != nil {
		handleDockerPortGuardError(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) OperateDockerPortGuard(c *gin.Context) {
	var request dto.DockerPortGuardOperation
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	if err := dockerPortGuardService.Operate(c.Request.Context(), request); err != nil {
		handleDockerPortGuardError(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) DeleteDockerPortGuardPolicies(c *gin.Context) {
	var request dto.DockerPortGuardPolicyBatchDelete
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	if err := dockerPortGuardService.DeletePolicies(c.Request.Context(), request); err != nil {
		handleDockerPortGuardError(c, err)
		return
	}
	helper.Success(c)
}

func (b *BaseApi) UpsertDockerPortGuardPolicies(c *gin.Context) {
	var request dto.DockerPortGuardPolicyBatch
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	if err := dockerPortGuardService.UpsertPolicies(c.Request.Context(), request); err != nil {
		handleDockerPortGuardError(c, err)
		return
	}
	helper.Success(c)
}

func handleDockerPortGuardError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrDockerGuardInvalid) {
		helper.ErrorWithBusinessCode(c, http.StatusBadRequest, "FW_DOCKER_GUARD_INVALID", "ErrInvalidParams", err)
		return
	}
	helper.ErrorWithBusinessCode(c, http.StatusInternalServerError, "FW_DOCKER_GUARD_FAILED", "ErrInternalServer", err)
}
