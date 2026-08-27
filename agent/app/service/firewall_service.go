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
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	filterfirewalld "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/firewalld"
	filteriptables "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/iptables"
	filternftables "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/nftables"
	filterufw "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/ufw"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/nftables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/ping"
	"gorm.io/gorm"
)

type FirewallService struct {
	rules                  repo.IFirewallRuleRepo
	adapters               firewallRuleRuntimeRegistry
	forwardingSync         firewallDatabaseSyncAdapter
	dockerSync             firewallDatabaseSyncAdapter
	selectedProvider       func(context.Context) (filter.Provider, error)
	protectedPorts         func() ([]firewall.PortWhitelist, error)
	iptablesHelper         *iptables_helper.Manager
	cleanupBackend         func(string) error
	cleanupInactiveBackend func(string) error
	resetBackend           func(string) error
	baseClient             func() (lifecycle.Client, error)
	installedProviders     func() []string
}

var firewallRuleMutationMu sync.Mutex

type IFirewallService interface {
	LoadBaseInfo(chainGroup string) (dto.FirewallSubsystemStatus, error)
	OperateFirewall(request dto.FirewallLifecycleOperation) error
	OperateFilterChain(request dto.FilterChainOperation) error
	Inventory(context.Context, dto.FirewallRuleInventory) (dto.FirewallRuleInventoryResponse, error)
	LoadFirewallNativeDetail(context.Context, dto.FirewallNativeDetail) (string, error)
	Check(context.Context, string, dto.FirewallRuleCheck) (dto.FirewallRuleCheckResponse, error)
	Create(context.Context, dto.FirewallRuleCreate) (dto.FirewallRuleCreateResponse, error)
	PreviewRuleSync(context.Context, string, dto.FirewallRuleSyncRequest) (dto.FirewallRuleSyncPreview, error)
	SyncRules(context.Context, string, dto.FirewallRuleSyncRequest) (dto.FirewallRuleSyncResult, error)
	CurrentRuleSyncTask() (dto.FirewallRuleSyncTask, error)
	Delete(context.Context, dto.FirewallRuleDelete) (dto.FirewallRuleDeleteResponse, error)
	Update(context.Context, string, dto.FirewallRuleUpdate) error
	Reorder(context.Context, string, dto.FirewallRuleReorder) error
	Reset(context.Context, dto.FirewallRuleReset) (dto.FirewallRuleResetResponse, error)
}

func NewIFirewallService() IFirewallService {
	return newFirewallService()
}

func newFirewallService() *FirewallService {
	return &FirewallService{
		rules:                  repo.NewIFirewallRuleRepo(),
		adapters:               newFirewallRuleRuntimeRegistry(firewallRuleSnapshotPolicy),
		forwardingSync:         newForwardingService(),
		dockerSync:             newDockerPortGuardService(),
		selectedProvider:       firewallRuleSelectedProvider,
		protectedPorts:         loadFirewallPortWhiteList,
		iptablesHelper:         newIptablesHelperManager(),
		cleanupBackend:         cleanupSystemBackend,
		cleanupInactiveBackend: cleanupInactiveSystemBackend,
		resetBackend:           resetServiceFirewallBackend,
		baseClient:             selectedSystemFirewallClient,
		installedProviders:     lifecycle.InstalledProviders,
	}
}

func (s *FirewallService) LoadBaseInfo(chainGroup string) (dto.FirewallSubsystemStatus, error) {
	status := dto.FirewallSubsystemStatus{Version: "-", Name: "-", Backend: "-"}
	loadClient := s.baseClient
	if loadClient == nil {
		loadClient = selectedSystemFirewallClient
	}
	client, err := loadClient()
	if err != nil {
		if global.LOG != nil {
			global.LOG.Errorf("load firewall failed, err: %v", err)
		}
		loadInstalled := s.installedProviders
		if loadInstalled == nil {
			loadInstalled = lifecycle.InstalledProviders
		}
		if len(loadInstalled()) > 0 {
			status.IsExist = true
			status.Message = err.Error()
		}
		return status, nil
	}
	status.IsExist = true
	runtimeStatus, err := lifecycle.LoadStatus(client)
	if err != nil {
		return status, err
	}
	status.Name, status.Backend = runtimeStatus.Name, runtimeStatus.Name
	status.Version, status.PingStatus = runtimeStatus.Version, ping.LoadStatus()
	status.IsActive = runtimeStatus.IsActive
	if supportsManagedFilterChains(runtimeStatus.Name) {
		initialized, bound, err := loadFirewallInitStatus(runtimeStatus.Name, chainGroup)
		if err != nil {
			return status, err
		}
		status.IsInit, status.IsBind = initialized, bound
		status.IPv4 = loadSystemFirewallFamilyInfo(status.Name, constant.FirewallFamilyIPv4)
		status.IPv6 = loadSystemFirewallFamilyInfo(status.Name, constant.FirewallFamilyIPv6)
		if chainGroup == "base" {
			status.ConflictBackend = conflictingDirectFirewallBackend(runtimeStatus.Name)
		}
	}
	return status, nil
}

func conflictingDirectFirewallBackend(provider string) string {
	other := ""
	switch provider {
	case constant.FirewallProviderIptables:
		other = constant.FirewallProviderNftables
	case constant.FirewallProviderNftables:
		other = constant.FirewallProviderIptables
	default:
		return ""
	}
	for _, family := range []string{constant.FirewallFamilyIPv4, constant.FirewallFamilyIPv6} {
		var (
			bound bool
			err   error
		)
		if other == constant.FirewallProviderIptables {
			bound, err = iptables_helper.LoadFamilyBindStatus(family)
		} else {
			bound, err = nftables_helper.LoadFamilyBindStatus(filter.Family(family))
		}
		if err == nil && bound {
			return other
		}
	}
	return ""
}

func (s *FirewallService) OperateFirewall(request dto.FirewallLifecycleOperation) error {
	switch request.Operation {
	case "disableBanPing":
		if err := ping.UpdateStatus("0"); err != nil {
			return err
		}
		return settingRepo.Update(constant.FirewallPingStatusKey, constant.StatusDisable)
	case "enableBanPing":
		if err := ping.UpdateStatus("1"); err != nil {
			return err
		}
		return settingRepo.Update(constant.FirewallPingStatusKey, constant.StatusEnable)
	}
	client, err := selectedSystemFirewallClient()
	if err != nil {
		return err
	}
	if err := lifecycle.NewOperator(client).Operate(lifecycle.Operation(request.Operation), request.WithDockerRestart, s.addPortsBeforeStart); err != nil {
		return err
	}
	if request.Operation == "start" || request.Operation == "restart" {
		ReconcileDockerPortGuardBestEffort(context.Background())
	}
	return nil
}

func (s *FirewallService) OperateFilterChain(request dto.FilterChainOperation) error {
	provider, err := selectedSystemFirewallProvider()
	if err != nil {
		return err
	}
	if !supportsManagedFilterChains(provider) {
		return fmt.Errorf("filter chain operations are not supported for %s", provider)
	}
	if provider == constant.FirewallProviderNftables {
		if err := newNftablesHelperManager().Operate(firewall.BaseOperation(request.Operate)); err != nil {
			return err
		}
	} else if err := s.iptablesHelper.Operate(firewall.BaseOperation(request.Operate)); err != nil {
		return err
	}
	if request.Operate != string(firewall.BaseOperationInit) && request.Operate != string(firewall.BaseOperationBind) {
		return nil
	}
	configured, err := loadConfiguredFirewallPortWhiteList()
	if err != nil {
		return err
	}
	required, err := loadRequiredFirewallPortWhiteList()
	if err != nil {
		return err
	}
	ports := excludeFirewallPorts(configured, required)
	return s.SyncSystemPorts(context.Background(), nil, systemPorts(ports))
}

func (s *FirewallService) Reset(ctx context.Context, request dto.FirewallRuleReset) (dto.FirewallRuleResetResponse, error) {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()

	provider := request.Provider
	selected := provider
	if provider == "" {
		var err error
		selected, err = s.selectedProvider(ctx)
		if err != nil {
			return dto.FirewallRuleResetResponse{}, err
		}
		provider = selected
	} else if isDirectFirewallProvider(provider) {
		var err error
		selected, err = s.selectedProvider(ctx)
		if err != nil {
			return dto.FirewallRuleResetResponse{}, err
		}
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		return dto.FirewallRuleResetResponse{}, err
	}
	if provider == filter.ProviderIptables || provider == filter.ProviderNftables {
		cleanup := s.cleanupBackend
		if isDirectFirewallProvider(selected) && selected != provider {
			cleanup = s.cleanupInactiveBackend
			if cleanup == nil {
				cleanup = cleanupInactiveSystemBackend
			}
		} else if cleanup == nil {
			cleanup = cleanupSystemBackend
		}
		if err := cleanup(string(provider)); err != nil {
			return dto.FirewallRuleResetResponse{}, err
		}
		return dto.FirewallRuleResetResponse{Removed: len(stored), Disabled: true}, nil
	}
	if provider != filter.ProviderUFW && provider != filter.ProviderFirewalld {
		return dto.FirewallRuleResetResponse{}, fmt.Errorf("%w: unsupported firewall provider %s", filter.ErrProviderUnavailable, provider)
	}
	reset := s.resetBackend
	if reset == nil {
		reset = resetServiceFirewallBackend
	}
	if err := reset(string(provider)); err != nil {
		return dto.FirewallRuleResetResponse{}, err
	}
	return dto.FirewallRuleResetResponse{Removed: len(stored), Disabled: true}, nil
}

func isDirectFirewallProvider(provider filter.Provider) bool {
	return provider == filter.ProviderIptables || provider == filter.ProviderNftables
}

func resetServiceFirewallBackend(provider string) error {
	client, err := lifecycle.NewClientFor(provider)
	if err != nil {
		return err
	}
	resetter, ok := client.(lifecycle.Resetter)
	if !ok {
		return fmt.Errorf("firewall provider %s does not support reset", provider)
	}
	if provider == constant.FirewallProviderFirewalld {
		if err := lifecycle.NewOperator(client).Operate(lifecycle.OperationStop, false, nil); err != nil {
			return err
		}
	}
	return resetter.Reset()
}

func (s *FirewallService) deleteFirewallRuleRecords(
	ctx context.Context,
	stored []model.FirewallRule,
) (int, error) {
	deleted := 0
	for _, record := range stored {
		if err := s.rules.DeleteWithRevision(ctx, record.UUID, record.Revision); err != nil {
			return deleted, fmt.Errorf("delete reset firewall rule %q: %w", record.UUID, err)
		}
		deleted++
	}
	return deleted, nil
}

func (s *FirewallService) loadProtectedPorts() ([]firewall.PortWhitelist, error) {
	if s.protectedPorts == nil {
		return nil, nil
	}
	return s.protectedPorts()
}

func (s *FirewallService) Inventory(ctx context.Context, request dto.FirewallRuleInventory) (dto.FirewallRuleInventoryResponse, error) {
	scope := request.Scope.Normalize()
	if isCombinedUFWInventoryScope(scope) {
		if err := s.checkSelectedProvider(ctx, scope.Provider); err != nil {
			return dto.FirewallRuleInventoryResponse{}, err
		}
		runtime, err := s.adapters.Resolve(scope.Provider)
		if err != nil {
			return dto.FirewallRuleInventoryResponse{}, err
		}
		return s.combinedUFWInventory(ctx, runtime, scope)
	}
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
	stored, err := s.rules.List(ctx)
	if err != nil {
		return dto.FirewallRuleInventoryResponse{}, err
	}
	desired, err := s.desiredFirewallRulesForScope(ctx, stored, scope)
	if err != nil {
		return dto.FirewallRuleInventoryResponse{}, err
	}
	items, err := filter.MergeInventory(filter.InventoryMergeInput{Observed: snapshot.Rules, Desired: desired})
	if err != nil {
		return dto.FirewallRuleInventoryResponse{}, err
	}
	return dto.FirewallRuleInventoryResponse{Items: items, Notices: snapshot.Notices}, nil
}

func isCombinedUFWInventoryScope(scope filter.Scope) bool {
	scope = scope.Normalize()
	return scope.Provider == filter.ProviderUFW && scope.Family == filter.FamilyInet && scope.Table == "" &&
		scope.Zone == "" && scope.Chain == filter.UFWInputChain && scope.Direction == filter.DirectionInput
}

func (s *FirewallService) combinedUFWInventory(
	ctx context.Context,
	runtime *firewallRuleRuntime,
	scope filter.Scope,
) (dto.FirewallRuleInventoryResponse, error) {
	scopes := []filter.Scope{scope, scope}
	scopes[0].Family = filter.FamilyIPv4
	scopes[1].Family = filter.FamilyIPv6
	snapshots, err := runtime.ObserveScopes(ctx, scopes)
	if err != nil {
		return dto.FirewallRuleInventoryResponse{}, err
	}
	if len(snapshots) != len(scopes) {
		return dto.FirewallRuleInventoryResponse{}, fmt.Errorf("%w: incomplete UFW multi-family inventory", filter.ErrAdapterUnavailable)
	}

	response := dto.FirewallRuleInventoryResponse{}
	seenNotices := make(map[string]struct{})
	for index, snapshot := range snapshots {
		if snapshot.Scope.Key() != scopes[index].Key() {
			return dto.FirewallRuleInventoryResponse{}, fmt.Errorf("%w: unexpected UFW inventory scope %q", filter.ErrInvalidScope, snapshot.Scope.Key())
		}
		stored, err := s.rules.List(ctx)
		if err != nil {
			return dto.FirewallRuleInventoryResponse{}, err
		}
		desired, err := s.desiredFirewallRulesForScope(ctx, stored, snapshot.Scope)
		if err != nil {
			return dto.FirewallRuleInventoryResponse{}, err
		}
		items, err := filter.MergeInventory(filter.InventoryMergeInput{Observed: snapshot.Rules, Desired: desired})
		if err != nil {
			return dto.FirewallRuleInventoryResponse{}, err
		}
		response.Items = append(response.Items, items...)
		for _, notice := range snapshot.Notices {
			key := string(notice.Code) + "\x00" + strings.Join(notice.Values, "\x00")
			if _, exists := seenNotices[key]; exists {
				continue
			}
			seenNotices[key] = struct{}{}
			response.Notices = append(response.Notices, notice)
		}
	}
	return response, nil
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

func (s *FirewallService) checkUpdate(
	ctx context.Context,
	clientIP string,
	ruleUUID string,
	requestedRule filter.FirewallRule,
) (dto.FirewallRuleCheckResult, error) {
	prepared, err := s.prepareManagedUpdate(ctx, clientIP, ruleUUID, requestedRule)
	if err != nil {
		return dto.FirewallRuleCheckResult{}, err
	}
	semantic, err := model.FirewallRuleFromDomain(prepared.After)
	if err != nil {
		return dto.FirewallRuleCheckResult{}, err
	}
	if err := s.ensureFirewallRuleIdentityAvailable(ctx, semantic, prepared.Stored.UUID); err != nil {
		return dto.FirewallRuleCheckResult{}, err
	}
	return dto.FirewallRuleCheckResult{
		Decision:         filter.CheckDecisionReady,
		Classification:   filter.CheckClassificationNone,
		Reason:           "update_ready",
		RequestedRule:    prepared.After,
		RequestedRuleKey: semantic.PolicyKey(),
	}, nil
}

func (s *FirewallService) Check(
	ctx context.Context,
	clientIP string,
	request dto.FirewallRuleCheck,
) (dto.FirewallRuleCheckResponse, error) {
	response := dto.FirewallRuleCheckResponse{Items: make([]dto.FirewallRuleCheckResult, 0, len(request.Items))}
	var protectedPorts []firewall.PortWhitelist
	var selectedProvider filter.Provider
	createStateLoaded := false
	type checkState struct {
		snapshot        filter.Snapshot
		desired         []filter.DesiredRule
		managedRevision string
	}
	states := make(map[string]checkState)
	for _, item := range request.Items {
		if ruleUUID := strings.TrimSpace(item.UUID); ruleUUID != "" {
			result, updateErr := s.checkUpdate(ctx, clientIP, ruleUUID, item.Rule)
			if updateErr != nil {
				return dto.FirewallRuleCheckResponse{}, updateErr
			}
			response.Items = append(response.Items, result)
			continue
		}
		if !createStateLoaded {
			var err error
			protectedPorts, err = s.loadProtectedPorts()
			if err != nil {
				return dto.FirewallRuleCheckResponse{}, err
			}
			selectedProvider, err = s.selectedProvider(ctx)
			if err != nil {
				return dto.FirewallRuleCheckResponse{}, err
			}
			createStateLoaded = true
		}
		item.Rule = applySelectedProviderScopeDefaults(item.Rule, selectedProvider)
		rule, err := filter.NormalizeRule(item.Rule)
		if err != nil {
			return dto.FirewallRuleCheckResponse{}, err
		}
		if rule.Scope.Provider != selectedProvider {
			return dto.FirewallRuleCheckResponse{}, fmt.Errorf(
				"%w: selected provider is %s, requested %s", filter.ErrProviderUnavailable, selectedProvider, rule.Scope.Provider,
			)
		}
		runtime, err := s.adapters.Resolve(rule.Scope.Provider)
		if err != nil {
			return dto.FirewallRuleCheckResponse{}, err
		}
		rule, err = runtime.Prepare(rule)
		if err != nil {
			return dto.FirewallRuleCheckResponse{}, err
		}
		if err = runtime.CheckRule(ctx, rule); err != nil {
			return dto.FirewallRuleCheckResponse{}, err
		}

		scopeKey := rule.Scope.Key()
		state, exists := states[scopeKey]
		if !exists {
			snapshot, observeErr := runtime.ObserveMutation(ctx, rule.Scope)
			if observeErr != nil {
				return dto.FirewallRuleCheckResponse{}, observeErr
			}
			stored, listErr := s.rules.List(ctx)
			if listErr != nil {
				return dto.FirewallRuleCheckResponse{}, listErr
			}
			desired, desiredErr := s.desiredFirewallRulesForScope(ctx, stored, rule.Scope)
			if desiredErr != nil {
				return dto.FirewallRuleCheckResponse{}, desiredErr
			}
			managedRevision, revisionErr := firewallManagedRevision(stored)
			if revisionErr != nil {
				return dto.FirewallRuleCheckResponse{}, revisionErr
			}
			state = checkState{snapshot: snapshot, desired: desired, managedRevision: managedRevision}
			states[scopeKey] = state
		}
		checked, checkErr := filter.CheckCreate(state.snapshot, rule, state.desired, clientIP, protectedPorts...)
		if checkErr != nil {
			return dto.FirewallRuleCheckResponse{}, checkErr
		}
		checkFlag, signErr := signFirewallRuleCheck(checked, state.snapshot, state.managedRevision)
		if signErr != nil {
			return dto.FirewallRuleCheckResponse{}, signErr
		}
		result := dto.FirewallRuleCheckResult{
			Decision: checked.Decision, Classification: checked.Classification, Reason: checked.Reason,
			RequestedRule: checked.RequestedRule, RequestedRuleKey: checked.RequestedRuleKey,
			ExistingRuleUUID: checked.ExistingRuleUUID, Candidates: checked.Candidates,
			AllowedActions: checked.AllowedActions, CheckFlag: checkFlag,
		}
		response.Items = append(response.Items, result)
	}
	return response, nil
}

func applySelectedProviderScopeDefaults(rule filter.FirewallRule, selected filter.Provider) filter.FirewallRule {
	scope := rule.Scope
	if scope.Provider == "" {
		scope.Provider = selected
	}
	if scope.Provider != selected {
		return rule
	}
	if scope.Direction == "" {
		scope.Direction = filter.DirectionInput
	}
	if scope.Family == "" {
		scope.Family = defaultFirewallRuleFamily(rule, selected)
	}
	switch selected {
	case filter.ProviderIptables, filter.ProviderNftables:
		if scope.Table == "" {
			scope.Table = "filter"
		}
		if scope.Chain == "" {
			scope.Chain = filter.IptablesInputChain
		}
	case filter.ProviderFirewalld:
		if scope.Zone == "" {
			scope.Zone = filter.FirewalldInputZone
		}
	case filter.ProviderUFW:
		if scope.Chain == "" {
			scope.Chain = filter.UFWInputChain
		}
	}
	rule.Scope = scope
	return rule
}

func defaultFirewallRuleFamily(rule filter.FirewallRule, provider filter.Provider) filter.Family {
	if strings.EqualFold(strings.TrimSpace(rule.Protocol), "icmpv6") ||
		strings.Contains(rule.SourceAddress, ":") || strings.Contains(rule.DestinationAddress, ":") {
		return filter.FamilyIPv6
	}
	if provider == filter.ProviderFirewalld {
		return filter.FamilyInet
	}
	return filter.FamilyIPv4
}

func (s *FirewallService) checkRule(
	ctx context.Context,
	clientIP string,
	item dto.FirewallRuleCheckItem,
) (dto.FirewallRuleCheckResult, error) {
	response, err := s.Check(ctx, clientIP, dto.FirewallRuleCheck{Items: []dto.FirewallRuleCheckItem{item}})
	if err != nil {
		return dto.FirewallRuleCheckResult{}, err
	}
	if len(response.Items) != 1 {
		return dto.FirewallRuleCheckResult{}, errors.New("firewall rule check returned no result")
	}
	return response.Items[0], nil
}

type preparedFirewallRuleCreate struct {
	request       dto.FirewallRuleCreateItem
	runtime       *firewallRuleRuntime
	snapshot      filter.Snapshot
	authorization firewallRuleCreateAuthorization
}

func (s *FirewallService) Create(
	ctx context.Context,
	request dto.FirewallRuleCreate,
) (dto.FirewallRuleCreateResponse, error) {
	result, _ := s.create(ctx, request)
	return result, nil
}

func (s *FirewallService) create(
	ctx context.Context,
	request dto.FirewallRuleCreate,
) (dto.FirewallRuleCreateResponse, error) {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()

	prepared, failedIndex, err := s.prepareCreate(ctx, request.Items)
	if err != nil {
		return firewallCreatePrepareFailure(request.Items, failedIndex, err), err
	}

	result := dto.FirewallRuleCreateResponse{}
	var createErr error
	for index := 0; index < len(prepared); {
		batchEnd := nativeCreateBatchEnd(prepared, index)
		if batchEnd-index > 1 {
			if err := s.createNativeRuleBatch(ctx, prepared[index:batchEnd]); err != nil {
				failed := firewallCreateExecutionFailure(request.Items, index, err)
				result.Failed += failed.Failed
				result.Skipped += failed.Skipped
				result.Errors = append(result.Errors, failed.Errors...)
				if global.LOG != nil {
					global.LOG.Errorf("batch create firewall rules %d-%d failed: %v", index+1, batchEnd, err)
				}
				createErr = err
				break
			}
			result.Succeeded += batchEnd - index
			index = batchEnd
			continue
		}

		entry := prepared[index]
		snapshot, err := entry.runtime.ObserveMutation(ctx, entry.request.Rule.Scope)
		if err == nil {
			entry.authorization, err = refreshCreateAuthorization(snapshot, entry)
		}
		if err == nil {
			err = s.createRule(ctx, entry.runtime, snapshot, entry.request, entry.authorization)
		}
		if err != nil {
			failed := firewallCreateExecutionFailure(request.Items, index, err)
			result.Failed += failed.Failed
			result.Skipped += failed.Skipped
			result.Errors = append(result.Errors, failed.Errors...)
			if global.LOG != nil {
				global.LOG.Errorf("batch create firewall rule item %d failed: %v", index+1, err)
			}
			createErr = err
			break
		}
		result.Succeeded++
		index++
	}
	return result, createErr
}

func (s *FirewallService) prepareCreate(
	ctx context.Context,
	items []dto.FirewallRuleCreateItem,
) ([]preparedFirewallRuleCreate, int, error) {
	type prepareState struct {
		runtime         *firewallRuleRuntime
		snapshot        filter.Snapshot
		managedRevision string
	}
	states := make(map[string]prepareState)
	prepared := make([]preparedFirewallRuleCreate, 0, len(items))
	selectedProvider, err := s.selectedProvider(ctx)
	if err != nil {
		return nil, 0, err
	}
	for index, request := range items {
		if request.CheckFlag == "" {
			return nil, index, filter.ErrRuleCheckRequired
		}
		rule, err := filter.NormalizeRule(request.Rule)
		if err != nil {
			return nil, index, err
		}
		if rule.Scope.Provider != selectedProvider {
			return nil, index, fmt.Errorf(
				"%w: selected provider is %s, requested %s", filter.ErrProviderUnavailable, selectedProvider, rule.Scope.Provider,
			)
		}
		runtime, err := s.adapters.Resolve(rule.Scope.Provider)
		if err != nil {
			return nil, index, err
		}
		rule, err = runtime.Prepare(rule)
		if err != nil {
			return nil, index, err
		}
		if err = runtime.CheckRule(ctx, rule); err != nil {
			return nil, index, err
		}

		scopeKey := rule.Scope.Key()
		state, exists := states[scopeKey]
		if !exists {
			snapshot, observeErr := runtime.ObserveMutation(ctx, rule.Scope)
			if observeErr != nil {
				return nil, index, observeErr
			}
			stored, listErr := s.rules.List(ctx)
			if listErr != nil {
				return nil, index, listErr
			}
			managedRevision, revisionErr := firewallManagedRevision(stored)
			if revisionErr != nil {
				return nil, index, revisionErr
			}
			state = prepareState{runtime: runtime, snapshot: snapshot, managedRevision: managedRevision}
			states[scopeKey] = state
		}
		authorization, authorizeErr := authorizeFirewallRuleCreate(
			request.CheckFlag, request.Action, request.AdoptInstanceKey, rule, state.snapshot, state.managedRevision,
		)
		if authorizeErr != nil {
			return nil, index, authorizeErr
		}
		sourceKind := request.SourceKind
		if sourceKind == "" {
			sourceKind = constant.FirewallRuleSourceUser
		}
		request.Rule = rule
		request.SourceKind = sourceKind
		prepared = append(prepared, preparedFirewallRuleCreate{
			request: request, runtime: state.runtime, snapshot: state.snapshot, authorization: authorization,
		})
	}
	return prepared, -1, nil
}

func nativeCreateBatchEnd(prepared []preparedFirewallRuleCreate, start int) int {
	if start < 0 || start >= len(prepared) {
		return start
	}
	first := prepared[start]
	if first.runtime == nil || first.runtime.adapter == nil || !supportsNativeRuleBatch(first.runtime.adapter.Provider()) ||
		first.authorization.Operation != filter.ChangeCreate || first.request.Rule.OrderIndex != nil {
		return start + 1
	}
	scopeKey := first.request.Rule.Scope.Key()
	seen := make(map[string]struct{}, len(prepared)-start)
	for index := start; index < len(prepared); index++ {
		entry := prepared[index]
		if entry.runtime != first.runtime || entry.authorization.Operation != filter.ChangeCreate ||
			entry.request.Rule.Scope.Key() != scopeKey || entry.request.Rule.OrderIndex != nil {
			return index
		}
		ruleKey, err := filter.RuleKey(entry.request.Rule)
		if err != nil {
			return index
		}
		if _, exists := seen[ruleKey]; exists {
			return index
		}
		seen[ruleKey] = struct{}{}
	}
	return len(prepared)
}

type createdFirewallBatchRule struct {
	record model.FirewallRule
	rule   filter.FirewallRule
}

func (s *FirewallService) createNativeRuleBatch(ctx context.Context, prepared []preparedFirewallRuleCreate) error {
	if len(prepared) < 2 {
		return fmt.Errorf("%w: native firewall batch requires at least two rules", filter.ErrInvalidRule)
	}
	runtime := prepared[0].runtime
	snapshot, err := runtime.ObserveMutation(ctx, prepared[0].request.Rule.Scope)
	if err != nil {
		return err
	}
	created := make([]createdFirewallBatchRule, 0, len(prepared))
	nextSequence, err := s.nextFirewallRuleSequence(ctx)
	if err != nil {
		return err
	}
	for _, entry := range prepared {
		domainRule := entry.request.Rule
		record, recordErr := firewallRuleModelForCreate(domainRule, entry.request, constant.FirewallRuleOriginCreated)
		if recordErr != nil {
			return s.cleanupFirewallBatchRecords(ctx, created, recordErr)
		}
		if domainRule.Scope.Provider != filter.ProviderFirewalld {
			sequence := nextSequence
			record.Sequence = &sequence
			nextSequence += model.FirewallRuleSequenceStep
		}
		if recordErr = s.ensureFirewallRuleIdentityAvailable(ctx, record, ""); recordErr != nil {
			return s.cleanupFirewallBatchRecords(ctx, created, recordErr)
		}
		if recordErr = s.rules.Create(ctx, &record); recordErr != nil {
			return s.cleanupFirewallBatchRecords(ctx, created, recordErr)
		}
		domainRule.UUID = record.UUID
		created = append(created, createdFirewallBatchRule{record: record, rule: domainRule})
	}

	changes := make([]filter.DesiredChange, 0, len(created))
	for index := range created {
		changes = append(changes, filter.DesiredChange{Operation: filter.ChangeCreate, After: &created[index].rule})
	}
	backendPlan, verification, err := runtime.Execute(ctx, snapshot, changes)
	if err != nil {
		return s.cleanupFirewallBatchRecords(ctx, created, err)
	}
	if !verification.Matched {
		return s.cleanupFirewallBatchRecords(ctx, created, filter.ErrVerificationFailed)
	}

	for index := range created {
		_, commitErr := findBatchCommittedObserved(verification.Snapshot, created[index].rule.UUID)
		if commitErr != nil {
			commitErr = rollbackFirewallPlan(ctx, runtime, backendPlan, commitErr)
			return s.cleanupFirewallBatchRecords(ctx, created, commitErr)
		}
	}
	return nil
}

func findBatchCommittedObserved(snapshot filter.Snapshot, ruleUUID string) (filter.ObservedRule, error) {
	marker := "1panel-rule:" + ruleUUID
	matches := make([]filter.ObservedRule, 0, 1)
	for _, observed := range snapshot.Rules {
		if observed.Marker == marker {
			matches = append(matches, observed)
		}
	}
	if len(matches) != 1 {
		return filter.ObservedRule{}, fmt.Errorf("%w: expected one committed batch rule, found %d", filter.ErrVerificationFailed, len(matches))
	}
	return matches[0], nil
}

func (s *FirewallService) cleanupFirewallBatchRecords(
	ctx context.Context,
	created []createdFirewallBatchRule,
	cause error,
) error {
	cleanupErrors := make([]error, 0)
	for index := len(created) - 1; index >= 0; index-- {
		if err := s.rules.DeleteWithRevision(ctx, created[index].record.UUID, created[index].record.Revision); err != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("cleanup failed firewall rule %q: %w", created[index].record.UUID, err))
		}
	}
	if len(cleanupErrors) == 0 {
		return cause
	}
	return errors.Join(append([]error{cause}, cleanupErrors...)...)
}

func firewallCreatePrepareFailure(
	items []dto.FirewallRuleCreateItem,
	failedIndex int,
	cause error,
) dto.FirewallRuleCreateResponse {
	result := dto.FirewallRuleCreateResponse{
		Failed: 1, Skipped: len(items) - 1,
		Errors: make([]dto.FirewallRuleCreateFailure, 0, len(items)),
	}
	for index := range items {
		failure := dto.FirewallRuleCreateFailure{
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

func firewallCreateExecutionFailure(
	items []dto.FirewallRuleCreateItem,
	failedIndex int,
	cause error,
) dto.FirewallRuleCreateResponse {
	result := dto.FirewallRuleCreateResponse{
		Failed: 1, Skipped: len(items) - failedIndex - 1,
		Errors: make([]dto.FirewallRuleCreateFailure, 0, len(items)-failedIndex),
	}
	for index := failedIndex; index < len(items); index++ {
		failure := dto.FirewallRuleCreateFailure{
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

type preparedFirewallRuleDelete struct {
	index    int
	stored   model.FirewallRule
	desired  filter.DesiredRule
	runtime  *firewallRuleRuntime
	compiled int
}

func (s *FirewallService) Delete(
	ctx context.Context,
	request dto.FirewallRuleDelete,
) (dto.FirewallRuleDeleteResponse, error) {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()

	result := dto.FirewallRuleDeleteResponse{}
	selectedProvider, err := s.selectedProvider(ctx)
	if err != nil {
		return dto.FirewallRuleDeleteResponse{}, err
	}
	type deleteGroup struct {
		items []preparedFirewallRuleDelete
	}
	groups := make([]deleteGroup, 0)
	groupIndexes := make(map[string]int)
	seen := make(map[string]struct{}, len(request.UUIDs))
	for index, value := range request.UUIDs {
		ruleUUID := strings.TrimSpace(value)
		if _, exists := seen[ruleUUID]; exists {
			result.Failed++
			result.Errors = append(result.Errors, dto.FirewallRuleDeleteFailure{
				Index: index, UUID: ruleUUID, Error: "duplicate firewall rule UUID",
			})
			continue
		}
		seen[ruleUUID] = struct{}{}
		prepared, err := s.prepareDelete(ctx, index, ruleUUID, selectedProvider)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, dto.FirewallRuleDeleteFailure{Index: index, UUID: ruleUUID, Error: err.Error()})
			continue
		}
		groupKey := string(prepared.desired.Rule.Scope.Provider) + ":" + prepared.desired.Rule.Scope.Key()
		groupIndex, exists := groupIndexes[groupKey]
		if !exists {
			groupIndex = len(groups)
			groupIndexes[groupKey] = groupIndex
			groups = append(groups, deleteGroup{})
		}
		groups[groupIndex].items = append(groups[groupIndex].items, prepared)
	}

	for _, group := range groups {
		if len(group.items) > 1 && group.items[0].compiled == 1 && supportsNativeRuleBatch(group.items[0].desired.Rule.Scope.Provider) {
			if err := s.deleteNativeRuleBatch(ctx, group.items); err != nil {
				result.Failed += len(group.items)
				for _, item := range group.items {
					result.Errors = append(result.Errors, dto.FirewallRuleDeleteFailure{
						Index: item.index, UUID: item.stored.UUID, Error: err.Error(),
					})
				}
				continue
			}
			result.Succeeded += len(group.items)
			continue
		}
		for _, item := range group.items {
			if err := s.deleteRule(ctx, item.stored.UUID); err != nil {
				result.Failed++
				result.Errors = append(result.Errors, dto.FirewallRuleDeleteFailure{
					Index: item.index, UUID: item.stored.UUID, Error: err.Error(),
				})
				continue
			}
			result.Succeeded++
		}
	}
	sort.SliceStable(result.Errors, func(i, j int) bool { return result.Errors[i].Index < result.Errors[j].Index })
	return result, nil
}

func (s *FirewallService) prepareDelete(
	ctx context.Context,
	index int,
	ruleUUID string,
	selectedProvider filter.Provider,
) (preparedFirewallRuleDelete, error) {
	if ruleUUID == "" {
		return preparedFirewallRuleDelete{}, fmt.Errorf("%w: rule UUID is required", repo.ErrFirewallPersistenceInvalid)
	}
	stored, err := s.rules.GetByUUID(ctx, ruleUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return preparedFirewallRuleDelete{}, fmt.Errorf("%w: managed rule %q was not found", filter.ErrInvalidRule, ruleUUID)
		}
		return preparedFirewallRuleDelete{}, err
	}
	if stored.Origin != constant.FirewallRuleOriginCreated && stored.Origin != constant.FirewallRuleOriginAdopted {
		return preparedFirewallRuleDelete{}, fmt.Errorf("%w: only created or adopted rules can be deleted", filter.ErrInvalidRule)
	}
	desiredRules, err := s.compileStoredFirewallRules(ctx, stored, selectedProvider)
	if err != nil {
		return preparedFirewallRuleDelete{}, err
	}
	if len(desiredRules) == 0 {
		return preparedFirewallRuleDelete{}, fmt.Errorf("%w: policy %q has no compiled target rules", filter.ErrInvalidRule, ruleUUID)
	}
	desired := desiredRules[0]
	runtime, err := s.adapters.Resolve(desired.Rule.Scope.Provider)
	if err != nil {
		return preparedFirewallRuleDelete{}, err
	}
	return preparedFirewallRuleDelete{
		index: index, stored: stored, desired: desired, runtime: runtime, compiled: len(desiredRules),
	}, nil
}

func (s *FirewallService) deleteNativeRuleBatch(ctx context.Context, prepared []preparedFirewallRuleDelete) error {
	runtime := prepared[0].runtime
	snapshot, err := runtime.ObserveMutation(ctx, prepared[0].desired.Rule.Scope)
	if err != nil {
		return err
	}
	type positionedDelete struct {
		position int
		change   filter.DesiredChange
	}
	positioned := make([]positionedDelete, 0, len(prepared))
	for _, item := range prepared {
		observed, observeErr := filter.ManagedObserved(snapshot, item.desired)
		if observeErr != nil {
			return observeErr
		}
		if observed.Locator.Position == nil {
			return fmt.Errorf("%w: managed native firewall rule has no position", filter.ErrRuleStale)
		}
		before := item.desired.Rule
		locator := observed.Locator
		positioned = append(positioned, positionedDelete{
			position: *observed.Locator.Position,
			change: filter.DesiredChange{
				Operation: filter.ChangeDelete, Before: &before, Locator: &locator,
			},
		})
	}
	sort.Slice(positioned, func(i, j int) bool { return positioned[i].position > positioned[j].position })
	changes := make([]filter.DesiredChange, 0, len(positioned))
	for _, item := range positioned {
		changes = append(changes, item.change)
	}
	backendPlan, verification, err := runtime.Execute(ctx, snapshot, changes)
	if err != nil {
		return err
	}
	if !verification.Matched {
		return filter.ErrVerificationFailed
	}

	deleted := make([]model.FirewallRule, 0, len(prepared))
	for _, item := range prepared {
		if err = s.rules.DeleteWithRevision(ctx, item.stored.UUID, item.stored.Revision); err != nil {
			err = rollbackFirewallPlan(ctx, runtime, backendPlan, err)
			return s.restoreDeletedFirewallRecords(ctx, deleted, err)
		}
		deleted = append(deleted, item.stored)
	}
	return nil
}

func supportsNativeRuleBatch(provider filter.Provider) bool {
	return provider == filter.ProviderIptables || provider == filter.ProviderNftables
}

func (s *FirewallService) restoreDeletedFirewallRecords(
	ctx context.Context,
	deleted []model.FirewallRule,
	cause error,
) error {
	restoreErrors := make([]error, 0)
	for index := range deleted {
		record := deleted[index]
		if err := s.rules.Create(ctx, &record); err != nil {
			restoreErrors = append(restoreErrors, fmt.Errorf("restore deleted firewall rule %q: %w", record.UUID, err))
		}
	}
	if len(restoreErrors) == 0 {
		return cause
	}
	return errors.Join(append([]error{cause}, restoreErrors...)...)
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

func (s *FirewallService) createRule(
	ctx context.Context,
	runtime *firewallRuleRuntime,
	snapshot filter.Snapshot,
	request dto.FirewallRuleCreateItem,
	authorization firewallRuleCreateAuthorization,
) error {
	domainRule := request.Rule
	appendRule := false
	if authorization.Operation == filter.ChangeCreate && domainRule.Scope.Provider == filter.ProviderUFW && domainRule.OrderIndex == nil {
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
	if err := s.ensureFirewallRuleIdentityAvailable(ctx, ruleRecord, ""); err != nil {
		return err
	}
	if domainRule.Scope.Provider != filter.ProviderFirewalld {
		sequence, sequenceErr := s.sequenceForCreatedFirewallRule(ctx, snapshot, domainRule)
		if sequenceErr != nil {
			return sequenceErr
		}
		ruleRecord.Sequence = &sequence
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
	_, err = filter.FindCommittedObserved(verification.Snapshot, domainRule, backendPlan)
	if err != nil {
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
	selected, err := s.selectedProviderForStoredRule(ctx, stored)
	if err != nil {
		return err
	}
	desiredRules, err := s.compileStoredFirewallRules(ctx, stored, selected)
	if err != nil {
		return err
	}
	type appliedDelete struct {
		runtime *firewallRuleRuntime
		plan    filter.BackendPlan
	}
	applied := make([]appliedDelete, 0, len(desiredRules))
	rollback := func(cause error) error {
		for index := len(applied) - 1; index >= 0; index-- {
			cause = rollbackFirewallPlan(ctx, applied[index].runtime, applied[index].plan, cause)
		}
		return cause
	}
	for _, desired := range desiredRules {
		runtime, runtimeErr := s.resolveRuntime(ctx, desired.Rule.Scope.Provider)
		if runtimeErr != nil {
			return rollback(runtimeErr)
		}
		snapshot, observeErr := runtime.ObserveMutation(ctx, desired.Rule.Scope)
		if observeErr != nil {
			return rollback(observeErr)
		}
		observed, managedErr := filter.ManagedObserved(snapshot, desired)
		if managedErr != nil {
			if errors.Is(managedErr, filter.ErrRuleStale) {
				items, mergeErr := filter.MergeInventory(filter.InventoryMergeInput{
					Observed: snapshot.Rules,
					Desired:  []filter.DesiredRule{desired},
				})
				if mergeErr != nil {
					return rollback(mergeErr)
				}
				missing := false
				for _, item := range items {
					if item.Desired != nil && item.Desired.UUID == desired.UUID {
						missing = item.Match == filter.InventoryMatchMissing && item.Observed == nil
						break
					}
				}
				if missing {
					continue
				}
			}
			return rollback(managedErr)
		}
		restoreAtEnd := false
		if desired.Rule.Scope.Provider == filter.ProviderUFW && observed.Locator.Position != nil {
			maxPosition, maxErr := maxPositionForRule(ctx, runtime, snapshot, desired.Rule)
			if maxErr != nil {
				return rollback(maxErr)
			}
			restoreAtEnd = int64(*observed.Locator.Position) == maxPosition
		}
		locator := observed.Locator
		before := desired.Rule
		backendPlan, verification, executeErr := runtime.Execute(ctx, snapshot, []filter.DesiredChange{{
			Operation: filter.ChangeDelete, Before: &before, Locator: &locator, RestoreAtEnd: restoreAtEnd,
		}})
		if executeErr != nil {
			return rollback(executeErr)
		}
		if !verification.Matched {
			return rollback(filter.ErrVerificationFailed)
		}
		applied = append(applied, appliedDelete{runtime: runtime, plan: backendPlan})
	}
	if err := s.rules.DeleteWithRevision(ctx, stored.UUID, stored.Revision); err != nil {
		return rollback(err)
	}
	return nil
}

func (s *FirewallService) updateRule(ctx context.Context, clientIP, ruleUUID string, requestedRule filter.FirewallRule) error {
	prepared, err := s.prepareManagedUpdate(ctx, clientIP, ruleUUID, requestedRule)
	if err != nil {
		return err
	}
	metadataOnly, err := isUFWMetadataOnlyUpdate(prepared.Before.Rule, prepared.After, prepared.Observed.Locator)
	if err != nil {
		return err
	}
	if metadataOnly {
		if prepared.After.Description == prepared.Before.Rule.Description {
			return nil
		}
		return s.rules.UpdateWithRevision(ctx, prepared.Stored.UUID, prepared.Stored.Revision, map[string]interface{}{
			"description": prepared.After.Description,
		})
	}
	return s.executeManagedMutation(ctx, managedMutationRequest{
		Stored: prepared.Stored, Before: prepared.Before.Rule, After: prepared.After,
		Snapshot: prepared.Snapshot, Locator: prepared.Observed.Locator,
		AdapterOperation: filter.ChangeUpdate, Runtime: prepared.Runtime,
	})
}

func isUFWMetadataOnlyUpdate(before, after filter.FirewallRule, locator filter.Locator) (bool, error) {
	if after.Scope.Provider != filter.ProviderUFW || locator.Position == nil || after.OrderIndex == nil {
		return false, nil
	}
	beforeKey, err := filter.RuleKey(before)
	if err != nil {
		return false, err
	}
	afterKey, err := filter.RuleKey(after)
	if err != nil {
		return false, err
	}
	return beforeKey == afterKey && *after.OrderIndex == int64(*locator.Position), nil
}

func (s *FirewallService) reorderRule(ctx context.Context, clientIP, ruleUUID string, targetPosition *int64, priority *int) error {
	if ruleUUID == "" {
		return fmt.Errorf("%w: rule UUID is required", repo.ErrFirewallPersistenceInvalid)
	}
	if priority != nil {
		stored, err := s.rules.GetByUUID(ctx, ruleUUID)
		if err != nil {
			return err
		}
		if stored.Priority == nil {
			return fmt.Errorf("%w: rule has no explicit reorderable priority", filter.ErrUnsupportedScope)
		}
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
		after.Priority = priority
		adapterOperation = filter.ChangeUpdate
	default:
		return fmt.Errorf("%w: provider does not support rule reordering", filter.ErrUnsupportedScope)
	}
	after, err = runtime.Prepare(after)
	if err != nil {
		return err
	}
	protectedPorts, err := s.loadProtectedPorts()
	if err != nil {
		return err
	}
	if err := filter.GuardMutation(snapshot, observed, after, clientIP, protectedPorts...); err != nil {
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

type preparedManagedUpdate struct {
	Stored   model.FirewallRule
	Before   filter.DesiredRule
	After    filter.FirewallRule
	Snapshot filter.Snapshot
	Observed filter.ObservedRule
	Runtime  *firewallRuleRuntime
}

func (s *FirewallService) prepareManagedUpdate(
	ctx context.Context,
	clientIP string,
	ruleUUID string,
	requestedRule filter.FirewallRule,
) (preparedManagedUpdate, error) {
	ruleUUID = strings.TrimSpace(ruleUUID)
	if ruleUUID == "" {
		return preparedManagedUpdate{}, fmt.Errorf("%w: rule UUID is required", repo.ErrFirewallPersistenceInvalid)
	}
	stored, before, snapshot, observed, runtime, err := s.loadManagedMutation(ctx, ruleUUID)
	if err != nil {
		return preparedManagedUpdate{}, err
	}
	after, err := filter.NormalizeRule(requestedRule)
	if err != nil {
		return preparedManagedUpdate{}, err
	}
	after.UUID = stored.UUID
	after, err = runtime.Prepare(after)
	if err != nil {
		return preparedManagedUpdate{}, err
	}
	if err := runtime.CheckRule(ctx, after); err != nil {
		return preparedManagedUpdate{}, err
	}
	if after.Scope.Key() != before.Rule.Scope.Key() {
		return preparedManagedUpdate{}, fmt.Errorf("%w: managed rule scope cannot be changed", filter.ErrUnsupportedScope)
	}
	if after.NativeKind != before.Rule.NativeKind {
		return preparedManagedUpdate{}, fmt.Errorf("%w: native rule conversion requires an explicit workflow", filter.ErrUnsupportedScope)
	}
	capabilities, err := runtime.Capabilities(ctx)
	if err != nil {
		return preparedManagedUpdate{}, err
	}
	if capabilities.ExplicitPosition || capabilities.OwnedChains {
		if observed.Locator.Position == nil {
			return preparedManagedUpdate{}, fmt.Errorf("%w: managed rule has no positional locator", filter.ErrInvalidRule)
		}
		currentPosition := int64(*observed.Locator.Position)
		if after.OrderIndex == nil {
			after.OrderIndex = &currentPosition
		} else if *after.OrderIndex != currentPosition {
			if err := validatePositionTarget(ctx, runtime, snapshot, before.Rule, *after.OrderIndex); err != nil {
				return preparedManagedUpdate{}, err
			}
		}
	}
	protectedPorts, err := s.loadProtectedPorts()
	if err != nil {
		return preparedManagedUpdate{}, err
	}
	if err := filter.GuardMutation(snapshot, observed, after, clientIP, protectedPorts...); err != nil {
		return preparedManagedUpdate{}, err
	}
	return preparedManagedUpdate{
		Stored: stored, Before: before, After: after, Snapshot: snapshot, Observed: observed, Runtime: runtime,
	}, nil
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
	selected, err := s.selectedProviderForStoredRule(ctx, stored)
	if err != nil {
		return model.FirewallRule{}, filter.DesiredRule{}, filter.Snapshot{}, filter.ObservedRule{}, nil, err
	}
	desiredRules, err := s.compileStoredFirewallRules(ctx, stored, selected)
	if err != nil {
		return model.FirewallRule{}, filter.DesiredRule{}, filter.Snapshot{}, filter.ObservedRule{}, nil, err
	}
	if len(desiredRules) != 1 {
		return model.FirewallRule{}, filter.DesiredRule{}, filter.Snapshot{}, filter.ObservedRule{}, nil,
			fmt.Errorf("%w: policy %q expands to %d target rules and cannot be edited atomically", filter.ErrUnsupportedScope, ruleUUID, len(desiredRules))
	}
	desired := desiredRules[0]
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

func (s *FirewallService) selectedProviderForStoredRule(
	ctx context.Context,
	_ model.FirewallRule,
) (filter.Provider, error) {
	if s.selectedProvider != nil {
		return s.selectedProvider(ctx)
	}
	if len(s.adapters) == 1 {
		for provider := range s.adapters {
			return provider, nil
		}
	}
	return "", fmt.Errorf("%w: selected provider is unavailable", filter.ErrProviderUnavailable)
}

func (s *FirewallService) executeManagedMutation(ctx context.Context, request managedMutationRequest) error {
	before, after := request.Before, request.After
	semantic, err := model.FirewallRuleFromDomain(after)
	if err != nil {
		return err
	}
	if err := s.ensureFirewallRuleIdentityAvailable(ctx, semantic, request.Stored.UUID); err != nil {
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
	_, err = filter.FindCommittedObserved(verification.Snapshot, request.After, backendPlan)
	if err != nil {
		return rollbackFirewallPlan(ctx, request.Runtime, backendPlan, err)
	}
	updates, err := firewallRuleSemanticUpdates(request.After)
	if err != nil {
		return rollbackFirewallPlan(ctx, request.Runtime, backendPlan, err)
	}
	if request.After.Scope.Provider == filter.ProviderFirewalld {
		updates["sequence"] = nil
	} else {
		position, positionErr := firewallRuleMarkerPosition(verification.Snapshot, request.Stored.UUID)
		if positionErr != nil {
			return rollbackFirewallPlan(ctx, request.Runtime, backendPlan, positionErr)
		}
		sequence, sequenceErr := s.sequenceForFirewallRulePosition(
			ctx, verification.Snapshot, position, request.Stored.UUID, request.Stored.Sequence,
		)
		if sequenceErr != nil {
			return rollbackFirewallPlan(ctx, request.Runtime, backendPlan, sequenceErr)
		}
		updates["sequence"] = sequence
	}
	if err := s.rules.UpdateWithRevision(ctx, request.Stored.UUID, request.Stored.Revision, updates); err != nil {
		return rollbackFirewallPlan(ctx, request.Runtime, backendPlan, err)
	}
	return nil
}

func (s *FirewallService) ensureFirewallRuleIdentityAvailable(
	ctx context.Context,
	requested model.FirewallRule,
	excludedUUID string,
) error {
	stored, err := s.rules.List(ctx)
	if err != nil {
		return err
	}
	requestedKey := requested.PolicyKey()
	for _, candidate := range stored {
		if candidate.UUID == excludedUUID {
			continue
		}
		if candidate.PolicyKey() == requestedKey {
			return fmt.Errorf("%w: equivalent managed rule already exists", filter.ErrRuleOperation)
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

func (s *FirewallService) ensureSystemPort(ctx context.Context, port dto.FirewallSystemPort) error {
	create, err := s.prepareSystemPortCreate(ctx, port)
	if err != nil || create == nil {
		return err
	}
	return s.createFirewallRuleItem(ctx, *create)
}

func (s *FirewallService) createFirewallRuleItem(ctx context.Context, item dto.FirewallRuleCreateItem) error {
	result, err := s.create(ctx, dto.FirewallRuleCreate{Items: []dto.FirewallRuleCreateItem{item}})
	if err != nil {
		return err
	}
	if result.Failed == 0 && result.Skipped == 0 {
		return nil
	}
	if len(result.Errors) > 0 && result.Errors[0].Error != "" {
		return errors.New(result.Errors[0].Error)
	}
	return errors.New("create firewall rule failed")
}

func (s *FirewallService) prepareSystemPortCreate(ctx context.Context, port dto.FirewallSystemPort) (*dto.FirewallRuleCreateItem, error) {
	provider, err := s.selectedProvider(ctx)
	if err != nil {
		return nil, err
	}
	rule := systemPortRule(provider, port)
	check, err := s.checkRule(ctx, "", dto.FirewallRuleCheckItem{Rule: rule})
	if err != nil {
		return nil, err
	}

	create := dto.FirewallRuleCreateItem{
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
			return nil, fmt.Errorf("%w: external port rule has no candidate", filter.ErrRuleStale)
		}
		if len(check.Candidates) == 1 {
			create.Action = filter.CheckActionAdopt
		} else {
			create.Action = filter.CheckActionSelectAdopt
		}
		create.AdoptInstanceKey = check.Candidates[0].InstanceKey
	case filter.CheckClassificationCovered:
		create.Action = filter.CheckActionCreateAnyway
	case filter.CheckClassificationConflict:
		if !containsFirewallCheckAction(check.AllowedActions, filter.CheckActionCreateAnyway) {
			return nil, fmt.Errorf("cannot manage accepted port %s/%s: %s", port.Port, port.Protocol, check.Reason)
		}
		create.Action = filter.CheckActionCreateAnyway
	case filter.CheckClassificationExactManaged:
		return nil, nil
	case filter.CheckClassificationProtected:
		if len(check.Candidates) > 0 {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: protected accepted port %s", filter.ErrProtectedRule, port.Port)
	default:
		return nil, fmt.Errorf("cannot manage accepted port %s/%s: %s", port.Port, port.Protocol, check.Reason)
	}
	return &create, nil
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
			result, err := s.Delete(ctx, dto.FirewallRuleDelete{UUIDs: []string{rule.UUID}})
			if err != nil {
				return err
			}
			if result.Failed > 0 {
				if len(result.Errors) > 0 {
					return errors.New(result.Errors[0].Error)
				}
				return errors.New("delete firewall rule failed")
			}
		}
	}
}

func (s *FirewallService) systemPortRecords(ctx context.Context, port dto.FirewallSystemPort) ([]model.FirewallRule, error) {
	records := make([]model.FirewallRule, 0)
	sourceIDs := []string{constant.FirewallSystemAcceptedPortSourcePrefix + systemPortKey(port)}
	if port.Family == constant.FirewallFamilyIPv4 {
		sourceIDs = append(sourceIDs, constant.FirewallSystemAcceptedPortSourcePrefix+legacySystemPortKey(port))
	}
	seen := make(map[string]struct{})
	for _, sourceID := range sourceIDs {
		items, listErr := s.rules.List(ctx,
			repo.WithFirewallRuleSource(constant.FirewallRuleSourceSecurity, sourceID),
		)
		if listErr != nil {
			return nil, listErr
		}
		for _, item := range items {
			if _, exists := seen[item.UUID]; exists {
				continue
			}
			seen[item.UUID] = struct{}{}
			records = append(records, item)
		}
	}
	return records, nil
}

func (s *FirewallService) adoptExternalSystemPort(ctx context.Context, port dto.FirewallSystemPort) (bool, error) {
	provider, err := s.selectedProvider(ctx)
	if err != nil {
		return false, err
	}
	check, err := s.checkRule(ctx, "", dto.FirewallRuleCheckItem{Rule: systemPortRule(provider, port)})
	if err != nil {
		return false, err
	}
	if check.Classification != filter.CheckClassificationExactExternal || len(check.Candidates) == 0 {
		return false, nil
	}
	create := dto.FirewallRuleCreateItem{
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
	if err := s.createFirewallRuleItem(ctx, create); err != nil {
		return false, err
	}
	return true, nil
}

func systemPortRule(provider filter.Provider, port dto.FirewallSystemPort) filter.FirewallRule {
	scope := filter.Scope{Provider: provider, Direction: filter.DirectionInput}
	family := filter.Family(strings.ToLower(strings.TrimSpace(port.Family)))
	switch provider {
	case filter.ProviderIptables, filter.ProviderNftables:
		if family != filter.FamilyIPv6 {
			family = filter.FamilyIPv4
		}
		scope.Family = family
		scope.Table = "filter"
	case filter.ProviderFirewalld:
		if family != filter.FamilyIPv4 && family != filter.FamilyIPv6 {
			family = filter.FamilyInet
		}
		scope.Family = family
		scope.Zone = filter.FirewalldInputZone
	case filter.ProviderUFW:
		if family != filter.FamilyIPv6 {
			family = filter.FamilyIPv4
		}
		scope.Family = family
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
		family := strings.ToLower(strings.TrimSpace(port.Family))
		if family != "" {
			family = string(normalized.Scope.Family)
		}
		item := dto.FirewallSystemPort{
			Family: family, Port: normalized.DestinationPort, Protocol: normalized.Protocol,
		}
		result[systemPortKey(item)] = item
	}
	return result, nil
}

func systemPortKey(port dto.FirewallSystemPort) string {
	key := legacySystemPortKey(port)
	if family := strings.ToLower(strings.TrimSpace(port.Family)); family != "" {
		return family + "/" + key
	}
	return key
}

func legacySystemPortKey(port dto.FirewallSystemPort) string {
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

func firewallRuleModelForCreate(rule filter.FirewallRule, request dto.FirewallRuleCreateItem, origin string) (model.FirewallRule, error) {
	record, err := model.FirewallRuleFromDomain(rule)
	if err != nil {
		return model.FirewallRule{}, err
	}
	record.Origin = origin
	record.Owner = model.FirewallRuleOwner(request.SourceKind, request.SourceID)
	return record, nil
}

func firewallRuleSemanticUpdates(rule filter.FirewallRule) (map[string]interface{}, error) {
	record, err := model.FirewallRuleFromDomain(rule)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"family": record.Family, "protocol": record.Protocol,
		"source_address": record.SourceAddress, "source_port": record.SourcePort,
		"destination_address": record.DestinationAddress, "destination_port": record.DestinationPort,
		"interface": record.Interface, "connection_states": record.ConnectionStates, "action": record.Action,
		"description": record.Description, "compatibility_error": "", "priority": record.Priority,
	}, nil
}

func (s *FirewallService) nextFirewallRuleSequence(ctx context.Context) (int64, error) {
	stored, err := s.rules.List(ctx)
	if err != nil {
		return 0, err
	}
	var maximum int64
	for _, record := range stored {
		if record.Sequence != nil && *record.Sequence > maximum {
			maximum = *record.Sequence
		}
	}
	return maximum + model.FirewallRuleSequenceStep, nil
}

func (s *FirewallService) sequenceForCreatedFirewallRule(
	ctx context.Context,
	snapshot filter.Snapshot,
	rule filter.FirewallRule,
) (int64, error) {
	if rule.OrderIndex == nil {
		return s.nextFirewallRuleSequence(ctx)
	}
	return s.sequenceForFirewallRulePosition(ctx, snapshot, int(*rule.OrderIndex), "", nil)
}

func (s *FirewallService) sequenceForFirewallRulePosition(
	ctx context.Context,
	snapshot filter.Snapshot,
	targetPosition int,
	excludedUUID string,
	current *int64,
) (int64, error) {
	stored, err := s.rules.List(ctx)
	if err != nil {
		return 0, err
	}
	byUUID := make(map[string]model.FirewallRule, len(stored))
	for _, record := range stored {
		byUUID[record.UUID] = record
	}
	var previous, next *model.FirewallRule
	needsRebalance := false
	for _, observed := range snapshot.Rules {
		uuid := strings.TrimPrefix(observed.Marker, "1panel-rule:")
		if observed.Marker == uuid || uuid == excludedUUID || observed.Locator.Position == nil {
			continue
		}
		record, exists := byUUID[uuid]
		if !exists {
			continue
		}
		position := *observed.Locator.Position
		if position < targetPosition {
			copy := record
			previous = &copy
		} else if position > targetPosition || excludedUUID == "" {
			copy := record
			next = &copy
			break
		}
	}
	if previous != nil && previous.Sequence == nil || next != nil && next.Sequence == nil {
		needsRebalance = true
	}
	if !needsRebalance && current != nil &&
		(previous == nil || *previous.Sequence < *current) && (next == nil || *current < *next.Sequence) {
		return *current, nil
	}
	if !needsRebalance {
		switch {
		case previous == nil && next == nil:
			return model.FirewallRuleSequenceStep, nil
		case previous == nil:
			return *next.Sequence - model.FirewallRuleSequenceStep, nil
		case next == nil:
			return *previous.Sequence + model.FirewallRuleSequenceStep, nil
		case *next.Sequence-*previous.Sequence > 1:
			return *previous.Sequence + (*next.Sequence-*previous.Sequence)/2, nil
		default:
			needsRebalance = true
		}
	}
	if needsRebalance {
		return s.rebalanceFirewallRuleSequences(ctx, snapshot, targetPosition, excludedUUID, byUUID)
	}
	return 0, fmt.Errorf("%w: cannot allocate firewall rule sequence", filter.ErrRuleOperation)
}

func (s *FirewallService) rebalanceFirewallRuleSequences(
	ctx context.Context,
	snapshot filter.Snapshot,
	targetPosition int,
	excludedUUID string,
	byUUID map[string]model.FirewallRule,
) (int64, error) {
	targetSequence := int64(targetPosition) * model.FirewallRuleSequenceStep
	for _, observed := range snapshot.Rules {
		if observed.Locator.Position == nil {
			continue
		}
		uuid := strings.TrimPrefix(observed.Marker, "1panel-rule:")
		if observed.Marker == uuid || uuid == excludedUUID {
			continue
		}
		record, exists := byUUID[uuid]
		if !exists {
			continue
		}
		position := *observed.Locator.Position
		if excludedUUID == "" && position >= targetPosition {
			position++
		}
		sequence := int64(position) * model.FirewallRuleSequenceStep
		if record.Sequence != nil && *record.Sequence == sequence {
			continue
		}
		if err := s.rules.UpdateWithRevision(ctx, record.UUID, record.Revision, map[string]interface{}{
			"sequence": sequence,
		}); err != nil {
			return 0, err
		}
	}
	return targetSequence, nil
}

func firewallRuleMarkerPosition(snapshot filter.Snapshot, ruleUUID string) (int, error) {
	marker := "1panel-rule:" + ruleUUID
	for _, observed := range snapshot.Rules {
		if observed.Marker == marker && observed.Locator.Position != nil {
			return *observed.Locator.Position, nil
		}
	}
	return 0, fmt.Errorf("%w: committed firewall rule %q has no position", filter.ErrVerificationFailed, ruleUUID)
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
	rules, err := firewallPolicyRulesForProvider(stored, filter.ProviderIptables)
	if err != nil {
		return filter.DesiredRule{}, err
	}
	if len(rules) != 1 {
		return filter.DesiredRule{}, fmt.Errorf(
			"%w: policy %q expands to %d canonical rules", filter.ErrUnsupportedScope, stored.UUID, len(rules),
		)
	}
	rule := rules[0]
	rule.UUID = stored.UUID
	ruleKey, err := filter.RuleKey(rule)
	if err != nil {
		return filter.DesiredRule{}, err
	}
	return filter.DesiredRule{
		UUID: stored.UUID, Rule: rule, RuleKey: ruleKey, Origin: filter.RuleOrigin(stored.Origin),
		Marker: "1panel-rule:" + stored.UUID,
	}, nil
}

func (s *FirewallService) compileStoredFirewallRules(
	ctx context.Context,
	stored model.FirewallRule,
	target filter.Provider,
) ([]filter.DesiredRule, error) {
	rules, err := firewallPolicyRulesForProvider(stored, target)
	if err != nil {
		return nil, err
	}
	runtime, err := s.adapters.Resolve(target)
	if err != nil {
		return nil, err
	}
	result := make([]filter.DesiredRule, 0, len(rules))
	scopeOrdinals := make(map[string]int)
	for _, rule := range rules {
		rule, err = runtime.Prepare(rule)
		if err != nil {
			return nil, err
		}
		if err = runtime.CheckRule(ctx, rule); err != nil {
			return nil, err
		}
		ruleKey, keyErr := filter.RuleKey(rule)
		if keyErr != nil {
			return nil, keyErr
		}
		scopeKey := rule.Scope.Key()
		ordinal := scopeOrdinals[scopeKey]
		scopeOrdinals[scopeKey] = ordinal + 1
		rule.UUID = compiledFirewallRuleUUID(stored.UUID, ruleKey, ordinal)
		result = append(result, filter.DesiredRule{
			UUID: stored.UUID, Rule: rule, RuleKey: ruleKey, Origin: filter.RuleOrigin(stored.Origin),
			Marker: "1panel-rule:" + rule.UUID,
		})
	}
	return result, nil
}

func compiledFirewallRuleUUID(policyUUID, ruleKey string, scopeOrdinal int) string {
	if scopeOrdinal == 0 {
		return policyUUID
	}
	const suffixLength = 12
	if len(ruleKey) > suffixLength {
		ruleKey = ruleKey[:suffixLength]
	}
	return fmt.Sprintf("%s-%d-%s", policyUUID, scopeOrdinal+1, ruleKey)
}

func (s *FirewallService) desiredFirewallRulesForScope(
	ctx context.Context,
	stored []model.FirewallRule,
	scope filter.Scope,
) ([]filter.DesiredRule, error) {
	sortFirewallPolicies(stored, scope.Provider)
	desired := make([]filter.DesiredRule, 0, len(stored))
	for _, record := range stored {
		compiled, err := s.compileStoredFirewallRules(ctx, record, scope.Provider)
		if err != nil {
			return nil, err
		}
		for _, rule := range compiled {
			if rule.Rule.Scope.Key() == scope.Key() {
				desired = append(desired, rule)
			}
		}
	}
	return desired, nil
}

func firewallRuleSnapshotPolicy(_ context.Context, snapshot filter.Snapshot) (filter.Snapshot, error) {
	ports, err := loadFirewallPortWhiteList()
	if err != nil {
		return filter.Snapshot{}, err
	}
	return filter.ProtectSnapshot(snapshot, ports)
}

func firewallRuleSelectedProvider(context.Context) (filter.Provider, error) {
	return selectedRuleProvider()
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

type firewallSnapshotPolicy func(context.Context, filter.Snapshot) (filter.Snapshot, error)

type firewallRuleRuntime struct {
	adapter filter.Adapter
	policy  firewallSnapshotPolicy
}

type firewallRuleRuntimeRegistry map[filter.Provider]*firewallRuleRuntime

func newFirewallRuleRuntimeRegistry(policy firewallSnapshotPolicy) firewallRuleRuntimeRegistry {
	return firewallRuleRuntimeRegistry{
		filter.ProviderIptables:  newFirewallRuleRuntime(filteriptables.NewAdapter(), policy),
		filter.ProviderNftables:  newFirewallRuleRuntime(filternftables.NewAdapter(), policy),
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

func (r *firewallRuleRuntime) ObserveScopes(ctx context.Context, scopes []filter.Scope) ([]filter.Snapshot, error) {
	observer, ok := r.adapter.(filter.MultiScopeObserver)
	if !ok {
		return nil, fmt.Errorf("%w: %s multi-scope inventory", filter.ErrAdapterUnavailable, r.adapter.Provider())
	}
	snapshots, err := observer.ObserveScopes(ctx, scopes)
	if err != nil {
		return nil, err
	}
	if r.policy == nil {
		return snapshots, nil
	}
	for index := range snapshots {
		snapshots[index], err = r.policy(ctx, snapshots[index])
		if err != nil {
			return nil, err
		}
	}
	return snapshots, nil
}

func (r *firewallRuleRuntime) ObserveMutation(ctx context.Context, scope filter.Scope) (filter.Snapshot, error) {
	snapshot, err := r.Observe(ctx, scope)
	if err != nil {
		return filter.Snapshot{}, err
	}
	for _, notice := range snapshot.Notices {
		if notice.Code == filter.ScopeNoticeManagedScopeInactive || notice.Code == filter.ScopeNoticeManagedScopeMissing {
			return filter.Snapshot{}, fmt.Errorf("%w: managed firewall scope is unavailable", filter.ErrProviderUnavailable)
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
	provider, err := selectedSystemFirewallProvider()
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

func refreshCreateAuthorization(
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

func OperateFirewallPort(oldPorts, newPorts []int) error {
	client, err := selectedSystemFirewallClient()
	if err != nil {
		return err
	}
	state, err := lifecycle.LoadState(client)
	if err != nil {
		return err
	}
	if state.Name == constant.FirewallProviderIptables || state.Name == constant.FirewallProviderNftables {
		isInit, _, err := loadDirectFirewallInitStatus(state.Name)
		if err != nil {
			return err
		}
		if !isInit {
			return nil
		}
		if state.Name == constant.FirewallProviderIptables {
			if err := newIptablesHelperManager().SyncRequiredPorts(true); err != nil {
				return err
			}
		} else if err := newNftablesHelperManager().SyncRequiredPorts(); err != nil {
			return err
		}
	} else if !state.IsActive {
		return nil
	}
	current, err := loadFirewallPortWhiteList()
	if err != nil {
		return err
	}
	previous := make([]firewall.PortWhitelist, 0, len(oldPorts))
	for _, port := range oldPorts {
		item := firewall.PortWhitelist{Port: strconv.Itoa(port), Protocol: "tcp"}
		if !containsFirewallPort(current, item) {
			previous = append(previous, item)
		}
	}
	added := make([]firewall.PortWhitelist, 0, len(newPorts))
	for _, port := range newPorts {
		added = append(added, firewall.PortWhitelist{Port: strconv.Itoa(port), Protocol: "tcp"})
	}
	if state.Name == constant.FirewallProviderIptables || state.Name == constant.FirewallProviderNftables {
		required, err := loadRequiredFirewallPortWhiteList()
		if err != nil {
			return err
		}
		added = excludeFirewallPorts(added, required)
	}
	return syncManagedAcceptedPorts(previous, added)
}

func containsFirewallPort(ports []firewall.PortWhitelist, target firewall.PortWhitelist) bool {
	for _, item := range ports {
		familyMatches := item.Family == "" || target.Family == "" || item.Family == target.Family
		if familyMatches && item.Port == target.Port && item.Protocol == target.Protocol {
			return true
		}
	}
	return false
}

func LoadPanelPort() string {
	if !global.IsMaster {
		return global.CONF.Base.Port
	}
	var portSetting model.Setting
	_ = global.CoreDB.Where("key = ?", "ServerPort").First(&portSetting).Error
	return portSetting.Value
}

func (s *FirewallService) SyncSystemPorts(ctx context.Context, previous, current []dto.FirewallSystemPort) error {
	previousSet, err := normalizeSystemPorts(previous)
	if err != nil {
		return err
	}
	currentSet, err := normalizeSystemPorts(current)
	if err != nil {
		return err
	}

	provider, err := s.selectedProvider(ctx)
	if err != nil {
		return err
	}
	if !supportsNativeRuleBatch(provider) {
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

	creates := make([]dto.FirewallRuleCreateItem, 0)
	for _, key := range sortedSystemPortKeys(currentSet) {
		if _, exists := previousSet[key]; exists {
			continue
		}
		create, prepareErr := s.prepareSystemPortCreate(ctx, currentSet[key])
		if prepareErr != nil {
			return prepareErr
		}
		if create != nil {
			creates = append(creates, *create)
		}
	}
	if len(creates) > 0 {
		result, batchErr := s.Create(ctx, dto.FirewallRuleCreate{Items: creates})
		if batchErr != nil {
			return batchErr
		}
		if result.Failed > 0 || result.Skipped > 0 {
			if len(result.Errors) > 0 {
				return fmt.Errorf("batch create accepted firewall ports: %s", result.Errors[0].Error)
			}
			return fmt.Errorf("batch create accepted firewall ports failed")
		}
	}

	removed := make([]dto.FirewallSystemPort, 0)
	deleteUUIDs := make([]string, 0)
	for _, key := range sortedSystemPortKeys(previousSet) {
		if _, exists := currentSet[key]; exists {
			continue
		}
		port := previousSet[key]
		removed = append(removed, port)
		stored, listErr := s.systemPortRecords(ctx, port)
		if listErr != nil {
			return listErr
		}
		for _, rule := range stored {
			deleteUUIDs = append(deleteUUIDs, rule.UUID)
		}
	}
	if len(deleteUUIDs) > 0 {
		result, batchErr := s.Delete(ctx, dto.FirewallRuleDelete{UUIDs: deleteUUIDs})
		if batchErr != nil {
			return batchErr
		}
		if result.Failed > 0 {
			if len(result.Errors) > 0 {
				return fmt.Errorf("batch delete accepted firewall ports: %s", result.Errors[0].Error)
			}
			return fmt.Errorf("batch delete accepted firewall ports failed")
		}
	}
	for _, port := range removed {
		if err := s.deleteSystemPort(ctx, port); err != nil {
			return err
		}
	}
	return nil
}

func loadConfiguredFirewallPortWhiteList() ([]firewall.PortWhitelist, error) {
	value, err := settingRepo.GetValueByKey(constant.FirewallPortWhiteList)
	if err != nil {
		value = constant.FirewallPortWhiteListValue
		if err := settingRepo.UpdateOrCreate(constant.FirewallPortWhiteList, value); err != nil {
			return nil, err
		}
	}
	return firewall.ParsePortWhitelist(value)
}

func loadFirewallPortWhiteList() ([]firewall.PortWhitelist, error) {
	configured, err := loadConfiguredFirewallPortWhiteList()
	if err != nil {
		return nil, err
	}
	required, err := loadRequiredFirewallPortWhiteList()
	if err != nil {
		return nil, err
	}
	return firewall.NormalizePortWhitelist(append(configured, required...)), nil
}

func loadRequiredFirewallPortWhiteList() ([]firewall.PortWhitelist, error) {
	panelPort := LoadPanelPort()
	if panelPort == "" {
		return nil, fmt.Errorf("find 1panel service port failed")
	}
	return firewall.NormalizePortWhitelist([]firewall.PortWhitelist{
		{Port: panelPort, Protocol: "tcp"},
		{Port: loadSSHPort(), Protocol: "tcp"},
	}), nil
}

func SyncFirewallPortWhitelistAfterUpdate(oldValue string) error {
	client, err := selectedSystemFirewallClient()
	if err != nil {
		return err
	}
	state, err := lifecycle.LoadState(client)
	if err != nil {
		return err
	}
	if state.Name == constant.FirewallProviderIptables || state.Name == constant.FirewallProviderNftables {
		isInit, _, err := loadDirectFirewallInitStatus(state.Name)
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
	oldPorts, err := firewall.ParsePortWhitelist(oldValue)
	if err != nil {
		return err
	}
	required, err := loadRequiredFirewallPortWhiteList()
	if err != nil {
		return err
	}
	if state.Name == constant.FirewallProviderIptables || state.Name == constant.FirewallProviderNftables {
		ports = excludeFirewallPorts(ports, required)
		oldPorts = excludeFirewallPorts(oldPorts, required)
	} else {
		ports = firewall.NormalizePortWhitelist(append(ports, required...))
		oldPorts = firewall.NormalizePortWhitelist(append(oldPorts, required...))
	}
	return syncManagedAcceptedPorts(oldPorts, ports)
}

func newIptablesHelperManager() *iptables_helper.Manager {
	return &iptables_helper.Manager{
		UpdateSetting:     settingRepo.Update,
		PanelPort:         LoadPanelPort,
		LoadRequiredPorts: loadRequiredFirewallPortWhiteList,
	}
}

func newNftablesHelperManager() *nftables_helper.Manager {
	return &nftables_helper.Manager{
		UpdateSetting:     settingRepo.Update,
		LoadRequiredPorts: loadRequiredFirewallPortWhiteList,
	}
}

func loadDirectFirewallInitStatus(provider string) (bool, bool, error) {
	return loadFirewallInitStatus(provider, "base")
}

func loadFirewallInitStatus(provider, tab string) (bool, bool, error) {
	switch provider {
	case constant.FirewallProviderNftables:
		return nftables_helper.LoadInitStatus(tab)
	case constant.FirewallProviderIptables:
		return iptables_helper.LoadInitStatus(tab)
	default:
		return false, false, fmt.Errorf("unsupported firewall provider: %s", provider)
	}
}

func supportsManagedFilterChains(provider string) bool {
	return provider == constant.FirewallProviderIptables || provider == constant.FirewallProviderNftables
}

func (s *FirewallService) addPortsBeforeStart(client lifecycle.Client) error {
	if client.Name() == constant.FirewallProviderIptables || client.Name() == constant.FirewallProviderNftables {
		isInit, _, err := loadDirectFirewallInitStatus(client.Name())
		if err != nil {
			return err
		}
		if !isInit {
			return nil
		}
		if client.Name() == constant.FirewallProviderIptables {
			if err := newIptablesHelperManager().SyncRequiredPorts(true); err != nil {
				return err
			}
		} else if err := newNftablesHelperManager().SyncRequiredPorts(); err != nil {
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
		return s.SyncSystemPorts(context.Background(), nil, systemPorts(excludeFirewallPorts(configured, required)))
	}
	portWhitelist, err := loadFirewallPortWhiteList()
	if err != nil {
		return err
	}
	return s.SyncSystemPorts(context.Background(), nil, systemPorts(portWhitelist))
}

func syncManagedAcceptedPorts(previous, current []firewall.PortWhitelist) error {
	return newFirewallService().
		SyncSystemPorts(context.Background(), systemPorts(previous), systemPorts(current))
}

func systemPorts(ports []firewall.PortWhitelist) []dto.FirewallSystemPort {
	result := make([]dto.FirewallSystemPort, 0, len(ports))
	for _, port := range ports {
		result = append(result, dto.FirewallSystemPort{Family: port.Family, Port: port.Port, Protocol: port.Protocol})
	}
	return result
}

func excludeFirewallPorts(ports, excluded []firewall.PortWhitelist) []firewall.PortWhitelist {
	result := make([]firewall.PortWhitelist, 0, len(ports))
	for _, port := range ports {
		exists := false
		for _, item := range excluded {
			familyMatches := item.Family == "" || port.Family == "" || item.Family == port.Family
			if familyMatches && item.Port == port.Port && item.Protocol == port.Protocol {
				exists = true
				break
			}
		}
		if !exists {
			result = append(result, port)
		}
	}
	return result
}

const (
	sshConfigPath          = "/etc/ssh/sshd_config"
	defaultFirewallSSHPort = "22"
)

func loadSSHPort() string {
	content, err := os.ReadFile(sshConfigPath)
	if err != nil {
		return defaultFirewallSSHPort
	}
	for _, line := range strings.Split(string(content), "\n") {
		if !strings.HasPrefix(line, "Port ") {
			continue
		}
		port := strings.TrimSpace(strings.TrimPrefix(line, "Port "))
		value, _ := strconv.Atoi(port)
		if value > 0 && value < 65535 {
			return port
		}
	}
	return defaultFirewallSSHPort
}
