package iptables_helper

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
)

func (m *Manager) EnsureIPv6BaseChains() error {
	return EnsureIPv6BaseChains(m.panelPort())
}

func RepairIPv6BaseChains(panelPort string) error {
	initialized, bound, err := LoadFamilyInitStatus(constant.FirewallFamilyIPv6, "base")
	if err != nil {
		return err
	}
	return repairIPv6BaseChains(initialized, bound, BindIPv6BaseChains, func() error {
		return EnsureIPv6BaseChains(panelPort)
	})
}

func repairIPv6BaseChains(initialized, bound bool, bind, ensure func() error) error {
	if initialized {
		if bound {
			return nil
		}
		return bind()
	}
	return ensure()
}

func EnsureIPv6BaseChains(panelPort string) error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil || !commands.IPv6Available() {
		return fmt.Errorf("ip6tables and ip6tables-restore are required")
	}
	if panelPort == "" {
		return fmt.Errorf("panel port is required")
	}
	if err := ensureBaseChainsFamily(true); err != nil {
		return err
	}
	script, err := buildBaseChainsRestoreScript(global.Dir.FirewallDir, panelPort, true)
	if err != nil {
		return err
	}
	protectedRules := []string{
		"-A " + BasicBeforeChain + " " + IoRuleIn,
		"-A " + BasicBeforeChain + " " + EstablishedRule,
	}
	for _, rule := range protectedRules {
		if !containsIptablesRule(script, rule) {
			script = strings.Replace(script, "COMMIT\n", rule+"\nCOMMIT\n", 1)
		}
	}
	if err := restoreRules(commands.Restore6, script); err != nil {
		return fmt.Errorf("batch initialize IPv6 base chains: %w", err)
	}
	if err := setBaseChainBindings(true, true); err != nil {
		return err
	}
	for _, chain := range []struct{ name, file string }{
		{BasicBeforeChain, IPv6FileName(BasicBeforeFileName)},
		{BasicChain, IPv6FileName(BasicFileName)},
		{BasicAfterChain, IPv6FileName(BasicAfterFileName)},
	} {
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
	return setBaseChainBindings(true, false)
}

func BindIPv6BaseChains() error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil || !commands.IPv6Available() {
		return nil
	}
	return setBaseChainBindings(true, true)
}
