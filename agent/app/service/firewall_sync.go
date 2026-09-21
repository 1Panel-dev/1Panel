package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	dockerfirewall "github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	firewallsync "github.com/1Panel-dev/1Panel/agent/utils/firewall/sync"
	"github.com/google/uuid"
)

var (
	firewallRuleSyncTaskMu sync.Mutex
	firewallRuleSyncTaskID string
)

type firewallDatabaseSyncAdapter interface {
	previewRuleSync(context.Context, dto.FirewallRuleSyncRequest) (dto.FirewallRuleSyncPreview, error)
	syncRules(context.Context, dto.FirewallRuleSyncRequest) (dto.FirewallRuleSyncResult, error)
}

func (s *FirewallService) SyncPortWhitelist(ctx context.Context) error {
	firewallWhitelistMu.Lock()
	defer firewallWhitelistMu.Unlock()

	ports, err := loadFirewallPortWhiteList()
	if err != nil {
		return err
	}
	required, err := firewall.RequiredPortWhitelist(ports)
	if err != nil {
		return err
	}
	provider, err := s.selectedProvider(ctx)
	if err != nil {
		return err
	}
	rules := whitelistRules(provider, firewall.ExpandPortWhitelist(customWhitelist(ports)), firewall.ExpandPortWhitelist(required))
	if provider != filter.ProviderIptables && provider != filter.ProviderNftables && len(rules) > 0 {
		client, err := s.baseClient()
		if err != nil {
			return err
		}
		active, err := client.Status()
		if err != nil || !active {
			return err
		}
	}
	client, err := s.firewallAdapter(provider)
	if err != nil {
		return err
	}
	prepared := make([]dto.FirewallRuleCreateItem, 0, len(rules))
	var failures []error
	activeFamilies := make(map[filter.Family]bool)
	for _, rule := range rules {
		if err := ctx.Err(); err != nil {
			return err
		}
		if provider == filter.ProviderIptables || provider == filter.ProviderNftables {
			active, inspected := activeFamilies[rule.Scope.Family]
			if !inspected {
				initialized, bound, err := loadSystemFirewallFamilyStatus(string(provider), string(rule.Scope.Family))
				if err != nil {
					return err
				}
				active = initialized && bound
				activeFamilies[rule.Scope.Family] = active
			}
			if !active {
				continue
			}
		}
		port := firewall.SystemPort{Family: string(rule.Scope.Family), Port: rule.DestinationPort, Protocol: rule.Protocol, SourceAddress: rule.SourceAddress}
		item, err := prepareFirewallCreateRule(ctx, client, dto.FirewallRuleCreateItem{
			Rule: rule, SourceKind: constant.FirewallRuleSourceSecurity, SourceID: constant.FirewallSystemAcceptedPortSourcePrefix + firewall.SystemPortKey(firewall.SystemPort(port)),
		})
		if err != nil {
			failures = append(failures, fmt.Errorf("prepare whitelist rule %s: %w", firewall.SystemPortKey(port), err))
			continue
		}
		prepared = append(prepared, item)
	}
	if len(failures) > 0 {
		return errors.Join(failures...)
	}
	if len(prepared) == 0 {
		return nil
	}
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()

	scopes := make([]filter.Scope, 0, len(prepared))
	for _, item := range prepared {
		scopes = append(scopes, item.Rule.Scope)
	}
	snapshots, err := readMutableFirewallRuleScopes(client, ctx, scopes)
	if err != nil {
		return err
	}
	byScope := make(map[string]filter.RuleSet, len(snapshots))
	byScopeIdentity := make(map[string]firewallRuleCollisionIndex, len(snapshots))
	for _, snapshot := range snapshots {
		identities, err := observedFirewallRuleCollisionIndex(snapshot)
		if err != nil {
			return err
		}
		byScope[snapshot.Scope.Key()] = snapshot
		byScopeIdentity[snapshot.Scope.Key()] = identities
	}
	var byMatch map[string][]filter.DesiredRule
	type whitelistBatch struct {
		snapshot filter.RuleSet
		changes  []filter.RuleChange
		records  []*model.FirewallRule
	}
	batches := make([]whitelistBatch, 0)
	batchByScope := make(map[string]int)
	for _, item := range prepared {
		if err := ctx.Err(); err != nil {
			return err
		}
		rule := item.Rule
		scope := rule.Scope
		identities := byScopeIdentity[scope.Key()]
		if err := identities.CheckDuplicate(rule); errors.Is(err, filter.ErrRuleOperation) {
			continue
		} else if err != nil {
			return err
		}
		if err := identities.Check(rule); err != nil {
			return err
		}
		key, err := filter.RuleMatchKey(rule)
		if err != nil {
			return err
		}
		if byMatch == nil {
			stored, err := s.rules.List(ctx)
			if err != nil {
				return err
			}
			byMatch = make(map[string][]filter.DesiredRule)
			for _, record := range stored {
				rules, err := compileStoredFirewallRules(ctx, record, client)
				if isFirewallPolicyIncompatible(err) {
					continue
				}
				if err != nil {
					return err
				}
				for _, candidate := range rules {
					key, err := filter.RuleMatchKey(candidate.Rule)
					if err != nil {
						return err
					}
					byMatch[key] = append(byMatch[key], candidate)
				}
			}
		}
		var existing *filter.DesiredRule
		for _, candidate := range byMatch[key] {
			if collision := checkCollisionActions(rule.Action, candidate.Rule.Action); errors.Is(collision, filter.ErrRuleOperation) {
				existing = &candidate
			} else if collision != nil {
				return collision
			}
		}
		var record *model.FirewallRule
		if existing != nil && scope.Chain != filter.BasicBeforeChain {
			rule = existing.Rule
		} else {
			rule.UUID = uuid.NewString()
			if scope.Chain != filter.BasicBeforeChain {
				created, err := firewallRuleModelForCreate(rule, item, constant.FirewallRuleOriginCreated)
				if err != nil {
					return err
				}
				created.UUID = rule.UUID
				record = &created
			}
		}
		if provider != filter.ProviderFirewalld {
			position := int64(1)
			rule.OrderIndex = &position
		}
		index, exists := batchByScope[scope.Key()]
		if !exists || provider != filter.ProviderIptables && provider != filter.ProviderNftables {
			index = len(batches)
			batchByScope[scope.Key()] = index
			batches = append(batches, whitelistBatch{snapshot: byScope[scope.Key()]})
		}
		batches[index].changes = append(batches[index].changes, filter.RuleChange{Operation: filter.ChangeCreate, After: &rule, CommandOnly: true})
		batches[index].records = append(batches[index].records, record)
		if err := identities.Add(rule); err != nil {
			return err
		}
	}
	saveRecords := func(batch whitelistBatch) {
		saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		for _, record := range batch.records {
			if record == nil {
				continue
			}
			if err := s.saveFirewallRule(saveCtx, record); err != nil {
				failures = append(failures, err)
				if global.LOG != nil {
					global.LOG.Errorf("save firewall whitelist rule %s failed: %v", record.UUID, err)
				}
			}
		}
		cancel()
	}
	_, savesRules := client.(filter.RuleSaver)
	plans := make([]filter.CommandBatch, 0, len(batches))
	completed := make([]int, 0, len(batches))
	for batchIndex, batch := range batches {
		if err := ctx.Err(); err != nil {
			failures = append(failures, err)
			break
		}
		commands, err := client.BuildCommands(batch.snapshot, batch.changes)
		commands.CommandOnly = true
		if err == nil {
			err = client.RunCommands(ctx, commands)
		}
		if err != nil {
			failures = append(failures, err)
			if global.LOG != nil {
				global.LOG.Errorf("create firewall whitelist rules for %s failed: %v", batch.snapshot.Scope.Key(), err)
			}
			continue
		}
		verification := filter.CommandBatch{Provider: commands.Provider, Scope: commands.Scope}
		for _, command := range commands.Rules {
			expected := command.Expected
			expected.Locator.Position = nil
			verification.Rules = append(verification.Rules, filter.RuleCommands{Operation: filter.ChangeCreate, Expected: expected})
		}
		plans = append(plans, verification)
		completed = append(completed, batchIndex)
		if !savesRules {
			saveRecords(batch)
		}
	}
	persistenceErrors := persistFirewallRuleBatches(ctx, client, plans)
	for index, batchIndex := range completed {
		if err := persistenceErrors[index]; err != nil {
			failures = append(failures, err)
			continue
		}
		if savesRules {
			saveRecords(batches[batchIndex])
		}
	}

	verification, err := verifyFirewallCommands(ctx, client, plans...)
	if err != nil {
		failures = append(failures, err)
	} else if !verification.Matched {
		failures = append(failures, filter.ErrVerificationFailed)
	}
	return errors.Join(failures...)
}

func (s *FirewallService) PreviewRuleSync(ctx context.Context, clientIP string, request dto.FirewallRuleSyncRequest) (dto.FirewallRuleSyncPreview, error) {
	switch strings.TrimSpace(request.Subsystem) {
	case "forwarding":
		service := s.forwardingSync
		if service == nil {
			service = newForwardingService()
		}
		return service.previewRuleSync(ctx, request)
	case "docker":
		service := s.dockerSync
		if service == nil {
			service = newDockerPortGuardService()
		}
		return service.previewRuleSync(ctx, request)
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

func (s *FirewallService) SyncRules(ctx context.Context, clientIP string, request dto.FirewallRuleSyncRequest) (dto.FirewallRuleSyncResult, error) {
	switch strings.TrimSpace(request.Subsystem) {
	case "forwarding":
		service := s.forwardingSync
		if service == nil {
			service = newForwardingService()
		}
		return service.syncRules(ctx, request)
	case "docker":
		service := s.dockerSync
		if service == nil {
			service = newDockerPortGuardService()
		}
		return service.syncRules(ctx, request)
	default:
		return s.syncSystemRules(ctx, clientIP, request)
	}
}

func (s *FirewallService) CurrentRuleSyncTask() (dto.FirewallRuleSyncTask, error) {
	firewallRuleSyncTaskMu.Lock()
	defer firewallRuleSyncTaskMu.Unlock()
	return currentFirewallRuleSyncTaskLocked()
}

func (s *DockerPortGuardService) Reconcile(ctx context.Context) error {
	dockerPortGuardServiceMu.Lock()
	defer dockerPortGuardServiceMu.Unlock()
	return s.reconcileLocked(ctx)
}

func (s *ForwardingService) Restore(ctx context.Context) error {
	forwardingMutationMu.Lock()
	defer forwardingMutationMu.Unlock()
	enabled, err := s.forwardingEnabled()
	if err != nil || !enabled {
		if err != nil {
			recordForwardingSyncError(err)
		}
		return err
	}
	manager, err := s.clientFactory()
	if err != nil {
		recordForwardingSyncError(err)
		return err
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		recordForwardingSyncError(err)
		return err
	}
	if err := s.initializeForwarding(manager); err != nil {
		recordForwardingSyncError(err)
		return err
	}
	err = manager.ReplaceRules(forwardingRulesFromModels(stored))
	recordForwardingSyncError(err)
	return err
}

func (s *ForwardingService) previewRuleSync(ctx context.Context, request dto.FirewallRuleSyncRequest) (dto.FirewallRuleSyncPreview, error) {
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

func (s *ForwardingService) syncRules(ctx context.Context, request dto.FirewallRuleSyncRequest) (dto.FirewallRuleSyncResult, error) {
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
			if err := s.initializeForwarding(target); err != nil {
				return err
			}
		}
		if err := target.ReplaceRules(desired); err != nil {
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

func (s *DockerPortGuardService) previewRuleSync(ctx context.Context, request dto.FirewallRuleSyncRequest) (dto.FirewallRuleSyncPreview, error) {
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

func (s *DockerPortGuardService) syncRules(ctx context.Context, request dto.FirewallRuleSyncRequest) (dto.FirewallRuleSyncResult, error) {
	dockerPortGuardServiceMu.Lock()
	defer dockerPortGuardServiceMu.Unlock()

	target, policies, targetRuntime, err := s.loadRuleSyncCandidates(ctx, request)
	if err != nil {
		return dto.FirewallRuleSyncResult{}, err
	}
	runtimePolicies := make([]dockerfirewall.Policy, 0, len(policies))
	for _, policy := range policies {
		sources := []string{}
		_ = json.Unmarshal([]byte(policy.Sources), &sources)
		runtimePolicies = append(runtimePolicies, dockerfirewall.Policy{
			UUID: policy.UUID, Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort,
			Protocol: policy.Protocol, Mode: policy.Mode, Sources: sources,
		})
	}
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
		if err := reconcileDockerFirewall(target, runtimePolicies, targetRuntime, targetInventory); err != nil {
			return err
		}
		if err := verifyDockerFirewall(targetRuntime, runtimePolicies, targetInventory.ReadOnly); err != nil {
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
	if reconcileErr != nil {
		return result, reconcileErr
	}
	return result, nil
}

func (s *FirewallService) syncSystemRules(ctx context.Context, clientIP string, request dto.FirewallRuleSyncRequest) (dto.FirewallRuleSyncResult, error) {
	subsystem := strings.TrimSpace(request.Subsystem)
	if subsystem == "" {
		subsystem = "system"
	}
	if err := lockFirewallLifecycleIdle(); err != nil {
		return dto.FirewallRuleSyncResult{}, err
	}
	defer firewallLifecycleTaskMu.Unlock()
	firewallRuleSyncTaskMu.Lock()
	defer firewallRuleSyncTaskMu.Unlock()

	running, err := currentFirewallRuleSyncTaskLocked()
	if err != nil {
		return dto.FirewallRuleSyncResult{}, err
	}
	if running.Executing {
		return dto.FirewallRuleSyncResult{
			Subsystem:      subsystem,
			TargetProvider: request.TargetProvider,
			TaskID:         running.TaskID,
			Queued:         true,
		}, nil
	}
	if subsystem != "system" {
		return dto.FirewallRuleSyncResult{}, fmt.Errorf("%w: firewall synchronization tasks are only available for the system firewall", filter.ErrInvalidRule)
	}
	taskItem, err := task.NewTask(firewallTaskName(task.TaskSync, firewallTaskHost, string(request.TargetProvider)), task.TaskSync, task.TaskScopeFirewall, "", 0)
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

func verifyForwardingRuleSync(target forwarding.Adapter, desired []forwarding.Rule) error {
	actual, err := target.List()
	if err != nil {
		return fmt.Errorf("verify synchronized forwarding rules: %w", err)
	}
	actual, err = normalizeForwardingRuntimeRules(actual)
	if err != nil {
		return fmt.Errorf("verify synchronized forwarding rules: %w", err)
	}
	if !firewallRuleStatesEqual(actual, desired, func(rule forwarding.Rule) string { return rule.Identity() }) {
		return fmt.Errorf("verify synchronized forwarding rules: target rules do not match the database")
	}
	return nil
}

func reconcileDockerFirewall(backend string, policies []dockerfirewall.Policy, runtime dockerfirewall.Runtime, inventory dockerfirewall.PolicyInventory) error {
	families := make(map[string]struct{}, len(policies))
	needsInitialize, needsBind := false, false
	for _, policy := range policies {
		families[policy.Family] = struct{}{}
	}
	if len(families) == 0 {
		initialized := false
		for _, family := range []string{dockerfirewall.FamilyIPv4, dockerfirewall.FamilyIPv6} {
			status := runtime.Status(family)
			if status.Reason == dockerfirewall.ReasonInspectFailed {
				return fmt.Errorf("inspect Docker firewall target %s for %s failed", backend, family)
			}
			initialized = initialized || status.Initialized
		}
		if initialized {
			return runtime.ReplacePolicies(nil, inventory)
		}
		return nil
	}
	for family := range families {
		status := runtime.Status(family)
		needsInitialize = needsInitialize || !status.Initialized
		needsBind = needsBind || !status.Bound || !status.Effective
	}
	var err error
	if needsInitialize {
		err = runtime.Initialize(policies, inventory)
	} else {
		if needsBind {
			err = runtime.Bind()
		}
		if err == nil {
			err = runtime.ReplacePolicies(policies, inventory)
		}
	}
	if err != nil {
		return err
	}
	for family := range families {
		if !runtime.Status(family).Effective {
			return fmt.Errorf("Docker firewall target %s is not effective for %s", backend, family)
		}
	}
	return nil
}

func ReconcileDockerPortGuardBestEffort(ctx context.Context) {
	if err := ReconcileDockerPortGuard(ctx); err != nil {
		global.LOG.Warnf("reconcile Docker port guard failed, err: %v", err)
	}
}
