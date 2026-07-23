package client

import (
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/client/iptables"
)

// PortWhiteListEntry is one always open port of the 1Panel port whitelist.
type PortWhiteListEntry struct {
	Port     string
	Protocol string
}

// PortWhiteList is the whitelist state a provider needs in order to sync itself.
// Required entries (1Panel and SSH ports) are never removed, Previous is the
// configured list before the change and stays empty when the whole list is
// re-applied instead of diffed.
type PortWhiteList struct {
	Configured []PortWhiteListEntry
	Required   []PortWhiteListEntry
	Previous   []PortWhiteListEntry
}

func (e PortWhiteListEntry) key() string {
	return e.Port + "/" + e.Protocol
}

func (e PortWhiteListEntry) rule() FireInfo {
	return FireInfo{Port: e.Port, Protocol: e.Protocol, Strategy: "accept"}
}

func (l PortWhiteList) desired() []PortWhiteListEntry {
	return normalizePortWhiteList(append(append([]PortWhiteListEntry{}, l.Configured...), l.Required...))
}

func (l PortWhiteList) previous() []PortWhiteListEntry {
	return normalizePortWhiteList(append(append([]PortWhiteListEntry{}, l.Previous...), l.Required...))
}

func normalizePortWhiteList(entries []PortWhiteListEntry) []PortWhiteListEntry {
	ports := make([]PortWhiteListEntry, 0, len(entries))
	exists := make(map[string]struct{})
	for _, item := range entries {
		if item.Port == "" {
			continue
		}
		if _, ok := exists[item.key()]; ok {
			continue
		}
		exists[item.key()] = struct{}{}
		ports = append(ports, item)
	}
	return ports
}

func portWhiteListKeys(entries []PortWhiteListEntry) map[string]struct{} {
	keys := make(map[string]struct{})
	for _, item := range entries {
		keys[item.key()] = struct{}{}
	}
	return keys
}

// nativePortWriter is the surface the ufw/firewalld whitelist needs; both write
// the whitelist through the same native port command they use for user rules.
type nativePortWriter interface {
	Status() (bool, error)
	Reload() error
	Port(port FireInfo, operation string) error
}

func addNativePortWhiteList(client nativePortWriter, list PortWhiteList) error {
	for _, item := range list.desired() {
		if err := client.Port(item.rule(), "add"); err != nil {
			return err
		}
	}
	return client.Reload()
}

func syncNativePortWhiteList(client nativePortWriter, list PortWhiteList) error {
	isActive, _ := client.Status()
	if !isActive {
		return nil
	}
	desired, previous := list.desired(), list.previous()
	desiredKeys, previousKeys := portWhiteListKeys(desired), portWhiteListKeys(previous)
	for _, item := range previous {
		if _, ok := desiredKeys[item.key()]; ok {
			continue
		}
		if err := client.Port(item.rule(), "remove"); err != nil {
			return err
		}
	}
	for _, item := range desired {
		if _, ok := previousKeys[item.key()]; ok {
			continue
		}
		if err := client.Port(item.rule(), "add"); err != nil {
			return err
		}
	}
	return client.Reload()
}

// SyncIptablesPortWhiteList writes the whitelist into the managed chains: required
// ports into 1PANEL_BASIC_BEFORE, configured ports into 1PANEL_BASIC. withSave is
// false while the chains are still being built and the caller persists them itself.
func SyncIptablesPortWhiteList(list PortWhiteList, withSave bool) error {
	if err := applyRequiredIptablesPortWhiteList(normalizePortWhiteList(list.Required), withSave); err != nil {
		return err
	}
	return applyIptablesPortWhiteList(list.Configured, list.Previous, withSave)
}

func applyRequiredIptablesPortWhiteList(entries []PortWhiteListEntry, withSave bool) error {
	if err := cleanRequiredIptablesPortWhiteList(entries); err != nil {
		return err
	}
	for _, item := range entries {
		if err := iptables.AddRule(iptables.FilterTab, iptables.Chain1PanelBasicBefore, "-p", item.Protocol, "-m", item.Protocol, "--dport", item.Port, "-j", "ACCEPT"); err != nil {
			return err
		}
	}
	if !withSave {
		return nil
	}
	if err := iptables.SaveRulesToFile(iptables.FilterTab, iptables.Chain1PanelBasicBefore, iptables.BasicBeforeFileName); err != nil {
		return err
	}
	return iptables.SaveRulesToFile(iptables.FilterTab, iptables.Chain1PanelBasicAfter, iptables.BasicAfterFileName)
}

func applyIptablesPortWhiteList(entries, previous []PortWhiteListEntry, withSave bool) error {
	if err := cleanIptablesPortWhiteList(entries, previous); err != nil {
		return err
	}
	for _, item := range entries {
		if err := iptables.AddRule(iptables.FilterTab, iptables.Chain1PanelBasic, "-p", item.Protocol, "-m", item.Protocol, "--dport", item.Port, "-j", "ACCEPT"); err != nil {
			return err
		}
	}
	if !withSave {
		return nil
	}
	return iptables.SaveRulesToFile(iptables.FilterTab, iptables.Chain1PanelBasic, iptables.BasicFileName)
}

func cleanRequiredIptablesPortWhiteList(entries []PortWhiteListEntry) error {
	tcpWhitelist := make(map[string]struct{})
	udpWhitelist := make(map[string]struct{})
	for _, item := range entries {
		if item.Protocol == "udp" {
			udpWhitelist[item.Port] = struct{}{}
			continue
		}
		tcpWhitelist[item.Port] = struct{}{}
	}

	if err := cleanExtraIptablesPortRules(iptables.Chain1PanelBasicBefore, "tcp", tcpWhitelist); err != nil {
		return err
	}
	if err := cleanExtraIptablesPortRules(iptables.Chain1PanelBasicBefore, "udp", udpWhitelist); err != nil {
		return err
	}
	return cleanExtraIptablesPortRules(iptables.Chain1PanelBasicAfter, "udp", map[string]struct{}{})
}

func cleanIptablesPortWhiteList(entries, previous []PortWhiteListEntry) error {
	if len(previous) == 0 {
		return nil
	}
	keys := portWhiteListKeys(entries)
	for _, item := range previous {
		if _, ok := keys[item.key()]; ok {
			continue
		}
		if !iptables.CheckRuleExist(iptables.FilterTab, iptables.Chain1PanelBasic, "-p", item.Protocol, "--dport", item.Port, "-j", "ACCEPT") {
			continue
		}
		if err := iptables.DeleteRule(iptables.FilterTab, iptables.Chain1PanelBasic, "-p", item.Protocol, "--dport", item.Port, "-j", "ACCEPT"); err != nil {
			return err
		}
	}
	return nil
}

func cleanExtraIptablesPortRules(chain, protocol string, whitelist map[string]struct{}) error {
	rules, err := iptables.ReadFilterRulesByChain(chain)
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
		if err := iptables.DeleteRule(iptables.FilterTab, chain, "-p", protocol, "-m", protocol, "--dport", rule.DstPort, "-j", "ACCEPT"); err != nil {
			return err
		}
	}
	return nil
}
