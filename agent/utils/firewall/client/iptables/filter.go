package iptables

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type FilterRules struct {
	ID          uint   `json:"id"`
	Protocol    string `json:"protocol"`
	SrcPort     uint   `json:"srcPort"`
	DstPort     uint   `json:"dstPort"`
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
	if policy.SrcPort != 0 {
		iptablesArg += fmt.Sprintf(" --sport %d", policy.SrcPort)
	}
	if policy.DstPort != 0 {
		iptablesArg += fmt.Sprintf(" --dport %d", policy.DstPort)
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
	if policy.SrcPort != 0 {
		iptablesArg += fmt.Sprintf(" --sport %d", policy.SrcPort)
	}
	if policy.DstPort != 0 {
		iptablesArg += fmt.Sprintf(" --dport %d", policy.DstPort)
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
	stdout, err := cmdMgr.RunWithStdoutBashCf("%s iptables -t %s -L %s", cmd.SudoHandleCmd(), FilterTab, chain)
	if err != nil {
		return rules, fmt.Errorf("load filter fules by chain %s failed, %v", chain, err)
	}
	lines := strings.Split(stdout, "\n")
	for i := 0; i < len(lines); i++ {
		fields := strings.Fields(lines[i])
		if len(fields) < 7 {
			continue
		}
		itemRule := FilterRules{
			Protocol: fields[1],
			SrcPort:  loadPort("src", fields[6]),
			DstPort:  loadPort("dst", fields[6]),
			SrcIP:    loadIP(fields[3]),
			DstIP:    loadIP(fields[4]),
			Strategy: fields[0],
		}
		rules = append(rules, itemRule)
	}
	return rules, nil
}

func LoadDefaultStrategy(chain string) (string, error) {
	cmdMgr := cmd.NewCommandMgr(cmd.WithIgnoreExist1(), cmd.WithTimeout(20*time.Second))
	stdout, err := cmdMgr.RunWithStdoutBashCf("%s iptables -t %s -L %s", cmd.SudoHandleCmd(), FilterTab, chain)
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

func loadPort(position, portStr string) uint {
	var portItem string
	if strings.Contains(portStr, "spt:") && position == "src" {
		portItem = strings.ReplaceAll(portStr, "spt:", "")
	}
	if strings.Contains(portStr, "dpt:") && position == "dst" {
		portItem = strings.ReplaceAll(portStr, "dpt:", "")
	}
	if len(portItem) == 0 {
		return 0
	}
	port, _ := strconv.Atoi(portItem)
	return uint(port)
}

func loadIP(ipStr string) string {
	if ipStr == ANYWHERE {
		return ""
	}
	return ipStr
}

func validateRuleSafety(rule FilterRules, chain string) error {
	if strings.ToUpper(rule.Strategy) != "DROP" {
		return nil
	}

	if chain == ChainInput || chain == Chain1PanelInput || chain == Chain1PanelBasic {
		if rule.SrcIP == "0.0.0.0/0" && rule.SrcPort == 0 && rule.DstPort == 0 {
			return fmt.Errorf("unsafe DROP is not allowed")
		}
	}

	if chain == ChainOutput || chain == Chain1PanelOutput || chain == Chain1PanelBasicAfter {
		if rule.DstIP == "0.0.0.0/0" && rule.DstPort == 0 && rule.SrcPort == 0 {
			return fmt.Errorf("unsafe DROP is not allowed")
		}
	}

	return nil
}
