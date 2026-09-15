package iptables_helper

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
)

func LoadInitStatus(tab string) (bool, bool, error) {
	return loadInitStatus(tab, RunWithStd, true)
}

func LoadFamilyInitStatus(family, tab string) (bool, bool, error) {
	switch family {
	case constant.FirewallFamilyIPv4:
		return loadInitStatus(tab, RunWithStd, true)
	case constant.FirewallFamilyIPv6:
		return loadInitStatus(tab, RunIPv6WithStd, true)
	default:
		return false, false, fmt.Errorf("unsupported iptables family %q", family)
	}
}

func LoadFamilyBindStatus(family string) (bool, error) {
	var (
		output string
		err    error
	)
	switch family {
	case constant.FirewallFamilyIPv4:
		output, err = RunWithStd(FilterTab, "-S", InputChain)
	case constant.FirewallFamilyIPv6:
		output, err = RunIPv6WithStd(FilterTab, "-S", InputChain)
	default:
		return false, fmt.Errorf("unsupported iptables family %q", family)
	}
	if err != nil {
		return false, err
	}
	return hasBaseChainBinding(output), nil
}

func hasBaseChainBinding(output string) bool {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		for _, chain := range BasicChains() {
			if line == fmt.Sprintf("-A %s -j %s", InputChain, chain) {
				return true
			}
		}
	}
	return false
}

func loadInitStatus(tab string, runner func(string, ...string) (string, error), requireTerminalRules bool) (bool, bool, error) {
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
		}
		if requireTerminalRules {
			initRules = append(initRules,
				fmt.Sprintf("-A %s %s", BasicAfterChain, DropAllTcp),
				fmt.Sprintf("-A %s %s", BasicAfterChain, DropAllUdp),
			)
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
			if global.LOG != nil {
				global.LOG.Debugf("not found init rule: %s", rule)
			}
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
			if global.LOG != nil {
				global.LOG.Debugf("not found bind rule: %s", rule)
			}
			return true, false
		}
	}
	return true, true
}
