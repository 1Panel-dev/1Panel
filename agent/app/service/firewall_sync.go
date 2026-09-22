package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	dockerfirewall "github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	firewallsync "github.com/1Panel-dev/1Panel/agent/utils/firewall/sync"
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
	provider, err := s.selectedProvider(ctx)
	if err != nil {
		return err
	}
	_, err = s.syncPortWhitelist(ctx, provider, ports)
	return err
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
