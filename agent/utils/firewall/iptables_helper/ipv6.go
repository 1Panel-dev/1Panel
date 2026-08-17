package iptables_helper

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
)

func EnsureIPv6BaseChains(panelPort string) error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil || !commands.IPv6Available() {
		return fmt.Errorf("ip6tables and ip6tables-restore are required")
	}
	if panelPort == "" {
		return fmt.Errorf("panel port is required")
	}
	chains := []struct {
		name string
		file string
	}{
		{name: BasicBeforeChain, file: IPv6FileName(BasicBeforeFileName)},
		{name: BasicChain, file: IPv6FileName(BasicFileName)},
		{name: BasicAfterChain, file: IPv6FileName(BasicAfterFileName)},
	}
	for _, chain := range chains {
		exists, err := CheckIPv6ChainExist(FilterTab, chain.name)
		if err != nil {
			return err
		}
		if !exists {
			if err := LoadIPv6RulesFromFile(FilterTab, chain.name, chain.file); err != nil {
				return err
			}
		}
	}
	protectedRules := [][]string{
		{"-i", "lo", "-j", "ACCEPT", "-m", "comment", "--comment", "Loopback Whitelist"},
		{"-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT", "-m", "comment", "--comment", "ESTABLISHED Whitelist"},
		{"-p", "tcp", "-m", "tcp", "--dport", panelPort, "-j", "ACCEPT", "-m", "comment", "--comment", "1Panel Port Whitelist"},
	}
	for _, rule := range protectedRules {
		if err := AddIPv6Rule(FilterTab, BasicBeforeChain, rule...); err != nil {
			return err
		}
	}
	for index, chain := range chains {
		if err := BindIPv6Chain(FilterTab, InputChain, chain.name, index+1); err != nil {
			return err
		}
		if err := SaveIPv6RulesToFile(FilterTab, chain.name, chain.file); err != nil {
			return err
		}
	}
	return nil
}

func UnbindIPv6BaseChains() error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil || !commands.IPv6Available() {
		return nil
	}
	for _, chain := range []string{BasicAfterChain, BasicChain, BasicBeforeChain} {
		if err := UnbindIPv6Chain(FilterTab, InputChain, chain); err != nil {
			return err
		}
	}
	return nil
}
