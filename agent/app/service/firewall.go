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
	filterufw "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/ufw"
	filterruntime "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/runtime"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/nftables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/ping"
	firewallsync "github.com/1Panel-dev/1Panel/agent/utils/firewall/sync"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FirewallService struct {
	rules                  repo.IFirewallRuleRepo
	adapters               firewallRuleRuntimeResolver
	forwardingSync         firewallDatabaseSyncAdapter
	dockerSync             firewallDatabaseSyncAdapter
	selectedProvider       func(context.Context) (filter.Provider, error)
	requiredPorts          func() ([]firewall.PortWhitelist, error)
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
	Resolve(filter.Provider) (*filterruntime.Engine, error)
	Providers() []filter.Provider
}

var firewallRuleMutationMu sync.Mutex

type IFirewallService interface {
	UpdatePanelPort(context.Context, uint, uint) error
	LoadBaseInfo(chainGroup string) (dto.FirewallSubsystemStatus, error)
	OperateFirewall(request dto.FirewallLifecycleOperation) error
	OperateFilterChain(request dto.FilterChainOperation) error
	QueueFilterChainInitialization(request dto.FilterChainOperation) (dto.FilterChainOperationResponse, error)
	Reset(context.Context, dto.FirewallRuleReset) (dto.FirewallRuleResetResponse, error)
	Inventory(context.Context, dto.FirewallRuleInventory) (dto.FirewallRuleInventoryResponse, error)
	LoadFirewallNativeDetail(context.Context, dto.FirewallNativeDetail) (string, error)
	Adopt(context.Context, dto.FirewallRuleAdopt) error
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
		adapters:               filterruntime.NewRegistry(firewallRuleSnapshotPolicy),
		forwardingSync:         newForwardingService(),
		dockerSync:             newDockerPortGuardService(),
		selectedProvider:       firewallRuleSelectedProvider,
		requiredPorts:          LoadRequiredFirewallPortWhiteList,
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

type firewallLifecycleClient struct{ lifecycle.Client }

func (c firewallLifecycleClient) Start() error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	return c.Client.Start()
}

func (c firewallLifecycleClient) Stop() error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	return c.Client.Stop()
}

func (c firewallLifecycleClient) Restart() error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	return c.Client.Restart()
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
	baseClient := s.baseClient
	if baseClient == nil {
		baseClient = selectedSystemFirewallClient
	}
	client, err := baseClient()
	if err != nil {
		return err
	}
	operation := lifecycle.Operation(request.Operation)
	operationErr := lifecycle.NewOperator(firewallLifecycleClient{client}).Operate(operation, request.WithDockerRestart, s.addPortsBeforeStart)
	restoreFirewalld := client.Name() == lifecycle.ProviderFirewalld &&
		(operation == lifecycle.OperationStart || operation == lifecycle.OperationRestart)
	if operation != lifecycle.OperationStart && operation != lifecycle.OperationRestart {
		return operationErr
	}
	if operationErr != nil {
		var completedErr *lifecycle.CompletedOperationError
		var dockerRestartErr *lifecycle.DockerRestartError
		if !errors.As(operationErr, &completedErr) && !errors.As(operationErr, &dockerRestartErr) {
			return operationErr
		}
		if global.LOG != nil {
			global.LOG.Warnf("firewall %s completed with post-start recovery errors: %v", operation, operationErr)
		}
	}
	if restoreFirewalld {
		restoreErr := s.restoreFirewalldRuntimeDependents(context.Background(), operation)
		if restoreErr != nil && global.LOG != nil {
			global.LOG.Errorf("restore firewalld runtime dependents after %s failed: %v", operation, restoreErr)
		}
		return nil
	}
	ReconcileDockerPortGuardBestEffort(context.Background())
	return nil
}

func (s *FirewallService) restoreFirewalldRuntimeDependents(ctx context.Context, operation lifecycle.Operation) error {
	restoreForwarding := s.restoreForwarding
	if restoreForwarding == nil {
		restoreForwarding = func(ctx context.Context) error { return newForwardingService().Restore(ctx) }
	}
	restoreDockerGuard := s.restoreDockerGuard
	if restoreDockerGuard == nil {
		restoreDockerGuard = ReconcileDockerPortGuard
	}
	dockerActive := s.dockerActive
	if dockerActive == nil {
		dockerActive = func() (bool, error) { return controller.CheckActive("docker") }
	}

	active, err := dockerActive()
	restoreErr := restoreFirewalldDependents(
		ctx, fmt.Sprintf("after firewalld %s", operation), err == nil && active, restoreForwarding, restoreDockerGuard,
	)
	if err != nil {
		return errors.Join(fmt.Errorf("check Docker status after firewalld %s: %w", operation, err), restoreErr)
	}
	return restoreErr
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

	taskItem, err := task.NewTask(firewallTaskName(task.TaskExec, firewallTaskHost, provider), task.TaskExec, task.TaskScopeFirewall, request.TaskID, 0)
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
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	return s.operateFilterChainBaseLocked(provider, request)
}

func (s *FirewallService) operateFilterChainBaseLocked(provider string, request dto.FilterChainOperation) error {
	if err := s.checkSelectedProvider(context.Background(), filter.Provider(provider)); err != nil {
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
	return nil
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
		restoreErr := restoreFirewalldDependents(
			ctx, "after resetting firewalld", restartDocker, restoreForwarding, restoreDockerGuard,
		)
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
	reason string,
	restoreDocker bool,
	restoreForwarding func(context.Context) error,
	restoreDockerGuard func(context.Context) error,
) error {
	var errs []error
	if err := restoreForwarding(ctx); err != nil {
		errs = append(errs, fmt.Errorf("restore port forwarding %s: %w", reason, err))
	}
	if restoreDocker {
		if err := restoreDockerGuard(ctx); err != nil {
			errs = append(errs, fmt.Errorf("restore Docker port guard %s: %w", reason, err))
		}
	}
	return errors.Join(errs...)
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
	desiredByScope, failures := s.desiredFirewallRulesByScope(ctx, stored, provider)
	response := dto.FirewallRuleInventoryResponse{Items: failures}
	unavailable := make(map[filter.Family]error)
	for _, scope := range scopes {
		var snapshot filter.Snapshot
		err := unavailable[scope.Family]
		if err == nil {
			snapshot, err = runtime.Observe(ctx, scope)
		}
		if errors.Is(err, filter.ErrFamilyUnavailable) {
			if unavailable[scope.Family] == nil {
				response.Notices = append(response.Notices, filter.ScopeNotice{Code: filter.ScopeNoticeFamilyUnavailable, Values: []string{string(scope.Family), err.Error()}})
				unavailable[scope.Family] = err
			}
			for _, desired := range desiredByScope[scope.Key()] {
				response.Items = append(response.Items, filter.InventoryItem{
					Rule: desired.Rule, Desired: &desired, State: filter.InventoryStateDrifted,
					Match: filter.InventoryMatchNone, Error: err.Error(),
				})
			}
			continue
		}
		if err != nil {
			return dto.FirewallRuleInventoryResponse{}, err
		}
		desired := desiredByScope[scope.Key()]
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
	provider := request.Scope.Provider
	if len(request.Scopes) > 0 {
		provider = request.Scopes[0].Provider
	}
	response.IPv4Range, response.IPv6Range = filter.InventoryPositionRanges(provider, response.Items)
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
		if requested == "deny" && (action == filter.ActionDrop || action == filter.ActionReject) {
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
	runtime *filterruntime.Engine,
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

	stored, err := s.rules.List(ctx)
	if err != nil {
		return dto.FirewallRuleInventoryResponse{}, err
	}
	desiredByScope, failures := s.desiredFirewallRulesByScope(ctx, stored, scope.Provider)
	response := dto.FirewallRuleInventoryResponse{Items: failures}
	seenNotices := make(map[string]struct{})
	for index, snapshot := range snapshots {
		if snapshot.Scope.Key() != scopes[index].Key() {
			return dto.FirewallRuleInventoryResponse{}, fmt.Errorf("%w: unexpected UFW inventory scope %q", filter.ErrInvalidScope, snapshot.Scope.Key())
		}
		desired := desiredByScope[snapshot.Scope.Key()]
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

type preparedFirewallRuleCreate struct {
	request dto.FirewallRuleCreateItem
	runtime *filterruntime.Engine
}

func (s *FirewallService) Create(
	ctx context.Context,
	request dto.FirewallRuleCreate,
) (dto.FirewallRuleCreateResponse, error) {
	taskItem, err := task.NewTask(firewallTaskName(task.TaskCreate, firewallTaskHost, ""), task.TaskCreate, task.TaskScopeFirewall, "", 0)
	if err != nil {
		return dto.FirewallRuleCreateResponse{}, err
	}
	taskItem.AddSubTaskWithOps(i18n.GetMsgByKey("FirewallCreateRulesStep"), func(t *task.Task) error {
		_, err := s.runCreateTask(t, request)
		return err
	}, nil, 0, 0)
	if err := repo.NewITaskRepo().Save(context.Background(), taskItem.Task); err != nil {
		taskItem.LogFailedWithErr(taskItem.Name, err)
		closeUnstartedFirewallTask(taskItem)
		return dto.FirewallRuleCreateResponse{}, fmt.Errorf("save firewall creation task: %w", err)
	}
	go func() {
		if err := taskItem.Execute(); err != nil && global.LOG != nil {
			global.LOG.Errorf("firewall creation task %s failed: %v", taskItem.TaskID, err)
		}
	}()
	return dto.FirewallRuleCreateResponse{TaskID: taskItem.TaskID, Queued: true}, nil
}

func (s *FirewallService) runCreateTask(t *task.Task, request dto.FirewallRuleCreate) (dto.FirewallRuleCreateResponse, error) {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	return s.createRules(t.TaskCtx, request, t)
}

func (s *FirewallService) createRules(ctx context.Context, request dto.FirewallRuleCreate, t *task.Task) (result dto.FirewallRuleCreateResponse, taskErr error) {
	var firstFailure error
	defer func() {
		if t != nil {
			t.Log(i18n.GetMsgWithMap("FirewallCreateRulesResult", map[string]interface{}{
				"succeeded": result.Succeeded, "failed": result.Failed, "skipped": result.Skipped,
			}))
		}
		if taskErr == nil {
			taskErr = firstFailure
		}
	}()
	type itemOrigin struct {
		index, part, count int
		rule               filter.FirewallRule
	}
	describe := func(rule filter.FirewallRule) string {
		return fmt.Sprintf("%s %s %s:%s -> %s:%s %s", rule.Scope.Family, rule.Protocol,
			rule.SourceAddress, rule.SourcePort, rule.DestinationAddress, rule.DestinationPort, rule.Action)
	}
	record := func(origin itemOrigin, status string, err error) {
		rule := origin.rule
		label := fmt.Sprintf("[%d/%d]", origin.index+1, len(request.Items))
		if origin.count > 1 {
			label += fmt.Sprintf("[%d/%d]", origin.part+1, origin.count)
		}
		label += fmt.Sprintf(" %s %s", rule.Scope.Provider, describe(rule))
		switch status {
		case "succeeded":
			result.Succeeded++
			if t != nil {
				t.LogSuccess(label)
			}
		case "failed":
			if firstFailure == nil {
				firstFailure = err
			}
			result.Failed++
			if t != nil {
				t.LogFailedWithErr(label, err)
			}
		case "skipped":
			result.Skipped++
			if t != nil {
				t.Logf("%s %s: %v", label, i18n.GetMsgByKey("FirewallCreateRuleSkipped"), err)
			}
		}
		if err != nil {
			result.Errors = append(result.Errors, dto.FirewallRuleCreateFailure{
				Index: origin.index, Status: status, Rule: rule, Error: err.Error(),
			})
		}
	}
	selected, err := s.selectedProvider(ctx)
	if err != nil {
		for index := range request.Items {
			record(itemOrigin{index: index, rule: request.Items[index].Rule}, "skipped", err)
		}
		return result, err
	}
	var stop error
	var prepared []preparedFirewallRuleCreate
	var origins []itemOrigin
	flush := func() {
		if len(prepared) == 0 {
			return
		}
		defer func() { prepared, origins = nil, nil }()
		if stop == nil {
			stop = ctx.Err()
		}
		if stop != nil {
			for _, origin := range origins {
				record(origin, "skipped", stop)
			}
			return
		}
		runtime := prepared[0].runtime
		scope := prepared[0].request.Rule.Scope
		snapshot, err := runtime.ObserveMutation(ctx, scope)
		if err != nil {
			for _, origin := range origins {
				record(origin, "failed", err)
			}
			if !errors.Is(err, filter.ErrFamilyUnavailable) {
				stop = err
			}
			return
		}
		stored, err := s.rules.List(ctx)
		if err != nil {
			for _, origin := range origins {
				record(origin, "failed", err)
			}
			stop = err
			return
		}
		identities, err := firewallRuleCollisions(stored, runtime.Provider(), "")
		if err != nil {
			for _, origin := range origins {
				record(origin, "failed", err)
			}
			stop = err
			return
		}
		valid := prepared[:0]
		validOrigins := origins[:0]
		for index, entry := range prepared {
			rule := entry.request.Rule
			checkErr := identities.Check(rule)
			if checkErr == nil {
				checkErr = filter.CheckObservedRuleCollisions(snapshot, entry.request.Rule, nil)
			}
			if checkErr == nil {
				checkErr = identities.Add(rule)
			}
			if checkErr != nil {
				record(origins[index], "failed", checkErr)
				continue
			}
			valid = append(valid, entry)
			validOrigins = append(validOrigins, origins[index])
		}
		if len(valid) == 0 {
			return
		}
		if len(valid) > 1 && t != nil {
			t.Log(i18n.GetMsgWithMap("FirewallCreateBatchStep", map[string]interface{}{
				"backend": runtime.Provider(), "count": len(valid),
			}))
		}
		itemErrors := s.applyCreateRules(ctx, runtime, snapshot, stored, valid)
		for offset, origin := range validOrigins {
			if err := itemErrors[offset]; err != nil {
				record(origin, "failed", err)
				if firewallCreateUnavailable(err) {
					stop = err
				}
			} else {
				record(origin, "succeeded", nil)
			}
		}
	}
	for index, item := range request.Items {
		if stop == nil {
			stop = ctx.Err()
		}
		origin := itemOrigin{index: index, rule: item.Rule}
		if stop != nil {
			record(origin, "skipped", stop)
			continue
		}
		rules := []filter.FirewallRule{item.Rule}
		if item.SourceKind == constant.FirewallRuleSourceImported {
			rules, err = convertImportedFirewallRule(item.Rule, selected)
			if err != nil {
				record(origin, "failed", err)
				continue
			}
			if t != nil {
				t.Log(i18n.GetMsgWithMap("FirewallImportRuleConversion", map[string]interface{}{
					"index": index + 1, "total": len(request.Items), "source": item.Rule.Scope.Provider,
					"target": selected, "rule": describe(item.Rule), "count": len(rules),
				}))
			}
		}
		for part, rule := range rules {
			origin := itemOrigin{index: index, part: part, count: len(rules), rule: rule}
			scope := applySelectedProviderScopeDefaults(rule, selected).Scope.Normalize()
			if len(prepared) > 0 && (prepared[0].request.Rule.Scope != scope || rule.OrderIndex != nil) {
				flush()
			}
			if stop == nil {
				stop = ctx.Err()
			}
			if stop != nil {
				record(origin, "skipped", stop)
				continue
			}
			child := item
			child.Rule = rule
			entry, prepareErr := s.prepareCreate(ctx, selected, child)
			if prepareErr != nil {
				record(origin, "failed", prepareErr)
				if firewallCreateUnavailable(prepareErr) {
					stop = prepareErr
				}
				continue
			}
			origin.rule = entry.request.Rule
			prepared = append(prepared, entry)
			origins = append(origins, origin)
			if !supportsNativeRuleBatch(selected) || rule.OrderIndex != nil {
				flush()
			}
		}
	}
	flush()
	return result, stop
}

func convertImportedFirewallRule(rule filter.FirewallRule, selected filter.Provider) ([]filter.FirewallRule, error) {
	source := rule.Scope.Normalize().Provider
	if source == "" {
		source = selected
	}
	rules, err := filter.ExpandAtomicRules(applySelectedProviderScopeDefaults(rule, source))
	if err != nil {
		return nil, err
	}
	var converted []filter.FirewallRule
	for _, sourceRule := range rules {
		policy, err := model.FirewallRuleFromDomain(sourceRule)
		if err != nil {
			return nil, err
		}
		targetRules, err := policy.RulesForProvider(selected)
		if err != nil {
			return nil, err
		}
		converted = append(converted, targetRules...)
	}
	return converted, nil
}

func firewallCreateUnavailable(err error) bool {
	return errors.Is(err, filter.ErrProviderUnavailable) || errors.Is(err, filter.ErrAdapterUnavailable) ||
		errors.Is(err, filter.ErrInventoryUnavailable) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func (s *FirewallService) prepareCreate(ctx context.Context, selected filter.Provider, request dto.FirewallRuleCreateItem) (preparedFirewallRuleCreate, error) {
	rule, err := filter.NormalizeRule(applySelectedProviderScopeDefaults(request.Rule, selected))
	if err != nil {
		return preparedFirewallRuleCreate{}, err
	}
	if rule.Scope.Provider != selected {
		return preparedFirewallRuleCreate{}, fmt.Errorf("%w: selected provider is %s", filter.ErrInvalidRule, selected)
	}
	runtime, err := s.adapters.Resolve(selected)
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
	rule.UUID = ""
	request.Rule = rule
	if request.SourceKind == "" {
		request.SourceKind = constant.FirewallRuleSourceUser
	}
	return preparedFirewallRuleCreate{request: request, runtime: runtime}, nil
}

func (s *FirewallService) applyCreateRules(ctx context.Context, runtime *filterruntime.Engine, snapshot filter.Snapshot, stored []model.FirewallRule, prepared []preparedFirewallRuleCreate) []error {
	results := make([]error, len(prepared))
	failAll := func(err error) []error {
		for index := range results {
			results[index] = err
		}
		return results
	}
	var maximumSequence int64
	for _, record := range stored {
		if record.Sequence != nil && *record.Sequence > maximumSequence {
			maximumSequence = *record.Sequence
		}
	}
	records := make([]model.FirewallRule, 0, len(prepared))
	changes := make([]filter.DesiredChange, 0, len(prepared))
	for _, entry := range prepared {
		rule := entry.request.Rule
		appendRule := false
		if rule.Scope.Provider == filter.ProviderUFW && rule.OrderIndex == nil {
			position, err := runtime.AppendPosition(ctx, snapshot, rule)
			if err != nil {
				return failAll(err)
			}
			rule.OrderIndex, appendRule = &position, true
		} else if rule.OrderIndex != nil {
			maximum, err := runtime.MaxPosition(ctx, snapshot, rule)
			if err != nil {
				return failAll(err)
			}
			if *rule.OrderIndex < 1 || *rule.OrderIndex > maximum+1 {
				return failAll(fmt.Errorf("%w: create target position %d is out of range 1-%d", filter.ErrInvalidRule, *rule.OrderIndex, maximum+1))
			}
			appendRule = rule.Scope.Provider == filter.ProviderUFW && *rule.OrderIndex == maximum+1
		}
		record, err := firewallRuleModelForCreate(rule, entry.request, constant.FirewallRuleOriginCreated)
		if err != nil {
			return failAll(err)
		}
		if rule.Scope.Provider != filter.ProviderFirewalld {
			maximumSequence += model.FirewallRuleSequenceStep
			sequence := maximumSequence
			if rule.OrderIndex != nil {
				sequence, err = s.sequenceForCreatedFirewallRule(ctx, snapshot, rule)
				if err != nil {
					return failAll(err)
				}
			}
			record.Sequence = &sequence
		}
		record.UUID = uuid.NewString()
		rule.UUID = record.UUID
		records = append(records, record)
		changes = append(changes, filter.DesiredChange{Operation: filter.ChangeCreate, After: &rule, Append: appendRule})
	}
	if err := runtime.ExecuteCreate(ctx, snapshot, changes); err != nil {
		return failAll(firewallCreateExecutionError(err))
	}
	for index := range records {
		results[index] = s.saveFirewallRule(ctx, &records[index])
	}
	return results
}

func firewallCreateExecutionError(err error) error {
	return fmt.Errorf("%s: %w", i18n.GetMsgByKey("FirewallCreateRuleExecutionFailed"), err)
}

func (s *FirewallService) saveFirewallRule(ctx context.Context, record *model.FirewallRule) error {
	if err := s.rules.Create(ctx, record); err != nil {
		message := "FirewallCreateRulePersistenceFailed"
		if record.Origin == constant.FirewallRuleOriginAdopted {
			message = "FirewallAdoptRulePersistenceFailed"
		}
		return fmt.Errorf("%s: %w", i18n.GetMsgByKey(message), err)
	}
	return nil
}

type preparedFirewallRuleDelete struct {
	index    int
	stored   model.FirewallRule
	desired  filter.DesiredRule
	runtime  *filterruntime.Engine
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
		if prepared.compiled != 1 {
			groupKey += ":" + prepared.stored.UUID
		}
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
	metadata := request.Description != nil || request.OrderIndex != nil || request.Priority != nil
	if (request.Rule != nil) == metadata || request.OrderIndex != nil && request.Priority != nil {
		return fmt.Errorf("%w: provide a rule or description/ordering fields", filter.ErrInvalidRule)
	}
	if request.Rule != nil {
		rule := *request.Rule
		rule.UUID = request.UUID
		return s.updateRule(ctx, clientIP, request.UUID, rule)
	}
	if request.OrderIndex != nil || request.Priority != nil {
		return s.updateRuleOrder(ctx, request.UUID, request.OrderIndex, request.Priority, request.Description)
	}
	return s.updateRuleDescription(ctx, request.UUID, *request.Description)
}

func (s *FirewallService) updateRuleDescription(ctx context.Context, ruleUUID, description string) error {
	stored, err := s.rules.GetByUUID(ctx, ruleUUID)
	if err != nil {
		return err
	}
	if isProtectedSystemFirewallRule(stored) {
		return filter.ErrProtectedRule
	}
	if stored.Origin != constant.FirewallRuleOriginCreated && stored.Origin != constant.FirewallRuleOriginAdopted {
		return fmt.Errorf("%w: only created or adopted rules can be changed", filter.ErrInvalidRule)
	}
	description = strings.TrimSpace(description)
	if stored.Description == description {
		return nil
	}
	return s.rules.UpdateWithRevision(ctx, stored.UUID, stored.Revision, map[string]interface{}{"description": description})
}

func (s *FirewallService) Reorder(ctx context.Context, clientIP string, request dto.FirewallRuleReorder) error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	return s.updateRuleOrder(ctx, request.UUID, request.TargetPosition, request.Priority, nil)
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

func (s *FirewallService) Adopt(ctx context.Context, request dto.FirewallRuleAdopt) error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	if err := s.checkSelectedProvider(ctx, request.Scope.Provider); err != nil {
		return err
	}
	runtime, err := s.adapters.Resolve(request.Scope.Provider)
	if err != nil {
		return err
	}
	snapshot, err := runtime.ObserveMutation(ctx, request.Scope)
	if err != nil {
		return err
	}
	observed, err := filter.FindCandidate(snapshot.Rules, request.InstanceKey)
	if err != nil {
		return filter.ErrRuleStale
	}
	return s.adoptRule(ctx, runtime, snapshot, observed, dto.FirewallRuleCreateItem{SourceKind: constant.FirewallRuleSourceUser})
}

func (s *FirewallService) adoptRule(ctx context.Context, runtime *filterruntime.Engine, snapshot filter.Snapshot, observed filter.ObservedRule, source dto.FirewallRuleCreateItem) error {
	if observed.Protected {
		return filter.ErrProtectedRule
	}
	if observed.ParseStatus != filter.ParseStatusSupported ||
		(observed.Persistence != "" && observed.Persistence != filter.PersistenceStatusConverged) {
		return fmt.Errorf("%w: rule cannot be managed", filter.ErrRuleOperation)
	}
	rule, err := runtime.Prepare(observed.Rule)
	if err != nil {
		return err
	}
	if err := runtime.CheckRule(ctx, rule); err != nil {
		return err
	}
	if rule.Scope.Provider != filter.ProviderFirewalld && observed.Locator.Position != nil {
		position := int64(*observed.Locator.Position)
		rule.OrderIndex = &position
	}
	record, err := firewallRuleModelForCreate(rule, source, constant.FirewallRuleOriginAdopted)
	if err != nil {
		return err
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		return err
	}
	if err := filter.CheckAdoptDuplicates(snapshot, rule); err != nil {
		return err
	}
	identities, err := firewallRuleCollisions(stored, rule.Scope.Provider, "")
	if err != nil {
		return err
	}
	for _, existing := range stored {
		marker := "1panel-rule:" + existing.UUID
		if existing.UUID != "" && (observed.Marker == marker || strings.HasPrefix(observed.Marker, marker+"-")) {
			return fmt.Errorf("%w: rule is already managed", filter.ErrRuleOperation)
		}
	}
	if err := identities.CheckDuplicate(rule); err != nil {
		if errors.Is(err, filter.ErrRuleOperation) {
			return filter.ErrDuplicateAdoption
		}
		return err
	}
	if rule.Scope.Provider != filter.ProviderFirewalld {
		sequence, err := s.sequenceForCreatedFirewallRule(ctx, snapshot, rule)
		if err != nil {
			return err
		}
		record.Sequence = &sequence
	}
	record.UUID = uuid.NewString()
	rule.UUID = record.UUID
	plan, verification, err := runtime.Execute(ctx, snapshot, []filter.DesiredChange{{
		Operation: filter.ChangeAdopt, After: &rule, Locator: &observed.Locator, PreviousMarker: observed.Marker,
	}})
	if err != nil {
		return err
	}
	if !verification.Matched {
		return filter.ErrVerificationFailed
	}
	if _, err := filter.FindCommittedObserved(verification.Snapshot, rule, plan); err != nil {
		return rollbackFirewallPlan(ctx, runtime, plan, err)
	}
	return s.saveFirewallRule(ctx, &record)
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
	if ports, ok := ctx.Value(panelPortWhitelistKey{}).([]firewall.PortWhitelist); ok {
		for _, desired := range desiredRules {
			if panelRuleStillRequired(desired.Rule, ports) {
				return filter.ErrProtectedRule
			}
		}
	}
	type appliedDelete struct {
		runtime *filterruntime.Engine
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
	requestedRule, err := filter.NormalizeRule(requestedRule)
	if err != nil {
		return err
	}
	stored, err := s.rules.GetByUUID(ctx, ruleUUID)
	if err != nil {
		return err
	}
	previousRules, compileErr := stored.RulesForProvider(requestedRule.Scope.Provider)
	if compileErr == nil && len(previousRules) == 1 {
		sameContent, err := filter.SameRuleContent(previousRules[0], requestedRule)
		if err != nil {
			return err
		}
		if sameContent {
			if requestedRule.Scope.Provider == filter.ProviderFirewalld {
				beforePriority, afterPriority := 0, 0
				if stored.Priority != nil {
					beforePriority = *stored.Priority
				}
				if requestedRule.Priority != nil {
					afterPriority = *requestedRule.Priority
				}
				if beforePriority != afterPriority {
					return s.updateRuleOrder(ctx, ruleUUID, nil, &afterPriority, &requestedRule.Description)
				}
			} else if requestedRule.OrderIndex != nil {
				return s.updateRuleOrder(ctx, ruleUUID, requestedRule.OrderIndex, nil, &requestedRule.Description)
			}
			return s.updateRuleDescription(ctx, ruleUUID, requestedRule.Description)
		}
	}
	prepared, err := s.prepareManagedUpdate(ctx, clientIP, ruleUUID, requestedRule)
	if err != nil {
		return err
	}
	metadataOnly, err := isFirewallMetadataOnlyUpdate(prepared.Before.Rule, prepared.After, prepared.Observed.Locator)
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

func isFirewallMetadataOnlyUpdate(before, after filter.FirewallRule, locator filter.Locator) (bool, error) {
	beforeKey, err := filter.RuleKey(before)
	if err != nil {
		return false, err
	}
	afterKey, err := filter.RuleKey(after)
	if err != nil {
		return false, err
	}
	if beforeKey != afterKey {
		return false, nil
	}
	if after.Scope.Provider == filter.ProviderFirewalld {
		return true, nil
	}
	return locator.Position != nil && after.OrderIndex != nil && *after.OrderIndex == int64(*locator.Position), nil
}

func (s *FirewallService) updateRuleOrder(ctx context.Context, ruleUUID string, targetPosition *int64, priority *int, description *string) error {
	if (targetPosition == nil) == (priority == nil) {
		return fmt.Errorf("%w: provide either position or priority", filter.ErrInvalidRule)
	}
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
	case capabilities.ExplicitPosition || capabilities.OwnedChains:
		if targetPosition == nil || *targetPosition < 1 {
			return fmt.Errorf("%w: target position is required", filter.ErrInvalidRule)
		}
		if err := runtime.ValidatePosition(ctx, snapshot, before.Rule, *targetPosition); err != nil {
			return err
		}
		after.OrderIndex = targetPosition
	case capabilities.ExplicitPriority:
		if before.Rule.NativeKind != filter.NativeKindRichRule {
			return fmt.Errorf("%w: only rich rules support explicit priority", filter.ErrUnsupportedScope)
		}
		if priority == nil {
			return fmt.Errorf("%w: priority is required", filter.ErrInvalidRule)
		}
		after.Priority = priority
		adapterOperation = filter.ChangeUpdate
	default:
		return fmt.Errorf("%w: provider does not support rule reordering", filter.ErrUnsupportedScope)
	}
	if description != nil {
		after.Description = strings.TrimSpace(*description)
	}
	after, err = runtime.Prepare(after)
	if err != nil {
		return err
	}
	if err := runtime.CheckRule(ctx, after); err != nil {
		return err
	}
	metadataOnly, err := isFirewallMetadataOnlyUpdate(before.Rule, after, observed.Locator)
	if err != nil {
		return err
	}
	if metadataOnly {
		return s.updateRuleDescription(ctx, stored.UUID, after.Description)
	}
	if err := filter.GuardMutation(observed); err != nil {
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
	Runtime          *filterruntime.Engine
}

type preparedManagedUpdate struct {
	Stored   model.FirewallRule
	Before   filter.DesiredRule
	After    filter.FirewallRule
	Snapshot filter.Snapshot
	Observed filter.ObservedRule
	Runtime  *filterruntime.Engine
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
		return preparedManagedUpdate{}, filter.ErrManagedScopeChange
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
	if err := filter.GuardMutation(observed); err != nil {
		return preparedManagedUpdate{}, err
	}
	if err := s.checkManagedMutationCollisions(ctx, before.Rule, after, snapshot, observed.Locator, stored.UUID); err != nil {
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
) (model.FirewallRule, filter.DesiredRule, filter.Snapshot, filter.ObservedRule, *filterruntime.Engine, error) {
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
	appendRule, restoreAtEnd := false, false
	if after.Scope.Provider == filter.ProviderUFW && (request.AdapterOperation == filter.ChangeUpdate || request.AdapterOperation == filter.ChangeReorder) {
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

func (s *FirewallService) checkManagedMutationCollisions(
	ctx context.Context,
	before, after filter.FirewallRule,
	snapshot filter.Snapshot,
	locator filter.Locator,
	excludedUUID string,
) error {
	sameContent, err := filter.SameRuleContent(before, after)
	if err != nil {
		return err
	}
	if sameContent {
		return nil
	}
	if err := filter.CheckObservedRuleCollisions(snapshot, after, &locator); err != nil {
		return err
	}
	return s.ensureFirewallRuleIdentityAvailable(ctx, after, excludedUUID)
}

func (s *FirewallService) ensureFirewallRuleIdentityAvailable(ctx context.Context, requested filter.FirewallRule, excludedUUID string) error {
	stored, err := s.rules.List(ctx)
	if err != nil {
		return err
	}
	identities, err := firewallRuleCollisions(stored, requested.Scope.Provider, excludedUUID)
	if err != nil {
		return err
	}
	return identities.Check(requested)
}

func firewallRuleCollisions(stored []model.FirewallRule, provider filter.Provider, excludedUUID string) (filter.RuleCollisionIndex, error) {
	identities := make(filter.RuleCollisionIndex, len(stored))
	for _, candidate := range stored {
		if candidate.UUID == excludedUUID {
			continue
		}
		rules, err := candidate.RulesForProvider(provider)
		if err != nil {
			continue
		}
		for _, rule := range rules {
			if err := identities.Add(rule); err != nil {
				return nil, err
			}
		}
	}
	return identities, nil
}

func (s *FirewallService) resolveRuntime(ctx context.Context, provider filter.Provider) (*filterruntime.Engine, error) {
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

func (s *FirewallService) ensureSystemPort(ctx context.Context, port dto.FirewallSystemPort) error {
	firewallRuleMutationMu.Lock()
	err := s.ensureSystemPortLocked(ctx, port)
	firewallRuleMutationMu.Unlock()
	if err == nil {
		return nil
	}
	if errors.Is(err, filter.ErrInventoryUnavailable) {
		err = s.appendUFWSystemPortUnverified(ctx, port, err)
	}
	if port.Family == constant.FirewallFamilyIPv6 && filterufw.IsIPv6Unavailable(err) {
		if global.LOG != nil {
			global.LOG.Warnf("skip accepted UFW IPv6 port %s/%s: %v", port.Port, port.Protocol, err)
		}
		return nil
	}
	return err
}

func (s *FirewallService) appendUFWSystemPortUnverified(
	ctx context.Context,
	port dto.FirewallSystemPort,
	cause error,
) error {
	if s.selectedProvider == nil || s.adapters == nil {
		return cause
	}
	provider, providerErr := s.selectedProvider(ctx)
	if providerErr != nil {
		return errors.Join(cause, providerErr)
	}
	if provider != filter.ProviderUFW {
		return cause
	}
	if global.LOG != nil {
		global.LOG.Warnf(
			"UFW inventory is unavailable while restoring accepted port %s/%s; attempting a restricted direct allow: %v",
			port.Port, port.Protocol, cause,
		)
	}
	runtime, resolveErr := s.adapters.Resolve(provider)
	if resolveErr != nil {
		return errors.Join(cause, resolveErr)
	}
	comment := "1panel-system-port:" + systemPortKey(port)
	if appendErr := runtime.AppendUnverified(ctx, systemPortRule(provider, port), comment); appendErr != nil {
		if global.LOG != nil {
			global.LOG.Errorf(
				"restore accepted UFW port %s/%s without rule inventory failed: %v; original error: %v",
				port.Port, port.Protocol, appendErr, cause,
			)
		}
		return errors.Join(cause, fmt.Errorf("append accepted UFW port without rule inventory: %w", appendErr))
	}
	if global.LOG != nil {
		global.LOG.Warnf(
			"restored accepted UFW port %s/%s without rule inventory; normal rule management failed: %v",
			port.Port, port.Protocol, cause,
		)
	}
	return nil
}

func (s *FirewallService) ensureSystemPortLocked(ctx context.Context, port dto.FirewallSystemPort) error {
	provider, err := s.selectedProvider(ctx)
	if err != nil {
		return err
	}
	source := dto.FirewallRuleCreateItem{
		Rule: systemPortRule(provider, port), SourceKind: constant.FirewallRuleSourceSecurity,
		SourceID: constant.FirewallSystemAcceptedPortSourcePrefix + systemPortKey(port),
	}
	prepared, err := s.prepareCreate(ctx, provider, source)
	if err != nil {
		return err
	}
	rule, runtime := prepared.request.Rule, prepared.runtime
	snapshot, err := runtime.ObserveMutation(ctx, rule.Scope)
	if err != nil {
		return err
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		return err
	}
	matchKey, err := filter.RuleMatchKey(rule)
	if err != nil {
		return err
	}
	for _, record := range stored {
		candidates, err := record.RulesForProvider(provider)
		if err != nil {
			continue
		}
		matched := false
		for _, candidate := range candidates {
			key, err := filter.RuleMatchKey(candidate)
			if err != nil {
				return err
			}
			if key == matchKey {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		if record.Action != string(rule.Action) {
			return filter.ErrRuleConflict
		}
		desired, err := s.compileStoredFirewallRules(ctx, record, provider)
		if err != nil {
			return err
		}
		items, err := filter.MergeInventory(filter.InventoryMergeInput{Observed: snapshot.Rules, Desired: desired})
		if err != nil {
			return err
		}
		for _, item := range items {
			if item.Desired != nil && item.Match == filter.InventoryMatchExact && item.State != filter.InventoryStateDrifted && item.Observed != nil {
				return nil
			}
		}
		return filter.ErrRuleStale
	}
	matches, err := filter.MatchObservedByRuleKey(snapshot.Rules, rule)
	if err != nil {
		return err
	}
	if len(matches) > 0 {
		if matches[0].Protected {
			if matches[0].Persistence != "" && matches[0].Persistence != filter.PersistenceStatusConverged {
				return filter.ErrRuleStale
			}
			return nil
		}
		return s.adoptRule(ctx, runtime, snapshot, matches[0], source)
	}
	_, err = s.createRules(ctx, dto.FirewallRuleCreate{Items: []dto.FirewallRuleCreateItem{source}}, nil)
	return err
}

func (s *FirewallService) deleteSystemPort(ctx context.Context, port dto.FirewallSystemPort) error {
	stored, err := s.systemPortRecords(ctx, port)
	if err != nil {
		return err
	}
	for _, rule := range stored {
		if err := s.deleteProtectedSystemPortRule(ctx, rule.UUID); err != nil {
			if errors.Is(err, filter.ErrProtectedRule) && ctx.Value(panelPortWhitelistKey{}) != nil {
				continue
			}
			return err
		}
	}
	return nil
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

func isFirewallPolicyIncompatible(err error) bool {
	return errors.Is(err, filter.ErrInvalidRule) || errors.Is(err, filter.ErrUnsupportedScope) ||
		errors.Is(err, filter.ErrInvalidScope) || errors.Is(err, filter.ErrCompositeRule)
}

func (s *FirewallService) compileRestorableFirewallRules(
	ctx context.Context,
	stored model.FirewallRule,
	provider filter.Provider,
) (restorable, preserved []filter.DesiredRule, err error) {
	compiled, err := s.compileStoredFirewallRules(ctx, stored, provider)
	if err != nil {
		return nil, nil, err
	}
	if !supportsManagedFilterChains(string(provider)) || !isProtectedSystemFirewallRule(stored) {
		return compiled, nil, nil
	}
	loadRequired := s.requiredPorts
	if loadRequired == nil {
		loadRequired = LoadRequiredFirewallPortWhiteList
	}
	required, err := loadRequired()
	if err != nil {
		return nil, nil, err
	}
	requiredPorts := systemPorts(required)
	for _, desired := range compiled {
		covered := false
		for _, port := range requiredPorts {
			covered, err = filter.SameRuleContent(desired.Rule, systemPortRule(provider, port))
			if err != nil {
				return nil, nil, err
			}
			if covered {
				break
			}
		}
		if covered {
			preserved = append(preserved, desired)
		} else {
			restorable = append(restorable, desired)
		}
	}
	return restorable, preserved, nil
}

func (s *FirewallService) desiredFirewallRulesByScope(
	ctx context.Context,
	stored []model.FirewallRule,
	provider filter.Provider,
) (map[string][]filter.DesiredRule, []filter.InventoryItem) {
	model.SortFirewallRules(stored, provider)
	desired := make(map[string][]filter.DesiredRule)
	var failures []filter.InventoryItem
	for _, record := range stored {
		compiled, _, err := s.compileRestorableFirewallRules(ctx, record, provider)
		if err != nil {
			rule := filter.FirewallRule{
				UUID:     record.UUID,
				Scope:    filter.Scope{Provider: provider, Family: filter.Family(record.Family), Direction: filter.DirectionInput}.Normalize(),
				Protocol: record.Protocol, SourceAddress: record.SourceAddress, SourcePort: record.SourcePort,
				DestinationAddress: record.DestinationAddress, DestinationPort: record.DestinationPort,
				Interface: record.Interface, ConnectionStates: strings.FieldsFunc(record.ConnectionStates, func(r rune) bool { return r == ',' }),
				Action: filter.Action(record.Action), Description: record.Description, Priority: record.Priority,
			}
			failures = append(failures, filter.InventoryItem{
				Incompatible: isFirewallPolicyIncompatible(err),
				Rule:         rule, State: filter.InventoryStateDrifted, Match: filter.InventoryMatchNone,
				Desired: &filter.DesiredRule{UUID: record.UUID, Rule: rule, Origin: filter.RuleOrigin(record.Origin), Protected: isProtectedSystemFirewallRule(record)},
				Error:   fmt.Sprintf("policy %s: %v", record.UUID, err),
			})
			continue
		}
		for _, rule := range compiled {
			rule.Protected = isProtectedSystemFirewallRule(record)
			rule.Expanded = len(compiled) > 1
			key := rule.Rule.Scope.Key()
			desired[key] = append(desired[key], rule)
		}
	}
	return desired, failures
}

func firewallRuleSnapshotPolicy(ctx context.Context, snapshot filter.Snapshot) (filter.Snapshot, error) {
	if ports, ok := ctx.Value(panelPortWhitelistKey{}).([]firewall.PortWhitelist); ok {
		return filter.ProtectSnapshot(snapshot, ports)
	}
	ports, err := loadFirewallPortWhiteList()
	if err != nil {
		return filter.Snapshot{}, err
	}
	return filter.ProtectSnapshot(snapshot, ports)
}

func firewallRuleSelectedProvider(context.Context) (filter.Provider, error) {
	return selectedRuleProvider()
}

func rollbackFirewallPlan(ctx context.Context, runtime *filterruntime.Engine, plan filter.BackendPlan, cause error) error {
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
		required, err := LoadRequiredFirewallPortWhiteList()
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
	required, err := LoadRequiredFirewallPortWhiteList()
	if err != nil {
		return nil, err
	}
	return firewall.NormalizePortWhitelist(append(configured, required...)), nil
}

func LoadRequiredFirewallPortWhiteList() ([]firewall.PortWhitelist, error) {
	return loadRequiredFirewallPorts(LoadPanelPort())
}

func loadRequiredFirewallPorts(panelPort string) ([]firewall.PortWhitelist, error) {
	if panelPort == "" {
		return nil, fmt.Errorf("find 1panel service port failed")
	}
	directives, _, err := parseSSHConfigTree(sshPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load required SSH ports: %w", err)
	}
	ports := []firewall.PortWhitelist{{Port: panelPort, Protocol: "tcp"}}
	for _, port := range loadSSHPortValues(directives) {
		ports = append(ports, firewall.PortWhitelist{Port: port, Protocol: "tcp"})
	}
	return firewall.NormalizeRequiredPorts(ports)
}

func (s *FirewallService) releaseSystemPorts(ctx context.Context, ports []dto.FirewallSystemPort) error {
	portSet, err := normalizeSystemPorts(ports)
	if err != nil || len(portSet) == 0 {
		return err
	}

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
		LoadRequiredPorts: LoadRequiredFirewallPortWhiteList,
	}
}

func newNftablesHelperManager() *nftables_helper.Manager {
	return &nftables_helper.Manager{
		UpdateSetting:     settingRepo.Update,
		LoadRequiredPorts: LoadRequiredFirewallPortWhiteList,
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
	var recoveryErrors []error
	recordFailure := func(stage string, err error) {
		if err == nil {
			return
		}
		wrapped := fmt.Errorf("%s for %s: %w", stage, provider, err)
		recoveryErrors = append(recoveryErrors, wrapped)
		if global.LOG != nil {
			global.LOG.Errorf("firewall post-start recovery failed: %v", wrapped)
		}
	}
	if provider == filter.ProviderIptables || provider == filter.ProviderNftables {
		isInit, _, err := loadDirectFirewallInitStatus(string(provider))
		if err != nil {
			recordFailure("load managed chain status", err)
			return errors.Join(recoveryErrors...)
		}
		if !isInit {
			return nil
		}
	}
	if err := s.restoreStoredFirewallRules(ctx, provider); err != nil {
		recordFailure("restore stored firewall rules", err)
	}
	if provider == filter.ProviderIptables || provider == filter.ProviderNftables {
		if provider == filter.ProviderIptables {
			if err := newIptablesHelperManager().SyncRequiredPorts(true); err != nil {
				recordFailure("synchronize required ports", err)
			}
		} else if err := newNftablesHelperManager().SyncRequiredPorts(); err != nil {
			recordFailure("synchronize required ports", err)
		}
		configured, err := loadConfiguredFirewallPortWhiteList()
		if err != nil {
			recordFailure("load configured accepted ports", err)
			return errors.Join(recoveryErrors...)
		}
		required, err := LoadRequiredFirewallPortWhiteList()
		if err != nil {
			recordFailure("load required accepted ports", err)
			return errors.Join(recoveryErrors...)
		}
		recordFailure(
			"restore configured accepted ports",
			s.SyncSystemPorts(ctx, nil, systemPorts(excludeFirewallPorts(configured, required))),
		)
		return errors.Join(recoveryErrors...)
	}
	portWhitelist, err := loadFirewallPortWhiteList()
	if err != nil {
		recordFailure("load accepted ports", err)
		return errors.Join(recoveryErrors...)
	}
	recordFailure("restore accepted ports", s.SyncSystemPorts(ctx, nil, systemPorts(portWhitelist)))
	return errors.Join(recoveryErrors...)
}

func systemPorts(ports []firewall.PortWhitelist) []dto.FirewallSystemPort {
	result := make([]dto.FirewallSystemPort, 0, len(ports))
	for _, port := range ports {
		families := []string{port.Family}
		if port.Family == "" {
			families = []string{constant.FirewallFamilyIPv4, constant.FirewallFamilyIPv6}
		}
		for _, family := range families {
			result = append(result, dto.FirewallSystemPort{Family: family, Port: port.Port, Protocol: port.Protocol})
		}
	}
	return result
}

func excludeFirewallPorts(ports, excluded []firewall.PortWhitelist) []firewall.PortWhitelist {
	return firewall.ExcludePorts(ports, excluded)
}

func AdoptLegacyHostFirewallRuleOwnership(ctx context.Context) error {
	return newFirewallService().adoptLegacyHostFirewallRuleOwnership(ctx)
}

func (s *FirewallService) adoptLegacyHostFirewallRuleOwnership(ctx context.Context) error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()

	selected, err := s.selectedProvider(ctx)
	if err != nil {
		return err
	}
	if selected != filter.ProviderIptables && selected != filter.ProviderUFW {
		return fmt.Errorf("%w: selected provider %s does not require legacy ownership transfer", filter.ErrProviderUnavailable, selected)
	}
	runtime, err := s.adapters.Resolve(selected)
	if err != nil {
		return err
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		return err
	}
	desiredByScope, failures := s.desiredFirewallRulesByScope(ctx, stored, selected)
	if len(failures) > 0 {
		return errors.New(failures[0].Error)
	}
	for _, scope := range filter.ManagedInputScopes(selected) {
		desired := desiredByScope[scope.Key()]
		if len(desired) == 0 {
			continue
		}
		snapshot, err := runtime.ObserveMutation(ctx, scope)
		if err != nil {
			return err
		}
		for {
			items, err := filter.MergeInventory(filter.InventoryMergeInput{
				Observed: snapshot.Rules,
				Desired:  desired,
			})
			if err != nil {
				return err
			}
			var candidate *filter.InventoryItem
			for index := range items {
				item := &items[index]
				if item.Match != filter.InventoryMatchChanged || item.Desired == nil || item.Observed == nil ||
					item.Desired.Origin != filter.RuleOriginAdopted || strings.TrimSpace(item.Desired.Marker) == "" ||
					strings.TrimSpace(item.Observed.Marker) != "" || item.Observed.Protected ||
					!filter.ObservedRuleMatchesExpected(*item.Observed, item.Desired.Rule) {
					continue
				}
				candidate = item
				break
			}
			if candidate == nil {
				break
			}
			after := candidate.Desired.Rule
			before := firewallsync.ObservedRule(*candidate.Observed)
			locator := candidate.Observed.Locator
			_, verification, err := runtime.Execute(ctx, snapshot, []filter.DesiredChange{{
				Operation:      filter.ChangeAdopt,
				Before:         &before,
				After:          &after,
				Locator:        &locator,
				PreviousMarker: candidate.Observed.Marker,
			}})
			if err != nil {
				return err
			}
			if !verification.Matched {
				return filter.ErrVerificationFailed
			}
			snapshot = verification.Snapshot
		}
	}
	if selected == filter.ProviderIptables {
		if err := iptables_helper.CleanupLegacyAdvancedChains(ctx); err != nil {
			return err
		}
	}
	return nil
}
