package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	filterfirewalld "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/firewalld"
	filteriptables "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/iptables"
	filterufw "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/ufw"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/ping"
	"gorm.io/gorm"
)

type FirewallService struct {
	rules            repo.IFirewallRuleRepo
	adapters         firewallRuleRuntimeRegistry
	selectedProvider func(context.Context) (filter.Provider, error)
	iptablesHelper   *iptables_helper.IptablesManager
}

var firewallRuleMutationMu sync.Mutex

type IFirewallService interface {
	LoadBaseInfo(tab string) (dto.FirewallBaseInfo, error)
	OperateFirewall(req dto.FirewallOperation) error
	OperateFilterChain(req dto.IptablesOp) error
	Inventory(context.Context, dto.FirewallRuleInventory) (dto.FirewallRuleInventoryResponse, error)
	LoadFirewallNativeDetail(context.Context, dto.FirewallNativeDetail) (string, error)
	Check(context.Context, string, dto.FirewallRuleCheck) (dto.FirewallRuleCheckResponse, error)
	CheckBatch(context.Context, string, dto.FirewallRuleBatchCheck) (dto.FirewallRuleBatchCheckResponse, error)
	Create(context.Context, dto.FirewallRuleCreate) error
	CreateBatch(context.Context, dto.FirewallRuleBatchCreate) (dto.FirewallRuleBatchCreateResponse, error)
	Delete(context.Context, dto.FirewallRuleDelete) error
	Update(context.Context, string, dto.FirewallRuleUpdate) error
	Reorder(context.Context, string, dto.FirewallRuleReorder) error
}

func NewIFirewallService() IFirewallService {
	return newFirewallService()
}

func newFirewallService() *FirewallService {
	return &FirewallService{
		rules:            repo.NewIFirewallRuleRepo(),
		adapters:         newFirewallRuleRuntimeRegistry(firewallRuleSnapshotPolicy),
		selectedProvider: firewallRuleSelectedProvider,
		iptablesHelper:   newIptablesHelperManager(),
	}
}

func (u *FirewallService) LoadBaseInfo(tab string) (dto.FirewallBaseInfo, error) {
	var baseInfo dto.FirewallBaseInfo
	baseInfo.Version = "-"
	baseInfo.Name = "-"
	client, err := lifecycle.NewClient()
	if err != nil {
		global.LOG.Errorf("load firewall failed, err: %v", err)
		baseInfo.IsExist = false
		return baseInfo, nil
	}
	baseInfo.IsExist = true
	status, err := lifecycle.LoadStatus(client)
	if err != nil {
		return baseInfo, err
	}
	isInit, isBind, err := iptables_helper.LoadInitStatus(status.Name, tab)
	if err != nil {
		return baseInfo, err
	}
	baseInfo.Name, baseInfo.Version, baseInfo.PingStatus = status.Name, status.Version, ping.LoadStatus()
	baseInfo.IsActive, baseInfo.IsInit, baseInfo.IsBind = status.IsActive, isInit, isBind
	return baseInfo, nil
}

func (u *FirewallService) OperateFirewall(req dto.FirewallOperation) error {
	switch req.Operation {
	case "disableBanPing":
		if err := ping.UpdateStatus("0"); err != nil {
			return err
		}
		return settingRepo.Update("BanPing", constant.StatusDisable)
	case "enableBanPing":
		if err := ping.UpdateStatus("1"); err != nil {
			return err
		}
		return settingRepo.Update("BanPing", constant.StatusEnable)
	}
	client, err := lifecycle.NewClient()
	if err != nil {
		return err
	}
	return lifecycle.NewOperator(client).Operate(lifecycle.Operation(req.Operation), req.WithDockerRestart, u.addPortsBeforeStart)
}

func (u *FirewallService) OperateFilterChain(req dto.IptablesOp) error {
	operation := iptables_helper.BaseOperation(req.Operate)
	if err := u.iptablesHelper.Operate(operation); err != nil {
		return err
	}
	if operation != iptables_helper.BaseOperationInit && operation != iptables_helper.BaseOperationBind {
		return nil
	}
	ports, err := loadConfiguredFirewallPortWhiteList()
	if err != nil {
		return err
	}
	required, err := loadRequiredFirewallPortWhiteList()
	if err != nil {
		return err
	}
	return u.SyncSystemPorts(context.Background(), nil, systemPorts(excludeFirewallPorts(ports, required)))
}

func (s *FirewallService) Inventory(ctx context.Context, request dto.FirewallRuleInventory) (dto.FirewallRuleInventoryResponse, error) {
	scope := request.Scope.Normalize()
	if err := scope.ValidateMVP(); err != nil {
		return dto.FirewallRuleInventoryResponse{}, err
	}
	if err := s.checkSelectedProvider(ctx, scope.Provider); err != nil {
		return dto.FirewallRuleInventoryResponse{}, err
	}
	runtime, err := s.adapters.Resolve(scope.Provider)
	if err != nil {
		return dto.FirewallRuleInventoryResponse{}, err
	}
	snapshot, err := runtime.Observe(ctx, scope)
	if err != nil {
		return dto.FirewallRuleInventoryResponse{}, err
	}
	stored, err := s.rules.List(ctx,
		repo.WithFirewallRuleScope(scope.Key()),
	)
	if err != nil {
		return dto.FirewallRuleInventoryResponse{}, err
	}
	items, err := mergeFirewallInventory(snapshot.Rules, stored, nil, nil)
	if err != nil {
		return dto.FirewallRuleInventoryResponse{}, err
	}
	return dto.FirewallRuleInventoryResponse{Items: items, Notices: snapshot.Notices}, nil
}

func (s *FirewallService) LoadFirewallNativeDetail(ctx context.Context, request dto.FirewallNativeDetail) (string, error) {
	provider := filter.Provider(strings.ToLower(strings.TrimSpace(string(request.Provider))))
	nativeKind := filter.NativeKind(strings.ToLower(strings.TrimSpace(string(request.NativeKind))))
	switch provider {
	case filter.ProviderFirewalld:
		if nativeKind != filter.NativeKindZoneService {
			return "", fmt.Errorf("%w: firewalld detail kind %q", filter.ErrInvalidRule, nativeKind)
		}
	case filter.ProviderUFW:
		if nativeKind != filter.NativeKindUFWApplication {
			return "", fmt.Errorf("%w: UFW detail kind %q", filter.ErrInvalidRule, nativeKind)
		}
	default:
		return "", fmt.Errorf("%w: native details for %s", filter.ErrUnsupportedScope, provider)
	}
	if err := s.checkSelectedProvider(ctx, provider); err != nil {
		return "", err
	}
	runtime, err := s.adapters.Resolve(provider)
	if err != nil {
		return "", err
	}
	informer, ok := runtime.adapter.(filter.NativeDetailReader)
	if !ok {
		return "", fmt.Errorf("%w: native details for %s", filter.ErrAdapterUnavailable, provider)
	}
	return informer.NativeDetail(ctx, request.Name, request.Permanent)
}

func (s *FirewallService) Check(ctx context.Context, clientIP string, request dto.FirewallRuleCheck) (dto.FirewallRuleCheckResponse, error) {
	if strings.TrimSpace(request.UUID) != "" {
		return s.checkUpdate(ctx, clientIP, strings.TrimSpace(request.UUID), request.Rule)
	}
	rule, err := filter.NormalizeRule(request.Rule)
	if err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	if err := s.checkSelectedProvider(ctx, rule.Scope.Provider); err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	runtime, err := s.adapters.Resolve(rule.Scope.Provider)
	if err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	rule, err = runtime.Prepare(rule)
	if err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	if err := runtime.CheckRule(ctx, rule); err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	snapshot, err := runtime.ObserveMutation(ctx, rule.Scope)
	if err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	stored, err := s.rules.List(ctx,
		repo.WithFirewallRuleScope(rule.Scope.Key()),
	)
	if err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	desired, err := desiredFirewallRulesFromModels(stored)
	if err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	result, err := filter.CheckCreate(snapshot, rule, desired, clientIP)
	if err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	managedRevision, err := firewallManagedRevision(stored)
	if err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	checkFlag, err := signFirewallRuleCheck(result, snapshot, managedRevision)
	if err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	return dto.FirewallRuleCheckResponse{
		Decision: result.Decision, Classification: result.Classification, Reason: result.Reason,
		RequestedRule: result.RequestedRule, RequestedRuleKey: result.RequestedRuleKey,
		ExistingRuleUUID: result.ExistingRuleUUID, Candidates: result.Candidates,
		AllowedActions: result.AllowedActions, CheckFlag: checkFlag,
	}, nil
}

func (s *FirewallService) checkUpdate(
	ctx context.Context,
	clientIP string,
	ruleUUID string,
	requestedRule filter.FirewallRule,
) (dto.FirewallRuleCheckResponse, error) {
	stored, before, snapshot, observed, runtime, err := s.loadManagedMutation(ctx, ruleUUID)
	if err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	after, err := filter.NormalizeRule(requestedRule)
	if err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	after.UUID = stored.UUID
	after, err = runtime.Prepare(after)
	if err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	if err := runtime.CheckRule(ctx, after); err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	if after.Scope.Key() != before.Rule.Scope.Key() {
		return dto.FirewallRuleCheckResponse{}, fmt.Errorf("%w: managed rule scope cannot be changed", filter.ErrUnsupportedScope)
	}
	if after.NativeKind != before.Rule.NativeKind {
		return dto.FirewallRuleCheckResponse{}, fmt.Errorf("%w: native rule conversion requires an explicit workflow", filter.ErrUnsupportedScope)
	}
	capabilities, err := runtime.Capabilities(ctx)
	if err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	if capabilities.ExplicitPosition || capabilities.OwnedChains {
		if observed.Locator.Position == nil {
			return dto.FirewallRuleCheckResponse{}, fmt.Errorf("%w: managed rule has no positional locator", filter.ErrInvalidRule)
		}
		currentPosition := int64(*observed.Locator.Position)
		if after.OrderIndex == nil {
			after.OrderIndex = &currentPosition
		} else if *after.OrderIndex != currentPosition {
			if err := validatePositionTarget(ctx, runtime, snapshot, before.Rule, *after.OrderIndex); err != nil {
				return dto.FirewallRuleCheckResponse{}, err
			}
		}
	}
	if err := filter.GuardMutation(snapshot, observed, after, clientIP); err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	semantic, err := firewallRuleSemanticModel(after)
	if err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	if err := s.ensureFirewallRuleIdentityAvailable(ctx, semantic.ScopeKey, semantic.RuleKey, "", stored.UUID); err != nil {
		return dto.FirewallRuleCheckResponse{}, err
	}
	return dto.FirewallRuleCheckResponse{
		Decision:         filter.CheckDecisionReady,
		Classification:   filter.CheckClassificationNone,
		Reason:           "update_ready",
		RequestedRule:    after,
		RequestedRuleKey: semantic.RuleKey,
	}, nil
}

func (s *FirewallService) CheckBatch(
	ctx context.Context,
	clientIP string,
	request dto.FirewallRuleBatchCheck,
) (dto.FirewallRuleBatchCheckResponse, error) {
	response := dto.FirewallRuleBatchCheckResponse{Items: make([]dto.FirewallRuleCheckResponse, 0, len(request.Rules))}
	for _, rule := range request.Rules {
		result, err := s.Check(ctx, clientIP, dto.FirewallRuleCheck{Rule: rule})
		if err != nil {
			return dto.FirewallRuleBatchCheckResponse{}, err
		}
		response.Items = append(response.Items, result)
	}
	return response, nil
}

func (s *FirewallService) Create(ctx context.Context, request dto.FirewallRuleCreate) error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	prepared, err := s.prepareCreate(ctx, request)
	if err != nil {
		return err
	}
	return s.createRule(ctx, prepared.runtime, prepared.snapshot, prepared.request, prepared.authorization)
}

type preparedFirewallRuleCreate struct {
	request       dto.FirewallRuleCreate
	runtime       *firewallRuleRuntime
	snapshot      filter.Snapshot
	authorization firewallRuleCreateAuthorization
}

func (s *FirewallService) prepareCreate(ctx context.Context, request dto.FirewallRuleCreate) (preparedFirewallRuleCreate, error) {
	if request.CheckFlag == "" {
		return preparedFirewallRuleCreate{}, filter.ErrRuleCheckRequired
	}
	rule, err := filter.NormalizeRule(request.Rule)
	if err != nil {
		return preparedFirewallRuleCreate{}, err
	}
	if err := s.checkSelectedProvider(ctx, rule.Scope.Provider); err != nil {
		return preparedFirewallRuleCreate{}, err
	}
	runtime, err := s.adapters.Resolve(rule.Scope.Provider)
	if err != nil {
		return preparedFirewallRuleCreate{}, err
	}
	rule, err = runtime.Prepare(rule)
	if err != nil {
		return preparedFirewallRuleCreate{}, err
	}
	if err := runtime.CheckRule(ctx, rule); err != nil {
		return preparedFirewallRuleCreate{}, err
	}
	snapshot, err := runtime.ObserveMutation(ctx, rule.Scope)
	if err != nil {
		return preparedFirewallRuleCreate{}, err
	}
	stored, err := s.rules.List(ctx,
		repo.WithFirewallRuleScope(rule.Scope.Key()),
	)
	if err != nil {
		return preparedFirewallRuleCreate{}, err
	}
	managedRevision, err := firewallManagedRevision(stored)
	if err != nil {
		return preparedFirewallRuleCreate{}, err
	}
	authorization, err := authorizeFirewallRuleCreate(
		request.CheckFlag, request.Action, request.AdoptInstanceKey, rule, snapshot, managedRevision,
	)
	if err != nil {
		return preparedFirewallRuleCreate{}, err
	}
	sourceKind := request.SourceKind
	if sourceKind == "" {
		sourceKind = constant.FirewallRuleSourceUser
	}
	request.Rule = rule
	request.SourceKind = sourceKind
	return preparedFirewallRuleCreate{
		request: request, runtime: runtime, snapshot: snapshot, authorization: authorization,
	}, nil
}

func (s *FirewallService) CreateBatch(
	ctx context.Context,
	request dto.FirewallRuleBatchCreate,
) (dto.FirewallRuleBatchCreateResponse, error) {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()

	prepared := make([]preparedFirewallRuleCreate, 0, len(request.Items))
	for index, item := range request.Items {
		entry, err := s.prepareCreate(ctx, item)
		if err != nil {
			return firewallBatchPrepareFailure(request.Items, index, err), nil
		}
		prepared = append(prepared, entry)
	}

	result := dto.FirewallRuleBatchCreateResponse{}
	for index, entry := range prepared {
		snapshot, err := entry.runtime.ObserveMutation(ctx, entry.request.Rule.Scope)
		if err == nil {
			entry.authorization, err = refreshBatchCreateAuthorization(snapshot, entry)
		}
		if err == nil {
			err = s.createRule(ctx, entry.runtime, snapshot, entry.request, entry.authorization)
		}
		if err != nil {
			failed := firewallBatchExecutionFailure(request.Items, index, err)
			result.Failed += failed.Failed
			result.Skipped += failed.Skipped
			result.Errors = append(result.Errors, failed.Errors...)
			if global.LOG != nil {
				global.LOG.Errorf("batch create firewall rule item %d failed: %v", index+1, err)
			}
			break
		}
		result.Succeeded++
	}
	return result, nil
}

func firewallBatchPrepareFailure(
	items []dto.FirewallRuleCreate,
	failedIndex int,
	cause error,
) dto.FirewallRuleBatchCreateResponse {
	result := dto.FirewallRuleBatchCreateResponse{
		Failed: 1, Skipped: len(items) - 1,
		Errors: make([]dto.FirewallRuleBatchCreateFailure, 0, len(items)),
	}
	for index := range items {
		failure := dto.FirewallRuleBatchCreateFailure{
			Index: index, Status: "skipped", Rule: items[index].Rule,
		}
		if index == failedIndex {
			failure.Status = "failed"
			failure.Error = cause.Error()
		}
		result.Errors = append(result.Errors, failure)
	}
	return result
}

func firewallBatchExecutionFailure(
	items []dto.FirewallRuleCreate,
	failedIndex int,
	cause error,
) dto.FirewallRuleBatchCreateResponse {
	result := dto.FirewallRuleBatchCreateResponse{
		Failed: 1, Skipped: len(items) - failedIndex - 1,
		Errors: make([]dto.FirewallRuleBatchCreateFailure, 0, len(items)-failedIndex),
	}
	for index := failedIndex; index < len(items); index++ {
		failure := dto.FirewallRuleBatchCreateFailure{
			Index: index, Status: "skipped", Rule: items[index].Rule,
		}
		if index == failedIndex {
			failure.Status = "failed"
			failure.Error = cause.Error()
		}
		result.Errors = append(result.Errors, failure)
	}
	return result
}

func (s *FirewallService) Delete(ctx context.Context, request dto.FirewallRuleDelete) error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	return s.deleteRule(ctx, request.UUID)
}

func (s *FirewallService) Update(ctx context.Context, clientIP string, request dto.FirewallRuleUpdate) error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	rule := request.Rule
	rule.UUID = request.UUID
	return s.updateRule(ctx, clientIP, request.UUID, rule)
}

func (s *FirewallService) Reorder(ctx context.Context, clientIP string, request dto.FirewallRuleReorder) error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	return s.reorderRule(ctx, clientIP, request.UUID, request.TargetPosition, request.Priority)
}

func OperateFirewallPort(oldPorts, newPorts []int) error {
	client, err := lifecycle.NewClient()
	if err != nil {
		return err
	}
	state, err := lifecycle.LoadState(client)
	if err != nil {
		return err
	}
	if state.Name == "iptables" {
		isInit, _, err := iptables_helper.LoadInitStatus(state.Name, "base")
		if err != nil {
			return err
		}
		if !isInit {
			return nil
		}
		if err := newIptablesHelperManager().SyncRequiredPorts(true); err != nil {
			return err
		}
	} else if !state.IsActive {
		return nil
	}
	current, err := loadFirewallPortWhiteList()
	if err != nil {
		return err
	}
	currentSet := iptables_helper.PortWhitelistMap(current)
	previous := make([]iptables_helper.PortWhitelist, 0, len(oldPorts))
	for _, port := range oldPorts {
		item := iptables_helper.PortWhitelist{Port: strconv.Itoa(port), Protocol: "tcp"}
		if _, stillRequired := currentSet[iptables_helper.PortWhitelistKey(item)]; !stillRequired {
			previous = append(previous, item)
		}
	}
	added := make([]iptables_helper.PortWhitelist, 0, len(newPorts))
	for _, port := range newPorts {
		added = append(added, iptables_helper.PortWhitelist{Port: strconv.Itoa(port), Protocol: "tcp"})
	}
	if state.Name == "iptables" {
		required, err := loadRequiredFirewallPortWhiteList()
		if err != nil {
			return err
		}
		added = excludeFirewallPorts(added, required)
	}
	return syncManagedAcceptedPorts(previous, added)
}

func LoadPanelPort() string {
	if !global.IsMaster {
		return global.CONF.Base.Port
	}
	var portSetting model.Setting
	_ = global.CoreDB.Where("key = ?", "ServerPort").First(&portSetting).Error
	return portSetting.Value
}

// SyncSystemPorts adds new ports before removing obsolete ports. This ordering
// preserves access when a management port is changed.
func (s *FirewallService) SyncSystemPorts(ctx context.Context, previous, current []dto.FirewallSystemPort) error {
	previousSet, err := normalizeSystemPorts(previous)
	if err != nil {
		return err
	}
	currentSet, err := normalizeSystemPorts(current)
	if err != nil {
		return err
	}

	for _, key := range sortedSystemPortKeys(currentSet) {
		if _, exists := previousSet[key]; exists {
			continue
		}
		if err := s.ensureSystemPort(ctx, currentSet[key]); err != nil {
			return err
		}
	}
	for _, key := range sortedSystemPortKeys(previousSet) {
		if _, exists := currentSet[key]; exists {
			continue
		}
		if err := s.deleteSystemPort(ctx, previousSet[key]); err != nil {
			return err
		}
	}
	return nil
}

func firewallRuleSnapshotPolicy(_ context.Context, snapshot filter.Snapshot) (filter.Snapshot, error) {
	ports, err := loadFirewallPortWhiteList()
	if err != nil {
		return filter.Snapshot{}, err
	}
	protectedPorts := make([]filter.ProtectedPort, 0, len(ports))
	for _, port := range ports {
		protectedPorts = append(protectedPorts, filter.ProtectedPort{Port: port.Port, Protocol: port.Protocol})
	}
	return filter.ProtectSnapshot(snapshot, protectedPorts)
}

func loadConfiguredFirewallPortWhiteList() ([]iptables_helper.PortWhitelist, error) {
	value, err := settingRepo.GetValueByKey(constant.FirewallPortWhiteList)
	if err != nil {
		value = constant.FirewallPortWhiteListValue
		if err := settingRepo.UpdateOrCreate(constant.FirewallPortWhiteList, value); err != nil {
			return nil, err
		}
	}
	return iptables_helper.ParsePortWhitelist(value)
}

func loadFirewallPortWhiteList() ([]iptables_helper.PortWhitelist, error) {
	configured, err := loadConfiguredFirewallPortWhiteList()
	if err != nil {
		return nil, err
	}
	required, err := loadRequiredFirewallPortWhiteList()
	if err != nil {
		return nil, err
	}
	return iptables_helper.NormalizePortWhitelist(append(configured, required...)), nil
}

func loadRequiredFirewallPortWhiteList() ([]iptables_helper.PortWhitelist, error) {
	panelPort := LoadPanelPort()
	if panelPort == "" {
		return nil, fmt.Errorf("find 1panel service port failed")
	}
	return iptables_helper.NormalizePortWhitelist([]iptables_helper.PortWhitelist{
		{Port: panelPort, Protocol: "tcp"},
		{Port: loadSSHPort(), Protocol: "tcp"},
	}), nil
}

func syncFirewallPortWhiteListAfterUpdate(oldValue string) error {
	client, err := lifecycle.NewClient()
	if err != nil {
		return err
	}
	state, err := lifecycle.LoadState(client)
	if err != nil {
		return err
	}
	if state.Name == "iptables" {
		isInit, _, err := iptables_helper.LoadInitStatus(state.Name, "base")
		if err != nil {
			return err
		}
		if !isInit {
			return nil
		}
	} else if !state.IsActive {
		return nil
	}
	ports, err := loadConfiguredFirewallPortWhiteList()
	if err != nil {
		return err
	}
	oldPorts, err := iptables_helper.ParsePortWhitelist(oldValue)
	if err != nil {
		return err
	}
	required, err := loadRequiredFirewallPortWhiteList()
	if err != nil {
		return err
	}
	if state.Name == "iptables" {
		ports = excludeFirewallPorts(ports, required)
		oldPorts = excludeFirewallPorts(oldPorts, required)
	} else {
		ports = iptables_helper.NormalizePortWhitelist(append(ports, required...))
		oldPorts = iptables_helper.NormalizePortWhitelist(append(oldPorts, required...))
	}
	return syncManagedAcceptedPorts(oldPorts, ports)
}

func newIptablesHelperManager() *iptables_helper.IptablesManager {
	return &iptables_helper.IptablesManager{
		UpdateSetting:     settingRepo.Update,
		PanelPort:         LoadPanelPort,
		LoadRequiredPorts: loadRequiredFirewallPortWhiteList,
	}
}

func (u *FirewallService) addPortsBeforeStart(client lifecycle.Client) error {
	if client.Name() == "iptables" {
		isInit, _, err := iptables_helper.LoadInitStatus(client.Name(), "base")
		if err != nil {
			return err
		}
		if !isInit {
			return nil
		}
		if err := newIptablesHelperManager().SyncRequiredPorts(true); err != nil {
			return err
		}
		configured, err := loadConfiguredFirewallPortWhiteList()
		if err != nil {
			return err
		}
		required, err := loadRequiredFirewallPortWhiteList()
		if err != nil {
			return err
		}
		return u.SyncSystemPorts(context.Background(), nil, systemPorts(excludeFirewallPorts(configured, required)))
	}
	portWhiteList, err := loadFirewallPortWhiteList()
	if err != nil {
		return err
	}
	return u.SyncSystemPorts(context.Background(), nil, systemPorts(portWhiteList))
}

func syncManagedAcceptedPorts(previous, current []iptables_helper.PortWhitelist) error {
	return newFirewallService().
		SyncSystemPorts(context.Background(), systemPorts(previous), systemPorts(current))
}

func systemPorts(ports []iptables_helper.PortWhitelist) []dto.FirewallSystemPort {
	result := make([]dto.FirewallSystemPort, 0, len(ports))
	for _, port := range ports {
		result = append(result, dto.FirewallSystemPort{Port: port.Port, Protocol: port.Protocol})
	}
	return result
}

func excludeFirewallPorts(ports, excluded []iptables_helper.PortWhitelist) []iptables_helper.PortWhitelist {
	excludedSet := iptables_helper.PortWhitelistMap(excluded)
	result := make([]iptables_helper.PortWhitelist, 0, len(ports))
	for _, port := range ports {
		if _, exists := excludedSet[iptables_helper.PortWhitelistKey(port)]; !exists {
			result = append(result, port)
		}
	}
	return result
}

func (s *FirewallService) checkSelectedProvider(ctx context.Context, requested filter.Provider) error {
	selected, err := s.selectedProvider(ctx)
	if err != nil {
		return err
	}
	if selected != requested {
		return fmt.Errorf("%w: selected provider is %s, requested %s", filter.ErrProviderUnavailable, selected, requested)
	}
	return nil
}

func firewallRuleSelectedProvider(context.Context) (filter.Provider, error) {
	return selectedRuleProvider()
}

func desiredFirewallRulesFromModels(stored []model.FirewallRule) ([]filter.DesiredRule, error) {
	desired := make([]filter.DesiredRule, 0, len(stored))
	for _, rule := range stored {
		converted, err := desiredFirewallRuleFromModel(rule)
		if err != nil {
			return nil, err
		}
		desired = append(desired, converted)
	}
	return desired, nil
}

func (s *FirewallService) createRule(
	ctx context.Context,
	runtime *firewallRuleRuntime,
	snapshot filter.Snapshot,
	request dto.FirewallRuleCreate,
	authorization firewallRuleCreateAuthorization,
) error {
	domainRule := request.Rule
	appendRule := false
	if authorization.Operation == filter.ChangeCreate && domainRule.Scope.Provider == filter.ProviderUFW {
		appendPosition, err := ufwAppendPosition(ctx, runtime, snapshot, domainRule)
		if err != nil {
			return err
		}
		domainRule.OrderIndex = &appendPosition
		appendRule = true
	} else if authorization.Operation == filter.ChangeCreate && domainRule.OrderIndex != nil {
		maxPosition, err := maxPositionForRule(ctx, runtime, snapshot, domainRule)
		if err != nil {
			return err
		}
		if *domainRule.OrderIndex < 1 || *domainRule.OrderIndex > maxPosition+1 {
			return fmt.Errorf("%w: create target position %d is out of range 1-%d", filter.ErrInvalidRule, *domainRule.OrderIndex, maxPosition+1)
		}
		appendRule = domainRule.Scope.Provider == filter.ProviderUFW && *domainRule.OrderIndex == maxPosition+1
	}
	origin := constant.FirewallRuleOriginCreated
	change := filter.DesiredChange{
		Operation: authorization.Operation,
		After:     &domainRule,
		Locator:   authorization.Locator,
		Append:    appendRule,
	}
	if authorization.Operation == filter.ChangeAdopt {
		origin = constant.FirewallRuleOriginAdopted
	}
	ruleRecord, err := firewallRuleModelForCreate(domainRule, request, origin)
	if err != nil {
		return err
	}
	if err := s.ensureFirewallRuleIdentityAvailable(ctx, ruleRecord.ScopeKey, ruleRecord.RuleKey, "", ""); err != nil {
		return err
	}
	if err := s.rules.Create(ctx, &ruleRecord); err != nil {
		return err
	}
	domainRule.UUID = ruleRecord.UUID
	change.After = &domainRule
	backendPlan, verification, err := runtime.Execute(ctx, snapshot, []filter.DesiredChange{change})
	if err != nil {
		return s.cleanupFailedCreate(ctx, ruleRecord, err)
	}
	if !verification.Matched {
		return s.cleanupFailedCreate(ctx, ruleRecord, filter.ErrVerificationFailed)
	}
	observed, err := filter.FindCommittedObserved(verification.Snapshot, domainRule, backendPlan)
	if err != nil {
		return s.cleanupAppliedCreate(ctx, runtime, backendPlan, ruleRecord, err)
	}
	updates, err := appliedFirewallRuleUpdates(observed)
	if err != nil {
		return s.cleanupAppliedCreate(ctx, runtime, backendPlan, ruleRecord, err)
	}
	matchKey, _ := updates["match_key"].(string)
	if err := s.ensureFirewallRuleIdentityAvailable(ctx, ruleRecord.ScopeKey, "", matchKey, ruleRecord.UUID); err != nil {
		return s.cleanupAppliedCreate(ctx, runtime, backendPlan, ruleRecord, err)
	}
	if err := s.rules.UpdateWithRevision(ctx, ruleRecord.UUID, ruleRecord.Revision, updates); err != nil {
		return s.cleanupAppliedCreate(ctx, runtime, backendPlan, ruleRecord, err)
	}
	return nil
}

func (s *FirewallService) deleteRule(ctx context.Context, ruleUUID string) error {
	if ruleUUID == "" {
		return fmt.Errorf("%w: rule UUID is required", repo.ErrFirewallPersistenceInvalid)
	}
	stored, err := s.rules.GetByUUID(ctx, ruleUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: managed rule %q was not found", filter.ErrInvalidRule, ruleUUID)
		}
		return err
	}
	if stored.Origin != constant.FirewallRuleOriginCreated && stored.Origin != constant.FirewallRuleOriginAdopted {
		return fmt.Errorf("%w: only created or adopted rules can be deleted", filter.ErrInvalidRule)
	}
	desired, err := desiredFirewallRuleFromModel(stored)
	if err != nil {
		return err
	}
	runtime, err := s.resolveRuntime(ctx, desired.Rule.Scope.Provider)
	if err != nil {
		return err
	}
	snapshot, err := runtime.ObserveMutation(ctx, desired.Rule.Scope)
	if err != nil {
		return err
	}
	observed, err := filter.ManagedObserved(snapshot, desired)
	if err != nil {
		return err
	}
	restoreAtEnd := false
	if desired.Rule.Scope.Provider == filter.ProviderUFW && observed.Locator.Position != nil {
		maxPosition, maxErr := maxPositionForRule(ctx, runtime, snapshot, desired.Rule)
		if maxErr != nil {
			return maxErr
		}
		restoreAtEnd = int64(*observed.Locator.Position) == maxPosition
	}
	locator := observed.Locator
	before := desired.Rule
	backendPlan, verification, err := runtime.Execute(ctx, snapshot, []filter.DesiredChange{{
		Operation:    filter.ChangeDelete,
		Before:       &before,
		Locator:      &locator,
		RestoreAtEnd: restoreAtEnd,
	}})
	if err != nil {
		return err
	}
	if !verification.Matched {
		return filter.ErrVerificationFailed
	}
	if err := s.rules.DeleteWithRevision(ctx, stored.UUID, stored.Revision); err != nil {
		return rollbackFirewallPlan(ctx, runtime, backendPlan, err)
	}
	return nil
}

func (s *FirewallService) updateRule(ctx context.Context, clientIP, ruleUUID string, requestedRule filter.FirewallRule) error {
	if ruleUUID == "" {
		return fmt.Errorf("%w: rule UUID is required", repo.ErrFirewallPersistenceInvalid)
	}
	stored, before, snapshot, observed, runtime, err := s.loadManagedMutation(ctx, ruleUUID)
	if err != nil {
		return err
	}
	after, err := filter.NormalizeRule(requestedRule)
	if err != nil {
		return err
	}
	after.UUID = stored.UUID
	after, err = runtime.Prepare(after)
	if err != nil {
		return err
	}
	if err := runtime.CheckRule(ctx, after); err != nil {
		return err
	}
	if after.Scope.Key() != before.Rule.Scope.Key() {
		return fmt.Errorf("%w: managed rule scope cannot be changed", filter.ErrUnsupportedScope)
	}
	if after.NativeKind != before.Rule.NativeKind {
		return fmt.Errorf("%w: native rule conversion requires an explicit workflow", filter.ErrUnsupportedScope)
	}
	capabilities, err := runtime.Capabilities(ctx)
	if err != nil {
		return err
	}
	if capabilities.ExplicitPosition || capabilities.OwnedChains {
		if observed.Locator.Position == nil {
			return fmt.Errorf("%w: managed rule has no positional locator", filter.ErrInvalidRule)
		}
		currentPosition := int64(*observed.Locator.Position)
		if after.OrderIndex == nil {
			after.OrderIndex = &currentPosition
		} else if *after.OrderIndex != currentPosition {
			if err := validatePositionTarget(ctx, runtime, snapshot, before.Rule, *after.OrderIndex); err != nil {
				return err
			}
		}
	}
	if err := filter.GuardMutation(snapshot, observed, after, clientIP); err != nil {
		return err
	}
	return s.executeManagedMutation(ctx, managedMutationRequest{
		Stored: stored, Before: before.Rule, After: after, Snapshot: snapshot, Locator: observed.Locator,
		AdapterOperation: filter.ChangeUpdate, Runtime: runtime,
	})
}

func (s *FirewallService) reorderRule(ctx context.Context, clientIP, ruleUUID string, targetPosition *int64, priority *int) error {
	if ruleUUID == "" {
		return fmt.Errorf("%w: rule UUID is required", repo.ErrFirewallPersistenceInvalid)
	}
	stored, before, snapshot, observed, runtime, err := s.loadManagedMutation(ctx, ruleUUID)
	if err != nil {
		return err
	}
	capabilities, err := runtime.Capabilities(ctx)
	if err != nil {
		return err
	}
	after := before.Rule
	adapterOperation := filter.ChangeReorder
	switch {
	case capabilities.OwnedChains:
		if targetPosition == nil || *targetPosition < 1 {
			return fmt.Errorf("%w: target position is required", filter.ErrInvalidRule)
		}
		if err := validatePositionTarget(ctx, runtime, snapshot, before.Rule, *targetPosition); err != nil {
			return err
		}
		after.OrderIndex = targetPosition
	case capabilities.ExplicitPriority:
		if priority == nil {
			return fmt.Errorf("%w: priority is required", filter.ErrInvalidRule)
		}
		if before.Rule.Priority == nil {
			return fmt.Errorf("%w: rule has no explicit reorderable priority", filter.ErrUnsupportedScope)
		}
		after.Priority = priority
		adapterOperation = filter.ChangeUpdate
	default:
		return fmt.Errorf("%w: provider does not support rule reordering", filter.ErrUnsupportedScope)
	}
	after, err = runtime.Prepare(after)
	if err != nil {
		return err
	}
	if err := filter.GuardMutation(snapshot, observed, after, clientIP); err != nil {
		return err
	}
	return s.executeManagedMutation(ctx, managedMutationRequest{
		Stored: stored, Before: before.Rule, After: after, Snapshot: snapshot, Locator: observed.Locator,
		AdapterOperation: adapterOperation, Runtime: runtime,
	})
}

func validatePositionTarget(
	ctx context.Context,
	runtime *firewallRuleRuntime,
	snapshot filter.Snapshot,
	rule filter.FirewallRule,
	targetPosition int64,
) error {
	if rule.Scope.Provider == filter.ProviderUFW {
		minPosition, maxPosition := snapshotPositionBounds(snapshot)
		if targetPosition < minPosition || targetPosition > maxPosition {
			return fmt.Errorf(
				"%w: target position %d is outside the %s range %d-%d",
				filter.ErrInvalidRule, targetPosition, rule.Scope.Family, minPosition, maxPosition,
			)
		}
		return nil
	}
	maxPosition, err := maxPositionForRule(ctx, runtime, snapshot, rule)
	if err != nil {
		return err
	}
	if targetPosition > maxPosition {
		return fmt.Errorf("%w: target position %d is out of range 1-%d", filter.ErrInvalidRule, targetPosition, maxPosition)
	}
	return nil
}

func ufwAppendPosition(
	ctx context.Context,
	runtime *firewallRuleRuntime,
	snapshot filter.Snapshot,
	rule filter.FirewallRule,
) (int64, error) {
	if rule.Scope.Family == filter.FamilyIPv4 {
		return snapshotMaxPosition(snapshot) + 1, nil
	}
	maxPosition, err := maxPositionForRule(ctx, runtime, snapshot, rule)
	if err != nil {
		return 0, err
	}
	return maxPosition + 1, nil
}

func snapshotPositionBounds(snapshot filter.Snapshot) (int64, int64) {
	minPosition, maxPosition := int64(0), int64(0)
	for _, observed := range snapshot.Rules {
		if observed.Locator.Position == nil {
			continue
		}
		position := int64(*observed.Locator.Position)
		if minPosition == 0 || position < minPosition {
			minPosition = position
		}
		if position > maxPosition {
			maxPosition = position
		}
	}
	return minPosition, maxPosition
}

func maxPositionForRule(
	ctx context.Context,
	runtime *firewallRuleRuntime,
	snapshot filter.Snapshot,
	rule filter.FirewallRule,
) (int64, error) {
	maxPosition := snapshotMaxPosition(snapshot)
	if rule.Scope.Provider == filter.ProviderUFW {
		relatedScope := rule.Scope
		if relatedScope.Family == filter.FamilyIPv4 {
			relatedScope.Family = filter.FamilyIPv6
		} else {
			relatedScope.Family = filter.FamilyIPv4
		}
		relatedSnapshot, err := runtime.ObserveMutation(ctx, relatedScope)
		if err != nil {
			return 0, err
		}
		if relatedMax := snapshotMaxPosition(relatedSnapshot); relatedMax > maxPosition {
			maxPosition = relatedMax
		}
	}
	return maxPosition, nil
}

func snapshotMaxPosition(snapshot filter.Snapshot) int64 {
	var maxPosition int64
	for _, observed := range snapshot.Rules {
		if observed.Locator.Position != nil && int64(*observed.Locator.Position) > maxPosition {
			maxPosition = int64(*observed.Locator.Position)
		}
	}
	return maxPosition
}

type managedMutationRequest struct {
	Stored           model.FirewallRule
	Before           filter.FirewallRule
	After            filter.FirewallRule
	Snapshot         filter.Snapshot
	Locator          filter.Locator
	AdapterOperation filter.ChangeOperation
	Runtime          *firewallRuleRuntime
}

func (s *FirewallService) loadManagedMutation(
	ctx context.Context,
	ruleUUID string,
) (model.FirewallRule, filter.DesiredRule, filter.Snapshot, filter.ObservedRule, *firewallRuleRuntime, error) {
	stored, err := s.rules.GetByUUID(ctx, ruleUUID)
	if err != nil {
		return model.FirewallRule{}, filter.DesiredRule{}, filter.Snapshot{}, filter.ObservedRule{}, nil, err
	}
	if stored.Origin != constant.FirewallRuleOriginCreated && stored.Origin != constant.FirewallRuleOriginAdopted {
		return model.FirewallRule{}, filter.DesiredRule{}, filter.Snapshot{}, filter.ObservedRule{}, nil,
			fmt.Errorf("%w: only created or adopted rules can be changed", filter.ErrInvalidRule)
	}
	desired, err := desiredFirewallRuleFromModel(stored)
	if err != nil {
		return model.FirewallRule{}, filter.DesiredRule{}, filter.Snapshot{}, filter.ObservedRule{}, nil, err
	}
	runtime, err := s.resolveRuntime(ctx, desired.Rule.Scope.Provider)
	if err != nil {
		return model.FirewallRule{}, filter.DesiredRule{}, filter.Snapshot{}, filter.ObservedRule{}, nil, err
	}
	snapshot, err := runtime.ObserveMutation(ctx, desired.Rule.Scope)
	if err != nil {
		return model.FirewallRule{}, filter.DesiredRule{}, filter.Snapshot{}, filter.ObservedRule{}, nil, err
	}
	observed, err := filter.ManagedObserved(snapshot, desired)
	if err != nil {
		return model.FirewallRule{}, filter.DesiredRule{}, filter.Snapshot{}, filter.ObservedRule{}, nil, err
	}
	return stored, desired, snapshot, observed, runtime, nil
}

func (s *FirewallService) executeManagedMutation(ctx context.Context, request managedMutationRequest) error {
	before, after := request.Before, request.After
	semantic, err := firewallRuleSemanticModel(after)
	if err != nil {
		return err
	}
	if err := s.ensureFirewallRuleIdentityAvailable(
		ctx, semantic.ScopeKey, semantic.RuleKey, "", request.Stored.UUID,
	); err != nil {
		return err
	}
	appendRule, restoreAtEnd := false, false
	if after.Scope.Provider == filter.ProviderUFW && request.AdapterOperation == filter.ChangeUpdate {
		maxPosition, maxErr := maxPositionForRule(ctx, request.Runtime, request.Snapshot, after)
		if maxErr != nil {
			return maxErr
		}
		appendRule = after.OrderIndex != nil && *after.OrderIndex == maxPosition
		restoreAtEnd = request.Locator.Position != nil && int64(*request.Locator.Position) == maxPosition
	}
	backendPlan, verification, err := request.Runtime.Execute(ctx, request.Snapshot, []filter.DesiredChange{{
		Operation:    request.AdapterOperation,
		Before:       &before,
		After:        &after,
		Locator:      &request.Locator,
		Append:       appendRule,
		RestoreAtEnd: restoreAtEnd,
	}})
	if err != nil {
		return err
	}
	if !verification.Matched {
		return filter.ErrVerificationFailed
	}
	observed, err := filter.FindCommittedObserved(verification.Snapshot, request.After, backendPlan)
	if err != nil {
		return rollbackFirewallPlan(ctx, request.Runtime, backendPlan, err)
	}
	updates, err := firewallRuleSemanticUpdates(request.After)
	if err != nil {
		return rollbackFirewallPlan(ctx, request.Runtime, backendPlan, err)
	}
	applied, err := appliedFirewallRuleUpdates(observed)
	if err != nil {
		return rollbackFirewallPlan(ctx, request.Runtime, backendPlan, err)
	}
	for key, value := range applied {
		updates[key] = value
	}
	matchKey, _ := updates["match_key"].(string)
	if err := s.ensureFirewallRuleIdentityAvailable(
		ctx, semantic.ScopeKey, "", matchKey, request.Stored.UUID,
	); err != nil {
		return rollbackFirewallPlan(ctx, request.Runtime, backendPlan, err)
	}
	if err := s.rules.UpdateWithRevision(ctx, request.Stored.UUID, request.Stored.Revision, updates); err != nil {
		return rollbackFirewallPlan(ctx, request.Runtime, backendPlan, err)
	}
	return nil
}

func (s *FirewallService) ensureFirewallRuleIdentityAvailable(
	ctx context.Context,
	scopeKey string,
	ruleKey string,
	matchKey string,
	excludedUUID string,
) error {
	stored, err := s.rules.List(ctx, repo.WithFirewallRuleScope(scopeKey))
	if err != nil {
		return err
	}
	for _, candidate := range stored {
		if candidate.UUID == excludedUUID {
			continue
		}
		if ruleKey != "" && candidate.RuleKey == ruleKey {
			return fmt.Errorf("%w: equivalent managed rule already exists", filter.ErrRuleOperation)
		}
		if matchKey != "" && candidate.MatchKey == matchKey {
			return fmt.Errorf("%w: firewall rule instance is already managed", filter.ErrRuleOperation)
		}
	}
	return nil
}

func (s *FirewallService) resolveRuntime(ctx context.Context, provider filter.Provider) (*firewallRuleRuntime, error) {
	if s.selectedProvider != nil {
		selected, err := s.selectedProvider(ctx)
		if err != nil {
			return nil, err
		}
		if selected != provider {
			return nil, fmt.Errorf("%w: selected provider is %s, requested %s", filter.ErrProviderUnavailable, selected, provider)
		}
	}
	if s.adapters == nil {
		return nil, filter.ErrAdapterUnavailable
	}
	return s.adapters.Resolve(provider)
}

func (s *FirewallService) cleanupFailedCreate(ctx context.Context, rule model.FirewallRule, cause error) error {
	if err := s.rules.DeleteWithRevision(ctx, rule.UUID, rule.Revision); err != nil {
		return errors.Join(cause, fmt.Errorf("cleanup failed firewall rule %q: %w", rule.UUID, err))
	}
	return cause
}

func (s *FirewallService) cleanupAppliedCreate(
	ctx context.Context,
	runtime *firewallRuleRuntime,
	plan filter.BackendPlan,
	rule model.FirewallRule,
	cause error,
) error {
	cause = rollbackFirewallPlan(ctx, runtime, plan, cause)
	return s.cleanupFailedCreate(ctx, rule, cause)
}

func rollbackFirewallPlan(ctx context.Context, runtime *firewallRuleRuntime, plan filter.BackendPlan, cause error) error {
	if runtime == nil {
		return cause
	}
	if err := runtime.Rollback(ctx, plan); err != nil {
		return errors.Join(cause, fmt.Errorf("rollback applied firewall plan: %w", err))
	}
	return cause
}

func (s *FirewallService) ensureSystemPort(ctx context.Context, port dto.FirewallSystemPort) error {
	provider, err := s.selectedProvider(ctx)
	if err != nil {
		return err
	}
	rule := systemPortRule(provider, port)
	check, err := s.Check(ctx, "", dto.FirewallRuleCheck{Rule: rule})
	if err != nil {
		return err
	}

	create := dto.FirewallRuleCreate{
		Rule:       check.RequestedRule,
		CheckFlag:  check.CheckFlag,
		SourceKind: constant.FirewallRuleSourceSecurity,
		SourceID:   constant.FirewallSystemAcceptedPortSourcePrefix + systemPortKey(port),
	}
	switch check.Classification {
	case filter.CheckClassificationNone:
		create.Action = filter.CheckActionCreate
	case filter.CheckClassificationExactExternal:
		if len(check.Candidates) == 0 {
			return fmt.Errorf("%w: external port rule has no candidate", filter.ErrRuleStale)
		}
		if len(check.Candidates) == 1 {
			create.Action = filter.CheckActionAdopt
		} else {
			create.Action = filter.CheckActionSelectAdopt
		}
		create.AdoptInstanceKey = check.Candidates[0].InstanceKey
	case filter.CheckClassificationCovered:
		create.Action = filter.CheckActionCreateAnyway
	case filter.CheckClassificationExactManaged:
		return nil
	case filter.CheckClassificationProtected:
		// A required/configured port can already exist as a legacy rule. It is
		// safe to leave that protected equivalent in place without taking it over.
		if len(check.Candidates) > 0 {
			return nil
		}
		return fmt.Errorf("%w: protected accepted port %s", filter.ErrProtectedRule, port.Port)
	default:
		return fmt.Errorf("cannot manage accepted port %s/%s: %s", port.Port, port.Protocol, check.Reason)
	}
	return s.Create(ctx, create)
}

func (s *FirewallService) deleteSystemPort(ctx context.Context, port dto.FirewallSystemPort) error {
	for {
		stored, err := s.systemPortRecords(ctx, port)
		if err != nil {
			return err
		}
		if len(stored) == 0 {
			adopted, err := s.adoptExternalSystemPort(ctx, port)
			if err != nil || !adopted {
				return err
			}
			continue
		}
		for _, rule := range stored {
			if err := s.Delete(ctx, dto.FirewallRuleDelete{UUID: rule.UUID}); err != nil {
				return err
			}
		}
		// Re-plan once ownership is gone so duplicate legacy rules are removed
		// through the same adopt/delete workflow as the first one.
	}
}

func (s *FirewallService) systemPortRecords(ctx context.Context, port dto.FirewallSystemPort) ([]model.FirewallRule, error) {
	return s.rules.List(ctx,
		repo.WithFirewallRuleSource(constant.FirewallRuleSourceSecurity, constant.FirewallSystemAcceptedPortSourcePrefix+systemPortKey(port)),
	)
}

func (s *FirewallService) adoptExternalSystemPort(ctx context.Context, port dto.FirewallSystemPort) (bool, error) {
	provider, err := s.selectedProvider(ctx)
	if err != nil {
		return false, err
	}
	check, err := s.Check(ctx, "", dto.FirewallRuleCheck{Rule: systemPortRule(provider, port)})
	if err != nil {
		return false, err
	}
	if check.Classification != filter.CheckClassificationExactExternal || len(check.Candidates) == 0 {
		return false, nil
	}
	create := dto.FirewallRuleCreate{
		Rule:             check.RequestedRule,
		CheckFlag:        check.CheckFlag,
		AdoptInstanceKey: check.Candidates[0].InstanceKey,
		SourceKind:       constant.FirewallRuleSourceSecurity,
		SourceID:         constant.FirewallSystemAcceptedPortSourcePrefix + systemPortKey(port),
	}
	if len(check.Candidates) == 1 {
		create.Action = filter.CheckActionAdopt
	} else {
		create.Action = filter.CheckActionSelectAdopt
	}
	if err := s.Create(ctx, create); err != nil {
		return false, err
	}
	return true, nil
}

func systemPortRule(provider filter.Provider, port dto.FirewallSystemPort) filter.FirewallRule {
	scope := filter.Scope{Provider: provider, Direction: filter.DirectionInput}
	switch provider {
	case filter.ProviderIptables:
		scope.Family = filter.FamilyIPv4
		scope.Table = "filter"
	case filter.ProviderFirewalld:
		scope.Family = filter.FamilyInet
		scope.Zone = filter.FirewalldInputZone
	case filter.ProviderUFW:
		scope.Family = filter.FamilyIPv4
	}
	return filter.FirewallRule{
		Scope: scope, Protocol: port.Protocol, DestinationPort: port.Port,
		Action: filter.ActionAccept, Description: "1Panel managed accepted port",
	}
}

func normalizeSystemPorts(ports []dto.FirewallSystemPort) (map[string]dto.FirewallSystemPort, error) {
	result := make(map[string]dto.FirewallSystemPort, len(ports))
	for _, port := range ports {
		normalized, err := filter.NormalizeRule(systemPortRule(filter.ProviderIptables, port))
		if err != nil {
			return nil, err
		}
		item := dto.FirewallSystemPort{Port: normalized.DestinationPort, Protocol: normalized.Protocol}
		result[systemPortKey(item)] = item
	}
	return result, nil
}

func systemPortKey(port dto.FirewallSystemPort) string {
	return strings.ToLower(strings.TrimSpace(port.Protocol)) + "/" + strings.TrimSpace(port.Port)
}

func sortedSystemPortKeys(ports map[string]dto.FirewallSystemPort) []string {
	keys := make([]string, 0, len(ports))
	for key := range ports {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

type firewallSnapshotPolicy func(context.Context, filter.Snapshot) (filter.Snapshot, error)

type firewallRuleRuntime struct {
	adapter filter.Adapter
	policy  firewallSnapshotPolicy
}

type firewallRuleRuntimeRegistry map[filter.Provider]*firewallRuleRuntime

func newFirewallRuleRuntimeRegistry(policy firewallSnapshotPolicy) firewallRuleRuntimeRegistry {
	return firewallRuleRuntimeRegistry{
		filter.ProviderIptables:  newFirewallRuleRuntime(filteriptables.NewAdapter(), policy),
		filter.ProviderFirewalld: newFirewallRuleRuntime(filterfirewalld.NewAdapter(), policy),
		filter.ProviderUFW:       newFirewallRuleRuntime(filterufw.NewAdapter(), policy),
	}
}

func newFirewallRuleRuntime(adapter filter.Adapter, policy firewallSnapshotPolicy) *firewallRuleRuntime {
	return &firewallRuleRuntime{adapter: adapter, policy: policy}
}

func (r firewallRuleRuntimeRegistry) Resolve(provider filter.Provider) (*firewallRuleRuntime, error) {
	runtime, exists := r[provider]
	if !exists || runtime == nil || runtime.adapter == nil {
		return nil, fmt.Errorf("%w: %s", filter.ErrAdapterUnavailable, provider)
	}
	return runtime, nil
}

func (r *firewallRuleRuntime) Observe(ctx context.Context, scope filter.Scope) (filter.Snapshot, error) {
	snapshot, err := r.adapter.Observe(ctx, scope)
	if err != nil {
		return filter.Snapshot{}, err
	}
	if r.policy == nil {
		return snapshot, nil
	}
	return r.policy(ctx, snapshot)
}

func (r *firewallRuleRuntime) ObserveMutation(ctx context.Context, scope filter.Scope) (filter.Snapshot, error) {
	snapshot, err := r.Observe(ctx, scope)
	if err != nil {
		return filter.Snapshot{}, err
	}
	for _, notice := range snapshot.Notices {
		if notice.Code == filter.ScopeNoticeManagedScopeInactive {
			return filter.Snapshot{}, fmt.Errorf("%w: managed firewall scope is inactive", filter.ErrProviderUnavailable)
		}
	}
	return snapshot, nil
}

func (r *firewallRuleRuntime) Prepare(rule filter.FirewallRule) (filter.FirewallRule, error) {
	preparer, ok := r.adapter.(filter.RulePreparer)
	if !ok {
		return rule, nil
	}
	return preparer.PrepareRule(rule)
}

func (r *firewallRuleRuntime) CheckRule(ctx context.Context, rule filter.FirewallRule) error {
	checker, ok := r.adapter.(filter.RuleChecker)
	if !ok {
		return nil
	}
	return checker.CheckRule(ctx, rule)
}

func (r *firewallRuleRuntime) Capabilities(ctx context.Context) (filter.Capabilities, error) {
	return r.adapter.Capabilities(ctx)
}

func (r *firewallRuleRuntime) Execute(ctx context.Context, snapshot filter.Snapshot, changes []filter.DesiredChange) (filter.BackendPlan, filter.VerifyResult, error) {
	plan, err := r.adapter.Compile(snapshot, changes)
	if err != nil {
		return filter.BackendPlan{}, filter.VerifyResult{}, err
	}
	result, err := r.adapter.Apply(ctx, plan)
	if err != nil {
		return plan, filter.VerifyResult{}, err
	}
	if result.Verification != nil {
		if !result.Verification.Matched {
			if rollbackErr := r.Rollback(ctx, plan); rollbackErr != nil {
				return plan, *result.Verification, errors.Join(filter.ErrVerificationFailed, rollbackErr)
			}
		}
		return plan, *result.Verification, nil
	}
	verification, err := r.adapter.Verify(ctx, plan)
	if err != nil {
		return plan, verification, rollbackFirewallPlan(ctx, r, plan, err)
	}
	if !verification.Matched {
		if rollbackErr := r.Rollback(ctx, plan); rollbackErr != nil {
			return plan, verification, errors.Join(filter.ErrVerificationFailed, rollbackErr)
		}
	}
	return plan, verification, nil
}

func (r *firewallRuleRuntime) Rollback(ctx context.Context, plan filter.BackendPlan) error {
	rollbacker, ok := r.adapter.(filter.PlanRollbacker)
	if !ok {
		return fmt.Errorf("%w: provider %s does not support applied-plan rollback", filter.ErrAdapterUnavailable, r.adapter.Provider())
	}
	return rollbacker.Rollback(ctx, plan)
}

func selectedRuleProvider() (filter.Provider, error) {
	provider, err := lifecycle.DetectProvider()
	if err != nil {
		return "", fmt.Errorf("%w: %v", filter.ErrProviderUnavailable, err)
	}
	return filter.Provider(provider), nil
}

type firewallRuleCheckClaims struct {
	Version            int                         `json:"version"`
	Provider           filter.Provider             `json:"provider"`
	ScopeKey           string                      `json:"scopeKey"`
	RuleDigest         string                      `json:"ruleDigest"`
	SnapshotRevision   string                      `json:"snapshotRevision"`
	ManagedRevision    string                      `json:"managedRevision"`
	Decision           filter.CheckDecision        `json:"decision"`
	Classification     filter.CheckClassification  `json:"classification"`
	AllowedActions     []filter.CheckAction        `json:"allowedActions"`
	AdoptionCandidates []firewallAdoptionCandidate `json:"adoptionCandidates,omitempty"`
}

type firewallAdoptionCandidate struct {
	InstanceKey string         `json:"instanceKey"`
	Locator     filter.Locator `json:"locator"`
}

type firewallRuleCreateAuthorization struct {
	Operation filter.ChangeOperation
	Locator   *filter.Locator
}

func refreshBatchCreateAuthorization(
	snapshot filter.Snapshot,
	prepared preparedFirewallRuleCreate,
) (firewallRuleCreateAuthorization, error) {
	authorization := prepared.authorization
	if authorization.Operation != filter.ChangeAdopt {
		return authorization, nil
	}
	if candidate, err := filter.FindCandidate(snapshot.Rules, prepared.request.AdoptInstanceKey); err == nil {
		locator := candidate.Locator
		authorization.Locator = &locator
		return authorization, nil
	}

	ruleKey, err := filter.RuleKey(prepared.request.Rule)
	if err != nil {
		return firewallRuleCreateAuthorization{}, err
	}
	candidates := make([]filter.ObservedRule, 0)
	for _, observed := range snapshot.Rules {
		if observed.Marker != "" || observed.Protected || observed.ParseStatus != filter.ParseStatusSupported {
			continue
		}
		observedKey, keyErr := filter.RuleKey(observed.Rule)
		if keyErr == nil && observedKey == ruleKey {
			candidates = append(candidates, observed)
		}
	}
	if len(candidates) == 1 {
		locator := candidates[0].Locator
		authorization.Locator = &locator
		return authorization, nil
	}
	if authorization.Locator != nil {
		for _, candidate := range candidates {
			if authorization.Locator.Canonical != "" && candidate.Locator.Canonical == authorization.Locator.Canonical ||
				authorization.Locator.NativeID != "" && candidate.Locator.NativeID == authorization.Locator.NativeID {
				locator := candidate.Locator
				authorization.Locator = &locator
				return authorization, nil
			}
		}
	}
	return firewallRuleCreateAuthorization{}, filter.ErrRuleStale
}

func signFirewallRuleCheck(result filter.RuleCheckResult, snapshot filter.Snapshot, managedRevision string) (string, error) {
	ruleDigest, err := firewallRuleDigest(result.RequestedRule)
	if err != nil {
		return "", err
	}
	claims := firewallRuleCheckClaims{
		Version:          constant.FirewallRuleCheckVersion,
		Provider:         result.RequestedRule.Scope.Provider,
		ScopeKey:         result.RequestedRule.Scope.Key(),
		RuleDigest:       ruleDigest,
		SnapshotRevision: snapshot.Revision,
		ManagedRevision:  managedRevision,
		Decision:         result.Decision,
		Classification:   result.Classification,
		AllowedActions:   result.AllowedActions,
	}
	if result.Classification == filter.CheckClassificationExactExternal {
		claims.AdoptionCandidates = make([]firewallAdoptionCandidate, 0, len(result.Candidates))
		for _, candidate := range result.Candidates {
			claims.AdoptionCandidates = append(claims.AdoptionCandidates, firewallAdoptionCandidate{
				InstanceKey: candidate.InstanceKey,
				Locator:     candidate.Locator,
			})
		}
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	signature := firewallRuleCheckSignature(payload)
	return base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func authorizeFirewallRuleCreate(
	checkFlag string,
	action filter.CheckAction,
	adoptInstanceKey string,
	rule filter.FirewallRule,
	snapshot filter.Snapshot,
	managedRevision string,
) (firewallRuleCreateAuthorization, error) {
	claims, err := parseFirewallRuleCheck(checkFlag)
	if err != nil {
		return firewallRuleCreateAuthorization{}, err
	}
	ruleDigest, err := firewallRuleDigest(rule)
	if err != nil {
		return firewallRuleCreateAuthorization{}, err
	}
	if claims.Version != constant.FirewallRuleCheckVersion ||
		claims.Provider != rule.Scope.Provider ||
		claims.ScopeKey != rule.Scope.Key() ||
		claims.RuleDigest != ruleDigest ||
		claims.SnapshotRevision != snapshot.Revision ||
		claims.ManagedRevision != managedRevision {
		return firewallRuleCreateAuthorization{}, fmt.Errorf("%w: firewall or managed rules changed", filter.ErrRuleCheckRequired)
	}
	if claims.Decision != filter.CheckDecisionReady && claims.Decision != filter.CheckDecisionConfirmationRequired {
		return firewallRuleCreateAuthorization{}, filter.ErrRuleOperation
	}
	if !containsFirewallCheckAction(claims.AllowedActions, action) {
		return firewallRuleCreateAuthorization{}, filter.ErrRuleOperation
	}

	switch action {
	case filter.CheckActionCreate, filter.CheckActionCreateAnyway:
		if strings.TrimSpace(adoptInstanceKey) != "" {
			return firewallRuleCreateAuthorization{}, filter.ErrRuleOperation
		}
		return firewallRuleCreateAuthorization{Operation: filter.ChangeCreate}, nil
	case filter.CheckActionAdopt, filter.CheckActionSelectAdopt:
		for _, candidate := range claims.AdoptionCandidates {
			if candidate.InstanceKey == adoptInstanceKey && adoptInstanceKey != "" {
				locator := candidate.Locator
				return firewallRuleCreateAuthorization{Operation: filter.ChangeAdopt, Locator: &locator}, nil
			}
		}
		return firewallRuleCreateAuthorization{}, filter.ErrRuleOperation
	default:
		return firewallRuleCreateAuthorization{}, filter.ErrRuleOperation
	}
}

func parseFirewallRuleCheck(checkFlag string) (firewallRuleCheckClaims, error) {
	parts := strings.Split(strings.TrimSpace(checkFlag), ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return firewallRuleCheckClaims{}, filter.ErrRuleCheckRequired
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return firewallRuleCheckClaims{}, filter.ErrRuleCheckRequired
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(signature, firewallRuleCheckSignature(payload)) {
		return firewallRuleCheckClaims{}, filter.ErrRuleCheckRequired
	}
	var claims firewallRuleCheckClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return firewallRuleCheckClaims{}, filter.ErrRuleCheckRequired
	}
	return claims, nil
}

func firewallRuleCheckSignature(payload []byte) []byte {
	mac := hmac.New(sha256.New, []byte(global.CONF.Base.EncryptKey+"\x00firewall-rule-check-v1"))
	_, _ = mac.Write(payload)
	return mac.Sum(nil)
}

func firewallRuleDigest(rule filter.FirewallRule) (string, error) {
	payload, err := json.Marshal(rule)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func firewallManagedRevision(rules []model.FirewallRule) (string, error) {
	ordered := append([]model.FirewallRule(nil), rules...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].UUID < ordered[j].UUID
	})
	payload, err := json.Marshal(ordered)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func containsFirewallCheckAction(actions []filter.CheckAction, expected filter.CheckAction) bool {
	for _, action := range actions {
		if action == expected {
			return true
		}
	}
	return false
}

func firewallRuleModelForCreate(rule filter.FirewallRule, request dto.FirewallRuleCreate, origin string) (model.FirewallRule, error) {
	record, err := firewallRuleSemanticModel(rule)
	if err != nil {
		return model.FirewallRule{}, err
	}
	record.Origin = origin
	record.Owner = model.FirewallRuleOwner(request.SourceKind, request.SourceID)
	return record, nil
}

func firewallRuleSemanticModel(rule filter.FirewallRule) (model.FirewallRule, error) {
	normalized, err := filter.NormalizeRule(rule)
	if err != nil {
		return model.FirewallRule{}, err
	}
	ruleKey, err := filter.RuleKey(normalized)
	if err != nil {
		return model.FirewallRule{}, err
	}
	return model.FirewallRule{
		ScopeKey:           normalized.Scope.Key(),
		Provider:           string(normalized.Scope.Provider),
		Family:             string(normalized.Scope.Family),
		Location:           firewallRuleLocation(normalized.Scope),
		NativeKind:         string(normalized.NativeKind),
		Protocol:           normalized.Protocol,
		SourceAddress:      normalized.SourceAddress,
		SourcePort:         normalized.SourcePort,
		DestinationAddress: normalized.DestinationAddress,
		DestinationPort:    normalized.DestinationPort,
		Interface:          normalized.Interface,
		ConnectionStates:   strings.Join(normalized.ConnectionStates, ","),
		Action:             string(normalized.Action),
		Priority:           normalized.Priority,
		OrderIndex:         persistedFirewallOrderIndex(normalized),
		OrderBucket:        normalized.OrderBucket,
		Description:        normalized.Description,
		RuleKey:            ruleKey,
	}, nil
}

func persistedFirewallOrderIndex(rule filter.FirewallRule) *int64 {
	if rule.Scope.Provider == filter.ProviderIptables || rule.Scope.Provider == filter.ProviderUFW {
		return nil
	}
	return rule.OrderIndex
}

func firewallRuleLocation(scope filter.Scope) string {
	switch scope.Provider {
	case filter.ProviderFirewalld:
		return scope.Zone
	case filter.ProviderIptables, filter.ProviderUFW:
		return scope.Chain
	default:
		return ""
	}
}

func appliedFirewallRuleUpdates(observed filter.ObservedRule) (map[string]interface{}, error) {
	instanceKey, err := filter.InstanceKey(observed)
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{
		"match_key": firewallRuleMatchKey(observed.Marker, instanceKey),
	}
	return updates, nil
}

const (
	firewallRuleMarkerMatchPrefix   = "marker:"
	firewallRuleInstanceMatchPrefix = "instance:"
)

func firewallRuleMatchKey(marker, instanceKey string) string {
	if marker = strings.TrimSpace(marker); marker != "" {
		return firewallRuleMarkerMatchPrefix + marker
	}
	if instanceKey = strings.TrimSpace(instanceKey); instanceKey != "" {
		return firewallRuleInstanceMatchPrefix + instanceKey
	}
	return ""
}

func firewallRuleMatchValues(matchKey string) (marker, instanceKey string) {
	switch {
	case strings.HasPrefix(matchKey, firewallRuleMarkerMatchPrefix):
		return strings.TrimPrefix(matchKey, firewallRuleMarkerMatchPrefix), ""
	case strings.HasPrefix(matchKey, firewallRuleInstanceMatchPrefix):
		return "", strings.TrimPrefix(matchKey, firewallRuleInstanceMatchPrefix)
	default:
		return "", ""
	}
}

func firewallRuleSemanticUpdates(rule filter.FirewallRule) (map[string]interface{}, error) {
	record, err := firewallRuleSemanticModel(rule)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"scope_key": record.ScopeKey, "provider": record.Provider, "family": record.Family, "location": record.Location,
		"native_kind": record.NativeKind, "protocol": record.Protocol,
		"source_address": record.SourceAddress, "source_port": record.SourcePort,
		"destination_address": record.DestinationAddress, "destination_port": record.DestinationPort,
		"interface": record.Interface, "connection_states": record.ConnectionStates, "action": record.Action,
		"priority": record.Priority, "order_index": record.OrderIndex, "order_bucket": record.OrderBucket,
		"description": record.Description, "rule_key": record.RuleKey,
	}, nil
}

func mergeFirewallInventory(
	observed []filter.ObservedRule,
	stored []model.FirewallRule,
	protectedObservedKeys map[string]struct{},
	usage map[string]filter.RuntimeUsage,
) ([]filter.InventoryItem, error) {
	desired := make([]filter.DesiredRule, 0, len(stored))
	for _, storedRule := range stored {
		rule, err := desiredFirewallRuleFromModel(storedRule)
		if err != nil {
			return nil, err
		}
		desired = append(desired, rule)
	}
	items, err := filter.MergeInventory(filter.InventoryMergeInput{
		Observed:              observed,
		Desired:               desired,
		ProtectedObservedKeys: protectedObservedKeys,
	})
	if err != nil {
		return nil, err
	}
	return filter.AttachRuntimeUsage(items, usage), nil
}

func desiredFirewallRuleFromModel(stored model.FirewallRule) (filter.DesiredRule, error) {
	connectionStates := make([]string, 0)
	if stored.ConnectionStates != "" {
		connectionStates = strings.Split(stored.ConnectionStates, ",")
	}
	scope := filter.Scope{
		Provider: filter.Provider(stored.Provider), Family: filter.Family(stored.Family), Direction: filter.DirectionInput,
	}
	switch scope.Provider {
	case filter.ProviderIptables:
		scope.Table = "filter"
		scope.Chain = stored.Location
	case filter.ProviderFirewalld:
		scope.Zone = stored.Location
	case filter.ProviderUFW:
		scope.Chain = stored.Location
	}
	rule := filter.FirewallRule{
		UUID:               stored.UUID,
		Scope:              scope,
		NativeKind:         filter.NativeKind(stored.NativeKind),
		Protocol:           stored.Protocol,
		SourceAddress:      stored.SourceAddress,
		SourcePort:         stored.SourcePort,
		DestinationAddress: stored.DestinationAddress,
		DestinationPort:    stored.DestinationPort,
		Interface:          stored.Interface,
		ConnectionStates:   connectionStates,
		Action:             filter.Action(stored.Action),
		Priority:           stored.Priority,
		OrderIndex:         stored.OrderIndex,
		OrderBucket:        stored.OrderBucket,
		Description:        stored.Description,
	}
	normalized, err := filter.NormalizeRule(rule)
	if err != nil {
		return filter.DesiredRule{}, fmt.Errorf("normalize firewall rule %q payload: %w", stored.UUID, err)
	}
	if normalized.Scope.Key() != stored.ScopeKey {
		return filter.DesiredRule{}, fmt.Errorf("%w: stored firewall rule %q scope key does not match payload", filter.ErrInvalidRule, stored.UUID)
	}
	ruleKey, err := filter.RuleKey(normalized)
	if err != nil {
		return filter.DesiredRule{}, err
	}
	if ruleKey != stored.RuleKey {
		return filter.DesiredRule{}, fmt.Errorf("%w: stored firewall rule %q key does not match payload", filter.ErrInvalidRule, stored.UUID)
	}
	marker, observedInstanceKey := firewallRuleMatchValues(stored.MatchKey)
	return filter.DesiredRule{
		UUID:                stored.UUID,
		Rule:                normalized,
		RuleKey:             stored.RuleKey,
		Origin:              filter.RuleOrigin(stored.Origin),
		Marker:              marker,
		ObservedInstanceKey: observedInstanceKey,
	}, nil
}
