package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"os"
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
	firewallsync "github.com/1Panel-dev/1Panel/agent/utils/firewall/sync"
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

type managedMutationRequest struct {
	Stored           model.FirewallRule
	Before           filter.FirewallRule
	After            filter.FirewallRule
	RuleSet          filter.RuleSet
	Locator          filter.Locator
	AdapterOperation filter.ChangeOperation
	Runtime          filter.Adapter
}

type firewallDockerRestartError struct {
	Err error
}

type firewallCompletedOperationError struct {
	Operation string
	Err       error
}

type firewallVerification struct {
	RuleSet filter.RuleSet
	Matched bool
}

type firewallRuleBatchItem struct {
	snapshot filter.RuleSet
	change   filter.RuleChange
}

type firewallSyncRule struct {
	dto.FirewallRuleSyncItem
	desired  filter.DesiredRule
	observed *filter.ObservedRule
	done     bool
}

type forwardingRuleSyncCandidate struct {
	rule forwarding.Rule
	err  error
}

type forwardingInventoryItem struct {
	ID        uint
	Rule      forwarding.Rule
	IsDesired bool
	IsRuntime bool
}

type observedInventoryCandidate struct {
	rule        filter.ObservedRule
	ruleKey     string
	instanceKey string
	claimed     bool
}

type firewallRuleCollisionIndex map[string][]filter.Action

type firewallSyncDesired[T any] struct {
	Value   T
	Payload dto.FirewallRuleSyncItem
	Err     error
}

func LoadPanelPort() string {
	if !global.IsMaster {
		return global.CONF.Base.Port
	}
	var portSetting model.Setting
	_ = global.CoreDB.Where("key = ?", "ServerPort").First(&portSetting).Error
	return portSetting.Value
}

func (s *FirewallService) syncPortWhitelist(ctx context.Context, provider filter.Provider, ports []firewall.PortWhitelist) ([]filter.FirewallRule, error) {
	required, err := firewall.RequiredPortWhitelist(ports)
	if err != nil {
		return nil, err
	}
	rules := whitelistRules(provider, firewall.ExpandPortWhitelist(customWhitelist(ports)), firewall.ExpandPortWhitelist(required))
	if provider != filter.ProviderIptables && provider != filter.ProviderNftables && len(rules) > 0 {
		client, err := s.baseClient()
		if err != nil {
			return nil, err
		}
		active, err := client.Status()
		if err != nil || !active {
			return nil, err
		}
	}
	client, err := s.firewallAdapter(provider)
	if err != nil {
		return nil, err
	}
	prepared := make([]dto.FirewallRuleCreateItem, 0, len(rules))
	var failures []error
	activeFamilies := make(map[filter.Family]bool)
	for _, rule := range rules {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if provider == filter.ProviderIptables || provider == filter.ProviderNftables {
			active, inspected := activeFamilies[rule.Scope.Family]
			if !inspected {
				initialized, bound, err := loadSystemFirewallFamilyStatus(string(provider), string(rule.Scope.Family))
				if err != nil {
					return nil, err
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
		return nil, errors.Join(failures...)
	}
	if len(prepared) == 0 {
		return nil, nil
	}
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()

	scopes := make([]filter.Scope, 0, len(prepared))
	for _, item := range prepared {
		scopes = append(scopes, item.Rule.Scope)
	}
	snapshots, err := readMutableFirewallRuleScopes(client, ctx, scopes)
	if err != nil {
		return nil, err
	}
	byScope := make(map[string]filter.RuleSet, len(snapshots))
	byScopeIdentity := make(map[string]firewallRuleCollisionIndex, len(snapshots))
	for _, snapshot := range snapshots {
		identities, err := observedFirewallRuleCollisionIndex(snapshot)
		if err != nil {
			return nil, err
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
			return nil, err
		}
		rule := item.Rule
		scope := rule.Scope
		identities := byScopeIdentity[scope.Key()]
		if err := identities.CheckDuplicate(rule); errors.Is(err, filter.ErrRuleOperation) {
			continue
		} else if err != nil {
			return nil, err
		}
		if err := identities.Check(rule); err != nil {
			return nil, err
		}
		key, err := filter.RuleMatchKey(rule)
		if err != nil {
			return nil, err
		}
		if byMatch == nil {
			stored, err := s.rules.List(ctx)
			if err != nil {
				return nil, err
			}
			byMatch = make(map[string][]filter.DesiredRule)
			for _, record := range stored {
				rules, err := compileStoredFirewallRules(ctx, record, client)
				if isFirewallPolicyIncompatible(err) {
					continue
				}
				if err != nil {
					return nil, err
				}
				for _, candidate := range rules {
					key, err := filter.RuleMatchKey(candidate.Rule)
					if err != nil {
						return nil, err
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
				return nil, collision
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
					return nil, err
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
			return nil, err
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
		return nil, errors.Join(append(failures, err)...)
	}
	if !verification.Matched {
		return nil, errors.Join(append(failures, filter.ErrVerificationFailed)...)
	}
	created := make([]filter.FirewallRule, 0)
	for index, plan := range plans {
		if persistenceErrors[index] != nil {
			continue
		}
		for _, command := range plan.Rules {
			created = append(created, command.Expected.Rule)
		}
	}
	return created, errors.Join(failures...)
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
		servicePorts = append(servicePorts, firewall.PortWhitelist{
			Type: serviceType, Port: ports[0], Protocol: "tcp",
			Sources: []string{"0.0.0.0/0", "::/0"},
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
	client, err := s.baseClient()
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
		if _, err := s.syncPortWhitelist(ctx, filter.Provider(client.Name()), allowances); err != nil {
			return err
		}
	}
	return settingRepo.UpdateOrCreate(constant.FirewallPortWhiteList, string(value))
}

func loadPortWhitelistSetting(db *gorm.DB) ([]firewall.PortWhitelist, error) {
	var setting model.Setting
	if err := db.Where("key = ?", constant.FirewallPortWhiteList).First(&setting).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		setting.Value = constant.FirewallPortWhiteListValue
	} else if err != nil {
		return nil, err
	}
	var rules []firewall.PortWhitelist
	err := json.Unmarshal([]byte(setting.Value), &rules)
	return rules, err
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
	switch provider {
	case constant.FirewallProviderNftables:
		return nftables_helper.LoadInitStatus(tab)
	case constant.FirewallProviderIptables:
		return iptables_helper.LoadInitStatus(tab)
	default:
		return false, false, fmt.Errorf("unsupported firewall provider: %s", provider)
	}
}

func loadSystemFirewallOverview(provider, chainGroup string) (dto.FirewallSubsystemStatus, error) {
	var status dto.FirewallSubsystemStatus
	var ipv4Err, ipv6Err error
	status.IPv4, ipv4Err = loadSystemFirewallFamilyInfo(provider, constant.FirewallFamilyIPv4)
	status.IPv6, ipv6Err = loadSystemFirewallFamilyInfo(provider, constant.FirewallFamilyIPv6)
	if chainGroup != "base" {
		return status, nil
	}
	if ipv4Err != nil {
		return status, ipv4Err
	}
	if provider == constant.FirewallProviderIptables {
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
	initialized, bound, err := loadSystemFirewallFamilyStatus(provider, family)
	return dto.FirewallBackendFamilyStatus{
		Available:   err == nil,
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

func (s *FirewallService) updateRuleOrder(ctx context.Context, ruleUUID string, targetPosition *int64, priority *int, description *string) error {
	if (targetPosition == nil) == (priority == nil) {
		return fmt.Errorf("%w: provide either position or priority", filter.ErrInvalidRule)
	}
	if ruleUUID == "" {
		return fmt.Errorf("%w: rule UUID is required", repo.ErrFirewallPersistenceInvalid)
	}
	stored, before, runtime, err := s.loadManagedRule(ctx, ruleUUID)
	if err != nil {
		return err
	}
	var snapshot filter.RuleSet
	var observed filter.ObservedRule
	if runtime.Provider() == filter.ProviderNftables {
		snapshot, err = readMutableFirewallRules(runtime, ctx, before.Rule.Scope)
		if err != nil {
			return err
		}
		observed, err = managedFirewallObserved(snapshot, before)
		if err != nil {
			return err
		}
	}
	capabilities, err := runtime.Capabilities(ctx)
	if err != nil {
		return err
	}
	after := before.Rule
	adapterOperation := filter.ChangeReorder
	switch {
	case capabilities.ExplicitPosition || capabilities.OwnedChains:
		if targetPosition == nil || *targetPosition < 1 {
			return fmt.Errorf("%w: target position is required", filter.ErrInvalidRule)
		}
		if runtime.Provider() == filter.ProviderNftables {
			if err := validateFirewallRulePosition(snapshot, before.Rule, *targetPosition); err != nil {
				return err
			}
		}
		after.OrderIndex = targetPosition
	case capabilities.ExplicitPriority:
		if before.Rule.NativeKind != filter.NativeKindRichRule {
			return fmt.Errorf("%w: only rich rules support explicit priority", filter.ErrUnsupportedScope)
		}
		if priority == nil {
			return fmt.Errorf("%w: priority is required", filter.ErrInvalidRule)
		}
		after.Priority = priority
		adapterOperation = filter.ChangeUpdate
	default:
		return fmt.Errorf("%w: provider does not support rule reordering", filter.ErrUnsupportedScope)
	}
	if description != nil {
		after.Description = strings.TrimSpace(*description)
	}
	after, err = prepareFirewallBackendRule(ctx, runtime, after)
	if err != nil {
		return err
	}
	metadataOnly, err := isFirewallMetadataOnlyUpdate(before.Rule, after, observed.Locator)
	if err != nil {
		return err
	}
	if metadataOnly {
		return s.updateRuleDescription(ctx, stored.UUID, after.Description)
	}
	if runtime.Provider() != filter.ProviderNftables {
		return s.replaceManagedRule(ctx, preparedManagedUpdate{Stored: stored, Before: before, After: after, Runtime: runtime})
	}
	if err := filter.GuardMutation(observed); err != nil {
		return err
	}
	return s.executeManagedMutation(ctx, managedMutationRequest{
		Stored: stored, Before: before.Rule, After: after, RuleSet: snapshot, Locator: observed.Locator,
		AdapterOperation: adapterOperation, Runtime: runtime,
	})
}

func (s *FirewallService) loadManagedRule(ctx context.Context, ruleUUID string) (model.FirewallRule, filter.DesiredRule, filter.Adapter, error) {
	stored, err := s.rules.GetByUUID(ctx, ruleUUID)
	if err != nil {
		return model.FirewallRule{}, filter.DesiredRule{}, nil, err
	}
	if stored.Origin != constant.FirewallRuleOriginCreated && stored.Origin != constant.FirewallRuleOriginAdopted {
		return model.FirewallRule{}, filter.DesiredRule{}, nil,
			fmt.Errorf("%w: only created or adopted rules can be changed", filter.ErrInvalidRule)
	}
	selected, err := s.selectedProviderForStoredRule(ctx, stored)
	if err != nil {
		return model.FirewallRule{}, filter.DesiredRule{}, nil, err
	}
	if err := checkFirewallRuleWhitelistProtection(selected, stored); err != nil {
		return model.FirewallRule{}, filter.DesiredRule{}, nil, err
	}
	runtime, err := s.resolveRuntime(ctx, selected)
	if err != nil {
		return model.FirewallRule{}, filter.DesiredRule{}, nil, err
	}
	desiredRules, err := compileStoredFirewallRules(ctx, stored, runtime)
	if err != nil {
		return model.FirewallRule{}, filter.DesiredRule{}, nil, err
	}
	if len(desiredRules) != 1 {
		return model.FirewallRule{}, filter.DesiredRule{}, nil,
			fmt.Errorf("%w: policy %q expands to %d target rules and cannot be edited atomically", filter.ErrUnsupportedScope, ruleUUID, len(desiredRules))
	}
	return stored, desiredRules[0], runtime, nil
}

func (s *FirewallService) selectedProviderForStoredRule(ctx context.Context, _ model.FirewallRule) (filter.Provider, error) {
	if s.selectedProvider != nil {
		return s.selectedProvider(ctx)
	}
	if s.adapters != nil {
		providers := make([]filter.Provider, 0, len(s.adapters))
		for provider := range s.adapters {
			providers = append(providers, provider)
		}
		if len(providers) == 1 {
			return providers[0], nil
		}
	}
	return "", fmt.Errorf("%w: selected provider is unavailable", filter.ErrProviderUnavailable)
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

func checkFirewallRuleWhitelistProtection(provider filter.Provider, record model.FirewallRule) error {
	ports, err := loadFirewallPortWhiteList()
	if err != nil {
		return err
	}
	rules, err := expandStoredFirewallRule(record, provider)
	if err != nil {
		return err
	}
	whitelist := filter.NewPortWhitelistIndex(ports)
	for _, rule := range rules {
		if whitelist.Matches(rule) {
			return filter.ErrProtectedRule
		}
	}
	return nil
}

func loadFirewallPortWhiteList() ([]firewall.PortWhitelist, error) {
	ports, err := loadPortWhitelistSetting(global.DB)
	if err != nil {
		return nil, err
	}
	return firewall.ValidatePortWhitelist(ports)
}

func expandStoredFirewallRule(rule model.FirewallRule, provider filter.Provider) ([]filter.FirewallRule, error) {
	if rule.CompatibilityError != "" {
		return nil, fmt.Errorf("%w: %s", filter.ErrUnsupportedScope, rule.CompatibilityError)
	}
	connectionStates := make([]string, 0)
	if rule.ConnectionStates != "" {
		connectionStates = strings.Split(rule.ConnectionStates, ",")
	}
	base := filter.FirewallRule{
		Protocol: rule.Protocol, SourceAddress: rule.SourceAddress, SourcePort: rule.SourcePort,
		DestinationAddress: rule.DestinationAddress, DestinationPort: rule.DestinationPort,
		Interface: rule.Interface, ConnectionStates: connectionStates,
		Action: filter.Action(rule.Action), Description: rule.Description,
	}
	if provider != filter.ProviderUFW && strings.EqualFold(strings.TrimSpace(base.Protocol), "all") &&
		strings.TrimSpace(base.SourcePort) == "" && strings.TrimSpace(base.DestinationPort) != "" {
		base.Protocol = "tcp/udp"
	}
	if provider == filter.ProviderFirewalld {
		base.Priority = rule.Priority
	}
	families := []filter.Family{filter.Family(rule.Family)}
	if provider != filter.ProviderFirewalld && families[0] == filter.FamilyInet {
		hasIPv4, hasIPv6 := false, false
		for _, address := range []string{base.SourceAddress, base.DestinationAddress} {
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
		compiled := base
		compiled.Scope = filter.Scope{Provider: provider, Family: family, Direction: filter.DirectionInput}
		switch provider {
		case filter.ProviderIptables, filter.ProviderNftables:
			compiled.Scope.Table, compiled.Scope.Chain = "filter", filter.IptablesInputChain
		case filter.ProviderFirewalld:
			compiled.Scope.Zone = filter.FirewalldInputZone
		case filter.ProviderUFW:
			compiled.Scope.Chain = filter.UFWInputChain
		default:
			return nil, fmt.Errorf("%w: unsupported firewall provider %q", filter.ErrProviderUnavailable, provider)
		}
		expanded, err := filter.ExpandAtomicRules(compiled)
		if err != nil {
			return nil, err
		}
		result = append(result, expanded...)
	}
	return result, nil
}

func compileStoredFirewallRules(ctx context.Context, stored model.FirewallRule, client filter.Adapter) ([]filter.DesiredRule, error) {
	rules, err := expandStoredFirewallRule(stored, client.Provider())
	if err != nil {
		return nil, err
	}
	capabilities, err := client.Capabilities(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]filter.DesiredRule, 0, len(rules))
	for ordinal, rule := range rules {
		prepared, err := prepareFirewallBackendRule(ctx, client, rule)
		if err != nil {
			return nil, err
		}
		ruleKey, err := filter.RuleKey(prepared)
		if err != nil {
			return nil, err
		}
		prepared.UUID = stored.UUID
		if ordinal > 0 {
			suffix := ruleKey
			if len(suffix) > 12 {
				suffix = suffix[:12]
			}
			prepared.UUID = fmt.Sprintf("%s-%d-%s", stored.UUID, ordinal+1, suffix)
		}
		desired := filter.DesiredRule{UUID: stored.UUID, Rule: prepared, RuleKey: ruleKey, Origin: filter.RuleOrigin(stored.Origin)}
		if capabilities.Marker {
			desired.Marker = "1panel-rule:" + prepared.UUID
		}
		result = append(result, desired)
	}
	return result, nil
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
	snapshots, err := readMutableFirewallRuleScopes(client, ctx, []filter.Scope{scope})
	if err != nil {
		return filter.RuleSet{}, err
	}
	return snapshots[0], nil
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

func readMutableFirewallRuleScopes(client filter.Adapter, ctx context.Context, scopes []filter.Scope) ([]filter.RuleSet, error) {
	snapshots, err := readFirewallRuleScopes(client, ctx, scopes)
	if err != nil {
		return nil, err
	}
	for _, snapshot := range snapshots {
		for _, notice := range snapshot.Notices {
			if notice.Code == filter.ScopeNoticeManagedScopeInactive || notice.Code == filter.ScopeNoticeManagedScopeMissing {
				return nil, fmt.Errorf("%w: managed firewall scope is unavailable", filter.ErrProviderUnavailable)
			}
		}
	}
	return snapshots, nil
}

func readFirewallRules(client filter.Adapter, ctx context.Context, scope filter.Scope) (filter.RuleSet, error) {
	snapshot, err := client.ListRules(ctx, scope)
	if err != nil {
		return filter.RuleSet{}, err
	}
	ports, err := loadFirewallPortWhiteList()
	if err != nil {
		return filter.RuleSet{}, err
	}
	return filter.ProtectRuleSet(snapshot, ports)
}

func firewallRulesByMarker(rules []filter.ObservedRule) map[string][]filter.ObservedRule {
	index := make(map[string][]filter.ObservedRule)
	for _, rule := range rules {
		if rule.Marker != "" {
			index[rule.Marker] = append(index[rule.Marker], rule)
		}
	}
	return index
}

func managedFirewallObserved(snapshot filter.RuleSet, desired filter.DesiredRule) (filter.ObservedRule, error) {
	matches := make([]filter.ObservedRule, 0, 1)
	for _, observed := range snapshot.Rules {
		if desired.Marker != "" {
			if observed.Marker != desired.Marker {
				continue
			}
		} else if desired.ObservedInstanceKey != "" {
			key, err := filter.InstanceKey(observed)
			if err != nil || key != desired.ObservedInstanceKey {
				continue
			}
		} else {
			key, err := firewallInventoryRuleKey(observed.Rule)
			wanted, wantErr := firewallInventoryRuleKey(desired.Rule)
			if err != nil || wantErr != nil || key != wanted {
				continue
			}
		}
		matches = append(matches, observed)
	}
	if len(matches) != 1 {
		return filter.ObservedRule{}, filter.ErrRuleStale
	}
	observed := matches[0]
	if observed.Protected || desired.Protected {
		return filter.ObservedRule{}, filter.ErrProtectedRule
	}
	if observed.Rule.Scope.Provider != filter.ProviderFirewalld && observed.Persistence != "" && observed.Persistence != filter.PersistenceStatusConverged {
		return filter.ObservedRule{}, filter.ErrRuleStale
	}
	if desired.Marker != "" && observed.ParseStatus == filter.ParseStatusOpaque {
		position := observed.Rule.OrderIndex
		observed.Rule = desired.Rule
		observed.Rule.OrderIndex = position
		observed.ParseStatus = filter.ParseStatusSupported
		observed.UncertainFields = nil
	} else {
		expected := desired.Rule
		if expected.Scope.Provider == filter.ProviderFirewalld {
			expected.Priority = observed.Rule.Priority
			expected.NativeKind = observed.Rule.NativeKind
			expected.OrderBucket = observed.Rule.OrderBucket
		}
		if !filter.ObservedRuleMatchesExpected(observed, expected) {
			return filter.ObservedRule{}, filter.ErrRuleStale
		}
	}
	return observed, nil
}

func (s *FirewallService) updateRuleDescription(ctx context.Context, ruleUUID, description string) error {
	stored, err := s.rules.GetByUUID(ctx, ruleUUID)
	if err != nil {
		return err
	}
	selected, err := s.selectedProviderForStoredRule(ctx, stored)
	if err != nil {
		return err
	}
	if err := checkFirewallRuleWhitelistProtection(selected, stored); err != nil {
		return err
	}
	if stored.Origin != constant.FirewallRuleOriginCreated && stored.Origin != constant.FirewallRuleOriginAdopted {
		return fmt.Errorf("%w: only created or adopted rules can be changed", filter.ErrInvalidRule)
	}
	description = strings.TrimSpace(description)
	if stored.Description == description {
		return nil
	}
	return s.rules.UpdateWithRevision(ctx, stored.UUID, stored.Revision, map[string]interface{}{"description": description})
}

func (s *FirewallService) replaceManagedRule(ctx context.Context, prepared preparedManagedUpdate) error {
	runtime := prepared.Runtime
	snapshot := filter.RuleSet{Scope: prepared.Before.Rule.Scope}
	remove, err := runtime.BuildCommands(snapshot, []filter.RuleChange{{
		Operation: filter.ChangeDelete, Before: &prepared.Before.Rule, CommandOnly: true,
	}})
	if err != nil {
		return err
	}
	create, err := runtime.BuildCommands(snapshot, []filter.RuleChange{{
		Operation: filter.ChangeCreate, After: &prepared.After, CommandOnly: true, Append: prepared.After.OrderIndex == nil,
	}})
	if err != nil {
		return err
	}
	updates, err := firewallRuleSemanticUpdates(prepared.After)
	if err != nil {
		return err
	}
	if runtime.Provider() == filter.ProviderFirewalld {
		updates["priority"] = prepared.After.Priority
	}
	remove.CommandOnly, create.CommandOnly = true, true
	if err := runtime.RunCommands(ctx, remove); err != nil {
		return err
	}
	saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	err = s.rules.UpdateWithRevision(saveCtx, prepared.Stored.UUID, prepared.Stored.Revision, updates)
	cancel()
	if err != nil {
		return err
	}
	runErr := runtime.RunCommands(ctx, create)
	if err := errors.Join(runErr, persistFirewallRules(ctx, runtime, create)); err != nil {
		return buserr.WithDetail("ErrFirewallRuleSavedApplyFailed", err.Error(), err)
	}
	return nil
}

func (s *FirewallService) executeManagedMutation(ctx context.Context, request managedMutationRequest) (err error) {
	defer func() {
		if err != nil && global.LOG != nil {
			global.LOG.Errorf("update firewall rule %s failed: %v", request.Stored.UUID, err)
		}
	}()
	before, after := request.Before, request.After
	appendRule := after.Scope.Provider == filter.ProviderUFW && after.OrderIndex != nil && *after.OrderIndex == maxObservedFirewallPosition(request.RuleSet)
	backendPlan, err := request.Runtime.BuildCommands(request.RuleSet, []filter.RuleChange{{
		Operation: request.AdapterOperation, Before: &before, After: &after,
		Locator: &request.Locator, Append: appendRule, CommandOnly: true,
	}})
	if err != nil {
		return err
	}
	backendPlan.CommandOnly = true
	updates, err := firewallRuleSemanticUpdates(request.After)
	if err != nil {
		return err
	}
	if len(backendPlan.Rules) == 1 && len(backendPlan.Rules[0].Commands) == 2 {
		commands := backendPlan.Rules[0]
		backendPlan.Rules[0].Commands = commands.Commands[:1]
		backendPlan.Rules[0].RollbackCommands = commands.RollbackCommands[:1]
		if err := request.Runtime.RunCommands(ctx, backendPlan); err != nil {
			return err
		}
		if after.Scope.Provider == filter.ProviderFirewalld {
			updates["priority"] = after.Priority
		}
		saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		err = s.rules.UpdateWithRevision(saveCtx, request.Stored.UUID, request.Stored.Revision, updates)
		cancel()
		if err != nil {
			return err
		}
		backendPlan.Rules[0].Commands = commands.Commands[1:]
		backendPlan.Rules[0].RollbackCommands = commands.RollbackCommands[1:]
		runErr := request.Runtime.RunCommands(ctx, backendPlan)
		if err := errors.Join(runErr, persistFirewallRules(ctx, request.Runtime, backendPlan)); err != nil {
			return buserr.WithDetail("ErrFirewallRuleSavedApplyFailed", err.Error(), err)
		}
		return nil
	}
	runErr := request.Runtime.RunCommands(ctx, backendPlan)
	if runErr != nil && request.Runtime.Provider() != filter.ProviderFirewalld {
		return runErr
	}
	if err := errors.Join(runErr, persistFirewallRules(ctx, request.Runtime, backendPlan)); err != nil {
		return err
	}
	sameContent, err := filter.SameRuleContent(request.Before, request.After)
	if err != nil {
		return err
	}
	if sameContent && request.Stored.Description == request.After.Description {
		return nil
	}
	return s.rules.UpdateWithRevision(ctx, request.Stored.UUID, request.Stored.Revision, updates)
}

func isFirewallPolicyIncompatible(err error) bool {
	return errors.Is(err, filter.ErrInvalidRule) || errors.Is(err, filter.ErrUnsupportedScope) ||
		errors.Is(err, filter.ErrInvalidScope) || errors.Is(err, filter.ErrCompositeRule)
}

func maxObservedFirewallPosition(snapshot filter.RuleSet) int64 {
	maximum := int64(snapshot.LastPosition)
	for _, observed := range snapshot.Rules {
		if observed.Locator.Position != nil && int64(*observed.Locator.Position) > maximum {
			maximum = int64(*observed.Locator.Position)
		}
	}
	return maximum
}

func applyFirewallChanges(client filter.Adapter, ctx context.Context, snapshot filter.RuleSet, changes []filter.RuleChange) (filter.CommandBatch, firewallVerification, error) {
	plan, err := client.BuildCommands(snapshot, changes)
	if err != nil {
		return filter.CommandBatch{}, firewallVerification{}, err
	}
	err = client.RunCommands(ctx, plan)
	if err == nil {
		err = persistFirewallRules(ctx, client, plan)
	}
	if err != nil {
		return plan, firewallVerification{}, err
	}
	verification, err := verifyFirewallCommands(ctx, client, plan)
	if plan.Provider == filter.ProviderUFW && len(plan.Rules) == 1 && plan.Rules[0].Operation == filter.ChangeAdopt && (err != nil || !verification.Matched) {
		if err == nil {
			err = filter.ErrVerificationFailed
		}
		err = buserr.WithDetail("ErrUFWRuleAdopt", err.Error(), err)
	}
	if err != nil {
		if plan.CreatesOnly() {
			return plan, verification, err
		}
		return plan, verification, rollbackFirewallPlan(ctx, client, plan, err)
	}
	if !verification.Matched && !plan.CreatesOnly() {
		if rollbackErr := restoreFirewallCommands(client, ctx, plan); rollbackErr != nil {
			return plan, verification, errors.Join(filter.ErrVerificationFailed, rollbackErr)
		}
	}
	return plan, verification, nil
}

func persistFirewallRules(ctx context.Context, client filter.Adapter, commands filter.CommandBatch) error {
	saver, ok := client.(filter.RuleSaver)
	if !ok {
		return nil
	}
	err := saver.SaveRules(ctx, commands.Scope)
	if err == nil || commands.CreatesOnly() || commands.CommandOnly {
		return err
	}
	if rollbackErr := restoreFirewallCommands(client, ctx, commands); rollbackErr != nil {
		return errors.Join(err, rollbackErr)
	}
	return err
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

func restoreFirewallCommands(client filter.Adapter, ctx context.Context, plan filter.CommandBatch) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	return client.Rollback(ctx, plan)
}

func readFirewallCommandResults(ctx context.Context, client filter.Adapter, plans ...filter.CommandBatch) ([]filter.RuleSet, map[string]map[string][]filter.ObservedRule, error) {
	scopes := make([]filter.Scope, 0, len(plans))
	for _, plan := range plans {
		scopes = append(scopes, plan.Scope)
		if plan.Provider == filter.ProviderUFW {
			related := plan.Scope
			if related.Family == filter.FamilyIPv4 {
				related.Family = filter.FamilyIPv6
			} else {
				related.Family = filter.FamilyIPv4
			}
			scopes = append(scopes, related)
		}
	}
	snapshots, err := listFirewallRuleScopes(client, ctx, scopes)
	if err != nil {
		return nil, nil, err
	}
	byIdentity := make(map[string]map[string][]filter.ObservedRule, len(snapshots))
	for _, snapshot := range snapshots {
		if snapshot.Scope.Provider == filter.ProviderFirewalld {
			canonical := make(map[string][]filter.ObservedRule)
			for _, observed := range snapshot.Rules {
				canonical[observed.Locator.Canonical] = append(canonical[observed.Locator.Canonical], observed)
			}
			byIdentity[snapshot.Scope.Key()] = canonical
		} else {
			byIdentity[snapshot.Scope.Key()] = firewallRulesByMarker(snapshot.Rules)
		}
	}
	return snapshots, byIdentity, nil
}

func verifyFirewallCommands(ctx context.Context, client filter.Adapter, plans ...filter.CommandBatch) (firewallVerification, error) {
	snapshots, byIdentity, err := readFirewallCommandResults(ctx, client, plans...)
	if err != nil {
		return firewallVerification{}, err
	}
	result := firewallVerification{Matched: true}
	for _, plan := range plans {
		result.Matched = false
		for _, snapshot := range snapshots {
			if snapshot.Scope.Key() != plan.Scope.Key() {
				continue
			}
			result.RuleSet = snapshot
			result.Matched = firewallCommandsMatch(plan, snapshot, snapshots, byIdentity)
			break
		}
		if !result.Matched {
			return result, nil
		}
	}
	return result, nil
}

func firewallCommandsMatch(commands filter.CommandBatch, current filter.RuleSet, all []filter.RuleSet, byIdentity map[string]map[string][]filter.ObservedRule) bool {
	for _, command := range commands.Rules {
		if commands.Provider == filter.ProviderFirewalld {
			previous, matched := 0, 0
			canonical := byIdentity[current.Scope.Key()]
			if command.Previous != nil {
				previous = len(canonical[command.Previous.Locator.Canonical])
			}
			for _, observed := range canonical[command.Expected.Locator.Canonical] {
				if observed.Persistence == filter.PersistenceStatusConverged {
					want, wantErr := filter.RuleKey(command.Expected.Rule)
					got, gotErr := filter.RuleKey(observed.Rule)
					if wantErr == nil && gotErr == nil && want == got {
						matched++
					}
				}
			}
			if command.Operation == filter.ChangeDelete {
				if previous != 0 {
					return false
				}
				continue
			}
			if command.Operation == filter.ChangeUpdate && command.Previous != nil && command.Previous.Locator.Canonical != command.Expected.Locator.Canonical && previous != 0 {
				return false
			}
			if matched != 1 {
				return false
			}
			continue
		}
		markerMatches, semanticMatches := 0, 0
		positionMatches := true
		requiresPosition := command.Expected.Locator.Position != nil && (commands.Provider == filter.ProviderUFW || command.Operation == filter.ChangeCreate || commands.Provider == filter.ProviderIptables && (command.Operation == filter.ChangeReorder || command.Operation == filter.ChangeUpdate && command.Expected.Rule.OrderIndex != nil))
		if requiresPosition {
			positionMatches = false
		}
		for _, rules := range all {
			if commands.Provider != filter.ProviderUFW && rules.Scope.Key() != commands.Scope.Key() {
				continue
			}
			candidates := byIdentity[rules.Scope.Key()][command.Expected.Marker]
			if command.Expected.Marker == "" {
				candidates = rules.Rules
			}
			for _, observed := range candidates {
				if observed.Marker != command.Expected.Marker {
					continue
				}
				markerMatches++
				if rules.Scope.Key() != commands.Scope.Key() {
					continue
				}
				same := false
				if commands.Provider == filter.ProviderUFW {
					same = observed.ParseStatus == filter.ParseStatusOpaque || filter.ObservedRuleMatchesExpected(observed, command.Expected.Rule)
				} else {
					want, wantErr := filter.RuleKey(command.Expected.Rule)
					got, gotErr := filter.RuleKey(observed.Rule)
					same = wantErr == nil && gotErr == nil && want == got
				}
				if same {
					semanticMatches++
				}
				if observed.Locator.Position != nil && command.Expected.Locator.Position != nil && *observed.Locator.Position == *command.Expected.Locator.Position {
					positionMatches = true
				}
			}
		}
		if command.Operation == filter.ChangeDelete {
			if markerMatches != 0 {
				return false
			}
			continue
		}
		if markerMatches != 1 || semanticMatches != 1 || !positionMatches {
			return false
		}
	}
	return true
}

func rollbackFirewallPlan(ctx context.Context, runtime filter.Adapter, plan filter.CommandBatch, cause error) error {
	if runtime == nil {
		return cause
	}
	if err := restoreFirewallCommands(runtime, ctx, plan); err != nil {
		return errors.Join(cause, fmt.Errorf("rollback applied firewall plan: %w", err))
	}
	return cause
}

func firewallRuleSemanticUpdates(rule filter.FirewallRule) (map[string]interface{}, error) {
	record, err := firewallRuleFromDomain(rule)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"family": record.Family, "protocol": record.Protocol,
		"source_address": record.SourceAddress, "source_port": record.SourcePort,
		"destination_address": record.DestinationAddress, "destination_port": record.DestinationPort,
		"interface": record.Interface, "connection_states": record.ConnectionStates, "action": record.Action,
		"description": record.Description, "compatibility_error": "",
	}, nil
}

func validateFirewallRulePosition(snapshot filter.RuleSet, rule filter.FirewallRule, target int64) error {
	if target < 1 {
		return fmt.Errorf("%w: target position must be positive", filter.ErrInvalidRule)
	}
	if rule.Scope.Provider == filter.ProviderUFW {
		minimum, maximum := positionBounds(snapshot)
		if target < minimum || target > maximum {
			return fmt.Errorf(
				"%w: target position %d is outside the %s range %d-%d",
				filter.ErrInvalidRule, target, rule.Scope.Family, minimum, maximum,
			)
		}
		return nil
	}
	maximum := maxObservedFirewallPosition(snapshot)
	if target > maximum {
		return fmt.Errorf("%w: target position %d is out of range 1-%d", filter.ErrInvalidRule, target, maximum)
	}
	return nil
}

func positionBounds(snapshot filter.RuleSet) (int64, int64) {
	minimum, maximum := int64(0), int64(0)
	for _, observed := range snapshot.Rules {
		if observed.Locator.Position == nil {
			continue
		}
		position := int64(*observed.Locator.Position)
		if minimum == 0 || position < minimum {
			minimum = position
		}
		if position > maximum {
			maximum = position
		}
	}
	return minimum, maximum
}

func isFirewallMetadataOnlyUpdate(before, after filter.FirewallRule, locator filter.Locator) (bool, error) {
	beforeKey, err := filter.RuleKey(before)
	if err != nil {
		return false, err
	}
	afterKey, err := filter.RuleKey(after)
	if err != nil {
		return false, err
	}
	if beforeKey != afterKey {
		return false, nil
	}
	if after.Scope.Provider == filter.ProviderFirewalld {
		return true, nil
	}
	return locator.Position != nil && after.OrderIndex != nil && *after.OrderIndex == int64(*locator.Position), nil
}

func (index firewallRuleCollisionIndex) Check(rule filter.FirewallRule) error {
	key, err := filter.RuleMatchKey(rule)
	if err != nil {
		return err
	}
	for _, action := range index[key] {
		if err := checkCollisionActions(rule.Action, action); err != nil {
			return err
		}
	}
	return nil
}

func checkCollisionActions(requested, existing filter.Action) error {
	if requested == existing {
		return fmt.Errorf("%w: equivalent rule already exists", filter.ErrRuleOperation)
	}
	if filter.OppositeActions(requested, existing) {
		return filter.ErrRuleConflict
	}
	return nil
}

func firewallRuleCollisions(stored []model.FirewallRule, provider filter.Provider) (firewallRuleCollisionIndex, error) {
	identities := make(firewallRuleCollisionIndex, len(stored))
	for _, candidate := range stored {
		rules, err := expandStoredFirewallRule(candidate, provider)
		if err != nil {
			continue
		}
		for _, rule := range rules {
			if err := identities.Add(rule); err != nil {
				return nil, err
			}
		}
	}
	return identities, nil
}

func (index firewallRuleCollisionIndex) Add(rule filter.FirewallRule) error {
	key, err := filter.RuleMatchKey(rule)
	if err != nil {
		return err
	}
	index[key] = append(index[key], rule.Action)
	return nil
}

func (s *FirewallService) restoreStoredFirewallRules(ctx context.Context, provider filter.Provider, t *task.Task) error {
	result, err := s.syncRules(ctx, "", dto.FirewallRuleSyncRequest{TargetProvider: provider}, t)
	if err != nil {
		return fmt.Errorf("restore database firewall rules: %w", err)
	}
	failures := make([]error, 0, len(result.Errors))
	for _, failure := range result.Errors {
		failures = append(failures, fmt.Errorf("rule %s: %s", failure.SourceUUID, failure.Error))
	}
	return errors.Join(failures...)
}

func executeFirewallRuleBatches(ctx context.Context, runtime filter.Adapter, items []firewallRuleBatchItem, t *task.Task, record func(int, error)) {
	groups := make([][]int, 0)
	byScope := make(map[string]int)
	for index, item := range items {
		key := string(item.change.Operation) + ":" + item.snapshot.Scope.Key()
		if runtime.Provider() == filter.ProviderUFW && item.change.Operation == filter.ChangeDelete {
			key = string(item.change.Operation)
		}
		group, exists := byScope[key]
		if !exists {
			group = len(groups)
			byScope[key] = group
			groups = append(groups, nil)
		}
		groups[group] = append(groups[group], index)
	}
	var plans []filter.CommandBatch
	var completed [][]int
	_, savesRules := runtime.(filter.RuleSaver)
	for _, group := range groups {
		if items[group[0]].change.Operation == filter.ChangeDelete {
			sort.SliceStable(group, func(i, j int) bool {
				left, right := items[group[i]].change.Locator, items[group[j]].change.Locator
				if left == nil || right == nil || left.Position == nil || right.Position == nil {
					return false
				}
				return *left.Position > *right.Position
			})
		}
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
			err := ctx.Err()
			if err == nil {
				var plan filter.CommandBatch
				plan, err = runtime.BuildCommands(items[batch[0]].snapshot, changes)
				if err == nil {
					if len(batch) > 1 && changes[0].Operation == filter.ChangeCreate && t != nil {
						t.Log(i18n.GetMsgWithMap("FirewallCreateBatchStep", map[string]interface{}{"backend": runtime.Provider(), "count": len(batch)}))
					}
					plan.CommandOnly = true
					err = runtime.RunCommands(ctx, plan)
					if err == nil && savesRules {
						plans = append(plans, plan)
						completed = append(completed, batch)
						continue
					}
				}
			}
			for _, index := range batch {
				record(index, err)
			}
		}
	}
	failures := persistFirewallRuleBatches(ctx, runtime, plans)
	for index, batch := range completed {
		for _, item := range batch {
			record(item, failures[index])
		}
	}
}

func (s *FirewallService) syncRules(ctx context.Context, _ string, request dto.FirewallRuleSyncRequest, t *task.Task) (result dto.FirewallRuleSyncResult, err error) {
	if request.SourceProvider != "" || request.ResetSource {
		return result, fmt.Errorf("%w: system synchronization reads rules from the database", filter.ErrInvalidRule)
	}
	if err := s.checkSelectedProvider(ctx, request.TargetProvider); err != nil {
		return result, err
	}
	result = dto.FirewallRuleSyncResult{Subsystem: "system", TargetProvider: request.TargetProvider}

	created, removed, unexecuted := 0, 0, 0
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
		} else if operation == task.TaskSync && cause == nil {
			t.Logf("%s %s", label, i18n.GetMsgByKey("FirewallSyncRuleUnchanged"))
		} else {
			t.LogWithStatus(label, cause)
		}
	}
	firewallWhitelistMu.Lock()
	ports, whitelistErr := loadFirewallPortWhiteList()
	var whitelistCreated []filter.FirewallRule
	if whitelistErr == nil {
		whitelistCreated, whitelistErr = s.syncPortWhitelist(ctx, request.TargetProvider, ports)
	}
	firewallWhitelistMu.Unlock()
	createdWhitelistKeys := make(map[string]bool, len(whitelistCreated))
	for _, rule := range whitelistCreated {
		key, err := filter.RuleMatchKey(rule)
		if err != nil {
			return result, err
		}
		createdWhitelistKeys[key] = true
		result.Total++
		record("TaskCreate", &firewallSyncRule{FirewallRuleSyncItem: dto.FirewallRuleSyncItem{SourceUUID: rule.UUID, Rule: &rule}}, nil, false)
	}
	if t != nil {
		t.LogWithStatus(i18n.GetMsgByKey("FirewallSyncWhitelistStep"), whitelistErr)
	}
	if whitelistErr != nil {
		return result, whitelistErr
	}
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	if err := ctx.Err(); err != nil {
		return result, err
	}
	runtime, rules, snapshots, err := s.loadFirewallSyncRules(ctx, request)
	if err != nil {
		return result, err
	}

	defer func() {
		if t != nil {
			t.Log(i18n.GetMsgWithMap("FirewallSyncOperationsResult", map[string]interface{}{"created": created, "removed": removed, "failed": result.Failed, "skipped": unexecuted, "unchanged": result.Skipped - unexecuted}))
		}
	}()
	for _, rule := range rules {
		if rule.Status == firewallsync.StatusExisting && rule.Rule != nil {
			key, err := filter.RuleMatchKey(*rule.Rule)
			if err != nil {
				return result, err
			}
			if createdWhitelistKeys[key] {
				rule.done = true
				continue
			}
		}
		if rule.Status != firewallsync.StatusRemove {
			result.Total++
		}
		if rule.Status == firewallsync.StatusBlocked {
			record("TaskSync", rule, errors.New(rule.Reason), false)
		}
	}
	for _, rule := range rules {
		if !rule.done && rule.Status == firewallsync.StatusExisting {
			record("TaskSync", rule, nil, false)
		}
	}
	byScope := make(map[string]filter.RuleSet, len(snapshots))
	for _, snapshot := range snapshots {
		byScope[snapshot.Scope.Key()] = snapshot
	}
	for _, operation := range []filter.ChangeOperation{filter.ChangeDelete, filter.ChangeCreate} {
		name := "TaskCreate"
		if operation == filter.ChangeDelete {
			name = "TaskDelete"
		}
		pending := make([]*firewallSyncRule, 0)
		items := make([]firewallRuleBatchItem, 0)
		for _, rule := range rules {
			if rule.done {
				continue
			}
			if operation == filter.ChangeDelete && rule.observed == nil || operation == filter.ChangeCreate && rule.Status != firewallsync.StatusReady {
				continue
			}
			if operation == filter.ChangeCreate {
				var cause error
				for id, err := range failedRemovals {
					if id == rule.SourceUUID || strings.HasPrefix(id, rule.SourceUUID+"-") {
						cause = err
						break
					}
				}
				if cause != nil {
					record(name, rule, cause, true)
					continue
				}
			}
			item := firewallRuleBatchItem{snapshot: filter.RuleSet{Scope: rule.Rule.Scope}}
			if operation == filter.ChangeDelete {
				item.snapshot = byScope[rule.Rule.Scope.Key()]
				item.change, err = firewallDeleteChange(*rule.observed, rule.desired)
				if err != nil {
					record(name, rule, err, false)
					continue
				}
				if runtime.Provider() == filter.ProviderUFW {
					item.snapshot.Rules = []filter.ObservedRule{*rule.observed}
				}
			} else {
				after := *rule.Rule
				after.OrderIndex = nil
				item.change = filter.RuleChange{Operation: operation, After: &after, Append: true}
			}
			pending = append(pending, rule)
			items = append(items, item)
		}
		executeFirewallRuleBatches(ctx, runtime, items, t, func(index int, failure error) {
			record(name, pending[index], failure, false)
		})
	}
	return result, nil
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

func (s *FirewallService) saveFirewallRule(ctx context.Context, record *model.FirewallRule) error {
	if err := s.rules.Create(ctx, record); err != nil {
		message := "FirewallCreateRulePersistenceFailed"
		if record.Origin == constant.FirewallRuleOriginAdopted {
			message = "FirewallAdoptRulePersistenceFailed"
		}
		return fmt.Errorf("%s: %w", i18n.GetMsgByKey(message), err)
	}
	return nil
}

func (s *FirewallService) createRules(ctx context.Context, request dto.FirewallRuleCreate, t *task.Task) (result dto.FirewallRuleCreateResponse, taskErr error) {
	var firstFailure error
	defer func() {
		if t != nil {
			t.Log(i18n.GetMsgWithMap("FirewallCreateRulesResult", map[string]interface{}{
				"succeeded": result.Succeeded, "failed": result.Failed, "skipped": result.Skipped,
			}))
		}
		if taskErr == nil {
			taskErr = firstFailure
		}
	}()
	type createItem struct {
		index, part, count int
		request            dto.FirewallRuleCreateItem
		stored             model.FirewallRule
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
	items := make([]createItem, 0, len(request.Items))
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
			record(createItem{index: index, request: item}, "failed", err)
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
			prepared, err := prepareFirewallCreateRule(ctx, runtime, child)
			if err != nil {
				record(entry, "failed", err)
				continue
			}
			entry.request = prepared
			items = append(items, entry)
		}
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		for _, item := range items {
			record(item, "failed", err)
		}
		return result, err
	}
	identities, err := firewallRuleCollisions(stored, selected)
	if err != nil {
		return result, err
	}
	valid := items[:0]
	for _, item := range items {
		if stop == nil {
			stop = ctx.Err()
		}
		if stop != nil {
			record(item, "skipped", stop)
			continue
		}
		rule := item.request.Rule
		if err := identities.CheckDuplicate(rule); errors.Is(err, filter.ErrRuleOperation) {
			record(item, "skipped", err)
			continue
		} else if err != nil {
			record(item, "failed", err)
			continue
		}
		item.stored, err = firewallRuleModelForCreate(rule, item.request, constant.FirewallRuleOriginCreated)
		if err != nil {
			record(item, "failed", err)
			continue
		}
		item.stored.UUID = uuid.NewString()
		rule.UUID = item.stored.UUID
		item.request.Rule = rule
		if err := identities.Add(rule); err != nil {
			record(item, "failed", err)
			continue
		}
		valid = append(valid, item)
	}
	changes := make([]firewallRuleBatchItem, 0, len(valid))
	for index := range valid {
		rule := &valid[index].request.Rule
		changes = append(changes, firewallRuleBatchItem{
			snapshot: filter.RuleSet{Scope: rule.Scope},
			change:   filter.RuleChange{Operation: filter.ChangeCreate, After: rule, Append: true},
		})
	}
	executeFirewallRuleBatches(ctx, runtime, changes, t, func(index int, err error) {
		item := valid[index]
		if err != nil {
			if errors.Is(err, filterfirewalld.ErrAlreadyEnabled) {
				record(item, "skipped", err)
			} else if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				record(item, "skipped", err)
				stop = err
			} else {
				record(item, "failed", fmt.Errorf("%s: %w", i18n.GetMsgByKey("FirewallCreateRuleExecutionFailed"), err))
			}
			return
		}
		persistCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()
		if err := s.saveFirewallRule(persistCtx, &item.stored); err != nil {
			record(item, "failed", err)
		} else {
			record(item, "succeeded", nil)
		}
	})
	return result, stop
}

func validateFirewallCreateBatch(request dto.FirewallRuleCreate, provider filter.Provider) error {
	count := 0
	if len(request.Items) > filter.MaxAtomicExpansion {
		return fmt.Errorf("create or import at most %d rules per batch (after expansion)", filter.MaxAtomicExpansion)
	}
	for _, item := range request.Items {
		rules, expandErr := expandFirewallCreateRule(item, provider)
		if errors.Is(expandErr, filter.ErrExpansionLimit) {
			return fmt.Errorf("create or import at most %d rules per batch (after expansion)", filter.MaxAtomicExpansion)
		}
		count += len(rules)
		if count > filter.MaxAtomicExpansion {
			return fmt.Errorf("create or import at most %d rules per batch (after expansion)", filter.MaxAtomicExpansion)
		}
	}
	return nil
}

func expandFirewallCreateRule(item dto.FirewallRuleCreateItem, provider filter.Provider) ([]filter.FirewallRule, error) {
	if item.SourceKind == constant.FirewallRuleSourceImported {
		source := item.Rule.Scope.Normalize().Provider
		if source == "" {
			source = provider
		}
		rules, err := filter.ExpandAtomicRules(applySelectedProviderScopeDefaults(item.Rule, source))
		if err != nil {
			return nil, err
		}
		var converted []filter.FirewallRule
		for _, sourceRule := range rules {
			policy, err := firewallRuleFromDomain(sourceRule)
			if err != nil {
				return nil, err
			}
			targetRules, err := expandStoredFirewallRule(policy, provider)
			if err != nil {
				return nil, err
			}
			converted = append(converted, targetRules...)
		}
		return converted, nil
	}
	rule := applySelectedProviderScopeDefaults(item.Rule, provider)
	if provider == filter.ProviderUFW && strings.TrimSpace(rule.DestinationPort) != "" {
		protocol := strings.ToLower(strings.TrimSpace(rule.Protocol))
		if protocol == "" || protocol == "all" || protocol == "any" {
			return filter.ExpandAtomicRules(rule)
		}
	}
	return []filter.FirewallRule{rule}, nil
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
	selected := runtime.Provider()
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

func observedFirewallRuleCollisionIndex(snapshot filter.RuleSet) (firewallRuleCollisionIndex, error) {
	index := make(firewallRuleCollisionIndex, len(snapshot.Rules))
	for _, observed := range snapshot.Rules {
		if observed.ParseStatus != filter.ParseStatusSupported {
			continue
		}
		if err := index.Add(observed.Rule); err != nil {
			return nil, err
		}
	}
	return index, nil
}

func firewallRuleFromDomain(rule filter.FirewallRule) (model.FirewallRule, error) {
	normalized, err := filter.NormalizeRule(rule)
	if err != nil {
		return model.FirewallRule{}, err
	}
	switch normalized.NativeKind {
	case "", filter.NativeKindRule, filter.NativeKindZonePort, filter.NativeKindRichRule, filter.NativeKindUFWRule:
	default:
		return model.FirewallRule{}, fmt.Errorf("%w: native rule %q cannot be stored as a provider-neutral policy", filter.ErrUnsupportedScope, normalized.NativeKind)
	}
	record := model.FirewallRule{
		Family:             string(normalized.Scope.Family),
		Protocol:           normalized.Protocol,
		SourceAddress:      normalized.SourceAddress,
		SourcePort:         normalized.SourcePort,
		DestinationAddress: normalized.DestinationAddress,
		DestinationPort:    normalized.DestinationPort,
		Interface:          normalized.Interface,
		ConnectionStates:   strings.Join(normalized.ConnectionStates, ","),
		Action:             string(normalized.Action),
		Description:        normalized.Description,
	}
	return record, nil
}

func sortFirewallRules(rules []model.FirewallRule, provider filter.Provider) {
	sort.SliceStable(rules, func(i, j int) bool {
		left, right := rules[i], rules[j]
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
func firewallRuleModelForCreate(rule filter.FirewallRule, request dto.FirewallRuleCreateItem, origin string) (model.FirewallRule, error) {
	record, err := firewallRuleFromDomain(rule)
	if err != nil {
		return model.FirewallRule{}, err
	}
	record.Origin = origin
	record.Owner = strings.TrimSpace(request.SourceKind)
	if sourceID := strings.TrimSpace(request.SourceID); sourceID != "" {
		record.Owner += ":" + sourceID
	}
	return record, nil
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

func (s *FirewallService) loadFirewallSyncRules(ctx context.Context, request dto.FirewallRuleSyncRequest) (filter.Adapter, []*firewallSyncRule, []filter.RuleSet, error) {
	if request.SourceProvider != "" || request.ResetSource {
		return nil, nil, nil, fmt.Errorf("%w: system synchronization reads rules from the database", filter.ErrInvalidRule)
	}
	if err := s.checkSelectedProvider(ctx, request.TargetProvider); err != nil {
		return nil, nil, nil, err
	}
	runtime, err := s.firewallAdapter(request.TargetProvider)
	if err != nil {
		return nil, nil, nil, err
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	sortFirewallRules(stored, request.TargetProvider)
	ports, err := loadFirewallPortWhiteList()
	if err != nil {
		return nil, nil, nil, err
	}
	required, err := firewall.RequiredPortWhitelist(ports)
	if err != nil {
		return nil, nil, nil, err
	}
	whitelistDesired := whitelistRules(request.TargetProvider, firewall.ExpandPortWhitelist(customWhitelist(ports)), firewall.ExpandPortWhitelist(required))
	rules := make([]*firewallSyncRule, 0, len(stored))
	preservedMarkers := make(map[string]bool)
	compileFailed := false
	whitelist := filter.NewPortWhitelistIndex(ports)
	for _, record := range stored {
		desired, preserved, err := s.compileRestorableFirewallRules(ctx, record, runtime, required)
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
			native.Protected = whitelist.Matches(rule)
			rules = append(rules, &firewallSyncRule{desired: native, FirewallRuleSyncItem: dto.FirewallRuleSyncItem{SourceUUID: record.UUID, Rule: &rule}})
		}
	}
	identities := make(firewallRuleCollisionIndex, len(rules))
	for _, existing := range rules {
		if existing.Rule != nil {
			if err := identities.Add(*existing.Rule); err != nil {
				return nil, nil, nil, err
			}
		}
	}
	for _, candidate := range whitelistDesired {
		prepared, err := prepareFirewallCreateRule(ctx, runtime, dto.FirewallRuleCreateItem{Rule: candidate})
		if err != nil {
			return nil, nil, nil, err
		}
		rule := prepared.Rule
		if err := identities.CheckDuplicate(rule); errors.Is(err, filter.ErrRuleOperation) {
			continue
		} else if err != nil {
			return nil, nil, nil, err
		}
		if err := identities.Add(rule); err != nil {
			return nil, nil, nil, err
		}
		key, err := filter.RuleKey(rule)
		if err != nil {
			return nil, nil, nil, err
		}
		rule.UUID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(key)).String()
		rules = append(rules, &firewallSyncRule{
			desired:              filter.DesiredRule{UUID: rule.UUID, Rule: rule, RuleKey: key, Origin: filter.RuleOriginCreated, Protected: true},
			FirewallRuleSyncItem: dto.FirewallRuleSyncItem{SourceUUID: rule.UUID, Rule: &rule},
		})
	}
	scopes := filter.ManagedInputScopes(request.TargetProvider)
	if compileFailed {
		scopes = slices.DeleteFunc(scopes, func(scope filter.Scope) bool {
			return !slices.ContainsFunc(rules, func(rule *firewallSyncRule) bool {
				return rule.Rule != nil && rule.Rule.Scope.Key() == scope.Key()
			})
		})
	}
	snapshots := make([]filter.RuleSet, 0, len(scopes))
	for _, group := range firewallScopeReadGroups(scopes) {
		current, err := readMutableFirewallRuleScopes(runtime, ctx, group)
		if errors.Is(err, filter.ErrFamilyUnavailable) {
			for _, rule := range rules {
				if rule.Rule != nil && slices.ContainsFunc(group, func(scope filter.Scope) bool { return scope.Key() == rule.Rule.Scope.Key() }) {
					rule.Status, rule.Reason = firewallsync.StatusBlocked, err.Error()
				}
			}
			continue
		}
		if err != nil {
			return nil, nil, nil, err
		}
		snapshots = append(snapshots, current...)
	}
	for _, snapshot := range snapshots {
		scope := snapshot.Scope
		desired := make([]filter.DesiredRule, 0)
		byUUID := make(map[string]*firewallSyncRule)
		for _, rule := range rules {
			if rule.Rule != nil && rule.Rule.Scope.Key() == scope.Key() {
				desired = append(desired, rule.desired)
				byUUID[rule.desired.Rule.UUID] = rule
			}
		}
		inventory, err := mergeFirewallInventory(filter.InventoryMergeInput{Observed: snapshot.Rules, Desired: desired})
		if err != nil {
			return nil, nil, nil, err
		}
		for _, item := range inventory {
			if item.Desired == nil {
				if compileFailed || item.Observed == nil || !strings.HasPrefix(item.Observed.Marker, "1panel-rule:") || preservedMarkers[item.Observed.Marker] {
					continue
				}
				observed := item.Observed
				if observed.Rule.Scope.Chain == filter.BasicBeforeChain {
					continue
				}
				rule := &firewallSyncRule{observed: observed, FirewallRuleSyncItem: dto.FirewallRuleSyncItem{
					SourceUUID: strings.TrimPrefix(observed.Marker, "1panel-rule:"), Rule: &observed.Rule, Status: firewallsync.StatusRemove,
					ReasonCode: firewallsync.ReasonManagedOnlyInTarget, Reason: firewallSyncReasonMessage(firewallsync.ReasonManagedOnlyInTarget),
				}}
				if observed.Protected || observed.ParseStatus == filter.ParseStatusOpaque {
					rule.Status, rule.ReasonCode = firewallsync.StatusBlocked, firewallsync.ReasonUnsafeRemoval
					rule.Reason = firewallSyncReasonMessage(rule.ReasonCode)
				}
				rules = append(rules, rule)
				continue
			}
			rule := byUUID[item.Desired.Rule.UUID]
			rule.observed = item.Observed
			if item.Match == filter.InventoryMatchExact && item.Observed != nil && scope.Provider == filter.ProviderFirewalld {
				rule.Rule.Priority = item.Observed.Rule.Priority
				rule.Rule.NativeKind = item.Observed.Rule.NativeKind
				rule.Rule.OrderBucket = item.Observed.Rule.OrderBucket
				rule.desired.Rule = *rule.Rule
				rule.desired.RuleKey, err = filter.RuleKey(*rule.Rule)
				if err != nil {
					return nil, nil, nil, err
				}
			}
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
	}
	return runtime, rules, snapshots, nil
}

func (s *FirewallService) compileRestorableFirewallRules(ctx context.Context, stored model.FirewallRule, client filter.Adapter, required []firewall.PortWhitelist) (restorable, preserved []filter.DesiredRule, err error) {
	provider := client.Provider()
	compiled, err := compileStoredFirewallRules(ctx, stored, client)
	if err != nil {
		return nil, nil, err
	}
	if (provider != filter.ProviderIptables && provider != filter.ProviderNftables) || !strings.HasPrefix(stored.Owner, constant.FirewallRuleSourceSecurity+":"+constant.FirewallSystemAcceptedPortSourcePrefix) {
		return compiled, nil, nil
	}
	requiredPorts := firewall.ExpandPortWhitelist(required)
	for _, desired := range compiled {
		covered := false
		for _, port := range requiredPorts {
			covered, err = filter.SameRuleContent(desired.Rule, firewall.RuleForSystemPort(provider, firewall.SystemPort(port)))
			if err != nil {
				return nil, nil, err
			}
			if covered {
				break
			}
		}
		if covered {
			preserved = append(preserved, desired)
		} else {
			restorable = append(restorable, desired)
		}
	}
	return restorable, preserved, nil
}

func firewallInventoryRuleKey(rule filter.FirewallRule) (string, error) {
	if rule.Scope.Provider == filter.ProviderFirewalld {
		key, err := filter.RuleMatchKey(rule)
		return key + ":" + string(rule.Action), err
	}
	return filter.RuleKey(rule)
}

func mergeFirewallInventory(input filter.InventoryMergeInput) ([]filter.InventoryItem, error) {
	candidates := make([]observedInventoryCandidate, len(input.Observed))
	byRuleKey := make(map[string][]int)
	byInstanceKey := make(map[string][]int)
	byMarker := make(map[string][]int)
	bySemanticKey := make(map[string][]int)
	byPartialRuleKey := make(map[string][]int)
	for index, observed := range input.Observed {
		candidate := observedInventoryCandidate{rule: observed}
		if marker := strings.TrimSpace(candidate.rule.Marker); marker != "" {
			markerKey := candidate.rule.Rule.Scope.Key() + "\x00" + marker
			byMarker[markerKey] = append(byMarker[markerKey], index)
		}
		if observed.ParseStatus == filter.ParseStatusSupported {
			normalized, err := filter.NormalizeRule(observed.Rule)
			if err != nil {
				return nil, fmt.Errorf("normalize observed firewall rule %d: %w", index, err)
			}
			candidate.rule.Rule = normalized
			candidate.ruleKey, err = filter.RuleKey(normalized)
			if err != nil {
				return nil, err
			}
			key, err := firewallInventoryRuleKey(normalized)
			if err != nil {
				return nil, err
			}
			byRuleKey[key] = append(byRuleKey[key], index)
			if instanceKey, err := filter.InstanceKey(candidate.rule); err == nil {
				candidate.instanceKey = instanceKey
				candidate.rule.InstanceKey = instanceKey
				byInstanceKey[instanceKey] = append(byInstanceKey[instanceKey], index)
			}
		}
		if observed.ParseStatus != filter.ParseStatusOpaque {
			rule := candidate.rule.Rule
			protocolUnknown, supportedFields := false, true
			for _, field := range observed.UncertainFields {
				if field != filter.ObservedFieldProtocol {
					supportedFields = false
					break
				}
				protocolUnknown = true
			}
			if supportedFields {
				key := candidate.ruleKey
				var err error
				if protocolUnknown {
					rule.Protocol = "tcp"
					key, err = filter.RuleKey(rule)
				} else if key == "" {
					key, err = filter.RuleKey(rule)
				}
				if err == nil {
					key += "\x00" + strings.TrimSpace(observed.Marker)
					if protocolUnknown {
						key = "protocol\x00" + key
						byPartialRuleKey[key] = append(byPartialRuleKey[key], index)
					} else if observed.ParseStatus != filter.ParseStatusSupported {
						key = "exact\x00" + key
						byPartialRuleKey[key] = append(byPartialRuleKey[key], index)
					} else {
						bySemanticKey[key] = append(bySemanticKey[key], index)
					}
				}
			}
		}
		candidates[index] = candidate
	}

	normalizedDesired := make([]filter.DesiredRule, 0, len(input.Desired))
	desiredMatches := make(map[int]int)
	desiredMatchStates := make([]filter.InventoryMatch, 0, len(input.Desired))
	for _, desired := range input.Desired {
		normalized, err := filter.NormalizeRule(desired.Rule)
		if err != nil {
			return nil, fmt.Errorf("normalize desired firewall rule %q: %w", desired.UUID, err)
		}
		desired.Rule = normalized
		calculatedKey, err := filter.RuleKey(normalized)
		if err != nil {
			return nil, err
		}
		if desired.RuleKey != "" && desired.RuleKey != calculatedKey {
			return nil, fmt.Errorf("%w: desired rule %q key does not match its semantics", filter.ErrInvalidRule, desired.UUID)
		}
		desired.RuleKey = calculatedKey

		match, matchState := findObservedInventoryMatch(desired, candidates, byRuleKey, byInstanceKey, byMarker, bySemanticKey, byPartialRuleKey)
		normalizedIndex := len(normalizedDesired)
		normalizedDesired = append(normalizedDesired, desired)
		desiredMatchStates = append(desiredMatchStates, matchState)
		if match >= 0 {
			candidates[match].claimed = true
			desiredMatches[match] = normalizedIndex
		}
	}

	items := make([]filter.InventoryItem, 0, len(candidates)+len(normalizedDesired))
	matchedDesired := make(map[int]struct{}, len(desiredMatches))
	for index := range candidates {
		candidate := &candidates[index]
		if desiredIndex, exists := desiredMatches[index]; exists {
			desired := normalizedDesired[desiredIndex]
			observed := candidate.rule
			match := desiredMatchStates[desiredIndex]
			if match == filter.InventoryMatchExact && observed.ParseStatus != filter.ParseStatusSupported {
				orderIndex := observed.Rule.OrderIndex
				observed.Rule = desired.Rule
				observed.Rule.OrderIndex = orderIndex
				observed.ParseStatus = filter.ParseStatusSupported
				observed.UncertainFields = nil
			}
			displayRule := observed.Rule
			displayRule.Description = desired.Rule.Description
			state := inventoryStateForDesired(desired, match)
			if observed.Protected {
				state = filter.InventoryStateProtected
			} else if observed.Persistence != "" && observed.Persistence != filter.PersistenceStatusConverged {
				state = filter.InventoryStateDrifted
			}
			items = append(items, filter.InventoryItem{
				Rule:     displayRule,
				Observed: &observed,
				Desired:  &desired,
				State:    state,
				Match:    match,
			})
			matchedDesired[desiredIndex] = struct{}{}
			continue
		}
		observed := candidate.rule
		state := filter.InventoryStateExternal
		if observed.Protected {
			state = filter.InventoryStateProtected
		} else if _, protected := input.ProtectedObservedKeys[candidate.ruleKey]; protected {
			state = filter.InventoryStateProtected
		}
		match := filter.InventoryMatchNone
		if observed.ParseStatus != filter.ParseStatusSupported {
			match = filter.InventoryMatchOpaque
		}
		items = append(items, filter.InventoryItem{Rule: observed.Rule, Observed: &observed, State: state, Match: match})
	}
	for index, desired := range normalizedDesired {
		if _, matched := matchedDesired[index]; matched {
			continue
		}
		desiredCopy := desired
		match := desiredMatchStates[index]
		items = append(items, filter.InventoryItem{
			Rule:    desired.Rule,
			Desired: &desiredCopy,
			State:   inventoryStateForDesired(desired, match),
			Match:   match,
		})
	}
	return items, nil
}

func findObservedInventoryMatch(desired filter.DesiredRule, candidates []observedInventoryCandidate, byRuleKey, byInstanceKey, byMarker, bySemanticKey, byPartialRuleKey map[string][]int) (int, filter.InventoryMatch) {
	if marker := strings.TrimSpace(desired.Marker); marker != "" {
		match, status := matchUnclaimedSemanticCandidate(desired, marker, candidates, bySemanticKey, byPartialRuleKey)
		if status != filter.InventoryMatchMissing {
			return match, status
		}
		markerKey := desired.Rule.Scope.Key() + "\x00" + marker
		match, status = uniqueUnclaimedCandidate(byMarker, markerKey, candidates)
		if match >= 0 && candidates[match].rule.ParseStatus != filter.ParseStatusOpaque &&
			!filter.ObservedRuleMatchesExpected(candidates[match].rule, desired.Rule) {
			return match, filter.InventoryMatchChanged
		}
		if status != filter.InventoryMatchMissing {
			return match, status
		}
		if desired.Origin == filter.RuleOriginAdopted {
			match, status = matchUnclaimedSemanticCandidate(desired, "", candidates, bySemanticKey, byPartialRuleKey)
			if status != filter.InventoryMatchMissing {
				if match >= 0 {
					return match, filter.InventoryMatchChanged
				}
				return match, status
			}
		}
		legacyMarker := "1panel-rule:" + strings.TrimSpace(desired.UUID)
		if legacyMarker != "1panel-rule:" && legacyMarker != marker {
			match, status = matchUnclaimedSemanticCandidate(desired, legacyMarker, candidates, bySemanticKey, byPartialRuleKey)
			if status != filter.InventoryMatchMissing {
				if match >= 0 {
					return match, filter.InventoryMatchChanged
				}
				return match, status
			}
		}
		return match, status
	}
	var match int
	if desired.ObservedInstanceKey != "" {
		match = firstUnclaimedCandidate(byInstanceKey, desired.ObservedInstanceKey, candidates)
	} else {
		key, err := firewallInventoryRuleKey(desired.Rule)
		if err != nil {
			return -1, filter.InventoryMatchMissing
		}
		match = firstUnclaimedCandidate(byRuleKey, key, candidates)
	}
	if match < 0 {
		return -1, filter.InventoryMatchMissing
	}
	return match, filter.InventoryMatchExact
}

func firstUnclaimedCandidate(byKey map[string][]int, key string, candidates []observedInventoryCandidate) int {
	indices := byKey[key]
	for len(indices) > 0 && candidates[indices[0]].claimed {
		indices = indices[1:]
	}
	if len(indices) == 0 {
		delete(byKey, key)
		return -1
	}
	byKey[key] = indices
	return indices[0]
}

func uniqueUnclaimedCandidate(byKey map[string][]int, key string, candidates []observedInventoryCandidate) (int, filter.InventoryMatch) {
	match := firstUnclaimedCandidate(byKey, key, candidates)
	if match < 0 {
		return -1, filter.InventoryMatchMissing
	}
	for _, index := range byKey[key][1:] {
		if !candidates[index].claimed {
			return -1, filter.InventoryMatchAmbiguous
		}
	}
	return match, filter.InventoryMatchExact
}

func matchUnclaimedSemanticCandidate(desired filter.DesiredRule, marker string, candidates []observedInventoryCandidate, bySemanticKey, byPartialRuleKey map[string][]int) (int, filter.InventoryMatch) {
	key := desired.RuleKey + "\x00" + marker
	if match := firstUnclaimedCandidate(bySemanticKey, key, candidates); match >= 0 {
		return match, filter.InventoryMatchExact
	}
	if len(byPartialRuleKey) == 0 {
		return -1, filter.InventoryMatchMissing
	}
	match, status := uniqueUnclaimedCandidate(byPartialRuleKey, "exact\x00"+key, candidates)
	if status == filter.InventoryMatchAmbiguous {
		return match, status
	}
	protocolIndependent := desired.Rule
	protocolIndependent.Protocol = "tcp"
	partialKey, err := filter.RuleKey(protocolIndependent)
	if err != nil {
		return match, status
	}
	partial, partialStatus := uniqueUnclaimedCandidate(byPartialRuleKey, "protocol\x00"+partialKey+"\x00"+marker, candidates)
	if partialStatus == filter.InventoryMatchAmbiguous || (match >= 0 && partial >= 0) {
		return -1, filter.InventoryMatchAmbiguous
	}
	if match >= 0 {
		return match, status
	}
	return partial, partialStatus
}

func inventoryStateForDesired(desired filter.DesiredRule, match filter.InventoryMatch) filter.InventoryState {
	if match != filter.InventoryMatchExact {
		return filter.InventoryStateDrifted
	}
	if desired.Protected {
		return filter.InventoryStateProtected
	}
	switch desired.Origin {
	case filter.RuleOriginAdopted:
		return filter.InventoryStateAdopted
	default:
		return filter.InventoryStateManaged
	}
}

func firewallSyncReasonMessage(code firewallsync.ReasonCode) string {
	switch code {
	case firewallsync.ReasonAlreadyExists:
		return "rule already exists in target backend"
	case firewallsync.ReasonOnlyExistsInTarget:
		return "rule exists only in target backend"
	case firewallsync.ReasonManagedOnlyInTarget:
		return "managed rule exists only in target backend"
	case firewallsync.ReasonUnsafeRemoval:
		return "managed runtime rule cannot be safely removed"
	case firewallsync.ReasonReadOnlyRule:
		return "read-only runtime rule is preserved but cannot be synchronized"
	default:
		return ""
	}
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

func firewallDeleteChange(current filter.ObservedRule, desired filter.DesiredRule) (filter.RuleChange, error) {
	if err := filter.GuardMutation(current); err != nil {
		return filter.RuleChange{}, err
	}
	before := current.Rule
	if before.UUID == "" && strings.HasPrefix(current.Marker, "1panel-rule:") {
		before.UUID = strings.TrimSpace(strings.TrimPrefix(current.Marker, "1panel-rule:"))
	}
	if before.UUID == "" {
		before.UUID = desired.Rule.UUID
	}
	return filter.RuleChange{Operation: filter.ChangeDelete, Before: &before, Locator: &current.Locator, UnmarkedAdopted: current.Marker == "" && desired.Origin == filter.RuleOriginAdopted}, nil
}

func loadForwardingFirewallOverview(manager forwarding.Adapter) (dto.FirewallSubsystemStatus, error) {
	ipv4, ipv4Err := loadForwardingFamilyInfo(manager, manager.Name(), constant.FirewallFamilyIPv4)
	ipv6, ipv6Err := loadForwardingFamilyInfo(manager, manager.Name(), constant.FirewallFamilyIPv6)
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
	return dto.FirewallSubsystemStatus{IsInit: ipv4.Initialized || ipv6.Initialized, IsBind: ipv4.Bound || ipv6.Bound, IPv4: ipv4, IPv6: ipv6}, errors.Join(ipv4Err, ipv6Err)
}

func loadForwardingFamilyInfo(manager forwarding.Adapter, backend, family string) (dto.FirewallBackendFamilyStatus, error) {
	initialized, bound, err := manager.FamilyStatus(family)
	available := err == nil
	if backend == constant.FirewallProviderIptables && family == constant.FirewallFamilyIPv6 {
		commands, commandErr := lifecycle.ResolveIptablesCommands()
		available = available && commandErr == nil && commands.IPv6Available()
	}
	return dto.FirewallBackendFamilyStatus{Available: available, Initialized: initialized, Bound: bound}, err
}

func (s *ForwardingService) forwardingEnabled() (bool, error) {
	if s.enabled != nil {
		return s.enabled()
	}
	status, err := settingRepo.GetValueByKey(constant.FirewallForwardingInitializedKey)
	return status == constant.StatusEnable, err
}

func (s *ForwardingService) initializeForwarding(manager forwarding.Adapter) error {
	if err := s.saveForwardingBackend(manager.Name()); err != nil {
		return err
	}
	return manager.Enable()
}

func (s *ForwardingService) saveForwardingBackend(backend string) error {
	if s.persistBackend != nil {
		return s.persistBackend(backend)
	}
	return settingRepo.UpdateOrCreate(constant.FirewallForwardingBackendKey, backend)
}

func (s *ForwardingService) persistForwardingEnabled() error {
	if s.markEnabled != nil {
		return s.markEnabled()
	}
	return settingRepo.UpdateOrCreate(constant.FirewallForwardingInitializedKey, constant.StatusEnable)
}

func recordForwardingSyncError(err error) {
	forwardingSyncStateMu.Lock()
	forwardingLastSyncErr = err
	forwardingSyncStateMu.Unlock()
}

func forwardingRulesFromModels(stored []model.ForwardingRule) []forwarding.Rule {
	rules := make([]forwarding.Rule, 0, len(stored))
	for _, rule := range stored {
		rules = append(rules, forwarding.Rule{
			Family: rule.Family, Protocol: rule.Protocol, Port: rule.Port, TargetIP: rule.TargetIP,
			TargetPort: rule.TargetPort, Interface: rule.Interface,
		})
	}
	return rules
}

func newForwardingService() *ForwardingService {
	return &ForwardingService{
		clientFactory: newForwardingAdapter,
		rules:         forwardingRuleRepo,
		markEnabled: func() error {
			return settingRepo.UpdateOrCreate(constant.FirewallForwardingInitializedKey, constant.StatusEnable)
		},
		persistBackend: func(backend string) error {
			return settingRepo.UpdateOrCreate(constant.FirewallForwardingBackendKey, backend)
		},
	}
}

func newForwardingAdapter() (forwarding.Adapter, error) {
	selected, _ := settingRepo.GetValueByKey(constant.FirewallForwardingBackendKey)
	selected = strings.TrimSpace(selected)
	if selected == "" {
		selected = constant.FirewallProviderIptables
	}
	return newForwardingAdapterFor(selected)
}

func newForwardingAdapterFor(backend string) (forwarding.Adapter, error) {
	client, err := lifecycle.NewClient(backend)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: selected forwarding backend %s: %w",
			errForwardingBackendUnavailable, backend, err,
		)
	}
	switch client.Name() {
	case constant.FirewallProviderIptables:
		return forwarding.NewIptables(client.Name()), nil
	case constant.FirewallProviderNftables:
		return forwarding.NewNftables(), nil
	default:
		return nil, errForwardingBackendUnavailable
	}
}

func ReconcileDockerPortGuard(ctx context.Context) error {
	return NewIDockerPortGuardService().Reconcile(ctx)
}

func (s *DockerPortGuardService) reconcileLocked(ctx context.Context) error {
	persistedEnabled, err := dockerPortGuardPersistedEnabled()
	if err != nil {
		return fmt.Errorf("load Docker port guard persisted status: %w", err)
	}
	backends := []string{constant.FirewallProviderIptables, constant.FirewallProviderNftables}
	if s.runtime != nil {
		backends = []string{selectedDockerFirewallBackend(constant.FirewallProviderIptables)}
	}
	initializedByBackend := make(map[string]bool, len(backends))
	initialized := false
	for _, backend := range backends {
		initialized, err = s.guardRuntime(backend).Initialized(dockerfirewall.FamilyIPv4)
		if err != nil {
			return &dockerfirewall.FamilyError{Family: dockerfirewall.FamilyIPv4, Err: fmt.Errorf("inspect initialization: %w", err)}
		}
		initializedByBackend[backend] = initialized
		if initialized {
			break
		}
	}
	if !initialized && !persistedEnabled {
		return nil
	}
	runtime, backend, err := s.runtimeForDocker(ctx)
	if err != nil {
		return err
	}
	initialized, inspected := initializedByBackend[backend]
	if !inspected {
		initialized, err = runtime.Initialized(dockerfirewall.FamilyIPv4)
		if err != nil {
			return &dockerfirewall.FamilyError{Family: dockerfirewall.FamilyIPv4, Err: fmt.Errorf("inspect initialization: %w", err)}
		}
	}
	if !initialized && !persistedEnabled {
		return nil
	}
	policies, err := s.runtimePolicies(ctx)
	if err != nil {
		return err
	}
	inventory, err := runtime.ListPolicies()
	if err != nil {
		return err
	}
	if !initialized {
		err = runtime.Initialize(policies, inventory)
	} else {
		err = runtime.ReplacePolicies(policies, inventory)
	}
	if err != nil {
		return err
	}
	return verifyDockerFirewall(runtime, policies, inventory.ReadOnly)
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
	return s.guardRuntime(backend), backend, nil
}

func (s *DockerPortGuardService) guardRuntime(backend string) dockerfirewall.Runtime {
	if s.runtimeForBackend != nil {
		return s.runtimeForBackend(backend)
	}
	if s.runtime != nil {
		return s.runtime
	}
	return newDockerFirewallRuntime(backend)
}

func newDockerFirewallRuntime(backend string) dockerfirewall.Runtime {
	if backend == constant.FirewallProviderNftables {
		return dockerfirewall.NewNftables()
	}
	return dockerfirewall.NewIptables()
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

func (s *DockerPortGuardService) runtimePolicies(ctx context.Context) ([]dockerfirewall.Policy, error) {
	stored, err := s.policies.ListManaged(ctx)
	if err != nil {
		return nil, err
	}
	policies := make([]dockerfirewall.Policy, 0, len(stored))
	for _, policy := range stored {
		sources := []string{}
		_ = json.Unmarshal([]byte(policy.Sources), &sources)
		policies = append(policies, dockerfirewall.Policy{UUID: policy.UUID, Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol, Mode: policy.Mode, Sources: sources})
	}
	return policies, nil
}

func dockerGuardReadOnlyPolicyUUID(policy dockerfirewall.ReadOnlyPolicy) string {
	nativeRules, _ := json.Marshal(policy.NativeRules)
	fingerprint := strings.Join([]string{
		policy.Policy.Family, policy.Policy.HostIP, strconv.Itoa(int(policy.Policy.HostPort)),
		policy.Policy.Protocol, policy.Action, string(nativeRules),
	}, "\x00")
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(fingerprint)).String()
}

func dockerPortGuardPersistedEnabled() (bool, error) {
	status, err := settingRepo.GetValueByKey(constant.FirewallDockerPortGuardStatusKey)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return status == constant.StatusEnable, err
}

func verifyDockerFirewall(runtime dockerfirewall.Runtime, desired []dockerfirewall.Policy, preserved []dockerfirewall.ReadOnlyPolicy) error {
	inventory, err := runtime.ListPolicies()
	if err != nil {
		return fmt.Errorf("verify synchronized Docker firewall policies: %w", err)
	}
	if !dockerFirewallPoliciesEqual(inventory.Policies, desired) {
		return fmt.Errorf("verify synchronized Docker firewall policies: target policies do not match the database")
	}
	if !readOnlyStatesEqual(inventory.ReadOnly, preserved) {
		return fmt.Errorf("verify synchronized Docker firewall policies: read-only runtime rules changed")
	}
	return nil
}

func dockerFirewallPoliciesEqual(left, right []dockerfirewall.Policy) bool {
	if len(left) != len(right) {
		return false
	}
	counts := make(map[string]int, len(left))
	for _, policy := range left {
		counts[dockerFirewallPolicyKey(policy)]++
	}
	for _, policy := range right {
		key := dockerFirewallPolicyKey(policy)
		if counts[key] == 0 {
			return false
		}
		counts[key]--
	}
	return true
}

func dockerFirewallPolicyKey(policy dockerfirewall.Policy) string {
	mode := policy.Mode
	if mode == dockerfirewall.ModeAllow && len(policy.Sources) == 0 {
		mode = dockerfirewall.ModeAll
	}
	sources := make([]string, 0, len(policy.Sources))
	for _, source := range policy.Sources {
		source = strings.TrimSpace(source)
		if prefix, err := netip.ParsePrefix(source); err == nil {
			source = prefix.Masked().String()
		} else if address, err := netip.ParseAddr(source); err == nil {
			address = address.Unmap()
			source = netip.PrefixFrom(address, address.BitLen()).String()
		}
		sources = append(sources, source)
	}
	sort.Strings(sources)
	host := policy.HostIP
	if address, err := netip.ParseAddr(host); err == nil {
		host = address.String()
	}
	return strings.Join([]string{
		policy.UUID, policy.Family, host, strconv.Itoa(int(policy.HostPort)),
		policy.Protocol, mode, strings.Join(sources, ","),
	}, "\x00")
}

func readOnlyStatesEqual(left, right []dockerfirewall.ReadOnlyPolicy) bool {
	if len(left) != len(right) {
		return false
	}
	leftRules := flattenNativeRules(left)
	rightRules := flattenNativeRules(right)
	if len(leftRules) != len(rightRules) {
		return false
	}
	for index := range leftRules {
		if leftRules[index].Family != rightRules[index].Family || !slices.Equal(leftRules[index].Tokens, rightRules[index].Tokens) {
			return false
		}
	}
	return true
}

func flattenNativeRules(policies []dockerfirewall.ReadOnlyPolicy) []dockerfirewall.NativeRule {
	rules := make([]dockerfirewall.NativeRule, 0)
	for _, policy := range policies {
		rules = append(rules, policy.NativeRules...)
	}
	slices.SortStableFunc(rules, func(left, right dockerfirewall.NativeRule) int {
		if left.Family < right.Family {
			return -1
		}
		if left.Family > right.Family {
			return 1
		}
		if left.Order < right.Order {
			return -1
		}
		if left.Order > right.Order {
			return 1
		}
		return 0
	})
	return rules
}

func newDockerPortGuardService() *DockerPortGuardService {
	return &DockerPortGuardService{
		policies: repo.NewIDockerPortGuardRepo(),
		client:   docker.NewDockerClient,
		version:  dockerFirewallVersion,
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
	if operation == firewall.BaseOperationInit || operation == firewall.BaseOperationBind {
		var err error
		ports, err = LoadRequiredFirewallPortWhiteList()
		if err != nil {
			return err
		}
	}
	if provider == constant.FirewallProviderNftables {
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

func (s *FirewallService) desiredFirewallRulesByScope(ctx context.Context, stored []model.FirewallRule, runtime filter.Adapter) (map[string][]filter.DesiredRule, []filter.InventoryItem) {
	provider := runtime.Provider()
	desired := make(map[string][]filter.DesiredRule)
	var failures []filter.InventoryItem
	ports, protectionErr := loadFirewallPortWhiteList()
	var required []firewall.PortWhitelist
	if protectionErr == nil {
		loadRequired := s.requiredPorts
		if loadRequired != nil {
			required, protectionErr = loadRequired()
		} else {
			required, protectionErr = firewall.RequiredPortWhitelist(ports)
		}
	}
	whitelist := filter.NewPortWhitelistIndex(ports)
	for _, record := range stored {
		compiled, _, err := s.compileRestorableFirewallRules(ctx, record, runtime, required)
		if err == nil {
			err = protectionErr
		}
		if err != nil {
			rule := filter.FirewallRule{
				UUID:     record.UUID,
				Scope:    filter.Scope{Provider: provider, Family: filter.Family(record.Family), Direction: filter.DirectionInput}.Normalize(),
				Protocol: record.Protocol, SourceAddress: record.SourceAddress, SourcePort: record.SourcePort,
				DestinationAddress: record.DestinationAddress, DestinationPort: record.DestinationPort,
				Interface: record.Interface, ConnectionStates: strings.FieldsFunc(record.ConnectionStates, func(r rune) bool { return r == ',' }),
				Action: filter.Action(record.Action), Description: record.Description, Priority: record.Priority,
			}
			failures = append(failures, filter.InventoryItem{
				Incompatible: isFirewallPolicyIncompatible(err),
				Rule:         rule, State: filter.InventoryStateDrifted, Match: filter.InventoryMatchNone,
				Desired: &filter.DesiredRule{UUID: record.UUID, Rule: rule, Origin: filter.RuleOrigin(record.Origin), Protected: protectionErr != nil || whitelist.Matches(rule)},
				Error:   fmt.Sprintf("policy %s: %v", record.UUID, err),
			})
			continue
		}
		for _, rule := range compiled {
			rule.Protected = whitelist.Matches(rule.Rule)
			rule.Expanded = len(compiled) > 1
			key := rule.Rule.Scope.Key()
			desired[key] = append(desired[key], rule)
		}
	}
	return desired, failures
}

func firewallInventoryPositionRanges(provider filter.Provider, items []filter.InventoryItem) (ipv4, ipv6 filter.PositionRange) {
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
		if (provider == filter.ProviderIptables || provider == filter.ProviderNftables) &&
			(scope.Table != "filter" || scope.Chain != filter.IptablesInputChain) {
			continue
		}
		bounds := &ipv4
		if scope.Family == filter.FamilyIPv6 {
			bounds = &ipv6
		} else if scope.Family != filter.FamilyIPv4 {
			continue
		}
		position := *item.Observed.Locator.Position
		if position < 1 {
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

func (s *FirewallService) adoptRule(ctx context.Context, runtime filter.Adapter, snapshot filter.RuleSet, observed filter.ObservedRule, source dto.FirewallRuleCreateItem) error {
	if (observed.Rule.Scope.Provider == filter.ProviderIptables || observed.Rule.Scope.Provider == filter.ProviderNftables) &&
		(observed.Rule.Scope.Chain == filter.BasicBeforeChain || observed.Rule.Scope.Chain == filter.BasicAfterChain) {
		return fmt.Errorf("%w: system preset chains cannot be adopted", filter.ErrUnsupportedScope)
	}
	if observed.Protected {
		return filter.ErrProtectedRule
	}
	if observed.ParseStatus != filter.ParseStatusSupported ||
		(observed.Persistence != "" && observed.Persistence != filter.PersistenceStatusConverged) {
		return fmt.Errorf("%w: rule cannot be managed", filter.ErrRuleOperation)
	}
	rule, err := prepareFirewallBackendRule(ctx, runtime, observed.Rule)
	if err != nil {
		return err
	}
	if rule.Scope.Provider != filter.ProviderFirewalld && observed.Locator.Position != nil {
		position := int64(*observed.Locator.Position)
		rule.OrderIndex = &position
	}
	record, err := firewallRuleModelForCreate(rule, source, constant.FirewallRuleOriginAdopted)
	if err != nil {
		return err
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		return err
	}
	if runtime.Provider() == filter.ProviderNftables {
		duplicates := 0
		for _, candidate := range snapshot.Rules {
			if candidate.ParseStatus != filter.ParseStatusSupported {
				continue
			}
			same, err := filter.SameRuleContent(candidate.Rule, rule)
			if err != nil {
				return err
			}
			if same {
				duplicates++
				if duplicates > 1 {
					return buserr.WithDetail("ErrInvalidParams", "duplicate firewall rules prevent adoption; manually delete duplicate rules and retry", nil)
				}
			}
		}
	}
	identities, err := firewallRuleCollisions(stored, rule.Scope.Provider)
	if err != nil {
		return err
	}
	for _, existing := range stored {
		marker := "1panel-rule:" + existing.UUID
		if existing.UUID != "" && (observed.Marker == marker || strings.HasPrefix(observed.Marker, marker+"-")) {
			return fmt.Errorf("%w: rule is already managed", filter.ErrRuleOperation)
		}
	}
	if err := identities.CheckDuplicate(rule); err != nil {
		if errors.Is(err, filter.ErrRuleOperation) {
			return buserr.WithDetail("ErrInvalidParams", "duplicate firewall rules prevent adoption; manually delete duplicate rules and retry", nil)
		}
		return err
	}
	record.UUID = uuid.NewString()
	rule.UUID = record.UUID
	if runtime.Provider() != filter.ProviderNftables {
		ports, err := loadFirewallPortWhiteList()
		if err != nil {
			return err
		}
		if filter.RuleMatchesPortWhitelist(rule, ports) {
			return filter.ErrProtectedRule
		}
		before := observed.Rule
		before.UUID = record.UUID
		scope := filter.RuleSet{Scope: rule.Scope}
		remove, err := runtime.BuildCommands(scope, []filter.RuleChange{{
			Operation: filter.ChangeDelete, Before: &before, CommandOnly: true,
			UnmarkedAdopted: observed.Marker == "", PreviousMarker: observed.Marker,
		}})
		if err != nil {
			return err
		}
		create, err := runtime.BuildCommands(scope, []filter.RuleChange{{
			Operation: filter.ChangeCreate, After: &rule, CommandOnly: true, Append: rule.OrderIndex == nil,
		}})
		if err != nil {
			return err
		}
		remove.CommandOnly, create.CommandOnly = true, true
		if err := runtime.RunCommands(ctx, remove); err != nil {
			return err
		}
		record.Priority = rule.Priority
		saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		err = s.saveFirewallRule(saveCtx, &record)
		cancel()
		if err != nil {
			return err
		}
		runErr := runtime.RunCommands(ctx, create)
		if err := errors.Join(runErr, persistFirewallRules(ctx, runtime, create)); err != nil {
			return buserr.WithDetail("ErrFirewallRuleSavedApplyFailed", err.Error(), err)
		}
		return nil
	}
	plan, verification, err := applyFirewallChanges(runtime, ctx, snapshot, []filter.RuleChange{{
		Operation: filter.ChangeAdopt, After: &rule, Locator: &observed.Locator, PreviousMarker: observed.Marker,
	}})
	if err != nil {
		return err
	}
	if !verification.Matched {
		return filter.ErrVerificationFailed
	}
	if _, err := filter.FindCommittedObserved(verification.RuleSet, rule, plan); err != nil {
		return rollbackFirewallPlan(ctx, runtime, plan, err)
	}
	return s.saveFirewallRule(ctx, &record)
}

func (index firewallRuleCollisionIndex) CheckDuplicate(rule filter.FirewallRule) error {
	key, err := filter.RuleMatchKey(rule)
	if err != nil {
		return err
	}
	for _, action := range index[key] {
		if action == rule.Action {
			return checkCollisionActions(rule.Action, action)
		}
	}
	return nil
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
		rules:                  repo.NewIFirewallRuleRepo(),
		adapters:               nil,
		forwardingSync:         newForwardingService(),
		dockerSync:             newDockerPortGuardService(),
		selectedProvider:       firewallRuleSelectedProvider,
		requiredPorts:          LoadRequiredFirewallPortWhiteList,
		cleanupBackend:         cleanupSystemBackend,
		cleanupInactiveBackend: cleanupInactiveSystemBackend,
		resetBackend:           resetServiceFirewallBackend,
		dockerActive:           firewallDockerActive,
		restoreForwarding: func(ctx context.Context) error {
			return newForwardingService().Restore(ctx)
		},
		restoreDockerGuard: ReconcileDockerPortGuard,
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

func (s *ForwardingService) loadRuleSyncCandidates(ctx context.Context, targetProvider filter.Provider) (forwarding.Adapter, []forwardingRuleSyncCandidate, []forwarding.Rule, bool, error) {
	target, err := s.clientFactory()
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
	initialized, _, err := target.InitStatus()
	if err != nil {
		return nil, nil, nil, false, err
	}
	targetRules := make([]forwarding.Rule, 0)
	if initialized {
		targetRules, err = target.List()
		if err != nil {
			return nil, nil, nil, false, err
		}
		targetRules, err = normalizeForwardingRuntimeRules(targetRules)
		if err != nil {
			return nil, nil, nil, false, err
		}
	}
	return target, candidates, targetRules, initialized, nil
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

func forwardingSyncPreview(target filter.Provider, candidates []forwardingRuleSyncCandidate, actual []forwarding.Rule) dto.FirewallRuleSyncPreview {
	desired := make([]firewallSyncDesired[forwarding.Rule], 0, len(candidates))
	for _, candidate := range candidates {
		desired = append(desired, firewallSyncDesired[forwarding.Rule]{
			Value: candidate.rule,
			Payload: dto.FirewallRuleSyncItem{
				SourceUUID: candidate.rule.Identity(), ForwardRule: &dto.ForwardRule{Family: candidate.rule.Family, Protocol: candidate.rule.Protocol, Port: candidate.rule.Port, TargetIP: candidate.rule.TargetIP, TargetPort: candidate.rule.TargetPort, Interface: candidate.rule.Interface},
			},
			Err: candidate.err,
		})
	}
	return firewallDiffPreview(
		"forwarding", target, desired, actual,
		func(rule forwarding.Rule) string { return rule.Identity() },
		func(rule forwarding.Rule) dto.FirewallRuleSyncItem {
			return dto.FirewallRuleSyncItem{SourceUUID: rule.Identity(), ForwardRule: &dto.ForwardRule{Family: rule.Family, Protocol: rule.Protocol, Port: rule.Port, TargetIP: rule.TargetIP, TargetPort: rule.TargetPort, Interface: rule.Interface}}
		},
	)
}

func firewallDiffPreview[T any](subsystem string, target filter.Provider, desired []firewallSyncDesired[T], actual []T, key func(T) string, actualItem func(T) dto.FirewallRuleSyncItem) dto.FirewallRuleSyncPreview {
	preview := dto.FirewallRuleSyncPreview{Subsystem: subsystem, TargetProvider: target, Items: make([]dto.FirewallRuleSyncItem, 0, len(desired)+len(actual))}
	actualByKey := make(map[string][]int, len(actual))
	for index, value := range actual {
		actualByKey[key(value)] = append(actualByKey[key(value)], index)
	}
	matched := make([]bool, len(actual))
	for _, candidate := range desired {
		item := candidate.Payload
		item.Status, item.ReasonCode, item.Reason = "", "", ""
		switch {
		case candidate.Err != nil:
			item.Status, item.ReasonCode, item.Reason = firewallsync.StatusBlocked, firewallsync.ReasonInvalidPolicy, candidate.Err.Error()
		default:
			match := -1
			for _, index := range actualByKey[key(candidate.Value)] {
				if !matched[index] {
					match = index
					break
				}
			}
			if match >= 0 {
				matched[match] = true
				item.Status, item.ReasonCode = firewallsync.StatusExisting, firewallsync.ReasonAlreadyExists
				item.Reason = firewallSyncReasonMessage(item.ReasonCode)
			} else {
				item.Status = firewallsync.StatusReady
			}
		}
		preview.Add(item)
	}
	for index, value := range actual {
		if matched[index] {
			continue
		}
		item := actualItem(value)
		item.Status, item.ReasonCode = firewallsync.StatusRemove, firewallsync.ReasonOnlyExistsInTarget
		item.Reason = firewallSyncReasonMessage(item.ReasonCode)
		preview.Add(item)
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

func firewallRuleStatesEqual[T any](left, right []T, key func(T) string) bool {
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

func (s *DockerPortGuardService) loadRuleSyncCandidates(ctx context.Context, request dto.FirewallRuleSyncRequest) (string, []model.DockerPortGuardPolicy, dockerfirewall.Runtime, error) {
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
	selected, _ := settingRepo.GetValueByKey(constant.FirewallDockerBackendKey)
	selected = strings.ToLower(strings.TrimSpace(selected))
	if selected == constant.FirewallProviderIptables || selected == constant.FirewallProviderNftables {
		return selected, nil
	}
	if s.client == nil {
		return "", buserr.New("ErrDockerFailed")
	}
	cli, err := s.client()
	if err != nil {
		return "", buserr.WithDetail("ErrDockerFailed", err.Error(), err)
	}
	defer cli.Close()
	info, err := cli.Info(ctx)
	if err != nil {
		return "", buserr.WithDetail("ErrDockerFailed", err.Error(), err)
	}
	return selectedDockerFirewallBackend(dockerFirewallBackend(info)), nil
}

func dockerSyncPreview(target filter.Provider, policies []model.DockerPortGuardPolicy, inventory dockerfirewall.PolicyInventory) dto.FirewallRuleSyncPreview {
	desired := make([]firewallSyncDesired[dockerfirewall.Policy], 0, len(policies))
	for _, policy := range policies {
		sources := []string{}
		_ = json.Unmarshal([]byte(policy.Sources), &sources)
		desired = append(desired, firewallSyncDesired[dockerfirewall.Policy]{
			Value: dockerfirewall.Policy{
				UUID: policy.UUID, Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort,
				Protocol: policy.Protocol, Mode: policy.Mode, Sources: sources,
			},
			Payload: dto.FirewallRuleSyncItem{SourceUUID: policy.UUID, DockerRule: &dto.DockerPortGuardEndpoint{
				Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol,
				PolicyUUID: policy.UUID, Mode: policy.Mode, Sources: sources, Description: policy.Description,
				TrafficPath: dockerTrafficPathUnknown, ManagementTarget: dockerManagementNeedsDiagnosis,
				ManagementReason: dockerReasonNoMatchingPath,
			}},
		})
	}
	preview := firewallDiffPreview(
		"docker", target, desired, inventory.Policies, dockerFirewallPolicyKey,
		func(policy dockerfirewall.Policy) dto.FirewallRuleSyncItem {
			return dto.FirewallRuleSyncItem{SourceUUID: policy.UUID, DockerRule: &dto.DockerPortGuardEndpoint{
				Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol,
				PolicyUUID: policy.UUID, Mode: policy.Mode, Sources: append([]string(nil), policy.Sources...),
				TrafficPath: dockerTrafficPathUnknown, ManagementTarget: dockerManagementNeedsDiagnosis,
				ManagementReason: dockerReasonNoMatchingPath,
			}}
		},
	)
	for _, policy := range inventory.ReadOnly {
		preview.Add(dto.FirewallRuleSyncItem{
			SourceUUID: dockerGuardReadOnlyPolicyUUID(policy),
			DockerRule: &dto.DockerPortGuardEndpoint{
				Family: policy.Policy.Family, HostIP: policy.Policy.HostIP, HostPort: policy.Policy.HostPort,
				Protocol: policy.Policy.Protocol, PolicyUUID: dockerGuardReadOnlyPolicyUUID(policy), Sources: append([]string(nil), policy.Policy.Sources...),
				NativeAction: policy.Action, ReadOnly: true, TrafficPath: dockerTrafficPathUnknown,
				ManagementTarget: dockerManagementNeedsDiagnosis, ManagementReason: dockerReasonNoMatchingPath,
			},
			Status:     firewallsync.StatusBlocked,
			ReasonCode: firewallsync.ReasonReadOnlyRule,
			Reason:     firewallSyncReasonMessage(firewallsync.ReasonReadOnlyRule),
		})
	}
	return preview
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

func annotateDockerEndpointManagement(endpoints []dto.DockerPortGuardEndpoint, backend string) {
	rules := map[string]dockerfirewall.DNATRules{
		constant.FirewallFamilyIPv4: dockerfirewall.ReadDNATRules(backend, constant.FirewallFamilyIPv4),
		constant.FirewallFamilyIPv6: dockerfirewall.ReadDNATRules(backend, constant.FirewallFamilyIPv6),
	}
	proxies := dockerfirewall.ReadProxyEndpoints()
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
			policyKey := fmt.Sprintf("%t|%s|%s|%t|%s|%s|%s", endpoint.PolicyUUID != "", endpoint.Mode, strings.Join(sources, ","), endpoint.Effective, endpoint.Description, endpoint.ManagementTarget, endpoint.ManagementReason)
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

func queueFirewallRuleTask(subsystem, operation string, labels []string, apply func(context.Context) error) (dto.FilterChainOperationResponse, error) {
	taskItem, err := task.NewTask(firewallTaskName(operation, subsystem, ""), operation, task.TaskScopeFirewall, "", 0)
	if err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	taskItem.AddSubTaskWithOps(taskItem.Name, func(t *task.Task) error {
		t.Logf("rules=%d", len(labels))
		err := t.TaskCtx.Err()
		if err == nil {
			err = apply(t.TaskCtx)
		}
		succeeded, failed := 0, 0
		for _, label := range labels {
			if err != nil {
				failed++
				t.LogFailedWithErr(label, err)
			} else {
				succeeded++
				t.LogSuccess(label)
			}
		}
		t.Log(i18n.GetMsgWithMap("FirewallRuleOperationResult", map[string]interface{}{
			"succeeded": succeeded, "failed": failed,
		}))
		return err
	}, nil, 0, 0)
	if err := repo.NewITaskRepo().Save(context.Background(), taskItem.Task); err != nil {
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
		(policy.Mode != dockerfirewall.ModeAll && policy.Mode != dockerfirewall.ModeSources && policy.Mode != dockerfirewall.ModeAllow) {
		return dockerfirewall.Policy{}, buserr.WithDetail("ErrInvalidParams", "invalid policy fields", nil)
	}
	address, err := netip.ParseAddr(policy.HostIP)
	if err != nil || (policy.Family == dockerfirewall.FamilyIPv4) != address.Is4() {
		return dockerfirewall.Policy{}, buserr.WithDetail("ErrInvalidParams", "host IP does not match address family", nil)
	}
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
	if policy.Mode != dockerfirewall.ModeAll && len(normalizedSources) == 0 {
		return dockerfirewall.Policy{}, buserr.WithDetail("ErrInvalidParams", "source-based modes require at least one source", nil)
	}
	if policy.Mode == dockerfirewall.ModeAll {
		normalizedSources = []string{}
	}
	sort.Strings(normalizedSources)
	policy.Sources = normalizedSources
	return policy, nil
}

func (i forwardingInventoryItem) SyncStatus() string {
	switch {
	case i.IsDesired && i.IsRuntime:
		return forwardingSyncConverged
	case i.IsDesired:
		return forwardingSyncMissing
	default:
		return forwardingSyncRuntimeOnly
	}
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
