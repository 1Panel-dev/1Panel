package iptables

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type FilterRules struct {
	ID          uint   `json:"id"`
	Chain       string `json:"chain"`
	Protocol    string `json:"protocol"`
	SrcPort     string `json:"srcPort"`
	DstPort     string `json:"dstPort"`
	SrcIP       string `json:"srcIP"`
	DstIP       string `json:"dstIP"`
	Strategy    string `json:"strategy"`
	Description string `json:"description"`
}

func AddFilterRule(chain string, policy FilterRules) error {
	if err := validateRuleSafety(policy, chain); err != nil {
		return err
	}
	args := []string{"-A", chain}
	if policy.Protocol != "" {
		args = append(args, "-p", policy.Protocol)
	}
	if len(policy.SrcPort) != 0 {
		args = append(args, "--sport", policy.SrcPort)
	}
	if len(policy.DstPort) != 0 {
		args = append(args, "--dport", policy.DstPort)
	}
	if policy.SrcIP != "" {
		args = append(args, "-s", policy.SrcIP)
	}
	if policy.DstIP != "" {
		args = append(args, "-d", policy.DstIP)
	}
	args = append(args, "-j", policy.Strategy)

	return Run(FilterTab, args...)
}

func DeleteFilterRule(chain string, policy FilterRules) error {
	args := []string{"-D", chain}
	if policy.Protocol != "" {
		args = append(args, "-p", policy.Protocol)
	}
	if len(policy.SrcPort) != 0 {
		args = append(args, "--sport", policy.SrcPort)
	}
	if len(policy.DstPort) != 0 {
		args = append(args, "--dport", policy.DstPort)
	}
	if policy.SrcIP != "" {
		args = append(args, "-s", policy.SrcIP)
	}
	if policy.DstIP != "" {
		args = append(args, "-d", policy.DstIP)
	}
	args = append(args, "-j", policy.Strategy)

	return Run(FilterTab, args...)
}

func ReadFilterRulesByChain(chain string) ([]FilterRules, error) {
	var rules []FilterRules
	if cmd.CheckIllegal(chain) {
		return rules, buserr.New("ErrCmdIllegal")
	}
	stdout, err := RunWithStd(FilterTab, "-nL", chain)
	if err != nil {
		return rules, fmt.Errorf("load filter fules by chain %s failed, %v", chain, err)
	}
	lines := strings.Split(stdout, "\n")
	for i := 0; i < len(lines); i++ {
		fields := strings.Fields(lines[i])
		if len(fields) < 5 {
			continue
		}
		strategy := strings.ToLower(fields[0])
		if strategy != "accept" && strategy != "drop" && strategy != "reject" {
			continue
		}
		itemRule := FilterRules{
			Chain:    chain,
			Protocol: loadProtocol(fields[1]),
			SrcPort:  loadPort("src", fields),
			DstPort:  loadPort("dst", fields),
			SrcIP:    loadIP(fields[3]),
			DstIP:    loadIP(fields[4]),
			Strategy: strategy,
		}
		rules = append(rules, itemRule)
	}
	return rules, nil
}

func LoadDefaultStrategy(chain string) (string, error) {
	if cmd.CheckIllegal(chain) {
		return "", buserr.New("ErrCmdIllegal")
	}
	stdout, err := RunWithStd(FilterTab, "-L", chain)
	if err != nil {
		return "", fmt.Errorf("load filter fules by chain %s failed, %v", chain, err)
	}
	lines := strings.Split(stdout, "\n")
	for i := len(lines) - 1; i > 0; i-- {
		fields := strings.Fields(lines[i])
		if len(fields) < 5 {
			continue
		}
		if fields[0] == "DROP" && fields[1] == "all" && fields[3] == ANYWHERE && fields[4] == ANYWHERE {
			return DROP, nil
		}
	}
	return ACCEPT, nil
}

func LoadInitStatus(clientName, tab string) (bool, bool) {
	if clientName == "firewalld" {
		return true, true
	}
	if clientName == "ufw" {
		return true, true
	}
	switch tab {
	case "base":
		filterRules, err := RunWithStd(FilterTab, "-S")
		if err != nil {
			return false, false
		}
		lines := strings.Split(filterRules, "\n")
		initRules := []string{
			"-N " + Chain1PanelBasicBefore,
			"-N " + Chain1PanelBasic,
			"-N " + Chain1PanelBasicAfter,
			fmt.Sprintf("-A %s %s -j ACCEPT", Chain1PanelBasicBefore, strings.ReplaceAll(strings.ReplaceAll(IoRuleIn, "'", "\""), " -j ACCEPT", "")),
			fmt.Sprintf("-A %s %s -j ACCEPT", Chain1PanelBasicBefore, strings.ReplaceAll(strings.ReplaceAll(EstablishedRule, "'", "\""), " -j ACCEPT", "")),
			fmt.Sprintf("-A %s %s", Chain1PanelBasicAfter, DropAllTcp),
			fmt.Sprintf("-A %s %s", Chain1PanelBasicAfter, DropAllUdp),
		}
		bindRules := []string{
			fmt.Sprintf("-A %s -j %s", ChainInput, Chain1PanelBasicBefore),
			fmt.Sprintf("-A %s -j %s", ChainInput, Chain1PanelBasic),
			fmt.Sprintf("-A %s -j %s", ChainInput, Chain1PanelBasicAfter),
		}
		return checkWithInitAndBind(initRules, bindRules, lines)
	case "advance":
		filterRules, err := RunWithStd(FilterTab, "-S")
		if err != nil {
			return false, false
		}
		lines := strings.Split(filterRules, "\n")
		initRules := []string{
			"-N " + Chain1PanelInput,
			"-N " + Chain1PanelOutput,
		}
		bindRules := []string{
			fmt.Sprintf("-A %s -j %s", ChainInput, Chain1PanelInput),
			fmt.Sprintf("-A %s -j %s", ChainOutput, Chain1PanelOutput),
		}
		return checkWithInitAndBind(initRules, bindRules, lines)
	default:
		return false, false
	}
}

func checkWithInitAndBind(initRules, bindRules []string, lines []string) (bool, bool) {
	for _, rule := range initRules {
		found := false
		for _, line := range lines {
			if strings.TrimSpace(line) == strings.TrimSpace(rule) {
				found = true
				break
			}
		}
		if !found {
			global.LOG.Debugf("not found init rule: %s", rule)
			return false, false
		}
	}
	for _, rule := range bindRules {
		found := false
		for _, line := range lines {
			if strings.TrimSpace(line) == strings.TrimSpace(rule) {
				found = true
				break
			}
		}
		if !found {
			global.LOG.Debugf("not found bind rule: %s", rule)
			return true, false
		}
	}
	return true, true
}

func loadPort(position string, portStr []string) string {
	if len(portStr) < 7 {
		return ""
	}

	var portItem string
	if strings.Contains(portStr[6], "spt:") && position == "src" {
		portItem = strings.ReplaceAll(portStr[6], "spt:", "")
	}
	if strings.Contains(portStr[6], "dpt:") && position == "dst" {
		portItem = strings.ReplaceAll(portStr[6], "dpt:", "")
	}
	if strings.Contains(portStr[6], "spts:") && position == "src" {
		portItem = strings.ReplaceAll(portStr[6], "spts:", "")
	}
	if strings.Contains(portStr[6], "dpts:") && position == "dst" {
		portItem = strings.ReplaceAll(portStr[6], "dpts:", "")
	}
	portItem = strings.ReplaceAll(portItem, ":", "-")
	return portItem
}

func loadIP(ipStr string) string {
	if ipStr == ANYWHERE || ipStr == "0.0.0.0/0" {
		return ""
	}
	return ipStr
}

func loadProtocol(protocol string) string {
	switch protocol {
	case "0":
		return "all"
	case "1":
		return "icmp"
	case "6":
		return "tcp"
	case "17":
		return "udp"
	default:
		return protocol
	}
}

func validateRuleSafety(rule FilterRules, chain string) error {
	if strings.ToUpper(rule.Strategy) != "DROP" {
		return nil
	}

	if chain == ChainInput || chain == Chain1PanelInput || chain == Chain1PanelBasic {
		if rule.SrcIP == "0.0.0.0/0" && len(rule.SrcPort) == 0 && len(rule.DstPort) == 0 {
			return fmt.Errorf("unsafe DROP is not allowed")
		}
	}

	if chain == ChainOutput || chain == Chain1PanelOutput || chain == Chain1PanelBasicAfter {
		if rule.DstIP == "0.0.0.0/0" && len(rule.DstPort) == 0 && len(rule.SrcPort) == 0 {
			return fmt.Errorf("unsafe DROP is not allowed")
		}
	}

	return nil
}
