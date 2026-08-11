package iptables_helper

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

func EnsureIPv6BaseChains(panelPort string) error {
	if !cmd.Which("ip6tables") || !cmd.Which("ip6tables-restore") {
		return fmt.Errorf("ip6tables and ip6tables-restore are required")
	}
	if panelPort == "" {
		return fmt.Errorf("panel port is required")
	}
	chains := []struct {
		name string
		file string
	}{
		{name: Chain1PanelBasicBefore, file: IPv6FileName(BasicBeforeFileName)},
		{name: Chain1PanelBasic, file: IPv6FileName(BasicFileName)},
		{name: Chain1PanelBasicAfter, file: IPv6FileName(BasicAfterFileName)},
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
		if err := AddIPv6Rule(FilterTab, Chain1PanelBasicBefore, rule...); err != nil {
			return err
		}
	}
	for index, chain := range chains {
		if err := BindIPv6Chain(FilterTab, ChainInput, chain.name, index+1); err != nil {
			return err
		}
		if err := SaveIPv6RulesToFile(FilterTab, chain.name, chain.file); err != nil {
			return err
		}
	}
	return nil
}

func UnbindIPv6BaseChains() error {
	if !cmd.Which("ip6tables") {
		return nil
	}
	for _, chain := range []string{Chain1PanelBasicAfter, Chain1PanelBasic, Chain1PanelBasicBefore} {
		if err := UnbindIPv6Chain(FilterTab, ChainInput, chain); err != nil {
			return err
		}
	}
	return nil
}
