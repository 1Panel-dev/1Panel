package iptables_helper

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
)

type Manager struct {
	UpdateSetting     func(key, value string) error
	PanelPort         func() string
	LoadRequiredPorts func() ([]firewall.PortWhitelist, error)
}

func (m *Manager) Operate(operation firewall.BaseOperation) error {
	switch operation {
	case firewall.BaseOperationInit, firewall.BaseOperationBind:
		if _, err := lifecycle.ResolveIptablesCommands(); err != nil {
			return fmt.Errorf("failed to find iptables")
		}
		return m.enableBase(true)
	case firewall.BaseOperationBindWithoutInit:
		return m.enableBase(false)
	case firewall.BaseOperationUnbind:
		return m.disableBase()
	default:
		return fmt.Errorf("unsupported iptables base operation %q", operation)
	}
}

func (m *Manager) enableBase(prepare bool) error {
	if prepare {
		if err := ensureBaseChains(); err != nil {
			return err
		}
		if err := m.initPreRules(); err != nil {
			return err
		}
		if err := saveBaseChains(); err != nil {
			return err
		}
	}
	if err := bindBaseChains(); err != nil {
		return err
	}
	if err := m.ensureIPv6BaseChains(); err != nil {
		return err
	}
	return m.updateSetting("IptablesStatus", constant.StatusEnable)
}

func (m *Manager) disableBase() error {
	for _, item := range []struct{ parent, chain string }{
		{InputChain, BasicAfterChain},
		{InputChain, BasicBeforeChain},
		{InputChain, BasicChain},
	} {
		if err := UnbindChain(FilterTab, item.parent, item.chain); err != nil {
			return err
		}
	}
	if err := UnbindIPv6BaseChains(); err != nil {
		return err
	}
	return m.updateSetting("IptablesStatus", constant.StatusDisable)
}

func ensureBaseChains() error {
	for _, chain := range BasicChains() {
		if err := AddChain(FilterTab, chain); err != nil {
			return err
		}
	}
	return nil
}

func bindBaseChains() error {
	for _, item := range []struct {
		parent, chain string
		position      int
	}{
		{InputChain, BasicBeforeChain, 1},
		{InputChain, BasicChain, 2},
		{InputChain, BasicAfterChain, 3},
	} {
		if err := BindChain(FilterTab, item.parent, item.chain, item.position); err != nil {
			return err
		}
	}
	return nil
}

func saveBaseChains() error {
	for _, item := range []struct{ chain, file string }{
		{BasicBeforeChain, BasicBeforeFileName},
		{BasicChain, BasicFileName},
		{BasicAfterChain, BasicAfterFileName},
	} {
		if err := SaveRulesToFile(FilterTab, item.chain, item.file); err != nil {
			return err
		}
	}
	return nil
}

func RestoreIPv4BaseChains(panelPort string) error {
	if panelPort == "" {
		return fmt.Errorf("panel port is required")
	}
	for _, item := range []struct{ chain, file string }{
		{BasicBeforeChain, BasicBeforeFileName},
		{BasicChain, BasicFileName},
		{BasicAfterChain, BasicAfterFileName},
	} {
		if err := LoadRulesFromFile(FilterTab, item.chain, item.file); err != nil {
			return err
		}
	}
	return AddRule(
		FilterTab,
		BasicBeforeChain,
		"-p", "tcp", "-m", "tcp", "--dport", panelPort, "-j", "ACCEPT",
	)
}

func (m *Manager) initPreRules() error {
	if err := AddRule(FilterTab, BasicBeforeChain, "-i", "lo", "-j", "ACCEPT", "-m", "comment", "--comment", "Loopback Whitelist"); err != nil {
		return err
	}
	if err := AddRule(FilterTab, BasicBeforeChain, "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT", "-m", "comment", "--comment", "ESTABLISHED Whitelist"); err != nil {
		return err
	}
	if err := m.SyncRequiredPorts(false); err != nil {
		return err
	}
	if err := AddRule(FilterTab, BasicAfterChain, "-p", "tcp", "-j", "DROP"); err != nil {
		return err
	}
	if err := AddRule(FilterTab, BasicAfterChain, "-p", "udp", "-j", "DROP"); err != nil {
		return err
	}
	return nil
}

func (m *Manager) ensureIPv6BaseChains() error {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil || !commands.IPv6Available() {
		return nil
	}
	return EnsureIPv6BaseChains(m.panelPort())
}

func (m *Manager) SyncRequiredPorts(withSave bool) error {
	requiredPorts, err := m.loadRequiredPorts()
	if err != nil {
		return err
	}
	return applyRequiredFirewallPortWhiteListRules(requiredPorts, withSave)
}

func applyRequiredFirewallPortWhiteListRules(portWhiteList []firewall.PortWhitelist, withSave bool) error {
	if err := syncRequiredFirewallPortWhiteListRules(portWhiteList); err != nil {
		return err
	}
	for _, item := range portWhiteList {
		if err := AddRule(FilterTab, BasicBeforeChain, "-p", item.Protocol, "-m", item.Protocol, "--dport", item.Port, "-j", "ACCEPT"); err != nil {
			return err
		}
	}
	if !withSave {
		return nil
	}
	if err := SaveRulesToFile(FilterTab, BasicBeforeChain, BasicBeforeFileName); err != nil {
		return err
	}
	return SaveRulesToFile(FilterTab, BasicAfterChain, BasicAfterFileName)
}

func syncRequiredFirewallPortWhiteListRules(portWhiteList []firewall.PortWhitelist) error {
	tcpWhitelist := make(map[string]struct{})
	udpWhitelist := make(map[string]struct{})
	for _, item := range portWhiteList {
		if item.Protocol == "udp" {
			udpWhitelist[item.Port] = struct{}{}
			continue
		}
		tcpWhitelist[item.Port] = struct{}{}
	}

	if err := cleanExtraFirewallPortRules(BasicBeforeChain, "tcp", tcpWhitelist); err != nil {
		return err
	}
	if err := cleanExtraFirewallPortRules(BasicBeforeChain, "udp", udpWhitelist); err != nil {
		return err
	}
	return cleanExtraFirewallPortRules(BasicAfterChain, "udp", map[string]struct{}{})
}

func cleanExtraFirewallPortRules(chain, protocol string, whitelist map[string]struct{}) error {
	rules, err := ReadFilterRulesByChain(chain)
	if err != nil {
		return err
	}
	kept := make(map[string]struct{})
	for _, rule := range rules {
		if rule.Strategy != "accept" || rule.Protocol != protocol || rule.DstPort == "" || rule.SrcIP != "" || rule.DstIP != "" || rule.SrcPort != "" {
			continue
		}
		if _, ok := whitelist[rule.DstPort]; ok {
			if _, seen := kept[rule.DstPort]; !seen {
				kept[rule.DstPort] = struct{}{}
				continue
			}
		}
		if err := DeleteRule(FilterTab, chain, "-p", protocol, "-m", protocol, "--dport", rule.DstPort, "-j", "ACCEPT"); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) updateSetting(key, value string) error {
	if m != nil && m.UpdateSetting != nil {
		return m.UpdateSetting(key, value)
	}
	return nil
}

func (m *Manager) panelPort() string {
	if m != nil && m.PanelPort != nil {
		return m.PanelPort()
	}
	return ""
}

func (m *Manager) loadRequiredPorts() ([]firewall.PortWhitelist, error) {
	if m != nil && m.LoadRequiredPorts != nil {
		return m.LoadRequiredPorts()
	}
	return nil, fmt.Errorf("load required firewall ports is not configured")
}
