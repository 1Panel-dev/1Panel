package client

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

const (
	PreRoutingChain  = "1PANEL_PREROUTING"
	PostRoutingChain = "1PANEL_POSTROUTING"
	ForwardChain     = "1PANEL_FORWARD"
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
		global.LOG.Errorf("iptables failed, %v", err)
	}
	return stdout, nil
}

func (iptables *Iptables) run(tab, rule string) error {
	if _, err := iptables.out(tab, rule); err != nil {
		return err
	}
	return nil
}

// Run 导出的 run 方法供外部调用
func (iptables *Iptables) Run(tab, rule string) error {
	return iptables.run(tab, rule)
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

// test struct
/*
🧩 规则解释

-A INPUT / -A OUTPUT：追加规则到对应链。

-p tcp / -p udp：指定协议。

--dport / --sport：目标端口 / 源端口。

-d / -s：目标 IP / 源 IP。

-j DROP：丢弃数据包（不回应）。

若改成 -j REJECT，则会主动返回一个拒绝报文（对调试有用）。*
*/
type IptablesPolicy struct {
	Protocol string
	SrcPort  uint16
	DstPort  uint16
	SourceIP string
	DestIP   string
	Action   string // ACCEPT, DROP, REJECT
	Comment  string
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

// IptablesChain
type IptablesChain struct {
	Name          string
	DefaultPolicy string
	FirstRule     *IptablesPolicyChainItem
	LastRule      *IptablesPolicyChainItem
}

type IptablesPolicyChainItem struct {
	next *IptablesPolicyChainItem
	P    IptablesPolicy
}

func (item *IptablesPolicyChainItem) SetNext(next *IptablesPolicyChainItem) {
	item.next = next
}

func (item *IptablesPolicyChainItem) Next() *IptablesPolicyChainItem {
	return item.next
}

func (c *IptablesChain) ParseLine(line string) error {
	cmd := strings.Split(line, " ")

	if cmd[0] == "-P" {
		c.Name = cmd[1]
		c.DefaultPolicy = cmd[2]
		return nil
	}
	if cmd[0] == "-A" {
		if cmd[1] != c.Name {
			return fmt.Errorf("invalid chain name in rule line: %s", line)
		}
		policy := IptablesPolicy{}
		for i := 2; i < len(cmd); i++ {
			switch cmd[i] {
			case "-p":
				i++
				policy.Protocol = cmd[i]
			case "--dport":
				i++
				// parse port
				var port uint16
				fmt.Sscanf(cmd[i], "%d", &port)
				policy.DstPort = port
			case "--sport":
				i++
				var port uint16
				fmt.Sscanf(cmd[i], "%d", &port)
				policy.SrcPort = port
			case "-s":
				i++
				policy.SourceIP = cmd[i]
			case "-d":
				i++
				policy.DestIP = cmd[i]
			case "-j":
				i++
				policy.Action = cmd[i]
			case "-m":
				// skip
				i++
			case "--comment":
				i++
				policy.Comment = strings.Trim(cmd[i], "\"")
			}
		}
		newItem := &IptablesPolicyChainItem{
			P: policy,
		}
		if c.FirstRule == nil {
			c.FirstRule = newItem
			c.LastRule = newItem
		} else {
			current := c.LastRule
			current.SetNext(newItem)
			c.LastRule = newItem
		}
		return nil
	}
	return fmt.Errorf("invalid iptables rule line: %s", line)
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

// 检查链中是否有无条件 DROP/REJECT 规则
func checkChainSafety(chain IptablesChain) error {
	item := chain.FirstRule
	for item != nil {
		policy := item.P
		if policy.Action == "DROP" || policy.Action == "REJECT" {
			// 检查是否为无条件规则(所有字段都为空)
			if policy.Protocol == "" && policy.SrcPort == 0 && policy.DstPort == 0 &&
				policy.SourceIP == "" && policy.DestIP == "" {
				return fmt.Errorf("发现无条件 %s 规则,不允许应用", policy.Action)
			}
		}
		item = item.Next()
	}
	return nil
}

// 执行 iptables 命令
func runIptables(args ...string) error {
	cmd := exec.Command("iptables", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("执行 iptables 失败: %v, 输出: %s", err, string(output))
	}
	return nil
}

// RemovePolicyByComment 按 comment 删除规则
func (iptables *Iptables) RemovePolicyByComment(chain, comment string) error {
	stdout, err := iptables.out(FilterTab, fmt.Sprintf("-S %s", chain))
	if err != nil {
		return err
	}

	// 解析规则找到匹配 comment 的行
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		if strings.Contains(line, comment) && strings.HasPrefix(line, "-A") {
			// 替换 -A 为 -D 执行删除
			deleteRule := strings.Replace(line, "-A", "-D", 1)
			if err := iptables.run(FilterTab, deleteRule); err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("rule with comment '%s' not found in chain %s", comment, chain)
}

// RemovePolicyByNum 按规则编号删除
func (iptables *Iptables) RemovePolicyByNum(chain string, num int) error {
	return iptables.run(FilterTab, fmt.Sprintf("-D %s %d", chain, num))
}

// FindRuleNum 查找规则编号（返回第一个匹配的行号）
func (iptables *Iptables) FindRuleNum(chain, matchStr string) (int, error) {
	stdout, err := iptables.out(FilterTab, fmt.Sprintf("-L %s --line-numbers -n", chain))
	if err != nil {
		return 0, err
	}

	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		if strings.Contains(line, matchStr) {
			// 提取行号（第一列）
			fields := strings.Fields(line)
			if len(fields) > 0 {
				num, err := fmt.Sscanf(fields[0], "%d", new(int))
				if err == nil && num == 1 {
					var ruleNum int
					fmt.Sscanf(fields[0], "%d", &ruleNum)
					return ruleNum, nil
				}
			}
		}
	}
	return 0, fmt.Errorf("rule not found in chain %s", chain)
}

// ChainExists 检查链是否存在
func (iptables *Iptables) ChainExists(tab, chain string) (bool, error) {
	// iptables -t filter -S | grep 'N 1PANEL'
	stdout, err := iptables.out(tab, fmt.Sprintf("-S | grep 'N %s'", chain))
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(stdout) == "" {
		return false, nil
	}
	return true, nil
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
