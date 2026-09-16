package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	filterruntime "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/runtime"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/nftables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/ping"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IFirewallSettingService interface {
	CreatePortWhitelist(context.Context, dto.FirewallPortWhitelistCreate) (dto.FilterChainOperationResponse, error)
	UpdatePortWhitelist(context.Context, dto.FirewallPortWhitelistUpdate) (dto.FilterChainOperationResponse, error)
	DeletePortWhitelist(context.Context, dto.FirewallPortWhitelistDelete) (dto.FilterChainOperationResponse, error)
	Load(context.Context) (dto.FirewallSettings, error)
	Operate(context.Context, dto.FirewallBackendOperation) error
}

type FirewallSettingService struct{}

var firewallWhitelistTaskMu sync.Mutex

var ErrFirewallBackendCleanupRequired = errors.New("firewall backend cleanup required")

func firewallBackendCleanupRequired(current, target string) error {
	return fmt.Errorf(
		"%w: current backend %s still contains 1Panel runtime rules; clean it up before switching to %s",
		ErrFirewallBackendCleanupRequired,
		current,
		target,
	)
}

func NewIFirewallSettingService() IFirewallSettingService {
	return &FirewallSettingService{}
}

type portWhitelistChange func([]firewall.PortWhitelist) ([]firewall.PortWhitelist, error)

type portWhitelistPlan struct {
	current []firewall.PortWhitelist
	desired []firewall.PortWhitelist
}

type firewallWhitelistOverrideKey struct{}

func (s *FirewallSettingService) CreatePortWhitelist(ctx context.Context, request dto.FirewallPortWhitelistCreate) (dto.FilterChainOperationResponse, error) {
	return s.queuePortWhitelist(ctx, func(current []firewall.PortWhitelist) ([]firewall.PortWhitelist, error) {
		rule, err := initializeRequestedWhitelistRule(request.Rule)
		if err != nil {
			return nil, err
		}
		return append(slices.Clone(current), rule), nil
	})
}

func (s *FirewallSettingService) UpdatePortWhitelist(ctx context.Context, request dto.FirewallPortWhitelistUpdate) (dto.FilterChainOperationResponse, error) {
	return s.queuePortWhitelist(ctx, func(current []firewall.PortWhitelist) ([]firewall.PortWhitelist, error) {
		rule, err := initializeRequestedWhitelistRule(request.Rule)
		if err != nil {
			return nil, err
		}
		return replacePortWhitelistRule(current, request.OldRule, rule)
	})
}

func (s *FirewallSettingService) DeletePortWhitelist(ctx context.Context, request dto.FirewallPortWhitelistDelete) (dto.FilterChainOperationResponse, error) {
	if request.Rule == nil {
		return dto.FilterChainOperationResponse{}, fmt.Errorf("select one firewall port whitelist rule to delete")
	}
	return s.queuePortWhitelist(ctx, func(current []firewall.PortWhitelist) ([]firewall.PortWhitelist, error) {
		return removePortWhitelistRule(current, *request.Rule)
	})
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

func replacePortWhitelistRule(current []firewall.PortWhitelist, oldRule, rule firewall.PortWhitelist) ([]firewall.PortWhitelist, error) {
	index, err := findPortWhitelistRule(current, oldRule)
	if err != nil {
		return nil, err
	}
	rules := slices.Clone(current)
	rules[index] = rule
	return rules, nil
}

func removePortWhitelistRule(current []firewall.PortWhitelist, target firewall.PortWhitelist) ([]firewall.PortWhitelist, error) {
	index, err := findPortWhitelistRule(current, target)
	if err != nil {
		return nil, err
	}
	return slices.Delete(slices.Clone(current), index, index+1), nil
}

func (s *FirewallSettingService) queuePortWhitelist(ctx context.Context, change portWhitelistChange) (dto.FilterChainOperationResponse, error) {
	firewallService := newFirewallService()
	firewallWhitelistTaskMu.Lock()
	defer firewallWhitelistTaskMu.Unlock()
	if err := task.CheckScopeTaskIsExecuting(task.TaskScopeFirewall, 0); err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	plan, err := s.preparePortWhitelist(ctx, change)
	if err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	taskItem, err := task.NewTask(i18n.GetMsgByKey("FirewallWhitelistTask"), task.TaskUpdate, task.TaskScopeFirewall, "", 0)
	if err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	taskItem.AddSubTaskWithOps(taskItem.Name, func(t *task.Task) error {
		firewallWhitelistTaskMu.Lock()
		defer firewallWhitelistTaskMu.Unlock()
		succeeded, failed := 0, 0
		err := s.applyPortWhitelist(t.TaskCtx, plan, firewallService, true, func(status, label string, err error) {
			switch status {
			case "applied":
				succeeded++
				t.LogSuccess(label)
			case "failed":
				failed++
				t.LogFailedWithErr(label, err)
			default:
				t.Log(i18n.GetWithName(status, label))
			}
		})
		t.Log(i18n.GetMsgWithMap("FirewallRuleOperationResult", map[string]interface{}{
			"succeeded": succeeded, "failed": failed,
		}))
		return err
	}, nil, 0, 0)
	if err := repo.NewITaskRepo().Save(context.Background(), taskItem.Task); err != nil {
		closeUnstartedFirewallTask(taskItem)
		return dto.FilterChainOperationResponse{}, fmt.Errorf("save firewall whitelist task: %w", err)
	}
	go func() { _ = taskItem.Execute() }()
	return dto.FilterChainOperationResponse{TaskID: taskItem.TaskID, Queued: true}, nil
}

type whitelistReporter func(status, label string, err error)

func (s *FirewallSettingService) applyPortWhitelist(ctx context.Context, plan portWhitelistPlan, firewallService *FirewallService, requireActive bool, report whitelistReporter) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	ports := plan.desired
	required, err := firewall.RequiredPortWhitelist(ports)
	if err != nil {
		report("failed", "SSH / 1Panel", err)
		return err
	}
	ctx = context.WithValue(ctx, firewallWhitelistOverrideKey{}, plan.desired)
	provider, providerErr := firewallService.selectedProvider(ctx)
	if providerErr != nil && !requireActive && configuredSystemFirewallBackend() == "" && len(lifecycle.InstalledProviders()) == 0 {
		return s.executePortWhitelist(ctx, plan, func() error { return nil }, report)
	}
	rulesFor := func(entries []firewall.PortWhitelist) ([]filter.FirewallRule, error) {
		valid := make([]firewall.PortWhitelist, 0, len(entries))
		for _, entry := range entries {
			if normalized, err := firewall.ValidatePortWhitelist([]firewall.PortWhitelist{entry}); err == nil {
				valid = append(valid, normalized...)
			}
		}
		system, err := firewall.RequiredPortWhitelist(valid)
		if err != nil {
			return nil, err
		}
		return whitelistRules(provider, firewall.ExpandPortWhitelist(customWhitelist(valid)), firewall.ExpandPortWhitelist(system)), nil
	}
	committed, err := rulesFor(plan.current)
	if err != nil {
		return err
	}
	desired := whitelistRules(provider, firewall.ExpandPortWhitelist(customWhitelist(ports)), firewall.ExpandPortWhitelist(required))
	removals, additions := whitelistRuleChanges(committed, desired)
	ready := s.portWhitelistReadiness(provider, providerErr, firewallService)
	current, err := firewallService.prepareWhitelistRules(ctx, provider, additions, ready, report)
	if err != nil {
		return err
	}
	if err := checkPortWhitelist(ctx, firewallService, provider, current, report); err != nil {
		return err
	}
	previous, err := firewallService.prepareWhitelistRules(ctx, provider, removals, ready, report)
	if err != nil {
		return err
	}
	return s.executePortWhitelist(ctx, plan, func() error {
		if err := syncWhitelistRules(ctx, firewallService, previous, current, report); err != nil {
			return err
		}
		return firewallService.reconcilePortWhitelist(ctx, provider, ready, report)
	}, report)
}

func whitelistRules(provider filter.Provider, ports, required []firewall.SystemPort) []filter.FirewallRule {
	rules := make([]filter.FirewallRule, 0, len(ports)+len(required))
	for _, port := range required {
		rule := systemPortRule(provider, port)
		if isDirectFirewallProvider(provider) {
			rule.Scope.Chain = filter.BasicBeforeChain
		}
		rules = append(rules, rule)
	}
	for _, port := range ports {
		rules = append(rules, systemPortRule(provider, port))
	}
	return rules
}

func whitelistRuleChanges(committed, desired []filter.FirewallRule) (removals, additions []filter.FirewallRule) {
	for _, rule := range desired {
		if !whitelistContainsRule(committed, rule) {
			additions = append(additions, rule)
		}
	}
	for _, rule := range committed {
		if !whitelistContainsRule(desired, rule) && !whitelistContainsRule(removals, rule) {
			removals = append(removals, rule)
		}
	}
	return removals, additions
}

func (s *FirewallService) prepareWhitelistRules(
	ctx context.Context, provider filter.Provider, candidates []filter.FirewallRule,
	ready func(dto.FirewallSystemPort) (bool, error), report whitelistReporter,
) ([]preparedFirewallRuleCreate, error) {
	rules := make([]preparedFirewallRuleCreate, 0, len(candidates))
	seen := make(filter.RuleCollisionIndex)
	var failures []error
	for _, rule := range candidates {
		port := firewall.SystemPort{Family: string(rule.Scope.Family), Port: rule.DestinationPort, Protocol: rule.Protocol, SourceAddress: rule.SourceAddress}
		prepare := func() error {
			if err := ctx.Err(); err != nil {
				return err
			}
			active, err := ready(port)
			if err != nil {
				return err
			}
			if !active {
				report("FirewallWhitelistDeferred", whitelistPortLabel(port), nil)
				return nil
			}
			prepared, err := s.prepareCreate(ctx, provider, dto.FirewallRuleCreateItem{
				Rule: rule, SourceKind: constant.FirewallRuleSourceSecurity, SourceID: systemPortSourceID(port),
			})
			if err != nil {
				return err
			}
			if err := seen.CheckDuplicate(prepared.request.Rule); errors.Is(err, filter.ErrRuleOperation) {
				return nil
			} else if err != nil {
				return err
			}
			if err := seen.Add(prepared.request.Rule); err != nil {
				return err
			}
			rules = append(rules, prepared)
			return nil
		}
		if err := prepare(); err != nil {
			label := whitelistPortLabel(port)
			report("failed", label, err)
			failures = append(failures, fmt.Errorf("%s: %w", label, err))
		}
	}
	return rules, errors.Join(failures...)
}

func syncWhitelistRules(ctx context.Context, service *FirewallService,
	previous, current []preparedFirewallRuleCreate, report whitelistReporter,
) error {
	apply := func(rules []preparedFirewallRuleCreate, operation string, run func(context.Context, preparedFirewallRuleCreate) error) error {
		var failures []error
		for _, prepared := range rules {
			if err := ctx.Err(); err != nil {
				return err
			}
			rule := prepared.request.Rule
			port := firewall.SystemPort{Family: string(rule.Scope.Family), Port: rule.DestinationPort, Protocol: rule.Protocol, SourceAddress: rule.SourceAddress}
			label := whitelistPortLabel(port)
			if err := run(ctx, prepared); err != nil {
				report("failed", label, err)
				failures = append(failures, fmt.Errorf("%s: %w", label, err))
			} else {
				report("applied", i18n.GetMsgByKey(operation)+": "+label, nil)
			}
		}
		return errors.Join(failures...)
	}
	if err := apply(current, task.TaskCreate, service.addWhitelistRule); err != nil {
		return err
	}
	removals := make([]preparedFirewallRuleCreate, 0, len(previous))
	for _, prepared := range previous {
		if !slices.ContainsFunc(current, func(candidate preparedFirewallRuleCreate) bool {
			same, err := filter.SameRuleContent(candidate.request.Rule, prepared.request.Rule)
			return err == nil && same
		}) {
			removals = append(removals, prepared)
		}
	}
	return apply(removals, task.TaskDelete, service.deleteWhitelistRule)
}

func whitelistContainsRule(rules []filter.FirewallRule, target filter.FirewallRule) bool {
	for _, rule := range rules {
		if same, err := filter.SameRuleContent(rule, target); err == nil && same {
			return true
		}
	}
	return false
}

func (s *FirewallService) addWhitelistRule(ctx context.Context, prepared preparedFirewallRuleCreate) error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	rule, runtime := prepared.request.Rule, prepared.runtime
	snapshot, err := runtime.ObserveMutation(ctx, rule.Scope)
	if err != nil {
		return err
	}
	for _, observed := range snapshot.Rules {
		if observed.ParseStatus == filter.ParseStatusSupported {
			if same, err := filter.SameRuleContent(rule, observed.Rule); err == nil && same {
				return nil
			}
		}
	}
	if err := filter.CheckObservedRuleCollisions(snapshot, rule, nil); err != nil {
		return err
	}
	if rule.Scope.Chain == filter.BasicBeforeChain {
		rule.UUID = uuid.NewString()
		return runtime.ExecuteCreate(ctx, snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		return err
	}
	for _, record := range stored {
		compiled, err := s.compileStoredFirewallRules(ctx, record, rule.Scope.Provider)
		if err != nil {
			continue
		}
		for _, candidate := range compiled {
			if err := filter.CheckRuleCollision(rule, candidate.Rule); errors.Is(err, filter.ErrRuleOperation) {
				return runtime.ExecuteCreate(ctx, snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &candidate.Rule}})
			} else if err != nil {
				return err
			}
		}
	}
	return s.applyCreateRules(ctx, runtime, snapshot, stored, []preparedFirewallRuleCreate{prepared})[0]
}

func (s *FirewallService) deleteWhitelistRule(ctx context.Context, prepared preparedFirewallRuleCreate) error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	rule, runtime := prepared.request.Rule, prepared.runtime
	provider := rule.Scope.Provider
	ports, err := firewallWhitelistForProtection(ctx)
	if err != nil {
		return err
	}
	if rule.Scope.Chain != filter.BasicBeforeChain && filter.RuleMatchesPortWhitelist(rule, ports) {
		return nil
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		return err
	}
	cleanup, _ := ctx.Value(firewallWhitelistCleanupKey{}).(bool)
	if !cleanup || rule.Scope.Chain == filter.BasicBeforeChain {
		if err := deleteNativeWhitelistRule(ctx, runtime, rule); err != nil {
			return err
		}
	}
	for _, record := range stored {
		if cleanup && !hasSystemFirewallRuleOwner(record) {
			continue
		}
		compiled, err := s.compileStoredFirewallRules(ctx, record, provider)
		if err != nil {
			continue
		}
		remaining := make([]filter.FirewallRule, 0, len(compiled))
		for _, candidate := range compiled {
			if !matchesWhitelistRemoval(rule, candidate.Rule, record) || filter.RuleMatchesPortWhitelist(candidate.Rule, ports) {
				remaining = append(remaining, candidate.Rule)
				continue
			}
			if candidate.Rule.Scope.Key() != rule.Scope.Key() || cleanup && rule.Scope.Chain != filter.BasicBeforeChain {
				if err := deleteNativeWhitelistRule(ctx, runtime, candidate.Rule, candidate.Marker); err != nil {
					return err
				}
			}
		}
		if len(remaining) == len(compiled) {
			continue
		}
		if err := global.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			ctx := context.WithValue(ctx, constant.DB, tx)
			if err := s.rules.DeleteWithRevision(ctx, record.UUID, record.Revision); err != nil {
				return err
			}
			for _, rule := range remaining {
				kept, err := model.FirewallRuleFromDomain(rule)
				if err != nil {
					return err
				}
				kept.UUID, kept.Origin, kept.Owner, kept.Sequence = rule.UUID, record.Origin, record.Owner, record.Sequence
				if err := s.rules.Create(ctx, &kept); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func deleteNativeWhitelistRule(ctx context.Context, runtime *filterruntime.Engine, rule filter.FirewallRule, markers ...string) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		snapshot, err := runtime.ObserveMutation(ctx, rule.Scope)
		if err != nil {
			return err
		}
		matched := -1
		for index, observed := range snapshot.Rules {
			if len(markers) > 0 && observed.Marker != markers[0] {
				continue
			}
			if observed.ParseStatus == filter.ParseStatusSupported {
				if same, err := filter.SameRuleContent(rule, observed.Rule); err == nil && same {
					matched = index
					break
				}
			}
		}
		if matched < 0 {
			break
		}
		observed := snapshot.Rules[matched]
		before := observed.Rule
		before.UUID = strings.TrimPrefix(observed.Marker, "1panel-rule:")
		if before.UUID == "" {
			before.UUID = uuid.NewString()
		}
		snapshot.Rules[matched].Protected = false
		_, verification, err := runtime.Execute(ctx, snapshot, []filter.DesiredChange{{
			Operation: filter.ChangeDelete, Before: &before, Locator: &observed.Locator, UnmarkedAdopted: observed.Marker == "",
		}})
		if err != nil {
			return err
		}
		if !verification.Matched {
			return filter.ErrVerificationFailed
		}
	}
	return nil
}

type firewallWhitelistCleanupKey struct{}

func checkFirewallRuleWhitelistProtection(ctx context.Context, record model.FirewallRule) error {
	ports, err := firewallWhitelistForProtection(ctx)
	if err != nil {
		return err
	}
	rules, err := record.RulesForProvider(filter.ProviderIptables)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if filter.RuleMatchesPortWhitelist(rule, ports) {
			return filter.ErrProtectedRule
		}
	}
	return nil
}

func matchesWhitelistRemoval(requested, candidate filter.FirewallRule, record model.FirewallRule) bool {
	if isDirectFirewallProvider(requested.Scope.Provider) && requested.Scope.Chain == filter.BasicBeforeChain && hasSystemFirewallRuleOwner(record) {
		candidate.Scope.Chain = filter.BasicBeforeChain
	}
	return whitelistContainsRule([]filter.FirewallRule{requested}, candidate)
}

func isWhitelistPortAllowance(rule filter.FirewallRule) bool {
	normalized, err := filter.NormalizeRule(rule)
	return err == nil && normalized.Action == filter.ActionAccept &&
		(normalized.Protocol == "tcp" || normalized.Protocol == "udp") && normalized.DestinationPort != "" &&
		normalized.SourcePort == "" && normalized.DestinationAddress == "" && normalized.Interface == "" && len(normalized.ConnectionStates) == 0
}

func (s *FirewallService) reconcilePortWhitelist(ctx context.Context, provider filter.Provider,
	ready func(dto.FirewallSystemPort) (bool, error), report whitelistReporter,
) error {
	if !isDirectFirewallProvider(provider) {
		return nil
	}
	ports, err := firewallWhitelistForProtection(ctx)
	if err != nil {
		return err
	}
	runtime, err := s.adapters.Resolve(provider)
	if err != nil {
		return err
	}
	var obsolete []filter.FirewallRule
	for _, scope := range filter.ManagedInputScopes(provider) {
		if scope.Chain != filter.BasicBeforeChain {
			continue
		}
		active, err := ready(dto.FirewallSystemPort{Family: string(scope.Family)})
		if err != nil {
			return err
		}
		if !active {
			continue
		}
		snapshot, err := runtime.ObserveMutation(ctx, scope)
		if errors.Is(err, filter.ErrFamilyUnavailable) {
			continue
		}
		if err != nil {
			return err
		}
		for _, observed := range snapshot.Rules {
			if observed.ParseStatus == filter.ParseStatusSupported && isWhitelistPortAllowance(observed.Rule) &&
				!filter.RuleMatchesPortWhitelist(observed.Rule, ports) {
				obsolete = append(obsolete, observed.Rule)
			}
		}
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		return err
	}
	for _, record := range stored {
		if !hasSystemFirewallRuleOwner(record) {
			continue
		}
		compiled, err := s.compileStoredFirewallRules(ctx, record, provider)
		if err != nil {
			return err
		}
		for _, candidate := range compiled {
			if isWhitelistPortAllowance(candidate.Rule) && !filter.RuleMatchesPortWhitelist(candidate.Rule, ports) {
				obsolete = append(obsolete, candidate.Rule)
			}
		}
	}
	prepared, err := s.prepareWhitelistRules(ctx, provider, obsolete, ready, report)
	if err != nil {
		return err
	}
	return syncWhitelistRules(context.WithValue(ctx, firewallWhitelistCleanupKey{}, true), s, prepared, nil, report)
}

func (s *FirewallSettingService) preparePortWhitelist(ctx context.Context, change portWhitelistChange) (portWhitelistPlan, error) {
	var plan portWhitelistPlan
	var err error
	plan.current, err = loadPortWhitelistSetting(global.DB.WithContext(ctx))
	if err != nil {
		return plan, err
	}
	plan.desired, err = change(plan.current)
	if err != nil {
		return plan, err
	}
	plan.desired, err = firewall.ValidatePortWhitelist(plan.desired)
	return plan, err
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

func writePortWhitelistSetting(db *gorm.DB, rules []firewall.PortWhitelist) error {
	value, err := json.Marshal(rules)
	if err != nil {
		return err
	}
	return db.Where("key = ?", constant.FirewallPortWhiteList).Assign(map[string]interface{}{"value": string(value)}).
		FirstOrCreate(&model.Setting{Key: constant.FirewallPortWhiteList}).Error
}

func checkPortWhitelistPlan(db *gorm.DB, plan portWhitelistPlan) error {
	current, err := loadPortWhitelistSetting(db)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(current, plan.current) {
		return fmt.Errorf("firewall port whitelist changed while the task was running; refresh and retry")
	}
	return nil
}

func (s *FirewallSettingService) executePortWhitelist(ctx context.Context, plan portWhitelistPlan, apply func() error, report whitelistReporter) error {
	if err := checkPortWhitelistPlan(global.DB.WithContext(ctx), plan); err != nil {
		return err
	}
	if err := apply(); err != nil {
		return err
	}
	if err := global.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := checkPortWhitelistPlan(tx, plan); err != nil {
			return err
		}
		return writePortWhitelistSetting(tx, plan.desired)
	}); err != nil {
		return err
	}
	report("FirewallWhitelistSaved", "", nil)
	return nil
}

func checkPortWhitelist(
	ctx context.Context, service *FirewallService, provider filter.Provider,
	rules []preparedFirewallRuleCreate, report whitelistReporter,
) error {
	stored, err := service.rules.List(ctx)
	if err != nil {
		return err
	}
	var existing []filter.FirewallRule
	for _, record := range stored {
		if rules, err := record.RulesForProvider(provider); err == nil {
			existing = append(existing, rules...)
		}
	}
	snapshots := make(map[string]filter.Snapshot)
	var failures []error
	for _, prepared := range rules {
		rule := prepared.request.Rule
		check := func() error {
			if err := ctx.Err(); err != nil {
				return err
			}
			scope := rule.Scope.Key()
			snapshot, found := snapshots[scope]
			if !found {
				var err error
				snapshot, err = prepared.runtime.ObserveMutation(ctx, rule.Scope)
				if err != nil {
					return err
				}
				snapshots[scope] = snapshot
			}
			for _, candidate := range existing {
				if err := whitelistRuleConflict(rule, candidate); err != nil {
					return err
				}
			}
			for _, observed := range snapshot.Rules {
				if observed.ParseStatus != filter.ParseStatusSupported {
					continue
				}
				if err := whitelistRuleConflict(rule, observed.Rule); err != nil {
					return err
				}
			}
			return nil
		}
		if err := check(); err != nil {
			port := firewall.SystemPort{Family: string(rule.Scope.Family), Port: rule.DestinationPort, Protocol: rule.Protocol, SourceAddress: rule.SourceAddress}
			label := whitelistPortLabel(port)
			report("failed", label, err)
			failures = append(failures, fmt.Errorf("%s: %w", label, err))
		}
	}
	return errors.Join(failures...)
}

func whitelistRuleConflict(requested, existing filter.FirewallRule) error {
	if filter.OppositeActions(requested.Action, existing.Action) && filter.RulesOverlap(requested, existing) {
		return fmt.Errorf("%w: %s %s/%s [%s]", filter.ErrRuleConflict,
			existing.Action, existing.DestinationPort, existing.Protocol, existing.SourceAddress)
	}
	err := filter.CheckRuleCollision(requested, existing)
	if errors.Is(err, filter.ErrRuleOperation) {
		return nil
	}
	return err
}

func (s *FirewallSettingService) portWhitelistReadiness(provider filter.Provider, providerErr error, firewallService *FirewallService) func(dto.FirewallSystemPort) (bool, error) {
	type state struct {
		ready bool
		err   error
	}
	states := make(map[string]state)
	return func(port dto.FirewallSystemPort) (bool, error) {
		if providerErr != nil {
			return false, providerErr
		}
		key := "service"
		if isDirectFirewallProvider(provider) {
			key = port.Family
		}
		if cached, ok := states[key]; ok {
			return cached.ready, cached.err
		}
		var result state
		if isDirectFirewallProvider(provider) {
			initialized, bound, err := loadSystemFirewallFamilyStatus(string(provider), port.Family)
			result = state{ready: initialized && bound, err: err}
		} else {
			client, err := firewallService.baseClient()
			result.err = err
			if err == nil {
				result.ready, result.err = client.Status()
			}
		}
		states[key] = result
		return result.ready, result.err
	}
}

func whitelistPortLabel(port dto.FirewallSystemPort) string {
	return fmt.Sprintf("%s %s/%s [%s]", port.Family, port.Port, port.Protocol, port.SourceAddress)
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

func customWhitelist(entries []firewall.PortWhitelist) []firewall.PortWhitelist {
	result := make([]firewall.PortWhitelist, 0, len(entries))
	for _, entry := range entries {
		if entry.Type == "" {
			result = append(result, entry)
		}
	}
	return result
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

func initializeRequestedWhitelistRule(rule firewall.PortWhitelist) (firewall.PortWhitelist, error) {
	rule.Type = strings.ToLower(strings.TrimSpace(rule.Type))
	rules, err := InitializeFirewallWhitelistPorts([]firewall.PortWhitelist{rule})
	if err != nil {
		return rule, err
	}
	return rules[0], nil
}

func (service *FirewallService) syncSystemAccessPortTransition(ctx context.Context, serviceType string, ports []string) error {
	firewallWhitelistTaskMu.Lock()
	defer firewallWhitelistTaskMu.Unlock()
	settings := &FirewallSettingService{}
	plan, err := settings.preparePortWhitelist(ctx, func(entries []firewall.PortWhitelist) ([]firewall.PortWhitelist, error) {
		entries = slices.Clone(entries)
		for i := range entries {
			if entries[i].Type == serviceType {
				if len(ports) == 0 {
					return nil, fmt.Errorf("firewall whitelist %s requires a port", serviceType)
				}
				entries[i].Port = ports[0]
			}
		}
		return entries, nil
	})
	if err != nil {
		return err
	}
	if reflect.DeepEqual(plan.current, plan.desired) {
		return nil
	}
	return settings.applyPortWhitelist(ctx, plan, service, false, func(status, label string, err error) {
		if err != nil && global.LOG != nil {
			global.LOG.Errorf("sync service firewall rule %s: %v", label, err)
		}
	})
}

func (s *FirewallSettingService) Load(ctx context.Context) (dto.FirewallSettings, error) {
	result := dto.FirewallSettings{PingStatus: ping.LoadStatus()}

	installed := make(map[string]bool)
	for _, name := range lifecycle.InstalledProviders() {
		installed[name] = true
	}
	result.System.Selected = configuredSystemFirewallBackend()
	if result.System.Selected == "" {
		if client, err := lifecycle.NewClient(); err == nil {
			result.System.Selected = client.Name()
		}
	}
	result.System.Current = result.System.Selected
	for _, name := range []string{
		constant.FirewallProviderFirewalld,
		constant.FirewallProviderUFW,
		constant.FirewallProviderIptables,
		constant.FirewallProviderNftables,
	} {
		option := dto.FirewallBackendOption{Name: name, Installed: installed[name], Supported: true}
		if option.Installed && name == result.System.Selected {
			client, err := lifecycle.NewClientFor(name)
			if err != nil {
				option.Message = err.Error()
			} else if supportsManagedFilterChains(name) {
				option.Initialized, option.Bound, err = loadFirewallInitStatus(name, "base")
				if err != nil {
					option.Message = err.Error()
				}
				option.IPv4 = loadSystemFirewallFamilyInfo(name, constant.FirewallFamilyIPv4)
				option.IPv6 = loadSystemFirewallFamilyInfo(name, constant.FirewallFamilyIPv6)
			} else if option.Active, err = client.Status(); err != nil {
				option.Message = err.Error()
			}
		}
		if name == result.System.Selected && name == constant.FirewallProviderIptables {
			if commands, err := lifecycle.ResolveIptablesCommands(); err == nil {
				option.Implementation = commands.IPv4
			}
		}
		result.System.Options = append(result.System.Options, option)
	}

	result.Forwarding.Selected = configuredForwardingBackend()
	result.Forwarding.Current = result.Forwarding.Selected
	for _, name := range []string{constant.FirewallProviderIptables, constant.FirewallProviderNftables} {
		option := dto.FirewallBackendOption{Name: name, Installed: installed[name], Supported: true}
		if option.Installed && name == result.Forwarding.Selected {
			manager, err := newForwardingManagerFor(name)
			if err != nil {
				option.Message = err.Error()
			} else if status, err := manager.Status(); err != nil {
				option.Message = err.Error()
			} else {
				option.Initialized, option.Bound = status.IsInit, status.IsBind
				ipv4Init, ipv4Bound, ipv4Err := manager.FamilyStatus(constant.FirewallFamilyIPv4)
				ipv6Init, ipv6Bound, ipv6Err := manager.FamilyStatus(constant.FirewallFamilyIPv6)
				option.IPv4 = dto.FirewallBackendFamilyStatus{
					Available: ipv4Err == nil, Initialized: ipv4Init, Bound: ipv4Bound,
				}
				option.IPv6 = dto.FirewallBackendFamilyStatus{
					Available: ipv6Err == nil, Initialized: ipv6Init, Bound: ipv6Bound,
				}
				if name == constant.FirewallProviderIptables {
					if commands, commandErr := lifecycle.ResolveIptablesCommands(); commandErr == nil {
						option.IPv6.Available = option.IPv6.Available && commands.IPv6Available()
						if !commands.IPv6Available() {
							option.IPv6.Reason = docker_guard.ReasonCommandMissing
						}
					}
				}
			}
		}
		if name == result.Forwarding.Selected && name == constant.FirewallProviderIptables {
			if commands, err := lifecycle.ResolveIptablesCommands(); err == nil {
				option.Implementation = commands.IPv4
			}
		}
		result.Forwarding.Options = append(result.Forwarding.Options, option)
	}

	dockerInstalled := cmd.Which("docker")
	dockerVersion := ""
	if dockerInstalled {
		dockerVersion = loadDockerEngineVersion(ctx)
	}
	result.Docker.Selected = configuredDockerFirewallBackend()
	result.Docker.Current = result.Docker.Selected
	for _, name := range []string{constant.FirewallProviderIptables, constant.FirewallProviderNftables} {
		option := dto.FirewallBackendOption{
			Name: name, Installed: installed[name], Supported: dockerInstalled,
			Active: dockerInstalled && installed[name] && result.Docker.Selected == name,
		}
		if name == constant.FirewallProviderNftables && dockerInstalled && !dockerNftablesSupported(dockerVersion) {
			option.Supported = false
			option.SupportReason = "docker_version_unsupported"
			option.Active = false
		}
		if option.Active {
			guard := docker_guard.NewRuntime(name)
			ipv4, ipv6 := guard.Status(docker_guard.FamilyIPv4), guard.Status(docker_guard.FamilyIPv6)
			option.Initialized = ipv4.Initialized || ipv6.Initialized
			option.Bound = ipv4.Bound || ipv6.Bound
			option.IPv4.Initialized, option.IPv4.Bound = ipv4.Initialized, ipv4.Bound
			option.IPv6.Initialized, option.IPv6.Bound = ipv6.Initialized, ipv6.Bound
			option.IPv4.Available = ipv4.Reason != docker_guard.ReasonCommandMissing
			option.IPv6.Available = ipv6.Reason != docker_guard.ReasonCommandMissing
			option.IPv4.Reason, option.IPv6.Reason = ipv4.Reason, ipv6.Reason
		}
		result.Docker.Options = append(result.Docker.Options, option)
	}
	var err error
	result.PortWhitelist, err = loadPortWhitelistSetting(global.DB.WithContext(ctx))
	if err != nil {
		return result, err
	}
	result.PanelPort = LoadPanelPort()
	sshPort, sshErr := loadSSHWhitelistPortFrom(sshPath)
	if sshErr != nil {
		global.LOG.Warnf("load SSH port for firewall settings: %v", sshErr)
	} else {
		result.SSHPort = sshPort
	}
	return result, err
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

func loadSystemFirewallFamilyInfo(provider, family string) dto.FirewallBackendFamilyStatus {
	if provider == constant.FirewallProviderIptables && family == constant.FirewallFamilyIPv6 {
		commands, err := lifecycle.ResolveIptablesCommands()
		if err != nil || !commands.IPv6Available() {
			return dto.FirewallBackendFamilyStatus{Reason: docker_guard.ReasonCommandMissing}
		}
	}
	initialized, bound, err := loadSystemFirewallFamilyStatus(provider, family)
	return dto.FirewallBackendFamilyStatus{
		Available:   err == nil,
		Initialized: initialized,
		Bound:       bound,
	}
}

func (s *FirewallSettingService) Operate(ctx context.Context, request dto.FirewallBackendOperation) error {
	if err := lockFirewallLifecycleIdle(); err != nil {
		return err
	}
	defer firewallLifecycleTaskMu.Unlock()
	if request.Subsystem != "system" && request.Backend != constant.FirewallProviderIptables && request.Backend != constant.FirewallProviderNftables {
		return fmt.Errorf("%s only supports iptables or nftables", request.Subsystem)
	}
	if request.Subsystem == "system" && !supportsManagedFilterChains(request.Backend) && request.Operation != "select" {
		return fmt.Errorf("%s does not support initialization or cleanup", request.Backend)
	}
	switch request.Subsystem {
	case "system":
		if err := s.operateSystem(request); err != nil {
			return err
		}
		if request.Operation == "initialize" {
			service := newFirewallService()
			rulesErr := service.restoreStoredFirewallRules(ctx, filter.Provider(request.Backend), nil)
			whitelistErr := service.SyncPortWhitelist(ctx)
			return errors.Join(rulesErr, whitelistErr)
		}
		return nil
	case "forwarding":
		return s.operateForwarding(request)
	case "docker":
		return s.operateDocker(ctx, request)
	default:
		return fmt.Errorf("unsupported firewall subsystem %q", request.Subsystem)
	}
}

func (s *FirewallSettingService) operateDocker(ctx context.Context, request dto.FirewallBackendOperation) error {
	guard := docker_guard.NewRuntime(request.Backend)
	if request.Operation == "cleanup" {
		if err := guard.Cleanup(); err != nil {
			return err
		}
		return settingRepo.UpdateOrCreate(constant.FirewallDockerPortGuardStatusKey, constant.StatusDisable)
	}
	previous, _ := settingRepo.GetValueByKey(constant.FirewallDockerBackendKey)
	if request.Operation == "select" {
		current := previous
		if current == "" {
			current = alternateDirectBackend(request.Backend)
		}
		initialized, err := dockerGuardBackendInitialized(current)
		if err != nil {
			return err
		}
		if current != request.Backend && initialized {
			return firewallBackendCleanupRequired(current, request.Backend)
		}
	}
	if err := settingRepo.UpdateOrCreate(constant.FirewallDockerBackendKey, request.Backend); err != nil {
		return err
	}
	if request.Operation == "select" {
		if err := (&DockerService{}).UpdateFirewallBackend(request.Backend); err != nil {
			_ = settingRepo.UpdateOrCreate(constant.FirewallDockerBackendKey, previous)
			return err
		}
	}
	if request.Operation == "initialize" {
		if err := newDockerPortGuardService().Operate(ctx, dto.DockerPortGuardOperation{Operation: "initialize"}); err != nil {
			_ = settingRepo.UpdateOrCreate(constant.FirewallDockerBackendKey, previous)
			return err
		}
	}
	return nil
}

func dockerGuardBackendInitialized(backend string) (bool, error) {
	guard := docker_guard.NewRuntime(backend)
	for _, family := range []string{docker_guard.FamilyIPv4, docker_guard.FamilyIPv6} {
		initialized, err := guard.Initialized(family)
		if err != nil {
			return false, err
		}
		if initialized {
			return true, nil
		}
	}
	return false, nil
}

func (s *FirewallSettingService) operateSystem(request dto.FirewallBackendOperation) error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	if _, err := lifecycle.NewClientFor(request.Backend); err != nil {
		return err
	}
	if request.Operation == "cleanup" {
		return cleanupSystemBackend(request.Backend)
	}
	previous, _ := settingRepo.GetValueByKey(constant.FirewallSystemBackendKey)
	if previous == "" {
		if client, err := lifecycle.NewClient(); err == nil {
			previous = client.Name()
		}
	}
	if request.Operation == "select" && previous != "" && previous != request.Backend {
		initialized, err := systemFirewallBackendInitialized(previous)
		if err != nil {
			return err
		}
		if initialized {
			return firewallBackendCleanupRequired(previous, request.Backend)
		}
	}
	if err := settingRepo.UpdateOrCreate(constant.FirewallSystemBackendKey, request.Backend); err != nil {
		return err
	}
	rollback := func(err error) error {
		if err == nil {
			return nil
		}
		_ = settingRepo.UpdateOrCreate(constant.FirewallSystemBackendKey, previous)
		return err
	}
	if request.Operation == "select" {
		return nil
	}
	initErr := newFirewallService().operateFilterChainBaseLocked(request.Backend, dto.FilterChainOperation{
		Name: constant.FirewallBasicChain, Operate: string(firewall.BaseOperationInit),
	})
	if initErr != nil {
		return rollback(initErr)
	}
	return settingRepo.UpdateOrCreate(constant.FirewallFilterInitializedKey, constant.StatusEnable)
}

func systemFirewallBackendInitialized(backend string) (bool, error) {
	return systemFirewallBackendInitializedWithClientFactory(backend, lifecycle.NewClientFor)
}

func systemFirewallBackendInitializedWithClientFactory(
	backend string,
	newClient func(string) (lifecycle.Client, error),
) (bool, error) {
	client, err := newClient(backend)
	if err != nil {
		if errors.Is(err, lifecycle.ErrNotInstalled) {
			return false, nil
		}
		return false, err
	}
	if supportsManagedFilterChains(backend) {
		for _, family := range []string{constant.FirewallFamilyIPv4, constant.FirewallFamilyIPv6} {
			initialized, _, err := loadSystemFirewallFamilyStatus(backend, family)
			if family == constant.FirewallFamilyIPv6 && errors.Is(err, filter.ErrFamilyUnavailable) {
				continue
			}
			if err != nil {
				return false, err
			}
			if initialized {
				return true, nil
			}
		}
		return false, nil
	}
	return client.Status()
}

func cleanupSystemBackend(backend string) error {
	switch backend {
	case constant.FirewallProviderIptables:
		return newIptablesHelperManager().Cleanup()
	case constant.FirewallProviderNftables:
		return newNftablesHelperManager().Cleanup()
	default:
		return fmt.Errorf("cleanup is only available for 1Panel-owned iptables and nftables resources")
	}
}

func cleanupInactiveSystemBackend(backend string) error {
	switch backend {
	case constant.FirewallProviderIptables:
		return (&iptables_helper.Manager{}).Cleanup()
	case constant.FirewallProviderNftables:
		return (&nftables_helper.Manager{}).Cleanup()
	default:
		return fmt.Errorf("cleanup is only available for 1Panel-owned iptables and nftables resources")
	}
}

func (s *FirewallSettingService) operateForwarding(request dto.FirewallBackendOperation) error {
	manager, err := newForwardingManagerFor(request.Backend)
	if err != nil {
		return err
	}
	if request.Operation == "cleanup" {
		if err := manager.Cleanup(); err != nil {
			return err
		}
		if err := settingRepo.UpdateOrCreate(constant.FirewallForwardingInitializedKey, constant.StatusDisable); err != nil {
			return err
		}
		recordForwardingSyncError(nil)
		return nil
	}
	previous, _ := settingRepo.GetValueByKey(constant.FirewallForwardingBackendKey)
	if request.Operation == "select" {
		current := previous
		if current == "" {
			detected, err := newForwardingManager()
			if err != nil {
				return err
			}
			current = detected.Name()
		}
		initialized, err := forwardingBackendInitialized(current)
		if err != nil {
			return err
		}
		if current != request.Backend && initialized {
			return firewallBackendCleanupRequired(current, request.Backend)
		}
	}
	if err := settingRepo.UpdateOrCreate(constant.FirewallForwardingBackendKey, request.Backend); err != nil {
		return err
	}
	if request.Operation == "initialize" {
		return newForwardingService().Enable()
	}
	recordForwardingSyncError(nil)
	return nil
}

func forwardingBackendInitialized(backend string) (bool, error) {
	manager, err := newForwardingManagerFor(backend)
	if err != nil {
		if errors.Is(err, lifecycle.ErrNotInstalled) {
			return false, nil
		}
		return false, err
	}
	for _, family := range []string{constant.FirewallFamilyIPv4, constant.FirewallFamilyIPv6} {
		initialized, _, err := manager.FamilyStatus(family)
		if err != nil {
			return false, err
		}
		if initialized {
			return true, nil
		}
	}
	return false, nil
}

func alternateDirectBackend(backend string) string {
	if backend == constant.FirewallProviderNftables {
		return constant.FirewallProviderIptables
	}
	return constant.FirewallProviderNftables
}
