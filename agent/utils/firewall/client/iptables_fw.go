package client

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

const (
	Chain1PanelInput  = "1PFW-INPUT"
	Chain1PanelOutput = "1PFW-OUTPUT"
)

var (
	// 匹配 iptables 规则的正则表达式
	filterListRegex = regexp.MustCompile(`^(\d+)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)(?:\s+(.+?))?(?:\s+(.+?))?`)
)

type IptablesFw struct {
	iptables *Iptables
}

func NewIptablesFw() (*IptablesFw, error) {
	iptables, err := NewIptables()
	if err != nil {
		return nil, err
	}
	return &IptablesFw{
		iptables: iptables,
	}, nil
}

func (f *IptablesFw) Name() string {
	return "iptables"
}

func (f *IptablesFw) Status() (bool, error) {
	// 检查自定义链是否存在来判断防火墙是否启用
	exists, err := f.iptables.ChainExists(FilterTab, Chain1PanelInput)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (f *IptablesFw) Version() (string, error) {
	stdout, err := cmd.RunDefaultWithStdoutBashC("iptables --version")
	if err != nil {
		return "", fmt.Errorf("load the firewall version failed, %v", err)
	}
	// 提取版本号，例如 "iptables v1.8.7 (nf_tables)"
	parts := strings.Fields(stdout)
	if len(parts) >= 2 {
		return strings.TrimPrefix(parts[1], "v"), nil
	}
	return strings.TrimSpace(stdout), nil
}

func (f *IptablesFw) Start() error {
	// 创建自定义链
	if err := f.iptables.NewChain(FilterTab, Chain1PanelInput); err != nil {
		global.LOG.Debugf("chain %s may already exist: %v", Chain1PanelInput, err)
	}
	if err := f.iptables.NewChain(FilterTab, Chain1PanelOutput); err != nil {
		global.LOG.Debugf("chain %s may already exist: %v", Chain1PanelOutput, err)
	}

	// 将自定义链附加到 INPUT 和 OUTPUT 链
	if err := f.ensureChainAttached("INPUT", Chain1PanelInput); err != nil {
		return fmt.Errorf("failed to attach %s to INPUT: %v", Chain1PanelInput, err)
	}
	if err := f.ensureChainAttached("OUTPUT", Chain1PanelOutput); err != nil {
		return fmt.Errorf("failed to attach %s to OUTPUT: %v", Chain1PanelOutput, err)
	}

	// 重载规则（从数据库恢复）
	return f.Reload()
}

func (f *IptablesFw) Stop() error {
	// 从 INPUT 和 OUTPUT 链中移除跳转规则
	_ = f.iptables.Run(FilterTab, fmt.Sprintf("-D INPUT -j %s", Chain1PanelInput))
	_ = f.iptables.Run(FilterTab, fmt.Sprintf("-D OUTPUT -j %s", Chain1PanelOutput))

	// 清空并删除自定义链
	_ = f.iptables.FlushChain(FilterTab, Chain1PanelInput)
	_ = f.iptables.FlushChain(FilterTab, Chain1PanelOutput)
	_ = f.iptables.Run(FilterTab, fmt.Sprintf("-X %s", Chain1PanelInput))
	_ = f.iptables.Run(FilterTab, fmt.Sprintf("-X %s", Chain1PanelOutput))

	return nil
}

func (f *IptablesFw) Restart() error {
	if err := f.Stop(); err != nil {
		return err
	}
	return f.Start()
}

func (f *IptablesFw) Reload() error {
	// 清空链但保留链结构
	if err := f.iptables.FlushChain(FilterTab, Chain1PanelInput); err != nil {
		return err
	}
	if err := f.iptables.FlushChain(FilterTab, Chain1PanelOutput); err != nil {
		return err
	}

	// 从数据库加载规则并恢复
	var rules []model.IptablesRule
	if err := global.DB.Find(&rules).Error; err != nil {
		return fmt.Errorf("failed to load rules from database: %v", err)
	}

	for _, rule := range rules {
		fireInfo := FireInfo{
			Protocol: rule.Protocol,
			Port:     rule.Port,
			Strategy: rule.Strategy,
			Address:  rule.Address,
			Family:   rule.Family,
		}
		if rule.RuleType == "port" {
			if err := f.Port(fireInfo, "add"); err != nil {
				global.LOG.Errorf("failed to restore port rule %+v: %v", rule, err)
			}
		} else if rule.RuleType == "address" {
			if err := f.RichRules(fireInfo, "add"); err != nil {
				global.LOG.Errorf("failed to restore address rule %+v: %v", rule, err)
			}
		}
	}

	return nil
}

func (f *IptablesFw) ListPort() ([]FireInfo, error) {
	return f.listRules(Chain1PanelInput, "port")
}

func (f *IptablesFw) ListAddress() ([]FireInfo, error) {
	return f.listRules(Chain1PanelInput, "address")
}

func (f *IptablesFw) ListForward() ([]FireInfo, error) {
	// 复用 ufw 的转发实现
	if err := f.EnableForward(); err != nil {
		global.LOG.Errorf("init port forward failed, err: %v", err)
	}

	rules, err := f.iptables.NatList()
	if err != nil {
		return nil, err
	}

	var list []FireInfo
	for _, rule := range rules {
		dest := strings.Split(rule.DestPort, ":")
		if len(dest) < 2 {
			continue
		}
		if len(dest[0]) == 0 {
			dest[0] = "127.0.0.1"
		}
		list = append(list, FireInfo{
			Num:        rule.Num,
			Protocol:   rule.Protocol,
			Interface:  rule.InIface,
			Port:       rule.SrcPort,
			TargetIP:   dest[0],
			TargetPort: dest[1],
		})
	}
	return list, nil
}

func (f *IptablesFw) Port(port FireInfo, operation string) error {
	if cmd.CheckIllegal(operation, port.Protocol, port.Port) {
		return buserr.New("ErrCmdIllegal")
	}

	// 转换策略
	strategy := f.convertStrategy(port.Strategy)
	if strategy == "" {
		return fmt.Errorf("unsupported strategy %s", port.Strategy)
	}

	// 构建规则的唯一标识
	comment := fmt.Sprintf("1Panel_Port_%s_%s_%s", port.Protocol, port.Port, strategy)

	if operation == "add" {
		// 添加到 INPUT 链
		policy := IptablesPolicy{
			Protocol: port.Protocol,
			DstPort:  f.parsePort(port.Port),
			Action:   strategy,
			Comment:  comment,
		}
		if err := f.iptables.AddPolicy(Chain1PanelInput, policy); err != nil {
			return fmt.Errorf("add port rule to INPUT failed: %v", err)
		}

		// 保存到数据库
		rule := model.IptablesRule{
			RuleType: "port",
			Protocol: port.Protocol,
			Port:     port.Port,
			Strategy: port.Strategy,
			Family:   "ipv4",
		}
		if err := global.DB.Create(&rule).Error; err != nil {
			return fmt.Errorf("save rule to database failed: %v", err)
		}
	} else if operation == "remove" {
		// 从 INPUT 链删除
		if err := f.iptables.RemovePolicyByComment(Chain1PanelInput, comment); err != nil {
			return fmt.Errorf("remove port rule from INPUT failed: %v", err)
		}

		// 从数据库删除
		global.DB.Where("rule_type = ? AND protocol = ? AND port = ? AND strategy = ?",
			"port", port.Protocol, port.Port, port.Strategy).Delete(&model.IptablesRule{})
	}

	return nil
}

func (f *IptablesFw) RichRules(rule FireInfo, operation string) error {
	if cmd.CheckIllegal(operation, rule.Address, rule.Protocol, rule.Port) {
		return buserr.New("ErrCmdIllegal")
	}

	strategy := f.convertStrategy(rule.Strategy)
	if strategy == "" {
		return fmt.Errorf("unsupported strategy %s", rule.Strategy)
	}

	// 构建规则的唯一标识
	comment := fmt.Sprintf("1Panel_Addr_%s_%s_%s_%s", rule.Address, rule.Protocol, rule.Port, strategy)

	if operation == "add" {
		policy := IptablesPolicy{
			Protocol: rule.Protocol,
			SourceIP: rule.Address,
			Action:   strategy,
			Comment:  comment,
		}
		if len(rule.Port) > 0 {
			policy.DstPort = f.parsePort(rule.Port)
		}

		if err := f.iptables.AddPolicy(Chain1PanelInput, policy); err != nil {
			return fmt.Errorf("add address rule failed: %v", err)
		}

		// 保存到数据库
		dbRule := model.IptablesRule{
			RuleType: "address",
			Protocol: rule.Protocol,
			Port:     rule.Port,
			Strategy: rule.Strategy,
			Address:  rule.Address,
			Family:   rule.Family,
		}
		if dbRule.Family == "" {
			dbRule.Family = "ipv4"
		}
		if err := global.DB.Create(&dbRule).Error; err != nil {
			return fmt.Errorf("save rule to database failed: %v", err)
		}
	} else if operation == "remove" {
		if err := f.iptables.RemovePolicyByComment(Chain1PanelInput, comment); err != nil {
			return fmt.Errorf("remove address rule failed: %v", err)
		}

		// 从数据库删除
		global.DB.Where("rule_type = ? AND address = ? AND protocol = ? AND port = ? AND strategy = ?",
			"address", rule.Address, rule.Protocol, rule.Port, rule.Strategy).Delete(&model.IptablesRule{})
	}

	return nil
}

func (f *IptablesFw) PortForward(info Forward, operation string) error {
	// 直接复用 iptables 的 NAT 功能
	if operation == "add" {
		err := f.iptables.NatAdd(info.Protocol, info.Port, info.TargetIP, info.TargetPort, info.Interface, true)
		if err != nil {
			return fmt.Errorf("add port forward failed: %v", err)
		}
	} else if operation == "remove" {
		err := f.iptables.NatRemove(info.Num, info.Protocol, info.Port, info.TargetIP, info.TargetPort, info.Interface)
		if err != nil {
			return fmt.Errorf("remove port forward failed: %v", err)
		}
	}
	return nil
}

func (f *IptablesFw) EnableForward() error {
	// 复用 ufw 的转发初始化逻辑
	if err := f.iptables.Check(); err != nil {
		return err
	}

	_ = f.iptables.NewChain(NatTab, PreRoutingChain)
	_ = f.iptables.NewChain(NatTab, PostRoutingChain)
	_ = f.iptables.NewChain(FilterTab, ForwardChain)

	if err := f.enableForwardChain(); err != nil {
		return err
	}
	return f.iptables.Reload()
}

// 辅助方法：确保链已附加到目标链
func (f *IptablesFw) ensureChainAttached(targetChain, customChain string) error {
	// 检查是否已经附加
	rules, err := f.iptables.NatList(targetChain)
	if err == nil {
		for _, rule := range rules {
			if rule.Target == customChain {
				return nil
			}
		}
	}

	// 附加链
	return f.iptables.AppendChain(FilterTab, targetChain, customChain)
}

// 辅助方法：启用转发链
func (f *IptablesFw) enableForwardChain() error {
	rules, err := f.iptables.NatList("PREROUTING")
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if rule.Target == PreRoutingChain {
			return nil
		}
	}

	_ = f.iptables.AppendChain(NatTab, "PREROUTING", PreRoutingChain)
	_ = f.iptables.AppendChain(NatTab, "POSTROUTING", PostRoutingChain)
	_ = f.iptables.AppendChain(FilterTab, "FORWARD", ForwardChain)
	return nil
}

// 辅助方法：列出规则
func (f *IptablesFw) listRules(chain, ruleType string) ([]FireInfo, error) {
	stdout, err := cmd.RunDefaultWithStdoutBashCf("%s iptables -t filter -L %s -n -v --line-numbers", f.iptables.CmdStr, chain)
	if err != nil {
		return nil, fmt.Errorf("list rules failed: %v", err)
	}

	var datas []FireInfo
	lines := strings.Split(stdout, "\n")
	for i, line := range lines {
		// 跳过表头
		if i < 2 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}

		// 解析规则字段
		// num pkts bytes target prot opt in out source destination [extra]
		num := fields[0]
		target := fields[3]
		protocol := fields[4]
		source := fields[8]

		// 提取端口信息（从 extra 字段）
		var port string
		if len(fields) > 10 {
			for j := 10; j < len(fields); j++ {
				if strings.HasPrefix(fields[j], "dpt:") {
					port = strings.TrimPrefix(fields[j], "dpt:")
				} else if strings.HasPrefix(fields[j], "spt:") {
					port = strings.TrimPrefix(fields[j], "spt:")
				}
			}
		}

		// 根据类型过滤
		if ruleType == "port" {
			// 端口规则：有端口信息，源地址为 0.0.0.0/0
			if len(port) > 0 && (source == "0.0.0.0/0" || source == "anywhere") {
				datas = append(datas, FireInfo{
					Num:      num,
					Protocol: protocol,
					Port:     port,
					Strategy: f.convertStrategyReverse(target),
					Family:   "ipv4",
				})
			}
		} else if ruleType == "address" {
			// 地址规则：有源地址限制
			if source != "0.0.0.0/0" && source != "anywhere" {
				datas = append(datas, FireInfo{
					Num:      num,
					Protocol: protocol,
					Port:     port,
					Address:  source,
					Strategy: f.convertStrategyReverse(target),
					Family:   "ipv4",
				})
			}
		}
	}

	return datas, nil
}

// 辅助方法：转换策略名称 (accept/drop -> ACCEPT/DROP)
func (f *IptablesFw) convertStrategy(strategy string) string {
	switch strategy {
	case "accept":
		return "ACCEPT"
	case "drop":
		return "DROP"
	case "reject":
		return "REJECT"
	default:
		return ""
	}
}

// 辅助方法：反向转换策略名称 (ACCEPT/DROP -> accept/drop)
func (f *IptablesFw) convertStrategyReverse(strategy string) string {
	switch strategy {
	case "ACCEPT":
		return "accept"
	case "DROP":
		return "drop"
	case "REJECT":
		return "reject"
	default:
		return strategy
	}
}

// 辅助方法：解析端口字符串为 uint16
func (f *IptablesFw) parsePort(portStr string) uint16 {
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0
	}
	return uint16(port)
}
