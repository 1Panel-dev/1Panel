package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
)

type IForwardingService interface {
	LoadBaseInfo(ctx context.Context) (dto.FirewallSubsystemStatus, error)
	ExportBackup(context.Context, filter.Provider) (dto.FirewallSubsystemBackup, error)
	SearchRules(ctx context.Context, request dto.ForwardRuleSearch) (int64, []dto.ForwardRule, error)
	OperateRules(dto.ForwardRuleOperate) (dto.FilterChainOperationResponse, error)
	Enable(ctx context.Context) error
	QueueInitialization(dto.FirewallInitializationTask) (dto.FilterChainOperationResponse, error)
	Restore(context.Context) error
}

type ForwardingService struct {
	clientFactory func(context.Context) (forwarding.Adapter, error)
}

var errForwardingBackendUnavailable = errors.New("no supported forwarding backend detected")

var forwardingMutationMu sync.Mutex

func (s *ForwardingService) LoadBaseInfo(ctx context.Context) (dto.FirewallSubsystemStatus, error) {
	families, err := loadFirewallFamilies()
	if err != nil {
		return dto.FirewallSubsystemStatus{}, err
	}
	selected, _ := settingRepo.GetValueByKey(constant.FirewallForwardingBackendKey)
	selected = strings.TrimSpace(selected)
	if selected == "" {
		selected = constant.FirewallProviderIptables
	}
	baseInfo := dto.FirewallSubsystemStatus{
		Version: "-", Name: selected, Backend: selected, IPv6Enabled: len(families) > 1,
	}
	if selected == constant.FirewallProviderIptables || selected == constant.FirewallProviderNftables {
		baseInfo.Name += "-forward"
	}
	manager, err := s.clientFactory(ctx)
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
	status, statusErr := loadForwardingFirewallOverview(manager, families)
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
		if family.command == "ip6tables" && !baseInfo.IPv6Enabled {
			continue
		}
		policy, err := loadForwardPolicy(ctx, family.command)
		if err != nil {
			global.LOG.Warnf("inspect %s FORWARD policy: %v", family.command, err)
			continue
		}
		family.status.ForwardPolicy = policy
	}
	return baseInfo, nil
}

func (s *ForwardingService) ExportBackup(ctx context.Context, provider filter.Provider) (dto.FirewallSubsystemBackup, error) {
	var manager forwarding.Adapter
	var err error
	if provider == "" {
		manager, err = s.clientFactory(ctx)
	} else {
		manager, err = newForwardingAdapterFor(ctx, string(provider))
	}
	if err != nil {
		return dto.FirewallSubsystemBackup{}, err
	}
	rules, err := manager.List()
	if err != nil {
		return dto.FirewallSubsystemBackup{}, err
	}
	return forwardingBackup(manager, rules)
}

func (s *ForwardingService) SearchRules(ctx context.Context, request dto.ForwardRuleSearch) (int64, []dto.ForwardRule, error) {
	manager, err := s.clientFactory(ctx)
	if err != nil {
		return 0, nil, err
	}
	rules, err := manager.List()
	if err != nil {
		return 0, nil, err
	}
	items := make([]dto.ForwardRule, 0, len(rules))
	keyword := strings.ToLower(strings.TrimSpace(request.Info))
	for _, rule := range rules {
		if keyword != "" && !strings.Contains(strings.ToLower(strings.Join([]string{rule.Family, rule.Protocol, rule.Port, rule.TargetIP, rule.TargetPort, rule.Interface}, " ")), keyword) {
			continue
		}
		items = append(items, dto.ForwardRule{Num: strconv.Itoa(len(items) + 1), Family: rule.Family, Protocol: rule.Protocol, Port: rule.Port, TargetIP: rule.TargetIP, TargetPort: rule.TargetPort, Interface: rule.Interface})
	}
	total := len(items)
	if request.All {
		return int64(total), items, nil
	}
	start := min(max(request.Page-1, 0)*max(request.PageSize, 1), total)
	end := min(start+max(request.PageSize, 1), total)
	return int64(total), items[start:end], nil
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
		return s.operateRules(t.TaskCtx, request, t, false)
	}, nil, 0, 0)
	if err := taskRepo.Save(context.Background(), taskItem.Task); err != nil {
		taskItem.LogFailedWithErr(taskItem.Name, err)
		closeUnstartedFirewallTask(taskItem)
		return dto.FilterChainOperationResponse{}, err
	}
	go func() { _ = taskItem.Execute() }()
	return dto.FilterChainOperationResponse{TaskID: taskItem.TaskID, Queued: true}, nil
}

func (s *ForwardingService) Enable(ctx context.Context) error {
	forwardingMutationMu.Lock()
	defer forwardingMutationMu.Unlock()
	manager, err := s.clientFactory(ctx)
	if err != nil {
		return err
	}
	rules, err := manager.List()
	if err != nil {
		return err
	}
	families, err := loadFirewallFamilies()
	if err != nil {
		return err
	}
	if err := initializeForwarding(manager, families); err != nil {
		return err
	}
	return persistForwardingRules(manager, rules)
}

func (s *ForwardingService) QueueInitialization(request dto.FirewallInitializationTask) (dto.FilterChainOperationResponse, error) {
	if err := task.CheckScopeTaskIsExecuting(task.TaskScopeFirewall, 0); err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	var backup dto.FirewallSubsystemBackup
	if request.BackupFile != "" {
		var err error
		backup, err = readFirewallSubsystemBackup(request.BackupFile, "forwarding")
		if err != nil {
			return dto.FilterChainOperationResponse{}, err
		}
	}
	operations := dto.ForwardRuleOperate{Rules: make([]dto.ForwardRuleOperation, 0, len(backup.Forwarding))}
	for _, rule := range backup.Forwarding {
		operations.Rules = append(operations.Rules, dto.ForwardRuleOperation{Operation: "add", Family: rule.Family, Protocol: rule.Protocol, Port: rule.Port, TargetIP: rule.TargetIP, TargetPort: rule.TargetPort, Interface: rule.Interface})
	}
	return queueFirewallRuleTask(firewallTaskForwarding, task.TaskExec, request.TaskID, nil, func(t *task.Task) error {
		return s.operateRules(t.TaskCtx, operations, t, true, backup.Families...)
	})
}

func (s *ForwardingService) Restore(ctx context.Context) error {
	forwardingMutationMu.Lock()
	defer forwardingMutationMu.Unlock()
	status, err := settingRepo.GetValueByKey(constant.FirewallForwardingInitializedKey)
	if err != nil || status != constant.StatusEnable {
		return err
	}
	manager, err := s.clientFactory(ctx)
	if err != nil {
		return err
	}
	backup, err := readFirewallSubsystemBackup("forwarding-"+manager.Name()+".rules", "forwarding")
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	families, err := loadFirewallFamilies()
	if err != nil {
		return err
	}
	if !slices.Contains(families, constant.FirewallFamilyIPv6) {
		before := len(backup.Forwarding)
		backup.Forwarding = slices.DeleteFunc(backup.Forwarding, func(rule forwarding.Rule) bool { return rule.Family == constant.FirewallFamilyIPv6 })
		logFirewallIPv6Skipped(nil, "forwarding startup", before-len(backup.Forwarding))
	}
	missing := make(map[string]bool)
	needsBind := false
	for _, family := range families {
		initialized, bound, err := manager.FamilyStatus(family)
		if err != nil {
			return err
		}
		missing[family] = !initialized
		needsBind = needsBind || (initialized && !bound)
	}
	restoreIPv6 := slices.Contains(backup.Families, forwarding.FamilyIPv6)
	for _, rule := range backup.Forwarding {
		restoreIPv6 = restoreIPv6 || rule.Family == forwarding.FamilyIPv6
	}
	if !missing[forwarding.FamilyIPv4] && (!missing[forwarding.FamilyIPv6] || !restoreIPv6) && !needsBind {
		return nil
	}
	current, err := manager.List()
	if err != nil {
		return err
	}
	seen := make(map[string]int, len(current))
	for _, rule := range current {
		seen[rule.Identity()]++
	}
	for _, family := range families {
		if family == forwarding.FamilyIPv6 && missing[family] && !restoreIPv6 {
			continue
		}
		if err := manager.OperateFamily(family, missing[family]); err != nil {
			return err
		}
	}
	for _, family := range families {
		if !missing[family] {
			continue
		}
		rules := make([]forwarding.Rule, 0)
		for _, rule := range backup.Forwarding {
			if rule.Family != family {
				continue
			}
			if seen[rule.Identity()] > 0 {
				seen[rule.Identity()]--
				continue
			}
			rules = append(rules, rule)
		}
		if err := manager.CreateRules(ctx, rules); err != nil {
			return err
		}
	}
	return nil
}

func NewIForwardingService() IForwardingService {
	return newForwardingService()
}

func loadForwardPolicy(ctx context.Context, command string) (string, error) {
	if !cmd.Which(command) {
		command += "-nft"
		if !cmd.Which(command) {
			return "", nil
		}
	}
	output, err := cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(5*time.Second)).RunWithOptionalSudoAndStdout(command, "-t", "filter", "-w", "2", "-S", "FORWARD")
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
