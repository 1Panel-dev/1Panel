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
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"初始化并启用端口转发","formatEN":"initialize and enable port forwarding"}
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
// @Summary Check unified firewall v2 rules for duplicates and conflicts
// @Accept json
// @Param request body dto.FirewallRuleCheck true "request"
// @Success 200 {object} dto.FirewallRuleCheckResponse
// @Failure 400 {object} dto.Response
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules/check [post]
func (b *BaseApi) CheckFirewallRules(c *gin.Context) {
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
// @Summary Create unified firewall v2 rules
// @Accept json
// @Param request body dto.FirewallRuleCreate true "request"
// @Success 200 {object} dto.FirewallRuleCreateResponse
// @Failure 400 {object} dto.Response
// @Failure 409 {object} dto.Response
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"添加防火墙规则","formatEN":"create firewall rules"}
func (b *BaseApi) CreateFirewallRules(c *gin.Context) {
	var request dto.FirewallRuleCreate
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	result, err := firewallService.Create(c.Request.Context(), request)
	if err != nil {
		handleFirewallRuleError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

// @Tags Firewall
// @Summary Delete managed unified firewall v2 rules
// @Accept json
// @Param request body dto.FirewallRuleDelete true "request"
// @Success 200 {object} dto.FirewallRuleDeleteResponse
// @Failure 400 {object} dto.Response
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules/delete [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"删除防火墙规则","formatEN":"delete firewall rules"}
func (b *BaseApi) DeleteFirewallRules(c *gin.Context) {
	var request dto.FirewallRuleDelete
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	result, err := firewallService.Delete(c.Request.Context(), request)
	if err != nil {
		handleFirewallRuleError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

// @Tags Firewall
// @Summary Update a managed unified firewall v2 rule
// @Accept json
// @Param request body dto.FirewallRuleUpdate true "request"
// @Success 200
// @Failure 400 {object} dto.Response
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules/update [post]
// @x-panel-log {"bodyKeys":["uuid"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新防火墙规则 [uuid]","formatEN":"update firewall rule [uuid]"}
func (b *BaseApi) UpdateFirewallRule(c *gin.Context) {
	var request dto.FirewallRuleUpdate
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	request.UUID = strings.TrimSpace(request.UUID)
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
// @Param request body dto.FirewallRuleReorder true "request"
// @Success 200
// @Failure 400 {object} dto.Response
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules/reorder [post]
// @x-panel-log {"bodyKeys":["uuid"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"调整防火墙规则顺序 [uuid]","formatEN":"reorder firewall rule [uuid]"}
func (b *BaseApi) ReorderFirewallRule(c *gin.Context) {
	var request dto.FirewallRuleReorder
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	request.UUID = strings.TrimSpace(request.UUID)
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

// @Tags Firewall
// @Summary Load firewall settings
// @Success 200 {object} dto.FirewallSettings
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/settings [get]
func (b *BaseApi) LoadFirewallSettings(c *gin.Context) {
	data, err := firewallSettingService.Load(c.Request.Context())
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Firewall
// @Summary Operate firewall backend
// @Accept json
// @Param request body dto.FirewallBackendOperation true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/settings/operate [post]
// @x-panel-log {"bodyKeys":["subsystem","backend","operation"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"防火墙子系统 [subsystem] 后端 [operation] [backend]","formatEN":"[operation] firewall [subsystem] backend [backend]"}
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

// @Tags Firewall
// @Summary List Docker port guard status and policies
// @Success 200 {object} dto.DockerPortGuardList
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/docker/ports [get]
func (b *BaseApi) ListDockerPortGuard(c *gin.Context) {
	data, err := dockerPortGuardService.LoadOverview(c.Request.Context())
	if err != nil {
		handleDockerPortGuardError(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Firewall
// @Summary Sync Docker port guard rules
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/docker/sync [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"同步 Docker 端口防护规则","formatEN":"sync Docker port guard rules"}
func (b *BaseApi) SyncDockerPortGuard(c *gin.Context) {
	if err := dockerPortGuardService.Reconcile(c.Request.Context()); err != nil {
		handleDockerPortGuardError(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Firewall
// @Summary Operate Docker port guard
// @Accept json
// @Param request body dto.DockerPortGuardOperation true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/docker/operate [post]
// @x-panel-log {"bodyKeys":["operation"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operation] Docker 端口防护","formatEN":"[operation] Docker port guard"}
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

// @Tags Firewall
// @Summary Delete Docker port guard policies
// @Accept json
// @Param request body dto.DockerPortGuardPolicyBatchDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/docker/policies/delete/batch [post]
// @x-panel-log {"bodyKeys":["uuids"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"删除 Docker 端口防护策略 [uuids]","formatEN":"delete Docker port guard policies [uuids]"}
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

// @Tags Firewall
// @Summary Batch upsert Docker port guard policies
// @Accept json
// @Param request body dto.DockerPortGuardPolicyBatch true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/docker/policies/batch [post]
// @x-panel-log {"bodyKeys":["mode"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"批量更新 Docker 端口防护策略 [mode]","formatEN":"batch update Docker port guard policies [mode]"}
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
