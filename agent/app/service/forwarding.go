package service

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	forwardClient "github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
)

type IForwardingService interface {
	LoadBaseInfo() (dto.FirewallBaseInfo, error)
	SearchWithPage(search dto.ForwardRuleSearch) (int64, interface{}, error)
	Operate(req dto.ForwardRuleOperate) error
	Enable() error
	Replay() error
	CleanupOwned() error
}

type ForwardingService struct {
	adapterFactory func() (forwardClient.Adapter, error)
	filterFactory  func() (firewall.FilterClient, error)
}

func NewIForwardingService() IForwardingService {
	return &ForwardingService{
		adapterFactory: newForwardingAdapter,
		filterFactory:  firewall.NewFirewallClient,
	}
}

func newForwardingAdapter() (forwardClient.Adapter, error) {
	client, err := firewall.NewFirewallClient()
	if err != nil {
		return nil, err
	}
	return forwardClient.NewAdapter(client.Name())
}

func (s *ForwardingService) LoadBaseInfo() (dto.FirewallBaseInfo, error) {
	baseInfo := dto.FirewallBaseInfo{Version: "-", Name: "-"}
	adapter, err := s.adapterFactory()
	if err != nil {
		global.LOG.Errorf("load forwarding failed, err: %v", err)
		return baseInfo, nil
	}
	filter, err := s.filterFactory()
	if err != nil {
		global.LOG.Errorf("load firewall status failed, err: %v", err)
		return baseInfo, nil
	}
	baseInfo.IsExist = true
	baseInfo.Name = adapter.Name()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		baseInfo.PingStatus = firewall.LoadPingStatus()
		baseInfo.Version, _ = filter.Version()
	}()
	go func() {
		defer wg.Done()
		baseInfo.IsActive, _ = filter.Status()
		baseInfo.IsInit, baseInfo.IsBind = adapter.InitStatus()
	}()
	wg.Wait()
	return baseInfo, nil
}

func (s *ForwardingService) SearchWithPage(req dto.ForwardRuleSearch) (int64, interface{}, error) {
	adapter, err := s.adapterFactory()
	if err != nil {
		return 0, nil, err
	}
	rules, err := adapter.List()
	if err != nil {
		return 0, nil, err
	}
	if req.Strategy != "" {
		return 0, nil, nil
	}

	var filtered []forwardClient.Rule
	for _, rule := range rules {
		if req.Info != "" && !strings.Contains(rule.Port, req.Info) &&
			!strings.Contains(rule.TargetPort, req.Info) && !strings.Contains(rule.TargetIP, req.Info) {
			continue
		}
		filtered = append(filtered, rule)
	}
	total := len(filtered)
	start, end := (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		return int64(total), make([]dto.ForwardRule, 0), nil
	}
	if end > total {
		end = total
	}
	pageRules := filtered[start:end]
	var items []dto.ForwardRule
	if pageRules != nil {
		items = make([]dto.ForwardRule, 0, len(pageRules))
	}
	for _, rule := range pageRules {
		items = append(items, dto.ForwardRule{
			Num:        rule.Num,
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
	adapter, err := s.adapterFactory()
	if err != nil {
		return err
	}

	rules, _ := adapter.List()
	kept := rules[:0]
	for _, rule := range rules {
		shouldKeep := true
		for i := range req.Rules {
			reqRule := &req.Rules[i]
			if reqRule.TargetIP == "" {
				reqRule.TargetIP = "127.0.0.1"
			}
			if reqRule.Operation == "remove" && requestMatchesForwardRule(*reqRule, rule) {
				shouldKeep = false
				break
			}
		}
		if shouldKeep {
			kept = append(kept, rule)
		}
	}

	for _, rule := range kept {
		for _, reqRule := range req.Rules {
			if reqRule.Operation != "remove" && requestMatchesForwardRule(reqRule, rule) {
				return buserr.New("ErrRecordExist")
			}
		}
	}

	sort.SliceStable(req.Rules, func(i, j int) bool {
		if req.Rules[i].Operation == "remove" && req.Rules[j].Operation != "remove" {
			return true
		}
		if req.Rules[i].Operation != "remove" && req.Rules[j].Operation == "remove" {
			return false
		}
		n1, _ := strconv.Atoi(req.Rules[i].Num)
		n2, _ := strconv.Atoi(req.Rules[j].Num)
		return n1 > n2
	})

	for _, rule := range req.Rules {
		for _, protocol := range strings.Split(rule.Protocol, "/") {
			targetIP := rule.TargetIP
			if targetIP == "" {
				targetIP = "127.0.0.1"
			}
			err := adapter.Operate(forwardClient.Rule{
				Num:        rule.Num,
				Protocol:   protocol,
				Port:       rule.Port,
				TargetIP:   targetIP,
				TargetPort: rule.TargetPort,
				Interface:  rule.Interface,
			}, rule.Operation)
			if err == nil {
				continue
			}
			if req.ForceDelete {
				global.LOG.Error(err)
				continue
			}
			return err
		}
	}
	return nil
}

func requestMatchesForwardRule(req dto.ForwardRuleOperation, rule forwardClient.Rule) bool {
	for _, protocol := range strings.Split(req.Protocol, "/") {
		if req.Port == rule.Port && req.TargetPort == rule.TargetPort && req.TargetIP == rule.TargetIP &&
			protocol == rule.Protocol && req.Interface == rule.Interface {
			return true
		}
	}
	return false
}

func (s *ForwardingService) Enable() error {
	adapter, err := s.adapterFactory()
	if err != nil {
		return err
	}
	if err := adapter.Enable(); err != nil {
		return err
	}
	if adapter.Name() != "firewalld" {
		_ = settingRepo.Update("IptablesForwardStatus", constant.StatusEnable)
	}
	return nil
}

func (s *ForwardingService) Replay() error {
	adapter, err := s.adapterFactory()
	if err != nil {
		return err
	}
	if err := adapter.Replay(); err != nil {
		return err
	}
	if adapter.Name() == "firewalld" {
		return nil
	}
	status, _ := settingRepo.GetValueByKey("IptablesForwardStatus")
	if status == constant.StatusEnable {
		return adapter.Enable()
	}
	return nil
}

func (s *ForwardingService) CleanupOwned() error {
	adapter, err := s.adapterFactory()
	if err != nil {
		return fmt.Errorf("load forwarding adapter: %w", err)
	}
	return adapter.CleanupOwned()
}
