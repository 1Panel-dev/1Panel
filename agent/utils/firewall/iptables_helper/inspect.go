package iptables_helper

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
)

func LoadInitStatus(tab string) (bool, bool, error) {
	return loadInitStatus(tab, RunWithStd)
}

func LoadFamilyInitStatus(family, tab string) (bool, bool, error) {
	initialized, bound, _, err := LoadFamilyState(family, tab)
	return initialized, bound, err
}

func LoadFamilyState(family, tab string) (bool, bool, bool, error) {
	runner := RunWithStd
	switch family {
	case constant.FirewallFamilyIPv4:
	case constant.FirewallFamilyIPv6:
		runner = RunIPv6WithStd
	default:
		return false, false, false, fmt.Errorf("unsupported iptables family %q", family)
	}
	if tab != "base" {
		return false, false, false, nil
	}
	output, err := runner(FilterTab, "-S")
	if err != nil {
		return false, false, false, err
	}
	count := 0
	for _, chain := range BasicChains() {
		if containsIptablesRule(output, "-N "+chain) {
			count++
		}
	}
	initialized := count == len(BasicChains())
	_, bound := checkWithInitAndBind([]string{"-N " + BasicBeforeChain, "-N " + BasicChain, "-N " + BasicAfterChain}, []string{"-A INPUT -j " + BasicBeforeChain, "-A INPUT -j " + BasicChain, "-A INPUT -j " + BasicAfterChain}, strings.Split(output, "\n"))
	return initialized, bound, count > 0 && !initialized, nil
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
	output := strings.Join(lines, "\n")
	for _, rule := range initRules {
		if !containsIptablesRule(output, rule) {
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
