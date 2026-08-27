package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
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
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	"gorm.io/gorm"
)

var (
	firewallRuleSyncTaskMu sync.Mutex
	firewallRuleSyncTaskID string
)

const (
	firewallRuleSyncReady    = "ready"
	firewallRuleSyncExisting = "existing"
	firewallRuleSyncBlocked  = "blocked"
)

func firewallSyncSubsystem(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "system"
	}
	return value
}

type firewallRuleSyncOutcome string

const (
	firewallRuleSyncApplied firewallRuleSyncOutcome = "applied"
	firewallRuleSyncSkipped firewallRuleSyncOutcome = "skipped"
	firewallRuleSyncRemoved firewallRuleSyncOutcome = "removed"
	firewallRuleSyncFailed  firewallRuleSyncOutcome = "failed"
)

type firewallRuleSyncEntry struct {
	source  model.FirewallRule
	rule    filter.FirewallRule
	desired filter.DesiredRule
	remove  *filter.ObservedRule
	match   filter.InventoryMatch
	reorder bool
	item    dto.FirewallRuleSyncItem
	err     error
	outcome firewallRuleSyncOutcome
	failure error
}

type firewallRuleSyncScopePlan struct {
	runtime  *firewallRuleRuntime
	snapshot filter.Snapshot
	entries  []*firewallRuleSyncEntry
}

type firewallSystemSyncPlan struct {
	target   filter.Provider
	entries  []*firewallRuleSyncEntry
	database databaseSyncPlan
	scopes   []firewallRuleSyncScopePlan
}

func (s *FirewallService) PreviewRuleSync(
	ctx context.Context,
	clientIP string,
	request dto.FirewallRuleSyncRequest,
) (dto.FirewallRuleSyncPreview, error) {
	switch firewallSyncSubsystem(request.Subsystem) {
	case "forwarding":
		return s.forwardingRuleSyncService().previewRuleSync(ctx, request)
	case "docker":
		return s.dockerRuleSyncService().previewRuleSync(ctx, request)
	default:
		return s.previewSystemRuleSync(ctx, clientIP, request)
	}
}

func (s *FirewallService) previewSystemRuleSync(
	ctx context.Context,
	clientIP string,
	request dto.FirewallRuleSyncRequest,
) (dto.FirewallRuleSyncPreview, error) {
	plan, err := s.loadFirewallRuleSyncPlan(ctx, clientIP, request)
	if err != nil {
		return dto.FirewallRuleSyncPreview{}, err
	}
	return plan.database.preview(), nil
}

func (s *FirewallService) syncRules(
	ctx context.Context,
	clientIP string,
	request dto.FirewallRuleSyncRequest,
) (dto.FirewallRuleSyncResult, error) {
	plan, err := s.loadFirewallRuleSyncPlan(ctx, clientIP, request)
	if err != nil {
		return dto.FirewallRuleSyncResult{}, err
	}
	if plan.database.preview().Blocked > 0 {
		return plan.database.validationResult(), nil
	}
	return s.executeFirewallSystemSyncPlan(ctx, clientIP, plan), nil
}

func (s *FirewallService) loadFirewallRuleSyncPlan(
	ctx context.Context,
	clientIP string,
	request dto.FirewallRuleSyncRequest,
) (firewallSystemSyncPlan, error) {
	target, entries, hasCompileErrors, err := s.loadStoredFirewallRuleSyncCandidates(ctx, request)
	if err != nil {
		return firewallSystemSyncPlan{}, err
	}
	runtime, err := s.adapters.Resolve(target)
	if err != nil {
		return firewallSystemSyncPlan{}, err
	}

	entriesByScope := make(map[string][]*firewallRuleSyncEntry)
	expectedMarkers := make(map[string]struct{})
	for _, entry := range entries {
		entry.item = dto.FirewallRuleSyncItem{SourceUUID: entry.source.UUID, Rule: &entry.rule}
		if entry.err != nil {
			entry.item.Status, entry.item.Reason = firewallRuleSyncBlocked, entry.err.Error()
			continue
		}
		scopeKey := entry.rule.Scope.Key()
		entriesByScope[scopeKey] = append(entriesByScope[scopeKey], entry)
		expectedMarkers[scopeKey+"\x00"+entry.desired.Marker] = struct{}{}
	}

	scopes := firewallRuleSyncScopes(target)
	if hasCompileErrors {
		scopes = scopesWithFirewallSyncCandidates(scopes, entriesByScope)
	}
	scopePlans := make([]firewallRuleSyncScopePlan, 0, len(scopes))
	seenSnapshots := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		snapshot, observeErr := runtime.ObserveMutation(ctx, scope)
		if observeErr != nil {
			return firewallSystemSyncPlan{}, observeErr
		}
		scopeKey := snapshot.Scope.Key()
		if _, seen := seenSnapshots[scopeKey]; seen {
			continue
		}
		seenSnapshots[scopeKey] = struct{}{}
		scopeEntries := append([]*firewallRuleSyncEntry(nil), entriesByScope[scopeKey]...)
		desired := make([]filter.DesiredRule, 0, len(scopeEntries))
		byMarker := make(map[string]*firewallRuleSyncEntry, len(scopeEntries))
		for _, entry := range scopeEntries {
			desired = append(desired, entry.desired)
			byMarker[entry.desired.Marker] = entry
		}
		inventory, mergeErr := filter.MergeInventory(filter.InventoryMergeInput{
			Observed: snapshot.Rules, Desired: desired,
		})
		if mergeErr != nil {
			return firewallSystemSyncPlan{}, mergeErr
		}
		for itemIndex := range inventory {
			inventoryItem := inventory[itemIndex]
			if inventoryItem.Desired == nil {
				continue
			}
			entry, exists := byMarker[inventoryItem.Desired.Marker]
			if !exists {
				continue
			}
			entry.match = inventoryItem.Match
			entry.item.Status, entry.item.Reason = s.classifyFirewallRuleSyncCandidate(
				clientIP, snapshot, entry, inventoryItem,
			)
		}
		if !hasCompileErrors {
			for observedIndex := range snapshot.Rules {
				observed := snapshot.Rules[observedIndex]
				if !strings.HasPrefix(observed.Marker, "1panel-rule:") {
					continue
				}
				if _, exists := expectedMarkers[scopeKey+"\x00"+observed.Marker]; exists {
					continue
				}
				copy := observed
				entry := &firewallRuleSyncEntry{
					source: model.FirewallRule{UUID: strings.TrimPrefix(observed.Marker, "1panel-rule:")},
					rule:   observed.Rule, remove: &copy,
				}
				status, reason := "remove", "managed rule exists only in target backend"
				if observed.Protected || observed.ParseStatus == filter.ParseStatusOpaque {
					status, reason = firewallRuleSyncBlocked, "managed runtime rule cannot be safely removed"
					entry.err = filter.ErrProtectedRule
				}
				entry.item = dto.FirewallRuleSyncItem{
					SourceUUID: entry.source.UUID, Rule: &entry.rule,
					Status: status, Reason: reason,
				}
				entries = append(entries, entry)
				scopeEntries = append(scopeEntries, entry)
			}
		}
		planFirewallManagedOrder(snapshot, scopeEntries)
		scopePlans = append(scopePlans, firewallRuleSyncScopePlan{
			runtime: runtime, snapshot: snapshot, entries: scopeEntries,
		})
	}
	items := make([]dto.FirewallRuleSyncItem, 0, len(entries))
	for _, entry := range entries {
		items = append(items, entry.item)
	}
	return firewallSystemSyncPlan{
		target: target, entries: entries,
		database: databaseSyncPlan{subsystem: firewallSyncSubsystem(request.Subsystem), target: target, items: items},
		scopes:   scopePlans,
	}, nil
}

func firewallProviderHasOrderedManagedRules(provider filter.Provider) bool {
	return provider == filter.ProviderIptables || provider == filter.ProviderNftables || provider == filter.ProviderUFW
}

func planFirewallManagedOrder(
	snapshot filter.Snapshot,
	entries []*firewallRuleSyncEntry,
) {
	if !firewallProviderHasOrderedManagedRules(snapshot.Scope.Provider) {
		return
	}
	byMarker := make(map[string]*firewallRuleSyncEntry, len(entries))
	for _, entry := range entries {
		if entry.remove == nil && entry.err == nil && entry.desired.Marker != "" {
			byMarker[entry.desired.Marker] = entry
		}
	}
	if len(byMarker) < 2 {
		return
	}

	actual := make([]string, 0, len(byMarker))
	segments := make(map[string]int, len(byMarker))
	segment := 0
	for _, observed := range snapshot.Rules {
		_, expected := byMarker[observed.Marker]
		if expected {
			actual = append(actual, observed.Marker)
			if observed.Protected || observed.ParseStatus == filter.ParseStatusOpaque {
				segment++
				segments[observed.Marker] = segment
				segment++
			} else {
				segments[observed.Marker] = segment
			}
			continue
		}
		if strings.HasPrefix(observed.Marker, "1panel-rule:") &&
			!observed.Protected && observed.ParseStatus != filter.ParseStatusOpaque {
			continue
		}
		segment++
	}

	desired := make([]string, 0, len(actual))
	for _, entry := range entries {
		marker := entry.desired.Marker
		if _, exists := segments[marker]; exists {
			desired = append(desired, marker)
		}
	}
	if slices.Equal(actual, desired) {
		return
	}
	drifted := make(map[string]struct{}, len(desired))
	for index := range desired {
		if actual[index] != desired[index] {
			drifted[actual[index]] = struct{}{}
			drifted[desired[index]] = struct{}{}
		}
	}
	feasible, previousSegment := true, -1
	for _, marker := range desired {
		if segments[marker] < previousSegment {
			feasible = false
			break
		}
		previousSegment = segments[marker]
	}
	for _, marker := range desired {
		if _, exists := drifted[marker]; !exists {
			continue
		}
		entry := byMarker[marker]
		if !feasible {
			entry.item.Status = firewallRuleSyncBlocked
			entry.item.Reason = "managed rule order cannot cross external, opaque, or protected rules"
			continue
		}
		entry.reorder = true
		if entry.item.Status == firewallRuleSyncExisting {
			entry.item.Status = firewallRuleSyncReady
			entry.item.Reason = "managed rule order differs from database sequence"
		}
	}
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
	preview, err := s.previewSystemRuleSync(ctx, clientIP, request)
	if err != nil {
		return dto.FirewallRuleSyncResult{}, err
	}

	resourceName := fmt.Sprintf("database -> %s", request.TargetProvider)
	taskItem, err := task.NewTaskWithOps(resourceName, task.TaskSync, task.TaskScopeFirewall, request.TaskID, 0)
	if err != nil {
		return dto.FirewallRuleSyncResult{}, fmt.Errorf("create firewall migration task: %w", err)
	}
	taskRequest := request
	var syncResult dto.FirewallRuleSyncResult
	taskItem.AddSubTask(i18n.GetWithName("FirewallSyncStep", string(request.TargetProvider)), func(t *task.Task) error {
		var syncErr error
		syncResult, syncErr = s.syncRules(t.TaskCtx, clientIP, taskRequest)
		if syncErr != nil {
			return syncErr
		}
		t.Log(i18n.GetMsgWithMap("FirewallSyncResult", map[string]interface{}{
			"succeeded": syncResult.Succeeded,
			"existing":  syncResult.Skipped,
			"failed":    syncResult.Failed,
		}))
		if syncResult.Failed > 0 {
			for _, message := range firewallRuleSyncFailureMessages(syncResult.Errors) {
				t.Log(message)
			}
			return errors.New(i18n.GetMsgWithMap("FirewallSyncFailed", map[string]interface{}{"failed": syncResult.Failed}))
		}
		return nil
	}, nil)
	taskItem.AddSubTask(i18n.GetMsgByKey("FirewallVerifyTargetStep"), func(t *task.Task) error {
		verified, verifyErr := s.loadFirewallRuleSyncPlan(t.TaskCtx, clientIP, taskRequest)
		if verifyErr != nil {
			return verifyErr
		}
		for _, item := range verified.database.items {
			if item.Status == firewallRuleSyncExisting {
				continue
			}
			reason := item.Reason
			if reason == "" {
				reason = i18n.GetMsgByKey("FirewallTargetRuleIneffective")
			}
			return errors.New(i18n.GetMsgWithMap("FirewallVerifyRuleFailed", map[string]interface{}{
				"name": item.SourceUUID, "detail": reason,
			}))
		}
		return nil
	}, nil)

	if err := repo.NewITaskRepo().Save(context.Background(), taskItem.Task); err != nil {
		return dto.FirewallRuleSyncResult{}, fmt.Errorf("save firewall migration task: %w", err)
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
		_ = taskItem.Execute()
	}()
	return dto.FirewallRuleSyncResult{
		Subsystem:      "system",
		TargetProvider: request.TargetProvider,
		Total:          preview.Total,
		TaskID:         taskItem.TaskID,
		Queued:         true,
	}, nil
}

func firewallRuleSyncFailureMessages(failures []dto.FirewallRuleSyncFailure) []string {
	type failureGroup struct {
		detail string
		uuids  []string
	}
	groups := make([]failureGroup, 0, len(failures))
	groupByDetail := make(map[string]int, len(failures))
	for _, failure := range failures {
		detail := strings.TrimSpace(failure.Error)
		if detail == "" {
			detail = "database synchronization failed"
		}
		uuid := strings.TrimSpace(failure.SourceUUID)
		if uuid == "" {
			uuid = "-"
		}
		if index, exists := groupByDetail[detail]; exists {
			groups[index].uuids = append(groups[index].uuids, uuid)
			continue
		}
		groupByDetail[detail] = len(groups)
		groups = append(groups, failureGroup{detail: detail, uuids: []string{uuid}})
	}

	messages := make([]string, 0, len(groups))
	for _, group := range groups {
		messages = append(messages, fmt.Sprintf("UUID [%s]: %s", strings.Join(group.uuids, ", "), group.detail))
	}
	return messages
}

func (s *FirewallService) forwardingRuleSyncService() firewallDatabaseSyncAdapter {
	if s.forwardingSync == nil {
		return newForwardingService()
	}
	return s.forwardingSync
}

func (s *FirewallService) dockerRuleSyncService() firewallDatabaseSyncAdapter {
	if s.dockerSync == nil {
		return newDockerPortGuardService()
	}
	return s.dockerSync
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

func (s *FirewallService) loadStoredFirewallRuleSyncCandidates(
	ctx context.Context,
	request dto.FirewallRuleSyncRequest,
) (filter.Provider, []*firewallRuleSyncEntry, bool, error) {
	if request.SourceProvider != "" || request.ResetSource {
		return "", nil, false, fmt.Errorf("%w: system firewall synchronization reads desired rules from the database", filter.ErrInvalidRule)
	}
	selected, err := s.selectedProvider(ctx)
	if err != nil {
		return "", nil, false, err
	}
	if request.TargetProvider != selected {
		return "", nil, false, fmt.Errorf(
			"%w: selected provider is %s, requested target is %s",
			filter.ErrProviderUnavailable, selected, request.TargetProvider,
		)
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		return "", nil, false, err
	}
	sortFirewallPolicies(stored, selected)
	entries := make([]*firewallRuleSyncEntry, 0, len(stored))
	hasCompileErrors := false
	for _, record := range stored {
		desired, convertErr := s.compileStoredFirewallRules(ctx, record, selected)
		if convertErr != nil {
			hasCompileErrors = true
			entries = append(entries, &firewallRuleSyncEntry{source: record, err: convertErr})
			continue
		}
		for _, compiled := range desired {
			entries = append(entries, &firewallRuleSyncEntry{
				source: record, rule: compiled.Rule, desired: compiled,
			})
		}
	}
	return selected, entries, hasCompileErrors, nil
}

func scopesWithFirewallSyncCandidates(scopes []filter.Scope, candidates map[string][]*firewallRuleSyncEntry) []filter.Scope {
	result := make([]filter.Scope, 0, len(candidates))
	for _, scope := range scopes {
		if len(candidates[scope.Key()]) > 0 {
			result = append(result, scope)
		}
	}
	return result
}

func firewallRuleSyncScopes(provider filter.Provider) []filter.Scope {
	base := filter.Scope{Provider: provider, Direction: filter.DirectionInput}
	switch provider {
	case filter.ProviderIptables, filter.ProviderNftables:
		result := make([]filter.Scope, 0, 6)
		for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
			for _, chain := range []string{filter.BasicBeforeChain, filter.IptablesInputChain, filter.BasicAfterChain} {
				scope := base
				scope.Family, scope.Table, scope.Chain = family, "filter", chain
				result = append(result, scope)
			}
		}
		return result
	case filter.ProviderFirewalld:
		base.Family, base.Zone = filter.FamilyInet, filter.FirewalldInputZone
		return []filter.Scope{base}
	case filter.ProviderUFW:
		result := make([]filter.Scope, 0, 2)
		for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
			scope := base
			scope.Family, scope.Chain = family, filter.UFWInputChain
			result = append(result, scope)
		}
		return result
	default:
		return nil
	}
}

func firewallPolicyRulesForProvider(stored model.FirewallRule, provider filter.Provider) ([]filter.FirewallRule, error) {
	if stored.CompatibilityError != "" {
		return nil, fmt.Errorf("%w: %s", filter.ErrUnsupportedScope, stored.CompatibilityError)
	}
	connectionStates := make([]string, 0)
	if stored.ConnectionStates != "" {
		connectionStates = strings.Split(stored.ConnectionStates, ",")
	}
	base := filter.FirewallRule{
		Protocol: stored.Protocol, SourceAddress: stored.SourceAddress, SourcePort: stored.SourcePort,
		DestinationAddress: stored.DestinationAddress, DestinationPort: stored.DestinationPort,
		Interface: stored.Interface, ConnectionStates: connectionStates,
		Action: filter.Action(stored.Action), Description: stored.Description,
	}
	if provider == filter.ProviderFirewalld {
		base.Priority = stored.Priority
	}
	families := []filter.Family{filter.Family(stored.Family)}
	if provider != filter.ProviderFirewalld && len(families) == 1 && families[0] == filter.FamilyInet {
		hasIPv4, hasIPv6 := firewallRuleAddressFamilies(base)
		switch {
		case hasIPv4 && hasIPv6:
			return nil, fmt.Errorf("%w: inet policy contains both IPv4 and IPv6 addresses", filter.ErrUnsupportedScope)
		case hasIPv6 || strings.EqualFold(base.Protocol, "icmpv6"):
			families = []filter.Family{filter.FamilyIPv6}
		case hasIPv4:
			families = []filter.Family{filter.FamilyIPv4}
		default:
			families = []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6}
		}
	}
	result := make([]filter.FirewallRule, 0, len(families))
	for _, family := range families {
		rule := base
		rule.Scope = filter.Scope{Provider: provider, Family: family, Direction: filter.DirectionInput}
		switch provider {
		case filter.ProviderIptables, filter.ProviderNftables:
			rule.Scope.Table, rule.Scope.Chain = "filter", filter.IptablesInputChain
		case filter.ProviderFirewalld:
			rule.Scope.Zone = filter.FirewalldInputZone
		case filter.ProviderUFW:
			rule.Scope.Chain = filter.UFWInputChain
		default:
			return nil, fmt.Errorf("%w: unsupported firewall provider %q", filter.ErrProviderUnavailable, provider)
		}
		expanded, err := filter.ExpandAtomicRules(rule)
		if err != nil {
			return nil, err
		}
		result = append(result, expanded...)
	}
	return result, nil
}

func sortFirewallPolicies(stored []model.FirewallRule, provider filter.Provider) {
	sort.SliceStable(stored, func(i, j int) bool {
		left, right := stored[i], stored[j]
		if provider == filter.ProviderFirewalld {
			switch {
			case left.Priority == nil && right.Priority != nil:
				return false
			case left.Priority != nil && right.Priority == nil:
				return true
			case left.Priority != nil && right.Priority != nil && *left.Priority != *right.Priority:
				return *left.Priority < *right.Priority
			}
		} else {
			switch {
			case left.Sequence == nil && right.Sequence != nil:
				return false
			case left.Sequence != nil && right.Sequence == nil:
				return true
			case left.Sequence != nil && right.Sequence != nil && *left.Sequence != *right.Sequence:
				return *left.Sequence < *right.Sequence
			}
		}
		return left.UUID < right.UUID
	})
}

func firewallRuleAddressFamilies(rule filter.FirewallRule) (bool, bool) {
	hasIPv4, hasIPv6 := false, false
	for _, address := range []string{rule.SourceAddress, rule.DestinationAddress} {
		address = strings.TrimSpace(address)
		if address == "" {
			continue
		}
		if strings.Contains(address, ":") {
			hasIPv6 = true
		} else {
			hasIPv4 = true
		}
	}
	return hasIPv4, hasIPv6
}

func (s *FirewallService) classifyFirewallRuleSyncCandidate(
	clientIP string,
	snapshot filter.Snapshot,
	entry *firewallRuleSyncEntry,
	item filter.InventoryItem,
) (string, string) {
	switch item.Match {
	case filter.InventoryMatchExact:
		return firewallRuleSyncExisting, "rule already matches database policy"
	case filter.InventoryMatchMissing:
		protectedPorts, err := s.loadProtectedPorts()
		if err != nil {
			return firewallRuleSyncBlocked, err.Error()
		}
		if filter.RuleBlocksManagementConnection(entry.rule, clientIP, protectedPorts...) {
			return firewallRuleSyncBlocked, "rule may block the current management connection"
		}
		return firewallRuleSyncReady, "rule is missing from target backend"
	case filter.InventoryMatchChanged:
		if item.Observed == nil {
			return firewallRuleSyncBlocked, filter.ErrRuleStale.Error()
		}
		protectedPorts, err := s.loadProtectedPorts()
		if err != nil {
			return firewallRuleSyncBlocked, err.Error()
		}
		if err := filter.GuardMutation(snapshot, *item.Observed, entry.rule, clientIP, protectedPorts...); err != nil {
			return firewallRuleSyncBlocked, err.Error()
		}
		return firewallRuleSyncReady, "target rule differs from database policy"
	default:
		return firewallRuleSyncBlocked, fmt.Sprintf("target rule cannot be reconciled: %s", item.Match)
	}
}

func firewallRuleSyncInventoryFromSnapshot(
	snapshot filter.Snapshot,
	entry *firewallRuleSyncEntry,
) (filter.InventoryItem, error) {
	items, err := filter.MergeInventory(filter.InventoryMergeInput{
		Observed: snapshot.Rules,
		Desired:  []filter.DesiredRule{entry.desired},
	})
	if err != nil {
		return filter.InventoryItem{}, err
	}
	for _, item := range items {
		if item.Desired != nil && item.Desired.Marker == entry.desired.Marker {
			return item, nil
		}
	}
	return filter.InventoryItem{}, fmt.Errorf(
		"%w: compiled database rule %q was not found in target inventory",
		filter.ErrVerificationFailed, entry.source.UUID,
	)
}

func (s *FirewallService) executeFirewallSystemSyncPlan(
	ctx context.Context,
	clientIP string,
	plan firewallSystemSyncPlan,
) dto.FirewallRuleSyncResult {
	protectedPorts, err := s.loadProtectedPorts()
	if err != nil {
		for _, entry := range plan.entries {
			if entry.item.Status == firewallRuleSyncReady {
				entry.fail(err)
			}
		}
	}

	for _, scopePlan := range plan.scopes {
		reconciler := firewallScopeReconciler{
			ctx: ctx, clientIP: clientIP, protectedPorts: protectedPorts,
			runtime: scopePlan.runtime, snapshot: scopePlan.snapshot, entries: scopePlan.entries,
		}
		reconciler.reconcile()
	}
	return plan.result()
}

func (p firewallSystemSyncPlan) result() dto.FirewallRuleSyncResult {
	result := p.database.baseResult()
	for _, entry := range p.entries {
		switch entry.outcome {
		case firewallRuleSyncApplied:
			result.Succeeded++
		case firewallRuleSyncSkipped:
			result.Skipped++
		case firewallRuleSyncRemoved:
			result.Removed++
		case firewallRuleSyncFailed:
			appendDatabaseSyncFailure(&result, entry.item, entry.failure)
		default:
			if entry.item.Status == firewallRuleSyncExisting {
				result.Skipped++
			}
		}
	}
	return result
}

func (e *firewallRuleSyncEntry) fail(err error) {
	if e.outcome == firewallRuleSyncFailed {
		return
	}
	e.outcome = firewallRuleSyncFailed
	e.failure = err
}

type databaseSyncDesired[T any] struct {
	value T
	item  dto.FirewallRuleSyncItem
	err   error
}

type databaseSyncPlan struct {
	subsystem string
	target    filter.Provider
	items     []dto.FirewallRuleSyncItem
}

type firewallDatabaseSyncAdapter interface {
	previewRuleSync(context.Context, dto.FirewallRuleSyncRequest) (dto.FirewallRuleSyncPreview, error)
	syncRules(context.Context, dto.FirewallRuleSyncRequest) (dto.FirewallRuleSyncResult, error)
}

func buildDatabaseSyncPlan[T any](
	subsystem string,
	target filter.Provider,
	desired []databaseSyncDesired[T],
	actual []T,
	key func(T) string,
	actualItem func(T) dto.FirewallRuleSyncItem,
) databaseSyncPlan {
	items := make([]dto.FirewallRuleSyncItem, 0, len(desired)+len(actual))
	actualByKey := make(map[string][]int, len(actual))
	for index, value := range actual {
		actualByKey[key(value)] = append(actualByKey[key(value)], index)
	}
	matched := make([]bool, len(actual))
	for _, candidate := range desired {
		item := candidate.item
		switch {
		case candidate.err != nil:
			item.Status, item.Reason = firewallRuleSyncBlocked, candidate.err.Error()
		default:
			match := unmatchedDatabaseSyncIndex(actualByKey[key(candidate.value)], matched)
			if match >= 0 {
				matched[match] = true
				item.Status, item.Reason = firewallRuleSyncExisting, "rule already exists in target backend"
			} else {
				item.Status = firewallRuleSyncReady
			}
		}
		items = append(items, item)
	}
	for index, value := range actual {
		if matched[index] {
			continue
		}
		item := actualItem(value)
		item.Status, item.Reason = "remove", "rule exists only in target backend"
		items = append(items, item)
	}
	return databaseSyncPlan{subsystem: subsystem, target: target, items: items}
}

func unmatchedDatabaseSyncIndex(indices []int, matched []bool) int {
	for _, index := range indices {
		if !matched[index] {
			return index
		}
	}
	return -1
}

func databaseSyncStatesEqual[T any](left, right []T, key func(T) string) bool {
	if len(left) != len(right) {
		return false
	}
	counts := make(map[string]int, len(left))
	for _, value := range left {
		counts[key(value)]++
	}
	for _, value := range right {
		valueKey := key(value)
		if counts[valueKey] == 0 {
			return false
		}
		counts[valueKey]--
	}
	return true
}

func (p databaseSyncPlan) preview() dto.FirewallRuleSyncPreview {
	result := dto.FirewallRuleSyncPreview{
		Subsystem: p.subsystem, TargetProvider: p.target,
		Items: append([]dto.FirewallRuleSyncItem(nil), p.items...),
	}
	for _, item := range p.items {
		switch item.Status {
		case firewallRuleSyncReady:
			result.Ready++
			result.Total++
		case firewallRuleSyncExisting:
			result.Existing++
			result.Total++
		case firewallRuleSyncBlocked:
			result.Blocked++
			result.Total++
		case "remove":
			result.Removed++
		}
	}
	return result
}

func (p databaseSyncPlan) baseResult() dto.FirewallRuleSyncResult {
	result := dto.FirewallRuleSyncResult{Subsystem: p.subsystem, TargetProvider: p.target}
	for _, item := range p.items {
		if item.Status != "remove" {
			result.Total++
		}
	}
	return result
}

func (p databaseSyncPlan) completedResult() dto.FirewallRuleSyncResult {
	result := p.baseResult()
	for _, item := range p.items {
		switch item.Status {
		case firewallRuleSyncReady:
			result.Succeeded++
		case firewallRuleSyncExisting:
			result.Skipped++
		case firewallRuleSyncBlocked:
			appendDatabaseSyncFailure(&result, item, errors.New(item.Reason))
		case "remove":
			result.Removed++
		}
	}
	return result
}

func (p databaseSyncPlan) validationResult() dto.FirewallRuleSyncResult {
	result := p.baseResult()
	for _, item := range p.items {
		switch item.Status {
		case firewallRuleSyncExisting:
			result.Skipped++
		case firewallRuleSyncBlocked:
			appendDatabaseSyncFailure(&result, item, errors.New(item.Reason))
		}
	}
	return result
}

func (p databaseSyncPlan) failedResult(cause error) dto.FirewallRuleSyncResult {
	result := p.baseResult()
	for _, item := range p.items {
		switch item.Status {
		case firewallRuleSyncExisting:
			result.Skipped++
		case firewallRuleSyncReady:
			appendDatabaseSyncFailure(&result, item, cause)
		case firewallRuleSyncBlocked:
			appendDatabaseSyncFailure(&result, item, errors.New(item.Reason))
		}
	}
	return result
}

func (p databaseSyncPlan) reconcile(run func() error) (dto.FirewallRuleSyncResult, error) {
	if p.preview().Blocked > 0 {
		return p.validationResult(), nil
	}
	if err := run(); err != nil {
		return p.failedResult(err), err
	}
	return p.completedResult(), nil
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
	return buildForwardingDatabaseSyncPlan(filter.Provider(target.Name()), candidates, targetRules).preview(), nil
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
	plan := buildForwardingDatabaseSyncPlan(filter.Provider(target.Name()), candidates, targetRules)
	desired := make([]forwarding.Rule, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.err == nil {
			desired = append(desired, candidate.rule)
		}
	}
	preview := plan.preview()
	if preview.Blocked > 0 {
		return plan.validationResult(), nil
	}
	if len(desired) == 0 && !targetInitialized {
		return plan.completedResult(), nil
	}
	result, reconcileErr := plan.reconcile(func() error {
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
	})
	recordForwardingSyncError(reconcileErr)
	if reconcileErr != nil {
		if preview.Ready == 0 {
			return result, reconcileErr
		}
		return result, nil
	}
	return result, nil
}

func buildForwardingDatabaseSyncPlan(
	target filter.Provider,
	candidates []forwardingRuleSyncCandidate,
	actual []forwarding.Rule,
) databaseSyncPlan {
	desired := make([]databaseSyncDesired[forwarding.Rule], 0, len(candidates))
	for _, candidate := range candidates {
		desired = append(desired, databaseSyncDesired[forwarding.Rule]{
			value: candidate.rule,
			item: dto.FirewallRuleSyncItem{
				SourceUUID: candidate.rule.Identity(), ForwardRule: forwardingRuleSyncDTO(candidate.rule),
			},
			err: candidate.err,
		})
	}
	return buildDatabaseSyncPlan(
		"forwarding", target, desired, actual,
		func(rule forwarding.Rule) string { return rule.Identity() },
		func(rule forwarding.Rule) dto.FirewallRuleSyncItem {
			return dto.FirewallRuleSyncItem{SourceUUID: rule.Identity(), ForwardRule: forwardingRuleSyncDTO(rule)}
		},
	)
}

func (s *DockerPortGuardService) previewRuleSync(
	ctx context.Context,
	request dto.FirewallRuleSyncRequest,
) (dto.FirewallRuleSyncPreview, error) {
	target, policies, runtime, err := s.loadRuleSyncCandidates(ctx, request)
	if err != nil {
		return dto.FirewallRuleSyncPreview{}, err
	}
	targetPolicies, err := runtime.ListPolicies()
	if err != nil {
		return dto.FirewallRuleSyncPreview{}, err
	}
	return buildDockerDatabaseSyncPlan(filter.Provider(target), policies, targetPolicies).preview(), nil
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
	targetPolicies, err := targetRuntime.ListPolicies()
	if err != nil {
		return dto.FirewallRuleSyncResult{}, err
	}
	plan := buildDockerDatabaseSyncPlan(filter.Provider(target), policies, targetPolicies)
	result, reconcileErr := plan.reconcile(func() error {
		if err := reconcileDockerGuardSyncTarget(target, runtimePolicies, targetRuntime); err != nil {
			return err
		}
		if err := verifyDockerGuardRuleSync(targetRuntime, runtimePolicies); err != nil {
			return err
		}
		if len(policies) == 0 {
			return nil
		}
		if err := settingRepo.UpdateOrCreate(constant.FirewallDockerBackendKey, target); err != nil {
			return err
		}
		return settingRepo.UpdateOrCreate(constant.FirewallDockerPortGuardStatusKey, constant.StatusEnable)
	})
	recordDockerPortGuardReconcileError(reconcileErr)
	if reconcileErr != nil {
		if len(policies) == 0 {
			return result, reconcileErr
		}
		return result, nil
	}
	return result, nil
}

func buildDockerDatabaseSyncPlan(
	target filter.Provider,
	policies []model.DockerPortGuardPolicy,
	actual []docker_guard.Policy,
) databaseSyncPlan {
	desired := make([]databaseSyncDesired[docker_guard.Policy], 0, len(policies))
	for _, policy := range policies {
		desired = append(desired, databaseSyncDesired[docker_guard.Policy]{
			value: dockerGuardPolicyFromModel(policy),
			item:  dto.FirewallRuleSyncItem{SourceUUID: policy.UUID, DockerRule: dockerGuardRuleSyncDTO(policy)},
		})
	}
	return buildDatabaseSyncPlan(
		"docker", target, desired, actual, dockerGuardPolicySyncKey,
		func(policy docker_guard.Policy) dto.FirewallRuleSyncItem {
			return dto.FirewallRuleSyncItem{SourceUUID: policy.UUID, DockerRule: dockerGuardRuntimeRuleSyncDTO(policy)}
		},
	)
}

type firewallScopeReconciler struct {
	ctx            context.Context
	clientIP       string
	protectedPorts []firewall.PortWhitelist
	runtime        *firewallRuleRuntime
	snapshot       filter.Snapshot
	entries        []*firewallRuleSyncEntry
}

func (r *firewallScopeReconciler) reconcile() {
	removes := make([]*firewallRuleSyncEntry, 0)
	updates := make([]*firewallRuleSyncEntry, 0)
	creates := make([]*firewallRuleSyncEntry, 0)
	for _, entry := range r.entries {
		if entry.outcome == firewallRuleSyncFailed {
			continue
		}
		switch entry.item.Status {
		case "remove":
			removes = append(removes, entry)
		case firewallRuleSyncReady:
			switch entry.match {
			case filter.InventoryMatchChanged:
				updates = append(updates, entry)
			case filter.InventoryMatchMissing:
				creates = append(creates, entry)
			}
		}
	}
	sort.SliceStable(removes, func(i, j int) bool {
		return firewallSyncRemovalPosition(removes[i]) > firewallSyncRemovalPosition(removes[j])
	})
	r.applyGroups(removes)
	for _, entry := range updates {
		r.applyGroups([]*firewallRuleSyncEntry{entry})
	}
	if len(r.snapshot.Rules) == 0 {
		r.applyGroups(creates)
	} else {
		for _, entry := range creates {
			r.applyGroups([]*firewallRuleSyncEntry{entry})
		}
	}
	r.restoreOrder()
}

func firewallSyncRemovalPosition(entry *firewallRuleSyncEntry) int {
	if entry.remove == nil || entry.remove.Locator.Position == nil {
		return 0
	}
	return *entry.remove.Locator.Position
}

func (r *firewallScopeReconciler) applyGroups(entries []*firewallRuleSyncEntry) {
	if len(entries) == 0 {
		return
	}
	batch := r.runtime.adapter.Provider() == filter.ProviderIptables || r.runtime.adapter.Provider() == filter.ProviderNftables
	if batch {
		r.apply(entries)
		return
	}
	for _, entry := range entries {
		r.apply([]*firewallRuleSyncEntry{entry})
	}
}

func (r *firewallScopeReconciler) apply(entries []*firewallRuleSyncEntry) {
	changes := make([]filter.DesiredChange, 0, len(entries))
	active := make([]*firewallRuleSyncEntry, 0, len(entries))
	for _, entry := range entries {
		change, changed, err := firewallRuleSyncChange(r.snapshot, r.entries, entry, r.clientIP, r.protectedPorts)
		if err != nil {
			entry.fail(err)
			continue
		}
		if !changed {
			if entry.item.Status == "remove" {
				entry.outcome = firewallRuleSyncRemoved
			} else {
				entry.outcome = firewallRuleSyncSkipped
			}
			continue
		}
		changes = append(changes, change)
		active = append(active, entry)
	}
	if len(changes) == 0 {
		return
	}
	_, verification, err := r.runtime.Execute(r.ctx, r.snapshot, changes)
	if err == nil && !verification.Matched {
		err = filter.ErrVerificationFailed
	}
	if err != nil {
		for _, entry := range active {
			entry.fail(err)
		}
		return
	}
	for _, entry := range active {
		if entry.item.Status == "remove" {
			entry.outcome = firewallRuleSyncRemoved
		} else {
			entry.outcome = firewallRuleSyncApplied
		}
	}
	r.snapshot = verification.Snapshot
}

func (r *firewallScopeReconciler) restoreOrder() {
	reorderEntries := make([]*firewallRuleSyncEntry, 0)
	desiredMarkers := make([]string, 0, len(r.entries))
	for _, entry := range r.entries {
		if entry.remove != nil || entry.err != nil || entry.desired.Marker == "" {
			continue
		}
		desiredMarkers = append(desiredMarkers, entry.desired.Marker)
		if entry.reorder {
			reorderEntries = append(reorderEntries, entry)
		}
	}
	if len(reorderEntries) == 0 {
		return
	}
	for _, entry := range reorderEntries {
		if entry.outcome == firewallRuleSyncFailed {
			r.failOrder(reorderEntries, errors.New("managed rule order was not synchronized because a preceding rule change failed"))
			return
		}
	}

	changed := false
	for step := 0; step < len(desiredMarkers); step++ {
		marker, position, converged, err := nextFirewallManagedOrderChange(r.snapshot, desiredMarkers)
		if err != nil {
			r.failOrder(reorderEntries, err)
			return
		}
		if converged {
			for _, entry := range reorderEntries {
				if changed {
					entry.outcome = firewallRuleSyncApplied
				} else if entry.outcome == "" {
					entry.outcome = firewallRuleSyncSkipped
				}
			}
			return
		}
		observed, _, exists := firewallRuleSyncObservedByMarker(r.snapshot, marker)
		if !exists {
			r.failOrder(reorderEntries, filter.ErrRuleStale)
			return
		}
		after := firewallRuleSyncObservedRule(observed)
		target := int64(position)
		after.OrderIndex = &target
		before := firewallRuleSyncObservedRule(observed)
		locator := observed.Locator
		operation := filter.ChangeReorder
		if r.runtime.adapter.Provider() == filter.ProviderUFW {
			operation = filter.ChangeUpdate
		}
		_, verification, executeErr := r.runtime.Execute(r.ctx, r.snapshot, []filter.DesiredChange{{
			Operation: operation, Before: &before, After: &after, Locator: &locator,
		}})
		if executeErr == nil && !verification.Matched {
			executeErr = filter.ErrVerificationFailed
		}
		if executeErr != nil {
			r.failOrder(reorderEntries, executeErr)
			return
		}
		r.snapshot = verification.Snapshot
		changed = true
	}
	r.failOrder(reorderEntries, fmt.Errorf("%w: managed rule order did not converge", filter.ErrVerificationFailed))
}

func nextFirewallManagedOrderChange(
	snapshot filter.Snapshot,
	desiredMarkers []string,
) (string, int, bool, error) {
	expected := make(map[string]struct{}, len(desiredMarkers))
	for _, marker := range desiredMarkers {
		expected[marker] = struct{}{}
	}
	actual := make([]string, 0, len(desiredMarkers))
	positions := make([]int, 0, len(desiredMarkers))
	for index, observed := range snapshot.Rules {
		if _, exists := expected[observed.Marker]; !exists {
			continue
		}
		actual = append(actual, observed.Marker)
		position := index + 1
		if observed.Locator.Position != nil {
			position = *observed.Locator.Position
		}
		positions = append(positions, position)
	}
	if len(actual) != len(desiredMarkers) {
		return "", 0, false, filter.ErrRuleStale
	}
	for index := range desiredMarkers {
		if actual[index] != desiredMarkers[index] {
			return desiredMarkers[index], positions[index], false, nil
		}
	}
	return "", 0, true, nil
}

func firewallRuleSyncObservedByMarker(snapshot filter.Snapshot, marker string) (filter.ObservedRule, int, bool) {
	for index, observed := range snapshot.Rules {
		if observed.Marker == marker {
			position := index + 1
			if observed.Locator.Position != nil {
				position = *observed.Locator.Position
			}
			return observed, position, true
		}
	}
	return filter.ObservedRule{}, 0, false
}

func firewallRuleSyncObservedRule(observed filter.ObservedRule) filter.FirewallRule {
	rule := observed.Rule
	if rule.UUID == "" && strings.HasPrefix(observed.Marker, "1panel-rule:") {
		rule.UUID = strings.TrimSpace(strings.TrimPrefix(observed.Marker, "1panel-rule:"))
	}
	return rule
}

func (r *firewallScopeReconciler) failOrder(entries []*firewallRuleSyncEntry, err error) {
	for _, entry := range entries {
		if entry.outcome != firewallRuleSyncFailed {
			entry.fail(err)
		}
	}
}

func firewallRuleSyncChange(
	snapshot filter.Snapshot,
	entries []*firewallRuleSyncEntry,
	entry *firewallRuleSyncEntry,
	clientIP string,
	protectedPorts []firewall.PortWhitelist,
) (filter.DesiredChange, bool, error) {
	if entry.remove != nil {
		for index := range snapshot.Rules {
			observed := snapshot.Rules[index]
			if observed.Marker != entry.remove.Marker {
				continue
			}
			if observed.Protected {
				return filter.DesiredChange{}, false, filter.ErrProtectedRule
			}
			before := firewallRuleSyncObservedRule(observed)
			locator := observed.Locator
			return filter.DesiredChange{
				Operation: filter.ChangeDelete, Before: &before, Locator: &locator,
			}, true, nil
		}
		return filter.DesiredChange{}, false, nil
	}
	item, err := firewallRuleSyncInventoryFromSnapshot(snapshot, entry)
	if err != nil {
		return filter.DesiredChange{}, false, err
	}
	after := entry.rule
	after.OrderIndex = nil
	change := filter.DesiredChange{After: &after}
	switch item.Match {
	case filter.InventoryMatchExact:
		return filter.DesiredChange{}, false, nil
	case filter.InventoryMatchMissing:
		if filter.RuleBlocksManagementConnection(entry.rule, clientIP, protectedPorts...) {
			return filter.DesiredChange{}, false, filter.ErrLockoutRisk
		}
		change.Operation = filter.ChangeCreate
		if position := firewallRuleSyncInsertionPosition(snapshot, entries, entry); position != nil {
			after.OrderIndex = position
			change.After = &after
		} else {
			change.Append = entry.rule.Scope.Provider == filter.ProviderUFW
		}
	case filter.InventoryMatchChanged:
		if item.Observed == nil {
			return filter.DesiredChange{}, false, filter.ErrRuleStale
		}
		if err := filter.GuardMutation(snapshot, *item.Observed, entry.rule, clientIP, protectedPorts...); err != nil {
			return filter.DesiredChange{}, false, err
		}
		before := firewallRuleSyncObservedRule(*item.Observed)
		locator := item.Observed.Locator
		change.Operation = filter.ChangeUpdate
		change.Before = &before
		change.Locator = &locator
	default:
		return filter.DesiredChange{}, false, fmt.Errorf("%w: target rule match is %s", filter.ErrRuleOperation, item.Match)
	}
	return change, true, nil
}

func firewallRuleSyncInsertionPosition(
	snapshot filter.Snapshot,
	entries []*firewallRuleSyncEntry,
	target *firewallRuleSyncEntry,
) *int64 {
	if !firewallProviderHasOrderedManagedRules(snapshot.Scope.Provider) {
		return nil
	}
	targetIndex := slices.Index(entries, target)
	for index := targetIndex - 1; index >= 0; index-- {
		entry := entries[index]
		if entry.remove != nil || entry.err != nil {
			continue
		}
		if _, position, exists := firewallRuleSyncObservedByMarker(snapshot, entry.desired.Marker); exists {
			value := int64(position + 1)
			return &value
		}
	}
	for index := targetIndex + 1; index < len(entries); index++ {
		entry := entries[index]
		if entry.remove != nil || entry.err != nil {
			continue
		}
		if _, position, exists := firewallRuleSyncObservedByMarker(snapshot, entry.desired.Marker); exists {
			value := int64(position)
			return &value
		}
	}
	return nil
}
