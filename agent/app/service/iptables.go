package service

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/client"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/client/iptables"
)

type IIptablesService interface {
	Search(req dto.SearchPageWithType) (int64, interface{}, error)
	OperateRule(req dto.IptablesRuleOp, withSave bool) error
	BatchOperate(req dto.IptablesBatchOperate) error
	LoadChainStatus(req dto.OperationWithName) dto.IptablesChainStatus

	Operate(req dto.IptablesOp) error
}

type IptablesService struct{}

func NewIIptablesService() IIptablesService {
	return &IptablesService{}
}

func (s *IptablesService) Search(req dto.SearchPageWithType) (int64, interface{}, error) {
	rules, err := iptables.ReadFilterRulesByChain(req.Type)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read iptables rules: %w", err)
	}
	var records []iptables.FilterRules
	total, start, end := len(rules), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		records = make([]iptables.FilterRules, 0)
	} else {
		if end >= total {
			end = total
		}
		records = rules[start:end]
	}

	rulesInDB, _ := hostRepo.ListFirewallRecord(hostRepo.WithByChain(req.Type))

	for i := 0; i < len(records); i++ {
		for _, item := range rulesInDB {
			if records[i].Strategy == item.Strategy &&
				records[i].DstIP == item.DstIP &&
				fmt.Sprintf("%v", records[i].DstPort) == item.DstPort &&
				records[i].Protocol == item.Protocol &&
				records[i].SrcIP == item.SrcIP &&
				fmt.Sprintf("%v", records[i].SrcPort) == item.SrcPort {
				records[i].ID = item.ID
				records[i].Description = item.Description
			}
		}
	}
	return int64(total), records, nil
}

func (s *IptablesService) OperateRule(req dto.IptablesRuleOp, withSave bool) error {
	action := "ACCEPT"
	if req.Strategy == "drop" {
		action = "DROP"
	}
	policy := iptables.FilterRules{
		Protocol: req.Protocol,
		SrcIP:    req.SrcIP,
		DstIP:    req.DstIP,
		Strategy: action,
	}
	if req.SrcPort != 0 {
		policy.SrcPort = fmt.Sprintf("%v", req.SrcPort)
	}
	if req.DstPort != 0 {
		policy.DstPort = fmt.Sprintf("%v", req.DstPort)
	}

	name := iptables.InputFileName
	if req.Chain == iptables.Chain1PanelOutput {
		name = iptables.OutputFileName
	}
	switch req.Operation {
	case "add":
		if err := s.validateRuleInput(&req); err != nil {
			return err
		}

		if err := iptables.AddFilterRule(req.Chain, policy); err != nil {
			return fmt.Errorf("failed to add iptables rule: %w", err)
		}

		if len(req.Description) != 0 {
			rule := &model.Firewall{
				Chain:       req.Chain,
				Protocol:    req.Protocol,
				SrcIP:       req.SrcIP,
				SrcPort:     policy.SrcPort,
				DstIP:       req.DstIP,
				DstPort:     policy.DstPort,
				Strategy:    req.Strategy,
				Description: req.Description,
			}

			if err := hostRepo.SaveFirewallRecord(rule); err != nil {
				return fmt.Errorf("failed to save rule to database: %w", err)
			}
		}
	case "remove":
		if err := iptables.DeleteFilterRule(req.Chain, policy); err != nil {
			return fmt.Errorf("failed to remove iptables rule: %w", err)
		}
		if req.ID != 0 {
			if err := hostRepo.DeleteFirewallRecordByID(req.ID); err != nil {
				return fmt.Errorf("failed to delete rule from database: %w", err)
			}
		}
	}

	if !withSave {
		return nil
	}
	if err := iptables.SaveRulesToFile(iptables.FilterTab, req.Chain, name); err != nil {
		global.LOG.Errorf("persistence for %s failed, err: %v", iptables.Chain1PanelBasic, err)
	}
	return nil
}

func (s *IptablesService) BatchOperate(req dto.IptablesBatchOperate) error {
	if len(req.Rules) == 0 {
		return errors.New("no rules to operate")
	}
	for _, rule := range req.Rules {
		if err := s.OperateRule(rule, false); err != nil {
			return err
		}
	}
	chain := iptables.Chain1PanelInput
	fileName := iptables.InputFileName
	if req.Rules[0].Chain == iptables.Chain1PanelOutput {
		chain = iptables.Chain1PanelOutput
		fileName = iptables.OutputFileName
	}
	if err := iptables.SaveRulesToFile(iptables.FilterTab, chain, fileName); err != nil {
		global.LOG.Errorf("persistence for %s failed, err: %v", iptables.Chain1PanelBasic, err)
	}
	return nil
}

func (s *IptablesService) Operate(req dto.IptablesOp) error {
	targetChain := iptables.ChainInput
	if req.Name == iptables.Chain1PanelOutput {
		targetChain = iptables.ChainOutput
	}
	switch req.Operate {
	case "init-base":
		if ok := cmd.Which("iptables"); !ok {
			return fmt.Errorf("failed to find iptables")
		}
		if err := iptables.AddChain(iptables.FilterTab, iptables.Chain1PanelBasicBefore); err != nil {
			return err
		}
		if err := iptables.AddChain(iptables.FilterTab, iptables.Chain1PanelBasic); err != nil {
			return err
		}
		if err := iptables.AddChain(iptables.FilterTab, iptables.Chain1PanelBasicAfter); err != nil {
			return err
		}
		if err := initPreRules(); err != nil {
			return err
		}
		if err := iptables.BindChain(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelBasicBefore, 1); err != nil {
			return err
		}
		if err := iptables.BindChain(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelBasic, 2); err != nil {
			return err
		}
		if err := iptables.BindChain(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelBasicAfter, 3); err != nil {
			return err
		}
		if err := iptables.SaveRulesToFile(iptables.FilterTab, iptables.Chain1PanelBasicBefore, iptables.BasicBeforeFileName); err != nil {
			return err
		}
		if err := iptables.SaveRulesToFile(iptables.FilterTab, iptables.Chain1PanelBasicAfter, iptables.BasicAfterFileName); err != nil {
			return err
		}
		_ = settingRepo.Update("IptablesStatus", constant.StatusEnable)
		return nil
	case "init-forward":
		if err := client.EnableIptablesForward(); err != nil {
			return err
		}
		_ = settingRepo.Update("IptablesForwardStatus", constant.StatusEnable)
		return nil
	case "init-advance":
		if err := iptables.AddChain(iptables.FilterTab, iptables.Chain1PanelInput); err != nil {
			return err
		}
		if err := iptables.AddChain(iptables.FilterTab, iptables.Chain1PanelOutput); err != nil {
			return err
		}
		if err := iptables.BindChain(iptables.FilterTab, iptables.ChainOutput, iptables.Chain1PanelOutput, 1); err != nil {
			return err
		}
		number := loadBindNumber(iptables.Chain1PanelInput)
		if err := iptables.BindChain(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelInput, number); err != nil {
			return err
		}
		_ = settingRepo.Update("IptablesInputStatus", constant.StatusEnable)
		_ = settingRepo.Update("IptablesOutputStatus", constant.StatusEnable)
		return nil
	case "bind-base":
		if err := initPreRules(); err != nil {
			return err
		}
		if err := iptables.BindChain(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelBasicBefore, 1); err != nil {
			return err
		}
		if err := iptables.BindChain(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelBasic, 2); err != nil {
			return err
		}
		if err := iptables.BindChain(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelBasicAfter, 3); err != nil {
			return err
		}
		_ = settingRepo.Update("IptablesStatus", constant.StatusEnable)
		return nil
	case "bind-base-without-init":
		if err := iptables.BindChain(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelBasicBefore, 1); err != nil {
			return err
		}
		if err := iptables.BindChain(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelBasic, 2); err != nil {
			return err
		}
		if err := iptables.BindChain(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelBasicAfter, 3); err != nil {
			return err
		}
		_ = settingRepo.Update("IptablesStatus", constant.StatusEnable)
		return nil
	case "unbind-base":
		if err := iptables.UnbindChain(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelBasicAfter); err != nil {
			return err
		}
		if err := iptables.UnbindChain(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelBasicBefore); err != nil {
			return err
		}
		if err := iptables.UnbindChain(iptables.FilterTab, iptables.ChainInput, iptables.Chain1PanelBasic); err != nil {
			return err
		}
		_ = settingRepo.Update("IptablesStatus", constant.StatusDisable)
		return nil
	case "bind":
		if err := iptables.BindChain(iptables.FilterTab, targetChain, req.Name, loadBindNumber(req.Name)); err != nil {
			return err
		}
		if req.Name == iptables.Chain1PanelInput {
			_ = settingRepo.Update("IptablesInputStatus", constant.StatusEnable)
		}
		if req.Name == iptables.Chain1PanelOutput {
			_ = settingRepo.Update("IptablesOutputStatus", constant.StatusEnable)
		}
		return nil
	case "unbind":
		if err := iptables.UnbindChain(iptables.FilterTab, targetChain, req.Name); err != nil {
			return err
		}
		if req.Name == iptables.Chain1PanelInput {
			_ = settingRepo.Update("IptablesInputStatus", constant.StatusDisable)
		}
		if req.Name == iptables.Chain1PanelOutput {
			_ = settingRepo.Update("IptablesOutputStatus", constant.StatusDisable)
		}
		return nil
	}
	return nil
}

func (s *IptablesService) LoadChainStatus(req dto.OperationWithName) dto.IptablesChainStatus {
	var data dto.IptablesChainStatus
	var err error
	data.DefaultStrategy, err = iptables.LoadDefaultStrategy(req.Name)
	if err != nil {
		global.LOG.Error(err)
	}
	switch req.Name {
	case iptables.Chain1PanelBasic:
		data.IsBind, err = iptables.CheckChainBind(iptables.FilterTab, iptables.ChainInput, req.Name)
	case iptables.Chain1PanelInput:
		data.IsBind, err = iptables.CheckChainBind(iptables.FilterTab, iptables.ChainInput, req.Name)
	case iptables.Chain1PanelOutput:
		data.IsBind, err = iptables.CheckChainBind(iptables.FilterTab, iptables.ChainOutput, req.Name)
	}
	return data
}

func (s *IptablesService) validateRuleInput(req *dto.IptablesRuleOp) error {
	if req.Protocol != "" {
		validProtocols := map[string]bool{"tcp": true, "udp": true, "icmp": true, "all": true}
		if !validProtocols[strings.ToLower(req.Protocol)] {
			return fmt.Errorf("invalid protocol: %s, must be tcp, udp, icmp or all", req.Protocol)
		}
	}
	if req.SrcIP != "" {
		if err := s.validateIPOrCIDR(req.SrcIP); err != nil {
			return fmt.Errorf("invalid source IP: %w", err)
		}
	}
	if req.DstIP != "" {
		if err := s.validateIPOrCIDR(req.DstIP); err != nil {
			return fmt.Errorf("invalid destination IP: %w", err)
		}
	}
	if req.SrcPort > 65535 {
		return fmt.Errorf("invalid source port: %d, must be between 1 and 65535", req.SrcPort)
	}
	if req.DstPort > 65535 {
		return fmt.Errorf("invalid destination port: %d, must be between 1 and 65535", req.DstPort)
	}
	if (req.SrcPort > 0 || req.DstPort > 0) && req.Protocol == "" {
		return fmt.Errorf("port specification requires protocol (tcp/udp)")
	}

	return nil
}

func (s *IptablesService) validateIPOrCIDR(ipStr string) error {
	if strings.Contains(ipStr, "/") {
		_, _, err := net.ParseCIDR(ipStr)
		if err != nil {
			return fmt.Errorf("invalid CIDR format: %w", err)
		}
		return nil
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return fmt.Errorf("invalid IP address format")
	}

	return nil
}

func loadBindNumber(chain string) int {
	if chain == iptables.Chain1PanelOutput {
		return 1
	}
	number := 1
	if exist, _ := iptables.CheckChainExist(iptables.FilterTab, iptables.Chain1PanelBasicBefore); exist {
		number++
	}
	if exist, _ := iptables.CheckChainExist(iptables.FilterTab, iptables.Chain1PanelBasic); exist {
		number++
	}
	return number
}

func initPreRules() error {
	if err := iptables.AddRule(iptables.FilterTab, iptables.Chain1PanelBasicBefore, iptables.IoRuleIn); err != nil {
		return err
	}
	if err := iptables.AddRule(iptables.FilterTab, iptables.Chain1PanelBasicBefore, iptables.EstablishedRule); err != nil {
		return err
	}
	panelPort := LoadPanelPort()
	if len(panelPort) == 0 {
		return errors.New("find 1panel service port failed")
	}
	ports := []string{"80", "443", panelPort, loadSSHPort()}
	for _, item := range ports {
		if err := iptables.AddRule(iptables.FilterTab, iptables.Chain1PanelBasicBefore, fmt.Sprintf("-p tcp -m tcp --dport %v -j ACCEPT", item)); err != nil {
			return err
		}
	}
	if err := iptables.AddRule(iptables.FilterTab, iptables.Chain1PanelBasicAfter, "-p udp -m udp --dport 443 -j ACCEPT"); err != nil {
		return err
	}
	if err := iptables.AddRule(iptables.FilterTab, iptables.Chain1PanelBasicAfter, iptables.DropAllTcp); err != nil {
		return err
	}
	if err := iptables.AddRule(iptables.FilterTab, iptables.Chain1PanelBasicAfter, iptables.DropAllUdp); err != nil {
		return err
	}
	return nil
}

func LoadPanelPort() string {
	if !global.IsMaster {
		return global.CONF.Base.Port
	} else {
		var portSetting model.Setting
		_ = global.CoreDB.Where("key = ?", "ServerPort").First(&portSetting).Error
		if len(portSetting.Value) != 0 {
			return portSetting.Value
		}
	}
	return ""
}
