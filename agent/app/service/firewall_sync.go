package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"gorm.io/gorm"
)

const (
	firewallRuleSyncReady    = "ready"
	firewallRuleSyncExisting = "existing"
	firewallRuleSyncBlocked  = "blocked"
)

var (
	firewallRuleSyncTaskMu sync.Mutex
	firewallRuleSyncTaskID string
)

type firewallRuleSyncCandidate struct {
	source model.FirewallRule
	rule   filter.FirewallRule
	err    error
}

func (s *FirewallService) PreviewRuleSync(
	ctx context.Context,
	clientIP string,
	request dto.FirewallRuleSyncRequest,
) (dto.FirewallRuleSyncPreview, error) {
	target, candidates, err := s.loadRuleSyncCandidates(ctx, request.SourceProvider, request.TargetProvider)
	if err != nil {
		return dto.FirewallRuleSyncPreview{}, err
	}
	result := dto.FirewallRuleSyncPreview{
		Subsystem:      firewallSyncSubsystem(request.Subsystem),
		SourceProvider: request.SourceProvider,
		TargetProvider: target,
		Items:          make([]dto.FirewallRuleSyncItem, 0, len(candidates)),
	}
	for _, candidate := range candidates {
		item := dto.FirewallRuleSyncItem{SourceUUID: candidate.source.UUID, Rule: &candidate.rule}
		if candidate.err != nil {
			item.Status, item.Reason = firewallRuleSyncBlocked, candidate.err.Error()
			result.Blocked++
		} else {
			check, checkErr := s.checkRule(ctx, clientIP, dto.FirewallRuleCheckItem{Rule: candidate.rule})
			item.Status, item.Reason = firewallRuleSyncCheckStatus(check, checkErr)
			switch item.Status {
			case firewallRuleSyncReady:
				result.Ready++
			case firewallRuleSyncExisting:
				result.Existing++
			default:
				result.Blocked++
			}
		}
		result.Items = append(result.Items, item)
	}
	result.Total = len(result.Items)
	return result, nil
}

func (s *FirewallService) syncRules(
	ctx context.Context,
	clientIP string,
	request dto.FirewallRuleSyncRequest,
) (dto.FirewallRuleSyncResult, error) {
	target, candidates, err := s.loadRuleSyncCandidates(ctx, request.SourceProvider, request.TargetProvider)
	if err != nil {
		return dto.FirewallRuleSyncResult{}, err
	}
	result := dto.FirewallRuleSyncResult{
		Subsystem:      firewallSyncSubsystem(request.Subsystem),
		SourceProvider: request.SourceProvider,
		TargetProvider: target,
		Total:          len(candidates),
	}
	for _, candidate := range candidates {
		if candidate.err != nil {
			appendFirewallRuleSyncFailure(&result, candidate, candidate.err)
			continue
		}
		check, checkErr := s.checkRule(ctx, clientIP, dto.FirewallRuleCheckItem{Rule: candidate.rule})
		status, reason := firewallRuleSyncCheckStatus(check, checkErr)
		switch status {
		case firewallRuleSyncExisting:
			result.Skipped++
			continue
		case firewallRuleSyncBlocked:
			if reason == "" {
				reason = "rule is not eligible for synchronization"
			}
			appendFirewallRuleSyncFailure(&result, candidate, errors.New(reason))
			continue
		}

		create, err := firewallRuleSyncCreateItem(candidate.source, check)
		if err != nil {
			appendFirewallRuleSyncFailure(&result, candidate, err)
			continue
		}
		if err := s.createFirewallRuleItem(ctx, create); err != nil {
			appendFirewallRuleSyncFailure(&result, candidate, err)
			continue
		}
		result.Succeeded++
	}
	return result, nil
}

func (s *FirewallService) SyncRules(
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
	preview, err := s.PreviewRuleSync(ctx, clientIP, request)
	if err != nil {
		return dto.FirewallRuleSyncResult{}, err
	}
	if request.ResetSource && preview.Blocked > 0 {
		return dto.FirewallRuleSyncResult{}, fmt.Errorf(
			"%w: %d source firewall rules cannot be synchronized; source reset was not scheduled",
			filter.ErrRuleOperation, preview.Blocked,
		)
	}
	_, candidates, err := s.loadRuleSyncCandidates(ctx, request.SourceProvider, request.TargetProvider)
	if err != nil {
		return dto.FirewallRuleSyncResult{}, err
	}

	resourceName := fmt.Sprintf("%s -> %s", request.SourceProvider, request.TargetProvider)
	taskItem, err := task.NewTaskWithOps(resourceName, task.TaskSync, task.TaskScopeFirewall, request.TaskID, 0)
	if err != nil {
		return dto.FirewallRuleSyncResult{}, fmt.Errorf("create firewall migration task: %w", err)
	}
	taskRequest := request
	taskRequest.ResetSource = false
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
			messageKey := "FirewallSyncFailed"
			if request.ResetSource {
				messageKey = "FirewallSyncFailedResetSkipped"
			}
			return errors.New(i18n.GetMsgWithMap(messageKey, map[string]interface{}{"failed": syncResult.Failed}))
		}
		return nil
	}, nil)
	if request.ResetSource {
		taskItem.AddSubTask(i18n.GetWithName("FirewallResetSourceStep", string(request.SourceProvider)), func(t *task.Task) error {
			result, resetErr := s.Reset(t.TaskCtx, dto.FirewallRuleReset{Provider: request.SourceProvider})
			if resetErr == nil {
				t.Log(i18n.GetMsgWithMap("FirewallResetSourceResult", map[string]interface{}{"removed": result.Removed}))
			}
			return resetErr
		}, nil)
	}
	taskItem.AddSubTask(i18n.GetMsgByKey("FirewallVerifyTargetStep"), func(t *task.Task) error {
		for _, candidate := range candidates {
			if candidate.err != nil {
				return candidate.err
			}
			check, checkErr := s.checkRule(t.TaskCtx, clientIP, dto.FirewallRuleCheckItem{Rule: candidate.rule})
			status, reason := firewallRuleSyncCheckStatus(check, checkErr)
			if status != firewallRuleSyncExisting {
				if reason == "" {
					reason = i18n.GetMsgByKey("FirewallTargetRuleIneffective")
				}
				return errors.New(i18n.GetMsgWithMap("FirewallVerifyRuleFailed", map[string]interface{}{
					"name": candidate.source.UUID, "detail": reason,
				}))
			}
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
		SourceProvider: request.SourceProvider,
		TargetProvider: request.TargetProvider,
		Total:          preview.Total,
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
		SourceProvider: request.SourceProvider,
		TargetProvider: request.TargetProvider,
		TaskID:         taskID,
		Queued:         true,
	}
}

func firewallSyncSubsystem(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "system"
	}
	return value
}

func (s *FirewallService) loadRuleSyncCandidates(
	ctx context.Context,
	source filter.Provider,
	target filter.Provider,
) (filter.Provider, []firewallRuleSyncCandidate, error) {
	if source == "" {
		return "", nil, fmt.Errorf("%w: source provider is required for system firewall synchronization", filter.ErrInvalidRule)
	}
	selected, err := s.selectedProvider(ctx)
	if err != nil {
		return "", nil, err
	}
	if target != selected {
		return "", nil, fmt.Errorf("%w: selected provider is %s, requested target is %s", filter.ErrProviderUnavailable, selected, target)
	}
	if source == target {
		return "", nil, fmt.Errorf("%w: source and target firewall providers are both %s", filter.ErrInvalidRule, target)
	}
	stored, err := s.rules.List(ctx, repo.WithByProvider(string(source)))
	if err != nil {
		return "", nil, err
	}
	candidates := make([]firewallRuleSyncCandidate, 0, len(stored))
	for _, record := range stored {
		desired, convertErr := desiredFirewallRuleFromModel(record)
		if convertErr != nil {
			candidates = append(candidates, firewallRuleSyncCandidate{source: record, err: convertErr})
			continue
		}
		rules, convertErr := firewallRulesForSyncProvider(desired.Rule, target)
		if convertErr != nil {
			candidates = append(candidates, firewallRuleSyncCandidate{source: record, rule: desired.Rule, err: convertErr})
			continue
		}
		for _, rule := range rules {
			candidates = append(candidates, firewallRuleSyncCandidate{source: record, rule: rule})
		}
	}
	return target, candidates, nil
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

func firewallRulesForSyncProvider(rule filter.FirewallRule, provider filter.Provider) ([]filter.FirewallRule, error) {
	families := []filter.Family{rule.Scope.Family}
	if provider != filter.ProviderFirewalld && rule.Scope.Family == filter.FamilyInet {
		hasIPv4, hasIPv6 := firewallRuleAddressFamilies(rule)
		switch {
		case hasIPv4 && hasIPv6:
			return nil, fmt.Errorf("%w: inet rule contains both IPv4 and IPv6 addresses", filter.ErrUnsupportedScope)
		case hasIPv6 || strings.EqualFold(rule.Protocol, "icmpv6"):
			families = []filter.Family{filter.FamilyIPv6}
		case hasIPv4:
			families = []filter.Family{filter.FamilyIPv4}
		default:
			families = []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6}
		}
	}

	converted := make([]filter.FirewallRule, 0, len(families))
	for _, family := range families {
		item := rule
		item.UUID = ""
		item.NativeKind = ""
		item.Priority = nil
		item.OrderIndex = nil
		item.OrderBucket = ""
		item.Scope = filter.Scope{Provider: provider, Family: family, Direction: filter.DirectionInput}
		switch provider {
		case filter.ProviderIptables, filter.ProviderNftables:
			item.Scope.Table = "filter"
			item.Scope.Chain = filter.IptablesInputChain
		case filter.ProviderFirewalld:
			item.Scope.Zone = filter.FirewalldInputZone
		case filter.ProviderUFW:
			item.Scope.Chain = filter.UFWInputChain
		default:
			return nil, fmt.Errorf("%w: unsupported firewall provider %q", filter.ErrProviderUnavailable, provider)
		}
		expanded, err := filter.ExpandAtomicRules(item)
		if err != nil {
			return nil, err
		}
		converted = append(converted, expanded...)
	}
	return converted, nil
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

func firewallRuleSyncCheckStatus(check dto.FirewallRuleCheckResult, err error) (string, string) {
	if err != nil {
		return firewallRuleSyncBlocked, err.Error()
	}
	if check.Decision == filter.CheckDecisionNoChange || check.Classification == filter.CheckClassificationExactManaged {
		return firewallRuleSyncExisting, check.Reason
	}
	if check.Decision == filter.CheckDecisionBlocked {
		return firewallRuleSyncBlocked, check.Reason
	}
	switch check.Classification {
	case filter.CheckClassificationNone:
		if containsFirewallCheckAction(check.AllowedActions, filter.CheckActionCreate) {
			return firewallRuleSyncReady, check.Reason
		}
	case filter.CheckClassificationCovered:
		if containsFirewallCheckAction(check.AllowedActions, filter.CheckActionCreateAnyway) {
			return firewallRuleSyncReady, check.Reason
		}
	case filter.CheckClassificationExactExternal:
		if len(check.Candidates) == 1 && containsFirewallCheckAction(check.AllowedActions, filter.CheckActionAdopt) {
			return firewallRuleSyncReady, check.Reason
		}
	}
	if check.Reason == "" {
		return firewallRuleSyncBlocked, "rule is not eligible for synchronization"
	}
	return firewallRuleSyncBlocked, check.Reason
}

func firewallRuleSyncCreateItem(
	source model.FirewallRule,
	check dto.FirewallRuleCheckResult,
) (dto.FirewallRuleCreateItem, error) {
	kind, id := firewallRuleOwnerParts(source.Owner)
	create := dto.FirewallRuleCreateItem{
		Rule: check.RequestedRule, CheckFlag: check.CheckFlag, SourceKind: kind, SourceID: id,
	}
	switch check.Classification {
	case filter.CheckClassificationNone:
		create.Action = filter.CheckActionCreate
	case filter.CheckClassificationCovered:
		create.Action = filter.CheckActionCreateAnyway
	case filter.CheckClassificationExactExternal:
		if len(check.Candidates) != 1 {
			return dto.FirewallRuleCreateItem{}, fmt.Errorf("%w: expected one equivalent external rule", filter.ErrRuleOperation)
		}
		create.Action = filter.CheckActionAdopt
		create.AdoptInstanceKey = check.Candidates[0].InstanceKey
	default:
		return dto.FirewallRuleCreateItem{}, fmt.Errorf("%w: %s", filter.ErrRuleOperation, check.Reason)
	}
	return create, nil
}

func firewallRuleOwnerParts(owner string) (string, string) {
	kind, id, found := strings.Cut(strings.TrimSpace(owner), ":")
	if !found {
		return kind, ""
	}
	return kind, id
}

func appendFirewallRuleSyncFailure(
	result *dto.FirewallRuleSyncResult,
	candidate firewallRuleSyncCandidate,
	err error,
) {
	result.Failed++
	result.Errors = append(result.Errors, dto.FirewallRuleSyncFailure{
		SourceUUID: candidate.source.UUID,
		Rule:       &candidate.rule,
		Error:      err.Error(),
	})
}
