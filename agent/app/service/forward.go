package service

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
)

const (
	forwardingSyncConverged   = "converged"
	forwardingSyncMissing     = "missing"
	forwardingSyncRuntimeOnly = "runtime_only"
)

type IForwardingService interface {
	LoadBaseInfo() (dto.FirewallSubsystemStatus, error)
	SearchRules(request dto.ForwardRuleSearch) (int64, []dto.ForwardRule, error)
	OperateRules(dto.ForwardRuleOperate) (dto.FilterChainOperationResponse, error)
	Enable() error
	QueueInitialization(dto.FirewallInitializationTask) (dto.FilterChainOperationResponse, error)
	Restore(context.Context) error
}

type ForwardingService struct {
	clientFactory  func() (forwarding.Adapter, error)
	rules          repo.IForwardingRuleRepo
	enabled        func() (bool, error)
	persistBackend func(string) error
	markEnabled    func() error
}

var errForwardingBackendUnavailable = errors.New("no supported forwarding backend detected")

var forwardingMutationMu sync.Mutex

var (
	forwardingSyncStateMu sync.RWMutex
	forwardingLastSyncErr error
)

func (s *ForwardingService) LoadBaseInfo() (dto.FirewallSubsystemStatus, error) {
	selected, _ := settingRepo.GetValueByKey(constant.FirewallForwardingBackendKey)
	selected = strings.TrimSpace(selected)
	if selected == "" {
		selected = constant.FirewallProviderIptables
	}
	baseInfo := dto.FirewallSubsystemStatus{
		Version: "-", Name: selected, Backend: selected, SyncError: lastForwardingSyncError(),
	}
	if selected == constant.FirewallProviderIptables || selected == constant.FirewallProviderNftables {
		baseInfo.Name += "-forward"
	}
	manager, err := s.clientFactory()
	if err != nil {
		if errors.Is(err, errForwardingBackendUnavailable) {
			baseInfo.Reason = constant.FirewallBackendNotInstalled
			return baseInfo, nil
		}
		return baseInfo, err
	}
	client, err := lifecycle.NewClient(manager.Name())
	if err != nil {
		return baseInfo, err
	}
	version, versionErr := client.Version()
	status, statusErr := loadForwardingFirewallOverview(manager)
	if err := errors.Join(versionErr, statusErr); err != nil {
		return baseInfo, err
	}
	baseInfo.IsExist = true
	baseInfo.Name, baseInfo.Backend = manager.Name(), manager.Name()
	if baseInfo.Backend == constant.FirewallProviderIptables || baseInfo.Backend == constant.FirewallProviderNftables {
		baseInfo.Name += "-forward"
	}
	baseInfo.Version = version
	baseInfo.PingStatus = firewall.LoadPingStatus()
	baseInfo.IsInit, baseInfo.IsBind = status.IsInit, status.IsBind
	baseInfo.IPv4, baseInfo.IPv6 = status.IPv4, status.IPv6
	for _, family := range []struct {
		command string
		status  *dto.FirewallBackendFamilyStatus
	}{
		{"iptables", &baseInfo.IPv4},
		{"ip6tables", &baseInfo.IPv6},
	} {
		policy, err := loadForwardPolicy(family.command)
		if err != nil {
			global.LOG.Warnf("inspect %s FORWARD policy: %v", family.command, err)
			continue
		}
		family.status.ForwardPolicy = policy
	}
	return baseInfo, nil
}

func (s *ForwardingService) SearchRules(request dto.ForwardRuleSearch) (int64, []dto.ForwardRule, error) {
	if request.Strategy != "" {
		return 0, nil, nil
	}
	stored, err := s.rules.List(context.Background())
	if err != nil {
		return 0, nil, err
	}
	manager, err := s.clientFactory()
	if err != nil {
		return 0, nil, err
	}
	runtime, err := manager.List()
	if err != nil {
		return 0, nil, err
	}
	inventory, err := mergeForwardingInventory(stored, runtime)
	if err != nil {
		return 0, nil, err
	}
	keyword := strings.ToLower(strings.TrimSpace(request.Info))
	filtered := inventory[:0]
	for _, item := range inventory {
		if keyword == "" || forwardingRuleMatchesKeyword(item, keyword) {
			filtered = append(filtered, item)
		}
	}
	inventory = filtered
	total := len(inventory)
	start, end := (request.Page-1)*request.PageSize, request.Page*request.PageSize
	if request.All {
		start, end = 0, total
	}
	if start > total {
		return int64(total), make([]dto.ForwardRule, 0), nil
	}
	if end > total {
		end = total
	}
	pageRules := inventory[start:end]
	var items []dto.ForwardRule
	if pageRules != nil {
		items = make([]dto.ForwardRule, 0, len(pageRules))
	}
	for index, item := range pageRules {
		items = append(items, dto.ForwardRule{
			ID:         item.ID,
			Num:        strconv.Itoa(start + index + 1),
			Family:     item.Rule.Family,
			Protocol:   item.Rule.Protocol,
			Port:       item.Rule.Port,
			TargetIP:   item.Rule.TargetIP,
			TargetPort: item.Rule.TargetPort,
			Interface:  item.Rule.Interface,
			IsDesired:  item.IsDesired,
			IsRuntime:  item.IsRuntime,
			SyncStatus: item.SyncStatus(),
		})
	}
	return int64(total), items, nil
}

func (s *ForwardingService) OperateRules(request dto.ForwardRuleOperate) (dto.FilterChainOperationResponse, error) {
	count := 0
	for _, rule := range request.Rules {
		if rule.Operation == "add" {
			count += strings.Count(rule.Protocol, "/") + 1
		}
		if count > filter.MaxAtomicExpansion {
			return dto.FilterChainOperationResponse{}, fmt.Errorf("create or import at most %d rules per batch (after expansion)", filter.MaxAtomicExpansion)
		}
	}
	operation := task.TaskCreate
	for _, rule := range request.Rules {
		if rule.Operation != "add" {
			operation = task.TaskUpdate
		}
	}
	if forwardingOperationsOnlyRemove(request.Rules) {
		operation = task.TaskDelete
	}
	taskItem, err := task.NewTask(firewallTaskName(operation, firewallTaskForwarding, ""), operation, task.TaskScopeFirewall, "", 0)
	if err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	taskItem.AddSubTaskWithOps(taskItem.Name, func(t *task.Task) error {
		return s.operateRules(t.TaskCtx, request, t)
	}, nil, 0, 0)
	if err := taskRepo.Save(context.Background(), taskItem.Task); err != nil {
		taskItem.LogFailedWithErr(taskItem.Name, err)
		closeUnstartedFirewallTask(taskItem)
		return dto.FilterChainOperationResponse{}, err
	}
	go func() { _ = taskItem.Execute() }()
	return dto.FilterChainOperationResponse{TaskID: taskItem.TaskID, Queued: true}, nil
}

func (s *ForwardingService) Enable() error {
	forwardingMutationMu.Lock()
	defer forwardingMutationMu.Unlock()
	manager, err := s.clientFactory()
	if err != nil {
		recordForwardingSyncError(err)
		return err
	}
	if err := s.persistForwardingEnabled(); err != nil {
		recordForwardingSyncError(err)
		return err
	}
	if err := s.initializeForwarding(manager); err != nil {
		recordForwardingSyncError(err)
		return err
	}
	rules, err := s.rules.List(context.Background())
	if err != nil {
		recordForwardingSyncError(err)
		return err
	}
	err = manager.ReplaceRules(forwardingRulesFromModels(rules))
	recordForwardingSyncError(err)
	return err
}

func (s *ForwardingService) QueueInitialization(request dto.FirewallInitializationTask) (dto.FilterChainOperationResponse, error) {
	if err := task.CheckScopeTaskIsExecuting(task.TaskScopeFirewall, 0); err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	taskItem, err := task.NewTask(firewallTaskName(task.TaskExec, firewallTaskForwarding, ""), task.TaskExec, task.TaskScopeFirewall, request.TaskID, 0)
	if err != nil {
		return dto.FilterChainOperationResponse{}, fmt.Errorf("create forwarding initialization task: %w", err)
	}
	var manager forwarding.Adapter
	var backend string
	taskItem.AddSubTask(i18n.GetMsgByKey("FirewallEnableForwardingStep"), func(t *task.Task) error {
		forwardingMutationMu.Lock()
		defer forwardingMutationMu.Unlock()
		var err error
		manager, err = s.clientFactory()
		if err != nil {
			recordForwardingSyncError(err)
			return err
		}
		backend = manager.Name()
		t.Logf("backend=%s", backend)
		if err := s.persistForwardingEnabled(); err != nil {
			recordForwardingSyncError(err)
			return err
		}
		if err := s.initializeForwarding(manager); err != nil {
			recordForwardingSyncError(err)
			return err
		}
		return nil
	}, nil)
	taskItem.AddSubTask(i18n.GetMsgByKey("FirewallRestoreForwardingRulesStep"), func(t *task.Task) error {
		forwardingMutationMu.Lock()
		defer forwardingMutationMu.Unlock()
		rules, err := s.rules.List(t.TaskCtx)
		if err != nil {
			recordForwardingSyncError(err)
			return err
		}
		err = manager.ReplaceRules(forwardingRulesFromModels(rules))
		recordForwardingSyncError(err)
		return err
	}, nil)
	if err := repo.NewITaskRepo().Save(context.Background(), taskItem.Task); err != nil {
		return dto.FilterChainOperationResponse{}, fmt.Errorf("save forwarding initialization task: %w", err)
	}
	go func() { _ = taskItem.Execute() }()
	return dto.FilterChainOperationResponse{TaskID: taskItem.TaskID, Queued: true}, nil
}

func NewIForwardingService() IForwardingService {
	return newForwardingService()
}

func loadForwardPolicy(command string) (string, error) {
	if !cmd.Which(command) {
		command += "-nft"
		if !cmd.Which(command) {
			return "", nil
		}
	}
	output, err := cmd.NewCommandMgr(cmd.WithTimeout(5*time.Second)).RunWithOptionalSudoAndStdout(command, "-t", "filter", "-w", "2", "-S", "FORWARD")
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 3 && fields[0] == "-P" && fields[1] == "FORWARD" {
			if fields[2] != "ACCEPT" && fields[2] != "DROP" {
				return "", fmt.Errorf("unexpected FORWARD policy: %s", fields[2])
			}
			return fields[2], nil
		}
	}
	return "", errors.New("FORWARD default policy was not found")
}

func lastForwardingSyncError() string {
	forwardingSyncStateMu.RLock()
	defer forwardingSyncStateMu.RUnlock()
	if forwardingLastSyncErr == nil {
		return ""
	}
	return forwardingLastSyncErr.Error()
}

func mergeForwardingInventory(stored []model.ForwardingRule, runtime []forwarding.Rule) ([]forwardingInventoryItem, error) {
	items := make([]forwardingInventoryItem, 0, len(stored)+len(runtime))
	byIdentity := make(map[string]int, len(stored)+len(runtime))
	for _, record := range stored {
		rule, err := forwarding.NormalizeRule(forwarding.Rule{
			Family: record.Family, Protocol: record.Protocol, Port: record.Port, TargetIP: record.TargetIP,
			TargetPort: record.TargetPort, Interface: record.Interface,
		})
		if err != nil {
			return nil, fmt.Errorf("normalize desired forwarding rule: %w", err)
		}
		key := rule.Identity()
		byIdentity[key] = len(items)
		items = append(items, forwardingInventoryItem{ID: record.ID, Rule: rule, IsDesired: true})
	}
	for _, observed := range runtime {
		rule, err := forwarding.NormalizeRule(observed)
		if err != nil {
			return nil, fmt.Errorf("normalize runtime forwarding rule: %w", err)
		}
		key := rule.Identity()
		if index, exists := byIdentity[key]; exists {
			items[index].IsRuntime = true
			continue
		}
		byIdentity[key] = len(items)
		items = append(items, forwardingInventoryItem{Rule: rule, IsRuntime: true})
	}
	return items, nil
}

func forwardingRuleMatchesKeyword(item forwardingInventoryItem, keyword string) bool {
	values := []string{
		item.Rule.Family, item.Rule.Protocol, item.Rule.Port, item.Rule.TargetIP,
		item.Rule.TargetPort, item.Rule.Interface, item.SyncStatus(),
	}
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), keyword) {
			return true
		}
	}
	return false
}

func (s *ForwardingService) operateRules(ctx context.Context, request dto.ForwardRuleOperate, t *task.Task) (resultErr error) {
	forwardingMutationMu.Lock()
	defer forwardingMutationMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	type operationBatch struct {
		operation forwarding.OperationType
		rules     []forwarding.Rule
	}
	groups := make([]operationBatch, 0)
	for _, operation := range request.Rules {
		kind := forwarding.OperationType(operation.Operation)
		if kind != forwarding.OperationAdd && kind != forwarding.OperationRemove {
			return fmt.Errorf("unsupported forwarding operation %q", operation.Operation)
		}
		if len(groups) == 0 || groups[len(groups)-1].operation != kind {
			groups = append(groups, operationBatch{operation: kind})
		}
		for _, protocol := range strings.Split(operation.Protocol, "/") {
			rule, err := forwarding.NormalizeRule(forwarding.Rule{
				Family: operation.Family, Protocol: protocol, Port: operation.Port,
				TargetIP: operation.TargetIP, TargetPort: operation.TargetPort, Interface: operation.Interface,
			})
			if err != nil {
				return err
			}
			groups[len(groups)-1].rules = append(groups[len(groups)-1].rules, rule)
		}
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		return err
	}
	byIdentity := make(map[string]model.ForwardingRule, len(stored))
	for index, rule := range forwardingRulesFromModels(stored) {
		normalized, err := forwarding.NormalizeRule(rule)
		if err != nil {
			return err
		}
		byIdentity[normalized.Identity()] = stored[index]
	}
	succeeded, failed, skipped := 0, 0, 0
	var nativeFailure error
	defer func() {
		recordForwardingSyncError(errors.Join(resultErr, nativeFailure))
		if t != nil {
			t.Log(i18n.GetMsgWithMap("FirewallRuleOperationResult", map[string]interface{}{"succeeded": succeeded, "failed": failed}))
			if skipped > 0 {
				t.Logf("%s: %d", i18n.GetMsgByKey("FirewallCreateRuleSkipped"), skipped)
			}
		}
	}()
	record := func(operation forwarding.OperationType, rule forwarding.Rule, status string, cause error) {
		label := fmt.Sprintf("%s %s %s %s -> %s:%s", operation, rule.Family, rule.Protocol, rule.Port, rule.TargetIP, rule.TargetPort)
		switch status {
		case "skipped":
			skipped++
			if t != nil {
				t.Logf("%s %s: %v", label, i18n.GetMsgByKey("FirewallCreateRuleSkipped"), cause)
			}
		case "failed":
			failed++
			if t != nil {
				t.LogFailedWithErr(label, cause)
			}
		default:
			succeeded++
			if t != nil {
				t.LogSuccess(label)
			}
		}
	}
	if len(request.Rules) == 2 && len(groups) == 2 && groups[0].operation == forwarding.OperationRemove && groups[1].operation == forwarding.OperationAdd {
		old := make(map[string]bool, len(groups[0].rules))
		for _, rule := range groups[0].rules {
			old[rule.Identity()] = true
		}
		unchanged := len(old) == len(groups[1].rules)
		duplicate := false
		for _, rule := range groups[1].rules {
			key := rule.Identity()
			unchanged = unchanged && old[key]
			if _, exists := byIdentity[key]; exists && !old[key] {
				duplicate = true
			}
		}
		if unchanged || duplicate {
			for _, group := range groups {
				for _, rule := range group.rules {
					record(group.operation, rule, "skipped", buserr.New("ErrRecordExist"))
				}
			}
			return nil
		}
	}
	var client forwarding.Adapter
	var failures []error
	for _, group := range groups {
		byFamily := make(map[string][]forwarding.Rule, 2)
		seen := make(map[string]bool, len(group.rules))
		for _, rule := range group.rules {
			key := rule.Identity()
			_, exists := byIdentity[key]
			if seen[key] || (group.operation == forwarding.OperationAdd && exists) {
				record(group.operation, rule, "skipped", buserr.New("ErrRecordExist"))
				continue
			}
			seen[key] = true
			byFamily[rule.Family] = append(byFamily[rule.Family], rule)
		}
		for _, family := range []string{forwarding.FamilyIPv4, forwarding.FamilyIPv6} {
			rules := byFamily[family]
			if len(rules) == 0 {
				continue
			}
			err := ctx.Err()
			if err == nil && client == nil {
				var enabled bool
				enabled, err = s.forwardingEnabled()
				if err == nil && !enabled {
					err = fmt.Errorf("%w: forwarding is not initialized", filter.ErrProviderUnavailable)
				}
				if err == nil {
					client, err = s.clientFactory()
				}
			}
			if err == nil {
				if group.operation == forwarding.OperationAdd {
					err = client.CreateRules(ctx, rules)
				} else {
					err = client.DeleteRules(ctx, rules)
				}
			}
			if err != nil {
				nativeFailure = errors.Join(nativeFailure, err)
				if !request.ForceDelete || !forwardingOperationsOnlyRemove(request.Rules) || ctx.Err() != nil {
					failures = append(failures, err)
					for _, rule := range rules {
						record(group.operation, rule, "failed", err)
					}
					continue
				}
				if t != nil {
					t.Logf("force delete database records: %v", err)
				}
			}
			for start := 0; start < len(rules); start += 500 {
				batch := rules[start:min(start+500, len(rules))]
				records := make([]model.ForwardingRule, 0, len(batch))
				ids := make([]uint, 0, len(batch))
				for _, rule := range batch {
					if group.operation == forwarding.OperationAdd {
						records = append(records, model.ForwardingRule{Family: rule.Family, Protocol: rule.Protocol, Port: rule.Port, TargetIP: rule.TargetIP, TargetPort: rule.TargetPort, Interface: rule.Interface})
					} else if stored, exists := byIdentity[rule.Identity()]; exists {
						ids = append(ids, stored.ID)
					}
				}
				if group.operation == forwarding.OperationAdd {
					err = s.rules.CreateBatch(context.WithoutCancel(ctx), records)
				} else {
					err = s.rules.DeleteBatch(context.WithoutCancel(ctx), ids)
				}
				if err != nil {
					failures = append(failures, err)
				}
				for index, rule := range batch {
					if err != nil {
						record(group.operation, rule, "failed", err)
						continue
					}
					if group.operation == forwarding.OperationAdd {
						byIdentity[rule.Identity()] = records[index]
					} else {
						delete(byIdentity, rule.Identity())
					}
					record(group.operation, rule, "succeeded", nil)
				}
			}
		}
		if group.operation == forwarding.OperationRemove && len(failures) > 0 {
			return errors.Join(failures...)
		}
	}
	return errors.Join(failures...)
}
