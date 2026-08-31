package service

import (
	"context"
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
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/controller"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	filterruntime "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/runtime"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/nftables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/ping"
	"gorm.io/gorm"
)

type FirewallService struct {
	rules                  repo.IFirewallRuleRepo
	adapters               firewallRuleRuntimeResolver
	forwardingSync         firewallDatabaseSyncAdapter
	dockerSync             firewallDatabaseSyncAdapter
	selectedProvider       func(context.Context) (filter.Provider, error)
	protectedPorts         func() ([]firewall.PortWhitelist, error)
	iptablesHelper         *iptables_helper.Manager
	cleanupBackend         func(string) error
	cleanupInactiveBackend func(string) error
	resetBackend           func(string, bool) error
	dockerActive           func() (bool, error)
	restoreForwarding      func(context.Context) error
	restoreDockerGuard     func(context.Context) error
	baseClient             func() (lifecycle.Client, error)
}

type firewallRuleRuntimeResolver interface {
	Resolve(filter.Provider) (*firewallRuleRuntime, error)
	Providers() []filter.Provider
}

var firewallRuleMutationMu sync.Mutex

type IFirewallService interface {
	LoadBaseInfo(chainGroup string) (dto.FirewallSubsystemStatus, error)
	OperateFirewall(request dto.FirewallLifecycleOperation) error
	OperateFilterChain(request dto.FilterChainOperation) error
	QueueFilterChainInitialization(request dto.FilterChainOperation) (dto.FilterChainOperationResponse, error)
	Reset(context.Context, dto.FirewallRuleReset) (dto.FirewallRuleResetResponse, error)
	Inventory(context.Context, dto.FirewallRuleInventory) (dto.FirewallRuleInventoryResponse, error)
	LoadFirewallNativeDetail(context.Context, dto.FirewallNativeDetail) (string, error)
	Check(context.Context, string, dto.FirewallRuleCheck) (dto.FirewallRuleCheckResponse, error)
	Create(context.Context, dto.FirewallRuleCreate) (dto.FirewallRuleCreateResponse, error)
	Delete(context.Context, dto.FirewallRuleDelete) (dto.FirewallRuleDeleteResponse, error)
	Update(context.Context, string, dto.FirewallRuleUpdate) error
	Reorder(context.Context, string, dto.FirewallRuleReorder) error
	PreviewRuleSync(context.Context, string, dto.FirewallRuleSyncRequest) (dto.FirewallRuleSyncPreview, error)
	SyncRules(context.Context, string, dto.FirewallRuleSyncRequest) (dto.FirewallRuleSyncResult, error)
	CurrentRuleSyncTask() (dto.FirewallRuleSyncTask, error)
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
		dockerActive: func() (bool, error) {
			return controller.CheckActive("docker")
		},
		restoreForwarding: func(ctx context.Context) error {
			return newForwardingService().Restore(ctx)
		},
		restoreDockerGuard: ReconcileDockerPortGuard,
		baseClient:         selectedSystemFirewallClient,
	}
}

func (s *FirewallService) LoadBaseInfo(chainGroup string) (dto.FirewallSubsystemStatus, error) {
	status := dto.FirewallSubsystemStatus{Version: "-", Name: "-", Backend: "-"}
	if selected := configuredSystemFirewallBackend(); selected != "" {
		status.Name, status.Backend = selected, selected
	}
	loadClient := s.baseClient
	if loadClient == nil {
		loadClient = selectedSystemFirewallClient
	}
	client, err := loadClient()
	if err != nil {
		if global.LOG != nil {
			global.LOG.Errorf("load firewall failed, err: %v", err)
		}
		if errors.Is(err, lifecycle.ErrNotInstalled) {
			status.Reason = constant.FirewallBackendNotInstalled
			return status, nil
		}
		status.IsExist = true
		status.Message = err.Error()
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
	}
	return status, nil
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
	if err := s.operateFilterChainBase(provider, request); err != nil {
		return err
	}
	if request.Operate != string(firewall.BaseOperationInit) && request.Operate != string(firewall.BaseOperationBind) {
		return nil
	}
	ctx := context.Background()
	if err := s.restoreStoredFirewallRules(ctx, filter.Provider(provider)); err != nil {
		return err
	}
	return s.syncConfiguredFirewallPorts(ctx)
}

func (s *FirewallService) QueueFilterChainInitialization(
	request dto.FilterChainOperation,
) (dto.FilterChainOperationResponse, error) {
	if request.Operate != string(firewall.BaseOperationInit) {
		return dto.FilterChainOperationResponse{}, fmt.Errorf("only filter chain initialization can be queued")
	}
	provider, err := selectedSystemFirewallProvider()
	if err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	if !supportsManagedFilterChains(provider) {
		return dto.FilterChainOperationResponse{}, fmt.Errorf("filter chain operations are not supported for %s", provider)
	}
	if err := task.CheckScopeTaskIsExecuting(task.TaskScopeFirewall, 0); err != nil {
		return dto.FilterChainOperationResponse{}, err
	}

	resourceName := fmt.Sprintf("%s filter", provider)
	taskItem, err := task.NewTaskWithOps(resourceName, task.TaskExec, task.TaskScopeFirewall, request.TaskID, 0)
	if err != nil {
		return dto.FilterChainOperationResponse{}, fmt.Errorf("create firewall initialization task: %w", err)
	}
	taskItem.AddSubTask(i18n.GetWithName("FirewallInitializeChainsStep", provider), func(t *task.Task) error {
		t.Logf("backend=%s", provider)
		return s.operateFilterChainBase(provider, request)
	}, nil)
	taskItem.AddSubTask(i18n.GetWithName("FirewallRestoreRulesStep", provider), func(t *task.Task) error {
		return s.restoreStoredFirewallRules(t.TaskCtx, filter.Provider(provider))
	}, nil)
	taskItem.AddSubTask(i18n.GetMsgByKey("FirewallSyncWhitelistStep"), func(t *task.Task) error {
		return s.syncConfiguredFirewallPorts(t.TaskCtx)
	}, nil)
	if err := repo.NewITaskRepo().Save(context.Background(), taskItem.Task); err != nil {
		return dto.FilterChainOperationResponse{}, fmt.Errorf("save firewall initialization task: %w", err)
	}
	go func() {
		_ = taskItem.Execute()
	}()
	return dto.FilterChainOperationResponse{TaskID: taskItem.TaskID, Queued: true}, nil
}

func (s *FirewallService) operateFilterChainBase(provider string, request dto.FilterChainOperation) error {
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
	return nil
}

func (s *FirewallService) syncConfiguredFirewallPorts(ctx context.Context) error {
	configured, err := loadConfiguredFirewallPortWhiteList()
	if err != nil {
		return err
	}
	required, err := loadRequiredFirewallPortWhiteList()
	if err != nil {
		return err
	}
	ports := excludeFirewallPorts(configured, required)
	return s.SyncSystemPorts(ctx, nil, systemPorts(ports))
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
	restartDocker := false
	if provider == filter.ProviderFirewalld && request.WithDockerRestart {
		dockerActive := s.dockerActive
		if dockerActive == nil {
			dockerActive = func() (bool, error) { return controller.CheckActive("docker") }
		}
		active, err := dockerActive()
		if err != nil {
			return dto.FirewallRuleResetResponse{}, fmt.Errorf("check Docker status before resetting firewalld: %w", err)
		}
		restartDocker = active
	}
	resetErr := reset(string(provider), restartDocker)
	if resetErr != nil {
		var dockerRestartErr *lifecycle.DockerRestartError
		if provider != filter.ProviderFirewalld || !errors.As(resetErr, &dockerRestartErr) {
			return dto.FirewallRuleResetResponse{}, resetErr
		}
	}
	if provider == filter.ProviderFirewalld {
		restoreForwarding := s.restoreForwarding
		if restoreForwarding == nil {
			restoreForwarding = func(ctx context.Context) error { return newForwardingService().Restore(ctx) }
		}
		restoreDockerGuard := s.restoreDockerGuard
		if restoreDockerGuard == nil {
			restoreDockerGuard = ReconcileDockerPortGuard
		}
		restoreErr := restoreFirewalldDependents(ctx, restartDocker, restoreForwarding, restoreDockerGuard)
		if err := errors.Join(resetErr, restoreErr); err != nil {
			return dto.FirewallRuleResetResponse{}, err
		}
	}
	return dto.FirewallRuleResetResponse{Removed: len(stored), Disabled: true}, nil
}

func isDirectFirewallProvider(provider filter.Provider) bool {
	return provider == filter.ProviderIptables || provider == filter.ProviderNftables
}

func resetServiceFirewallBackend(provider string, withDockerRestart bool) error {
	client, err := lifecycle.NewClientFor(provider)
	if err != nil {
		return err
	}
	return resetServiceFirewallClient(client, withDockerRestart, func(
		client lifecycle.Client,
		restartDocker bool,
		prepareStop func() error,
	) error {
		return lifecycle.NewOperator(client).StopWithPrepare(restartDocker, prepareStop)
	})
}

func resetServiceFirewallClient(
	client lifecycle.Client,
	withDockerRestart bool,
	stop func(lifecycle.Client, bool, func() error) error,
) error {
	resetter, ok := client.(lifecycle.Resetter)
	if !ok {
		return fmt.Errorf("firewall provider %s does not support reset", client.Name())
	}
	if resetBeforeStop, ok := client.(lifecycle.PreStopResetter); ok {
		if err := stop(client, withDockerRestart, resetBeforeStop.ResetBeforeStop); err != nil {
			return err
		}
		return nil
	}
	return resetter.Reset()
}

func restoreFirewalldDependents(
	ctx context.Context,
	restartDocker bool,
	restoreForwarding func(context.Context) error,
	restoreDockerGuard func(context.Context) error,
) error {
	var errs []error
	if err := restoreForwarding(ctx); err != nil {
		errs = append(errs, fmt.Errorf("restore port forwarding after resetting firewalld: %w", err))
	}
	if restartDocker {
		if err := restoreDockerGuard(ctx); err != nil {
			errs = append(errs, fmt.Errorf("restore Docker port guard after resetting firewalld: %w", err))
		}
	}
	return errors.Join(errs...)
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
	requestedScopes := request.Scopes
	if len(requestedScopes) == 0 && request.Scope.Provider != "" {
		requestedScopes = []filter.Scope{request.Scope}
	}
	if len(requestedScopes) == 0 {
		return dto.FirewallRuleInventoryResponse{}, filter.ErrInvalidScope
	}
	scopes := make([]filter.Scope, len(requestedScopes))
	for index, requested := range requestedScopes {
		scopes[index] = requested.Normalize()
	}
	if len(scopes) == 1 && isCombinedUFWInventoryScope(scopes[0]) {
		scope := scopes[0]
		if err := s.checkSelectedProvider(ctx, scope.Provider); err != nil {
			return dto.FirewallRuleInventoryResponse{}, err
		}
		runtime, err := s.adapters.Resolve(scope.Provider)
		if err != nil {
			return dto.FirewallRuleInventoryResponse{}, err
		}
		response, err := s.combinedUFWInventory(ctx, runtime, scope)
		if err != nil {
			return dto.FirewallRuleInventoryResponse{}, err
		}
		return finalizeFirewallInventory(response, request), nil
	}
	provider := scopes[0].Provider
	for _, scope := range scopes {
		if err := scope.ValidateMVP(); err != nil {
			return dto.FirewallRuleInventoryResponse{}, err
		}
		if scope.Provider != provider {
			return dto.FirewallRuleInventoryResponse{}, fmt.Errorf(
				"%w: inventory scopes must use the same provider", filter.ErrInvalidScope,
			)
		}
	}
	if err := s.checkSelectedProvider(ctx, provider); err != nil {
		return dto.FirewallRuleInventoryResponse{}, err
	}
	runtime, err := s.adapters.Resolve(provider)
	if err != nil {
		return dto.FirewallRuleInventoryResponse{}, err
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		return dto.FirewallRuleInventoryResponse{}, err
	}
	response := dto.FirewallRuleInventoryResponse{}
	for _, scope := range scopes {
		snapshot, err := runtime.Observe(ctx, scope)
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
		response.Items = append(response.Items, items...)
		response.Notices = append(response.Notices, snapshot.Notices...)
	}
	return finalizeFirewallInventory(response, request), nil
}

func finalizeFirewallInventory(
	response dto.FirewallRuleInventoryResponse,
	request dto.FirewallRuleInventory,
) dto.FirewallRuleInventoryResponse {
	response.AllTotal = int64(len(response.Items))
	for _, item := range response.Items {
		if isDeletableManagedInventoryItem(item) {
			response.ManagedTotal++
		}
	}
	filtered := make([]filter.InventoryItem, 0, len(response.Items))
	for _, item := range response.Items {
		if matchesFirewallInventoryRequest(item, request) {
			filtered = append(filtered, item)
		}
	}
	response.Total = int64(len(filtered))
	if request.All {
		response.Items = filtered
		return response
	}
	page, pageSize := max(1, request.Page), max(1, request.PageSize)
	start := (page - 1) * pageSize
	if start >= len(filtered) {
		response.Items = make([]filter.InventoryItem, 0)
		return response
	}
	end := min(start+pageSize, len(filtered))
	response.Items = filtered[start:end]
	return response
}

func matchesFirewallInventoryRequest(item filter.InventoryItem, request dto.FirewallRuleInventory) bool {
	if slicesContains(request.ExcludeChains, item.Rule.Scope.Chain) {
		return false
	}
	if len(request.Families) > 0 && !matchesFirewallInventoryFamily(item.Rule, request.Families) {
		return false
	}
	if len(request.Actions) > 0 && !matchesFirewallInventoryAction(item.Rule.Action, request.Actions) {
		return false
	}
	if len(request.States) > 0 && !slicesContains(request.States, item.State) {
		return false
	}
	keyword := strings.ToLower(strings.TrimSpace(request.Info))
	if keyword == "" {
		return true
	}
	rule := item.Rule
	values := []string{
		firewallInventoryProtocol(rule), rule.SourceAddress, rule.SourcePort, rule.DestinationAddress,
		rule.DestinationPort, rule.Description, string(rule.Action), string(item.State),
	}
	if item.Observed != nil {
		values = append(values, item.Observed.Rule.Description)
	}
	if item.Desired != nil {
		values = append(values, item.Desired.Rule.Description)
	}
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), keyword) {
			return true
		}
	}
	return false
}

func matchesFirewallInventoryFamily(rule filter.FirewallRule, families []filter.Family) bool {
	for _, family := range families {
		if rule.Scope.Family != filter.FamilyInet && rule.Scope.Family == family {
			return true
		}
		if rule.Scope.Family == filter.FamilyInet &&
			(rule.SourceAddress == "" || (family == filter.FamilyIPv6) == strings.Contains(rule.SourceAddress, ":")) {
			return true
		}
	}
	return false
}

func matchesFirewallInventoryAction(action filter.Action, actions []string) bool {
	for _, requested := range actions {
		if requested == "accept" && action == filter.ActionAccept {
			return true
		}
		if requested == "deny" && action != filter.ActionAccept {
			return true
		}
	}
	return false
}

func firewallInventoryProtocol(rule filter.FirewallRule) string {
	if rule.NativeKind == filter.NativeKindZoneService {
		return "service"
	}
	if rule.NativeKind == filter.NativeKindUFWApplication && rule.Protocol == "" {
		return "app"
	}
	if rule.Scope.Provider == filter.ProviderUFW && rule.Protocol == "all" && rule.DestinationPort != "" {
		return "tcp/udp"
	}
	return rule.Protocol
}

func slicesContains[T comparable](values []T, target T) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func isDeletableManagedInventoryItem(item filter.InventoryItem) bool {
	if item.Desired == nil || item.Desired.Protected || item.State == filter.InventoryStateProtected {
		return false
	}
	if item.Desired.Origin != filter.RuleOriginCreated && item.Desired.Origin != filter.RuleOriginAdopted {
		return false
	}
	if isIptablesSystemPresetInventoryScope(item.Rule.Scope) {
		return false
	}
	return item.State != filter.InventoryStateDrifted ||
		(item.Match == filter.InventoryMatchMissing && item.Observed == nil)
}

func isIptablesSystemPresetInventoryScope(scope filter.Scope) bool {
	return (scope.Provider == filter.ProviderIptables || scope.Provider == filter.ProviderNftables) &&
		(scope.Chain == filter.BasicBeforeChain || scope.Chain == filter.BasicAfterChain)
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
	return runtime.NativeDetail(ctx, request.Name, request.Permanent)
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
			managedRevision, revisionErr := model.FirewallRulesRevision(stored)
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
			managedRevision, revisionErr := model.FirewallRulesRevision(stored)
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
	if first.runtime == nil || !supportsNativeRuleBatch(first.runtime.Provider()) ||
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
			if err := s.deleteRule(ctx, item.stored.UUID, false); err != nil {
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
	if isProtectedSystemFirewallRule(stored) {
		return preparedFirewallRuleDelete{}, filter.ErrProtectedRule
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
			if errors.Is(observeErr, filter.ErrRuleStale) {
				missing, mergeErr := managedFirewallRuleMissing(snapshot, item.desired)
				if mergeErr != nil {
					return mergeErr
				}
				if missing {
					continue
				}
			}
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
	var backendPlan filter.BackendPlan
	if len(changes) > 0 {
		var verification filter.VerifyResult
		backendPlan, verification, err = runtime.Execute(ctx, snapshot, changes)
		if err != nil {
			return err
		}
		if !verification.Matched {
			return filter.ErrVerificationFailed
		}
	}

	deleted := make([]model.FirewallRule, 0, len(prepared))
	for _, item := range prepared {
		if err = s.rules.DeleteWithRevision(ctx, item.stored.UUID, item.stored.Revision); err != nil {
			if len(changes) > 0 {
				err = rollbackFirewallPlan(ctx, runtime, backendPlan, err)
			}
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
		appendPosition, err := runtime.AppendPosition(ctx, snapshot, domainRule)
		if err != nil {
			return err
		}
		domainRule.OrderIndex = &appendPosition
		appendRule = true
	} else if authorization.Operation == filter.ChangeCreate && domainRule.OrderIndex != nil {
		maxPosition, err := runtime.MaxPosition(ctx, snapshot, domainRule)
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

func (s *FirewallService) deleteRule(ctx context.Context, ruleUUID string, allowProtected bool) error {
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
		if allowProtected {
			desired.Protected = false
		}
		runtime, runtimeErr := s.resolveRuntime(ctx, desired.Rule.Scope.Provider)
		if runtimeErr != nil {
			return rollback(runtimeErr)
		}
		snapshot, observeErr := runtime.ObserveMutation(ctx, desired.Rule.Scope)
		if observeErr != nil {
			return rollback(observeErr)
		}
		if allowProtected {
			for index := range snapshot.Rules {
				snapshot.Rules[index].Protected = false
			}
		}
		observed, managedErr := filter.ManagedObserved(snapshot, desired)
		if managedErr != nil {
			if errors.Is(managedErr, filter.ErrRuleStale) {
				missing, mergeErr := managedFirewallRuleMissing(snapshot, desired)
				if mergeErr != nil {
					return rollback(mergeErr)
				}
				if missing {
					continue
				}
			}
			return rollback(managedErr)
		}
		restoreAtEnd := false
		if desired.Rule.Scope.Provider == filter.ProviderUFW && observed.Locator.Position != nil {
			maxPosition := maxObservedFirewallPosition(snapshot)
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

func managedFirewallRuleMissing(snapshot filter.Snapshot, desired filter.DesiredRule) (bool, error) {
	items, err := filter.MergeInventory(filter.InventoryMergeInput{
		Observed: snapshot.Rules,
		Desired:  []filter.DesiredRule{desired},
	})
	if err != nil {
		return false, err
	}
	for _, item := range items {
		if item.Desired != nil && item.Desired.UUID == desired.UUID {
			return item.Match == filter.InventoryMatchMissing && item.Observed == nil, nil
		}
	}
	return false, nil
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
		if err := runtime.ValidatePosition(ctx, snapshot, before.Rule, *targetPosition); err != nil {
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
	if !supportsManagedNativeKindTransition(before.Rule, after) {
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
			if err := runtime.ValidatePosition(ctx, snapshot, before.Rule, *after.OrderIndex); err != nil {
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

func supportsManagedNativeKindTransition(before, after filter.FirewallRule) bool {
	if before.NativeKind == after.NativeKind {
		return true
	}
	if before.Scope.Key() != after.Scope.Key() || before.Scope.Provider != filter.ProviderFirewalld {
		return false
	}
	return before.NativeKind == filter.NativeKindZonePort && after.NativeKind == filter.NativeKindRichRule ||
		before.NativeKind == filter.NativeKindRichRule && after.NativeKind == filter.NativeKindZonePort
}

func (s *FirewallService) loadManagedMutation(
	ctx context.Context,
	ruleUUID string,
) (model.FirewallRule, filter.DesiredRule, filter.Snapshot, filter.ObservedRule, *firewallRuleRuntime, error) {
	stored, err := s.rules.GetByUUID(ctx, ruleUUID)
	if err != nil {
		return model.FirewallRule{}, filter.DesiredRule{}, filter.Snapshot{}, filter.ObservedRule{}, nil, err
	}
	if isProtectedSystemFirewallRule(stored) {
		return model.FirewallRule{}, filter.DesiredRule{}, filter.Snapshot{}, filter.ObservedRule{}, nil,
			filter.ErrProtectedRule
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
	if s.adapters != nil {
		providers := s.adapters.Providers()
		if len(providers) == 1 {
			return providers[0], nil
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
		maxPosition := maxObservedFirewallPosition(request.Snapshot)
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

func maxObservedFirewallPosition(snapshot filter.Snapshot) int64 {
	var maximum int64
	for _, observed := range snapshot.Rules {
		if observed.Locator.Position != nil && int64(*observed.Locator.Position) > maximum {
			maximum = int64(*observed.Locator.Position)
		}
	}
	return maximum
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
	case filter.CheckClassificationConflict:
		return nil, fmt.Errorf("cannot manage accepted port %s/%s: %s", port.Port, port.Protocol, check.Reason)
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
			if err := s.deleteProtectedSystemPortRule(ctx, rule.UUID); err != nil {
				return err
			}
		}
	}
}

func (s *FirewallService) deleteProtectedSystemPortRule(ctx context.Context, ruleUUID string) error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	return s.deleteRule(ctx, ruleUUID, true)
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

func isProtectedSystemFirewallRule(rule model.FirewallRule) bool {
	ownerPrefix := model.FirewallRuleOwner(
		constant.FirewallRuleSourceSecurity,
		constant.FirewallSystemAcceptedPortSourcePrefix,
	)
	return strings.HasPrefix(rule.Owner, ownerPrefix)
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
	return firewall.RuleForSystemPort(provider, firewall.SystemPort(port))
}

func normalizeSystemPorts(ports []dto.FirewallSystemPort) (map[string]dto.FirewallSystemPort, error) {
	domainPorts := make([]firewall.SystemPort, 0, len(ports))
	for _, port := range ports {
		domainPorts = append(domainPorts, firewall.SystemPort(port))
	}
	normalized, err := firewall.NormalizeSystemPorts(domainPorts)
	if err != nil {
		return nil, err
	}
	result := make(map[string]dto.FirewallSystemPort, len(normalized))
	for key, port := range normalized {
		result[key] = dto.FirewallSystemPort(port)
	}
	return result, nil
}

func systemPortKey(port dto.FirewallSystemPort) string {
	return firewall.SystemPortKey(firewall.SystemPort(port))
}

func legacySystemPortKey(port dto.FirewallSystemPort) string {
	return firewall.LegacySystemPortKey(firewall.SystemPort(port))
}

func sortedSystemPortKeys(ports map[string]dto.FirewallSystemPort) []string {
	domainPorts := make(map[string]firewall.SystemPort, len(ports))
	for key, port := range ports {
		domainPorts[key] = firewall.SystemPort(port)
	}
	return firewall.SortedSystemPortKeys(domainPorts)
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
	rules, err := stored.RulesForProvider(filter.ProviderIptables)
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
	rules, err := stored.RulesForProvider(target)
	if err != nil {
		return nil, err
	}
	runtime, err := s.adapters.Resolve(target)
	if err != nil {
		return nil, err
	}
	return runtime.CompileDesired(ctx, stored.UUID, filter.RuleOrigin(stored.Origin), rules)
}

func (s *FirewallService) desiredFirewallRulesForScope(
	ctx context.Context,
	stored []model.FirewallRule,
	scope filter.Scope,
) ([]filter.DesiredRule, error) {
	model.SortFirewallRules(stored, scope.Provider)
	desired := make([]filter.DesiredRule, 0, len(stored))
	for _, record := range stored {
		compiled, err := s.compileStoredFirewallRules(ctx, record, scope.Provider)
		if err != nil {
			return nil, err
		}
		for _, rule := range compiled {
			if rule.Rule.Scope.Key() == scope.Key() {
				rule.Protected = isProtectedSystemFirewallRule(record)
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

type firewallSnapshotPolicy = filterruntime.SnapshotPolicy
type firewallRuleRuntime = filterruntime.Engine
type firewallRuleRuntimeRegistry = filterruntime.Registry

func newFirewallRuleRuntimeRegistry(policy firewallSnapshotPolicy) firewallRuleRuntimeRegistry {
	return filterruntime.NewRegistry(policy)
}

func newFirewallRuleRuntime(adapter filter.Adapter, policy firewallSnapshotPolicy) *firewallRuleRuntime {
	return filterruntime.New(adapter, policy)
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

func selectedRuleProvider() (filter.Provider, error) {
	provider, err := selectedSystemFirewallProvider()
	if err != nil {
		return "", fmt.Errorf("%w: %v", filter.ErrProviderUnavailable, err)
	}
	return filter.Provider(provider), nil
}

type firewallRuleCreateAuthorization = filter.CreateAuthorization

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
	return firewallCheckFlagCodec().Sign(result, snapshot, managedRevision)
}

func authorizeFirewallRuleCreate(
	checkFlag string,
	action filter.CheckAction,
	adoptInstanceKey string,
	rule filter.FirewallRule,
	snapshot filter.Snapshot,
	managedRevision string,
) (firewallRuleCreateAuthorization, error) {
	return firewallCheckFlagCodec().Authorize(checkFlag, action, adoptInstanceKey, rule, snapshot, managedRevision)
}

func firewallCheckFlagCodec() *filter.CheckFlagCodec {
	secret := []byte(global.CONF.Base.EncryptKey + "\x00firewall-rule-check-v1")
	return filter.NewCheckFlagCodec(secret, constant.FirewallRuleCheckVersion)
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
	return firewall.ContainsPort(ports, target)
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

func ReleaseFirewallPortWhitelistAfterUpdate(oldValue string) error {
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
	removed := excludeFirewallPorts(oldPorts, ports)
	removed = excludeFirewallPorts(removed, required)
	return newFirewallService().releaseSystemPorts(context.Background(), systemPorts(removed))
}

// releaseSystemPorts only removes whitelist ownership from persisted rules. It
// deliberately leaves the active native rules unchanged so users can delete
// them explicitly from the firewall rule list.
func (s *FirewallService) releaseSystemPorts(ctx context.Context, ports []dto.FirewallSystemPort) error {
	portSet, err := normalizeSystemPorts(ports)
	if err != nil || len(portSet) == 0 {
		return err
	}
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()

	owners := make(map[string]struct{}, len(portSet)*2)
	for _, port := range portSet {
		owners[model.FirewallRuleOwner(
			constant.FirewallRuleSourceSecurity,
			constant.FirewallSystemAcceptedPortSourcePrefix+systemPortKey(port),
		)] = struct{}{}
		if port.Family == constant.FirewallFamilyIPv4 {
			owners[model.FirewallRuleOwner(
				constant.FirewallRuleSourceSecurity,
				constant.FirewallSystemAcceptedPortSourcePrefix+legacySystemPortKey(port),
			)] = struct{}{}
		}
	}
	records, err := s.rules.List(ctx)
	if err != nil {
		return err
	}
	for _, record := range records {
		if _, exists := owners[record.Owner]; !exists {
			continue
		}
		if err := s.rules.UpdateWithRevision(ctx, record.UUID, record.Revision, map[string]interface{}{
			"owner": constant.FirewallRuleSourceUser,
		}); err != nil {
			return fmt.Errorf("release accepted firewall port rule %q: %w", record.UUID, err)
		}
	}
	return nil
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
	ctx := context.Background()
	provider := filter.Provider(client.Name())
	if provider == filter.ProviderIptables || provider == filter.ProviderNftables {
		isInit, _, err := loadDirectFirewallInitStatus(string(provider))
		if err != nil {
			return err
		}
		if !isInit {
			return nil
		}
	}
	if err := s.restoreStoredFirewallRules(ctx, provider); err != nil {
		return err
	}
	if provider == filter.ProviderIptables || provider == filter.ProviderNftables {
		if provider == filter.ProviderIptables {
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
		return s.SyncSystemPorts(ctx, nil, systemPorts(excludeFirewallPorts(configured, required)))
	}
	portWhitelist, err := loadFirewallPortWhiteList()
	if err != nil {
		return err
	}
	return s.SyncSystemPorts(ctx, nil, systemPorts(portWhitelist))
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
	return firewall.ExcludePorts(ports, excluded)
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
