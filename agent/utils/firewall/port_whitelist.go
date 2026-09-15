package firewall

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"sort"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

type PortWhitelist = filter.PortWhitelist

const (
	PortWhitelistTypePanel = "panel"
	PortWhitelistTypeSSH   = "ssh"
)

func ParsePortWhitelist(value string) ([]PortWhitelist, error) {
	var rules []PortWhitelist
	if err := json.Unmarshal([]byte(value), &rules); err != nil {
		return nil, err
	}
	return ValidatePortWhitelist(rules)
}

func ValidatePortWhitelist(rules []PortWhitelist) ([]PortWhitelist, error) {
	if rules == nil {
		return nil, fmt.Errorf("firewall port whitelist must be an array")
	}
	result := make([]PortWhitelist, 0, len(rules))
	seen := make(map[string]bool, len(rules))
	for _, rule := range rules {
		rule.Type = strings.ToLower(strings.TrimSpace(rule.Type))
		rule.Protocol = strings.ToLower(strings.TrimSpace(rule.Protocol))
		if rule.Type != "" && rule.Protocol == "" {
			rule.Protocol = "tcp"
		}
		if rule.Protocol != "tcp" && rule.Protocol != "udp" {
			return nil, fmt.Errorf("invalid firewall port whitelist protocol: %s", rule.Protocol)
		}
		if rule.Type != "" {
			if rule.Type != PortWhitelistTypePanel && rule.Type != PortWhitelistTypeSSH {
				return nil, fmt.Errorf("invalid firewall port whitelist type: %s", rule.Type)
			}
			if rule.Port != "" {
				port, err := parseWhitelistPort(rule.Port)
				if err != nil {
					return nil, err
				}
				rule.Port = strconv.Itoa(port)
			}
		} else {
			var err error
			rule.Port, err = normalizeWhitelistPort(rule.Port)
			if err != nil {
				return nil, err
			}
		}
		if len(rule.Sources) == 0 {
			return nil, fmt.Errorf("firewall port whitelist requires at least one source")
		}
		var err error
		rule.Sources, err = NormalizeWhitelistSources("", rule.Sources)
		if err != nil {
			return nil, err
		}
		rule.Family = ""
		key := rule.Type + "/" + rule.Protocol + "/" + rule.Port
		if rule.Type != "" {
			key = rule.Type + "/" + rule.Protocol
		}
		if seen[key] {
			return nil, fmt.Errorf("duplicate firewall port whitelist: %s", key)
		}
		seen[key] = true
		result = append(result, rule)
	}
	return result, nil
}

func normalizeWhitelistPort(value string) (string, error) {
	value = strings.TrimSpace(value)
	separator := ""
	if strings.Contains(value, "-") {
		separator = "-"
	} else if strings.Contains(value, ":") {
		separator = ":"
	}
	if separator == "" {
		port, err := parseWhitelistPort(value)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(port), nil
	}
	parts := strings.Split(value, separator)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid firewall port whitelist range: %s", value)
	}
	start, err := parseWhitelistPort(parts[0])
	if err != nil {
		return "", err
	}
	end, err := parseWhitelistPort(parts[1])
	if err != nil || start > end {
		return "", fmt.Errorf("invalid firewall port whitelist range: %s", value)
	}
	if start == end {
		return strconv.Itoa(start), nil
	}
	return fmt.Sprintf("%d-%d", start, end), nil
}

func parseWhitelistPort(value string) (int, error) {
	port, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("invalid firewall port whitelist: %s", value)
	}
	return port, nil
}

func NormalizePortWhitelist(items []PortWhitelist) []PortWhitelist {
	ports := make([]PortWhitelist, 0, len(items))
	for _, item := range items {
		if item.Port == "" {
			continue
		}
		baseKey := strings.Join([]string{item.Port, item.Protocol, strings.Join(item.Sources, ",")}, "/")
		duplicate := false
		for _, current := range ports {
			currentBaseKey := strings.Join([]string{current.Port, current.Protocol, strings.Join(current.Sources, ",")}, "/")
			if currentBaseKey == baseKey && (current.Family == "" || current.Family == item.Family) {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}
		if item.Family == "" {
			filtered := ports[:0]
			for _, current := range ports {
				currentBaseKey := strings.Join([]string{current.Port, current.Protocol, strings.Join(current.Sources, ",")}, "/")
				if currentBaseKey != baseKey {
					filtered = append(filtered, current)
				}
			}
			ports = filtered
		}
		ports = append(ports, item)
	}
	return ports
}

func NormalizeWhitelistSources(family string, sources []string) ([]string, error) {
	result := make([]string, 0, len(sources))
	seen := make(map[string]struct{}, len(sources))
	for _, source := range sources {
		source = strings.TrimSpace(source)
		prefix, err := netip.ParsePrefix(source)
		if err != nil {
			address, err := netip.ParseAddr(source)
			if err != nil {
				return nil, err
			}
			prefix = netip.PrefixFrom(address, address.BitLen())
		}
		sourceFamily := family
		if sourceFamily == "" {
			sourceFamily = constant.FirewallFamilyIPv6
			if prefix.Addr().Unmap().Is4() {
				sourceFamily = constant.FirewallFamilyIPv4
			}
		}
		rule, err := filter.NormalizeRule(RuleForSystemPort(filter.ProviderIptables, SystemPort{
			Family: sourceFamily, Port: "1", Protocol: "tcp", SourceAddress: source,
		}))
		if err != nil {
			return nil, err
		}
		if rule.SourceAddress == "" {
			rule.SourceAddress = "0.0.0.0/0"
			if sourceFamily == constant.FirewallFamilyIPv6 {
				rule.SourceAddress = "::/0"
			}
		}
		if _, exists := seen[rule.SourceAddress]; !exists {
			seen[rule.SourceAddress] = struct{}{}
			result = append(result, rule.SourceAddress)
		}
	}
	return result, nil
}

type SystemPort struct {
	Family        string
	Port          string
	Protocol      string
	SourceAddress string
}

func ExpandPortWhitelist(ports []PortWhitelist) []SystemPort {
	result := make([]SystemPort, 0, len(ports))
	for _, port := range ports {
		sources := port.Sources
		if len(sources) == 0 {
			sources = []string{"0.0.0.0/0", "::/0"}
		}
		for _, source := range sources {
			family := constant.FirewallFamilyIPv4
			if strings.Contains(source, ":") {
				family = constant.FirewallFamilyIPv6
			}
			if port.Family != "" && port.Family != family {
				continue
			}
			if source == "0.0.0.0/0" || source == "::/0" {
				source = ""
			}
			result = append(result, SystemPort{Family: family, Port: port.Port, Protocol: port.Protocol, SourceAddress: source})
		}
	}
	return result
}

func RuleForSystemPort(provider filter.Provider, port SystemPort) filter.FirewallRule {
	scope := filter.Scope{Provider: provider, Direction: filter.DirectionInput}
	family := filter.Family(strings.ToLower(strings.TrimSpace(port.Family)))
	switch provider {
	case filter.ProviderIptables, filter.ProviderNftables:
		if family != filter.FamilyIPv6 {
			family = filter.FamilyIPv4
		}
		scope.Family, scope.Table = family, "filter"
	case filter.ProviderFirewalld:
		if family != filter.FamilyIPv4 && family != filter.FamilyIPv6 {
			family = filter.FamilyInet
		}
		scope.Family, scope.Zone = family, filter.FirewalldInputZone
	case filter.ProviderUFW:
		if family != filter.FamilyIPv6 {
			family = filter.FamilyIPv4
		}
		scope.Family = family
	}
	return filter.FirewallRule{
		Scope: scope, Protocol: port.Protocol, DestinationPort: port.Port,
		SourceAddress: port.SourceAddress, Action: filter.ActionAccept,
	}
}

func NormalizeSystemPorts(ports []SystemPort) (map[string]SystemPort, error) {
	result := make(map[string]SystemPort, len(ports))
	for _, port := range ports {
		normalized, err := filter.NormalizeRule(RuleForSystemPort(filter.ProviderIptables, port))
		if err != nil {
			return nil, err
		}
		family := strings.ToLower(strings.TrimSpace(port.Family))
		if family != "" {
			family = string(normalized.Scope.Family)
		}
		item := SystemPort{
			Family: family, Port: normalized.DestinationPort,
			Protocol: normalized.Protocol, SourceAddress: normalized.SourceAddress,
		}
		result[SystemPortKey(item)] = item
	}
	return result, nil
}

func SystemPortKey(port SystemPort) string {
	key := LegacySystemPortKey(port)
	if family := strings.ToLower(strings.TrimSpace(port.Family)); family != "" {
		key = family + "/" + key
	}
	if source := strings.TrimSpace(port.SourceAddress); source != "" {
		key += "/" + source
	}
	return key
}

func LegacySystemPortKey(port SystemPort) string {
	return strings.ToLower(strings.TrimSpace(port.Protocol)) + "/" + strings.TrimSpace(port.Port)
}

func SortedSystemPortKeys(ports map[string]SystemPort) []string {
	keys := make([]string, 0, len(ports))
	for key := range ports {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func ContainsPort(ports []PortWhitelist, target PortWhitelist) bool {
	for _, port := range ports {
		familyMatches := port.Family == "" || target.Family == "" || port.Family == target.Family
		if familyMatches && port.Port == target.Port && port.Protocol == target.Protocol {
			return true
		}
	}
	return false
}

func ExcludePorts(ports, excluded []PortWhitelist) []PortWhitelist {
	result := make([]PortWhitelist, 0, len(ports))
	for _, port := range ports {
		if !ContainsPort(excluded, port) {
			result = append(result, port)
		}
	}
	return result
}

func NormalizeRequiredPorts(ports []PortWhitelist) ([]PortWhitelist, error) {
	result := make([]PortWhitelist, 0, len(ports))
	for _, port := range ports {
		port.Protocol = strings.ToLower(strings.TrimSpace(port.Protocol))
		if port.Protocol != "tcp" && port.Protocol != "udp" {
			return nil, fmt.Errorf("unsupported required firewall port protocol %q", port.Protocol)
		}
		portNumber, err := strconv.Atoi(strings.TrimSpace(port.Port))
		if err != nil || portNumber < 1 || portNumber > 65535 {
			return nil, fmt.Errorf("invalid required firewall port %q", port.Port)
		}
		port.Port = strconv.Itoa(portNumber)
		if len(port.Sources) > 0 {
			port.Sources, err = NormalizeWhitelistSources(port.Family, port.Sources)
			if err != nil {
				return nil, err
			}
		}
		result = append(result, port)
	}
	return NormalizePortWhitelist(result), nil
}

func RequiredPortWhitelist(entries []PortWhitelist) ([]PortWhitelist, error) {
	result := make([]PortWhitelist, 0, len(entries))
	for _, entry := range entries {
		if entry.Type == "" {
			continue
		}
		if entry.Port == "" {
			return nil, fmt.Errorf("firewall whitelist %s has no stored port", entry.Type)
		}
		protocol := strings.ToLower(strings.TrimSpace(entry.Protocol))
		if protocol == "" {
			protocol = "tcp"
		}
		result = append(result, PortWhitelist{Family: entry.Family, Port: entry.Port, Protocol: protocol, Sources: entry.Sources})
	}
	return NormalizeRequiredPorts(result)
}
