package service

import (
	"context"
	"fmt"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/client"
)

type IIptablesFilterService interface {
	GetFilterRules(chains []string) ([]dto.IptablesChainInfo, error)
	AddRule(req dto.IptablesFilterRuleOperate) error
	RemoveRule(id uint) error
	BatchOperate(req dto.IptablesFilterBatchOperate) error
	InitChains() error
	ApplyFirewall() error
	UnloadFirewall() error
	ReloadRules() error
}

type IptablesFilterService struct {
	repo           repo.IIptablesFilterRuleRepo
	iptablesClient *client.Iptables
}

func NewIIptablesFilterService() IIptablesFilterService {
	iptablesClient, _ := client.NewIptables()
	return &IptablesFilterService{
		repo:           repo.NewIIptablesFilterRuleRepo(),
		iptablesClient: iptablesClient,
	}
}

const (
	ChainInput        = "INPUT"
	ChainOutput       = "OUTPUT"
	Chain1PanelInput  = "1PANEL_INPUT"
	Chain1PanelOutput = "1PANEL_OUTPUT"
)

func (s *IptablesFilterService) GetFilterRules(chains []string) ([]dto.IptablesChainInfo, error) {
	if len(chains) == 0 {
		chains = []string{ChainInput, ChainOutput, Chain1PanelInput, Chain1PanelOutput}
	}

	// 获取 iptables 版本
	version, err := s.iptablesClient.GetVersion()
	if err != nil {
		global.LOG.Warnf("failed to get iptables version: %v", err)
		version = "unknown"
	}

	// 读取 iptables 规则
	iptablesChains, err := s.iptablesClient.ReadFilter(chains)
	if err != nil {
		return nil, fmt.Errorf("failed to read iptables rules: %w", err)
	}

	// 从数据库读取规则描述
	ctx := context.Background()
	dbRules, _ := s.repo.List(ctx, []string{Chain1PanelInput, Chain1PanelOutput})
	descMap := make(map[uint]string)
	for _, rule := range dbRules {
		descMap[rule.ID] = rule.Description
	}

	var result []dto.IptablesChainInfo
	for _, chainName := range chains {
		chain, ok := iptablesChains[chainName]
		if !ok {
			continue
		}

		chainInfo := dto.IptablesChainInfo{
			Version:       version,
			Name:          chainName,
			DefaultPolicy: chain.DefaultPolicy,
			Rules:         []dto.IptablesFilterRuleInfo{},
			IsApplied:     false,
		}

		// 检查是否已应用（INPUT/OUTPUT 链包含跳转规则）
		if chainName == ChainInput || chainName == ChainOutput {
			item := chain.FirstRule
			targetChain := Chain1PanelInput
			if chainName == ChainOutput {
				targetChain = Chain1PanelOutput
			}
			for item != nil {
				if item.P.Action == targetChain {
					chainInfo.IsApplied = true
					break
				}
				item = item.Next()
			}
		}

		// 转换规则
		item := chain.FirstRule
		order := 0
		for item != nil {
			p := item.P
			ruleInfo := dto.IptablesFilterRuleInfo{
				Protocol:   p.Protocol,
				SourceIP:   p.SourceIP,
				SourcePort: p.SrcPort,
				DestIP:     p.DestIP,
				DestPort:   p.DstPort,
				Action:     p.Action,
				Comment:    p.Comment,
				RuleOrder:  order,
			}

			// 从 comment 中提取 ID（格式：1Panel_FilterRule_<ID>）
			if p.Comment != "" {
				var id uint
				fmt.Sscanf(p.Comment, "1Panel_FilterRule_%d", &id)
				ruleInfo.ID = id
				if desc, ok := descMap[id]; ok {
					ruleInfo.Description = desc
				}
			}

			chainInfo.Rules = append(chainInfo.Rules, ruleInfo)
			order++
			item = item.Next()
		}

		result = append(result, chainInfo)
	}

	return result, nil
}

func (s *IptablesFilterService) AddRule(req dto.IptablesFilterRuleOperate) error {
	if req.Chain != Chain1PanelInput && req.Chain != Chain1PanelOutput {
		return fmt.Errorf("Only %s or %s chains are allowed", Chain1PanelInput, Chain1PanelOutput)
	}

	// 安全检查：防止无条件 DROP/REJECT
	if (req.Action == "DROP" || req.Action == "REJECT") &&
		req.Protocol == "" && req.SourceIP == "" && req.DestIP == "" &&
		req.SourcePort == 0 && req.DestPort == 0 {
		return fmt.Errorf("Iptables Rule Security Check: unconditional %s rules are not allowed, this may lock you out of the system", req.Action)
	}

	ctx := context.Background()

	// 获取最大规则顺序
	maxOrder, err := s.repo.GetMaxRuleOrder(ctx, req.Chain)
	if err != nil {
		return fmt.Errorf("failed to get max rule order: %w", err)
	}

	// 创建数据库记录
	rule := &model.IptablesFilterRule{
		Chain:       req.Chain,
		Protocol:    req.Protocol,
		SourceIP:    req.SourceIP,
		SourcePort:  req.SourcePort,
		DestIP:      req.DestIP,
		DestPort:    req.DestPort,
		Action:      req.Action,
		Description: req.Description,
		RuleOrder:   maxOrder + 1,
	}

	if err := s.repo.Create(ctx, rule); err != nil {
		return fmt.Errorf("failed to save rule to database: %w", err)
	}

	// 生成 comment
	rule.Comment = fmt.Sprintf("1Panel_FilterRule_%d", rule.ID)

	// 添加 iptables 规则
	policy := client.IptablesPolicy{
		Protocol: req.Protocol,
		SourceIP: req.SourceIP,
		SrcPort:  req.SourcePort,
		DestIP:   req.DestIP,
		DstPort:  req.DestPort,
		Action:   req.Action,
		Comment:  rule.Comment,
	}

	if err := s.iptablesClient.AddPolicy(req.Chain, policy); err != nil {
		// 回滚数据库
		_ = s.repo.Delete(ctx, rule.ID)
		return fmt.Errorf("failed to add iptables rule: %w", err)
	}

	// 更新数据库 comment
	rule.Comment = fmt.Sprintf("1Panel_FilterRule_%d", rule.ID)
	return global.DB.Model(rule).Update("comment", rule.Comment).Error
}

func (s *IptablesFilterService) RemoveRule(id uint) error {
	ctx := context.Background()

	// 从数据库查询规则
	rule, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("rule not found: %w", err)
	}

	// 按 comment 删除 iptables 规则
	if rule.Comment != "" {
		if err := s.iptablesClient.RemovePolicyByComment(rule.Chain, rule.Comment); err != nil {
			global.LOG.Warnf("failed to remove iptables rule by comment: %v, trying to delete by policy", err)
			// 如果按 comment 删除失败，尝试按规则内容删除
			policy := client.IptablesPolicy{
				Protocol: rule.Protocol,
				SourceIP: rule.SourceIP,
				SrcPort:  rule.SourcePort,
				DestIP:   rule.DestIP,
				DstPort:  rule.DestPort,
				Action:   rule.Action,
				Comment:  rule.Comment,
			}
			if err := s.iptablesClient.DeletePolicy(rule.Chain, policy); err != nil {
				return fmt.Errorf("failed to remove iptables rule: %w", err)
			}
		}
	}

	// 删除数据库记录
	return s.repo.Delete(ctx, id)
}

func (s *IptablesFilterService) BatchOperate(req dto.IptablesFilterBatchOperate) error {
	for _, ruleReq := range req.Rules {
		if ruleReq.Operation == "add" {
			if err := s.AddRule(ruleReq); err != nil {
				return err
			}
		} else if ruleReq.Operation == "remove" {
			if err := s.RemoveRule(ruleReq.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *IptablesFilterService) InitChains() error {
	// 检查并创建 1PANEL_INPUT
	exists, err := s.iptablesClient.ChainExists(client.FilterTab, Chain1PanelInput)
	if err != nil {
		return fmt.Errorf("failed to check chain %s: %w", Chain1PanelInput, err)
	}

	if !exists {
		if err := s.iptablesClient.NewChain(client.FilterTab, Chain1PanelInput); err != nil {
			return fmt.Errorf("failed to create chain %s: %w", Chain1PanelInput, err)
		}
	} else {
		// 清空已存在的链
		if err := s.iptablesClient.FlushChain(client.FilterTab, Chain1PanelInput); err != nil {
			return err
		}
	}

	// 检查并创建 1PANEL_OUTPUT
	exists, err = s.iptablesClient.ChainExists(client.FilterTab, Chain1PanelOutput)
	if err != nil {
		return fmt.Errorf("failed to check chain %s: %w", Chain1PanelOutput, err)
	}

	if !exists {
		if err := s.iptablesClient.NewChain(client.FilterTab, Chain1PanelOutput); err != nil {
			return fmt.Errorf("failed to create chain %s: %w", Chain1PanelOutput, err)
		}
	} else {
		if err := s.iptablesClient.FlushChain(client.FilterTab, Chain1PanelOutput); err != nil {
			return err
		}
	}

	return nil
}

func (s *IptablesFilterService) ApplyFirewall() error {
	// 安全检查
	chains, err := s.iptablesClient.ReadFilter([]string{Chain1PanelInput, Chain1PanelOutput})
	if err != nil {
		return fmt.Errorf("failed to read filter chains: %w", err)
	}

	// 检查 1PANEL_INPUT 安全性
	if inputChain, ok := chains[Chain1PanelInput]; ok {
		item := inputChain.FirstRule
		for item != nil {
			p := item.P
			if (p.Action == "DROP" || p.Action == "REJECT") &&
				p.Protocol == "" && p.SourceIP == "" && p.DestIP == "" &&
				p.SrcPort == 0 && p.DstPort == 0 {
				return fmt.Errorf("Chain %s includes unconditional %s rule, not allowed to apply", Chain1PanelInput, p.Action)
			}
			item = item.Next()
		}
	}

	// 检查 1PANEL_OUTPUT 安全性
	if outputChain, ok := chains[Chain1PanelOutput]; ok {
		item := outputChain.FirstRule
		for item != nil {
			p := item.P
			if (p.Action == "DROP" || p.Action == "REJECT") &&
				p.Protocol == "" && p.SourceIP == "" && p.DestIP == "" &&
				p.SrcPort == 0 && p.DstPort == 0 {
				return fmt.Errorf("Chain %s includes unconditional %s rule, not allowed to apply", Chain1PanelOutput, p.Action)
			}
			item = item.Next()
		}
	}

	if err := s.iptablesClient.Setup1PanelFirewallChains("input"); err != nil {
		return fmt.Errorf("failed to apply %s to %s: %w", Chain1PanelInput, ChainInput, err)
	}
	global.LOG.Infof("Applied %s to %s chain", Chain1PanelInput, ChainInput)

	if err := s.iptablesClient.Setup1PanelFirewallChains("output"); err != nil {
		return fmt.Errorf("failed to apply %s to %s: %w", Chain1PanelInput, ChainInput, err)
	}
	global.LOG.Infof("Applied %s to %s chain", Chain1PanelInput, ChainInput)

	return nil
}

func (s *IptablesFilterService) UnloadFirewall() error {
	// 从 INPUT 链删除跳转规则
	ruleNum, err := s.iptablesClient.FindRuleNum(ChainInput, Chain1PanelInput)
	if err == nil {
		if err := s.iptablesClient.RemovePolicyByNum(ChainInput, ruleNum); err != nil {
			return fmt.Errorf("failed to unload %s from %s: %w", Chain1PanelInput, ChainInput, err)
		}
		global.LOG.Infof("Unloaded %s from %s chain", Chain1PanelInput, ChainInput)
	} else {
		global.LOG.Warnf("%s not found in %s chain: %v", Chain1PanelInput, ChainInput, err)
	}

	// 从 OUTPUT 链删除跳转规则
	ruleNum, err = s.iptablesClient.FindRuleNum(ChainOutput, Chain1PanelOutput)
	if err == nil {
		if err := s.iptablesClient.RemovePolicyByNum(ChainOutput, ruleNum); err != nil {
			return fmt.Errorf("failed to unload %s from %s: %w", Chain1PanelOutput, ChainOutput, err)
		}
		global.LOG.Infof("Unloaded %s from %s chain", Chain1PanelOutput, ChainOutput)
	} else {
		global.LOG.Warnf("%s not found in %s chain: %v", Chain1PanelOutput, ChainOutput, err)
	}

	return nil
}

func (s *IptablesFilterService) ReloadRules() error {
	// 清空自定义链
	if err := s.iptablesClient.FlushChain(client.FilterTab, Chain1PanelInput); err != nil {
		return fmt.Errorf("failed to flush chain %s: %w", Chain1PanelInput, err)
	}
	if err := s.iptablesClient.FlushChain(client.FilterTab, Chain1PanelOutput); err != nil {
		return fmt.Errorf("failed to flush chain %s: %w", Chain1PanelOutput, err)
	}

	// 从数据库读取规则
	ctx := context.Background()
	rules, err := s.repo.List(ctx, []string{Chain1PanelInput, Chain1PanelOutput})
	if err != nil {
		return fmt.Errorf("failed to load rules from database: %w", err)
	}

	// 逐条恢复规则
	for _, rule := range rules {
		policy := client.IptablesPolicy{
			Protocol: rule.Protocol,
			SourceIP: rule.SourceIP,
			SrcPort:  rule.SourcePort,
			DestIP:   rule.DestIP,
			DstPort:  rule.DestPort,
			Action:   rule.Action,
			Comment:  rule.Comment,
		}

		if err := s.iptablesClient.AddPolicy(rule.Chain, policy); err != nil {
			global.LOG.Errorf("failed to reload rule %d: %v", rule.ID, err)
			continue
		}
	}

	global.LOG.Infof("Reloaded %d firewall rules from database", len(rules))
	return nil
}
