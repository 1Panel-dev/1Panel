package client

import "strings"

// PortUnit is one native operation a provider runs for a single logical port rule.
type PortUnit struct {
	Apply  FireInfo // rule handed to the provider command
	Record FireInfo // rule persisted in the 1Panel record, differs from Apply on ufw
	Chain  string   // owning 1PANEL chain, empty on providers without managed chains
}

// AddressUnit is the address rule counterpart of PortUnit. Applied and recorded
// shapes never differ for address rules, so there is no Record here.
type AddressUnit struct {
	Apply FireInfo
	Chain string
}

// splitRuleAddresses always yields at least one entry so a rule without a source
// still expands into a single "any source" unit.
func splitRuleAddresses(address string) []string {
	parts := strings.Split(strings.TrimSuffix(address, ","), ",")
	addresses := make([]string, 0, len(parts))
	for _, part := range parts {
		addresses = append(addresses, normalizeRuleAddress(part))
	}
	if len(addresses) == 0 {
		return []string{""}
	}
	return addresses
}

// normalizeRuleAddress maps the ufw "Anywhere" display value back to an empty source.
func normalizeRuleAddress(address string) string {
	address = strings.TrimSpace(address)
	if strings.EqualFold(address, "Anywhere") {
		return ""
	}
	return address
}

// expandPortRule is shared by firewalld and iptables: protocols x (port range or
// comma list) x addresses. A range is passed through untouched, everything else
// is split on commas.
func expandPortRule(rule FireInfo, chain string) []PortUnit {
	addresses := splitRuleAddresses(rule.Address)
	var units []PortUnit
	for _, protocol := range strings.Split(rule.Protocol, "/") {
		ports := []string{rule.Port}
		if !strings.Contains(rule.Port, "-") {
			ports = strings.Split(rule.Port, ",")
		}
		for _, port := range ports {
			if len(port) == 0 {
				continue
			}
			for _, address := range addresses {
				item := rule
				item.Port = port
				item.Protocol = protocol
				item.Address = address
				units = append(units, PortUnit{Apply: item, Record: item, Chain: chain})
			}
		}
	}
	return units
}

// expandUfwPortRule keeps the two ufw specifics: a port list or range is applied
// with ":" but recorded with "-", and a single port drops the protocol so that
// tcp/udp becomes one rule instead of two.
func expandUfwPortRule(rule FireInfo) []PortUnit {
	addresses := splitRuleAddresses(rule.Address)
	var units []PortUnit
	if strings.Contains(rule.Port, ",") || strings.Contains(rule.Port, "-") {
		for _, protocol := range strings.Split(rule.Protocol, "/") {
			for _, address := range addresses {
				apply, record := rule, rule
				apply.Protocol, record.Protocol = protocol, protocol
				apply.Port = strings.ReplaceAll(rule.Port, "-", ":")
				record.Port = strings.ReplaceAll(rule.Port, ":", "-")
				apply.Address, record.Address = address, ufwRecordAddress(address)
				units = append(units, PortUnit{Apply: apply, Record: record, Chain: rule.Chain})
			}
		}
		return units
	}
	for _, address := range addresses {
		apply, record := rule, rule
		if rule.Protocol == "tcp/udp" {
			apply.Protocol = ""
		}
		apply.Address, record.Address = address, ufwRecordAddress(address)
		units = append(units, PortUnit{Apply: apply, Record: record, Chain: rule.Chain})
	}
	return units
}

func ufwRecordAddress(address string) string {
	if len(address) == 0 {
		return "Anywhere"
	}
	return address
}

// expandAddressRule is shared by every provider: one native rule per source,
// empty entries are dropped instead of becoming an "any source" rule.
func expandAddressRule(rule FireInfo, chain string) []AddressUnit {
	var units []AddressUnit
	for _, address := range strings.Split(rule.Address, ",") {
		if len(address) == 0 {
			continue
		}
		item := rule
		item.Address = address
		units = append(units, AddressUnit{Apply: item, Chain: chain})
	}
	return units
}

// needsRichRule is the firewalld/iptables choice between the port shortcut and a
// full rule; ufw has its own rule in ufwNeedsRichRule.
func needsRichRule(rule FireInfo) bool {
	return len(rule.Address) != 0 || rule.Strategy == "drop"
}

// ufwNeedsRichRule is the ufw variant: ufw denies a port through the port
// shortcut as well, only a source forces the longer form.
func ufwNeedsRichRule(rule FireInfo) bool {
	return len(rule.Address) != 0 && !strings.EqualFold(rule.Address, "Anywhere")
}
