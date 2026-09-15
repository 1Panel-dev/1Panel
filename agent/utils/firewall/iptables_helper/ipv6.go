package iptables_helper

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
)

func (m *Manager) EnsureIPv6BaseChains() error {
	ports, err := m.loadRequiredPorts()
	if err != nil {
		return err
	}
	return EnsureIPv6BaseChains(ports)
}

func EnsureIPv6BaseChains(ports []firewall.PortWhitelist) error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil || !commands.IPv6Available() {
		return fmt.Errorf("ip6tables and ip6tables-restore are required")
	}
	output, err := RunIPv6WithStd(FilterTab, "-S")
	if err != nil {
		return err
	}
	if err := ensureBaseChainsFamily(true); err != nil {
		return err
	}
	script, err := buildIPv6BaseInitializationScript(global.Dir.FirewallDir, ports, output)
	if err != nil {
		return err
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

func buildIPv6BaseInitializationScript(dir string, ports []firewall.PortWhitelist, output string) (string, error) {
	for _, chain := range BasicChains() {
		if !containsIptablesRule(output, "-N "+chain) {
			return buildBaseChainsRestoreScript(dir, true, ports...)
		}
	}
	defaults, err := baseDefaultRules(ports, constant.FirewallFamilyIPv6)
	if err != nil {
		return "", err
	}
	var script strings.Builder
	script.WriteString("*filter\n")
	for _, rule := range defaults {
		if !containsIptablesRule(output, rule) {
			script.WriteString(rule + "\n")
		}
	}
	script.WriteString("COMMIT\n")
	return script.String(), nil
}

func baseDefaultRules(ports []firewall.PortWhitelist, family string) ([]string, error) {
	ports, err := firewall.NormalizeRequiredPorts(ports)
	if err != nil {
		return nil, err
	}
	rules := []string{"-A " + BasicBeforeChain + " " + IoRuleIn, "-A " + BasicBeforeChain + " " + EstablishedRule}
	for _, port := range firewall.ExpandPortWhitelist(ports) {
		if port.Family == family {
			rules = append(rules, iptablesSystemPortRuleLine(port))
		}
	}
	return append(rules, "-A "+BasicAfterChain+" "+DropAllTcp, "-A "+BasicAfterChain+" "+DropAllUdp), nil
}
