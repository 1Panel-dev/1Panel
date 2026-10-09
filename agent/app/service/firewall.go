package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
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
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	lifecycleproviders "github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle/providers"
	"github.com/google/uuid"
)

type FirewallService struct {
	adapters               map[filter.Provider]filter.Adapter
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
	ListRuleBackups(context.Context, string) (dto.FirewallRuleBackups, error)
	Reset(context.Context, dto.FirewallRuleReset) (dto.FirewallRuleResetResponse, error)
	ExportBackup(context.Context, filter.Provider) ([]dto.FirewallRuleExportItem, int, error)
	Inventory(context.Context, dto.FirewallRuleInventory) (dto.FirewallRuleInventoryResponse, error)
	LoadFirewallNativeDetail(context.Context, dto.FirewallNativeDetail) (string, error)
	Create(context.Context, dto.FirewallRuleCreate) (dto.FirewallRuleCreateResponse, error)
	Delete(context.Context, dto.FirewallRuleDelete) (dto.FirewallRuleDeleteResponse, error)
	Update(context.Context, dto.FirewallRuleUpdate) error
	Reorder(context.Context, dto.FirewallRuleReorder) error
}

func (s *FirewallService) SyncPortWhitelist(ctx context.Context) error {
	return s.syncPortWhitelist(ctx, nil)
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
	families, err := loadFirewallFamilies()
	if err != nil {
		return dto.FirewallSubsystemStatus{}, err
	}
	status := dto.FirewallSubsystemStatus{Version: "-", Name: "-", Backend: "-", IPv6Enabled: len(families) > 1}
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
		return status, err
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
		overview, err := loadSystemFirewallOverview(runtimeStatus.Name, chainGroup, families)
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
	if err := task.CheckScopeTaskIsExecuting(task.TaskScopeFirewall, 0); err != nil {
		return response, err
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
		return s.runFirewallLifecycleTask(t, client, request, true)
	}, nil, 0, 0)
	if err := taskRepo.Save(context.Background(), taskItem.Task); err != nil {
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
			_ = taskRepo.Update(context.Background(), taskItem.Task)
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
	return s.SyncPortWhitelist(ctx)
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
		families, err := loadFirewallFamilies()
		if err != nil {
			return err
		}
		if !slices.Contains(families, constant.FirewallFamilyIPv6) {
			t.Logf("IPv6 firewall support is disabled; skipping IPv6 host chain initialization")
		}
		return s.operateFilterChainBase(provider, request)
	}, nil)
	taskItem.AddSubTask(i18n.GetMsgByKey("TaskSync"), func(t *task.Task) error {
		whitelistErr := runFirewallLifecycleAction(t, i18n.GetMsgByKey("FirewallSyncWhitelistStep"), func() error {
			return s.SyncPortWhitelist(t.TaskCtx)
		})
		return whitelistErr
	}, nil)
	if err := taskRepo.Save(context.Background(), taskItem.Task); err != nil {
		return dto.FilterChainOperationResponse{}, fmt.Errorf("save firewall initialization task: %w", err)
	}
	go func() {
		_ = taskItem.Execute()
	}()
	return dto.FilterChainOperationResponse{TaskID: taskItem.TaskID, Queued: true}, nil
}

func (s *FirewallService) ListRuleBackups(ctx context.Context, subsystem string) (dto.FirewallRuleBackups, error) {
	if subsystem == "" {
		subsystem = "system"
	}
	directory, err := firewallBackupDirectory(subsystem)
	if err != nil {
		return dto.FirewallRuleBackups{}, err
	}
	result := dto.FirewallRuleBackups{Directory: directory, Files: make([]dto.FirewallRuleBackup, 0)}
	entries, err := os.ReadDir(directory)
	if errors.Is(err, os.ErrNotExist) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		if !entry.Type().IsRegular() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		var provider filter.Provider
		count := 0
		if subsystem == "system" {
			items, err := readFirewallBackup(entry.Name())
			if err != nil || len(items) == 0 {
				continue
			}
			provider, count = items[0].Scope.Provider, len(items)
		} else {
			backup, err := readFirewallSubsystemBackup(entry.Name(), subsystem)
			if err != nil {
				continue
			}
			provider, count = backup.Provider, len(backup.Forwarding)
			if backup.Docker != nil {
				count = len(backup.Docker.Policies)
			}
		}
		info, err := entry.Info()
		if err != nil {
			return result, err
		}
		result.Files = append(result.Files, dto.FirewallRuleBackup{Name: entry.Name(), Provider: provider, RuleCount: count, ModifiedAt: info.ModTime().UnixMilli()})
	}
	slices.SortFunc(result.Files, func(a, b dto.FirewallRuleBackup) int {
		if a.ModifiedAt > b.ModifiedAt {
			return -1
		}
		if a.ModifiedAt < b.ModifiedAt {
			return 1
		}
		return strings.Compare(b.Name, a.Name)
	})
	return result, nil
}

func (s *FirewallService) Reset(ctx context.Context, request dto.FirewallRuleReset) (dto.FirewallRuleResetResponse, error) {
	if request.Subsystem != "" && request.Subsystem != "system" {
		return s.resetSubsystemRules(ctx, request)
	}
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
	var backup string
	var count int
	if request.Backup == nil || *request.Backup {
		directory, directoryErr := firewallBackupDirectory("system")
		if directoryErr != nil {
			return dto.FirewallRuleResetResponse{}, directoryErr
		}
		var items []dto.FirewallRuleExportItem
		items, count, err = s.ExportBackup(ctx, provider)
		if err == nil {
			backup, err = writeFirewallJSON(directory, string(provider)+"-"+time.Now().Format("20060102-150405")+"-"+uuid.NewString()+".json", items)
		}
	} else {
		var inventory dto.FirewallRuleInventoryResponse
		inventory, err = s.readFirewallInventory(ctx, provider, filter.ManagedInputScopes(provider))
		count = len(inventory.Items)
	}
	if err != nil {
		return dto.FirewallRuleResetResponse{}, err
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
		return dto.FirewallRuleResetResponse{Removed: count, Disabled: true, BackupPath: backup}, nil
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
			restoreDockerGuard = RestoreDockerPortGuard
		}
		restoreErr := restoreFirewalldDependents(
			ctx, "after resetting firewalld", restartDocker, restoreForwarding, restoreDockerGuard,
		)
		if err := errors.Join(resetErr, restoreErr); err != nil {
			return dto.FirewallRuleResetResponse{}, err
		}
	}
	return dto.FirewallRuleResetResponse{Removed: count, Disabled: true, BackupPath: backup}, nil
}

func (s *FirewallService) ExportBackup(ctx context.Context, provider filter.Provider) ([]dto.FirewallRuleExportItem, int, error) {
	if provider == "" {
		var err error
		provider, err = s.selectedProvider(ctx)
		if err != nil {
			return nil, 0, err
		}
	}
	response, err := s.readFirewallInventory(ctx, provider, filter.ManagedInputScopes(provider))
	if err != nil {
		return nil, 0, err
	}
	descriptions, err := settingRepo.GetDescriptionList(repo.WithByType("firewall"))
	if err != nil {
		return nil, 0, err
	}
	byID := make(map[string]string, len(descriptions))
	for _, description := range descriptions {
		byID[description.ID] = description.Description
	}
	items := make([]dto.FirewallRuleExportItem, 0, len(response.Items))
	for _, item := range response.Items {
		if (provider == filter.ProviderIptables || provider == filter.ProviderNftables) && item.Rule.Scope.Chain != filter.IptablesInputChain {
			continue
		}
		id, err := filter.DescriptionID(*item.Observed)
		if err != nil {
			return nil, 0, err
		}
		rule := item.Rule
		if description, ok := byID[id]; ok {
			rule.Description = description
		}
		rule.UUID, rule.OrderIndex = "", nil
		items = append(items, dto.FirewallRuleExportItem{FirewallRule: rule, Raw: item.Observed.Raw, ParseStatus: item.Observed.ParseStatus})
	}
	return items, len(response.Items), nil
}

func (s *FirewallService) Inventory(ctx context.Context, request dto.FirewallRuleInventory) (dto.FirewallRuleInventoryResponse, error) {
	scopes := request.Scopes
	if len(scopes) == 0 && request.Scope.Provider != "" {
		scopes = []filter.Scope{request.Scope}
	}
	if len(scopes) == 0 {
		return dto.FirewallRuleInventoryResponse{}, filter.ErrInvalidScope
	}
	provider := scopes[0].Provider
	if err := s.checkSelectedProvider(ctx, provider); err != nil {
		return dto.FirewallRuleInventoryResponse{}, err
	}
	if len(scopes) == 1 && provider == filter.ProviderUFW && scopes[0].Family == filter.FamilyInet {
		scopes = filter.ManagedInputScopes(provider)
	}
	response, err := s.readFirewallInventory(ctx, provider, scopes)
	if err != nil {
		return response, err
	}
	searchDescription := strings.TrimSpace(request.Info) != ""
	if !searchDescription {
		response = finalizeFirewallInventory(response, request)
	}
	ids := make([]string, 0, len(response.Items))
	for index := range response.Items {
		item := &response.Items[index]
		item.DescriptionID, err = filter.DescriptionID(*item.Observed)
		if err != nil {
			return response, err
		}
		ids = append(ids, item.DescriptionID)
	}
	byID := make(map[string]string, len(ids))
	for start := 0; start < len(ids); start += 500 {
		descriptions, err := settingRepo.GetDescriptionList(repo.WithByType("firewall"), settingRepo.WithDescriptionIDs(ids[start:min(start+500, len(ids))]))
		if err != nil {
			return response, err
		}
		for _, description := range descriptions {
			byID[description.ID] = description.Description
		}
	}
	for index := range response.Items {
		item := &response.Items[index]
		if description, ok := byID[item.DescriptionID]; ok {
			item.Rule.Description = description
		}
	}
	if searchDescription {
		response = finalizeFirewallInventory(response, request)
	}
	ports, err := loadFirewallPortWhiteList()
	if err != nil {
		return response, err
	}
	whitelist := filter.NewPortWhitelistIndex(ports)
	for index := range response.Items {
		item := &response.Items[index]
		item.IsWhitelist = item.Observed.ParseStatus == filter.ParseStatusSupported && whitelist.Matches(item.Rule)
		item.Observed.Protected = item.Observed.Protected || item.IsWhitelist
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

func (s *FirewallService) Create(ctx context.Context, request dto.FirewallRuleCreate) (dto.FirewallRuleCreateResponse, error) {
	if request.BackupFile != "" {
		if len(request.Items) != 0 {
			return dto.FirewallRuleCreateResponse{}, filter.ErrInvalidRule
		}
		items, err := readFirewallBackup(request.BackupFile)
		if err != nil {
			return dto.FirewallRuleCreateResponse{}, err
		}
		for _, item := range items {
			request.Items = append(request.Items, dto.FirewallRuleCreateItem{Rule: item.FirewallRule, Raw: item.Raw, ParseStatus: item.ParseStatus, SourceKind: constant.FirewallRuleSourceImported})
		}
	}

	for index, item := range request.Items {
		if item.SourceKind != constant.FirewallRuleSourceImported || item.Rule.Scope.Provider != filter.ProviderUFW || item.ParseStatus != filter.ParseStatusPartial || item.Raw == "" {
			continue
		}
		observed := filterufw.ParseRules(item.Rule.Scope, item.Raw)
		if len(observed) == 1 && observed[0].ParseStatus == filter.ParseStatusSupported {
			observed[0].Rule.Description = item.Rule.Description
			request.Items[index].Rule = observed[0].Rule
			request.Items[index].ParseStatus, request.Items[index].Raw = filter.ParseStatusSupported, ""
		}
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
	if request.Initialize {
		taskItem.AddSubTask(i18n.GetWithName("FirewallInitializeChainsStep", string(provider)), func(t *task.Task) error {
			if err := s.checkSelectedProvider(t.TaskCtx, provider); err != nil {
				return err
			}
			families, err := loadFirewallFamilies()
			if err != nil {
				return err
			}
			if !slices.Contains(families, constant.FirewallFamilyIPv6) {
				t.Logf("IPv6 firewall support is disabled; skipping IPv6 host chain initialization")
			}
			if provider == filter.ProviderIptables || provider == filter.ProviderNftables {
				return s.operateFilterChainBase(string(provider), dto.FilterChainOperation{Name: "1PANEL_BASIC", Operate: string(firewall.BaseOperationInit)})
			}
			client, err := s.baseClient()
			if err != nil {
				return err
			}
			return s.runFirewallLifecycleTask(t, client, dto.FirewallLifecycleOperation{Operation: "start"}, false)
		}, nil)
	}
	ruleState := make(map[string]filter.RuleSet)
	taskItem.AddSubTaskWithOps(i18n.GetMsgByKey("FirewallCreateRulesStep"), func(t *task.Task) error {
		firewallRuleMutationMu.Lock()
		if err := s.checkSelectedProvider(t.TaskCtx, provider); err != nil {
			firewallRuleMutationMu.Unlock()
			return err
		}
		_, err := s.createRules(t.TaskCtx, request, t, ruleState)
		firewallRuleMutationMu.Unlock()
		if !request.Initialize || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		if err != nil {
			ruleState = nil
		}
		whitelistErr := s.syncPortWhitelist(t.TaskCtx, ruleState)
		if whitelistErr == nil && provider == filter.ProviderFirewalld {
			whitelistErr = lifecycleproviders.RemoveFirewalldSSHService()
		}
		t.LogWithStatus(i18n.GetMsgByKey("FirewallSyncWhitelistStep"), whitelistErr)
		return errors.Join(err, whitelistErr)
	}, nil, 0, 0)
	if err := taskRepo.Save(context.Background(), taskItem.Task); err != nil {
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
	if len(request.Targets) == 0 {
		return dto.FirewallRuleDeleteResponse{}, fmt.Errorf("%w: rules are required", filter.ErrInvalidRule)
	}
	request.Targets = append([]dto.FirewallRuleDeleteItem(nil), request.Targets...)
	taskItem, err := task.NewTask(firewallTaskName(task.TaskDelete, firewallTaskHost, ""), task.TaskDelete, task.TaskScopeFirewall, "", 0)
	if err != nil {
		return dto.FirewallRuleDeleteResponse{}, err
	}
	taskItem.AddSubTaskWithOps(taskItem.Name, func(t *task.Task) error {
		t.Logf("rules=%d", len(request.Targets))
		firewallRuleMutationMu.Lock()
		defer firewallRuleMutationMu.Unlock()
		if err := t.TaskCtx.Err(); err != nil {
			return err
		}
		_, err := s.deleteRules(t.TaskCtx, request, t)
		return err
	}, nil, 0, 0)
	if err := taskRepo.Save(context.Background(), taskItem.Task); err != nil {
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

func (s *FirewallService) Update(ctx context.Context, request dto.FirewallRuleUpdate) error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	return s.updateObservedRule(ctx, request)
}

func (s *FirewallService) Reorder(ctx context.Context, request dto.FirewallRuleReorder) error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	return s.updateObservedRule(ctx, dto.FirewallRuleUpdate{FirewallRuleDeleteTarget: request.FirewallRuleDeleteTarget, OrderIndex: request.TargetPosition, Priority: request.Priority})
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
	RestoreDockerPortGuardBestEffort(context.Background())
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
		restoreDockerGuard = RestoreDockerPortGuard
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

func (s *FirewallService) runFirewallLifecycleTask(t *task.Task, client lifecycle.Client, request dto.FirewallLifecycleOperation, syncWhitelist bool) error {
	ctx := t.TaskCtx
	provider := filter.Provider(client.Name())
	var prepareStart func(lifecycle.Client) error
	if syncWhitelist {
		prepareStart = func(lifecycle.Client) error {
			return runFirewallLifecycleAction(t, i18n.GetMsgByKey("FirewallSyncWhitelistStep"), func() error {
				return s.SyncPortWhitelist(ctx)
			})
		}
	}
	operationErr := operateFirewallLifecycle(firewallLifecycleClient{client}, request.Operation, request.WithDockerRestart, prepareStart, t)
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
		return RestoreDockerPortGuard(ctx)
	})
	return errors.Join(operationErr, forwardingErr, dockerErr)
}

func finalizeFirewallInventory(response dto.FirewallRuleInventoryResponse, request dto.FirewallRuleInventory) dto.FirewallRuleInventoryResponse {
	provider := request.Scope.Provider
	if len(request.Scopes) > 0 {
		provider = request.Scopes[0].Provider
	}
	chain := filter.IptablesInputChain
	if len(request.Scopes) == 1 {
		chain = request.Scopes[0].Normalize().Chain
	} else if len(request.Scopes) == 0 {
		chain = request.Scope.Normalize().Chain
	}
	response.IPv4Range, response.IPv6Range = firewallInventoryPositionRanges(provider, chain, response.Items)
	response.AllTotal = 0
	filtered := make([]filter.InventoryItem, 0, len(response.Items))
	for _, item := range response.Items {
		if item.Observed != nil && item.Observed.Protected {
			continue
		}
		response.AllTotal++
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
	keyword := strings.ToLower(strings.TrimSpace(request.Info))
	if keyword == "" {
		return true
	}
	rule := item.Rule
	values := []string{
		firewallInventoryProtocol(rule), rule.SourceAddress, rule.SourcePort, rule.DestinationAddress,
		rule.DestinationPort, rule.Description, string(rule.Action),
	}
	if item.Observed != nil {
		values = append(values, item.Observed.Rule.Description)
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

func (s *FirewallService) deleteRules(ctx context.Context, request dto.FirewallRuleDelete, t *task.Task) (dto.FirewallRuleDeleteResponse, error) {
	result := dto.FirewallRuleDeleteResponse{}
	provider, err := s.selectedProvider(ctx)
	if err != nil {
		return result, err
	}
	runtime, err := s.firewallAdapter(provider)
	if err != nil {
		return result, err
	}
	ports, err := loadFirewallPortWhiteList()
	if err != nil {
		return result, err
	}
	whitelist := filter.NewPortWhitelistIndex(ports)
	plans := make([]filter.CommandBatch, len(request.Targets))
	groups := make([][]int, 0)
	byScope := make(map[string]int)
	seen := make(map[string]bool)
	var failures []error
	record := func(index int, err error) {
		label := fmt.Sprintf("[%d/%d] %s", index+1, len(request.Targets), request.Targets[index].Scope.Key())
		if errors.Is(err, filter.ErrRuleNotFound) {
			if t != nil {
				t.Logf("%s: %v", label, err)
			} else if global.LOG != nil {
				global.LOG.Infof("%s: %v", label, err)
			}
			err = nil
		}
		if err == nil {
			result.Succeeded++
		} else {
			result.Failed++
			result.Errors = append(result.Errors, dto.FirewallRuleDeleteFailure{Index: index, InstanceKey: request.Targets[index].InstanceKey, Error: err.Error()})
			failures = append(failures, err)
		}
		if t != nil {
			t.LogWithStatus(label, err)
		}
	}
	for index, target := range request.Targets {
		if seen[target.InstanceKey] {
			continue
		}
		seen[target.InstanceKey] = true
		scope := target.Scope.Normalize()
		if scope.Provider != provider || target.Observed.Rule.Scope.Key() != scope.Key() {
			record(index, filter.ErrInvalidScope)
			continue
		}
		plan, err := runtime.BuildCommands(filter.RuleSet{Scope: scope}, []filter.RuleChange{{Operation: filter.ChangeDelete, Target: &target.Observed, CommandOnly: true}})
		if err != nil {
			record(index, err)
			continue
		}
		if plan.Rules[0].Expected.Protected || whitelist.Matches(plan.Rules[0].Expected.Rule) {
			record(index, filter.ErrProtectedRule)
			continue
		}
		plan.CommandOnly = true
		plans[index] = plan
		key := scope.Key()
		if provider == filter.ProviderFirewalld {
			key += ":" + string(plan.Rules[0].Expected.Rule.NativeKind)
		}
		group, exists := byScope[key]
		if !exists {
			group = len(groups)
			byScope[key] = group
			groups = append(groups, nil)
		}
		groups[group] = append(groups[group], index)
	}
	var attempted []filter.CommandBatch
	var indexes []int
	var results []error
	for _, group := range groups {
		for start := 0; start < len(group); {
			end := start + 1
			if provider != filter.ProviderUFW {
				end = min(start+filter.MaxAtomicExpansion, len(group))
			}
			batch := group[start:end]
			start = end
			plan := plans[batch[0]]
			err := ctx.Err()
			if err == nil && len(batch) > 1 {
				changes := make([]filter.RuleChange, 0, len(batch))
				for _, index := range batch {
					changes = append(changes, filter.RuleChange{Operation: filter.ChangeDelete, Target: &plans[index].Rules[0].Expected, CommandOnly: true})
				}
				plan, err = runtime.BuildCommands(filter.RuleSet{Scope: plan.Scope}, changes)
			}
			if err == nil {
				plan.CommandOnly = true
				err = runtime.RunCommands(ctx, plan)
			}
			retry := errors.Is(err, filter.ErrRuleNotFound)
			if err != nil && provider == filter.ProviderIptables {
				retry = retry || strings.Contains(err.Error(), "iptables-restore: line ") && strings.Contains(err.Error(), " failed")
			}
			for _, index := range batch {
				itemErr := err
				if len(batch) > 1 && retry {
					itemErr = ctx.Err()
					if itemErr == nil {
						itemErr = runtime.RunCommands(ctx, plans[index])
					}
				}
				if errors.Is(itemErr, filter.ErrRuleNotFound) {
					record(index, itemErr)
					continue
				}
				attempted = append(attempted, plans[index])
				indexes = append(indexes, index)
				results = append(results, itemErr)
			}
		}
	}
	for index, saveErr := range persistFirewallRuleBatches(ctx, runtime, attempted) {
		record(indexes[index], errors.Join(results[index], saveErr))
	}
	return result, errors.Join(failures...)
}

func (s *FirewallService) resetSubsystemRules(ctx context.Context, request dto.FirewallRuleReset) (dto.FirewallRuleResetResponse, error) {
	result := dto.FirewallRuleResetResponse{}
	if err := lockFirewallLifecycleIdle(); err != nil {
		return result, err
	}
	defer firewallLifecycleTaskMu.Unlock()
	if request.Subsystem != "forwarding" && request.Subsystem != "docker" {
		return result, filter.ErrInvalidRule
	}
	if request.Subsystem == "forwarding" {
		forwardingMutationMu.Lock()
		defer forwardingMutationMu.Unlock()
	} else {
		dockerPortGuardServiceMu.Lock()
		defer dockerPortGuardServiceMu.Unlock()
	}
	backend := string(request.Provider)
	if backend == "" {
		if request.Subsystem == "forwarding" {
			manager, err := newForwardingAdapter(ctx)
			if err != nil {
				return result, err
			}
			backend = manager.Name()
		} else {
			backend = selectedDockerFirewallBackend("")
		}
	}
	if backend != constant.FirewallProviderIptables && backend != constant.FirewallProviderNftables {
		return result, filter.ErrInvalidRule
	}
	backup := dto.FirewallSubsystemBackup{Subsystem: request.Subsystem, Provider: filter.Provider(backend)}
	var cleanup func() error
	if request.Subsystem == "forwarding" {
		manager, err := newForwardingAdapterFor(ctx, backend)
		if err != nil {
			return result, err
		}
		backup, err = newForwardingService().ExportBackup(ctx, filter.Provider(backend))
		if err != nil {
			return result, err
		}
		result.Removed = len(backup.Forwarding)
		cleanup = manager.Cleanup
	} else {
		runtime := newDockerFirewallRuntime(ctx, backend)
		var err error
		backup, err = newDockerPortGuardService().ExportBackup(ctx, filter.Provider(backend))
		if err != nil {
			return result, err
		}
		result.Removed = len(backup.Docker.Policies)
		cleanup = runtime.Cleanup
	}
	if request.Backup == nil || *request.Backup {
		directory, err := firewallBackupDirectory(request.Subsystem)
		if err != nil {
			return result, err
		}
		result.BackupPath, err = writeFirewallJSON(directory, request.Subsystem+"-"+backend+"-"+time.Now().Format("20060102-150405")+"-"+uuid.NewString()+".json", backup)
		if err != nil {
			return result, err
		}
	}
	if err := errors.Join(cleanup(), ctx.Err()); err != nil {
		return result, err
	}
	if err := os.Remove(filepath.Join(global.Dir.FirewallDir, request.Subsystem+"-"+backend+".rules")); err != nil && !errors.Is(err, os.ErrNotExist) {
		return result, err
	}
	key := constant.FirewallForwardingInitializedKey
	if request.Subsystem == "docker" {
		key = constant.FirewallDockerPortGuardStatusKey
	}
	if err := settingRepo.UpdateOrCreate(key, constant.StatusDisable); err != nil {
		return result, err
	}
	result.Disabled = true
	return result, nil
}
