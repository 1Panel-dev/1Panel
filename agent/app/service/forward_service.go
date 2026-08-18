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
	managerFactory func() (*forwarding.Manager, error)
	rules          repo.IForwardingRuleRepo
	enabled        func() (bool, error)
	persistBackend func(string) error
}

type forwardingCandidate struct {
	adapter forwarding.Adapter
	runtime forwarding.RuntimeClient
}

var errForwardingBackendConflict = errors.New("iptables and nftables forwarding backends are both initialized; remove one before continuing")
var forwardingMutationMu sync.Mutex

func NewIForwardingService() IForwardingService {
	return &ForwardingService{
		managerFactory: newForwardingManager,
		rules:          repo.NewIForwardingRuleRepo(),
		enabled:        forwardingPersistedEnabled,
		persistBackend: func(backend string) error { return settingRepo.UpdateOrCreate(settingForwardingBackend, backend) },
	}
}

func (s *ForwardingService) LoadBaseInfo() (dto.FirewallSubsystemStatus, error) {
	baseInfo := dto.FirewallSubsystemStatus{Version: "-", Name: "-", Backend: "-"}
	manager, err := s.manager()
	if err != nil {
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
	baseInfo.IsActive, baseInfo.IsInit, baseInfo.IsBind = status.IsActive, status.IsInit, status.IsBind
	return baseInfo, nil
}

func forwardingDisplayName(backend string) string {
	switch backend {
	case "iptables", "nftables":
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
	filtered := stored[:0]
	for _, rule := range stored {
		if request.Info == "" || strings.Contains(rule.Port, request.Info) || strings.Contains(rule.TargetPort, request.Info) || strings.Contains(rule.TargetIP, request.Info) {
			filtered = append(filtered, rule)
		}
	}
	stored = filtered
	total := len(stored)
	start, end := (request.Page-1)*request.PageSize, request.Page*request.PageSize
	if start > total {
		return int64(total), make([]dto.ForwardRule, 0), nil
	}
	if end > total {
		end = total
	}
	pageRules := stored[start:end]
	var items []dto.ForwardRule
	if pageRules != nil {
		items = make([]dto.ForwardRule, 0, len(pageRules))
	}
	for index, rule := range pageRules {
		items = append(items, dto.ForwardRule{
			ID:         rule.ID,
			Num:        strconv.Itoa(start + index + 1),
			Family:     rule.Family,
			Protocol:   rule.Protocol,
			Port:       rule.Port,
			TargetIP:   rule.TargetIP,
			TargetPort: rule.TargetPort,
			Interface:  rule.Interface,
		})
	}
	return int64(total), items, nil
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
		if request.ForceDelete && forwardingOperationsOnlyRemove(request.Rules) {
			if global.LOG != nil {
				global.LOG.Error(err)
			}
			return nil
		}
		return err
	}
	return nil
}

func (s *ForwardingService) Enable() error {
	forwardingMutationMu.Lock()
	defer forwardingMutationMu.Unlock()
	manager, err := s.manager()
	if err != nil {
		return err
	}
	if err := settingRepo.UpdateOrCreate(settingForwardingInitialized, constant.StatusEnable); err != nil {
		return err
	}
	if err := s.activateManager(manager); err != nil {
		return err
	}
	rules, err := s.rules.List(context.Background())
	if err != nil {
		return err
	}
	return manager.Reconcile(forwardingRulesFromModels(rules))
}

func (s *ForwardingService) Restore(ctx context.Context) error {
	forwardingMutationMu.Lock()
	defer forwardingMutationMu.Unlock()
	enabled, err := s.forwardingEnabled()
	if err != nil || !enabled {
		return err
	}
	manager, err := s.manager()
	if err != nil {
		return err
	}
	stored, err := s.rules.List(ctx)
	if err != nil {
		return err
	}
	if err := s.activateManager(manager); err != nil {
		return err
	}
	return manager.Reconcile(forwardingRulesFromModels(stored))
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
	return settingRepo.UpdateOrCreate(settingForwardingBackend, backend)
}

func forwardingPersistedEnabled() (bool, error) {
	status, err := settingRepo.GetValueByKey(settingForwardingInitialized)
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

func applyForwardingOperations(current []forwarding.Rule, requested []dto.ForwardRuleOperation) ([]forwarding.Rule, error) {
	desired := make([]forwarding.Rule, 0, len(current)+len(requested))
	for _, rule := range current {
		normalized, err := forwardingproviders.NormalizeRule(rule)
		if err != nil {
			return nil, fmt.Errorf("normalize persisted forwarding rule: %w", err)
		}
		desired = append(desired, normalized)
	}
	for _, operation := range requested {
		for _, protocol := range strings.Split(operation.Protocol, "/") {
			rule, err := forwardingproviders.NormalizeRule(forwarding.Rule{
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
	for index, rule := range rules {
		if rule.Family == wanted.Family && rule.Protocol == wanted.Protocol && rule.Port == wanted.Port &&
			rule.TargetIP == wanted.TargetIP && rule.TargetPort == wanted.TargetPort && rule.Interface == wanted.Interface {
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
	selected, _ := settingRepo.GetValueByKey(settingForwardingBackend)
	return newForwardingManagerFor(strings.TrimSpace(selected))
}

func newForwardingManagerFor(backend string) (*forwarding.Manager, error) {
	clients, err := lifecycle.NewNetfilterClients()
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("selected forwarding backend %s is not installed", backend)
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
