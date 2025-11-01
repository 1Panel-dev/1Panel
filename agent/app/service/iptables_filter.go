package service

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

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
	mu             sync.Mutex
}

func NewIIptablesFilterService() IIptablesFilterService {
	iptablesClient, err := client.NewIptables()
	if err != nil {
		global.LOG.Errorf("failed to create iptables client: %v", err)
		return nil
	}
	return &IptablesFilterService{
		repo:           repo.NewIIptablesFilterRuleRepo(),
		iptablesClient: iptablesClient,
	}
}

func (s *IptablesFilterService) GetFilterRules(chains []string) ([]dto.IptablesChainInfo, error) {
	if len(chains) == 0 {
		chains = []string{client.ChainInput, client.ChainOutput, client.Chain1PanelInput, client.Chain1PanelOutput}
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
	dbRules, _ := s.repo.List(ctx, []string{client.Chain1PanelInput, client.Chain1PanelOutput})
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
		if chainName == client.ChainInput || chainName == client.ChainOutput {
			item := chain.FirstRule
			targetChain := client.Chain1PanelInput
			if chainName == client.ChainOutput {
				targetChain = client.Chain1PanelOutput
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

func (s *IptablesFilterService) validateRuleInput(req *dto.IptablesFilterRuleOperate) error {
	// 验证协议
	if req.Protocol != "" {
		validProtocols := map[string]bool{"tcp": true, "udp": true, "icmp": true, "all": true}
		if !validProtocols[strings.ToLower(req.Protocol)] {
			return fmt.Errorf("invalid protocol: %s, must be tcp, udp, icmp or all", req.Protocol)
		}
	}

	// 验证源 IP
	if req.SourceIP != "" {
		if err := s.validateIPOrCIDR(req.SourceIP); err != nil {
			return fmt.Errorf("invalid source IP: %w", err)
		}
	}

	// 验证目标 IP
	if req.DestIP != "" {
		if err := s.validateIPOrCIDR(req.DestIP); err != nil {
			return fmt.Errorf("invalid destination IP: %w", err)
		}
	}

	// 验证端口范围（1-65535）
	if req.SourcePort > 65535 {
		return fmt.Errorf("invalid source port: %d, must be between 1 and 65535", req.SourcePort)
	}
	if req.DestPort > 65535 {
		return fmt.Errorf("invalid destination port: %d, must be between 1 and 65535", req.DestPort)
	}

	// 端口必须配合协议使用
	if (req.SourcePort > 0 || req.DestPort > 0) && req.Protocol == "" {
		return fmt.Errorf("port specification requires protocol (tcp/udp)")
	}

	return nil
}

func (s *IptablesFilterService) validateIPOrCIDR(ipStr string) error {
	// 检查是否为 CIDR 格式
	if strings.Contains(ipStr, "/") {
		_, _, err := net.ParseCIDR(ipStr)
		if err != nil {
			return fmt.Errorf("invalid CIDR format: %w", err)
		}
		return nil
	}

	// 检查是否为有效 IP 地址
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return fmt.Errorf("invalid IP address format")
	}

	return nil
}

func (s *IptablesFilterService) AddRule(req dto.IptablesFilterRuleOperate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.Chain != client.Chain1PanelInput && req.Chain != client.Chain1PanelOutput {
		return fmt.Errorf("Only %s or %s chains are allowed", client.Chain1PanelInput, client.Chain1PanelOutput)
	}

	// 安全检查：防止无条件 DROP/REJECT
	if (req.Action == "DROP" || req.Action == "REJECT") &&
		req.Protocol == "" && req.SourceIP == "" && req.DestIP == "" &&
		req.SourcePort == 0 && req.DestPort == 0 {
		return fmt.Errorf("Iptables Rule Security Check: unconditional %s rules are not allowed, this may lock you out of the system", req.Action)
	}

	// 验证输入格式
	if err := s.validateRuleInput(&req); err != nil {
		return err
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
	return global.DB.Model(rule).Update("comment", rule.Comment).Error
}

func (s *IptablesFilterService) RemoveRule(id uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

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

type operationRecord struct {
	operation string
	id        uint
	rule      *model.IptablesFilterRule
}

func (s *IptablesFilterService) BatchOperate(req dto.IptablesFilterBatchOperate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 记录成功的操作以便回滚
	var executed []operationRecord

	// 执行批量操作
	for _, ruleReq := range req.Rules {
		if ruleReq.Operation == "add" {
			// 不再调用 AddRule，直接在这里执行（避免重复加锁）
			if ruleReq.Chain != client.Chain1PanelInput && ruleReq.Chain != client.Chain1PanelOutput {
				// 回滚已执行的操作
				s.rollbackBatchOperations(executed)
				return fmt.Errorf("Only %s or %s chains are allowed", client.Chain1PanelInput, client.Chain1PanelOutput)
			}

			// 安全检查
			if (ruleReq.Action == "DROP" || ruleReq.Action == "REJECT") &&
				ruleReq.Protocol == "" && ruleReq.SourceIP == "" && ruleReq.DestIP == "" &&
				ruleReq.SourcePort == 0 && ruleReq.DestPort == 0 {
				s.rollbackBatchOperations(executed)
				return fmt.Errorf("Iptables Rule Security Check: unconditional %s rules are not allowed", ruleReq.Action)
			}

			// 验证输入格式
			if err := s.validateRuleInput(&ruleReq); err != nil {
				s.rollbackBatchOperations(executed)
				return err
			}

			ctx := context.Background()
			maxOrder, err := s.repo.GetMaxRuleOrder(ctx, ruleReq.Chain)
			if err != nil {
				s.rollbackBatchOperations(executed)
				return fmt.Errorf("failed to get max rule order: %w", err)
			}

			rule := &model.IptablesFilterRule{
				Chain:       ruleReq.Chain,
				Protocol:    ruleReq.Protocol,
				SourceIP:    ruleReq.SourceIP,
				SourcePort:  ruleReq.SourcePort,
				DestIP:      ruleReq.DestIP,
				DestPort:    ruleReq.DestPort,
				Action:      ruleReq.Action,
				Description: ruleReq.Description,
				RuleOrder:   maxOrder + 1,
			}

			if err := s.repo.Create(ctx, rule); err != nil {
				s.rollbackBatchOperations(executed)
				return fmt.Errorf("failed to save rule to database: %w", err)
			}

			rule.Comment = fmt.Sprintf("1Panel_FilterRule_%d", rule.ID)
			policy := client.IptablesPolicy{
				Protocol: ruleReq.Protocol,
				SourceIP: ruleReq.SourceIP,
				SrcPort:  ruleReq.SourcePort,
				DestIP:   ruleReq.DestIP,
				DstPort:  ruleReq.DestPort,
				Action:   ruleReq.Action,
				Comment:  rule.Comment,
			}

			if err := s.iptablesClient.AddPolicy(ruleReq.Chain, policy); err != nil {
				_ = s.repo.Delete(ctx, rule.ID)
				s.rollbackBatchOperations(executed)
				return fmt.Errorf("failed to add iptables rule: %w", err)
			}

			if err := global.DB.Model(rule).Update("comment", rule.Comment).Error; err != nil {
				_ = s.iptablesClient.RemovePolicyByComment(ruleReq.Chain, rule.Comment)
				_ = s.repo.Delete(ctx, rule.ID)
				s.rollbackBatchOperations(executed)
				return fmt.Errorf("failed to update comment: %w", err)
			}

			executed = append(executed, operationRecord{operation: "add", id: rule.ID, rule: rule})

		} else if ruleReq.Operation == "remove" {
			ctx := context.Background()
			rule, err := s.repo.GetByID(ctx, ruleReq.ID)
			if err != nil {
				s.rollbackBatchOperations(executed)
				return fmt.Errorf("rule not found: %w", err)
			}

			if rule.Comment != "" {
				if err := s.iptablesClient.RemovePolicyByComment(rule.Chain, rule.Comment); err != nil {
					global.LOG.Warnf("failed to remove iptables rule by comment: %v, trying to delete by policy", err)
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
						s.rollbackBatchOperations(executed)
						return fmt.Errorf("failed to remove iptables rule: %w", err)
					}
				}
			}

			if err := s.repo.Delete(ctx, ruleReq.ID); err != nil {
				s.rollbackBatchOperations(executed)
				return fmt.Errorf("failed to delete rule from database: %w", err)
			}

			executed = append(executed, operationRecord{operation: "remove", id: ruleReq.ID, rule: &rule})
		}
	}
	return nil
}

func (s *IptablesFilterService) rollbackBatchOperations(executed []operationRecord) {
	ctx := context.Background()
	for i := len(executed) - 1; i >= 0; i-- {
		record := executed[i]
		if record.operation == "add" {
			// 回滚添加操作：删除规则
			if record.rule.Comment != "" {
				_ = s.iptablesClient.RemovePolicyByComment(record.rule.Chain, record.rule.Comment)
			}
			_ = s.repo.Delete(ctx, record.id)
			global.LOG.Warnf("Rolled back add operation for rule %d", record.id)
		} else if record.operation == "remove" {
			// 回滚删除操作：重新添加规则
			policy := client.IptablesPolicy{
				Protocol: record.rule.Protocol,
				SourceIP: record.rule.SourceIP,
				SrcPort:  record.rule.SourcePort,
				DestIP:   record.rule.DestIP,
				DstPort:  record.rule.DestPort,
				Action:   record.rule.Action,
				Comment:  record.rule.Comment,
			}
			_ = s.iptablesClient.AddPolicy(record.rule.Chain, policy)
			_ = s.repo.Create(ctx, record.rule)
			global.LOG.Warnf("Rolled back remove operation for rule %d", record.id)
		}
	}
}

func (s *IptablesFilterService) InitChains() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查并创建 1PANEL_INPUT
	exists, err := s.iptablesClient.ChainExists(client.FilterTab, client.Chain1PanelInput)
	if err != nil {
		return fmt.Errorf("failed to check chain %s: %w", client.Chain1PanelInput, err)
	}

	if !exists {
		if err := s.iptablesClient.NewChain(client.FilterTab, client.Chain1PanelInput); err != nil {
			return fmt.Errorf("failed to create chain %s: %w", client.Chain1PanelInput, err)
		}
	} else {
		// 清空已存在的链
		if err := s.iptablesClient.FlushChain(client.FilterTab, client.Chain1PanelInput); err != nil {
			return err
		}
	}

	// 检查并创建 1PANEL_OUTPUT
	exists, err = s.iptablesClient.ChainExists(client.FilterTab, client.Chain1PanelOutput)
	if err != nil {
		return fmt.Errorf("failed to check chain %s: %w", client.Chain1PanelOutput, err)
	}

	if !exists {
		if err := s.iptablesClient.NewChain(client.FilterTab, client.Chain1PanelOutput); err != nil {
			return fmt.Errorf("failed to create chain %s: %w", client.Chain1PanelOutput, err)
		}
	} else {
		if err := s.iptablesClient.FlushChain(client.FilterTab, client.Chain1PanelOutput); err != nil {
			return err
		}
	}

	return nil
}

func (s *IptablesFilterService) ApplyFirewall() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 安全检查
	chains, err := s.iptablesClient.ReadFilter([]string{client.Chain1PanelInput, client.Chain1PanelOutput})
	if err != nil {
		return fmt.Errorf("failed to read filter chains: %w", err)
	}

	// 检查 1PANEL_INPUT 安全性
	if inputChain, ok := chains[client.Chain1PanelInput]; ok {
		item := inputChain.FirstRule
		for item != nil {
			p := item.P
			if (p.Action == "DROP" || p.Action == "REJECT") &&
				p.Protocol == "" && p.SourceIP == "" && p.DestIP == "" &&
				p.SrcPort == 0 && p.DstPort == 0 {
				return fmt.Errorf("Chain %s includes unconditional %s rule, not allowed to apply", client.Chain1PanelInput, p.Action)
			}
			item = item.Next()
		}
	}

	// 检查 1PANEL_OUTPUT 安全性
	if outputChain, ok := chains[client.Chain1PanelOutput]; ok {
		item := outputChain.FirstRule
		for item != nil {
			p := item.P
			if (p.Action == "DROP" || p.Action == "REJECT") &&
				p.Protocol == "" && p.SourceIP == "" && p.DestIP == "" &&
				p.SrcPort == 0 && p.DstPort == 0 {
				return fmt.Errorf("Chain %s includes unconditional %s rule, not allowed to apply", client.Chain1PanelOutput, p.Action)
			}
			item = item.Next()
		}
	}

	if err := s.iptablesClient.Setup1PanelFirewallChains("input"); err != nil {
		return fmt.Errorf("failed to apply %s to %s: %w", client.Chain1PanelInput, client.ChainInput, err)
	}
	global.LOG.Infof("Applied %s to %s chain", client.Chain1PanelInput, client.ChainInput)

	if err := s.iptablesClient.Setup1PanelFirewallChains("output"); err != nil {
		return fmt.Errorf("failed to apply %s to %s: %w", client.Chain1PanelOutput, client.ChainOutput, err)
	}
	global.LOG.Infof("Applied %s to %s chain", client.Chain1PanelOutput, client.ChainOutput)

	return nil
}

func (s *IptablesFilterService) UnloadFirewall() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 从 INPUT 链删除跳转规则
	ruleNum, err := s.iptablesClient.FindRuleNum(client.ChainInput, client.Chain1PanelInput)
	if err == nil {
		if err := s.iptablesClient.RemovePolicyByNum(client.ChainInput, ruleNum); err != nil {
			return fmt.Errorf("failed to unload %s from %s: %w", client.Chain1PanelInput, client.ChainInput, err)
		}
		global.LOG.Infof("Unloaded %s from %s chain", client.Chain1PanelInput, client.ChainInput)
	} else {
		global.LOG.Warnf("%s not found in %s chain: %v", client.Chain1PanelInput, client.ChainInput, err)
	}

	// 从 OUTPUT 链删除跳转规则
	ruleNum, err = s.iptablesClient.FindRuleNum(client.ChainOutput, client.Chain1PanelOutput)
	if err == nil {
		if err := s.iptablesClient.RemovePolicyByNum(client.ChainOutput, ruleNum); err != nil {
			return fmt.Errorf("failed to unload %s from %s: %w", client.Chain1PanelOutput, client.ChainOutput, err)
		}
		global.LOG.Infof("Unloaded %s from %s chain", client.Chain1PanelOutput, client.ChainOutput)
	} else {
		global.LOG.Warnf("%s not found in %s chain: %v", client.Chain1PanelOutput, client.ChainOutput, err)
	}

	return nil
}

func (s *IptablesFilterService) ReloadRules() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 清空自定义链
	if err := s.iptablesClient.FlushChain(client.FilterTab, client.Chain1PanelInput); err != nil {
		return fmt.Errorf("failed to flush chain %s: %w", client.Chain1PanelInput, err)
	}
	if err := s.iptablesClient.FlushChain(client.FilterTab, client.Chain1PanelOutput); err != nil {
		return fmt.Errorf("failed to flush chain %s: %w", client.Chain1PanelOutput, err)
	}

	// 从数据库读取规则
	ctx := context.Background()
	rules, err := s.repo.List(ctx, []string{client.Chain1PanelInput, client.Chain1PanelOutput})
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
