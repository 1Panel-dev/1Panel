package v2

import (
	"errors"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) UpdatePanelFirewallPort(c *gin.Context) {
	if !global.IsMaster {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	var request struct {
		OldPort uint `json:"oldPort" validate:"required,min=1,max=65535"`
		NewPort uint `json:"newPort" validate:"required,min=1,max=65535"`
	}
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	if err := firewallService.UpdatePanelPort(c.Request.Context(), request.OldPort, request.NewPort); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

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
// @Success 200 {object} dto.FirewallLifecycleOperationResponse
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/operate [post]
// @x-panel-log {"bodyKeys":["operation"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operation] 防火墙","formatEN":"[operation] firewall"}
func (b *BaseApi) OperateFirewall(c *gin.Context) {
	var request dto.FirewallLifecycleOperation
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}

	result, err := firewallService.QueueFirewallOperation(request)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}

	helper.SuccessWithData(c, result)
}

// @Tags Firewall
// @Summary Load forwarding base info
// @Accept json
// @Success 200 {object} dto.FirewallSubsystemStatus
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/forward/base [post]
func (b *BaseApi) LoadForwardingBaseInfo(c *gin.Context) {
	data, err := forwardingService.LoadBaseInfo(c.Request.Context())
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
	total, items, err := forwardingService.SearchRules(c.Request.Context(), request)
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
// @Success 200 {object} dto.FilterChainOperationResponse
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/forward/operate [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新端口转发规则","formatEN":"update port forward rules"}
func (b *BaseApi) OperateForwardingRules(c *gin.Context) {
	var request dto.ForwardRuleOperate
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}

	result, err := forwardingService.OperateRules(request)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

// @Tags Firewall
// @Summary Enable forwarding
// @Accept json
// @Param request body dto.FirewallInitializationTask true "request"
// @Success 200 {object} dto.FilterChainOperationResponse
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/forward/enable [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"初始化并启用端口转发","formatEN":"initialize and enable port forwarding"}
func (b *BaseApi) EnableForwarding(c *gin.Context) {
	var request dto.FirewallInitializationTask
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	result, err := forwardingService.QueueInitialization(request)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

// @Tags Firewall
// @Summary Apply/Unload/Init firewall filter chain
// @Accept json
// @Param request body dto.FilterChainOperation true "request"
// @Success 200 {object} dto.FilterChainOperationResponse
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/filter/operate [post]
// @x-panel-log {"bodyKeys":["operate"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operate] 防火墙过滤链","formatEN":"[operate] firewall filter chain"}
func (b *BaseApi) OperateFilterChain(c *gin.Context) {
	var request dto.FilterChainOperation
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	if request.Operate == "init-base" {
		result, err := firewallService.QueueFilterChainInitialization(request)
		if err != nil {
			helper.InternalServer(c, err)
			return
		}
		helper.SuccessWithData(c, result)
		return
	}
	if err := firewallService.OperateFilterChain(request); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, dto.FilterChainOperationResponse{})
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
// @Summary Reset firewall rules
// @Accept json
// @Param request body dto.FirewallRuleReset true "request"
// @Success 200 {object} dto.FirewallRuleResetResponse
// @Failure 400 {object} dto.Response
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules/reset [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"重置防火墙规则","formatEN":"reset firewall rules"}
func (b *BaseApi) ResetFirewallRules(c *gin.Context) {
	var request dto.FirewallRuleReset
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	result, err := firewallService.Reset(c.Request.Context(), request)
	if err != nil {
		handleFirewallRuleError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
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
// @Summary Queue firewall rule creation
// @Description Creation and import return a taskID immediately; validation and execution results are written to the task log.
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
// @Summary Queue firewall rule deletion
// @Description Deletes non-whitelist rules by scope and instance key. Returns a taskID immediately; results are written to the task log.
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
// @Summary Update a firewall rule
// @Accept json
// @Param request body dto.FirewallRuleUpdate true "request"
// @Success 200
// @Failure 400 {object} dto.Response
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules/update [post]
// @x-panel-log {"bodyKeys":["instanceKey"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"更新防火墙规则 [instanceKey]","formatEN":"update firewall rule [instanceKey]"}
func (b *BaseApi) UpdateFirewallRule(c *gin.Context) {
	var request dto.FirewallRuleUpdate
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	if err := firewallService.Update(c.Request.Context(), request); err != nil {
		handleFirewallRuleError(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Firewall
// @Summary Reorder a firewall rule
// @Accept json
// @Param request body dto.FirewallRuleReorder true "request"
// @Success 200
// @Failure 400 {object} dto.Response
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules/reorder [post]
// @x-panel-log {"bodyKeys":["instanceKey"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"调整防火墙规则顺序 [instanceKey]","formatEN":"reorder firewall rule [instanceKey]"}
func (b *BaseApi) ReorderFirewallRule(c *gin.Context) {
	var request dto.FirewallRuleReorder
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	if err := firewallService.Reorder(c.Request.Context(), request); err != nil {
		handleFirewallRuleError(c, err)
		return
	}
	helper.Success(c)
}

func handleFirewallRuleError(c *gin.Context, err error) {
	var businessErr buserr.BusinessError
	isBusinessError := errors.As(err, &businessErr)
	switch {
	case errors.Is(err, filter.ErrProtectedRule):
		helper.ErrorWithBusinessCode(c, http.StatusBadRequest, "FW_LOCKOUT_RISK", "ErrInvalidParams", err)
	case errors.Is(err, filter.ErrRuleStale):
		helper.ErrorWithBusinessCode(c, http.StatusConflict, "FW_RULE_STALE", "ErrInvalidParams", err)
	case errors.Is(err, filter.ErrUnsupportedScope), errors.Is(err, filter.ErrInvalidScope),
		errors.Is(err, filter.ErrProviderUnavailable), errors.Is(err, filter.ErrAdapterUnavailable):
		helper.ErrorWithBusinessCode(c, http.StatusBadRequest, "FW_SCOPE_UNSUPPORTED", "ErrInvalidParams", err)
	case errors.Is(err, filter.ErrInvalidRule), errors.Is(err, filter.ErrRuleOperation):
		helper.ErrorWithBusinessCode(c, http.StatusBadRequest, "FW_RULE_UNSUPPORTED", "ErrInvalidParams", err)
	case isBusinessError && businessErr.Msg == "ErrRecordExist":
		c.JSON(http.StatusOK, dto.Response{Code: http.StatusConflict, ErrorCode: "FW_RULE_DUPLICATE", Message: err.Error()})
		c.Abort()
	case isBusinessError && businessErr.Msg == "ErrFirewallRuleConflict":
		c.JSON(http.StatusOK, dto.Response{Code: http.StatusConflict, ErrorCode: "FW_RULE_CONFLICT", Message: err.Error()})
		c.Abort()
	case isBusinessError && businessErr.Msg == "ErrInvalidParams":
		c.JSON(http.StatusOK, dto.Response{Code: http.StatusBadRequest, ErrorCode: "FW_RULE_UNSUPPORTED", Message: err.Error()})
		c.Abort()
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
// @Summary Create firewall port whitelist rules
// @Description Saves whitelist configuration and applies missing allowances; existing rules are not removed.
// @Accept json
// @Param request body dto.FirewallPortWhitelistCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/settings/whitelist [post]
// @x-panel-log {"bodyKeys":["rule"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"创建防火墙端口白名单","formatEN":"create firewall port whitelist"}
func (b *BaseApi) CreateFirewallPortWhitelist(c *gin.Context) {
	var request dto.FirewallPortWhitelistCreate
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	if err := firewallSettingService.CreatePortWhitelist(c.Request.Context(), request); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Firewall
// @Summary Update firewall port whitelist rules
// @Description Saves whitelist configuration and applies missing allowances; existing rules are not removed.
// @Accept json
// @Param request body dto.FirewallPortWhitelistUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/settings/whitelist/update [post]
// @x-panel-log {"bodyKeys":["oldRule","rule"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"编辑防火墙端口白名单","formatEN":"update firewall port whitelist"}
func (b *BaseApi) UpdateFirewallPortWhitelist(c *gin.Context) {
	var request dto.FirewallPortWhitelistUpdate
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	if err := firewallSettingService.UpdatePortWhitelist(c.Request.Context(), request); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Firewall
// @Summary Delete firewall port whitelist rules
// @Description Removes whitelist configuration; existing firewall rules are not removed.
// @Accept json
// @Param request body dto.FirewallPortWhitelistDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/settings/whitelist/delete [post]
// @x-panel-log {"bodyKeys":["rules"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"删除防火墙端口白名单","formatEN":"delete firewall port whitelist"}
func (b *BaseApi) DeleteFirewallPortWhitelist(c *gin.Context) {
	var request dto.FirewallPortWhitelistDelete
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	if err := firewallSettingService.DeletePortWhitelist(c.Request.Context(), request); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
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
		var businessErr buserr.BusinessError
		if errors.As(err, &businessErr) && businessErr.Msg == "ErrFirewallBackendCleanupRequired" {
			c.JSON(http.StatusOK, dto.Response{Code: http.StatusConflict, ErrorCode: "FW_BACKEND_CLEANUP_REQUIRED", Message: err.Error()})
			c.Abort()
			return
		}
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
// @Summary List Docker published ports
// @Success 200 {array} dto.DockerPortGuardContainer
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/docker/endpoints [get]
func (b *BaseApi) ListDockerPublishedPorts(c *gin.Context) {
	data, err := dockerPortGuardService.LoadPublishedPorts(c.Request.Context())
	if err != nil {
		handleDockerPortGuardError(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Firewall
// @Summary Operate Docker port guard
// @Accept json
// @Param request body dto.DockerPortGuardOperation true "request"
// @Success 200 {object} dto.FilterChainOperationResponse
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/docker/operate [post]
// @x-panel-log {"bodyKeys":["operation"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operation] Docker 端口防护","formatEN":"[operation] Docker port guard"}
func (b *BaseApi) OperateDockerPortGuard(c *gin.Context) {
	var request dto.DockerPortGuardOperation
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	if request.Operation == "initialize" {
		result, err := dockerPortGuardService.QueueInitialization(request)
		if err != nil {
			handleDockerPortGuardError(c, err)
			return
		}
		helper.SuccessWithData(c, result)
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
// @Success 200 {object} dto.FilterChainOperationResponse
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/docker/policies/delete/batch [post]
// @x-panel-log {"bodyKeys":["uuids"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"删除 Docker 端口防护策略 [uuids]","formatEN":"delete Docker port guard policies [uuids]"}
func (b *BaseApi) DeleteDockerPortGuardPolicies(c *gin.Context) {
	var request dto.DockerPortGuardPolicyBatchDelete
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	result, err := dockerPortGuardService.DeletePolicies(request)
	if err != nil {
		handleDockerPortGuardError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

// @Tags Firewall
// @Summary Batch upsert Docker port guard policies
// @Accept json
// @Param request body dto.DockerPortGuardPolicyBatch true "request"
// @Success 200 {object} dto.FilterChainOperationResponse
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/docker/policies/batch [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"批量更新 Docker 端口防护策略","formatEN":"batch update Docker port guard policies"}
func (b *BaseApi) UpsertDockerPortGuardPolicies(c *gin.Context) {
	var request dto.DockerPortGuardPolicyBatch
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	result, err := dockerPortGuardService.UpsertPolicies(request)
	if err != nil {
		handleDockerPortGuardError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

func handleDockerPortGuardError(c *gin.Context, err error) {
	var businessErr buserr.BusinessError
	if errors.As(err, &businessErr) {
		code, errorCode := http.StatusInternalServerError, ""
		switch businessErr.Msg {
		case "ErrDockerIptablesChainUnavailable":
			code, errorCode = http.StatusServiceUnavailable, "FW_DOCKER_IPTABLES_CHAIN_UNAVAILABLE"
		case "ErrDockerNftablesChainUnavailable":
			code, errorCode = http.StatusServiceUnavailable, "FW_DOCKER_NFTABLES_CHAIN_UNAVAILABLE"
		case "ErrInvalidParams":
			code, errorCode = http.StatusBadRequest, "FW_DOCKER_GUARD_INVALID"
		case "ErrDockerFailed":
			code, errorCode = http.StatusServiceUnavailable, "FW_DOCKER_UNAVAILABLE"
		}
		if errorCode != "" {
			c.JSON(http.StatusOK, dto.Response{Code: code, ErrorCode: errorCode, Message: err.Error()})
			c.Abort()
			return
		}
	}
	if errors.Is(err, docker.ErrUnavailable) {
		helper.ErrorWithBusinessCode(c, http.StatusServiceUnavailable, "FW_DOCKER_UNAVAILABLE", "ErrDockerFailed", err)
		return
	}
	helper.ErrorWithBusinessCode(c, http.StatusInternalServerError, "FW_DOCKER_GUARD_FAILED", "ErrInternalServer", err)
}

// @Tags Firewall
// @Summary List firewall rule backups
// @Param subsystem query string false "Firewall subsystem" Enums(system,forwarding,docker) default(system)
// @Success 200 {object} dto.FirewallRuleBackups
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/rules/backups [get]
func (b *BaseApi) ListFirewallRuleBackups(c *gin.Context) {
	result, err := firewallService.ListRuleBackups(c.Request.Context(), c.DefaultQuery("subsystem", "system"))
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

// @Tags Firewall
// @Summary Initialize, repair or bind one firewall address family
// @Accept json
// @Param request body dto.FirewallFamilyOperation true "request"
// @Success 200 {object} dto.FilterChainOperationResponse
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/family/operate [post]
// @x-panel-log {"bodyKeys":["subsystem","family","operation"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"[operation] [subsystem] [family] 防火墙链","formatEN":"[operation] [subsystem] [family] firewall chains"}
func (b *BaseApi) OperateFirewallFamily(c *gin.Context) {
	var request dto.FirewallFamilyOperation
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	result, err := firewallSettingService.OperateFamily(request)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

// @Tags Firewall
// @Summary Update firewall IPv6 support
// @Accept json
// @Param request body dto.FirewallIPv6Operation true "request"
// @Success 200 {object} dto.FilterChainOperationResponse
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /hosts/firewall/settings/ipv6 [post]
// @x-panel-log {"bodyKeys":["status"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"设置防火墙 IPv6 支持为 [status]","formatEN":"Set firewall IPv6 support to [status]"}
func (b *BaseApi) OperateFirewallIPv6(c *gin.Context) {
	var request dto.FirewallIPv6Operation
	if err := helper.CheckBindAndValidate(&request, c); err != nil {
		return
	}
	result, err := firewallSettingService.OperateIPv6(request)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}
