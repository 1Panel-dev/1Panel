package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	forwardingproviders "github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding/providers"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/ping"
)

type IForwardingService interface {
	LoadBaseInfo() (dto.FirewallSubsystemStatus, error)
	SearchRules(request dto.ForwardRuleSearch) (int64, interface{}, error)
	OperateRules(request dto.ForwardRuleOperate) error
	Enable() error
	Restore(context.Context) error
}

type ForwardingService struct {
	managerFactory     func() (*forwarding.Manager, error)
	rules              repo.IForwardingRuleRepo
	enabled            func() (bool, error)
	persistBackend     func(string) error
	markEnabled        func() error
	installedProviders func() []string
}

type forwardingCandidate struct {
	adapter forwarding.Adapter
	runtime forwarding.RuntimeClient
}

var errForwardingBackendConflict = errors.New("iptables and nftables forwarding backends are both initialized; remove one before continuing")
var errForwardingBackendUnavailable = errors.New("no supported forwarding backend detected")
var forwardingMutationMu sync.Mutex

const (
	forwardingSyncConverged   = "converged"
	forwardingSyncMissing     = "missing"
	forwardingSyncRuntimeOnly = "runtime_only"
)

var (
	forwardingSyncStateMu sync.RWMutex
	forwardingLastSyncErr error
)

func NewIForwardingService() IForwardingService {
	return newForwardingService()
}

func newForwardingService() *ForwardingService {
	return &ForwardingService{
		managerFactory: newForwardingManager,
		rules:          repo.NewIForwardingRuleRepo(),
		enabled:        forwardingPersistedEnabled,
		markEnabled: func() error {
			return settingRepo.UpdateOrCreate(constant.FirewallForwardingInitializedKey, constant.StatusEnable)
		},
		persistBackend: func(backend string) error {
			return settingRepo.UpdateOrCreate(constant.FirewallForwardingBackendKey, backend)
		},
		installedProviders: lifecycle.InstalledProviders,
	}
}

type forwardingRuleSyncCandidate struct {
	rule forwarding.Rule
	err  error
}

func (s *ForwardingService) LoadBaseInfo() (dto.FirewallSubsystemStatus, error) {
	baseInfo := dto.FirewallSubsystemStatus{
		Version: "-", Name: "-", Backend: "-", SyncError: lastForwardingSyncError(),
	}
	manager, err := s.manager()
	if err != nil {
		if errors.Is(err, errForwardingBackendUnavailable) {
			loadInstalled := s.installedProviders
			if loadInstalled == nil {
				loadInstalled = lifecycle.InstalledProviders
			}
			for _, provider := range loadInstalled() {
				if provider == constant.FirewallProviderIptables || provider == constant.FirewallProviderNftables {
					baseInfo.IsExist = true
					baseInfo.Message = err.Error()
					break
				}
			}
			return baseInfo, nil
		}
		return baseInfo, err
	}
	status, err := manager.Status()
	if err != nil {
		return baseInfo, err
	}
	baseInfo.IsExist = true
	baseInfo.Name, baseInfo.Backend = forwardingDisplayName(status.Name), status.Name
	baseInfo.Version = status.Version
	baseInfo.PingStatus = ping.LoadStatus()
	baseInfo.IsInit, baseInfo.IsBind = status.IsInit, status.IsBind
	baseInfo.IPv4 = loadForwardingFamilyInfo(manager, status.Name, constant.FirewallFamilyIPv4)
	baseInfo.IPv6 = loadForwardingFamilyInfo(manager, status.Name, constant.FirewallFamilyIPv6)
	return baseInfo, nil
}

func loadForwardingFamilyInfo(manager *forwarding.Manager, backend, family string) dto.FirewallBackendFamilyStatus {
	initialized, bound, err := manager.FamilyStatus(family)
	available := err == nil
	if backend == constant.FirewallProviderIptables && family == constant.FirewallFamilyIPv6 {
		commands, commandErr := lifecycle.ResolveIptablesCommands()
		available = available && commandErr == nil && commands.IPv6Available()
	}
	return dto.FirewallBackendFamilyStatus{Available: available, Initialized: initialized, Bound: bound}
}

func forwardingDisplayName(backend string) string {
	switch backend {
	case constant.FirewallProviderIptables, constant.FirewallProviderNftables:
		return backend + "-forward"
	default:
		return backend
	}
}

func (s *ForwardingService) SearchRules(request dto.ForwardRuleSearch) (int64, interface{}, error) {
	if request.Strategy != "" {
		return 0, nil, nil
	}
	stored, err := s.rules.List(context.Background())
	if err != nil {
		return 0, nil, err
	}
	manager, err := s.manager()
	if err != nil {
		return 0, nil, err
	}
	runtime, err := manager.List("", "")
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

func (s *ForwardingService) OperateRules(request dto.ForwardRuleOperate) error {
	forwardingMutationMu.Lock()
	defer forwardingMutationMu.Unlock()
	ctx := context.Background()
	stored, err := s.rules.List(ctx)
	if err != nil {
		return err
	}
	desired, err := applyForwardingOperations(forwardingRulesFromModels(stored), request.Rules)
	if errors.Is(err, forwarding.ErrRuleExists) {
		return buserr.New("ErrRecordExist")
	} else if err != nil {
		return err
	}
	if err := s.rules.ReplaceAll(ctx, forwardingRuleModels(desired)); err != nil {
		return err
	}
	if err := s.reconcile(desired); err != nil {
		recordForwardingSyncError(err)
		if request.ForceDelete && forwardingOperationsOnlyRemove(request.Rules) {
			if global.LOG != nil {
				global.LOG.Error(err)
			}
			return nil
		}
		return err
	}
	recordForwardingSyncError(nil)
	return nil
}

func (s *ForwardingService) Enable() error {
	forwardingMutationMu.Lock()
	defer forwardingMutationMu.Unlock()
	manager, err := s.manager()
	if err != nil {
		recordForwardingSyncError(err)
		return err
	}
	if err := s.persistForwardingEnabled(); err != nil {
		recordForwardingSyncError(err)
		return err
	}
	if err := s.activateManager(manager); err != nil {
		recordForwardingSyncError(err)
		return err
	}
	rules, err := s.rules.List(context.Background())
	if err != nil {
		recordForwardingSyncError(err)
		return err
	}
	err = manager.Reconcile(forwardingRulesFromModels(rules))
	recordForwardingSyncError(err)
	return err
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
	manager, err := s.manager()
	if err != nil {
		recordForwardingSyncError(err)
		return err
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		recordForwardingSyncError(err)
		return err
	}
	if err := s.activateManager(manager); err != nil {
		recordForwardingSyncError(err)
		return err
	}
	err = manager.Reconcile(forwardingRulesFromModels(stored))
	recordForwardingSyncError(err)
	return err
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
	if !databaseSyncStatesEqual(actual, desired, func(rule forwarding.Rule) string { return rule.Identity() }) {
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

func (s *ForwardingService) reconcile(rules []forwarding.Rule) error {
	manager, err := s.manager()
	if err != nil {
		return err
	}
	return s.reconcileWithManager(manager, rules)
}

func (s *ForwardingService) reconcileWithManager(manager *forwarding.Manager, rules []forwarding.Rule) error {
	enabled, err := s.forwardingEnabled()
	if err != nil || !enabled {
		return err
	}
	if err := s.activateManager(manager); err != nil {
		return err
	}
	return manager.Reconcile(rules)
}

func (s *ForwardingService) activateManager(manager *forwarding.Manager) error {
	if err := s.saveForwardingBackend(manager.Name()); err != nil {
		return err
	}
	return manager.Enable()
}

func (s *ForwardingService) forwardingEnabled() (bool, error) {
	if s.enabled != nil {
		return s.enabled()
	}
	return forwardingPersistedEnabled()
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

func forwardingPersistedEnabled() (bool, error) {
	status, err := settingRepo.GetValueByKey(constant.FirewallForwardingInitializedKey)
	return status == constant.StatusEnable, err
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

func forwardingRuleModels(rules []forwarding.Rule) []model.ForwardingRule {
	stored := make([]model.ForwardingRule, 0, len(rules))
	for _, rule := range rules {
		stored = append(stored, model.ForwardingRule{
			Family: rule.Family, Protocol: rule.Protocol, Port: rule.Port, TargetIP: rule.TargetIP,
			TargetPort: rule.TargetPort, Interface: rule.Interface,
		})
	}
	return stored
}

type forwardingInventoryItem struct {
	ID        uint
	Rule      forwarding.Rule
	IsDesired bool
	IsRuntime bool
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

func mergeForwardingInventory(
	stored []model.ForwardingRule,
	runtime []forwarding.Rule,
) ([]forwardingInventoryItem, error) {
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

func recordForwardingSyncError(err error) {
	forwardingSyncStateMu.Lock()
	forwardingLastSyncErr = err
	forwardingSyncStateMu.Unlock()
}

func lastForwardingSyncError() string {
	forwardingSyncStateMu.RLock()
	defer forwardingSyncStateMu.RUnlock()
	if forwardingLastSyncErr == nil {
		return ""
	}
	return forwardingLastSyncErr.Error()
}

func applyForwardingOperations(current []forwarding.Rule, requested []dto.ForwardRuleOperation) ([]forwarding.Rule, error) {
	desired := make([]forwarding.Rule, 0, len(current)+len(requested))
	for _, rule := range current {
		normalized, err := forwarding.NormalizeRule(rule)
		if err != nil {
			return nil, fmt.Errorf("normalize persisted forwarding rule: %w", err)
		}
		desired = append(desired, normalized)
	}
	for _, operation := range requested {
		for _, protocol := range strings.Split(operation.Protocol, "/") {
			rule, err := forwarding.NormalizeRule(forwarding.Rule{
				Family: operation.Family, Protocol: protocol, Port: operation.Port, TargetIP: operation.TargetIP,
				TargetPort: operation.TargetPort, Interface: operation.Interface,
			})
			if err != nil {
				return nil, err
			}
			index := forwardingRuleIndex(desired, rule)
			switch forwarding.OperationType(operation.Operation) {
			case forwarding.OperationAdd:
				if index >= 0 {
					return nil, forwarding.ErrRuleExists
				}
				desired = append(desired, rule)
			case forwarding.OperationRemove:
				if index >= 0 {
					desired = append(desired[:index], desired[index+1:]...)
				}
			default:
				return nil, fmt.Errorf("unsupported forwarding operation %q", operation.Operation)
			}
		}
	}
	return desired, nil
}

func forwardingRuleIndex(rules []forwarding.Rule, wanted forwarding.Rule) int {
	wantedIdentity := wanted.Identity()
	for index, rule := range rules {
		if rule.Identity() == wantedIdentity {
			return index
		}
	}
	return -1
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

func (s *ForwardingService) manager() (*forwarding.Manager, error) {
	return s.managerFactory()
}

func newForwardingManager() (*forwarding.Manager, error) {
	selected, _ := settingRepo.GetValueByKey(constant.FirewallForwardingBackendKey)
	if strings.TrimSpace(selected) == "" {
		selected = constant.FirewallProviderIptables
	}
	return newForwardingManagerFor(strings.TrimSpace(selected))
}

func newForwardingManagerFor(backend string) (*forwarding.Manager, error) {
	clients, err := lifecycle.NewNetfilterClients()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errForwardingBackendUnavailable, err)
	}
	candidates := make([]forwardingCandidate, 0, len(clients))
	for _, client := range clients {
		adapter, err := forwardingproviders.New(client.Name())
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, forwardingCandidate{adapter: adapter, runtime: client})
	}
	if backend != "" {
		for _, candidate := range candidates {
			if candidate.adapter.Name() == backend {
				return forwarding.NewManager(candidate.adapter, candidate.runtime), nil
			}
		}
		return nil, fmt.Errorf(
			"%w: selected forwarding backend %s %w",
			errForwardingBackendUnavailable, backend, lifecycle.ErrNotInstalled,
		)
	}
	return selectForwardingManager(candidates)
}

func selectForwardingManager(candidates []forwardingCandidate) (*forwarding.Manager, error) {
	if len(candidates) == 0 {
		return nil, errors.New("no supported forwarding backend detected")
	}
	selected := -1
	for index, candidate := range candidates {
		ipv4Initialized, _, err := candidate.adapter.FamilyStatus(forwarding.FamilyIPv4)
		if err != nil {
			return nil, err
		}
		ipv6Initialized, _, err := candidate.adapter.FamilyStatus(forwarding.FamilyIPv6)
		if err != nil {
			return nil, err
		}
		initialized := ipv4Initialized || ipv6Initialized
		if !initialized {
			continue
		}
		if selected >= 0 {
			return nil, errForwardingBackendConflict
		}
		selected = index
	}
	if selected < 0 {
		selected = 0
	}
	candidate := candidates[selected]
	return forwarding.NewManager(candidate.adapter, candidate.runtime), nil
}
