package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	filterruntime "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/runtime"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/nftables_helper"
	firewallsync "github.com/1Panel-dev/1Panel/agent/utils/firewall/sync"
	"gorm.io/gorm"
)

var (
	firewallRuleSyncTaskMu sync.Mutex
	firewallRuleSyncTaskID string
)

type firewallDatabaseSyncAdapter interface {
	previewRuleSync(context.Context, dto.FirewallRuleSyncRequest) (dto.FirewallRuleSyncPreview, error)
	syncRules(context.Context, dto.FirewallRuleSyncRequest) (dto.FirewallRuleSyncResult, error)
}

type firewallSyncRule struct {
	dto.FirewallRuleSyncItem
	desired  filter.DesiredRule
	observed *filter.ObservedRule
	done     bool
}

func firewallSyncSubsystem(value string) string {
	if value = strings.TrimSpace(value); value == "" {
		return "system"
	}
	return value
}

func (s *FirewallService) PreviewRuleSync(ctx context.Context, clientIP string, request dto.FirewallRuleSyncRequest) (dto.FirewallRuleSyncPreview, error) {
	switch firewallSyncSubsystem(request.Subsystem) {
	case "forwarding":
		return s.forwardingRuleSyncService().previewRuleSync(ctx, request)
	case "docker":
		return s.dockerRuleSyncService().previewRuleSync(ctx, request)
	}
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	_, rules, _, err := s.loadFirewallSyncRules(ctx, request)
	preview := dto.FirewallRuleSyncPreview{Subsystem: "system", TargetProvider: request.TargetProvider, Items: make([]dto.FirewallRuleSyncItem, 0, len(rules))}
	for _, rule := range rules {
		preview.Add(rule.FirewallRuleSyncItem)
	}
	return preview, err
}

func (s *FirewallService) loadFirewallSyncRules(ctx context.Context, request dto.FirewallRuleSyncRequest) (*filterruntime.Engine, []*firewallSyncRule, []filter.Snapshot, error) {
	if request.SourceProvider != "" || request.ResetSource {
		return nil, nil, nil, fmt.Errorf("%w: system synchronization reads rules from the database", filter.ErrInvalidRule)
	}
	if err := s.checkSelectedProvider(ctx, request.TargetProvider); err != nil {
		return nil, nil, nil, err
	}
	runtime, err := s.adapters.Resolve(request.TargetProvider)
	if err != nil {
		return nil, nil, nil, err
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	model.SortFirewallRules(stored, request.TargetProvider)
	rules := make([]*firewallSyncRule, 0, len(stored))
	preservedMarkers := make(map[string]bool)
	compileFailed := false
	for _, record := range stored {
		desired, preserved, err := s.compileRestorableFirewallRules(ctx, record, request.TargetProvider)
		if err != nil {
			compileFailed = true
			rules = append(rules, &firewallSyncRule{FirewallRuleSyncItem: dto.FirewallRuleSyncItem{SourceUUID: record.UUID, Status: firewallsync.StatusBlocked, Reason: err.Error()}})
			continue
		}
		for _, rule := range preserved {
			preservedMarkers[rule.Marker] = true
		}
		for _, native := range desired {
			rule := native.Rule
			rules = append(rules, &firewallSyncRule{desired: native, FirewallRuleSyncItem: dto.FirewallRuleSyncItem{SourceUUID: record.UUID, Rule: &rule}})
		}
	}
	snapshots := make([]filter.Snapshot, 0)
	for _, scope := range filter.ManagedInputScopes(request.TargetProvider) {
		desired := make([]filter.DesiredRule, 0)
		scoped := make([]*firewallSyncRule, 0)
		byUUID := make(map[string]*firewallSyncRule)
		for _, rule := range rules {
			if rule.Rule != nil && rule.Rule.Scope.Key() == scope.Key() {
				scoped = append(scoped, rule)
				desired = append(desired, rule.desired)
				byUUID[rule.desired.Rule.UUID] = rule
			}
		}
		if compileFailed && len(desired) == 0 {
			continue
		}
		snapshot, err := runtime.ObserveMutation(ctx, scope)
		if errors.Is(err, filter.ErrFamilyUnavailable) {
			for _, rule := range scoped {
				rule.Status, rule.Reason = firewallsync.StatusBlocked, err.Error()
			}
			continue
		}
		if err != nil {
			return nil, nil, nil, err
		}
		snapshots = append(snapshots, snapshot)
		inventory, err := filter.MergeInventory(filter.InventoryMergeInput{Observed: snapshot.Rules, Desired: desired})
		if err != nil {
			return nil, nil, nil, err
		}
		matched := make(map[string]filter.InventoryItem)
		for _, item := range inventory {
			if item.Desired == nil {
				if compileFailed || item.Observed == nil || !strings.HasPrefix(item.Observed.Marker, "1panel-rule:") || preservedMarkers[item.Observed.Marker] {
					continue
				}
				observed := item.Observed
				rule := &firewallSyncRule{observed: observed, FirewallRuleSyncItem: dto.FirewallRuleSyncItem{
					SourceUUID: strings.TrimPrefix(observed.Marker, "1panel-rule:"), Rule: &observed.Rule, Status: firewallsync.StatusRemove,
					ReasonCode: firewallsync.ReasonManagedOnlyInTarget, Reason: firewallsync.ReasonMessage(firewallsync.ReasonManagedOnlyInTarget),
				}}
				if observed.Protected || observed.ParseStatus == filter.ParseStatusOpaque {
					rule.Status, rule.ReasonCode = firewallsync.StatusBlocked, firewallsync.ReasonUnsafeRemoval
					rule.Reason = firewallsync.ReasonMessage(rule.ReasonCode)
				}
				rules = append(rules, rule)
				continue
			}
			rule := byUUID[item.Desired.Rule.UUID]
			matched[item.Desired.Rule.UUID] = item
			rule.observed = item.Observed
			divergent := item.Observed != nil && item.Observed.Persistence != "" && item.Observed.Persistence != filter.PersistenceStatusConverged
			switch {
			case item.Match == filter.InventoryMatchExact && !divergent:
				rule.Status, rule.Reason = firewallsync.StatusExisting, "rule already matches database policy"
			case item.Match == filter.InventoryMatchMissing || item.Match == filter.InventoryMatchChanged || item.Match == filter.InventoryMatchExact:
				rule.Status, rule.Reason = firewallsync.StatusReady, "target rule differs from database policy"
				if item.Observed != nil && item.Observed.Protected {
					rule.Status, rule.Reason = firewallsync.StatusBlocked, filter.ErrProtectedRule.Error()
				}
			default:
				rule.Status, rule.Reason = firewallsync.StatusBlocked, fmt.Sprintf("target rule cannot be synchronized: %s", item.Match)
			}
		}
		ordered := make([]filter.InventoryItem, 0, len(scoped))
		for _, rule := range scoped {
			ordered = append(ordered, matched[rule.desired.Rule.UUID])
		}
		drifted := firewallsync.RuleOrder(snapshot, ordered)
		for _, rule := range scoped {
			if !drifted[rule.desired.Marker] {
				continue
			}
			if rule.Status == firewallsync.StatusExisting {
				rule.Status, rule.Reason = firewallsync.StatusReady, "managed rule order differs from database sequence"
			}
		}
	}
	return runtime, rules, snapshots, nil
}

func (s *FirewallService) syncRules(ctx context.Context, _ string, request dto.FirewallRuleSyncRequest, t *task.Task) (result dto.FirewallRuleSyncResult, err error) {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	result = dto.FirewallRuleSyncResult{Subsystem: "system", TargetProvider: request.TargetProvider}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	runtime, rules, snapshots, err := s.loadFirewallSyncRules(ctx, request)
	if err != nil {
		return result, err
	}
	created, removed, unexecuted := 0, 0, 0
	stopped := make(map[string]error)
	failedRemovals := make(map[string]error)
	record := func(operation string, rule *firewallSyncRule, cause error, skipped bool) {
		item := rule.FirewallRuleSyncItem
		if operation == "TaskDelete" && rule.observed != nil {
			item.Rule = &rule.observed.Rule
		}
		switch {
		case skipped:
			result.Skipped++
			unexecuted++
			rule.done = true
		case cause != nil:
			appendDatabaseSyncFailure(&result, item, cause)
			rule.done = true
			if operation == "TaskDelete" {
				failedRemovals[rule.SourceUUID] = cause
			}
		case operation == "TaskDelete":
			removed++
			if rule.Status == firewallsync.StatusRemove {
				result.Removed++
				rule.done = true
			}
		case operation == "TaskCreate":
			created++
			result.Succeeded++
			rule.done = true
		default:
			result.Skipped++
			rule.done = true
		}
		if t == nil {
			return
		}
		label := fmt.Sprintf("%s %s", i18n.GetMsgByKey(operation), item.SourceUUID)
		if r := item.Rule; r != nil {
			label += fmt.Sprintf(" [%s] %s %s:%s -> %s:%s %s", r.Scope.Key(), r.Protocol, r.SourceAddress, r.SourcePort, r.DestinationAddress, r.DestinationPort, r.Action)
		}
		if skipped {
			t.Logf("%s %s: %v", label, i18n.GetMsgByKey("FirewallCreateRuleSkipped"), cause)
		} else {
			t.LogWithStatus(label, cause)
		}
	}
	defer func() {
		if t != nil {
			t.Log(i18n.GetMsgWithMap("FirewallSyncOperationsResult", map[string]interface{}{"created": created, "removed": removed, "failed": result.Failed, "skipped": unexecuted}))
		}
	}()
	blocked := false
	for _, rule := range rules {
		if rule.Status != firewallsync.StatusRemove {
			result.Total++
		}
		if rule.Status == firewallsync.StatusBlocked {
			blocked = true
			record("TaskSync", rule, errors.New(rule.Reason), false)
		}
	}
	if blocked {
		for _, rule := range rules {
			if !rule.done {
				record("TaskSync", rule, filter.ErrRuleOperation, true)
			}
		}
		return result, nil
	}
	for _, rule := range rules {
		if rule.Status == firewallsync.StatusExisting {
			record("TaskSync", rule, nil, false)
		}
	}
	for _, operation := range []filter.ChangeOperation{filter.ChangeDelete, filter.ChangeCreate} {
		name := "TaskCreate"
		if operation == filter.ChangeDelete {
			name = "TaskDelete"
		}
		for _, initial := range snapshots {
			scope := initial.Scope
			queue := make([]*firewallSyncRule, 0)
			for _, rule := range rules {
				if rule.done || rule.Rule.Scope.Key() != scope.Key() {
					continue
				}
				if operation == filter.ChangeDelete && rule.observed != nil || operation == filter.ChangeCreate && rule.Status == firewallsync.StatusReady {
					queue = append(queue, rule)
				}
			}
			if operation == filter.ChangeDelete {
				sort.SliceStable(queue, func(i, j int) bool {
					return syncObservedPosition(queue[i].observed) > syncObservedPosition(queue[j].observed)
				})
			}
			for start := 0; start < len(queue); {
				rule := queue[start]
				cause := stopped[scope.Key()]
				if operation == filter.ChangeCreate && cause == nil {
					for id, err := range failedRemovals {
						if id == rule.SourceUUID || strings.HasPrefix(id, rule.SourceUUID+"-") {
							cause = err
							break
						}
					}
				}
				if cause != nil {
					record(name, rule, cause, true)
					start++
					continue
				}
				snapshot := initial
				err := ctx.Err()
				if err == nil {
					snapshot, err = runtime.ObserveMutation(ctx, scope)
				}
				unreadable := err != nil
				end := start + 1
				if supportsNativeRuleBatch(scope.Provider) && (operation == filter.ChangeDelete || len(snapshot.Rules) == 0) && len(failedRemovals) == 0 {
					end = min(start+filter.MaxAtomicExpansion, len(queue))
				}
				batch := queue[start:end]
				changes := make([]filter.DesiredChange, 0, len(batch))
				for _, entry := range batch {
					if err != nil {
						break
					}
					if operation == filter.ChangeDelete {
						var change filter.DesiredChange
						change, err = firewallsync.DeleteChange(snapshot, *entry.observed, entry.desired)
						changes = append(changes, change)
					} else {
						after := *entry.Rule
						markers := make([]string, 0)
						for _, candidate := range rules {
							if candidate.Rule != nil && candidate.Rule.Scope.Key() == scope.Key() && candidate.desired.Marker != "" {
								markers = append(markers, candidate.desired.Marker)
							}
						}
						after.OrderIndex = nil
						if len(batch) == 1 {
							after.OrderIndex = firewallsync.InsertionPosition(snapshot, markers, entry.desired.Marker)
						}
						changes = append(changes, filter.DesiredChange{Operation: operation, After: &after, Append: scope.Provider == filter.ProviderUFW && after.OrderIndex == nil})
					}
				}
				if err == nil {
					err = runtime.ExecuteSync(ctx, snapshot, changes)
				}
				for _, entry := range batch {
					record(name, entry, err, false)
				}
				if err != nil && (operation == filter.ChangeDelete || scope.Provider != filter.ProviderFirewalld || unreadable || firewallCreateUnavailable(err)) {
					stopped[scope.Key()] = err
				}
				start = end
			}
		}
	}
	return result, nil
}

func syncObservedPosition(rule *filter.ObservedRule) int {
	if rule.Locator.Position != nil {
		return *rule.Locator.Position
	}
	return 0
}

func (s *FirewallService) restoreStoredFirewallRules(ctx context.Context, provider filter.Provider) error {
	result, err := s.syncRules(ctx, "", dto.FirewallRuleSyncRequest{TargetProvider: provider}, nil)
	if err != nil {
		return fmt.Errorf("restore database firewall rules: %w", err)
	}
	failures := make([]error, 0, len(result.Errors))
	for _, failure := range result.Errors {
		failures = append(failures, fmt.Errorf("rule %s: %s", failure.SourceUUID, failure.Error))
	}
	return errors.Join(failures...)
}

func (s *FirewallService) SyncRules(
	ctx context.Context,
	clientIP string,
	request dto.FirewallRuleSyncRequest,
) (dto.FirewallRuleSyncResult, error) {
	switch firewallSyncSubsystem(request.Subsystem) {
	case "forwarding":
		return s.forwardingRuleSyncService().syncRules(ctx, request)
	case "docker":
		return s.dockerRuleSyncService().syncRules(ctx, request)
	default:
		return s.syncSystemRules(ctx, clientIP, request)
	}
}

func (s *FirewallService) syncSystemRules(
	ctx context.Context,
	clientIP string,
	request dto.FirewallRuleSyncRequest,
) (dto.FirewallRuleSyncResult, error) {
	firewallRuleSyncTaskMu.Lock()
	defer firewallRuleSyncTaskMu.Unlock()

	running, err := currentFirewallRuleSyncTaskLocked()
	if err != nil {
		return dto.FirewallRuleSyncResult{}, err
	}
	if running.Executing {
		return runningFirewallRuleSyncResult(request, running.TaskID), nil
	}
	if firewallSyncSubsystem(request.Subsystem) != "system" {
		return dto.FirewallRuleSyncResult{}, fmt.Errorf("%w: firewall synchronization tasks are only available for the system firewall", filter.ErrInvalidRule)
	}
	resourceName := fmt.Sprintf("database -> %s", request.TargetProvider)
	taskItem, err := task.NewTaskWithOps(resourceName, task.TaskSync, task.TaskScopeFirewall, "", 0)
	if err != nil {
		return dto.FirewallRuleSyncResult{}, fmt.Errorf("create firewall sync task: %w", err)
	}
	taskItem.AddSubTaskWithOps(i18n.GetWithName("FirewallSyncStep", string(request.TargetProvider)), func(t *task.Task) error {
		result, err := s.syncRules(t.TaskCtx, clientIP, request, t)
		if err != nil {
			return err
		}
		if result.Failed > 0 {
			return errors.New(i18n.GetMsgWithMap("FirewallSyncFailed", map[string]interface{}{"failed": result.Failed}))
		}
		return nil
	}, nil, 0, 0)

	if err := repo.NewITaskRepo().Save(context.Background(), taskItem.Task); err != nil {
		taskItem.LogFailedWithErr(taskItem.Name, err)
		closeUnstartedFirewallTask(taskItem)
		return dto.FirewallRuleSyncResult{}, fmt.Errorf("save firewall sync task: %w", err)
	}
	firewallRuleSyncTaskID = taskItem.TaskID
	go func() {
		defer func() {
			firewallRuleSyncTaskMu.Lock()
			if firewallRuleSyncTaskID == taskItem.TaskID {
				firewallRuleSyncTaskID = ""
			}
			firewallRuleSyncTaskMu.Unlock()
		}()
		if err := taskItem.Execute(); err != nil && global.LOG != nil {
			global.LOG.Errorf("firewall sync task %s failed: %v", taskItem.TaskID, err)
		}
	}()
	return dto.FirewallRuleSyncResult{
		Subsystem:      "system",
		TargetProvider: request.TargetProvider,
		TaskID:         taskItem.TaskID,
		Queued:         true,
	}, nil
}

func (s *FirewallService) CurrentRuleSyncTask() (dto.FirewallRuleSyncTask, error) {
	firewallRuleSyncTaskMu.Lock()
	defer firewallRuleSyncTaskMu.Unlock()
	return currentFirewallRuleSyncTaskLocked()
}

func currentFirewallRuleSyncTaskLocked() (dto.FirewallRuleSyncTask, error) {
	if firewallRuleSyncTaskID != "" {
		return dto.FirewallRuleSyncTask{TaskID: firewallRuleSyncTaskID, Executing: true}, nil
	}
	if global.TaskDB == nil {
		return dto.FirewallRuleSyncTask{}, nil
	}
	taskRepo := repo.NewITaskRepo()
	record, err := taskRepo.GetFirst(
		repo.WithByStatus(constant.StatusExecuting),
		repo.WithByType(task.TaskScopeFirewall),
		taskRepo.WithOperate(task.TaskSync),
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.FirewallRuleSyncTask{}, nil
		}
		return dto.FirewallRuleSyncTask{}, err
	}
	return dto.FirewallRuleSyncTask{TaskID: record.ID, Executing: true}, nil
}

func runningFirewallRuleSyncResult(request dto.FirewallRuleSyncRequest, taskID string) dto.FirewallRuleSyncResult {
	return dto.FirewallRuleSyncResult{
		Subsystem:      firewallSyncSubsystem(request.Subsystem),
		TargetProvider: request.TargetProvider,
		TaskID:         taskID,
		Queued:         true,
	}
}

func (s *FirewallService) syncConfiguredFirewallPorts(ctx context.Context) error {
	configured, err := loadConfiguredFirewallPortWhiteList()
	if err != nil {
		return err
	}
	required, err := LoadRequiredFirewallPortWhiteList()
	if err != nil {
		return err
	}
	ports := excludeFirewallPorts(configured, required)
	return s.SyncSystemPorts(ctx, nil, systemPorts(ports))
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
		var syncErrors []error
		for _, key := range sortedSystemPortKeys(currentSet) {
			if _, exists := previousSet[key]; exists {
				continue
			}
			if err := s.ensureSystemPort(ctx, currentSet[key]); err != nil {
				wrapped := fmt.Errorf("restore accepted firewall port %s: %w", key, err)
				syncErrors = append(syncErrors, wrapped)
				if global.LOG != nil {
					global.LOG.Errorf("%v", wrapped)
				}
			}
		}
		for _, key := range sortedSystemPortKeys(previousSet) {
			if _, exists := currentSet[key]; exists {
				continue
			}
			if err := s.deleteSystemPort(ctx, previousSet[key]); err != nil {
				wrapped := fmt.Errorf("release accepted firewall port %s: %w", key, err)
				syncErrors = append(syncErrors, wrapped)
				if global.LOG != nil {
					global.LOG.Errorf("%v", wrapped)
				}
			}
		}
		return errors.Join(syncErrors...)
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

func syncManagedAcceptedPorts(previous, current []firewall.PortWhitelist) error {
	return newFirewallService().
		SyncSystemPorts(context.Background(), systemPorts(previous), systemPorts(current))
}

func syncPanelRequiredPorts(provider string, ports []firewall.PortWhitelist) error {
	loadPorts := func() ([]firewall.PortWhitelist, error) { return ports, nil }
	if provider == constant.FirewallProviderIptables {
		manager := newIptablesHelperManager()
		manager.LoadRequiredPorts = loadPorts
		return manager.SyncRequiredPorts(true)
	}
	initialized := false
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		familyInitialized, _, err := nftables_helper.LoadFamilyInitStatus(family, "base")
		if family == filter.FamilyIPv6 && errors.Is(err, filter.ErrFamilyUnavailable) {
			continue
		}
		if err != nil {
			return err
		}
		initialized = initialized || familyInitialized
	}
	if !initialized {
		return nil
	}
	manager := newNftablesHelperManager()
	manager.LoadRequiredPorts = loadPorts
	return manager.SyncRequiredPorts()
}

type forwardingRuleSyncCandidate struct {
	rule forwarding.Rule
	err  error
}

func (s *ForwardingService) loadRuleSyncCandidates(
	ctx context.Context,
	targetProvider filter.Provider,
) (*forwarding.Manager, []forwardingRuleSyncCandidate, []forwarding.Rule, bool, error) {
	target, err := s.manager()
	if err != nil {
		return nil, nil, nil, false, err
	}
	if target.Name() != string(targetProvider) {
		return nil, nil, nil, false, fmt.Errorf(
			"%w: selected forwarding backend is %s, requested target is %s",
			filter.ErrProviderUnavailable, target.Name(), targetProvider,
		)
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		return nil, nil, nil, false, err
	}
	candidates := make([]forwardingRuleSyncCandidate, 0, len(stored))
	for _, record := range stored {
		rule := forwarding.Rule{
			Family: record.Family, Protocol: record.Protocol, Port: record.Port, TargetIP: record.TargetIP,
			TargetPort: record.TargetPort, Interface: record.Interface,
		}
		normalized, normalizeErr := forwarding.NormalizeRule(rule)
		candidates = append(candidates, forwardingRuleSyncCandidate{rule: normalized, err: normalizeErr})
	}
	targetStatus, err := target.Status()
	if err != nil {
		return nil, nil, nil, false, err
	}
	targetRules := make([]forwarding.Rule, 0)
	if targetStatus.IsInit {
		targetRules, err = target.List("", "")
		if err != nil {
			return nil, nil, nil, false, err
		}
		targetRules, err = normalizeForwardingRuntimeRules(targetRules)
		if err != nil {
			return nil, nil, nil, false, err
		}
	}
	return target, candidates, targetRules, targetStatus.IsInit, nil
}

func verifyForwardingRuleSync(target *forwarding.Manager, desired []forwarding.Rule) error {
	actual, err := target.List("", "")
	if err != nil {
		return fmt.Errorf("verify synchronized forwarding rules: %w", err)
	}
	actual, err = normalizeForwardingRuntimeRules(actual)
	if err != nil {
		return fmt.Errorf("verify synchronized forwarding rules: %w", err)
	}
	if !firewallsync.StatesEqual(actual, desired, func(rule forwarding.Rule) string { return rule.Identity() }) {
		return fmt.Errorf("verify synchronized forwarding rules: target rules do not match the database")
	}
	return nil
}

func normalizeForwardingRuntimeRules(rules []forwarding.Rule) ([]forwarding.Rule, error) {
	normalized := make([]forwarding.Rule, 0, len(rules))
	for _, rule := range rules {
		item, err := forwarding.NormalizeRule(rule)
		if err != nil {
			return nil, fmt.Errorf("normalize target forwarding rule %s: %w", rule.Identity(), err)
		}
		normalized = append(normalized, item)
	}
	return normalized, nil
}

func forwardingRuleSyncDTO(rule forwarding.Rule) *dto.ForwardRule {
	return &dto.ForwardRule{
		Family: rule.Family, Protocol: rule.Protocol, Port: rule.Port, TargetIP: rule.TargetIP,
		TargetPort: rule.TargetPort, Interface: rule.Interface,
	}
}

func (s *ForwardingService) previewRuleSync(
	ctx context.Context,
	request dto.FirewallRuleSyncRequest,
) (dto.FirewallRuleSyncPreview, error) {
	targetProvider, err := databaseRuleSyncTarget(request, "forwarding")
	if err != nil {
		return dto.FirewallRuleSyncPreview{}, err
	}
	target, candidates, targetRules, _, err := s.loadRuleSyncCandidates(ctx, targetProvider)
	if err != nil {
		return dto.FirewallRuleSyncPreview{}, err
	}
	return forwardingSyncPreview(filter.Provider(target.Name()), candidates, targetRules), nil
}

func (s *ForwardingService) syncRules(
	ctx context.Context,
	request dto.FirewallRuleSyncRequest,
) (dto.FirewallRuleSyncResult, error) {
	forwardingMutationMu.Lock()
	defer forwardingMutationMu.Unlock()

	targetProvider, err := databaseRuleSyncTarget(request, "forwarding")
	if err != nil {
		return dto.FirewallRuleSyncResult{}, err
	}
	target, candidates, targetRules, targetInitialized, err := s.loadRuleSyncCandidates(ctx, targetProvider)
	if err != nil {
		return dto.FirewallRuleSyncResult{}, err
	}
	preview := forwardingSyncPreview(filter.Provider(target.Name()), candidates, targetRules)
	desired := make([]forwarding.Rule, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.err == nil {
			desired = append(desired, candidate.rule)
		}
	}
	if preview.Blocked > 0 {
		return firewallSyncResult(preview, nil, false), nil
	}
	if len(desired) == 0 && !targetInitialized {
		return firewallSyncResult(preview, nil, true), nil
	}
	reconcileErr := func() error {
		if len(desired) > 0 {
			if err := s.persistForwardingEnabled(); err != nil {
				return err
			}
			if err := s.activateManager(target); err != nil {
				return err
			}
		}
		if err := target.Reconcile(desired); err != nil {
			return err
		}
		return verifyForwardingRuleSync(target, desired)
	}()
	result := firewallSyncResult(preview, reconcileErr, true)
	recordForwardingSyncError(reconcileErr)
	if reconcileErr != nil {
		if preview.Ready == 0 {
			return result, reconcileErr
		}
		return result, nil
	}
	return result, nil
}

func forwardingSyncPreview(
	target filter.Provider,
	candidates []forwardingRuleSyncCandidate,
	actual []forwarding.Rule,
) dto.FirewallRuleSyncPreview {
	desired := make([]firewallsync.Desired[forwarding.Rule, dto.FirewallRuleSyncItem], 0, len(candidates))
	for _, candidate := range candidates {
		desired = append(desired, firewallsync.Desired[forwarding.Rule, dto.FirewallRuleSyncItem]{
			Value: candidate.rule,
			Payload: dto.FirewallRuleSyncItem{
				SourceUUID: candidate.rule.Identity(), ForwardRule: forwardingRuleSyncDTO(candidate.rule),
			},
			Err: candidate.err,
		})
	}
	return firewallDiffPreview(
		"forwarding", target, desired, actual,
		func(rule forwarding.Rule) string { return rule.Identity() },
		func(rule forwarding.Rule) dto.FirewallRuleSyncItem {
			return dto.FirewallRuleSyncItem{SourceUUID: rule.Identity(), ForwardRule: forwardingRuleSyncDTO(rule)}
		},
	)
}

func (s *FirewallService) forwardingRuleSyncService() firewallDatabaseSyncAdapter {
	if s.forwardingSync == nil {
		return newForwardingService()
	}
	return s.forwardingSync
}

func (s *DockerPortGuardService) loadRuleSyncCandidates(
	ctx context.Context,
	request dto.FirewallRuleSyncRequest,
) (string, []model.DockerPortGuardPolicy, dockerGuardRuntime, error) {
	targetProvider, err := databaseRuleSyncTarget(request, "Docker")
	if err != nil {
		return "", nil, nil, err
	}
	target := string(targetProvider)
	selected, err := s.selectedRuleSyncBackend(ctx)
	if err != nil {
		return "", nil, nil, err
	}
	if target != selected {
		return "", nil, nil, fmt.Errorf(
			"%w: selected Docker firewall backend is %s, requested target is %s",
			filter.ErrProviderUnavailable, selected, target,
		)
	}
	policies, err := s.policies.ListManaged(ctx)
	if err != nil {
		return "", nil, nil, err
	}
	return target, policies, s.guardRuntime(target), nil
}

func (s *DockerPortGuardService) selectedRuleSyncBackend(ctx context.Context) (string, error) {
	if global.DB != nil {
		selected, _ := settingRepo.GetValueByKey(constant.FirewallDockerBackendKey)
		selected = strings.ToLower(strings.TrimSpace(selected))
		if selected == constant.FirewallProviderIptables || selected == constant.FirewallProviderNftables {
			return selected, nil
		}
	}
	if s.client == nil {
		return "", fmt.Errorf("%w: Docker firewall backend is unavailable", ErrDockerUnavailable)
	}
	cli, err := s.client()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDockerUnavailable, err)
	}
	defer cli.Close()
	info, err := cli.Info(ctx)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDockerUnavailable, err)
	}
	return selectedDockerFirewallBackend(dockerFirewallBackend(info)), nil
}

func dockerGuardPoliciesFromModels(policies []model.DockerPortGuardPolicy) []docker_guard.Policy {
	result := make([]docker_guard.Policy, 0, len(policies))
	for _, policy := range policies {
		result = append(result, dockerGuardPolicyFromModel(policy))
	}
	return result
}

func dockerGuardRuleSyncDTO(policy model.DockerPortGuardPolicy) *dto.DockerPortGuardEndpoint {
	return &dto.DockerPortGuardEndpoint{
		Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol,
		PolicyUUID: policy.UUID, Mode: policy.Mode, Sources: docker_guard.DecodeSources(policy.Sources), Description: policy.Description,
		TrafficPath: dockerTrafficPathUnknown, ManagementTarget: dockerManagementNeedsDiagnosis,
		ManagementReason: dockerReasonNoMatchingPath,
	}
}

func dockerGuardRuntimeRuleSyncDTO(policy docker_guard.Policy) *dto.DockerPortGuardEndpoint {
	return &dto.DockerPortGuardEndpoint{
		Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol,
		PolicyUUID: policy.UUID, Mode: policy.Mode, Sources: append([]string(nil), policy.Sources...),
		TrafficPath: dockerTrafficPathUnknown, ManagementTarget: dockerManagementNeedsDiagnosis,
		ManagementReason: dockerReasonNoMatchingPath,
	}
}

func dockerGuardReadOnlyRuleSyncDTO(policy docker_guard.ReadOnlyPolicy) *dto.DockerPortGuardEndpoint {
	return &dto.DockerPortGuardEndpoint{
		Family: policy.Policy.Family, HostIP: policy.Policy.HostIP, HostPort: policy.Policy.HostPort,
		Protocol: policy.Policy.Protocol, PolicyUUID: dockerGuardReadOnlyPolicyUUID(policy), Sources: append([]string(nil), policy.Policy.Sources...),
		NativeAction: policy.Action, ReadOnly: true, TrafficPath: dockerTrafficPathUnknown,
		ManagementTarget: dockerManagementNeedsDiagnosis, ManagementReason: dockerReasonNoMatchingPath,
	}
}

func (s *DockerPortGuardService) previewRuleSync(
	ctx context.Context,
	request dto.FirewallRuleSyncRequest,
) (dto.FirewallRuleSyncPreview, error) {
	target, policies, runtime, err := s.loadRuleSyncCandidates(ctx, request)
	if err != nil {
		return dto.FirewallRuleSyncPreview{}, err
	}
	targetInventory, err := runtime.ListPolicies()
	if err != nil {
		return dto.FirewallRuleSyncPreview{}, err
	}
	return dockerSyncPreview(filter.Provider(target), policies, targetInventory), nil
}

func (s *DockerPortGuardService) syncRules(
	ctx context.Context,
	request dto.FirewallRuleSyncRequest,
) (dto.FirewallRuleSyncResult, error) {
	dockerPortGuardServiceMu.Lock()
	defer dockerPortGuardServiceMu.Unlock()

	target, policies, targetRuntime, err := s.loadRuleSyncCandidates(ctx, request)
	if err != nil {
		return dto.FirewallRuleSyncResult{}, err
	}
	runtimePolicies := dockerGuardPoliciesFromModels(policies)
	targetInventory, err := targetRuntime.ListPolicies()
	if err != nil {
		return dto.FirewallRuleSyncResult{}, err
	}
	preview := dockerSyncPreview(filter.Provider(target), policies, targetInventory)
	for _, item := range preview.Items {
		if item.Status == firewallsync.StatusBlocked && item.ReasonCode != firewallsync.ReasonReadOnlyRule {
			return firewallSyncResult(preview, nil, false), nil
		}
	}
	if preview.Ready == 0 && preview.Removed == 0 {
		return firewallSyncResult(preview, nil, false), nil
	}
	reconcileErr := func() error {
		if err := s.replaceRuntimeReadOnlyPolicies(ctx, targetInventory.ReadOnly); err != nil {
			return err
		}
		if err := docker_guard.ReconcileTarget(target, runtimePolicies, targetRuntime); err != nil {
			return err
		}
		if err := docker_guard.Verify(targetRuntime, runtimePolicies, targetInventory.ReadOnly); err != nil {
			return err
		}
		if len(policies) == 0 {
			return nil
		}
		if err := settingRepo.UpdateOrCreate(constant.FirewallDockerBackendKey, target); err != nil {
			return err
		}
		return settingRepo.UpdateOrCreate(constant.FirewallDockerPortGuardStatusKey, constant.StatusEnable)
	}()
	result := firewallSyncResult(preview, reconcileErr, true)
	recordDockerPortGuardReconcileError(reconcileErr)
	if reconcileErr != nil {
		return result, reconcileErr
	}
	return result, nil
}

func dockerSyncPreview(
	target filter.Provider,
	policies []model.DockerPortGuardPolicy,
	inventory docker_guard.PolicyInventory,
) dto.FirewallRuleSyncPreview {
	desired := make([]firewallsync.Desired[docker_guard.Policy, dto.FirewallRuleSyncItem], 0, len(policies))
	for _, policy := range policies {
		desired = append(desired, firewallsync.Desired[docker_guard.Policy, dto.FirewallRuleSyncItem]{
			Value:   dockerGuardPolicyFromModel(policy),
			Payload: dto.FirewallRuleSyncItem{SourceUUID: policy.UUID, DockerRule: dockerGuardRuleSyncDTO(policy)},
		})
	}
	preview := firewallDiffPreview(
		"docker", target, desired, inventory.Policies, docker_guard.PolicySyncKey,
		func(policy docker_guard.Policy) dto.FirewallRuleSyncItem {
			return dto.FirewallRuleSyncItem{SourceUUID: policy.UUID, DockerRule: dockerGuardRuntimeRuleSyncDTO(policy)}
		},
	)
	for _, policy := range inventory.ReadOnly {
		preview.Add(dto.FirewallRuleSyncItem{
			SourceUUID: dockerGuardReadOnlyPolicyUUID(policy),
			DockerRule: dockerGuardReadOnlyRuleSyncDTO(policy),
			Status:     firewallsync.StatusBlocked,
			ReasonCode: firewallsync.ReasonReadOnlyRule,
			Reason:     firewallsync.ReasonMessage(firewallsync.ReasonReadOnlyRule),
		})
	}
	return preview
}

func (s *FirewallService) dockerRuleSyncService() firewallDatabaseSyncAdapter {
	if s.dockerSync == nil {
		return newDockerPortGuardService()
	}
	return s.dockerSync
}

func databaseRuleSyncTarget(request dto.FirewallRuleSyncRequest, subsystem string) (filter.Provider, error) {
	if request.SourceProvider != "" {
		return "", fmt.Errorf("%w: %s synchronization reads rules from the database and does not accept a source provider", filter.ErrInvalidRule, subsystem)
	}
	if request.ResetSource {
		return "", fmt.Errorf("%w: %s synchronization does not have a source firewall to reset", filter.ErrInvalidRule, subsystem)
	}
	if request.TargetProvider != filter.ProviderIptables && request.TargetProvider != filter.ProviderNftables {
		return "", fmt.Errorf("%w: %s synchronization only supports iptables and nftables targets", filter.ErrInvalidRule, subsystem)
	}
	return request.TargetProvider, nil
}

func appendDatabaseSyncFailure(result *dto.FirewallRuleSyncResult, item dto.FirewallRuleSyncItem, err error) {
	if err == nil {
		err = errors.New("database synchronization failed")
	}
	result.Failed++
	result.Errors = append(result.Errors, dto.FirewallRuleSyncFailure{
		SourceUUID: item.SourceUUID,
		Rule:       item.Rule, ForwardRule: item.ForwardRule, DockerRule: item.DockerRule,
		Error: err.Error(),
	})
}

func firewallDiffPreview[T any](subsystem string, target filter.Provider, desired []firewallsync.Desired[T, dto.FirewallRuleSyncItem], actual []T, key func(T) string, actualItem func(T) dto.FirewallRuleSyncItem) dto.FirewallRuleSyncPreview {
	preview := dto.FirewallRuleSyncPreview{Subsystem: subsystem, TargetProvider: target, Items: make([]dto.FirewallRuleSyncItem, 0)}
	for _, item := range firewallsync.Diff(desired, actual, key, actualItem) {
		row := item.Payload
		row.Status, row.ReasonCode, row.Reason = item.Status, item.ReasonCode, item.Reason
		preview.Add(row)
	}
	return preview
}

func firewallSyncResult(preview dto.FirewallRuleSyncPreview, cause error, executed bool) dto.FirewallRuleSyncResult {
	result := dto.FirewallRuleSyncResult{Subsystem: preview.Subsystem, TargetProvider: preview.TargetProvider}
	for _, item := range preview.Items {
		if item.Status != firewallsync.StatusRemove {
			result.Total++
		}
		switch item.Status {
		case firewallsync.StatusExisting:
			result.Skipped++
		case firewallsync.StatusBlocked:
			appendDatabaseSyncFailure(&result, item, errors.New(item.Reason))
		case firewallsync.StatusReady:
			if executed {
				if cause != nil {
					appendDatabaseSyncFailure(&result, item, cause)
				} else {
					result.Succeeded++
				}
			}
		case firewallsync.StatusRemove:
			if executed && cause == nil {
				result.Removed++
			}
		}
	}
	return result
}
