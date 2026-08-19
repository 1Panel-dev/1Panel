package iptables_helper

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
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

func LoadInitStatus(tab string) (bool, bool, error) {
	return loadInitStatus(tab, RunWithStd)
}

func LoadFamilyInitStatus(family, tab string) (bool, bool, error) {
	switch family {
	case constant.FirewallFamilyIPv4:
		return loadInitStatus(tab, RunWithStd)
	case constant.FirewallFamilyIPv6:
		return loadInitStatus(tab, RunIPv6WithStd)
	default:
		return false, false, fmt.Errorf("unsupported iptables family %q", family)
	}
}

func loadInitStatus(tab string, runner func(string, ...string) (string, error)) (bool, bool, error) {
	switch tab {
	case "base":
		filterRules, err := runner(FilterTab, "-S")
		if err != nil {
			return false, false, fmt.Errorf("load iptables initialization status: %w", err)
		}
		lines := strings.Split(filterRules, "\n")
		initRules := []string{
			"-N " + BasicBeforeChain,
			"-N " + BasicChain,
			"-N " + BasicAfterChain,
			fmt.Sprintf("-A %s %s -j ACCEPT", BasicBeforeChain, strings.ReplaceAll(strings.ReplaceAll(IoRuleIn, "'", "\""), " -j ACCEPT", "")),
			fmt.Sprintf("-A %s %s -j ACCEPT", BasicBeforeChain, strings.ReplaceAll(strings.ReplaceAll(EstablishedRule, "'", "\""), " -j ACCEPT", "")),
			fmt.Sprintf("-A %s %s", BasicAfterChain, DropAllTcp),
			fmt.Sprintf("-A %s %s", BasicAfterChain, DropAllUdp),
		}
		bindRules := []string{
			fmt.Sprintf("-A %s -j %s", InputChain, BasicBeforeChain),
			fmt.Sprintf("-A %s -j %s", InputChain, BasicChain),
			fmt.Sprintf("-A %s -j %s", InputChain, BasicAfterChain),
		}
		isInit, isBind := checkWithInitAndBind(initRules, bindRules, lines)
		return isInit, isBind, nil
	default:
		return false, false, nil
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
