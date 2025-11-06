package client

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

const (
	PreRoutingChain   = "1PANEL_PREROUTING"
	PostRoutingChain  = "1PANEL_POSTROUTING"
	ForwardChain      = "1PANEL_FORWARD"
	ChainInput        = "INPUT"
	ChainOutput       = "OUTPUT"
	Chain1PanelInput  = "1PANEL_INPUT"
	Chain1PanelOutput = "1PANEL_OUTPUT"
	Chain1PanelBasic  = "1PANEL_BASIC"
)

const (
	ACCEPT = "ACCEPT"
	DROP   = "DROP"
	REJECT = "REJECT"
)

const (
	FilterTab = "filter"
	NatTab    = "nat"
)

const NatChain = "1PANEL"

var (
	natListRegex = regexp.MustCompile(`^(\d+)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)\s+(.+?)(?:\s+(.+?) .+?:(\d{1,5}(?::\d+)?).+?[ :](.+-.+|(?:.+:)?\d{1,5}(?:-\d{1,5})?))?$`)
)

type Iptables struct {
	CmdStr string
}

func NewIptables() (*Iptables, error) {
	iptables := new(Iptables)
	iptables.CmdStr = cmd.SudoHandleCmd()

	return iptables, nil
}

func (iptables *Iptables) out(tab, rule string) (string, error) {
	cmdMgr := cmd.NewCommandMgr(cmd.WithIgnoreExist1(), cmd.WithTimeout(20*time.Second))
	stdout, err := cmdMgr.RunWithStdoutBashCf("%s iptables -t %s %s", iptables.CmdStr, tab, rule)
	if err != nil {
		global.LOG.Errorf("iptables command failed [table=%s, rule=%s]: %v", tab, rule, err)
		return stdout, err
	}
	return stdout, nil
}

func (iptables *Iptables) run(tab, rule string) error {
	if _, err := iptables.out(tab, rule); err != nil {
		return err
	}
	return nil
}

func (iptables *Iptables) Check() error {
	stdout, err := cmd.RunDefaultWithStdoutBashC("cat /proc/sys/net/ipv4/ip_forward")
	if err != nil {
		return fmt.Errorf("check ip_forward failed, %v", err)
	}
	if strings.TrimSpace(stdout) == "0" {
		return fmt.Errorf("ipv4 forward disabled")
	}

	chain, err := iptables.out(NatTab, fmt.Sprintf("-L -n | grep 'Chain %s'", PreRoutingChain))
	if err != nil {
		return fmt.Errorf("failed to check chain: %w", err)
	}
	if strings.TrimSpace(chain) != "" {
		return fmt.Errorf("chain %s already exists", PreRoutingChain)
	}

	return nil
}

func (iptables *Iptables) GetVersion() (string, error) {
	stdout, err := cmd.RunDefaultWithStdoutBashC("iptables --version")
	if err != nil {
		return "", fmt.Errorf("failed to get iptables version: %w", err)
	}
	// 提取版本号，例如 "iptables v1.8.7 (nf_tables)"
	parts := strings.Fields(stdout)
	if len(parts) >= 2 {
		return strings.TrimPrefix(parts[1], "v"), nil
	}
	return strings.TrimSpace(stdout), nil
}

func (iptables *Iptables) NewChain(tab, chain string) error {
	return iptables.run(tab, "-N "+chain)
}

func (iptables *Iptables) AppendChain(tab string, chain, chain1 string) error {
	return iptables.run(tab, fmt.Sprintf("-A %s -j %s", chain, chain1))
}

func (iptables *Iptables) NatList(chain ...string) ([]IptablesNatInfo, error) {
	if len(chain) == 0 {
		chain = append(chain, PreRoutingChain)
	}
	stdout, err := iptables.out(NatTab, fmt.Sprintf("-nvL %s --line-numbers", chain[0]))
	if err != nil {
		return nil, err
	}

	var forwardList []IptablesNatInfo
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimFunc(line, func(r rune) bool {
			return r <= 32
		})
		if natListRegex.MatchString(line) {
			match := natListRegex.FindStringSubmatch(line)
			if !strings.Contains(match[13], ":") {
				match[13] = fmt.Sprintf(":%s", match[13])
			}
			forwardList = append(forwardList, IptablesNatInfo{
				Num:         match[1],
				Target:      match[4],
				Protocol:    match[11],
				InIface:     match[7],
				OutIface:    match[8],
				Opt:         match[6],
				Source:      match[9],
				Destination: match[10],
				SrcPort:     match[12],
				DestPort:    match[13],
			})
		}
	}

	return forwardList, nil
}

func (iptables *Iptables) NatAdd(protocol, srcPort, dest, destPort, iface string, save bool) error {
	if dest != "" && dest != "127.0.0.1" && dest != "localhost" {
		iptablesArg := fmt.Sprintf("-A %s", PreRoutingChain)
		if iface != "" {
			iptablesArg += fmt.Sprintf(" -i %s", iface)
		}
		iptablesArg += fmt.Sprintf(" -p %s --dport %s -j DNAT --to-destination %s:%s", protocol, srcPort, dest, destPort)
		if err := iptables.run(NatTab, iptablesArg); err != nil {
			return err
		}

		if err := iptables.run(NatTab, fmt.Sprintf(
			"-A %s -d %s -p %s --dport %s -j MASQUERADE",
			PostRoutingChain,
			dest,
			protocol,
			destPort,
		)); err != nil {
			return err
		}

		if err := iptables.run(FilterTab, fmt.Sprintf(
			"-A %s -d %s -p %s --dport %s -j ACCEPT",
			ForwardChain,
			dest,
			protocol,
			destPort,
		)); err != nil {
			return err
		}

		if err := iptables.run(FilterTab, fmt.Sprintf(
			"-A %s -s %s -p %s --sport %s -j ACCEPT",
			ForwardChain,
			dest,
			protocol,
			destPort,
		)); err != nil {
			return err
		}
	} else {
		iptablesArg := fmt.Sprintf("-A %s", PreRoutingChain)
		if iface != "" {
			iptablesArg += fmt.Sprintf(" -i %s", iface)
		}
		iptablesArg += fmt.Sprintf(" -p %s --dport %s -j REDIRECT --to-port %s", protocol, srcPort, destPort)
		if err := iptables.run(NatTab, iptablesArg); err != nil {
			return err
		}
	}

	if save {
		return global.DB.Save(&model.Forward{
			Protocol:   protocol,
			Port:       srcPort,
			TargetIP:   dest,
			TargetPort: destPort,
			Interface:  iface,
		}).Error
	}
	return nil
}

func (iptables *Iptables) NatRemove(num string, protocol, srcPort, dest, destPort, iface string) error {
	if err := iptables.run(NatTab, fmt.Sprintf("-D %s %s", PreRoutingChain, num)); err != nil {
		return err
	}

	if dest != "" && dest != "127.0.0.1" && dest != "localhost" {
		if err := iptables.run(NatTab, fmt.Sprintf(
			"-D %s -d %s -p %s --dport %s -j MASQUERADE",
			PostRoutingChain,
			dest,
			protocol,
			destPort,
		)); err != nil {
			return err
		}

		if err := iptables.run(FilterTab, fmt.Sprintf(
			"-D %s -d %s -p %s --dport %s -j ACCEPT",
			ForwardChain,
			dest,
			protocol,
			destPort,
		)); err != nil {
			return err
		}

		if err := iptables.run(FilterTab, fmt.Sprintf(
			"-D %s -s %s -p %s --sport %s -j ACCEPT",
			ForwardChain,
			dest,
			protocol,
			destPort,
		)); err != nil {
			return err
		}
	}

	global.DB.Where(
		"protocol = ? AND port = ? AND target_ip = ? AND target_port = ? AND (interface = ? OR (interface IS NULL AND ? = ''))",
		protocol,
		srcPort,
		dest,
		destPort,
		iface,
		iface,
	).Delete(&model.Forward{})
	return nil
}

func (iptables *Iptables) AddPolicy(chain string, policy IptablesPolicy) error {
	iptablesArg := fmt.Sprintf("-A %s", chain)
	if policy.Protocol != "" {
		iptablesArg += fmt.Sprintf(" -p %s", policy.Protocol)
	}
	if policy.SrcPort != 0 {
		iptablesArg += fmt.Sprintf(" --sport %d", policy.SrcPort)
	}
	if policy.DstPort != 0 {
		iptablesArg += fmt.Sprintf(" --dport %d", policy.DstPort)
	}
	if policy.SourceIP != "" {
		iptablesArg += fmt.Sprintf(" -s %s", policy.SourceIP)
	}
	if policy.DestIP != "" {
		iptablesArg += fmt.Sprintf(" -d %s", policy.DestIP)
	}
	iptablesArg += fmt.Sprintf(" -j %s", policy.Action)
	iptablesArg += fmt.Sprintf(" -m comment --comment \"%s\"", policy.Comment)

	return iptables.run(FilterTab, iptablesArg)
}

func (iptables *Iptables) Reload() error {
	if err := iptables.run(NatTab, "-F "+PreRoutingChain); err != nil {
		return err
	}
	if err := iptables.run(NatTab, "-F "+PostRoutingChain); err != nil {
		return err
	}
	if err := iptables.run(FilterTab, "-F "+ForwardChain); err != nil {
		return err
	}

	var rules []model.Forward
	global.DB.Find(&rules)
	for _, forward := range rules {
		if err := iptables.NatAdd(forward.Protocol, forward.Port, forward.TargetIP, forward.TargetPort, forward.Interface, false); err != nil {
			return err
		}
	}
	return nil
}

// iptables filter 解析
func (iptables *Iptables) ReadFilter(chainName []string) (map[string]IptablesChain, error) {
	// iptables -S
	cmdMgr := cmd.NewCommandMgr(cmd.WithIgnoreExist1(), cmd.WithTimeout(20*time.Second))
	stdout, err := cmdMgr.RunWithStdoutBashCf("%s iptables -S", iptables.CmdStr)
	if err != nil {
		global.LOG.Errorf("iptables failed, %v", err)
	}
	// 解析内容
	chains := make(map[string]IptablesChain)
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-P") || strings.HasPrefix(line, "-A") {
			parts := strings.SplitN(line, " ", 3)
			chain := parts[1]
			if len(chainName) > 0 {
				found := false
				for _, name := range chainName {
					if chain == name {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			if _, exists := chains[chain]; !exists {
				chains[chain] = IptablesChain{
					Name: chain,
				}
			}
			chainStruct := chains[chain]
			if err := chainStruct.ParseLine(line); err != nil {
				return nil, err
			}
			chains[chain] = chainStruct
		}
	}
	return chains, nil
}

// RemovePolicyByComment 按 comment 删除规则
func (iptables *Iptables) RemovePolicyByComment(chain, comment string) error {
	stdout, err := iptables.out(FilterTab, fmt.Sprintf("-S %s", chain))
	if err != nil {
		return fmt.Errorf("failed to list rules in chain %s: %w", chain, err)
	}

	// 解析规则找到匹配 comment 的行
	lines := strings.Split(stdout, "\n")
	found := false
	for _, line := range lines {
		if strings.Contains(line, comment) && strings.HasPrefix(line, "-A") {
			// 替换 -A 为 -D 执行删除
			deleteRule := strings.Replace(line, "-A", "-D", 1)
			if err := iptables.run(FilterTab, deleteRule); err != nil {
				return fmt.Errorf("failed to delete rule: %w", err)
			}
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("rule with comment '%s' not found in chain %s", comment, chain)
	}
	return nil
}

// RemovePolicyByNum 按规则编号删除
func (iptables *Iptables) RemovePolicyByNum(chain string, num int) error {
	return iptables.run(FilterTab, fmt.Sprintf("-D %s %d", chain, num))
}

// FindRuleNum 查找规则编号（返回第一个匹配的行号）
func (iptables *Iptables) FindRuleNum(chain, matchStr string) (int, error) {
	stdout, err := iptables.out(FilterTab, fmt.Sprintf("-L %s --line-numbers -n", chain))
	if err != nil {
		return 0, fmt.Errorf("failed to list rules in chain %s: %w", chain, err)
	}

	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		if strings.Contains(line, matchStr) {
			// 提取行号（第一列）
			fields := strings.Fields(line)
			if len(fields) > 0 {
				var ruleNum int
				n, err := fmt.Sscanf(fields[0], "%d", &ruleNum)
				if err == nil && n == 1 && ruleNum > 0 {
					return ruleNum, nil
				}
			}
		}
	}
	return 0, fmt.Errorf("rule matching '%s' not found in chain %s", matchStr, chain)
}

// ChainExists 检查链是否存在
func (iptables *Iptables) ChainExists(tab, chain string) (bool, error) {
	// iptables -t filter -S | grep 'N 1PANEL'
	stdout, err := iptables.out(tab, fmt.Sprintf("-S | grep -w 'N %s'", chain))
	// grep 未找到会返回错误，这是正常情况
	if err != nil && strings.TrimSpace(stdout) == "" {
		return false, nil
	}
	if strings.TrimSpace(stdout) == "" {
		return false, nil
	}
	return true, nil
}

// ensureChain creates a chain if it doesn't exist
func (iptables *Iptables) ensureChain(tab, chain string) error {
	exists, err := iptables.ChainExists(tab, chain)
	if err != nil {
		return fmt.Errorf("failed to check chain %s: %w", chain, err)
	}
	if !exists {
		if err := iptables.NewChain(tab, chain); err != nil {
			return fmt.Errorf("failed to create chain %s: %w", chain, err)
		}
	}
	return nil
}

// ensureJumpRule adds a jump rule to targetChain at specified position if it doesn't exist
func (iptables *Iptables) ensureJumpRule(chain, targetChain string, position int) error {
	if !iptables.CheckPolicyExists(chain, fmt.Sprintf("-j %s", targetChain)) {
		return iptables.run(FilterTab, fmt.Sprintf("-I %s %d -j %s", chain, position, targetChain))
	}
	return nil
}

// ensureChainWithJump creates a custom chain and adds jump rule from parent chain if not exists
func (iptables *Iptables) ensureChainWithJump(tab, parentChain, customChain string) error {
	// Ensure custom chain exists
	exists, err := iptables.ChainExists(tab, customChain)
	if err != nil {
		return fmt.Errorf("failed to check chain %s: %w", customChain, err)
	}
	if !exists {
		if err := iptables.NewChain(tab, customChain); err != nil {
			return fmt.Errorf("failed to create chain %s: %w", customChain, err)
		}
		// Add jump rule from parent chain
		if err := iptables.AppendChain(tab, parentChain, customChain); err != nil {
			return fmt.Errorf("failed to append %s to %s: %w", customChain, parentChain, err)
		}
	}
	return nil
}

// FlushChain 清空链（保留链结构）
func (iptables *Iptables) FlushChain(tab, chain string) error {
	return iptables.run(tab, fmt.Sprintf("-F %s", chain))
}

// DeletePolicy 删除指定策略规则
func (iptables *Iptables) DeletePolicy(chain string, policy IptablesPolicy) error {
	iptablesArg := fmt.Sprintf("-D %s", chain)
	if policy.Protocol != "" {
		iptablesArg += fmt.Sprintf(" -p %s", policy.Protocol)
	}
	if policy.SrcPort != 0 {
		iptablesArg += fmt.Sprintf(" --sport %d", policy.SrcPort)
	}
	if policy.DstPort != 0 {
		iptablesArg += fmt.Sprintf(" --dport %d", policy.DstPort)
	}
	if policy.SourceIP != "" {
		iptablesArg += fmt.Sprintf(" -s %s", policy.SourceIP)
	}
	if policy.DestIP != "" {
		iptablesArg += fmt.Sprintf(" -d %s", policy.DestIP)
	}
	iptablesArg += fmt.Sprintf(" -j %s", policy.Action)
	if policy.Comment != "" {
		iptablesArg += fmt.Sprintf(" -m comment --comment \"%s\"", policy.Comment)
	}

	return iptables.run(FilterTab, iptablesArg)
}

// CheckPolicyExists 检查策略是否存在
func (iptables *Iptables) CheckPolicyExists(chain string, policyStr string) bool {
	line, err := iptables.FindRuleNum(chain, policyStr)
	if err != nil {
		return false
	}
	if line > 0 {
		return true
	}
	return false
}

const (
	establishedRule        = "-m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT -m comment --comment \"ESTABLISHED Whitelist\""
	ioRuleIn               = "-i lo -j ACCEPT -m comment --comment \"Loopback Whitelist\""
	ioRuleOut              = "-o lo -j ACCEPT -m comment --comment \"Loopback Whitelist\""
	establishedRuleComment = "ESTABLISHED Whitelist"
	ioRuleInComment        = "Loopback Whitelist"
	ioRuleOutComment       = "Loopback Whitelist"
)

func (iptables *Iptables) setupEstablishedRules(direction string) {
	if direction == "input" {
		if !iptables.CheckPolicyExists("INPUT", ioRuleInComment) {
			iptables.run(FilterTab, fmt.Sprintf("-I INPUT 1 %s", ioRuleIn))
		}

		if !iptables.CheckPolicyExists("INPUT", establishedRuleComment) {
			iptables.run(FilterTab, fmt.Sprintf("-I INPUT 2 %s", establishedRule))
		}

		// 末尾加上 DROP
		iptables.run(FilterTab, fmt.Sprintf("-A INPUT -j DROP -m comment --comment \"Default DROP\""))
	}

	if direction == "output" {
		if !iptables.CheckPolicyExists("OUTPUT", ioRuleOutComment) {
			iptables.run(FilterTab, fmt.Sprintf("-I OUTPUT 1 %s", ioRuleOut))
		}
		if !iptables.CheckPolicyExists("OUTPUT", establishedRuleComment) {
			iptables.run(FilterTab, fmt.Sprintf("-I OUTPUT 2 %s", establishedRule))
		}
	}
}

// Init1PanelBasicChains 初始化 1PANEL_BASIC 链 加入放行 22 端口

func (iptables *Iptables) Init1PanelBasicChains() error {
	// Ensure 1PANEL_BASIC chain exists
	if err := iptables.ensureChain(FilterTab, Chain1PanelBasic); err != nil {
		return err
	}

	// 检查是否已经存在 22 端口放行规则
	chains, err := iptables.ReadFilter([]string{Chain1PanelBasic})
	if err != nil {
		global.LOG.Warnf("failed to read 1PANEL_BASIC chain: %v, will attempt to add SSH rule", err)
	} else {
		basicChain, exists := chains[Chain1PanelBasic]
		if exists {
			// 遍历规则检查是否已有 22 端口的 ACCEPT 规则
			for rule := basicChain.FirstRule; rule != nil; rule = rule.Next() {
				policy := rule.P
				if policy.Protocol == "tcp" && policy.DstPort == 22 && policy.Action == ACCEPT {
					global.LOG.Infof("SSH port 22 rule already exists in %s chain", Chain1PanelBasic)
					return nil
				}
			}
		}
	}

	// Allow SSH (port 22) in 1PANEL_BASIC chain
	sshPolicy := IptablesPolicy{
		Protocol: "tcp",
		DstPort:  22,
		Action:   ACCEPT,
		Comment:  "Allow SSH",
	}
	panelPort, err := strconv.Atoi(global.CONF.Base.Port)
	if err != nil {
		return fmt.Errorf("invalid 1Panel port: %w", err)
	}

	PanelPolicy := IptablesPolicy{
		Protocol: "tcp",
		DstPort:  uint16(panelPort),
		Action:   ACCEPT,
		Comment:  "Allow 1Panel",
	}
	if err := iptables.AddPolicy(Chain1PanelBasic, sshPolicy); err != nil {
		return fmt.Errorf("failed to add SSH allow rule to %s: %w", Chain1PanelBasic, err)
	}
	if err := iptables.AddPolicy(Chain1PanelBasic, PanelPolicy); err != nil {
		return fmt.Errorf("failed to add 1Panel allow rule to %s: %w", Chain1PanelBasic, err)
	}
	global.LOG.Infof("SSH port 22/1Paenl rule added to %s chain", Chain1PanelBasic)

	return nil
}

func (iptables *Iptables) Setup1PanelFirewallChains(direction string) error {
	iptables.setupEstablishedRules(direction)

	if direction == "input" {
		// Ensure custom chains exist
		if err := iptables.ensureChain(FilterTab, Chain1PanelInput); err != nil {
			return err
		}
		if err := iptables.ensureChain(FilterTab, Chain1PanelBasic); err != nil {
			return err
		}

		// Add jump rules to INPUT chain
		if err := iptables.ensureJumpRule(ChainInput, Chain1PanelInput, 3); err != nil {
			return err
		}
		if err := iptables.ensureJumpRule(ChainInput, Chain1PanelBasic, 4); err != nil {
			return err
		}
	} else if direction == "output" {
		// Ensure custom chain exists
		if err := iptables.ensureChain(FilterTab, Chain1PanelOutput); err != nil {
			return err
		}

		// Add jump rule to OUTPUT chain
		if err := iptables.ensureJumpRule(ChainOutput, Chain1PanelOutput, 3); err != nil {
			return err
		}
	}
	return nil
}

// Teardown1PanelFirewallChains 撤销防火墙链设置
func (iptables *Iptables) Teardown1PanelFirewallChains(direction string) error {
	if direction == "input" {
		if err := iptables.RemovePolicyByComment(ChainInput, "Default DROP"); err != nil {
			global.LOG.Warnf("failed to remove default DROP rule from INPUT: %v", err)
		}

		err := iptables.run(FilterTab, fmt.Sprintf("-D %s -j %s", ChainInput, Chain1PanelBasic))
		if err != nil {
			global.LOG.Warnf("failed to remove jump rule from INPUT to 1PANEL_BASIC: %v", err)
		}

		err = iptables.run(FilterTab, fmt.Sprintf("-D %s -j %s", ChainInput, Chain1PanelInput))
		if err != nil {
			global.LOG.Warnf("failed to remove jump rule from INPUT to 1PANEL_INPUT: %v", err)
		}

	} else if direction == "output" {
		err := iptables.run(FilterTab, fmt.Sprintf("-D %s -j %s", ChainOutput, Chain1PanelOutput))
		if err != nil {
			global.LOG.Warnf("failed to remove jump rule from OUTPUT to 1PANEL_OUTPUT: %v", err)
		}
	}

	global.LOG.Infof("1Panel firewall chains teardown completed for %s direction", direction)
	return nil
}

// Check1PanelChainsApplied 检查自定义链是否已按正确顺序应用到主链
func (iptables *Iptables) Check1PanelInputChainsApplied() (bool, error) {
	chains, err := iptables.ReadFilter([]string{ChainInput})
	if err != nil {
		return false, err
	}

	inputChain, exists := chains[ChainInput]
	if !exists {
		return false, fmt.Errorf("INPUT chain not found")
	}

	// 需要找到跳转到 1PANEL_INPUT 和 1PANEL_BASIC 的规则
	// 并且 1PANEL_INPUT 应该在 1PANEL_BASIC 之前
	found1PanelInput := false
	found1PanelBasic := false
	position1PanelInput := -1
	position1PanelBasic := -1
	currentPosition := 0

	for rule := inputChain.FirstRule; rule != nil; rule = rule.Next() {
		policy := rule.P

		// 检查是否是跳转到 1PANEL_INPUT 的规则
		if policy.Action == Chain1PanelInput {
			found1PanelInput = true
			position1PanelInput = currentPosition
		}

		// 检查是否是跳转到 1PANEL_BASIC 的规则
		if policy.Action == Chain1PanelBasic {
			found1PanelBasic = true
			position1PanelBasic = currentPosition
		}

		currentPosition++
	}

	// 检查两个跳转规则是否都存在，且 1PANEL_INPUT 在 1PANEL_BASIC 之前
	if !found1PanelInput || !found1PanelBasic {
		global.LOG.Debugf("1Panel chains not fully applied: 1PANEL_INPUT=%v, 1PANEL_BASIC=%v",
			found1PanelInput, found1PanelBasic)
		return false, fmt.Errorf("1Panel chains not fully applied")
	}

	if position1PanelInput >= position1PanelBasic {
		global.LOG.Debugf("1Panel chains order incorrect: 1PANEL_INPUT at %d, 1PANEL_BASIC at %d",
			position1PanelInput, position1PanelBasic)
		return false, fmt.Errorf("1Panel chains order incorrect")
	}

	return true, nil

}

func (iptables *Iptables) SaftyCheck(ports []int) bool {
	chains, err := iptables.ReadFilter([]string{ChainInput, Chain1PanelInput, Chain1PanelBasic})
	if err != nil {
		global.LOG.Errorf("failed to read filter chains: %v", err)
		return false
	}

	inputChain, exists := chains[ChainInput]
	if !exists {
		global.LOG.Errorf("INPUT chain not found")
		return false
	}

	inputDefaultAccept := inputChain.DefaultPolicy == ACCEPT

	for _, port := range ports {
		// 检查所有相关链中是否有规则会阻止该端口
		if iptables.isPortBlocked(port, chains, inputDefaultAccept) {
			global.LOG.Warnf("port %d is blocked by firewall rules", port)
			return false
		}
	}

	return true
}

// isPortBlocked 检查指定端口是否被防火墙规则阻止
func (iptables *Iptables) isPortBlocked(port int, chains map[string]IptablesChain, inputDefaultAccept bool) bool {
	chainOrder := []string{ChainInput, Chain1PanelInput, Chain1PanelBasic}
	for _, chainName := range chainOrder {
		chain, exists := chains[chainName]
		if !exists {
			continue
		}
		for rule := chain.FirstRule; rule != nil; rule = rule.Next() {
			policy := rule.P
			portMatches := policy.DstPort == 0 || policy.DstPort == uint16(port)
			if portMatches {
				if policy.Action == DROP || policy.Action == REJECT {
					return true
				}
				if policy.Action == ACCEPT {
					return false
				}
			}
		}
	}
	return !inputDefaultAccept
}
