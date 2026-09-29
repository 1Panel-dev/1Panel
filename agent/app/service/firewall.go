package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	filterfirewalld "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/firewalld"
	filterufw "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/ufw"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FirewallService struct {
	rules                  repo.IFirewallRuleRepo
	adapters               map[filter.Provider]filter.Adapter
	forwardingSync         firewallDatabaseSyncAdapter
	dockerSync             firewallDatabaseSyncAdapter
	selectedProvider       func(context.Context) (filter.Provider, error)
	requiredPorts          func() ([]firewall.PortWhitelist, error)
	cleanupBackend         func(string) error
	cleanupInactiveBackend func(string) error
	resetBackend           func(string, bool) error
	dockerActive           func() (bool, error)
	restoreForwarding      func(context.Context) error
	restoreDockerGuard     func(context.Context) error
	baseClient             func() (lifecycle.Client, error)
}

var firewallRuleMutationMu sync.Mutex

var (
	firewallLifecycleTaskMu  sync.Mutex
	firewallLifecycleTaskID  string
	firewallLifecycleRequest dto.FirewallLifecycleOperation
)

type IFirewallService interface {
	SyncPortWhitelist(context.Context) error
	UpdatePanelPort(context.Context, uint, uint) error
	LoadBaseInfo(chainGroup string) (dto.FirewallSubsystemStatus, error)
	QueueFirewallOperation(request dto.FirewallLifecycleOperation) (dto.FirewallLifecycleOperationResponse, error)
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

type firewallRuleDeleteItem struct {
	index  int
	stored model.FirewallRule
	rules  []filter.DesiredRule
}

type preparedManagedUpdate struct {
	Stored   model.FirewallRule
	Before   filter.DesiredRule
	After    filter.FirewallRule
	RuleSet  filter.RuleSet
	Observed filter.ObservedRule
	Runtime  filter.Adapter
}

func (s *FirewallService) UpdatePanelPort(ctx context.Context, oldPort, port uint) error {
	if oldPort == 0 || oldPort > 65535 || port == 0 || port > 65535 {
		return fmt.Errorf("invalid panel port transition %d -> %d", oldPort, port)
	}
	if LoadPanelPort() != strconv.Itoa(int(oldPort)) {
		return fmt.Errorf("panel port changed before firewall update")
	}
	if oldPort == port {
		return nil
	}
	return s.updateSystemAccessPortWhitelist(ctx, firewall.PortWhitelistTypePanel, []string{strconv.Itoa(int(port))})
}

func (s *FirewallService) LoadBaseInfo(chainGroup string) (dto.FirewallSubsystemStatus, error) {
	status := dto.FirewallSubsystemStatus{Version: "-", Name: "-", Backend: "-"}
	status.LifecycleTaskID = currentFirewallLifecycleTaskID()
	selected, _ := settingRepo.GetValueByKey(constant.FirewallSystemBackendKey)
	if selected = strings.TrimSpace(selected); selected != "" {
		status.Name, status.Backend = selected, selected
	}
	loadClient := s.baseClient
	if loadClient == nil {
		loadClient = NewSelectedSystemFirewallClient
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
	status.Version, status.PingStatus = runtimeStatus.Version, firewall.LoadPingStatus()
	status.IsActive = runtimeStatus.IsActive
	if runtimeStatus.Name == constant.FirewallProviderIptables || runtimeStatus.Name == constant.FirewallProviderNftables {
		overview, err := loadSystemFirewallOverview(runtimeStatus.Name, chainGroup)
		if err != nil {
			return status, err
		}
		status.IsInit, status.IsBind = overview.IsInit, overview.IsBind
		status.IPv4, status.IPv6 = overview.IPv4, overview.IPv6
	}
	return status, nil
}

func (s *FirewallService) QueueFirewallOperation(request dto.FirewallLifecycleOperation) (dto.FirewallLifecycleOperationResponse, error) {
	response := dto.FirewallLifecycleOperationResponse{}
	if request.Operation == "disableBanPing" || request.Operation == "enableBanPing" {
		return response, s.OperateFirewall(request)
	}
	if !firewallLifecycleTaskMu.TryLock() {
		return response, buserr.New("TaskIsExecuting")
	}
	defer firewallLifecycleTaskMu.Unlock()
	if firewallLifecycleTaskID != "" {
		if request == firewallLifecycleRequest {
			return dto.FirewallLifecycleOperationResponse{TaskID: firewallLifecycleTaskID, Queued: true}, nil
		}
		return response, buserr.New("TaskIsExecuting")
	}
	running, err := s.CurrentRuleSyncTask()
	if err != nil {
		return response, err
	}
	if running.Executing {
		return response, buserr.New("TaskIsExecuting")
	}
	loadClient := s.baseClient
	if loadClient == nil {
		loadClient = NewSelectedSystemFirewallClient
	}
	client, err := loadClient()
	if err != nil {
		return response, err
	}
	if (client.Name() != lifecycle.ProviderFirewalld && client.Name() != lifecycle.ProviderUFW) ||
		(request.Operation != string(lifecycle.OperationStart) && request.Operation != string(lifecycle.OperationStop) &&
			request.Operation != string(lifecycle.OperationRestart)) {
		return response, s.OperateFirewall(request)
	}
	operation, label := task.TaskExec, "Start"
	switch request.Operation {
	case string(lifecycle.OperationStop):
		label = "Stop"
	case string(lifecycle.OperationRestart):
		operation, label = task.TaskRestart, task.TaskRestart
	}
	name := task.GetTaskName(client.Name(), label, task.TaskScopeFirewall)
	taskItem, err := task.NewTask(name, operation, task.TaskScopeFirewall, "", 0)
	if err != nil {
		return response, err
	}
	taskItem.AddSubTaskWithOps(name, func(t *task.Task) error {
		return s.runFirewallLifecycleTask(t, client, request)
	}, nil, 0, 0)
	if err := repo.NewITaskRepo().Save(context.Background(), taskItem.Task); err != nil {
		closeUnstartedFirewallTask(taskItem)
		return response, err
	}
	firewallLifecycleTaskID, firewallLifecycleRequest = taskItem.TaskID, request
	go func() {
		defer func() {
			closeUnstartedFirewallTask(taskItem)
			firewallLifecycleTaskMu.Lock()
			defer firewallLifecycleTaskMu.Unlock()
			if firewallLifecycleTaskID == taskItem.TaskID {
				firewallLifecycleTaskID = ""
			}
		}()
		if err := taskItem.Execute(); err != nil && taskItem.Task.Status == constant.StatusExecuting {
			taskItem.LogFailedWithErr(name, err)
			taskItem.Task.Status = constant.StatusFailed
			taskItem.Task.ErrorMsg = err.Error()
			taskItem.Task.EndAt = time.Now()
			_ = repo.NewITaskRepo().Update(context.Background(), taskItem.Task)
		}
	}()
	return dto.FirewallLifecycleOperationResponse{TaskID: taskItem.TaskID, Queued: true}, nil
}

func (s *FirewallService) OperateFilterChain(request dto.FilterChainOperation) error {
	client, err := NewSelectedSystemFirewallClient()
	if err != nil {
		return err
	}
	provider := client.Name()
	if err := s.operateFilterChainBase(provider, request); err != nil {
		return err
	}
	if request.Operate != string(firewall.BaseOperationInit) && request.Operate != string(firewall.BaseOperationBind) {
		return nil
	}
	ctx := context.Background()
	rulesErr := s.restoreStoredFirewallRules(ctx, filter.Provider(provider), nil)
	whitelistErr := s.SyncPortWhitelist(ctx)
	return errors.Join(rulesErr, whitelistErr)
}

func (s *FirewallService) QueueFilterChainInitialization(request dto.FilterChainOperation) (dto.FilterChainOperationResponse, error) {
	if request.Operate != string(firewall.BaseOperationInit) {
		return dto.FilterChainOperationResponse{}, fmt.Errorf("only filter chain initialization can be queued")
	}
	client, err := NewSelectedSystemFirewallClient()
	if err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	provider := client.Name()
	if provider != constant.FirewallProviderIptables && provider != constant.FirewallProviderNftables {
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
	taskItem.AddSubTask(i18n.GetMsgByKey("TaskSync"), func(t *task.Task) error {
		rulesErr := runFirewallLifecycleAction(t, i18n.GetWithName("FirewallRestoreRulesStep", provider), func() error {
			return s.restoreStoredFirewallRules(t.TaskCtx, filter.Provider(provider), t)
		})
		whitelistErr := runFirewallLifecycleAction(t, i18n.GetMsgByKey("FirewallSyncWhitelistStep"), func() error {
			return s.SyncPortWhitelist(t.TaskCtx)
		})
		return errors.Join(rulesErr, whitelistErr)
	}, nil)
	if err := repo.NewITaskRepo().Save(context.Background(), taskItem.Task); err != nil {
		return dto.FilterChainOperationResponse{}, fmt.Errorf("save firewall initialization task: %w", err)
	}
	go func() {
		_ = taskItem.Execute()
	}()
	return dto.FilterChainOperationResponse{TaskID: taskItem.TaskID, Queued: true}, nil
}

func (s *FirewallService) Reset(ctx context.Context, request dto.FirewallRuleReset) (dto.FirewallRuleResetResponse, error) {
	if err := lockFirewallLifecycleIdle(); err != nil {
		return dto.FirewallRuleResetResponse{}, err
	}
	defer firewallLifecycleTaskMu.Unlock()
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()

	selected, err := s.selectedProvider(ctx)
	if err != nil {
		return dto.FirewallRuleResetResponse{}, err
	}
	provider := request.Provider
	if provider == "" {
		provider = selected
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		return dto.FirewallRuleResetResponse{}, err
	}
	if selected == provider && len(stored) > 0 {
		if err := s.saveFirewallResetOrder(ctx, provider, stored); err != nil {
			return dto.FirewallRuleResetResponse{}, err
		}
	}
	if provider == filter.ProviderIptables || provider == filter.ProviderNftables {
		cleanup := s.cleanupBackend
		if (selected == filter.ProviderIptables || selected == filter.ProviderNftables) && selected != provider {
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
			dockerActive = firewallDockerActive
		}
		active, err := dockerActive()
		if err != nil {
			return dto.FirewallRuleResetResponse{}, fmt.Errorf("check Docker status before resetting firewalld: %w", err)
		}
		restartDocker = active
	}
	resetErr := reset(string(provider), restartDocker)
	if resetErr != nil {
		var dockerRestartErr *firewallDockerRestartError
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
	if len(scopes) == 1 && scopes[0].Provider == filter.ProviderUFW && scopes[0].Family == filter.FamilyInet && scopes[0].Table == "" &&
		scopes[0].Zone == "" && scopes[0].Chain == filter.UFWInputChain && scopes[0].Direction == filter.DirectionInput {
		scope := scopes[0]
		if err := s.checkSelectedProvider(ctx, scope.Provider); err != nil {
			return dto.FirewallRuleInventoryResponse{}, err
		}
		runtime, err := s.firewallAdapter(scope.Provider)
		if err != nil {
			return dto.FirewallRuleInventoryResponse{}, err
		}
		response, err := s.combinedUFWInventory(ctx, runtime, scope, request.Refresh)
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
	runtime, err := s.firewallAdapter(provider)
	if err != nil {
		return dto.FirewallRuleInventoryResponse{}, err
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		return dto.FirewallRuleInventoryResponse{}, err
	}
	desiredByScope, failures := s.desiredFirewallRulesByScope(ctx, stored, runtime)
	response := dto.FirewallRuleInventoryResponse{Items: failures}
	unavailable := make(map[filter.Family]error)
	for _, group := range firewallScopeReadGroups(scopes) {
		snapshots, err := readFirewallRuleScopes(runtime, ctx, group)
		if errors.Is(err, filter.ErrFamilyUnavailable) {
			for _, scope := range group {
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
			}
			continue
		}
		if err != nil {
			return dto.FirewallRuleInventoryResponse{}, err
		}
		for _, snapshot := range snapshots {
			scope := snapshot.Scope
			desired := desiredByScope[scope.Key()]
			items, err := mergeFirewallInventory(filter.InventoryMergeInput{Observed: snapshot.Rules, Desired: desired})
			if err != nil {
				return dto.FirewallRuleInventoryResponse{}, err
			}
			response.Items = append(response.Items, items...)
			response.Notices = append(response.Notices, snapshot.Notices...)
		}
	}
	return finalizeFirewallInventory(response, request), nil
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
	runtime, err := s.firewallAdapter(provider)
	if err != nil {
		return "", err
	}
	reader, ok := runtime.(filter.NativeDetailReader)
	if !ok {
		return "", fmt.Errorf("%w: native details for %s", filter.ErrAdapterUnavailable, runtime.Provider())
	}
	return reader.NativeDetail(ctx, request.Name, request.Permanent)
}

func (s *FirewallService) Adopt(ctx context.Context, request dto.FirewallRuleAdopt) error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	if err := s.checkSelectedProvider(ctx, request.Scope.Provider); err != nil {
		return err
	}
	runtime, err := s.firewallAdapter(request.Scope.Provider)
	if err != nil {
		return err
	}
	if runtime.Provider() != filter.ProviderNftables {
		if request.Rule == nil {
			return fmt.Errorf("%w: original rule is required for adoption", filter.ErrInvalidRule)
		}
		rule, err := filter.NormalizeRule(*request.Rule)
		if err != nil {
			return err
		}
		if rule.Scope.Key() != request.Scope.Key() {
			return fmt.Errorf("%w: adoption rule scope mismatch", filter.ErrInvalidScope)
		}
		observed := filter.ObservedRule{Rule: rule, Marker: request.Marker, ParseStatus: filter.ParseStatusSupported}
		return s.adoptRule(ctx, runtime, filter.RuleSet{Scope: rule.Scope}, observed, dto.FirewallRuleCreateItem{SourceKind: constant.FirewallRuleSourceUser})
	}
	if request.InstanceKey == "" {
		return fmt.Errorf("%w: instance key is required for nftables adoption", filter.ErrInvalidRule)
	}
	snapshot, err := readMutableFirewallRules(runtime, ctx, request.Scope)
	if err != nil {
		return err
	}
	observed, err := filter.FindCandidate(snapshot.Rules, request.InstanceKey)
	if err != nil {
		return filter.ErrRuleStale
	}
	return s.adoptRule(ctx, runtime, snapshot, observed, dto.FirewallRuleCreateItem{SourceKind: constant.FirewallRuleSourceUser})
}

func (s *FirewallService) Create(ctx context.Context, request dto.FirewallRuleCreate) (dto.FirewallRuleCreateResponse, error) {
	if len(request.Items) > filter.MaxAtomicExpansion {
		return dto.FirewallRuleCreateResponse{}, fmt.Errorf("create or import at most %d rules per batch (after expansion)", filter.MaxAtomicExpansion)
	}
	provider, err := s.selectedProvider(ctx)
	if err != nil {
		return dto.FirewallRuleCreateResponse{}, err
	}
	if err := validateFirewallCreateBatch(request, provider); err != nil {
		return dto.FirewallRuleCreateResponse{}, err
	}
	taskItem, err := task.NewTask(firewallTaskName(task.TaskCreate, firewallTaskHost, ""), task.TaskCreate, task.TaskScopeFirewall, "", 0)
	if err != nil {
		return dto.FirewallRuleCreateResponse{}, err
	}
	taskItem.AddSubTaskWithOps(i18n.GetMsgByKey("FirewallCreateRulesStep"), func(t *task.Task) error {
		firewallRuleMutationMu.Lock()
		defer firewallRuleMutationMu.Unlock()
		_, err := s.createRules(t.TaskCtx, request, t)
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

func (s *FirewallService) Delete(ctx context.Context, request dto.FirewallRuleDelete) (dto.FirewallRuleDeleteResponse, error) {
	if err := ctx.Err(); err != nil {
		return dto.FirewallRuleDeleteResponse{}, err
	}
	if len(request.UUIDs) == 0 && len(request.BeforeRules) == 0 {
		return dto.FirewallRuleDeleteResponse{}, fmt.Errorf("%w: rules are required", filter.ErrInvalidRule)
	}
	request.UUIDs = append([]string(nil), request.UUIDs...)
	request.BeforeRules = append([]dto.FirewallRuleDeleteTarget(nil), request.BeforeRules...)
	taskItem, err := task.NewTask(firewallTaskName(task.TaskDelete, firewallTaskHost, ""), task.TaskDelete, task.TaskScopeFirewall, "", 0)
	if err != nil {
		return dto.FirewallRuleDeleteResponse{}, err
	}
	taskItem.AddSubTaskWithOps(taskItem.Name, func(t *task.Task) error {
		t.Logf("rules=%d", len(request.UUIDs)+len(request.BeforeRules))
		firewallRuleMutationMu.Lock()
		defer firewallRuleMutationMu.Unlock()
		if err := t.TaskCtx.Err(); err != nil {
			return err
		}
		_, err := s.deleteRules(t.TaskCtx, request, t)
		return err
	}, nil, 0, 0)
	if err := repo.NewITaskRepo().Save(context.Background(), taskItem.Task); err != nil {
		taskItem.LogFailedWithErr(taskItem.Name, err)
		closeUnstartedFirewallTask(taskItem)
		return dto.FirewallRuleDeleteResponse{}, fmt.Errorf("save firewall deletion task: %w", err)
	}
	go func() {
		if err := taskItem.Execute(); err != nil && global.LOG != nil {
			global.LOG.Errorf("firewall deletion task %s failed: %v", taskItem.TaskID, err)
		}
	}()
	return dto.FirewallRuleDeleteResponse{TaskID: taskItem.TaskID, Queued: true}, nil
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

func (s *FirewallService) Reorder(ctx context.Context, clientIP string, request dto.FirewallRuleReorder) error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	return s.updateRuleOrder(ctx, request.UUID, request.TargetPosition, request.Priority, nil)
}

func NewIFirewallService() IFirewallService {
	return newFirewallService()
}

func currentFirewallLifecycleTaskID() string {
	firewallLifecycleTaskMu.Lock()
	defer firewallLifecycleTaskMu.Unlock()
	return firewallLifecycleTaskID
}

func (s *FirewallService) OperateFirewall(request dto.FirewallLifecycleOperation) error {
	switch request.Operation {
	case "disableBanPing":
		if err := firewall.UpdatePingStatus("0"); err != nil {
			return err
		}
		return settingRepo.Update(constant.FirewallPingStatusKey, constant.StatusDisable)
	case "enableBanPing":
		if err := firewall.UpdatePingStatus("1"); err != nil {
			return err
		}
		return settingRepo.Update(constant.FirewallPingStatusKey, constant.StatusEnable)
	}
	baseClient := s.baseClient
	if baseClient == nil {
		baseClient = NewSelectedSystemFirewallClient
	}
	client, err := baseClient()
	if err != nil {
		return err
	}
	operation := request.Operation
	operationErr := operateFirewallLifecycle(firewallLifecycleClient{client}, operation, request.WithDockerRestart, s.restoreFirewallAfterStart, nil)
	restoreFirewalld := client.Name() == lifecycle.ProviderFirewalld &&
		(operation == string(lifecycle.OperationStart) || operation == string(lifecycle.OperationRestart))
	if operation != string(lifecycle.OperationStart) && operation != string(lifecycle.OperationRestart) {
		return operationErr
	}
	if operationErr != nil {
		var completedErr *firewallCompletedOperationError
		var dockerRestartErr *firewallDockerRestartError
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

func (s *FirewallService) restoreFirewallAfterStart(client lifecycle.Client) error {
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
		isInit, _, err := loadFirewallInitStatus(string(provider), "base")
		if err != nil {
			recordFailure("load managed chain status", err)
			return errors.Join(recoveryErrors...)
		}
		if !isInit {
			return nil
		}
	}
	if err := s.restoreStoredFirewallRules(ctx, provider, nil); err != nil {
		recordFailure("restore stored firewall rules", err)
	}
	recordFailure("restore whitelist allowances", s.SyncPortWhitelist(ctx))
	return errors.Join(recoveryErrors...)
}

func (s *FirewallService) restoreFirewalldRuntimeDependents(ctx context.Context, operation string) error {
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
		dockerActive = firewallDockerActive
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

func (s *FirewallService) runFirewallLifecycleTask(t *task.Task, client lifecycle.Client, request dto.FirewallLifecycleOperation) error {
	ctx := t.TaskCtx
	provider := filter.Provider(client.Name())
	operationErr := operateFirewallLifecycle(firewallLifecycleClient{client}, request.Operation, request.WithDockerRestart, func(lifecycle.Client) error {
		rulesErr := runFirewallLifecycleAction(t, i18n.GetWithName("FirewallRestoreRulesStep", client.Name()), func() error {
			return s.restoreStoredFirewallRules(ctx, provider, t)
		})
		whitelistErr := runFirewallLifecycleAction(t, i18n.GetMsgByKey("FirewallSyncWhitelistStep"), func() error {
			return s.SyncPortWhitelist(ctx)
		})
		return errors.Join(rulesErr, whitelistErr)
	}, t)
	if request.Operation == string(lifecycle.OperationStop) {
		return operationErr
	}
	var recoveryErr *firewallCompletedOperationError
	if operationErr != nil && !errors.As(operationErr, &recoveryErr) {
		return operationErr
	}
	var forwardingErr error
	if provider == filter.ProviderFirewalld {
		forwardingErr = runFirewallLifecycleAction(t, i18n.GetMsgByKey("FirewallRestoreForwardingRulesStep"), func() error {
			if s.restoreForwarding != nil {
				return s.restoreForwarding(ctx)
			}
			return newForwardingService().Restore(ctx)
		})
	}
	dockerErr := runFirewallLifecycleAction(t, i18n.GetMsgByKey("FirewallInspectDockerGuardStep"), func() error {
		if provider == filter.ProviderFirewalld {
			active := s.dockerActive
			if active == nil {
				active = firewallDockerActive
			}
			running, err := active()
			if err != nil || !running {
				return err
			}
		}
		if s.restoreDockerGuard != nil {
			return s.restoreDockerGuard(ctx)
		}
		return ReconcileDockerPortGuard(ctx)
	})
	return errors.Join(operationErr, forwardingErr, dockerErr)
}

func (s *FirewallService) saveFirewallResetOrder(ctx context.Context, provider filter.Provider, stored []model.FirewallRule) error {
	runtime, err := s.firewallAdapter(provider)
	if err != nil {
		return err
	}
	desiredByScope := make(map[string][]filter.DesiredRule)
	byUUID := make(map[string]model.FirewallRule, len(stored))
	for _, record := range stored {
		if record.Origin != constant.FirewallRuleOriginCreated && record.Origin != constant.FirewallRuleOriginAdopted {
			continue
		}
		desired, err := compileStoredFirewallRules(ctx, record, runtime)
		if isFirewallPolicyIncompatible(err) {
			continue
		}
		if err != nil {
			return err
		}
		byUUID[record.UUID] = record
		for _, rule := range desired {
			key := rule.Rule.Scope.Key()
			desiredByScope[key] = append(desiredByScope[key], rule)
		}
	}
	captured := make(map[string]model.FirewallRule)
	for _, group := range firewallScopeReadGroups(filter.ManagedInputScopes(provider)) {
		snapshots, err := listFirewallRuleScopes(runtime, ctx, group)
		if errors.Is(err, filter.ErrFamilyUnavailable) {
			continue
		}
		if err != nil {
			return err
		}
		for _, snapshot := range snapshots {
			items, err := mergeFirewallInventory(filter.InventoryMergeInput{Observed: snapshot.Rules, Desired: desiredByScope[snapshot.Scope.Key()]})
			if err != nil {
				return err
			}
			for _, item := range items {
				if item.Desired == nil || item.Observed == nil {
					continue
				}
				record := byUUID[item.Desired.UUID]
				if provider == filter.ProviderFirewalld {
					record.Priority = item.Observed.Rule.Priority
					record.Sequence = nil
				} else {
					if item.Observed.Locator.Position == nil {
						return filter.ErrVerificationFailed
					}
					position := int64(*item.Observed.Locator.Position)
					if previous, exists := captured[record.UUID]; exists && *previous.Sequence <= position {
						continue
					}
					record.Sequence, record.Priority = &position, nil
				}
				captured[record.UUID] = record
			}
		}
	}
	orders := make([]model.FirewallRule, 0, len(captured))
	for _, record := range stored {
		if snapshot, exists := captured[record.UUID]; exists {
			orders = append(orders, snapshot)
		}
	}
	return s.rules.SaveResetOrder(ctx, orders)
}

func (s *FirewallService) combinedUFWInventory(ctx context.Context, runtime filter.Adapter, scope filter.Scope, refresh bool) (dto.FirewallRuleInventoryResponse, error) {
	scopes := []filter.Scope{scope, scope}
	scopes[0].Family = filter.FamilyIPv4
	scopes[1].Family = filter.FamilyIPv6
	snapshots, err := readFirewallRuleScopes(runtime, ctx, scopes)
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
	desiredByScope, failures := s.desiredFirewallRulesByScope(ctx, stored, runtime)
	response := dto.FirewallRuleInventoryResponse{Items: failures}
	seenNotices := make(map[string]struct{})
	for index, snapshot := range snapshots {
		if snapshot.Scope.Key() != scopes[index].Key() {
			return dto.FirewallRuleInventoryResponse{}, fmt.Errorf("%w: unexpected UFW inventory scope %q", filter.ErrInvalidScope, snapshot.Scope.Key())
		}
		desired := desiredByScope[snapshot.Scope.Key()]
		items, err := mergeFirewallInventory(filter.InventoryMergeInput{Observed: snapshot.Rules, Desired: desired})
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

func finalizeFirewallInventory(response dto.FirewallRuleInventoryResponse, request dto.FirewallRuleInventory) dto.FirewallRuleInventoryResponse {
	provider := request.Scope.Provider
	if len(request.Scopes) > 0 {
		provider = request.Scopes[0].Provider
	}
	response.IPv4Range, response.IPv6Range = firewallInventoryPositionRanges(provider, response.Items)
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
	response.Items = append([]filter.InventoryItem(nil), filtered[start:end]...)
	return response
}

func isDeletableManagedInventoryItem(item filter.InventoryItem) bool {
	if item.Desired == nil || item.Desired.Protected || item.State == filter.InventoryStateProtected {
		return false
	}
	if item.Desired.Origin != filter.RuleOriginCreated && item.Desired.Origin != filter.RuleOriginAdopted {
		return false
	}
	if (item.Rule.Scope.Provider == filter.ProviderIptables || item.Rule.Scope.Provider == filter.ProviderNftables) &&
		(item.Rule.Scope.Chain == filter.BasicBeforeChain || item.Rule.Scope.Chain == filter.BasicAfterChain) {
		return false
	}
	return item.State != filter.InventoryStateDrifted ||
		(item.Match == filter.InventoryMatchMissing && item.Observed == nil)
}

func matchesFirewallInventoryRequest(item filter.InventoryItem, request dto.FirewallRuleInventory) bool {
	if slices.Contains(request.ExcludeChains, item.Rule.Scope.Chain) {
		return false
	}
	if len(request.Families) > 0 && !matchesFirewallInventoryFamily(item.Rule, request.Families) {
		return false
	}
	if len(request.Actions) > 0 && !matchesFirewallInventoryAction(item.Rule.Action, request.Actions) {
		return false
	}
	if len(request.States) > 0 && !slices.Contains(request.States, item.State) {
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

func (s *FirewallService) deleteRules(ctx context.Context, request dto.FirewallRuleDelete, t *task.Task) (result dto.FirewallRuleDeleteResponse, taskErr error) {
	var firstFailure error
	defer func() {
		sort.SliceStable(result.Errors, func(i, j int) bool { return result.Errors[i].Index < result.Errors[j].Index })
		if t != nil {
			t.Log(i18n.GetMsgWithMap("FirewallRuleOperationResult", map[string]interface{}{
				"succeeded": result.Succeeded, "failed": result.Failed,
			}))
		}
		if taskErr == nil {
			taskErr = firstFailure
		}
	}()
	record := func(index int, ruleUUID string, err error) {
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, dto.FirewallRuleDeleteFailure{Index: index, UUID: ruleUUID, Error: err.Error()})
			if firstFailure == nil {
				firstFailure = err
			}
		} else {
			result.Succeeded++
		}
		if t != nil {
			label := fmt.Sprintf("[%d/%d] %s", result.Succeeded+result.Failed, len(request.UUIDs)+len(request.BeforeRules), ruleUUID)
			t.LogWithStatus(label, err)
		}
	}
	selectedProvider, err := s.selectedProvider(ctx)
	if err != nil {
		return result, err
	}
	type beforeGroup struct {
		targets []dto.FirewallRuleDeleteTarget
		indexes []int
	}
	beforeGroups := make(map[string]*beforeGroup)
	for index, target := range request.BeforeRules {
		key := target.Scope.Normalize().Key()
		if beforeGroups[key] == nil {
			beforeGroups[key] = &beforeGroup{}
		}
		group := beforeGroups[key]
		group.targets = append(group.targets, target)
		group.indexes = append(group.indexes, len(request.UUIDs)+index)
	}
	for _, group := range beforeGroups {
		failures, batchErr := s.deleteBeforeRules(ctx, selectedProvider, group.targets)
		for index, target := range group.targets {
			err := batchErr
			if failures != nil && failures[index] != nil {
				err = failures[index]
			}
			record(group.indexes[index], target.InstanceKey, err)
		}
	}
	if len(request.UUIDs) == 0 {
		return result, nil
	}
	ports, err := loadFirewallPortWhiteList()
	var runtime filter.Adapter
	if err == nil {
		runtime, err = s.firewallAdapter(selectedProvider)
	}
	if err != nil {
		for index, value := range request.UUIDs {
			record(index, strings.TrimSpace(value), err)
		}
		return result, err
	}
	whitelist := filter.NewPortWhitelistIndex(ports)
	groups := make([][]firewallRuleDeleteItem, 0)
	groupIndexes := make(map[string]int)
	seen := make(map[string]bool, len(request.UUIDs))
	for index, value := range request.UUIDs {
		ruleUUID := strings.TrimSpace(value)
		if err := ctx.Err(); err != nil {
			record(index, ruleUUID, err)
			continue
		}
		if ruleUUID == "" {
			record(index, ruleUUID, fmt.Errorf("%w: rule UUID is required", repo.ErrFirewallPersistenceInvalid))
			continue
		}
		if seen[ruleUUID] {
			record(index, ruleUUID, fmt.Errorf("duplicate firewall rule UUID"))
			continue
		}
		seen[ruleUUID] = true
		stored, err := s.rules.GetByUUID(ctx, ruleUUID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = fmt.Errorf("%w: managed rule %q was not found", filter.ErrInvalidRule, ruleUUID)
			}
			record(index, ruleUUID, err)
			continue
		}
		if stored.Origin != constant.FirewallRuleOriginCreated && stored.Origin != constant.FirewallRuleOriginAdopted {
			record(index, ruleUUID, fmt.Errorf("%w: only created or adopted rules can be deleted", filter.ErrInvalidRule))
			continue
		}
		rules, err := compileStoredFirewallRules(ctx, stored, runtime)
		if err == nil && len(rules) == 0 {
			err = fmt.Errorf("%w: policy %q has no compiled target rules", filter.ErrInvalidRule, ruleUUID)
		}
		if err == nil {
			for _, rule := range rules {
				if whitelist.Matches(rule.Rule) {
					err = filter.ErrProtectedRule
					break
				}
			}
		}
		if err != nil {
			record(index, ruleUUID, err)
			continue
		}
		key := rules[0].Rule.Scope.Key()
		if selectedProvider == filter.ProviderUFW {
			key = string(selectedProvider)
		}
		groupIndex, exists := groupIndexes[key]
		if !exists {
			groupIndex = len(groups)
			groupIndexes[key] = groupIndex
			groups = append(groups, nil)
		}
		groups[groupIndex] = append(groups[groupIndex], firewallRuleDeleteItem{index: index, stored: stored, rules: rules})
	}
	for _, group := range groups {
		failures, batchErr := s.deleteFirewallRuleBatch(ctx, runtime, group)
		for index, item := range group {
			err := batchErr
			if failures != nil && failures[index] != nil {
				err = failures[index]
			}
			record(item.index, item.stored.UUID, err)
		}
	}

	return result, nil
}

func (s *FirewallService) deleteBeforeRules(ctx context.Context, provider filter.Provider, targets []dto.FirewallRuleDeleteTarget) ([]error, error) {
	scope := targets[0].Scope.Normalize()
	if scope.Provider != provider || (provider != filter.ProviderIptables && provider != filter.ProviderNftables) || scope.Chain != filter.BasicBeforeChain {
		return nil, fmt.Errorf("%w: native deletion only supports the selected firewall before chain", filter.ErrUnsupportedScope)
	}
	runtime, err := s.firewallAdapter(provider)
	if err != nil {
		return nil, err
	}
	snapshot, err := readMutableFirewallRules(runtime, ctx, scope)
	if err != nil {
		return nil, err
	}
	failures := make([]error, len(targets))
	changes := make([]filter.RuleChange, 0, len(targets))
	seen := make(map[string]bool, len(targets))
	byInstance := make(map[string][]int, len(snapshot.Rules))
	for index, observed := range snapshot.Rules {
		key, err := filter.InstanceKey(observed)
		if err == nil {
			byInstance[key] = append(byInstance[key], index)
		}
	}
	for index, target := range targets {
		if target.Scope.Normalize().Key() != scope.Key() || seen[target.InstanceKey] {
			failures[index] = fmt.Errorf("%w: duplicate or mismatched before rule", filter.ErrInvalidRule)
			continue
		}
		seen[target.InstanceKey] = true
		matches := byInstance[target.InstanceKey]
		if len(matches) != 1 {
			failures[index] = filter.ErrRuleStale
			continue
		}
		observed := snapshot.Rules[matches[0]]
		if err := filter.GuardMutation(observed); err != nil {
			failures[index] = err
			continue
		}
		if observed.ParseStatus != filter.ParseStatusSupported || observed.Locator.Position == nil {
			failures[index] = fmt.Errorf("%w: before rule cannot be deleted", filter.ErrUnsupportedScope)
			continue
		}
		rule := observed.Rule
		if rule.UUID == "" && strings.HasPrefix(observed.Marker, "1panel-rule:") {
			rule.UUID = strings.TrimSpace(strings.TrimPrefix(observed.Marker, "1panel-rule:"))
		}
		if rule.UUID == "" {
			rule.UUID = uuid.NewString()
		}
		locator := observed.Locator
		changes = append(changes, filter.RuleChange{
			Operation: filter.ChangeDelete, Before: &rule, Locator: &locator, UnmarkedAdopted: observed.Marker == "", CommandOnly: true,
		})
	}
	if len(changes) == 0 {
		return failures, nil
	}
	sort.Slice(changes, func(i, j int) bool { return *changes[i].Locator.Position > *changes[j].Locator.Position })
	if err := ctx.Err(); err != nil {
		return failures, err
	}
	plan, err := runtime.BuildCommands(snapshot, changes)
	if err == nil {
		plan.CommandOnly = true
		err = runtime.RunCommands(ctx, plan)
	}
	if err == nil {
		err = persistFirewallRules(ctx, runtime, plan)
	}
	return failures, err
}

func (s *FirewallService) deleteFirewallRuleBatch(ctx context.Context, runtime filter.Adapter, items []firewallRuleDeleteItem) ([]error, error) {
	byScope := make(map[string]filter.RuleSet)
	byMarker := make(map[string]map[string][]filter.ObservedRule)
	if runtime.Provider() == filter.ProviderNftables {
		scopes := make([]filter.Scope, 0)
		for _, item := range items {
			for _, desired := range item.rules {
				scopes = append(scopes, desired.Rule.Scope)
			}
		}
		snapshots, err := readMutableFirewallRuleScopes(runtime, ctx, scopes)
		if err != nil {
			return nil, err
		}
		for _, snapshot := range snapshots {
			byScope[snapshot.Scope.Key()] = snapshot
			byMarker[snapshot.Scope.Key()] = firewallRulesByMarker(snapshot.Rules)
		}
	}
	failures := make([]error, len(items))
	changes := make([]firewallRuleBatchItem, 0, len(items))
	indexes := make([]int, 0, len(items))
	for index, item := range items {
		for _, desired := range item.rules {
			if desired.Protected {
				failures[index] = filter.ErrProtectedRule
				break
			}
			snapshot := filter.RuleSet{Scope: desired.Rule.Scope}
			change := filter.RuleChange{Operation: filter.ChangeDelete, Before: &desired.Rule, CommandOnly: true}
			if runtime.Provider() == filter.ProviderNftables {
				snapshot = byScope[desired.Rule.Scope.Key()]
				candidates := snapshot
				if desired.Marker != "" {
					candidates.Rules = byMarker[snapshot.Scope.Key()][desired.Marker]
				}
				observed, err := managedFirewallObserved(candidates, desired)
				if errors.Is(err, filter.ErrRuleStale) {
					missing, missingErr := managedFirewallRuleMissing(snapshot, desired)
					if missingErr != nil {
						err = missingErr
					} else if missing {
						continue
					}
				}
				if err == nil {
					change, err = firewallDeleteChange(observed, desired)
				}
				if err != nil {
					failures[index] = err
					break
				}
			}
			changes = append(changes, firewallRuleBatchItem{snapshot: snapshot, change: change})
			indexes = append(indexes, index)
		}
	}
	validChanges, validIndexes := changes[:0], indexes[:0]
	for index, change := range changes {
		if failures[indexes[index]] == nil {
			validChanges = append(validChanges, change)
			validIndexes = append(validIndexes, indexes[index])
		}
	}
	executeFirewallRuleBatches(ctx, runtime, validChanges, nil, func(index int, failure error) {
		if failure != nil {
			failures[validIndexes[index]] = failure
		}
	})
	deleted := make([]model.FirewallRule, 0, len(items))
	for index, item := range items {
		if failures[index] == nil {
			deleted = append(deleted, item.stored)
		}
	}
	persistFailures := s.rules.DeleteBatchWithRevision(context.WithoutCancel(ctx), deleted)
	for index, item := range items {
		if failures[index] == nil {
			failures[index] = persistFailures[item.stored.UUID]
		}
	}
	return failures, nil
}

func managedFirewallRuleMissing(snapshot filter.RuleSet, desired filter.DesiredRule) (bool, error) {
	wanted, err := filter.RuleKey(desired.Rule)
	if err != nil {
		return false, err
	}
	if desired.RuleKey != "" && desired.RuleKey != wanted {
		return false, filter.ErrInvalidRule
	}
	wanted, err = firewallInventoryRuleKey(desired.Rule)
	if err != nil {
		return false, err
	}
	for _, observed := range snapshot.Rules {
		if desired.Marker != "" {
			if observed.Marker == desired.Marker {
				return false, nil
			}
			adopted := desired.Origin == filter.RuleOriginAdopted && observed.Marker == ""
			legacy := observed.Marker == "1panel-rule:"+desired.UUID && observed.Marker != desired.Marker
			if (adopted || legacy) && filter.ObservedRuleMatchesExpected(observed, desired.Rule) {
				return false, nil
			}
			continue
		}
		if desired.ObservedInstanceKey != "" {
			key, err := filter.InstanceKey(observed)
			if err == nil && key == desired.ObservedInstanceKey {
				return false, nil
			}
			continue
		}
		if observed.ParseStatus != filter.ParseStatusSupported {
			continue
		}
		key, err := firewallInventoryRuleKey(observed.Rule)
		if err != nil {
			return false, err
		}
		if key == wanted {
			return false, nil
		}
	}
	return true, nil
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
	previousRules, compileErr := expandStoredFirewallRule(stored, requestedRule.Scope.Provider)
	if compileErr == nil && len(previousRules) == 1 {
		sameContent, err := filter.SameRuleContent(previousRules[0], requestedRule)
		if err != nil {
			return err
		}
		if sameContent {
			if requestedRule.Scope.Provider == filter.ProviderFirewalld {
				if requestedRule.Priority != nil {
					return s.updateRuleOrder(ctx, ruleUUID, nil, requestedRule.Priority, &requestedRule.Description)
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
	if prepared.Runtime.Provider() != filter.ProviderNftables {
		return s.replaceManagedRule(ctx, prepared)
	}
	return s.executeManagedMutation(ctx, managedMutationRequest{
		Stored: prepared.Stored, Before: prepared.Before.Rule, After: prepared.After,
		RuleSet: prepared.RuleSet, Locator: prepared.Observed.Locator,
		AdapterOperation: filter.ChangeUpdate, Runtime: prepared.Runtime,
	})
}

func (s *FirewallService) prepareManagedUpdate(ctx context.Context, clientIP string, ruleUUID string, requestedRule filter.FirewallRule) (preparedManagedUpdate, error) {
	ruleUUID = strings.TrimSpace(ruleUUID)
	if ruleUUID == "" {
		return preparedManagedUpdate{}, fmt.Errorf("%w: rule UUID is required", repo.ErrFirewallPersistenceInvalid)
	}
	stored, before, runtime, err := s.loadManagedRule(ctx, ruleUUID)
	if err != nil {
		return preparedManagedUpdate{}, err
	}
	after, err := filter.NormalizeRule(requestedRule)
	if err != nil {
		return preparedManagedUpdate{}, err
	}
	after.UUID = stored.UUID
	after, err = prepareFirewallBackendRule(ctx, runtime, after)
	if err != nil {
		return preparedManagedUpdate{}, err
	}
	if after.Scope.Key() != before.Rule.Scope.Key() {
		return preparedManagedUpdate{}, buserr.New("ErrFirewallRuleScopeChange")
	}
	if !supportsManagedNativeKindTransition(before.Rule, after) {
		return preparedManagedUpdate{}, fmt.Errorf("%w: native rule conversion requires an explicit workflow", filter.ErrUnsupportedScope)
	}
	if runtime.Provider() != filter.ProviderNftables {
		if after.OrderIndex != nil && *after.OrderIndex < 1 {
			return preparedManagedUpdate{}, fmt.Errorf("%w: target position must be positive", filter.ErrInvalidRule)
		}
		return preparedManagedUpdate{Stored: stored, Before: before, After: after, Runtime: runtime}, nil
	}
	snapshot, err := readMutableFirewallRules(runtime, ctx, before.Rule.Scope)
	if err != nil {
		return preparedManagedUpdate{}, err
	}
	observed, err := managedFirewallObserved(snapshot, before)
	if err != nil {
		return preparedManagedUpdate{}, err
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
			if err := validateFirewallRulePosition(snapshot, before.Rule, *after.OrderIndex); err != nil {
				return preparedManagedUpdate{}, err
			}
		}
	}
	if err := filter.GuardMutation(observed); err != nil {
		return preparedManagedUpdate{}, err
	}
	return preparedManagedUpdate{
		Stored: stored, Before: before, After: after, RuleSet: snapshot, Observed: observed, Runtime: runtime,
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

func (s *FirewallService) ensureWebsitePorts(ctx context.Context, ports map[string]firewall.SystemPort) error {
	if len(ports) == 0 {
		return nil
	}
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()

	provider, err := s.selectedProvider(ctx)
	if err != nil {
		return err
	}
	runtime, err := s.firewallAdapter(provider)
	if err != nil {
		return err
	}
	const comment = "1panel website"
	rules := make([]filter.FirewallRule, 0, len(ports))
	var scopes []filter.Scope
	seenScopes := make(map[string]bool)
	for _, key := range firewall.SortedSystemPortKeys(ports) {
		prepared, err := prepareFirewallCreateRule(ctx, runtime, dto.FirewallRuleCreateItem{Rule: firewall.RuleForSystemPort(provider, ports[key])})
		if err != nil {
			return err
		}
		rule := prepared.Rule
		rule.Description = comment
		rules = append(rules, rule)
		if !seenScopes[rule.Scope.Key()] {
			seenScopes[rule.Scope.Key()] = true
			scopes = append(scopes, rule.Scope)
		}
	}
	external, supportsExternal := runtime.(filter.ExternalRuleAdapter)
	existing := make(map[string]bool)
	if provider != filter.ProviderFirewalld {
		if !supportsExternal {
			return fmt.Errorf("%w: %s does not support external rules", filter.ErrAdapterUnavailable, provider)
		}
		observed, err := external.ListRulesByComment(ctx, scopes, comment)
		if err != nil {
			return err
		}
		for _, candidate := range observed {
			if candidate.ParseStatus != filter.ParseStatusSupported || candidate.Rule.Description != comment || candidate.Rule.Action != filter.ActionAccept {
				continue
			}
			key, err := filter.RuleMatchKey(candidate.Rule)
			if err != nil {
				return err
			}
			existing[key] = true
		}
	}
	var failures []error
	changedScopes := make(map[string]filter.Scope)
	for _, rule := range rules {
		key, err := filter.RuleMatchKey(rule)
		if err != nil {
			return err
		}
		if existing[key] {
			continue
		}
		if provider == filter.ProviderFirewalld {
			rule.UUID = uuid.NewString()
			plan, buildErr := runtime.BuildCommands(filter.RuleSet{Scope: rule.Scope}, []filter.RuleChange{{Operation: filter.ChangeCreate, After: &rule, CommandOnly: true, Append: true}})
			if buildErr != nil {
				return buildErr
			}
			plan.CommandOnly = true
			err = runtime.RunCommands(ctx, plan)
			if errors.Is(err, filterfirewalld.ErrAlreadyEnabled) {
				err = nil
			}
		} else {
			err = external.AppendUnverified(ctx, rule, comment)
		}
		if rule.Scope.Family == filter.FamilyIPv6 && filterufw.IsIPv6Unavailable(err) {
			continue
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("allow website port %s/%s (%s): %w", rule.DestinationPort, rule.Protocol, rule.Scope.Family, err))
			continue
		}
		existing[key] = true
		saveKey := rule.Scope.Key()
		if provider == filter.ProviderNftables {
			saveKey = string(provider)
		}
		changedScopes[saveKey] = rule.Scope
	}
	if saver, ok := runtime.(filter.RuleSaver); ok {
		for _, scope := range changedScopes {
			saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
			err := saver.SaveRules(saveCtx, scope)
			cancel()
			if err != nil {
				failures = append(failures, err)
			}
		}
	}
	return errors.Join(failures...)
}

func ensureFirewallPorts(ports []int) error {
	if len(ports) == 0 {
		return nil
	}
	client, err := NewSelectedSystemFirewallClient()
	if err != nil {
		return err
	}
	state, err := lifecycle.LoadState(client)
	if err != nil {
		return err
	}
	if state.Name == constant.FirewallProviderIptables || state.Name == constant.FirewallProviderNftables {
		isInit, _, err := loadFirewallInitStatus(state.Name, "base")
		if err != nil {
			return err
		}
		if !isInit {
			return nil
		}
	} else if !state.IsActive {
		return nil
	}

	added := make([]firewall.PortWhitelist, 0, len(ports))
	for _, port := range ports {
		added = append(added, firewall.PortWhitelist{Port: strconv.Itoa(port), Protocol: "tcp"})
	}

	normalized, err := firewall.NormalizeSystemPorts(firewall.ExpandPortWhitelist(added))
	if err != nil {
		return err
	}
	return newFirewallService().ensureWebsitePorts(context.Background(), normalized)
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
	runtime, err := s.firewallAdapter(selected)
	if err != nil {
		return err
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		return err
	}
	desiredByScope, failures := s.desiredFirewallRulesByScope(ctx, stored, runtime)
	if len(failures) > 0 {
		return errors.New(failures[0].Error)
	}
	for _, scope := range filter.ManagedInputScopes(selected) {
		desired := desiredByScope[scope.Key()]
		if len(desired) == 0 {
			continue
		}
		snapshot, err := readMutableFirewallRules(runtime, ctx, scope)
		if err != nil {
			return err
		}
		items, err := mergeFirewallInventory(filter.InventoryMergeInput{Observed: snapshot.Rules, Desired: desired})
		if err != nil {
			return err
		}
		for _, item := range items {
			if item.Match != filter.InventoryMatchChanged || item.Desired == nil || item.Observed == nil ||
				item.Desired.Origin != filter.RuleOriginAdopted || strings.TrimSpace(item.Desired.Marker) == "" ||
				strings.TrimSpace(item.Observed.Marker) != "" || item.Observed.Protected ||
				!filter.ObservedRuleMatchesExpected(*item.Observed, item.Desired.Rule) {
				continue
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			matches := make([]filter.ObservedRule, 0, 1)
			for _, observed := range snapshot.Rules {
				if observed.Marker != "" || observed.Rule.SourceAddress != item.Observed.Rule.SourceAddress ||
					observed.Rule.DestinationAddress != item.Observed.Rule.DestinationAddress ||
					observed.Rule.SourcePort != item.Observed.Rule.SourcePort || observed.Rule.DestinationPort != item.Observed.Rule.DestinationPort {
					continue
				}
				if filter.ObservedRuleMatchesExpected(observed, item.Desired.Rule) {
					matches = append(matches, observed)
				}
			}
			if len(matches) != 1 {
				return filter.ErrRuleStale
			}
			after := item.Desired.Rule
			before := matches[0].Rule
			locator := matches[0].Locator
			_, verification, err := applyFirewallChanges(runtime, ctx, snapshot, []filter.RuleChange{{
				Operation: filter.ChangeAdopt, Before: &before, After: &after, Locator: &locator, PreviousMarker: matches[0].Marker,
			}})
			if err != nil {
				return err
			}
			if !verification.Matched {
				return filter.ErrVerificationFailed
			}
			snapshot = verification.RuleSet
		}
	}
	if selected == filter.ProviderIptables {
		if err := iptables_helper.CleanupLegacyAdvancedChains(ctx); err != nil {
			return err
		}
	}
	return nil
}
