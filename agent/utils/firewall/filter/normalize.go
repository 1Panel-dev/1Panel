package filter

import (
	"fmt"
	"net/netip"
	"sort"
	"strconv"
	"strings"
)

const MaxAtomicExpansion = 256

func NormalizeRule(rule FirewallRule) (FirewallRule, error) {
	rule.Scope = rule.Scope.Normalize()
	if err := rule.Scope.ValidateMVP(); err != nil {
		return FirewallRule{}, err
	}

	if hasCompositeValue(rule.SourceAddress) || hasCompositeValue(rule.DestinationAddress) ||
		hasCompositeValue(rule.SourcePort) ||
		(hasCompositeValue(rule.DestinationPort) && !supportsNativeDestinationPortSet(rule.Scope.Provider)) ||
		isCompositeProtocol(rule.Protocol) {
		return FirewallRule{}, fmt.Errorf("%w: expand addresses, ports and protocols before normalization", ErrCompositeRule)
	}

	protocol, err := normalizeProtocol(rule.Protocol)
	if err != nil {
		return FirewallRule{}, err
	}
	rule.Protocol = protocol

	rule.SourceAddress, err = normalizeAddress(rule.SourceAddress, rule.Scope.Family)
	if err != nil {
		return FirewallRule{}, fmt.Errorf("%w: source address: %v", ErrInvalidRule, err)
	}
	rule.DestinationAddress, err = normalizeAddress(rule.DestinationAddress, rule.Scope.Family)
	if err != nil {
		return FirewallRule{}, fmt.Errorf("%w: destination address: %v", ErrInvalidRule, err)
	}
	rule.SourcePort, err = normalizePort(rule.SourcePort)
	if err != nil {
		return FirewallRule{}, fmt.Errorf("%w: source port: %v", ErrInvalidRule, err)
	}
	rule.DestinationPort, err = normalizePortValue(rule.DestinationPort, supportsNativeDestinationPortSet(rule.Scope.Provider))
	if err != nil {
		return FirewallRule{}, fmt.Errorf("%w: destination port: %v", ErrInvalidRule, err)
	}
	ufwAllProtocolDestinationPort := rule.Scope.Provider == ProviderUFW && protocol == "all" &&
		rule.SourcePort == "" && rule.DestinationPort != "" && !strings.Contains(rule.DestinationPort, ",")
	if (rule.SourcePort != "" || rule.DestinationPort != "") && protocol != "tcp" && protocol != "udp" &&
		!ufwAllProtocolDestinationPort {
		return FirewallRule{}, fmt.Errorf("%w: ports require tcp or udp protocol", ErrInvalidRule)
	}

	action, err := normalizeAction(rule.Action)
	if err != nil {
		return FirewallRule{}, err
	}
	rule.Action = action

	rule.NativeKind = NativeKind(strings.ToLower(strings.TrimSpace(string(rule.NativeKind))))
	if rule.NativeKind == "" {
		switch rule.Scope.Provider {
		case ProviderUFW:
			rule.NativeKind = NativeKindUFWRule
		default:
			rule.NativeKind = NativeKindRule
		}
	}
	if rule.Scope.Provider == ProviderFirewalld {
		switch rule.NativeKind {
		case NativeKindZonePort:
			rule.Priority = nil
			rule.OrderBucket = OrderBucketZonePrimitiveAllow
		case NativeKindRichRule:
			if rule.Priority == nil {
				priority := 0
				rule.Priority = &priority
			}
			if *rule.Priority < -32768 || *rule.Priority > 32767 {
				return FirewallRule{}, fmt.Errorf("%w: firewalld priority %d is out of range", ErrInvalidRule, *rule.Priority)
			}
			rule.OrderBucket = firewalldRichOrderBucket(*rule.Priority, rule.Action)
		}
	}
	rule.Interface = strings.TrimSpace(rule.Interface)
	if rule.Interface == "*" || strings.EqualFold(rule.Interface, "all") || strings.EqualFold(rule.Interface, "any") {
		rule.Interface = ""
	}
	rule.ConnectionStates, err = normalizeConnectionStates(rule.ConnectionStates, rule.Scope.Provider)
	if err != nil {
		return FirewallRule{}, err
	}
	rule.OrderBucket = strings.ToLower(strings.TrimSpace(rule.OrderBucket))
	rule.Description = strings.TrimSpace(rule.Description)
	return rule, nil
}

func firewalldRichOrderBucket(priority int, action Action) string {
	if priority < 0 {
		return OrderBucketRichPre
	}
	if priority > 0 {
		return OrderBucketRichPost
	}
	if action == ActionDrop || action == ActionReject {
		return OrderBucketRichZeroDeny
	}
	return OrderBucketRichZeroAllow
}

func ExpandAtomicRules(input FirewallRule) ([]FirewallRule, error) {
	input.Scope = input.Scope.Normalize()
	families := []Family{input.Scope.Family}
	expandedFamilies := input.Scope.Provider == ProviderUFW && input.Scope.Family == FamilyInet
	if expandedFamilies {
		families = []Family{FamilyIPv4, FamilyIPv6}
	}
	protocols := splitProtocols(input.Protocol)
	sourceAddresses := splitValues(input.SourceAddress)
	destinationAddresses := splitValues(input.DestinationAddress)
	sourcePorts := splitValues(input.SourcePort)
	destinationPorts := splitValues(input.DestinationPort)
	if supportsNativeDestinationPortSet(input.Scope.Provider) {
		destinationPorts = []string{input.DestinationPort}
	}

	count := len(families) * len(protocols) * len(sourceAddresses) * len(destinationAddresses) * len(sourcePorts) * len(destinationPorts)
	if count > MaxAtomicExpansion {
		return nil, fmt.Errorf("%w: %d rules exceeds maximum %d", ErrExpansionLimit, count, MaxAtomicExpansion)
	}

	rules := make([]FirewallRule, 0, count)
	seen := make(map[string]struct{}, count)
	var lastErr error
	for _, family := range families {
		for _, protocol := range protocols {
			for _, sourceAddress := range sourceAddresses {
				for _, destinationAddress := range destinationAddresses {
					for _, sourcePort := range sourcePorts {
						for _, destinationPort := range destinationPorts {
							rule := input
							rule.Scope.Family = family
							rule.Protocol = protocol
							rule.SourceAddress = sourceAddress
							rule.DestinationAddress = destinationAddress
							rule.SourcePort = sourcePort
							rule.DestinationPort = destinationPort
							normalized, err := NormalizeRule(rule)
							if err != nil {
								if expandedFamilies {
									lastErr = err
									continue
								}
								return nil, err
							}
							key, err := RuleKey(normalized)
							if err != nil {
								return nil, err
							}
							if _, exists := seen[key]; exists {
								continue
							}
							seen[key] = struct{}{}
							rules = append(rules, normalized)
						}
					}
				}
			}
		}
	}
	if len(rules) == 0 && lastErr != nil {
		return nil, lastErr
	}
	return rules, nil
}

func normalizeProtocol(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || value == "any" {
		return "all", nil
	}
	switch value {
	case "tcp", "udp", "icmp", "icmpv6", "all":
		return value, nil
	default:
		return "", fmt.Errorf("%w: unsupported protocol %q", ErrInvalidRule, value)
	}
}

func normalizeAction(value Action) (Action, error) {
	switch strings.ToLower(strings.TrimSpace(string(value))) {
	case "accept", "allow":
		return ActionAccept, nil
	case "drop", "deny", "block":
		return ActionDrop, nil
	case "reject":
		return ActionReject, nil
	default:
		return "", fmt.Errorf("%w: unsupported action %q", ErrInvalidRule, value)
	}
}

func normalizeAddress(value string, family Family) (string, error) {
	value = strings.TrimSpace(value)
	lower := strings.ToLower(value)
	if value == "" || lower == "any" || strings.HasPrefix(lower, "anywhere") {
		return "", nil
	}

	if prefix, err := netip.ParsePrefix(value); err == nil {
		address := prefix.Addr()
		bits := prefix.Bits()
		if address.Is4In6() {
			if bits < 96 {
				return "", fmt.Errorf("invalid IPv4-mapped prefix %q", value)
			}
			address = address.Unmap()
			bits -= 96
		}
		if err := validateAddressFamily(address, family); err != nil {
			return "", err
		}
		prefix = netip.PrefixFrom(address, bits).Masked()
		if prefix.Bits() == 0 {
			return "", nil
		}
		return prefix.String(), nil
	}

	address, err := netip.ParseAddr(value)
	if err != nil {
		return "", fmt.Errorf("invalid address %q", value)
	}
	address = address.Unmap()
	if err := validateAddressFamily(address, family); err != nil {
		return "", err
	}
	return netip.PrefixFrom(address, address.BitLen()).String(), nil
}

func validateAddressFamily(address netip.Addr, family Family) error {
	switch family {
	case FamilyIPv4:
		if !address.Is4() {
			return fmt.Errorf("address %q is not IPv4", address)
		}
	case FamilyIPv6:
		if !address.Is6() || address.Is4In6() {
			return fmt.Errorf("address %q is not IPv6", address)
		}
	case FamilyInet:
		return nil
	default:
		return fmt.Errorf("invalid family %q", family)
	}
	return nil
}

func normalizePort(value string) (string, error) {
	return normalizePortValue(value, false)
}

func normalizePortValue(value string, allowSet bool) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "any") || strings.EqualFold(value, "anywhere") {
		return "", nil
	}
	if strings.Contains(value, ",") {
		if !allowSet {
			return "", ErrCompositeRule
		}
		parts := strings.Split(value, ",")
		normalized := make([]string, 0, len(parts))
		seen := make(map[string]struct{}, len(parts))
		portSlots := 0
		for _, part := range parts {
			port, err := normalizePortValue(part, false)
			if err != nil {
				return "", err
			}
			if port == "" {
				return "", fmt.Errorf("invalid empty port in %q", value)
			}
			if _, exists := seen[port]; exists {
				continue
			}
			seen[port] = struct{}{}
			normalized = append(normalized, port)
			portSlots++
			if strings.Contains(port, "-") {
				portSlots++
			}
		}
		if portSlots > 15 {
			return "", fmt.Errorf("port set uses %d slots; maximum is 15", portSlots)
		}
		return strings.Join(normalized, ","), nil
	}

	separator := ""
	if strings.Contains(value, "-") {
		separator = "-"
	} else if strings.Contains(value, ":") {
		separator = ":"
	}
	if separator == "" {
		port, err := parsePort(value)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(port), nil
	}

	parts := strings.Split(value, separator)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid port range %q", value)
	}
	start, err := parsePort(parts[0])
	if err != nil {
		return "", err
	}
	end, err := parsePort(parts[1])
	if err != nil {
		return "", err
	}
	if start > end {
		return "", fmt.Errorf("invalid descending port range %q", value)
	}
	if start == end {
		return strconv.Itoa(start), nil
	}
	return fmt.Sprintf("%d-%d", start, end), nil
}

func supportsNativeDestinationPortSet(provider Provider) bool {
	return provider == ProviderIptables || provider == ProviderUFW
}

func parsePort(value string) (int, error) {
	port, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("invalid port %q", value)
	}
	return port, nil
}

func normalizeSet(values []string, lower bool) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if lower {
			value = strings.ToLower(value)
		}
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func normalizeConnectionStates(values []string, provider Provider) ([]string, error) {
	states := normalizeSet(values, true)
	if len(states) == 0 {
		return nil, nil
	}
	allowed := map[string]struct{}{
		"new": {}, "established": {}, "related": {}, "invalid": {}, "untracked": {},
	}
	if provider == ProviderIptables {
		allowed["snat"] = struct{}{}
		allowed["dnat"] = struct{}{}
	}
	for _, state := range states {
		if _, ok := allowed[state]; !ok {
			return nil, fmt.Errorf("%w: unsupported connection state %q", ErrInvalidRule, state)
		}
	}
	return states, nil
}

func hasCompositeValue(value string) bool {
	return strings.Contains(value, ",")
}

func isCompositeProtocol(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "tcp/udp" || value == "udp/tcp" || strings.Contains(value, ",")
}

func splitProtocols(value string) []string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "tcp/udp" || value == "udp/tcp" {
		return []string{"tcp", "udp"}
	}
	return splitValues(value)
}

func splitValues(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if _, exists := seen[part]; exists {
			continue
		}
		seen[part] = struct{}{}
		result = append(result, part)
	}
	if len(result) == 0 {
		return []string{""}
	}
	return result
}
