package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	forwardClient "github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	forwardProviders "github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding/providers"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/ping"
)

type IForwardingService interface {
	LoadBaseInfo() (dto.FirewallBaseInfo, error)
	SearchWithPage(search dto.ForwardRuleSearch) (int64, interface{}, error)
	Operate(req dto.ForwardRuleOperate) error
	Enable() error
	Replay() error
}

type ForwardingService struct {
	managerFactory func() (*forwardClient.Manager, error)
}

type forwardingCandidate struct {
	adapter forwardClient.Adapter
	runtime forwardClient.RuntimeClient
}

var errForwardingBackendConflict = errors.New("iptables and nftables forwarding backends are both initialized; remove one before continuing")

func NewIForwardingService() IForwardingService {
	return &ForwardingService{
		managerFactory: newForwardingManager,
	}
}

func (s *ForwardingService) LoadBaseInfo() (dto.FirewallBaseInfo, error) {
	baseInfo := dto.FirewallBaseInfo{Version: "-", Name: "-", Backend: "-"}
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

func (s *ForwardingService) SearchWithPage(req dto.ForwardRuleSearch) (int64, interface{}, error) {
	manager, err := s.manager()
	if err != nil {
		return 0, nil, err
	}
	if req.Strategy != "" {
		return 0, nil, nil
	}
	rules, err := manager.List(req.Info, req.Strategy)
	if err != nil {
		return 0, nil, err
	}
	total := len(rules)
	start, end := (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		return int64(total), make([]dto.ForwardRule, 0), nil
	}
	if end > total {
		end = total
	}
	pageRules := rules[start:end]
	var items []dto.ForwardRule
	if pageRules != nil {
		items = make([]dto.ForwardRule, 0, len(pageRules))
	}
	for _, rule := range pageRules {
		items = append(items, dto.ForwardRule{
			Num:        rule.Num,
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

func (s *ForwardingService) Operate(req dto.ForwardRuleOperate) error {
	manager, err := s.manager()
	if err != nil {
		return err
	}

	operations := make([]forwardClient.Operation, 0, len(req.Rules))
	for _, rule := range req.Rules {
		operations = append(operations, forwardClient.Operation{Rule: forwardClient.Rule{
			Num: rule.Num, Family: rule.Family, Protocol: rule.Protocol, Port: rule.Port, TargetIP: rule.TargetIP,
			TargetPort: rule.TargetPort, Interface: rule.Interface,
		}, Operation: forwardClient.OperationType(rule.Operation)})
	}
	if err := manager.Operate(operations, req.ForceDelete); errors.Is(err, forwardClient.ErrRuleExists) {
		return buserr.New("ErrRecordExist")
	} else if err != nil {
		return err
	}
	return nil
}

func (s *ForwardingService) Enable() error {
	manager, err := s.manager()
	if err != nil {
		return err
	}
	if err := manager.Enable(); err != nil {
		return err
	}
	return settingRepo.Update("IptablesForwardStatus", constant.StatusEnable)
}

func (s *ForwardingService) Replay() error {
	manager, err := s.manager()
	if err != nil {
		return err
	}
	if err := manager.Replay(); err != nil {
		return err
	}
	status, err := settingRepo.GetValueByKey("IptablesForwardStatus")
	if err != nil {
		return err
	}
	if status == constant.StatusEnable {
		return manager.Enable()
	}
	return nil
}

func (s *ForwardingService) manager() (*forwardClient.Manager, error) {
	return s.managerFactory()
}

func newForwardingManager() (*forwardClient.Manager, error) {
	selected, _ := settingRepo.GetValueByKey(settingForwardingBackend)
	return newForwardingManagerFor(strings.TrimSpace(selected))
}

func newForwardingManagerFor(backend string) (*forwardClient.Manager, error) {
	clients, err := lifecycle.NewNetfilterClients()
	if err != nil {
		return nil, err
	}
	candidates := make([]forwardingCandidate, 0, len(clients))
	for _, client := range clients {
		adapter, err := forwardProviders.New(client.Name())
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, forwardingCandidate{adapter: adapter, runtime: client})
	}
	if backend != "" {
		for _, candidate := range candidates {
			if candidate.adapter.Name() == backend {
				return forwardClient.NewManager(candidate.adapter, candidate.runtime), nil
			}
		}
		return nil, fmt.Errorf("selected forwarding backend %s is not installed", backend)
	}
	return selectForwardingManager(candidates)
}

func selectForwardingManager(candidates []forwardingCandidate) (*forwardClient.Manager, error) {
	if len(candidates) == 0 {
		return nil, errors.New("no supported forwarding backend detected")
	}
	selected := -1
	for index, candidate := range candidates {
		ipv4Initialized, _, err := candidate.adapter.FamilyStatus(forwardClient.FamilyIPv4)
		if err != nil {
			return nil, err
		}
		ipv6Initialized, _, err := candidate.adapter.FamilyStatus(forwardClient.FamilyIPv6)
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
	return forwardClient.NewManager(candidate.adapter, candidate.runtime), nil
}
