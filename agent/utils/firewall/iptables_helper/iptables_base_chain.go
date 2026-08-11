package iptables_helper

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type IptablesManager struct {
	UpdateSetting     func(key, value string) error
	PanelPort         func() string
	LoadRequiredPorts func() ([]PortWhitelist, error)
}

type BaseOperation string

const (
	BaseOperationInit            BaseOperation = "init-base"
	BaseOperationBind            BaseOperation = "bind-base"
	BaseOperationBindWithoutInit BaseOperation = "bind-base-without-init"
	BaseOperationUnbind          BaseOperation = "unbind-base"
)

func (m *IptablesManager) Operate(operation BaseOperation) error {
	switch operation {
	case BaseOperationInit, BaseOperationBind:
		if !cmd.Which("iptables") {
			return fmt.Errorf("failed to find iptables")
		}
		return m.enableBase(true)
	case BaseOperationBindWithoutInit:
		return m.enableBase(false)
	case BaseOperationUnbind:
		return m.disableBase()
	default:
		return fmt.Errorf("unsupported iptables base operation %q", operation)
	}
}

func (m *IptablesManager) enableBase(prepare bool) error {
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

func (m *IptablesManager) disableBase() error {
	for _, item := range []struct{ parent, chain string }{
		{ChainInput, Chain1PanelBasicAfter},
		{ChainInput, Chain1PanelBasicBefore},
		{ChainInput, Chain1PanelBasic},
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
	for _, chain := range []string{
		Chain1PanelBasicBefore,
		Chain1PanelBasic,
		Chain1PanelBasicAfter,
	} {
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
		{ChainInput, Chain1PanelBasicBefore, 1},
		{ChainInput, Chain1PanelBasic, 2},
		{ChainInput, Chain1PanelBasicAfter, 3},
	} {
		if err := BindChain(FilterTab, item.parent, item.chain, item.position); err != nil {
			return err
		}
	}
	return nil
}

func saveBaseChains() error {
	for _, item := range []struct{ chain, file string }{
		{Chain1PanelBasicBefore, BasicBeforeFileName},
		{Chain1PanelBasic, BasicFileName},
		{Chain1PanelBasicAfter, BasicAfterFileName},
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
		{Chain1PanelBasicBefore, BasicBeforeFileName},
		{Chain1PanelBasic, BasicFileName},
		{Chain1PanelBasicAfter, BasicAfterFileName},
	} {
		if err := LoadRulesFromFile(FilterTab, item.chain, item.file); err != nil {
			return err
		}
	}
	return AddRule(
		FilterTab,
		Chain1PanelBasicBefore,
		"-p", "tcp", "-m", "tcp", "--dport", panelPort, "-j", "ACCEPT",
	)
}

func (m *IptablesManager) initPreRules() error {
	if err := AddRule(FilterTab, Chain1PanelBasicBefore, "-i", "lo", "-j", "ACCEPT", "-m", "comment", "--comment", "Loopback Whitelist"); err != nil {
		return err
	}
	if err := AddRule(FilterTab, Chain1PanelBasicBefore, "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT", "-m", "comment", "--comment", "ESTABLISHED Whitelist"); err != nil {
		return err
	}
	if err := m.SyncRequiredPorts(false); err != nil {
		return err
	}
	if err := AddRule(FilterTab, Chain1PanelBasicAfter, "-p", "tcp", "-j", "DROP"); err != nil {
		return err
	}
	if err := AddRule(FilterTab, Chain1PanelBasicAfter, "-p", "udp", "-j", "DROP"); err != nil {
		return err
	}
	return nil
}

func (m *IptablesManager) ensureIPv6BaseChains() error {
	if !cmd.Which("ip6tables") || !cmd.Which("ip6tables-restore") {
		return nil
	}
	return EnsureIPv6BaseChains(m.panelPort())
}

func (m *IptablesManager) SyncRequiredPorts(withSave bool) error {
	requiredPorts, err := m.loadRequiredPorts()
	if err != nil {
		return err
	}
	return applyRequiredFirewallPortWhiteListRules(requiredPorts, withSave)
}

func applyRequiredFirewallPortWhiteListRules(portWhiteList []PortWhitelist, withSave bool) error {
	if err := syncRequiredFirewallPortWhiteListRules(portWhiteList); err != nil {
		return err
	}
	for _, item := range portWhiteList {
		if err := AddRule(FilterTab, Chain1PanelBasicBefore, "-p", item.Protocol, "-m", item.Protocol, "--dport", item.Port, "-j", "ACCEPT"); err != nil {
			return err
		}
	}
	if !withSave {
		return nil
	}
	if err := SaveRulesToFile(FilterTab, Chain1PanelBasicBefore, BasicBeforeFileName); err != nil {
		return err
	}
	return SaveRulesToFile(FilterTab, Chain1PanelBasicAfter, BasicAfterFileName)
}

func syncRequiredFirewallPortWhiteListRules(portWhiteList []PortWhitelist) error {
	tcpWhitelist := make(map[string]struct{})
	udpWhitelist := make(map[string]struct{})
	for _, item := range portWhiteList {
		if item.Protocol == "udp" {
			udpWhitelist[item.Port] = struct{}{}
			continue
		}
		tcpWhitelist[item.Port] = struct{}{}
	}

	if err := cleanExtraFirewallPortRules(Chain1PanelBasicBefore, "tcp", tcpWhitelist); err != nil {
		return err
	}
	if err := cleanExtraFirewallPortRules(Chain1PanelBasicBefore, "udp", udpWhitelist); err != nil {
		return err
	}
	return cleanExtraFirewallPortRules(Chain1PanelBasicAfter, "udp", map[string]struct{}{})
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

func (m *IptablesManager) updateSetting(key, value string) error {
	if m != nil && m.UpdateSetting != nil {
		return m.UpdateSetting(key, value)
	}
	return nil
}

func (m *IptablesManager) panelPort() string {
	if m != nil && m.PanelPort != nil {
		return m.PanelPort()
	}
	return ""
}

func (m *IptablesManager) loadRequiredPorts() ([]PortWhitelist, error) {
	if m != nil && m.LoadRequiredPorts != nil {
		return m.LoadRequiredPorts()
	}
	return nil, fmt.Errorf("load required firewall ports is not configured")
}
