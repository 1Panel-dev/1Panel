package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
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
	"github.com/1Panel-dev/1Panel/agent/utils/controller"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	dockerfirewall "github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	filterfirewalld "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/firewalld"
	filteriptables "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/iptables"
	filternftables "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/nftables"
	filterufw "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/ufw"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	lifecycleproviders "github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle/providers"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/nftables_helper"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	firewallTaskHost       = "FirewallTaskHost"
	firewallTaskForwarding = "FirewallTaskForwarding"
	firewallTaskDocker     = "FirewallTaskDocker"
)

const fail2BanRestoreWithFirewallMarker = "/run/1panel_fail2ban_restore_with_firewall"

type firewallLifecycleClient struct{ lifecycle.Client }

type firewallDockerRestartError struct {
	Err error
}

type firewallCompletedOperationError struct {
	Operation string
	Err       error
}

type firewallRuleBatchItem struct {
	snapshot filter.RuleSet
	change   filter.RuleChange
}

func LoadPanelPort() string {
	if !global.IsMaster {
		return global.CONF.Base.Port
	}
	var portSetting model.Setting
	_ = global.CoreDB.Where("key = ?", "ServerPort").First(&portSetting).Error
	return portSetting.Value
}

func (s *FirewallService) updateSystemAccessPortWhitelist(ctx context.Context, serviceType string, ports []string) error {
	if len(ports) == 0 {
		return fmt.Errorf("firewall whitelist %s requires a port", serviceType)
	}
	firewallWhitelistMu.Lock()
	defer firewallWhitelistMu.Unlock()
	entries, err := loadFirewallPortWhiteList()
	if err != nil {
		return err
	}
	servicePorts := make([]firewall.PortWhitelist, 0)
	for index := range entries {
		if entries[index].Type != serviceType {
			continue
		}
		entries[index].Port = ports[0]
		servicePorts = append(servicePorts, entries[index])
	}
	if len(servicePorts) == 0 {
		families, err := loadFirewallFamilies()
		if err != nil {
			return err
		}
		sources := []string{"0.0.0.0/0"}
		if slices.Contains(families, constant.FirewallFamilyIPv6) {
			sources = append(sources, "::/0")
		}
		servicePorts = append(servicePorts, firewall.PortWhitelist{
			Type: serviceType, Port: ports[0], Protocol: "tcp",
			Sources: sources,
		})
		entries = append(entries, servicePorts...)
	}
	entries, err = firewall.ValidatePortWhitelist(entries)
	if err != nil {
		return err
	}
	value, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	_, err = s.baseClient()
	if err != nil && !errors.Is(err, lifecycle.ErrNotInstalled) {
		return err
	}
	if err == nil {
		allowances := make([]firewall.PortWhitelist, 0, len(servicePorts)*len(ports))
		for _, rule := range servicePorts {
			for _, port := range ports {
				rule.Port = port
				allowances = append(allowances, rule)
			}
		}
		if err := s.applyPortWhitelist(ctx, allowances, nil); err != nil {
			return err
		}
	}
	return settingRepo.UpdateOrCreate(constant.FirewallPortWhiteList, string(value))
}

func loadPortWhitelistSetting() ([]firewall.PortWhitelist, error) {
	setting, err := settingRepo.Get(settingRepo.WithByKey(constant.FirewallPortWhiteList))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		setting.Value = constant.FirewallPortWhiteListValue
	} else if err != nil {
		return nil, err
	}
	var rules []firewall.PortWhitelist
	if err = json.Unmarshal([]byte(setting.Value), &rules); err != nil {
		return nil, err
	}
	families, err := loadFirewallFamilies()
	if err != nil {
		return nil, err
	}
	if !slices.Contains(families, constant.FirewallFamilyIPv6) {
		rules = ipv4PortWhitelist(rules)
	}
	return rules, nil
}

func NewSelectedSystemFirewallClient() (lifecycle.Client, error) {
	provider, _ := settingRepo.GetValueByKey(constant.FirewallSystemBackendKey)
	provider = strings.TrimSpace(provider)
	client, err := lifecycle.NewClient(provider)
	if err != nil {
		return nil, err
	}
	if provider == "" {
		_ = settingRepo.UpdateOrCreate(constant.FirewallSystemBackendKey, client.Name())
	}
	return client, nil
}

func loadFirewallInitStatus(provider, tab string) (bool, bool, error) {
	families, err := loadFirewallFamilies()
	if err != nil {
		return false, false, err
	}
	if tab == "base" && len(families) == 1 {
		return loadSystemFirewallFamilyStatus(provider, constant.FirewallFamilyIPv4)
	}
	switch provider {
	case constant.FirewallProviderNftables:
		return nftables_helper.LoadInitStatus(tab)
	case constant.FirewallProviderIptables:
		return iptables_helper.LoadInitStatus(tab)
	default:
		return false, false, fmt.Errorf("unsupported firewall provider: %s", provider)
	}
}

func loadSystemFirewallOverview(provider, chainGroup string, families []string) (dto.FirewallSubsystemStatus, error) {
	status := dto.FirewallSubsystemStatus{IPv6Enabled: len(families) > 1}
	var ipv4Err, ipv6Err error
	status.IPv4, ipv4Err = loadSystemFirewallFamilyInfo(provider, constant.FirewallFamilyIPv4)
	if status.IPv6Enabled {
		status.IPv6, ipv6Err = loadSystemFirewallFamilyInfo(provider, constant.FirewallFamilyIPv6)
	}
	if chainGroup != "base" {
		return status, nil
	}
	if ipv4Err != nil {
		return status, ipv4Err
	}
	if provider == constant.FirewallProviderIptables || !status.IPv6Enabled {
		status.IsInit, status.IsBind = status.IPv4.Initialized, status.IPv4.Bound
		return status, nil
	}
	if !status.IPv4.Initialized {
		return status, nil
	}
	if !status.IPv4.Bound {
		status.IsInit = true
		return status, nil
	}
	if ipv6Err != nil {
		return status, ipv6Err
	}
	if status.IPv6.Initialized {
		status.IsInit, status.IsBind = true, status.IPv6.Bound
	}
	return status, nil
}

func loadSystemFirewallFamilyInfo(provider, family string) (dto.FirewallBackendFamilyStatus, error) {
	if provider == constant.FirewallProviderIptables && family == constant.FirewallFamilyIPv6 {
		commands, err := lifecycle.ResolveIptablesCommands()
		if err != nil || !commands.IPv6Available() {
			return dto.FirewallBackendFamilyStatus{Reason: dockerfirewall.ReasonCommandMissing}, nil
		}
	}
	var initialized, bound, partial bool
	var err error
	if provider == constant.FirewallProviderIptables {
		initialized, bound, partial, err = iptables_helper.LoadFamilyState(family, "base")
	} else {
		initialized, bound, partial, err = nftables_helper.LoadFamilyState(filter.Family(family))
	}
	return dto.FirewallBackendFamilyStatus{
		Available:   err == nil,
		Partial:     partial,
		Initialized: initialized,
		Bound:       bound,
	}, err
}

func loadSystemFirewallFamilyStatus(provider, family string) (bool, bool, error) {
	switch provider {
	case constant.FirewallProviderIptables:
		return iptables_helper.LoadFamilyInitStatus(family, "base")
	case constant.FirewallProviderNftables:
		return nftables_helper.LoadFamilyInitStatus(filter.Family(family), "base")
	default:
		return false, false, fmt.Errorf("unsupported firewall provider %q", provider)
	}
}

func (s *FirewallService) resolveRuntime(ctx context.Context, provider filter.Provider) (filter.Adapter, error) {
	if s.selectedProvider != nil {
		selected, err := s.selectedProvider(ctx)
		if err != nil {
			return nil, err
		}
		if selected != provider {
			return nil, fmt.Errorf("%w: selected provider is %s, requested %s", filter.ErrProviderUnavailable, selected, provider)
		}
	}
	return s.firewallAdapter(provider)
}

func (s *FirewallService) firewallAdapter(provider filter.Provider) (filter.Adapter, error) {
	if s.adapters != nil {
		if client := s.adapters[provider]; client != nil {
			return client, nil
		}
		return nil, fmt.Errorf("%w: %s", filter.ErrAdapterUnavailable, provider)
	}
	switch provider {
	case filter.ProviderUFW:
		return filterufw.NewAdapter(), nil
	case filter.ProviderFirewalld:
		return filterfirewalld.NewAdapter(), nil
	case filter.ProviderIptables:
		return filteriptables.NewAdapter(), nil
	case filter.ProviderNftables:
		return filternftables.NewAdapter(), nil
	default:
		return nil, fmt.Errorf("%w: %s", filter.ErrAdapterUnavailable, provider)
	}
}

func loadFirewallPortWhiteList() ([]firewall.PortWhitelist, error) {
	ports, err := loadPortWhitelistSetting()
	if err != nil {
		return nil, err
	}
	return firewall.ValidatePortWhitelist(ports)
}

func prepareFirewallBackendRule(ctx context.Context, client filter.Adapter, rule filter.FirewallRule) (filter.FirewallRule, error) {
	if preparer, ok := client.(filter.RulePreparer); ok {
		var err error
		rule, err = preparer.PrepareRule(rule)
		if err != nil {
			return filter.FirewallRule{}, err
		}
	}
	if checker, ok := client.(filter.RuleChecker); ok {
		if err := checker.CheckRule(ctx, rule); err != nil {
			return filter.FirewallRule{}, err
		}
	}
	return rule, nil
}

func readMutableFirewallRules(client filter.Adapter, ctx context.Context, scope filter.Scope) (filter.RuleSet, error) {
	snapshots, err := listFirewallRuleScopes(client, ctx, []filter.Scope{scope})
	if err != nil {
		return filter.RuleSet{}, err
	}
	snapshot := snapshots[0]
	for _, notice := range snapshot.Notices {
		if notice.Code == filter.ScopeNoticeManagedScopeInactive || notice.Code == filter.ScopeNoticeManagedScopeMissing {
			return filter.RuleSet{}, fmt.Errorf("%w: managed firewall scope is unavailable", filter.ErrProviderUnavailable)
		}
	}
	return snapshot, nil
}

func firewallScopeReadGroups(scopes []filter.Scope) [][]filter.Scope {
	groups := make([][]filter.Scope, 0)
	indexes := make(map[string]int)
	for _, scope := range scopes {
		scope = scope.Normalize()
		key := scope.Key()
		if scope.Provider == filter.ProviderUFW {
			key = string(scope.Provider)
		}
		if scope.Provider == filter.ProviderIptables || scope.Provider == filter.ProviderNftables {
			key = string(scope.Provider) + ":" + string(scope.Family) + ":" + scope.Table
		}
		if index, ok := indexes[key]; ok {
			groups[index] = append(groups[index], scope)
		} else {
			indexes[key] = len(groups)
			groups = append(groups, []filter.Scope{scope})
		}
	}
	return groups
}

func listFirewallRuleScopes(client filter.Adapter, ctx context.Context, scopes []filter.Scope) ([]filter.RuleSet, error) {
	unique := make([]filter.Scope, 0, len(scopes))
	seen := make(map[string]bool)
	for _, scope := range scopes {
		scope = scope.Normalize()
		if err := scope.ValidateMVP(); err != nil {
			return nil, err
		}
		if !seen[scope.Key()] {
			seen[scope.Key()] = true
			unique = append(unique, scope)
		}
	}
	if len(unique) == 0 {
		return nil, nil
	}
	if reader, ok := client.(filter.MultiScopeReader); ok {
		snapshots, err := reader.ListRuleScopes(ctx, unique)
		if err != nil {
			return nil, err
		}
		if len(snapshots) != len(unique) {
			return nil, filter.ErrInventoryUnavailable
		}
		return snapshots, nil
	}
	snapshots := make([]filter.RuleSet, 0, len(unique))
	for _, scope := range unique {
		snapshot, err := client.ListRules(ctx, scope)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, nil
}

func applyFirewallChanges(client filter.Adapter, ctx context.Context, snapshot filter.RuleSet, changes []filter.RuleChange) (filter.CommandBatch, error) {
	plan, err := client.BuildCommands(snapshot, changes)
	if err != nil {
		return filter.CommandBatch{}, err
	}
	if err := client.RunCommands(ctx, plan); err != nil {
		return plan, err
	}
	if saver, ok := client.(filter.RuleSaver); ok {
		return plan, saver.SaveRules(ctx, plan.Scope)
	}
	return plan, nil
}

func persistFirewallRuleBatches(ctx context.Context, client filter.Adapter, plans []filter.CommandBatch) []error {
	failures := make([]error, len(plans))
	saver, ok := client.(filter.RuleSaver)
	if !ok {
		return failures
	}
	saved := make(map[string]error)
	for index, plan := range plans {
		key := plan.Scope.Key()
		if plan.Provider == filter.ProviderNftables {
			key = string(plan.Provider)
		}
		err, exists := saved[key]
		if !exists {
			saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
			err = saver.SaveRules(saveCtx, plan.Scope)
			cancel()
			saved[key] = err
		}
		failures[index] = err
	}
	return failures
}

func executeFirewallRuleBatches(ctx context.Context, runtime filter.Adapter, items []firewallRuleBatchItem) []error {
	groups := make([][]int, 0)
	groupIndexes := make(map[string]int)
	for index, item := range items {
		key := string(item.change.Operation) + ":" + item.snapshot.Scope.Key()
		if runtime.Provider() == filter.ProviderUFW {
			key = string(item.change.Operation)
		} else if runtime.Provider() == filter.ProviderFirewalld {
			key += ":" + string(item.change.After.NativeKind)
		}
		group, exists := groupIndexes[key]
		if !exists {
			group = len(groups)
			groupIndexes[key] = group
			groups = append(groups, nil)
		}
		groups[group] = append(groups[group], index)
	}
	failures := make([]error, len(items))
	type appliedBatch struct {
		plan          filter.CommandBatch
		sourceIndexes []int
	}
	var applied []appliedBatch
	for _, group := range groups {
		for start := 0; start < len(group); {
			end := start + 1
			if runtime.Provider() != filter.ProviderUFW {
				end = min(start+filter.MaxAtomicExpansion, len(group))
			}
			batch := group[start:end]
			start = end
			changes := make([]filter.RuleChange, 0, len(batch))
			for _, index := range batch {
				change := items[index].change
				change.CommandOnly = true
				changes = append(changes, change)
			}
			var plan filter.CommandBatch
			err := ctx.Err()
			if err == nil {
				plan, err = runtime.BuildCommands(items[batch[0]].snapshot, changes)
			}
			if err == nil {
				plan.CommandOnly = true
				err = runtime.RunCommands(ctx, plan)
			}
			if err != nil {
				for _, index := range batch {
					failures[index] = err
				}
				continue
			}
			applied = append(applied, appliedBatch{plan: plan, sourceIndexes: batch})
		}
	}
	if len(applied) == 0 {
		return failures
	}
	plans := make([]filter.CommandBatch, len(applied))
	for index, batch := range applied {
		plans[index] = batch.plan
	}
	saved := persistFirewallRuleBatches(ctx, runtime, plans)
	for index, batch := range applied {
		for _, item := range batch.sourceIndexes {
			failures[item] = saved[index]
		}
	}
	return failures
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

func (s *FirewallService) createRules(ctx context.Context, request dto.FirewallRuleCreate, t *task.Task, ruleState map[string]filter.RuleSet) (result dto.FirewallRuleCreateResponse, taskErr error) {
	var firstFailure error
	defer func() {
		summary := i18n.GetMsgWithMap("FirewallCreateRulesResult", map[string]interface{}{
			"succeeded": result.Succeeded, "failed": result.Failed, "skipped": result.Skipped,
		})
		if t != nil {
			t.Log(summary)
		}
		if result.Failed > 0 {
			taskErr = errors.Join(errors.New(summary), taskErr, firstFailure)
		}
	}()
	type createItem struct {
		index, part, count int
		request            dto.FirewallRuleCreateItem
	}
	describe := func(rule filter.FirewallRule) string {
		return fmt.Sprintf("%s %s %s:%s -> %s:%s %s", rule.Scope.Family, rule.Protocol,
			rule.SourceAddress, rule.SourcePort, rule.DestinationAddress, rule.DestinationPort, rule.Action)
	}
	record := func(item createItem, status string, err error) {
		rule := item.request.Rule
		label := fmt.Sprintf("[%d/%d]", item.index+1, len(request.Items))
		if item.count > 1 {
			label += fmt.Sprintf("[%d/%d]", item.part+1, item.count)
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
				Index: item.index, Status: status, Rule: rule, Error: err.Error(),
			})
		}
	}
	selected, err := s.selectedProvider(ctx)
	if err != nil {
		for index, item := range request.Items {
			record(createItem{index: index, request: item}, "skipped", err)
		}
		return result, err
	}
	runtime, err := s.firewallAdapter(selected)
	if err != nil {
		for index, item := range request.Items {
			record(createItem{index: index, request: item}, "skipped", err)
		}
		return result, err
	}
	families, err := loadFirewallFamilies()
	if err != nil {
		return result, err
	}
	preparedItems := make([]*createItem, 0, len(request.Items))
	var stop error
	for index, item := range request.Items {
		if stop == nil {
			stop = ctx.Err()
		}
		if stop != nil {
			record(createItem{index: index, request: item}, "skipped", stop)
			continue
		}
		item.Rule.OrderIndex = nil
		rules, err := expandFirewallCreateRule(item, selected)
		if err != nil {
			status := "failed"
			if item.SourceKind == constant.FirewallRuleSourceImported && errors.Is(err, filter.ErrUnsupportedScope) {
				status = "skipped"
			}
			record(createItem{index: index, request: item}, status, err)
			continue
		}
		if item.SourceKind == constant.FirewallRuleSourceImported && t != nil {
			t.Log(i18n.GetMsgWithMap("FirewallImportRuleConversion", map[string]interface{}{
				"index": index + 1, "total": len(request.Items), "source": item.Rule.Scope.Provider,
				"target": selected, "rule": describe(item.Rule), "count": len(rules),
			}))
		}
		for part, rule := range rules {
			child := item
			child.Rule = rule
			entry := createItem{index: index, part: part, count: len(rules), request: child}
			if (request.Initialize || item.SourceKind == constant.FirewallRuleSourceImported) && rule.Scope.Family == filter.FamilyIPv6 && !slices.Contains(families, constant.FirewallFamilyIPv6) {
				record(entry, "skipped", fmt.Errorf("IPv6 firewall support is disabled"))
				continue
			}
			prepared, err := prepareFirewallCreateRule(ctx, runtime, child)
			if err != nil {
				status := "failed"
				if item.SourceKind == constant.FirewallRuleSourceImported && (errors.Is(err, filter.ErrUnsupportedScope) || errors.Is(err, filter.ErrInvalidRule)) {
					status = "skipped"
				}
				record(entry, status, err)
				continue
			}
			entry.request = prepared
			if child.ParseStatus == "" || child.ParseStatus == filter.ParseStatusSupported {
				entry.request.Raw = ""
			}
			preparedItems = append(preparedItems, &entry)
		}
	}
	if len(preparedItems) == 0 {
		return result, stop
	}
	if ruleState == nil {
		ruleState = make(map[string]filter.RuleSet)
	}
	if len(ruleState) == 0 {
		if err := loadFirewallRuleState(ctx, runtime, ruleState); err != nil {
			return result, err
		}
	}
	existing := make(map[string]bool)
	ufwActions := make(map[string]map[filter.Action]bool)
	for _, snapshot := range ruleState {
		for _, observed := range snapshot.Rules {
			key, err := firewallRuleDuplicateKey(observed, false)
			if err != nil {
				return result, err
			}
			existing[key] = true
			if selected == filter.ProviderUFW && observed.ParseStatus == filter.ParseStatusSupported {
				matchKey, err := filter.RuleMatchKey(observed.Rule)
				if err != nil {
					return result, err
				}
				if ufwActions[matchKey] == nil {
					ufwActions[matchKey] = make(map[filter.Action]bool)
				}
				ufwActions[matchKey][observed.Rule.Action] = true
			}
			if selected == filter.ProviderFirewalld && observed.ParseStatus == filter.ParseStatusSupported {
				families := []filter.Family{observed.Rule.Scope.Family}
				if families[0] == filter.FamilyInet {
					families = append(families, filter.FamilyIPv4, filter.FamilyIPv6)
				}
				for _, family := range families {
					candidate := observed
					candidate.Rule.Scope.Family = family
					key, err := firewallRuleDuplicateKey(candidate, true)
					if err != nil {
						return result, err
					}
					existing[key] = true
				}
			}
		}
	}
	mergedItems := make([]*createItem, 0, len(preparedItems))
	pending := make(map[string]bool)
	priority := 0
	for _, item := range preparedItems {
		converted := item.request.SourceKind == constant.FirewallRuleSourceImported && selected == filter.ProviderFirewalld && request.Items[item.index].Rule.Scope.Provider != selected
		observed := filter.ObservedRule{Rule: item.request.Rule, Raw: item.request.Raw, ParseStatus: item.request.ParseStatus}
		key, err := firewallRuleDuplicateKey(observed, converted)
		if err != nil {
			return result, err
		}
		if pending[key] {
			record(*item, "skipped", buserr.New("ErrRecordExist"))
			continue
		}
		if converted {
			priority++
		}
		if existing[key] {
			pending[key] = true
			record(*item, "skipped", buserr.New("ErrRecordExist"))
			continue
		}
		if selected == filter.ProviderUFW && item.request.ParseStatus == filter.ParseStatusSupported {
			matchKey, err := filter.RuleMatchKey(item.request.Rule)
			if err != nil {
				return result, err
			}
			actions := ufwActions[matchKey]
			if len(actions) > 1 || len(actions) == 1 && !actions[item.request.Rule.Action] {
				record(*item, "failed", buserr.New("ErrFirewallRuleConflict"))
				continue
			}
			ufwActions[matchKey] = map[filter.Action]bool{item.request.Rule.Action: true}
		}
		if converted {
			if priority > 32767 {
				return result, fmt.Errorf("%w: firewalld import order exceeds the priority range", filter.ErrExpansionLimit)
			}
			value := priority
			item.request.Rule.Priority = &value
			prepared, err := prepareFirewallCreateRule(ctx, runtime, item.request)
			if err != nil {
				status := "failed"
				if errors.Is(err, filter.ErrUnsupportedScope) || errors.Is(err, filter.ErrInvalidRule) {
					status = "skipped"
				}
				record(*item, status, err)
				continue
			}
			item.request = prepared
			actualKey, err := firewallRuleDuplicateKey(filter.ObservedRule{Rule: prepared.Rule, ParseStatus: prepared.ParseStatus}, false)
			if err != nil {
				return result, err
			}
			pending[key] = true
			if existing[actualKey] || pending[actualKey] {
				record(*item, "skipped", buserr.New("ErrRecordExist"))
				continue
			}
			pending[actualKey] = true
		}
		pending[key] = true
		mergedItems = append(mergedItems, item)
	}
	for _, item := range mergedItems {
		item.request.Rule.UUID = uuid.NewString()
	}
	creates := mergedItems
	rawScopes := make([]filter.Scope, 0)
	rawByScope := make(map[string]filter.RuleSet)
	for _, item := range creates {
		if item.request.Raw != "" {
			scope := item.request.Rule.Scope
			if snapshot, exists := ruleState[scope.Key()]; exists {
				rawByScope[scope.Key()] = snapshot
			} else {
				rawScopes = append(rawScopes, scope)
			}
		}
	}
	rawSnapshots, err := listFirewallRuleScopes(runtime, ctx, rawScopes)
	if err != nil {
		return result, err
	}
	for _, snapshot := range rawSnapshots {
		rawByScope[snapshot.Scope.Key()] = snapshot
	}
	changes := make([]firewallRuleBatchItem, 0, len(creates))
	for _, item := range creates {
		rule := &item.request.Rule
		snapshot := filter.RuleSet{Scope: rule.Scope}
		if current, exists := rawByScope[rule.Scope.Key()]; exists {
			snapshot = current
		}
		changes = append(changes, firewallRuleBatchItem{
			snapshot: snapshot,
			change:   filter.RuleChange{Operation: filter.ChangeCreate, After: rule, Append: true, Raw: item.request.Raw},
		})
	}
	descriptions := make([]model.CommonDescription, 0, len(creates))
	for index, err := range executeFirewallRuleBatches(ctx, runtime, changes) {
		item := creates[index]
		if err != nil {
			if errors.Is(err, filterfirewalld.ErrAlreadyEnabled) || errors.Is(err, filterufw.ErrAlreadyEnabled) {
				record(*item, "skipped", err)
			} else if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				record(*item, "skipped", err)
				stop = err
			} else {
				record(*item, "failed", fmt.Errorf("%s: %w", i18n.GetMsgByKey("FirewallCreateRuleExecutionFailed"), err))
			}
			continue
		}
		scope := item.request.Rule.Scope
		snapshot := ruleState[scope.Key()]
		snapshot.Scope = scope
		snapshot.Rules = append(snapshot.Rules, filter.ObservedRule{Rule: item.request.Rule, Raw: item.request.Raw, ParseStatus: item.request.ParseStatus})
		ruleState[scope.Key()] = snapshot
		if item.request.Rule.Description != "" {
			observed := filter.ObservedRule{Rule: item.request.Rule, Raw: item.request.Raw, ParseStatus: item.request.ParseStatus}
			id, err := filter.DescriptionID(observed)
			if err != nil {
				return result, err
			}
			descriptions = append(descriptions, model.CommonDescription{ID: id, Type: "firewall", Description: item.request.Rule.Description})
		}
		record(*item, "succeeded", nil)
	}
	if len(descriptions) > 0 {
		persistCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		err := settingRepo.SaveDescriptions(persistCtx, descriptions)
		cancel()
		if err != nil {
			if t != nil {
				t.Logf("save firewall rule descriptions: %v", err)
			}
			global.LOG.Warnf("save firewall rule descriptions: %v", err)
		}
	}
	return result, stop
}

func validateFirewallCreateBatch(request dto.FirewallRuleCreate, provider filter.Provider) error {
	if len(request.Items) == 0 && !request.Initialize {
		return filter.ErrInvalidRule
	}
	count := 0
	importing := true
	for _, item := range request.Items {
		importing = importing && item.SourceKind == constant.FirewallRuleSourceImported
		rules, err := expandFirewallCreateRule(item, provider)
		if errors.Is(err, filter.ErrExpansionLimit) {
			return err
		}
		count += len(rules)
	}
	if !importing && count > filter.MaxAtomicExpansion {
		return filter.ErrExpansionLimit
	}
	return nil
}

func expandFirewallCreateRule(item dto.FirewallRuleCreateItem, provider filter.Provider) ([]filter.FirewallRule, error) {
	rule := item.Rule
	if item.Raw != "" && item.ParseStatus != "" && item.ParseStatus != filter.ParseStatusSupported {
		if item.SourceKind != constant.FirewallRuleSourceImported || rule.Scope.Provider != provider {
			return nil, fmt.Errorf("%w: native rule can only be imported into its original backend", filter.ErrUnsupportedScope)
		}
		return []filter.FirewallRule{rule}, rule.Scope.ValidateMVP()
	}
	source := rule.Scope.Provider
	if item.SourceKind != constant.FirewallRuleSourceImported || source == "" || source == provider {
		return filter.ExpandAtomicRules(applySelectedProviderScopeDefaults(rule, provider))
	}
	if rule.NativeKind == filter.NativeKindOpaque || rule.NativeKind == filter.NativeKindZoneService || rule.NativeKind == filter.NativeKindUFWApplication {
		return nil, fmt.Errorf("%w: native rule cannot be converted to %s", filter.ErrUnsupportedScope, provider)
	}
	if provider == filter.ProviderFirewalld {
		priority := 0
		rule.Priority = &priority
	}
	families := []filter.Family{rule.Scope.Family}
	if families[0] == filter.FamilyInet && provider != filter.ProviderFirewalld {
		switch {
		case strings.Contains(rule.SourceAddress, ":") || strings.Contains(rule.DestinationAddress, ":") || rule.Protocol == "icmpv6":
			families = []filter.Family{filter.FamilyIPv6}
		case rule.SourceAddress != "" || rule.DestinationAddress != "":
			families = []filter.Family{filter.FamilyIPv4}
		default:
			families = []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6}
		}
	}
	var result []filter.FirewallRule
	for _, family := range families {
		converted := rule
		converted.Scope = filter.Scope{Provider: provider, Family: family, Direction: filter.DirectionInput}
		if (source == filter.ProviderIptables || source == filter.ProviderNftables) && (provider == filter.ProviderIptables || provider == filter.ProviderNftables) {
			converted.Scope.Chain = rule.Scope.Chain
		}
		converted.NativeKind, converted.OrderBucket = "", ""
		converted.UUID, converted.OrderIndex = "", nil
		if provider != filter.ProviderFirewalld {
			converted.Priority = nil
		}
		expanded, err := filter.ExpandAtomicRules(applySelectedProviderScopeDefaults(converted, provider))
		if err != nil {
			return nil, err
		}
		result = append(result, expanded...)
	}
	return result, nil
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
		switch {
		case strings.EqualFold(strings.TrimSpace(rule.Protocol), "icmpv6"), strings.Contains(rule.SourceAddress, ":"), strings.Contains(rule.DestinationAddress, ":"):
			scope.Family = filter.FamilyIPv6
		case selected == filter.ProviderFirewalld:
			scope.Family = filter.FamilyInet
		default:
			scope.Family = filter.FamilyIPv4
		}
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

func prepareFirewallCreateRule(ctx context.Context, runtime filter.Adapter, request dto.FirewallRuleCreateItem) (dto.FirewallRuleCreateItem, error) {
	request.Rule.Description = strings.TrimSpace(request.Rule.Description)
	selected := runtime.Provider()
	if (selected == filter.ProviderIptables || selected == filter.ProviderNftables) && request.Rule.Scope.Chain != "" && request.Rule.Scope.Chain != filter.IptablesInputChain && request.Rule.Description != "" {
		return dto.FirewallRuleCreateItem{}, filter.ErrRuleOperation
	}
	if request.Raw != "" && request.ParseStatus != "" && request.ParseStatus != filter.ParseStatusSupported {
		if request.Rule.Scope.Provider != selected {
			return dto.FirewallRuleCreateItem{}, filter.ErrUnsupportedScope
		}
		plan, err := runtime.BuildCommands(filter.RuleSet{Scope: request.Rule.Scope}, []filter.RuleChange{{Operation: filter.ChangeCreate, After: &request.Rule, Raw: request.Raw, CommandOnly: true, Append: true}})
		if err != nil {
			return dto.FirewallRuleCreateItem{}, err
		}
		description := request.Rule.Description
		request.Rule = plan.Rules[0].Expected.Rule
		request.Rule.Description = description
		if plan.Rules[0].Expected.ParseStatus == filter.ParseStatusSupported {
			request.ParseStatus, request.Raw = filter.ParseStatusSupported, ""
		}
		return request, nil
	}
	rule, err := filter.NormalizeRule(applySelectedProviderScopeDefaults(request.Rule, selected))
	if err != nil {
		return dto.FirewallRuleCreateItem{}, err
	}
	if rule.Scope.Provider != selected {
		return dto.FirewallRuleCreateItem{}, fmt.Errorf("%w: selected provider is %s", filter.ErrInvalidRule, selected)
	}
	rule, err = prepareFirewallBackendRule(ctx, runtime, rule)
	if err != nil {
		return dto.FirewallRuleCreateItem{}, err
	}
	rule.UUID = ""
	request.Rule = rule
	request.ParseStatus = filter.ParseStatusSupported
	if request.SourceKind == "" {
		request.SourceKind = constant.FirewallRuleSourceUser
	}
	return request, nil
}

func readFirewallRuleScopes(client filter.Adapter, ctx context.Context, scopes []filter.Scope) ([]filter.RuleSet, error) {
	snapshots, err := listFirewallRuleScopes(client, ctx, scopes)
	if err != nil || len(snapshots) == 0 {
		return snapshots, err
	}
	ports, err := loadFirewallPortWhiteList()
	if err != nil {
		return nil, err
	}
	for index := range snapshots {
		snapshots[index], err = filter.ProtectRuleSet(snapshots[index], ports)
		if err != nil {
			return nil, err
		}
	}
	return snapshots, nil
}

func firewallTaskName(operation, subsystem, backend string) string {
	name := i18n.GetMsgByKey(subsystem)
	if backend != "" {
		name += " · " + backend
	}
	key := "FirewallRule" + operation
	if operation == task.TaskExec {
		key = "FirewallTaskInitialize"
	}
	return i18n.GetMsgWithMap(key, map[string]interface{}{"name": name})
}

func closeUnstartedFirewallTask(t *task.Task) {
	if cancel, ok := global.LoadTaskCancel(t.TaskID); ok {
		cancel()
	}
	global.RemoveTaskCancel(t.TaskID)
	if closer, ok := t.Logger.Out.(io.Closer); ok {
		_ = closer.Close()
	}
}

func loadFirewallRuleState(ctx context.Context, runtime filter.Adapter, state map[string]filter.RuleSet) error {
	for _, group := range firewallScopeReadGroups(filter.ManagedInputScopes(runtime.Provider())) {
		snapshots, err := listFirewallRuleScopes(runtime, ctx, group)
		if errors.Is(err, filter.ErrFamilyUnavailable) {
			continue
		}
		if err != nil {
			return err
		}
		for _, snapshot := range snapshots {
			missing := false
			for _, notice := range snapshot.Notices {
				missing = missing || notice.Code == filter.ScopeNoticeManagedScopeMissing || notice.Code == filter.ScopeNoticeManagedScopeInactive
			}
			if !missing {
				state[snapshot.Scope.Key()] = snapshot
			}
		}
	}
	return nil
}

func firewallRuleDuplicateKey(observed filter.ObservedRule, converted bool) (string, error) {
	if converted && observed.ParseStatus == filter.ParseStatusSupported && observed.Rule.Scope.Provider == filter.ProviderFirewalld {
		priority := 0
		observed.Rule.Priority = &priority
		observed.Rule.NativeKind, observed.Rule.OrderBucket = filter.NativeKindRichRule, ""
	}
	key, err := filter.DescriptionID(observed)
	if converted {
		key = "converted:" + key
	}
	return key, err
}

func (s *FirewallService) syncPortWhitelist(ctx context.Context, ruleState map[string]filter.RuleSet) error {
	firewallWhitelistMu.Lock()
	defer firewallWhitelistMu.Unlock()
	ports, err := loadFirewallPortWhiteList()
	if err != nil {
		return err
	}
	return s.applyPortWhitelist(ctx, ports, ruleState)
}

func (s *FirewallService) applyPortWhitelist(ctx context.Context, ports []firewall.PortWhitelist, ruleState map[string]filter.RuleSet, families ...filter.Family) error {
	required, err := firewall.RequiredPortWhitelist(ports)
	if err != nil {
		return err
	}
	provider, err := s.selectedProvider(ctx)
	if err != nil {
		return err
	}
	direct := provider == filter.ProviderIptables || provider == filter.ProviderNftables
	if !direct {
		client, err := s.baseClient()
		if err != nil {
			return err
		}
		active, err := client.Status()
		if err != nil || !active {
			return err
		}
	}
	runtime, err := s.firewallAdapter(provider)
	if err != nil {
		return err
	}
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	if ruleState == nil {
		ruleState = make(map[string]filter.RuleSet)
		if err := loadFirewallRuleState(ctx, runtime, ruleState); err != nil {
			return err
		}
	}
	keys := make(map[string]bool)
	for _, snapshot := range ruleState {
		for _, observed := range snapshot.Rules {
			if observed.ParseStatus != filter.ParseStatusSupported {
				continue
			}
			families := []filter.Family{observed.Rule.Scope.Family}
			if families[0] == filter.FamilyInet {
				families = []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6}
			}
			for _, family := range families {
				rule := observed.Rule
				rule.Scope.Family = family
				key, err := filter.RuleMatchKey(rule)
				if err == nil {
					keys[string(rule.Action)+key] = true
				}
			}
		}
	}
	rules := whitelistRules(provider, firewall.ExpandPortWhitelist(customWhitelist(ports)), firewall.ExpandPortWhitelist(required))
	var changes []firewallRuleBatchItem
	for _, rule := range rules {
		if len(families) > 0 && !slices.Contains(families, rule.Scope.Family) {
			continue
		}
		item, err := prepareFirewallCreateRule(ctx, runtime, dto.FirewallRuleCreateItem{Rule: rule, SourceKind: constant.FirewallRuleSourceSecurity})
		if err != nil {
			return err
		}
		rule = item.Rule
		snapshot, available := ruleState[rule.Scope.Key()]
		if !available {
			continue
		}
		key, err := filter.RuleMatchKey(rule)
		if err != nil {
			return err
		}
		key = string(rule.Action) + key
		if keys[key] {
			continue
		}
		keys[key] = true
		rule.UUID = uuid.NewString()
		if provider != filter.ProviderFirewalld {
			position := int64(1)
			rule.OrderIndex = &position
		}
		changes = append(changes, firewallRuleBatchItem{snapshot: snapshot, change: filter.RuleChange{Operation: filter.ChangeCreate, After: &rule}})
	}
	var failures []error
	for _, err := range executeFirewallRuleBatches(ctx, runtime, changes) {
		if err != nil && !errors.Is(err, filterfirewalld.ErrAlreadyEnabled) && !errors.Is(err, filterufw.ErrAlreadyEnabled) {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

func whitelistRules(provider filter.Provider, ports, required []firewall.SystemPort) []filter.FirewallRule {
	rules := make([]filter.FirewallRule, 0, len(ports)+len(required))
	for _, port := range required {
		rule := firewall.RuleForSystemPort(provider, firewall.SystemPort(port))
		if provider == filter.ProviderIptables || provider == filter.ProviderNftables {
			rule.Scope.Chain = filter.BasicBeforeChain
		}
		rules = append(rules, rule)
	}
	for _, port := range ports {
		rules = append(rules, firewall.RuleForSystemPort(provider, firewall.SystemPort(port)))
	}
	return rules
}

func customWhitelist(entries []firewall.PortWhitelist) []firewall.PortWhitelist {
	result := make([]firewall.PortWhitelist, 0, len(entries))
	for _, entry := range entries {
		if entry.Type == "" {
			result = append(result, entry)
		}
	}
	return result
}

func loadForwardingFirewallOverview(manager forwarding.Adapter, families []string) (dto.FirewallSubsystemStatus, error) {
	ipv4, ipv4Err := loadForwardingFamilyInfo(manager, constant.FirewallFamilyIPv4)
	if len(families) == 1 {
		return dto.FirewallSubsystemStatus{IPv6Enabled: false, IsInit: ipv4.Initialized, IsBind: ipv4.Bound, IPv4: ipv4}, ipv4Err
	}
	ipv6, ipv6Err := loadForwardingFamilyInfo(manager, constant.FirewallFamilyIPv6)
	if ipv6.Available {
		interfaces, err := forwarding.IPv6RAInterfaces(os.ReadFile)
		var pathError *os.PathError
		switch {
		case errors.As(err, &pathError) && errors.Is(err, os.ErrNotExist) && pathError.Path == "/proc/net/if_inet6":
			ipv6.Available, ipv6.Initialized, ipv6.Bound = false, false, false
		case err != nil:
			ipv6.Reason = "ipv6_ra_check_failed"
			ipv6.Bound = false
		case len(interfaces) > 0:
			ipv6.Reason, ipv6.RAInterfaces = "ipv6_ra_required", interfaces
			ipv6.Bound = false
		case !ipv6.Bound:
			ipv6.Reason = "ipv6_forwarding_not_enabled"
		}
	}
	return dto.FirewallSubsystemStatus{IPv6Enabled: len(families) > 1, IsInit: ipv4.Initialized || ipv6.Initialized, IsBind: ipv4.Bound || ipv6.Bound, IPv4: ipv4, IPv6: ipv6}, errors.Join(ipv4Err, ipv6Err)
}

func loadForwardingFamilyInfo(manager forwarding.Adapter, family string) (dto.FirewallBackendFamilyStatus, error) {
	initialized, bound, partial, err := manager.FamilyState(family)
	if family == constant.FirewallFamilyIPv6 && errors.Is(err, filter.ErrFamilyUnavailable) {
		return dto.FirewallBackendFamilyStatus{}, nil
	}
	available := err == nil
	if manager.Name() == constant.FirewallProviderIptables && family == constant.FirewallFamilyIPv6 {
		commands, commandErr := lifecycle.ResolveIptablesCommands()
		available = available && commandErr == nil && commands.IPv6Available()
	}
	return dto.FirewallBackendFamilyStatus{Available: available, Initialized: initialized, Bound: bound, Partial: partial}, err
}

func (s *ForwardingService) operateRules(ctx context.Context, request dto.ForwardRuleOperate, t *task.Task, initialize bool, families ...string) (resultErr error) {
	defer func() { resultErr = errors.Join(resultErr, ctx.Err()) }()
	forwardingMutationMu.Lock()
	defer forwardingMutationMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	var allowed []string
	if initialize || request.Import {
		var err error
		allowed, err = loadFirewallFamilies()
		if err != nil {
			return err
		}
	}
	skippedIPv6 := 0
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
			if len(allowed) > 0 && rule.Family == forwarding.FamilyIPv6 && !slices.Contains(allowed, forwarding.FamilyIPv6) {
				skippedIPv6++
				continue
			}
			groups[len(groups)-1].rules = append(groups[len(groups)-1].rules, rule)
		}
	}
	logFirewallIPv6Skipped(t, "forwarding", skippedIPv6)
	groups = slices.DeleteFunc(groups, func(group operationBatch) bool { return len(group.rules) == 0 })
	if !initialize && len(groups) == 0 && skippedIPv6 > 0 {
		return nil
	}
	manager, err := s.clientFactory(ctx)
	if err != nil {
		return err
	}
	if initialize {
		if err := initializeForwarding(manager, allowed); err != nil {
			return err
		}
		if !slices.Contains(allowed, forwarding.FamilyIPv6) && t != nil {
			t.Logf("IPv6 firewall support is disabled; skipping IPv6 forwarding chain initialization")
		}
		for _, family := range families {
			if !slices.Contains(allowed, family) {
				continue
			}
			if err := manager.OperateFamily(family, true); err != nil {
				return err
			}
		}
	}
	initialized, _, err := manager.InitStatus()
	if err != nil {
		return err
	}
	if !initialized {
		return filter.ErrProviderUnavailable
	}
	checkedFamilies := make(map[string]bool)
	for _, group := range groups {
		for _, rule := range group.rules {
			if checkedFamilies[rule.Family] {
				continue
			}
			checkedFamilies[rule.Family] = true
			ready, _, err := manager.FamilyStatus(rule.Family)
			if err != nil {
				return err
			}
			if !ready {
				if !initialize {
					return fmt.Errorf("%s forwarding chains are not initialized", rule.Family)
				}
				if err := manager.OperateFamily(rule.Family, true); err != nil {
					return err
				}
			}
		}
	}
	stored, err := manager.List()
	if err != nil {
		return err
	}
	stored = slices.Clone(stored)
	counts := make(map[string]int, len(stored))
	for index, rule := range stored {
		normalized, err := forwarding.NormalizeRule(rule)
		if err != nil {
			return err
		}
		stored[index] = normalized
		counts[normalized.Identity()]++
	}
	succeeded, failed, skipped := 0, 0, 0
	defer func() {
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
	var failures []error
	var interfaces map[string]bool
	for _, group := range groups {
		if group.operation != forwarding.OperationAdd {
			continue
		}
		for _, rule := range group.rules {
			if rule.Interface == "" {
				continue
			}
			if counts[rule.Identity()] > 0 {
				continue
			}
			if interfaces == nil {
				available, err := net.Interfaces()
				if err != nil {
					return fmt.Errorf("load forwarding interfaces: %w", err)
				}
				interfaces = make(map[string]bool, len(available))
				for _, item := range available {
					interfaces[item.Name] = true
				}
			}
			if !interfaces[rule.Interface] {
				err := buserr.WithMap("ErrForwardInterfaceNotFound", map[string]interface{}{"name": rule.Interface}, nil)
				failures = append(failures, err)
				record(group.operation, rule, "failed", err)
			}
		}
	}
	if len(failures) > 0 {
		return errors.Join(failures...)
	}
	isEdit := len(request.Rules) == 2 && len(groups) == 2 && groups[0].operation == forwarding.OperationRemove && groups[1].operation == forwarding.OperationAdd
	if isEdit {
		old := make(map[string]bool, len(groups[0].rules))
		for _, rule := range groups[0].rules {
			old[rule.Identity()] = true
		}
		unchanged := len(old) == len(groups[1].rules)
		duplicate := false
		next := make(map[string]bool, len(groups[1].rules))
		for _, rule := range groups[1].rules {
			key := rule.Identity()
			next[key] = true
			unchanged = unchanged && old[key]
			if counts[key] > 0 && !old[key] {
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
		for _, rule := range groups[0].rules {
			if counts[rule.Identity()] == 0 {
				record(forwarding.OperationRemove, rule, "failed", filter.ErrRuleStale)
				return filter.ErrRuleStale
			}
		}
		removals := make([]forwarding.Rule, 0, len(groups[0].rules))
		for _, rule := range groups[0].rules {
			if !next[rule.Identity()] {
				removals = append(removals, rule)
			}
		}
		groups[0], groups[1] = groups[1], operationBatch{operation: forwarding.OperationRemove, rules: removals}
	}
	for _, group := range groups {
		byFamily := make(map[string][]forwarding.Rule, 2)
		seen := make(map[string]bool, len(group.rules))
		for _, rule := range group.rules {
			key := rule.Identity()
			exists := counts[key] > 0
			if seen[key] || (group.operation == forwarding.OperationAdd && exists) {
				record(group.operation, rule, "skipped", buserr.New("ErrRecordExist"))
				continue
			}
			if group.operation == forwarding.OperationRemove && !exists {
				failures = append(failures, filter.ErrRuleStale)
				record(group.operation, rule, "failed", filter.ErrRuleStale)
				continue
			}
			seen[key] = true
			copies := 1
			if group.operation == forwarding.OperationRemove && manager.Name() == constant.FirewallProviderIptables {
				copies = counts[key]
			}
			for i := 0; i < copies; i++ {
				byFamily[rule.Family] = append(byFamily[rule.Family], rule)
			}
		}
		for _, family := range []string{forwarding.FamilyIPv4, forwarding.FamilyIPv6} {
			rules := byFamily[family]
			if len(rules) == 0 {
				continue
			}
			err := ctx.Err()

			if err == nil {
				if group.operation == forwarding.OperationAdd {
					err = manager.CreateRules(ctx, rules)
				} else {
					err = manager.DeleteRules(ctx, rules)
				}
			}
			if err != nil {
				failures = append(failures, err)
				for _, rule := range rules {
					record(group.operation, rule, "failed", err)
				}
				continue
			}
			for _, rule := range rules {
				if group.operation == forwarding.OperationAdd {
					counts[rule.Identity()]++
					stored = append(stored, rule)
				} else {
					delete(counts, rule.Identity())
					stored = slices.DeleteFunc(stored, func(item forwarding.Rule) bool { return item.Identity() == rule.Identity() })
				}
				record(group.operation, rule, "succeeded", nil)
			}
		}
		if len(failures) > 0 && (isEdit || group.operation == forwarding.OperationRemove) {
			break
		}
	}
	return errors.Join(errors.Join(failures...), persistForwardingRules(manager, stored))
}

func initializeForwarding(manager forwarding.Adapter, families []string) error {
	if err := settingRepo.UpdateOrCreate(constant.FirewallForwardingBackendKey, manager.Name()); err != nil {
		return err
	}
	if !slices.Contains(families, forwarding.FamilyIPv6) {
		global.LOG.Info("IPv6 firewall support is disabled; skipping IPv6 forwarding chain initialization")
	}
	if err := manager.Enable(families...); err != nil {
		return err
	}
	return settingRepo.UpdateOrCreate(constant.FirewallForwardingInitializedKey, constant.StatusEnable)
}

func newForwardingService() *ForwardingService {
	return &ForwardingService{
		clientFactory: newForwardingAdapter,
	}
}

func newForwardingAdapter(ctx context.Context) (forwarding.Adapter, error) {
	selected, _ := settingRepo.GetValueByKey(constant.FirewallForwardingBackendKey)
	selected = strings.TrimSpace(selected)
	if selected == "" {
		selected = constant.FirewallProviderIptables
	}
	return newForwardingAdapterFor(ctx, selected)
}

func newForwardingAdapterFor(ctx context.Context, backend string) (forwarding.Adapter, error) {
	client, err := lifecycle.NewClient(backend)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: selected forwarding backend %s: %w",
			errForwardingBackendUnavailable, backend, err,
		)
	}
	switch client.Name() {
	case constant.FirewallProviderIptables:
		return forwarding.NewIptables(ctx, client.Name()), nil
	case constant.FirewallProviderNftables:
		return forwarding.NewNftables(ctx), nil
	default:
		return nil, errForwardingBackendUnavailable
	}
}

func (s *DockerPortGuardService) runtimeForDocker(ctx context.Context) (dockerfirewall.Runtime, string, error) {
	if s.runtime != nil {
		return s.runtime, selectedDockerFirewallBackend(constant.FirewallProviderIptables), nil
	}
	cli, err := s.client()
	if err != nil {
		return nil, "", buserr.WithDetail("ErrDockerFailed", err.Error(), err)
	}
	defer cli.Close()
	info, err := cli.Info(ctx)
	if err != nil {
		return nil, "", buserr.WithDetail("ErrDockerFailed", err.Error(), err)
	}
	backend := selectedDockerFirewallBackend(dockerFirewallBackend(info))
	if backend != constant.FirewallProviderIptables && backend != constant.FirewallProviderNftables {
		return nil, backend, fmt.Errorf("Docker firewall backend %q is not supported", backend)
	}
	return s.guardRuntime(ctx, backend), backend, nil
}

func (s *DockerPortGuardService) guardRuntime(ctx context.Context, backend string) dockerfirewall.Runtime {
	if s.runtimeForBackend != nil {
		return s.runtimeForBackend(ctx, backend)
	}
	if s.runtime != nil {
		return s.runtime
	}
	return newDockerFirewallRuntime(ctx, backend)
}

func newDockerFirewallRuntime(ctx context.Context, backend string) dockerfirewall.Runtime {
	if backend == constant.FirewallProviderNftables {
		return dockerfirewall.NewNftables(ctx)
	}
	return dockerfirewall.NewIptables(ctx)
}

func selectedDockerFirewallBackend(fallback string) string {
	selected, _ := settingRepo.GetValueByKey(constant.FirewallDockerBackendKey)
	selected = strings.ToLower(strings.TrimSpace(selected))
	if selected == constant.FirewallProviderIptables || selected == constant.FirewallProviderNftables {
		return selected
	}
	fallback = strings.ToLower(strings.TrimSpace(fallback))
	if fallback == constant.FirewallProviderNftables {
		return fallback
	}
	return constant.FirewallProviderIptables
}

func dockerFirewallBackend(info system.Info) string {
	if info.FirewallBackend == nil || info.FirewallBackend.Driver == "" {
		return constant.FirewallProviderIptables
	}
	return strings.ToLower(info.FirewallBackend.Driver)
}

func dockerPortGuardPersistedEnabled() (bool, error) {
	status, err := settingRepo.GetValueByKey(constant.FirewallDockerPortGuardStatusKey)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return status == constant.StatusEnable, err
}

func newDockerPortGuardService() *DockerPortGuardService {
	return &DockerPortGuardService{
		client:  docker.NewDockerClient,
		version: dockerFirewallVersion,
	}
}

func dockerFirewallVersion(backend string) string {
	client, err := lifecycle.NewClient(backend)
	if err != nil {
		return "-"
	}
	version, err := client.Version()
	if err != nil || strings.TrimSpace(version) == "" {
		return "-"
	}
	return version
}

func firewallDockerActive() (bool, error) {
	if !cmd.Which("docker") {
		return false, nil
	}
	return controller.CheckActive("docker")
}

func restoreFirewalldDependents(ctx context.Context, reason string, restoreDocker bool, restoreForwarding func(context.Context) error, restoreDockerGuard func(context.Context) error) error {
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

func operateFirewallLifecycle(client lifecycle.Client, operation string, withDockerRestart bool, prepareStart func(lifecycle.Client) error, t *task.Task) error {
	run := func(operation, name string, action func() error) error {
		if t != nil {
			return runFirewallLifecycleAction(t, task.GetTaskName(name, operation, ""), action)
		}
		return action()
	}
	switch operation {
	case string(lifecycle.OperationStart):
		if err := run("Start", client.Name(), client.Start); err != nil {
			return err
		}
	case string(lifecycle.OperationRestart):
		if err := run("TaskRestart", client.Name(), client.Restart); err != nil {
			return err
		}
	case string(lifecycle.OperationStop):
		return stopFirewallLifecycle(client, withDockerRestart, nil, t)
	default:
		return fmt.Errorf("not supported operation: %s", operation)
	}
	var recoveryErrors []error
	if prepareStart != nil {
		err := prepareStart(client)
		if err == nil && client.Name() == constant.FirewallProviderFirewalld {
			err = lifecycleproviders.RemoveFirewalldSSHService()
		}
		if err != nil {
			recoveryErrors = append(recoveryErrors, fmt.Errorf("prepare firewall after %s: %w", operation, err))
		}
	}
	if withDockerRestart {
		if err := run("TaskRestart", "Docker", func() error { return controller.HandleRestart("docker") }); err != nil {
			recoveryErrors = append(recoveryErrors, &firewallDockerRestartError{Err: err})
		}
	}
	if client.Name() == constant.FirewallProviderFirewalld && operation == string(lifecycle.OperationStart) {
		if err := run("TaskRecover", "Fail2Ban", restoreFail2BanAfterFirewallStart); err != nil {
			recoveryErrors = append(recoveryErrors, err)
		}
	}
	if err := errors.Join(recoveryErrors...); err != nil {
		return &firewallCompletedOperationError{Operation: operation, Err: err}
	}
	return nil
}

func (c firewallLifecycleClient) Start() error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	return c.Client.Start()
}

func (c firewallLifecycleClient) Restart() error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	return c.Client.Restart()
}

func runFirewallLifecycleAction(t *task.Task, name string, action func() error) error {
	t.Log(i18n.GetWithName("TaskStart", name))
	started := time.Now()
	err := t.TaskCtx.Err()
	if err == nil {
		err = action()
	}
	t.LogWithStatus(fmt.Sprintf("%s (%.2fs)", name, time.Since(started).Seconds()), err)
	return err
}

func stopFirewallLifecycle(client lifecycle.Client, withDockerRestart bool, prepareStop func() error, t *task.Task) error {
	if client.Name() == constant.FirewallProviderFirewalld {
		if err := rememberFail2BanBeforeFirewallStop(); err != nil {
			return err
		}
	}
	if prepareStop != nil {
		if err := prepareStop(); err != nil {
			return err
		}
	}
	stop := client.Stop
	if t != nil {
		stop = func() error {
			return runFirewallLifecycleAction(t, task.GetTaskName(client.Name(), "Stop", ""), client.Stop)
		}
	}
	if err := stop(); err != nil {
		return err
	}
	if withDockerRestart {
		restart := func() error { return controller.HandleRestart("docker") }
		var err error
		if t != nil {
			err = runFirewallLifecycleAction(t, task.GetTaskName("Docker", "TaskRestart", ""), restart)
		} else {
			err = restart()
		}
		if err != nil {
			return &firewallDockerRestartError{Err: err}
		}
	}
	return nil
}

func (c firewallLifecycleClient) Stop() error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	return c.Client.Stop()
}

func rememberFail2BanBeforeFirewallStop() error {
	exists, err := controller.CheckExist("fail2ban.service")
	if err != nil {
		global.LOG.Warnf("check fail2ban.service installation before stopping the firewall failed: %v", err)
	}
	if !exists {
		return nil
	}
	active, err := controller.CheckActive("fail2ban.service")
	if err != nil {
		global.LOG.Warnf("check fail2ban.service status before stopping the firewall failed: %v", err)
	}
	if !active {
		return nil
	}
	if err := os.WriteFile(fail2BanRestoreWithFirewallMarker, nil, 0600); err != nil {
		return fmt.Errorf("mark Fail2Ban for restoration with the firewall: %w", err)
	}
	return nil
}

func restoreFail2BanAfterFirewallStart() error {
	if _, err := os.Stat(fail2BanRestoreWithFirewallMarker); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("load Fail2Ban restore marker after starting the firewall: %w", err)
	}
	if err := controller.HandleStart("fail2ban.service"); err != nil {
		return fmt.Errorf("restore Fail2Ban after starting the firewall: %w", err)
	}
	if err := os.Remove(fail2BanRestoreWithFirewallMarker); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clear Fail2Ban firewall restore status: %w", err)
	}
	return nil
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
	if provider != constant.FirewallProviderIptables && provider != constant.FirewallProviderNftables {
		return fmt.Errorf("filter chain operations are not supported for %s", provider)
	}
	operation := firewall.BaseOperation(request.Operate)
	var ports []firewall.PortWhitelist
	if operation == firewall.BaseOperationInit {
		var err error
		ports, err = LoadRequiredFirewallPortWhiteList()
		if err != nil {
			return err
		}
	}
	families, err := loadFirewallFamilies()
	if err != nil {
		return err
	}
	if operation == firewall.BaseOperationInit && !slices.Contains(families, constant.FirewallFamilyIPv6) {
		global.LOG.Info("IPv6 firewall support is disabled; skipping IPv6 host chain initialization")
	}
	if operation == firewall.BaseOperationBind {
		bound := false
		for _, family := range families {
			status, err := loadSystemFirewallFamilyInfo(provider, family)
			if err != nil {
				return err
			}
			if !status.Initialized {
				continue
			}
			if provider == constant.FirewallProviderNftables {
				err = nftables_helper.OperateFamily(filter.Family(family), false, nil)
			} else {
				err = iptables_helper.OperateFamily(family, false, nil)
			}
			if err != nil {
				return err
			}
			bound = true
		}
		if !bound {
			return filter.ErrProviderUnavailable
		}
	} else if len(families) == 1 && operation != firewall.BaseOperationUnbind {
		initialize := operation == firewall.BaseOperationInit
		if provider == constant.FirewallProviderNftables {
			err = nftables_helper.OperateFamily(filter.FamilyIPv4, initialize, ports)
		} else {
			err = iptables_helper.OperateFamily(constant.FirewallFamilyIPv4, initialize, ports)
		}
		if err != nil {
			return err
		}
	} else if provider == constant.FirewallProviderNftables {
		if err := nftables_helper.Operate(operation, ports); err != nil {
			return err
		}
	} else if err := iptables_helper.Operate(operation, ports); err != nil {
		return err
	}
	status := constant.StatusEnable
	if operation == firewall.BaseOperationUnbind {
		status = constant.StatusDisable
	}
	return settingRepo.Update("IptablesStatus", status)
}

func LoadRequiredFirewallPortWhiteList() ([]firewall.PortWhitelist, error) {
	ports, err := loadFirewallPortWhiteList()
	if err != nil {
		return nil, err
	}
	return firewall.RequiredPortWhitelist(ports)
}

func lockFirewallLifecycleIdle() error {
	if !firewallLifecycleTaskMu.TryLock() {
		return buserr.New("TaskIsExecuting")
	}
	if firewallLifecycleTaskID != "" {
		firewallLifecycleTaskMu.Unlock()
		return buserr.New("TaskIsExecuting")
	}
	return nil
}

func cleanupInactiveSystemBackend(backend string) error {
	switch backend {
	case constant.FirewallProviderIptables:
		return iptables_helper.Cleanup()
	case constant.FirewallProviderNftables:
		return nftables_helper.Cleanup()
	default:
		return fmt.Errorf("cleanup is only available for 1Panel-owned iptables and nftables resources")
	}
}

func cleanupSystemBackend(backend string) error {
	switch backend {
	case constant.FirewallProviderIptables:
		if err := iptables_helper.Cleanup(); err != nil {
			return err
		}
	case constant.FirewallProviderNftables:
		if err := nftables_helper.Cleanup(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("cleanup is only available for 1Panel-owned iptables and nftables resources")
	}
	return settingRepo.Update("IptablesStatus", constant.StatusDisable)
}

func resetServiceFirewallBackend(provider string, withDockerRestart bool) error {
	client, err := lifecycle.NewClient(provider)
	if err != nil {
		return err
	}
	return resetServiceFirewallClient(client, withDockerRestart, func(
		client lifecycle.Client,
		restartDocker bool,
		prepareStop func() error,
	) error {
		return stopFirewallLifecycle(client, restartDocker, prepareStop, nil)
	})
}

func resetServiceFirewallClient(client lifecycle.Client, withDockerRestart bool, stop func(lifecycle.Client, bool, func() error) error) error {
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

func firewallInventoryPositionRanges(provider filter.Provider, chain string, items []filter.InventoryItem) (ipv4, ipv6 filter.PositionRange) {
	if provider == filter.ProviderFirewalld {
		return filter.PositionRange{Min: -32768, Max: 32767}, filter.PositionRange{Min: -32768, Max: 32767}
	}
	for _, item := range items {
		if item.Observed == nil || item.Observed.Locator.Position == nil {
			continue
		}
		scope := item.Observed.Rule.Scope
		if scope.Provider != provider || scope.Direction != filter.DirectionInput {
			continue
		}
		position := *item.Observed.Locator.Position
		if position < 1 {
			continue
		}
		if (provider == filter.ProviderIptables || provider == filter.ProviderNftables) &&
			(scope.Table != "filter" || scope.Chain != chain) {
			continue
		}
		bounds := &ipv4
		if scope.Family == filter.FamilyIPv6 {
			bounds = &ipv6
		} else if scope.Family != filter.FamilyIPv4 {
			continue
		}
		if bounds.Min == 0 || position < bounds.Min {
			bounds.Min = position
		}
		bounds.Max = max(bounds.Max, position)
		if provider != filter.ProviderUFW {
			bounds.Min = 1
		}
	}
	return
}

func findPortWhitelistRule(rules []firewall.PortWhitelist, target firewall.PortWhitelist) (int, error) {
	index := slices.IndexFunc(rules, func(rule firewall.PortWhitelist) bool {
		return samePortWhitelistRule(rule, target)
	})
	if index < 0 {
		return -1, fmt.Errorf("firewall port whitelist rule has changed or no longer exists; refresh and retry")
	}
	return index, nil
}

func samePortWhitelistRule(left, right firewall.PortWhitelist) bool {
	if reflect.DeepEqual(left, right) {
		return true
	}
	normalizedLeft, err := firewall.ValidatePortWhitelist([]firewall.PortWhitelist{left})
	if err != nil {
		return false
	}
	normalizedRight, err := firewall.ValidatePortWhitelist([]firewall.PortWhitelist{right})
	if err != nil {
		return false
	}
	slices.Sort(normalizedLeft[0].Sources)
	slices.Sort(normalizedRight[0].Sources)
	return reflect.DeepEqual(normalizedLeft[0], normalizedRight[0])
}

func loadSSHWhitelistPortFrom(path string) (string, error) {
	directives, _, err := parseSSHConfigTree(path)
	if errors.Is(err, os.ErrNotExist) {
		return defaultSSHPort, nil
	}
	if err != nil {
		return "", err
	}
	return loadSSHPortValues(directives)[0], nil
}

func newFirewallService() *FirewallService {
	return &FirewallService{
		adapters:               nil,
		selectedProvider:       firewallRuleSelectedProvider,
		requiredPorts:          LoadRequiredFirewallPortWhiteList,
		cleanupBackend:         cleanupSystemBackend,
		cleanupInactiveBackend: cleanupInactiveSystemBackend,
		resetBackend:           resetServiceFirewallBackend,
		dockerActive:           firewallDockerActive,
		restoreForwarding: func(ctx context.Context) error {
			return newForwardingService().Restore(ctx)
		},
		restoreDockerGuard: RestoreDockerPortGuard,
		baseClient:         NewSelectedSystemFirewallClient,
	}
}

func firewallRuleSelectedProvider(context.Context) (filter.Provider, error) {
	client, err := NewSelectedSystemFirewallClient()
	if err != nil {
		return "", fmt.Errorf("%w: %v", filter.ErrProviderUnavailable, err)
	}
	return filter.Provider(client.Name()), nil
}

func discoverDockerEndpoints(ctx context.Context, cli *client.Client, all bool) ([]dto.DockerPortGuardEndpoint, error) {
	containers, err := cli.ContainerList(ctx, containertypes.ListOptions{All: all})
	if err != nil {
		return nil, err
	}
	endpoints := make([]dto.DockerPortGuardEndpoint, 0)
	for _, item := range containers {
		name := ""
		if len(item.Names) > 0 {
			name = strings.TrimPrefix(item.Names[0], "/")
		}
		compose := item.Labels[dockerGuardComposeProjectLabel]
		application := ""
		if created, ok := item.Labels[dockerGuardComposeCreatedBy]; ok && created == "Apps" {
			application = compose
		}
		for _, port := range item.Ports {
			if port.PublicPort == 0 || (port.Type != "tcp" && port.Type != "udp") {
				continue
			}
			family := dockerfirewall.FamilyIPv4
			hostIP := port.IP
			if addr, err := netip.ParseAddr(hostIP); err == nil && addr.Is6() {
				family = dockerfirewall.FamilyIPv6
			} else if hostIP == "" {
				hostIP = "0.0.0.0"
			}
			endpoints = append(endpoints, dto.DockerPortGuardEndpoint{Family: family, HostIP: hostIP, HostPort: port.PublicPort, Protocol: port.Type, ContainerID: item.ID, ContainerName: name, ContainerState: item.State, ContainerPort: port.PrivatePort, Compose: compose, Application: application, Sources: []string{}})
		}
	}
	return endpoints, nil
}

func annotateDockerEndpointManagement(ctx context.Context, endpoints []dto.DockerPortGuardEndpoint, backend string) {
	rules := map[string]dockerfirewall.DNATRules{
		constant.FirewallFamilyIPv4: dockerfirewall.ReadDNATRules(ctx, backend, constant.FirewallFamilyIPv4),
		constant.FirewallFamilyIPv6: dockerfirewall.ReadDNATRules(ctx, backend, constant.FirewallFamilyIPv6),
	}
	proxies := dockerfirewall.ReadProxyEndpoints(ctx)
	inspections := make(map[string]dockerfirewall.EndpointInspection, len(rules))
	for family, familyRules := range rules {
		inspections[family] = dockerfirewall.InspectEndpoints(backend, family, familyRules, proxies)
	}
	for i := range endpoints {
		endpoints[i].TrafficPath, endpoints[i].ManagementTarget, endpoints[i].ManagementReason =
			dockerEndpointManagement(inspections[endpoints[i].Family], endpoints[i])
	}
}

func dockerEndpointManagement(inspection dockerfirewall.EndpointInspection, endpoint dto.DockerPortGuardEndpoint) (string, string, string) {
	if !inspection.DNATInspected {
		return dockerTrafficPathUnknown, dockerManagementNeedsDiagnosis, dockerReasonNATInspectFailed
	}
	dnatMatched := inspection.DNATMatches(endpoint.HostIP, endpoint.HostPort, endpoint.Protocol)
	if dnatMatched && inspection.IngressReachable {
		return dockerTrafficPathForward, dockerManagementContainerGuard, ""
	}
	if !inspection.ProxyInspected {
		return dockerTrafficPathUnknown, dockerManagementNeedsDiagnosis, dockerReasonProxyInspectFailed
	}
	if inspection.ProxyMatches(endpoint.HostIP, endpoint.HostPort, endpoint.Protocol) {
		return dockerTrafficPathInput, dockerManagementHostFirewall, ""
	}
	if dnatMatched {
		return dockerTrafficPathUnknown, dockerManagementNeedsDiagnosis, dockerReasonNATChainUnreachable
	}
	return dockerTrafficPathUnknown, dockerManagementNeedsDiagnosis, dockerReasonNoMatchingPath
}

func groupDockerGuardContainers(endpoints []dto.DockerPortGuardEndpoint) []dto.DockerPortGuardContainer {
	containers := make(map[string]*dto.DockerPortGuardContainer)
	order := make([]string, 0)
	for _, endpoint := range endpoints {
		key := endpoint.ContainerID
		if key == "" {
			key = "__orphan__"
		}
		container, ok := containers[key]
		if !ok {
			container = &dto.DockerPortGuardContainer{
				Key: key, Name: endpoint.ContainerName, Compose: endpoint.Compose,
				Application: endpoint.Application, Endpoints: []dto.DockerPortGuardEndpoint{},
			}
			containers[key] = container
			order = append(order, key)
		}
		container.Endpoints = append(container.Endpoints, endpoint)
	}
	sort.Slice(order, func(i, j int) bool {
		return containers[order[i]].Name < containers[order[j]].Name
	})

	result := make([]dto.DockerPortGuardContainer, 0, len(order))
	for _, key := range order {
		container := containers[key]
		items := make([]docker.PortRangeItem, 0, len(container.Endpoints))
		for i, endpoint := range container.Endpoints {
			sources := append([]string(nil), endpoint.Sources...)
			sort.Strings(sources)
			policyKey := fmt.Sprintf("%t|%s|%s|%t|%s|%s", endpoint.PolicyUUID != "", endpoint.Mode, strings.Join(sources, ","), endpoint.Effective, endpoint.ManagementTarget, endpoint.ManagementReason)
			items = append(items, docker.PortRangeItem{
				Key:        endpoint.Family + "|" + endpoint.HostIP + "|" + endpoint.Protocol + "|" + policyKey,
				PublicPort: endpoint.HostPort, PrivatePort: endpoint.ContainerPort,
				HasPrivatePort: endpoint.ContainerPort != 0, Position: i,
			})
		}
		container.PortGroups = make([]dto.DockerPortGuardPortGroup, 0, len(items))
		for _, portRange := range docker.MergePortRanges(items) {
			start := container.Endpoints[portRange.Start.Position]
			address := start.HostIP
			if strings.Contains(address, ":") {
				address = "[" + address + "]"
			}
			ports := fmt.Sprintf("%d", portRange.Start.PublicPort)
			if portRange.Start.PublicPort != portRange.End.PublicPort {
				ports = fmt.Sprintf("%d-%d", portRange.Start.PublicPort, portRange.End.PublicPort)
			}
			container.PortGroups = append(container.PortGroups, dto.DockerPortGuardPortGroup{
				Key:   fmt.Sprintf("%s|%d-%d", portRange.Start.Key, portRange.Start.PublicPort, portRange.End.PublicPort),
				Label: fmt.Sprintf("%s:%s/%s", address, ports, start.Protocol), Endpoint: start,
				Endpoints: func() []dto.DockerPortGuardEndpoint {
					members := make([]dto.DockerPortGuardEndpoint, 0, len(portRange.Items))
					for _, item := range portRange.Items {
						members = append(members, container.Endpoints[item.Position])
					}
					return members
				}(),
			})
		}
		result = append(result, *container)
	}
	return result
}

func normalizeDockerFirewallUUIDs(values []string) ([]string, error) {
	uuids := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, policyUUID := range values {
		policyUUID = strings.TrimSpace(policyUUID)
		if policyUUID == "" {
			return nil, buserr.WithDetail("ErrInvalidParams", "policy UUID cannot be empty", nil)
		}
		if _, exists := seen[policyUUID]; exists {
			continue
		}
		seen[policyUUID] = struct{}{}
		uuids = append(uuids, policyUUID)
	}
	if len(uuids) == 0 {
		return nil, buserr.WithDetail("ErrInvalidParams", "policy UUIDs cannot be empty", nil)
	}
	return uuids, nil
}

func queueFirewallRuleTask(subsystem, operation, taskID string, labels []string, apply func(*task.Task) error) (dto.FilterChainOperationResponse, error) {
	taskItem, err := task.NewTask(firewallTaskName(operation, subsystem, ""), operation, task.TaskScopeFirewall, taskID, 0)
	if err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	taskItem.AddSubTaskWithOps(taskItem.Name, func(t *task.Task) error {
		if labels != nil {
			t.Logf("rules=%d", len(labels))
		}
		err := t.TaskCtx.Err()
		if err == nil {
			err = apply(t)
		}
		err = errors.Join(err, t.TaskCtx.Err())
		if labels == nil {
			return err
		}
		succeeded, failed, skipped := 0, 0, 0
		for _, label := range labels {
			if label == "" {
				skipped++
				continue
			}
			if err != nil {
				failed++
				t.LogFailedWithErr(label, err)
			} else {
				succeeded++
				t.LogSuccess(label)
			}
		}
		t.Log(i18n.GetMsgWithMap("FirewallCreateRulesResult", map[string]interface{}{
			"succeeded": succeeded, "failed": failed, "skipped": skipped,
		}))
		return err
	}, nil, 0, 0)
	if err := taskRepo.Save(context.Background(), taskItem.Task); err != nil {
		taskItem.LogFailedWithErr(taskItem.Name, err)
		closeUnstartedFirewallTask(taskItem)
		return dto.FilterChainOperationResponse{}, fmt.Errorf("save firewall rule task: %w", err)
	}
	go func() { _ = taskItem.Execute() }()
	return dto.FilterChainOperationResponse{TaskID: taskItem.TaskID, Queued: true}, nil
}

func normalizeDockerFirewallPolicy(policy dockerfirewall.Policy) (dockerfirewall.Policy, error) {
	policy.Family = strings.ToLower(strings.TrimSpace(policy.Family))
	policy.HostIP = strings.TrimSpace(policy.HostIP)
	policy.Protocol = strings.ToLower(strings.TrimSpace(policy.Protocol))
	policy.Mode = strings.ToLower(strings.TrimSpace(policy.Mode))
	if policy.HostPort == 0 ||
		(policy.Protocol != "tcp" && policy.Protocol != "udp") ||
		(policy.Family != dockerfirewall.FamilyIPv4 && policy.Family != dockerfirewall.FamilyIPv6) ||
		(policy.Mode != dockerfirewall.ModeAll && policy.Mode != dockerfirewall.ModeSources && policy.Mode != dockerfirewall.ModeAllow && policy.Mode != dockerfirewall.ModeAcceptSources && policy.Mode != dockerfirewall.ModeAcceptAll) {
		return dockerfirewall.Policy{}, buserr.WithDetail("ErrInvalidParams", "invalid policy fields", nil)
	}
	address, err := netip.ParseAddr(policy.HostIP)
	if err != nil || (policy.Family == dockerfirewall.FamilyIPv4) != address.Is4() {
		return dockerfirewall.Policy{}, buserr.WithDetail("ErrInvalidParams", "host IP does not match address family", nil)
	}
	policy.HostIP = address.String()
	normalizedSources := make([]string, 0, len(policy.Sources))
	seen := make(map[string]struct{}, len(policy.Sources))
	for _, source := range policy.Sources {
		source = strings.TrimSpace(source)
		if source == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(source)
		if err != nil {
			if sourceAddress, addressErr := netip.ParseAddr(source); addressErr == nil {
				bits := 128
				if sourceAddress.Is4() {
					bits = 32
				}
				prefix = netip.PrefixFrom(sourceAddress, bits)
			} else {
				return dockerfirewall.Policy{}, buserr.WithDetail("ErrInvalidParams", fmt.Sprintf("invalid source address %q", source), nil)
			}
		}
		if (policy.Family == dockerfirewall.FamilyIPv4) != prefix.Addr().Is4() {
			return dockerfirewall.Policy{}, buserr.WithDetail("ErrInvalidParams", fmt.Sprintf("source %q does not match address family", source), nil)
		}
		canonical := prefix.Masked().String()
		if _, exists := seen[canonical]; !exists {
			seen[canonical] = struct{}{}
			normalizedSources = append(normalizedSources, canonical)
		}
	}
	if policy.Mode != dockerfirewall.ModeAll && policy.Mode != dockerfirewall.ModeAcceptAll && len(normalizedSources) == 0 {
		return dockerfirewall.Policy{}, buserr.WithDetail("ErrInvalidParams", "source-based modes require at least one source", nil)
	}
	if policy.Mode == dockerfirewall.ModeAll || policy.Mode == dockerfirewall.ModeAcceptAll {
		normalizedSources = []string{}
	}
	sort.Strings(normalizedSources)
	policy.Sources = normalizedSources
	return policy, nil
}

func forwardingOperationsOnlyRemove(operations []dto.ForwardRuleOperation) bool {
	if len(operations) == 0 {
		return false
	}
	for _, operation := range operations {
		if operation.Operation != string(forwarding.OperationRemove) {
			return false
		}
	}
	return true
}

func (e *firewallDockerRestartError) Error() string {
	return fmt.Sprintf("failed to restart Docker: %v", e.Err)
}

func (e *firewallDockerRestartError) Unwrap() error {
	return e.Err
}

func (e *firewallCompletedOperationError) Error() string {
	return fmt.Sprintf("firewall %s completed with recovery errors: %v", e.Operation, e.Err)
}

func (e *firewallCompletedOperationError) Unwrap() error {
	return e.Err
}

func InitializeFirewallWhitelistPorts(entries []firewall.PortWhitelist) ([]firewall.PortWhitelist, error) {
	entries = slices.Clone(entries)
	var sshPort string
	for i := range entries {
		entry := &entries[i]
		if entry.Type == "" || entry.Port != "" {
			continue
		}
		switch entry.Type {
		case firewall.PortWhitelistTypePanel:
			entry.Port = LoadPanelPort()
		case firewall.PortWhitelistTypeSSH:
			if sshPort == "" {
				var err error
				sshPort, err = loadSSHWhitelistPortFrom(sshPath)
				if err != nil {
					return nil, err
				}
			}
			entry.Port = sshPort
		}
	}
	return firewall.ValidatePortWhitelist(entries)
}
func (s *FirewallService) readFirewallInventory(ctx context.Context, provider filter.Provider, scopes []filter.Scope) (dto.FirewallRuleInventoryResponse, error) {
	response := dto.FirewallRuleInventoryResponse{Items: make([]filter.InventoryItem, 0)}
	runtime, err := s.firewallAdapter(provider)
	if err != nil {
		return response, err
	}
	for _, scope := range scopes {
		if scope.Provider != provider {
			return response, filter.ErrInvalidScope
		}
	}
	for _, group := range firewallScopeReadGroups(scopes) {
		snapshots, err := listFirewallRuleScopes(runtime, ctx, group)
		if errors.Is(err, filter.ErrFamilyUnavailable) {
			response.Notices = append(response.Notices, filter.ScopeNotice{Code: filter.ScopeNoticeFamilyUnavailable, Values: []string{string(group[0].Family), err.Error()}})
			continue
		}
		if err != nil {
			return response, err
		}
		for _, snapshot := range snapshots {
			response.Notices = append(response.Notices, snapshot.Notices...)
			for _, observed := range snapshot.Rules {
				observed.InstanceKey, err = filter.InstanceKey(observed)
				if err != nil {
					return response, err
				}
				response.Items = append(response.Items, filter.InventoryItem{Rule: observed.Rule, Observed: &observed})
			}
		}
	}
	if provider == filter.ProviderUFW {
		sort.SliceStable(response.Items, func(i, j int) bool {
			left, right := response.Items[i].Observed.Locator.Position, response.Items[j].Observed.Locator.Position
			return left != nil && right != nil && *left < *right
		})
	}
	return response, nil
}

func saveFirewallDescription(ctx context.Context, observed filter.ObservedRule, description string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	id, err := filter.DescriptionID(observed)
	if err != nil {
		return err
	}
	description = strings.TrimSpace(description)
	if description == "" {
		return settingRepo.DelDescription(id)
	}
	existing, err := settingRepo.GetDescription(settingRepo.WithByDescriptionID(id), repo.WithByType("firewall"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return settingRepo.CreateDescription(&model.CommonDescription{ID: id, Type: "firewall", Description: description})
	}
	if err != nil {
		return err
	}
	return settingRepo.UpdateDescription(existing.ID, map[string]interface{}{"description": description})
}

func (s *FirewallService) updateObservedRule(ctx context.Context, request dto.FirewallRuleUpdate) error {
	metadata := request.Description != nil || request.OrderIndex != nil || request.Priority != nil
	if (request.Rule != nil) == metadata || request.OrderIndex != nil && request.Priority != nil {
		return filter.ErrInvalidRule
	}
	runtime, err := s.resolveRuntime(ctx, request.Scope.Provider)
	if err != nil {
		return err
	}
	snapshot, err := readMutableFirewallRules(runtime, ctx, request.Scope)
	if err != nil {
		return err
	}
	observed, err := filter.FindCandidate(snapshot.Rules, request.InstanceKey)
	if err != nil {
		return filter.ErrRuleStale
	}
	if err := filter.GuardMutation(observed); err != nil {
		return err
	}
	if request.Rule != nil {
		ports, err := loadFirewallPortWhiteList()
		if err != nil {
			return err
		}
		if filter.NewPortWhitelistIndex(ports).Matches(observed.Rule) {
			return filter.ErrProtectedRule
		}
	}
	if request.Rule == nil && request.OrderIndex == nil && request.Priority == nil {
		return saveFirewallDescription(ctx, observed, *request.Description)
	}
	if observed.ParseStatus != filter.ParseStatusSupported {
		return filter.ErrUnsupportedScope
	}
	before := observed.Rule
	before.UUID = uuid.NewString()
	after := before
	operation := filter.ChangeReorder
	if runtime.Provider() == filter.ProviderFirewalld {
		operation = filter.ChangeUpdate
	}
	if request.Rule != nil {
		after, err = prepareFirewallBackendRule(ctx, runtime, *request.Rule)
		if err != nil {
			return err
		}
		operation = filter.ChangeUpdate
	} else {
		after.OrderIndex, after.Priority = request.OrderIndex, before.Priority
		if request.Priority != nil {
			after.Priority = request.Priority
		}
	}
	if after.Scope.Normalize().Key() != before.Scope.Key() {
		return filter.ErrUnsupportedScope
	}
	after.UUID = before.UUID
	after, err = filter.NormalizeRule(after)
	if err != nil {
		return err
	}
	beforeKey, err := firewallRuleDuplicateKey(observed, false)
	if err != nil {
		return err
	}
	afterKey, err := firewallRuleDuplicateKey(filter.ObservedRule{Rule: after, ParseStatus: filter.ParseStatusSupported}, false)
	if err != nil {
		return err
	}
	if beforeKey == afterKey {
		if after.OrderIndex == nil || observed.Locator.Position != nil && *after.OrderIndex == int64(*observed.Locator.Position) {
			if request.Rule != nil {
				return saveFirewallDescription(ctx, observed, request.Rule.Description)
			}
			if request.Description != nil {
				return saveFirewallDescription(ctx, observed, *request.Description)
			}
			return nil
		}
	} else {
		var matchKey string
		if runtime.Provider() == filter.ProviderUFW {
			matchKey, err = filter.RuleMatchKey(after)
			if err != nil {
				return err
			}
		}
		for _, candidate := range snapshot.Rules {
			instanceKey, err := filter.InstanceKey(candidate)
			if err != nil {
				return err
			}
			if instanceKey == request.InstanceKey {
				continue
			}
			key, err := firewallRuleDuplicateKey(candidate, false)
			if err != nil {
				return err
			}
			if key == afterKey {
				return buserr.New("ErrRecordExist")
			}
			if runtime.Provider() == filter.ProviderUFW && candidate.ParseStatus == filter.ParseStatusSupported {
				candidateMatch, err := filter.RuleMatchKey(candidate.Rule)
				if err != nil {
					return err
				}
				if candidateMatch == matchKey && candidate.Rule.Action != after.Action {
					return buserr.New("ErrFirewallRuleConflict")
				}
			}
		}
	}
	description := before.Description
	id, err := filter.DescriptionID(observed)
	if err != nil {
		return err
	}
	stored, err := settingRepo.GetDescription(settingRepo.WithByDescriptionID(id), repo.WithByType("firewall"))
	if err == nil {
		description = stored.Description
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if request.Rule != nil {
		description = request.Rule.Description
	}
	if request.Description != nil {
		description = *request.Description
	}
	plan, err := applyFirewallChanges(runtime, ctx, snapshot, []filter.RuleChange{{
		Operation: operation, Before: &before, After: &after, Locator: &observed.Locator, PreviousMarker: observed.Marker,
	}})
	if err != nil {
		return err
	}
	return saveFirewallDescription(ctx, plan.Rules[0].Expected, description)
}

func firewallBackupDirectory(subsystem string) (string, error) {
	var directory string
	switch subsystem {
	case "system":
		directory = "host"
	case "docker":
		directory = "docker"
	case "forwarding":
		directory = "forward"
	default:
		return "", filter.ErrInvalidRule
	}
	return filepath.Join(global.Dir.BaseDir, "1panel", "backup", "firewall", directory), nil
}

func readFirewallBackup(name string) ([]dto.FirewallRuleExportItem, error) {
	if name == "" || name != filepath.Base(name) || filepath.Ext(name) != ".json" {
		return nil, filter.ErrInvalidRule
	}
	directory, err := firewallBackupDirectory("system")
	if err != nil {
		return nil, err
	}
	file, err := os.OpenInRoot(directory, name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, filter.ErrInvalidRule
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var items []dto.FirewallRuleExportItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	if items == nil {
		return nil, filter.ErrInvalidRule
	}
	for _, item := range items {
		if err := item.Scope.ValidateMVP(); err != nil {
			return nil, err
		}
		switch item.ParseStatus {
		case "", filter.ParseStatusSupported:
			if _, err := filter.NormalizeRule(item.FirewallRule); err != nil {
				return nil, err
			}
		case filter.ParseStatusPartial, filter.ParseStatusOpaque:
			if item.Raw == "" {
				return nil, filter.ErrInvalidRule
			}
		default:
			return nil, filter.ErrInvalidRule
		}
	}
	return items, nil
}

func writeFirewallJSON(directory, name string, value any) (string, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return "", err
	}
	file, err := os.CreateTemp(directory, ".firewall-*.tmp")
	if err != nil {
		return "", err
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	_, writeErr := file.Write(data)
	syncErr := file.Sync()
	closeErr := file.Close()
	if err := errors.Join(writeErr, syncErr, closeErr); err != nil {
		return "", err
	}
	target := filepath.Join(directory, name)
	if err := os.Rename(temporary, target); err != nil {
		return "", err
	}
	return target, nil
}

func readFirewallSubsystemBackup(name, subsystem string) (dto.FirewallSubsystemBackup, error) {
	var backup dto.FirewallSubsystemBackup
	if name == "" || name != filepath.Base(name) || (filepath.Ext(name) != ".json" && name != subsystem+"-iptables.rules" && name != subsystem+"-nftables.rules") {
		return backup, filter.ErrInvalidRule
	}
	directory, err := firewallBackupDirectory(subsystem)
	if err != nil {
		return backup, err
	}
	if filepath.Ext(name) == ".rules" {
		directory = global.Dir.FirewallDir
	}
	file, err := os.OpenInRoot(directory, name)
	if err != nil {
		return backup, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return backup, err
	}
	if !info.Mode().IsRegular() {
		return backup, filter.ErrInvalidRule
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return backup, err
	}
	if err := json.Unmarshal(data, &backup); err != nil {
		return backup, err
	}
	if backup.Subsystem != subsystem || (backup.Provider != filter.ProviderIptables && backup.Provider != filter.ProviderNftables) {
		return backup, filter.ErrInvalidRule
	}
	for _, family := range backup.Families {
		if family != forwarding.FamilyIPv4 && family != forwarding.FamilyIPv6 {
			return backup, filter.ErrInvalidRule
		}
	}
	switch subsystem {
	case "forwarding":
		if backup.Forwarding == nil || backup.Docker != nil {
			return backup, filter.ErrInvalidRule
		}
		for i, rule := range backup.Forwarding {
			normalized, err := forwarding.NormalizeRule(rule)
			if err != nil {
				return backup, err
			}
			backup.Forwarding[i] = normalized
		}
	case "docker":
		if backup.Docker == nil || len(backup.Forwarding) > 0 {
			return backup, filter.ErrInvalidRule
		}
		seen := make(map[string]bool)
		for i, policy := range backup.Docker.Policies {
			if _, err := uuid.Parse(policy.UUID); err != nil {
				return backup, filter.ErrInvalidRule
			}
			normalized := policy
			if len(policy.NativeRules) == 0 {
				var err error
				normalized, err = normalizeDockerFirewallPolicy(policy)
				if err != nil {
					return backup, err
				}
			}
			if seen[normalized.UUID] {
				return backup, filter.ErrInvalidRule
			}
			seen[normalized.UUID] = true
			backup.Docker.Policies[i] = normalized
		}
		for _, policy := range backup.Docker.Policies {
			for _, rule := range policy.NativeRules {
				if rule.Family != dockerfirewall.FamilyIPv4 && rule.Family != dockerfirewall.FamilyIPv6 {
					return backup, filter.ErrInvalidRule
				}
				if len(rule.Tokens) == 0 {
					return backup, filter.ErrInvalidRule
				}
				for _, token := range rule.Tokens {
					if strings.ContainsAny(token, "\r\n\x00") {
						return backup, filter.ErrInvalidRule
					}
				}
				if backup.Provider == filter.ProviderIptables && (len(rule.Tokens) < 3 || rule.Tokens[0] != "-A" || rule.Tokens[1] != dockerfirewall.Chain) {
					return backup, filter.ErrInvalidRule
				}
			}
		}
	default:
		return backup, filter.ErrInvalidRule
	}
	return backup, nil
}

func forwardingBackup(manager forwarding.Adapter, rules []forwarding.Rule) (dto.FirewallSubsystemBackup, error) {
	if rules == nil {
		rules = []forwarding.Rule{}
	}
	backup := dto.FirewallSubsystemBackup{Subsystem: "forwarding", Provider: filter.Provider(manager.Name()), Forwarding: rules}
	for _, family := range []string{forwarding.FamilyIPv4, forwarding.FamilyIPv6} {
		initialized, _, err := manager.FamilyStatus(family)
		if family == forwarding.FamilyIPv6 && errors.Is(err, filter.ErrFamilyUnavailable) {
			continue
		}
		if err != nil {
			return backup, err
		}
		if initialized {
			backup.Families = append(backup.Families, family)
		}
	}
	return backup, nil
}

func persistForwardingRules(manager forwarding.Adapter, rules []forwarding.Rule) error {
	backup, err := forwardingBackup(manager, rules)
	if err != nil {
		return err
	}
	_, err = writeFirewallJSON(global.Dir.FirewallDir, "forwarding-"+manager.Name()+".rules", backup)
	return err
}

func persistDockerRules(backend string, inventory dockerfirewall.PolicyInventory) error {
	_, err := writeFirewallJSON(global.Dir.FirewallDir, "docker-"+backend+".rules", dto.FirewallSubsystemBackup{Subsystem: "docker", Provider: filter.Provider(backend), Docker: &inventory})
	return err
}

func dockerPolicyEndpointKey(policy dockerfirewall.Policy) string {
	return strings.Join([]string{policy.Family, policy.HostIP, strconv.Itoa(int(policy.HostPort)), policy.Protocol}, "\x00")
}

func dockerGuardInventoryEndpoints(inventory dockerfirewall.PolicyInventory) []dto.DockerPortGuardEndpoint {
	result := make([]dto.DockerPortGuardEndpoint, 0, len(inventory.Policies))
	for _, policy := range inventory.Policies {
		if _, err := normalizeDockerFirewallPolicy(policy); err != nil {
			policy.Mode = ""
		}
		result = append(result, dto.DockerPortGuardEndpoint{Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol, PolicyUUID: policy.UUID, Mode: policy.Mode, Sources: policy.Sources, TrafficPath: dockerTrafficPathUnknown, ManagementTarget: dockerManagementNeedsDiagnosis, ManagementReason: dockerReasonNoMatchingPath})
	}
	return result
}

func applyDockerPolicies(ctx context.Context, runtime dockerfirewall.Runtime, backend string, inventory dockerfirewall.PolicyInventory, desired []dockerfirewall.Policy) error {
	families := make(map[string]bool)
	for _, policy := range desired {
		families[policy.Family] = true
	}
	for family := range families {
		initialized, err := runtime.Initialized(family)
		if err != nil {
			return err
		}
		if !initialized {
			return filter.ErrProviderUnavailable
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := runtime.ReplacePolicies(desired, inventory); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	inventory.Policies = desired
	return persistDockerRules(backend, inventory)
}

func (s *DockerPortGuardService) initialize(ctx context.Context, request dto.DockerPortGuardOperation, t *task.Task) error {
	runtime, backend, err := s.runtimeForDocker(ctx)
	if err != nil {
		return err
	}
	inventory, err := runtime.ListPolicies()
	if err != nil {
		return err
	}
	if request.BackupFile != "" {
		backup, err := readFirewallSubsystemBackup(request.BackupFile, "docker")
		if err != nil {
			return err
		}
		inventory, err = mergeDockerBackup(inventory, backup, backend, t)
		if err != nil {
			return err
		}
	}
	families, err := loadFirewallFamilies()
	if err != nil {
		return err
	}
	if !slices.Contains(families, constant.FirewallFamilyIPv6) {
		if t != nil {
			t.Logf("IPv6 firewall support is disabled; skipping IPv6 Docker chain initialization")
		} else {
			global.LOG.Info("IPv6 firewall support is disabled; skipping IPv6 Docker chain initialization")
		}
	}
	if err := runtime.Initialize(inventory.Policies, inventory, families...); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := persistDockerRules(backend, inventory); err != nil {
		return err
	}
	if err := settingRepo.UpdateOrCreate(constant.FirewallDockerBackendKey, backend); err != nil {
		return err
	}
	if err := settingRepo.UpdateOrCreate(constant.FirewallDockerPortGuardStatusKey, constant.StatusEnable); err != nil {
		return err
	}
	return nil
}

func mergeDockerBackup(current dockerfirewall.PolicyInventory, backup dto.FirewallSubsystemBackup, backend string, t *task.Task) (dockerfirewall.PolicyInventory, error) {
	families, err := loadFirewallFamilies()
	if err != nil {
		return current, err
	}
	skippedIPv6 := 0
	defer func() { logFirewallIPv6Skipped(t, "Docker", skippedIPv6) }()
	empty := len(current.Policies) == 0
	byEndpoint := make(map[string]bool, len(current.Policies))
	byUUID := make(map[string]bool, len(current.Policies))
	for _, policy := range current.Policies {
		byEndpoint[dockerPolicyEndpointKey(policy)] = true
		byUUID[policy.UUID] = true
	}
	if current.RuleOrders == nil {
		current.RuleOrders = make(map[string][]int64)
	}
	var maxOrder int64
	for _, existing := range current.Policies {
		for _, rule := range existing.NativeRules {
			maxOrder = max(maxOrder, rule.Order)
		}
		for _, order := range current.RuleOrders[existing.Family+"\x00"+existing.UUID] {
			maxOrder = max(maxOrder, order)
		}
	}
	for _, policy := range backup.Docker.Policies {
		if policy.Family == dockerfirewall.FamilyIPv6 && !slices.Contains(families, dockerfirewall.FamilyIPv6) {
			skippedIPv6++
			continue
		}
		if byEndpoint[dockerPolicyEndpointKey(policy)] {
			if t != nil {
				t.Logf("%s %s %s:%d %s: %v", policy.Family, policy.Protocol, policy.HostIP, policy.HostPort, i18n.GetMsgByKey("FirewallCreateRuleSkipped"), buserr.New("ErrRecordExist"))
			}
			continue
		}
		if byUUID[policy.UUID] {
			return current, filter.ErrRuleStale
		}
		if string(backup.Provider) != backend {
			policy, err = normalizeDockerFirewallPolicy(policy)
			if err != nil {
				return current, err
			}
			policy, err = dockerfirewall.ConvertPolicyBackend(policy, string(backup.Provider), backend)
			if err != nil {
				return current, err
			}
		}
		if !empty {
			policy.NativeRules = append([]dockerfirewall.NativeRule(nil), policy.NativeRules...)
			for i := range policy.NativeRules {
				maxOrder++
				policy.NativeRules[i].Order = maxOrder
			}
		}
		current.Policies = append(current.Policies, policy)
		byEndpoint[dockerPolicyEndpointKey(policy)], byUUID[policy.UUID] = true, true
		key := policy.Family + "\x00" + policy.UUID
		if empty {
			current.RuleOrders[key] = backup.Docker.RuleOrders[key]
		}
	}
	return current, nil
}

func RestoreDockerPortGuard(ctx context.Context) error {
	return newDockerPortGuardService().Restore(ctx)
}

func RestoreDockerPortGuardBestEffort(ctx context.Context) {
	if err := RestoreDockerPortGuard(ctx); err != nil {
		global.LOG.Warnf("restore Docker port guard failed: %v", err)
	}
}

func loadFirewallFamilies() ([]string, error) {
	value, err := settingRepo.GetValueByKey(constant.FirewallIPv6SupportKey)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if value == constant.StatusDisable {
		return []string{constant.FirewallFamilyIPv4}, nil
	}
	return []string{constant.FirewallFamilyIPv4, constant.FirewallFamilyIPv6}, nil
}

func ipv4PortWhitelist(rules []firewall.PortWhitelist) []firewall.PortWhitelist {
	result := make([]firewall.PortWhitelist, 0, len(rules))
	for _, rule := range rules {
		sources := make([]string, 0, len(rule.Sources))
		for _, source := range rule.Sources {
			if !strings.Contains(source, ":") {
				sources = append(sources, source)
			}
		}
		if len(sources) == 0 {
			continue
		}
		rule.Sources = sources
		result = append(result, rule)
	}
	return result
}

func validateFirewallWhitelistFamilies(rules []firewall.PortWhitelist) error {
	families, err := loadFirewallFamilies()
	if err != nil {
		return err
	}
	if !slices.Contains(families, constant.FirewallFamilyIPv6) {
		for _, rule := range rules {
			for _, source := range rule.Sources {
				if strings.Contains(source, ":") {
					return fmt.Errorf("IPv6 firewall support is disabled")
				}
			}
		}
	}
	return nil
}

func logFirewallIPv6Skipped(t *task.Task, subsystem string, count int) {
	if count == 0 {
		return
	}
	message := fmt.Sprintf("IPv6 firewall support is disabled; skipped %d IPv6 %s rule(s)", count, subsystem)
	if t != nil {
		t.Log(message)
	} else {
		global.LOG.Info(message)
	}
}
