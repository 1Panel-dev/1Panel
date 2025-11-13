package iptables

import (
	"fmt"
	"strings"
	"time"

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
	iptablesArg := fmt.Sprintf("-A %s", chain)
	if policy.Protocol != "" {
		iptablesArg += fmt.Sprintf(" -p %s", policy.Protocol)
	}
	if len(policy.SrcPort) != 0 {
		iptablesArg += fmt.Sprintf(" --sport %s", policy.SrcPort)
	}
	if len(policy.DstPort) != 0 {
		iptablesArg += fmt.Sprintf(" --dport %s", policy.DstPort)
	}
	if policy.SrcIP != "" {
		iptablesArg += fmt.Sprintf(" -s %s", policy.SrcIP)
	}
	if policy.DstIP != "" {
		iptablesArg += fmt.Sprintf(" -d %s", policy.DstIP)
	}
	iptablesArg += fmt.Sprintf(" -j %s", policy.Strategy)

	return Run(FilterTab, iptablesArg)
}

func DeleteFilterRule(chain string, policy FilterRules) error {
	iptablesArg := fmt.Sprintf("-D %s", chain)
	if policy.Protocol != "" {
		iptablesArg += fmt.Sprintf(" -p %s", policy.Protocol)
	}
	if len(policy.SrcPort) != 0 {
		iptablesArg += fmt.Sprintf(" --sport %s", policy.SrcPort)
	}
	if len(policy.DstPort) != 0 {
		iptablesArg += fmt.Sprintf(" --dport %s", policy.DstPort)
	}
	if policy.SrcIP != "" {
		iptablesArg += fmt.Sprintf(" -s %s", policy.SrcIP)
	}
	if policy.DstIP != "" {
		iptablesArg += fmt.Sprintf(" -d %s", policy.DstIP)
	}
	iptablesArg += fmt.Sprintf(" -j %s", policy.Strategy)

	return Run(FilterTab, iptablesArg)
}

func ReadFilterRulesByChain(chain string) ([]FilterRules, error) {
	var rules []FilterRules
	cmdMgr := cmd.NewCommandMgr(cmd.WithIgnoreExist1(), cmd.WithTimeout(20*time.Second))
	stdout, err := cmdMgr.RunWithStdoutBashCf("%s iptables -w -t %s -nL %s", cmd.SudoHandleCmd(), FilterTab, chain)
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
	cmdMgr := cmd.NewCommandMgr(cmd.WithIgnoreExist1(), cmd.WithTimeout(20*time.Second))
	stdout, err := cmdMgr.RunWithStdoutBashCf("%s iptables -w -t %s -L %s", cmd.SudoHandleCmd(), FilterTab, chain)
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
