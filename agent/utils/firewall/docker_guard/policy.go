package docker_guard

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"sort"
	"strconv"
	"strings"
)

var ErrInvalidPolicy = errors.New("invalid Docker port guard request")

func NormalizePolicy(policy Policy) (Policy, error) {
	policy.Family = strings.ToLower(strings.TrimSpace(policy.Family))
	policy.HostIP = strings.TrimSpace(policy.HostIP)
	policy.Protocol = strings.ToLower(strings.TrimSpace(policy.Protocol))
	policy.Mode = strings.ToLower(strings.TrimSpace(policy.Mode))
	if policy.HostPort == 0 ||
		(policy.Protocol != "tcp" && policy.Protocol != "udp") ||
		(policy.Family != FamilyIPv4 && policy.Family != FamilyIPv6) ||
		(policy.Mode != ModeAll && policy.Mode != ModeSources && policy.Mode != ModeAllow) {
		return Policy{}, fmt.Errorf("%w: invalid policy fields", ErrInvalidPolicy)
	}
	address, err := netip.ParseAddr(policy.HostIP)
	if err != nil || (policy.Family == FamilyIPv4) != address.Is4() {
		return Policy{}, fmt.Errorf("%w: host IP does not match address family", ErrInvalidPolicy)
	}
	normalizedSources := make([]string, 0, len(policy.Sources))
	seen := make(map[string]struct{}, len(policy.Sources))
	for _, source := range policy.Sources {
		source = strings.TrimSpace(source)
		if source == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(source)
		if err != nil {
			if sourceAddress, addressErr := netip.ParseAddr(source); addressErr == nil {
				bits := 128
				if sourceAddress.Is4() {
					bits = 32
				}
				prefix = netip.PrefixFrom(sourceAddress, bits)
			} else {
				return Policy{}, fmt.Errorf("%w: invalid source address %q", ErrInvalidPolicy, source)
			}
		}
		if (policy.Family == FamilyIPv4) != prefix.Addr().Is4() {
			return Policy{}, fmt.Errorf("%w: source %q does not match address family", ErrInvalidPolicy, source)
		}
		canonical := prefix.Masked().String()
		if _, exists := seen[canonical]; !exists {
			seen[canonical] = struct{}{}
			normalizedSources = append(normalizedSources, canonical)
		}
	}
	if policy.Mode != ModeAll && len(normalizedSources) == 0 {
		return Policy{}, fmt.Errorf("%w: source-based modes require at least one source", ErrInvalidPolicy)
	}
	if policy.Mode == ModeAll {
		normalizedSources = []string{}
	}
	sort.Strings(normalizedSources)
	policy.Sources = normalizedSources
	return policy, nil
}

func NormalizePolicyUUIDs(values []string) ([]string, error) {
	uuids := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, policyUUID := range values {
		policyUUID = strings.TrimSpace(policyUUID)
		if policyUUID == "" {
			return nil, fmt.Errorf("%w: policy UUID cannot be empty", ErrInvalidPolicy)
		}
		if _, exists := seen[policyUUID]; exists {
			continue
		}
		seen[policyUUID] = struct{}{}
		uuids = append(uuids, policyUUID)
	}
	if len(uuids) == 0 {
		return nil, fmt.Errorf("%w: policy UUIDs cannot be empty", ErrInvalidPolicy)
	}
	return uuids, nil
}

func PolicySyncKey(policy Policy) string {
	mode := policy.Mode
	if mode == ModeAllow && len(policy.Sources) == 0 {
		mode = ModeAll
	}
	sources := append([]string(nil), policy.Sources...)
	sort.Strings(sources)
	return strings.Join([]string{
		policy.UUID, policy.Family, CanonicalHost(policy.HostIP), strconv.Itoa(int(policy.HostPort)),
		policy.Protocol, mode, strings.Join(sources, ","),
	}, "\x00")
}

func PolicyStatesEqual(left, right []Policy) bool {
	if len(left) != len(right) {
		return false
	}
	counts := make(map[string]int, len(left))
	for _, policy := range left {
		counts[PolicySyncKey(policy)]++
	}
	for _, policy := range right {
		key := PolicySyncKey(policy)
		if counts[key] == 0 {
			return false
		}
		counts[key]--
	}
	return true
}

func CanonicalHost(value string) string {
	if address, err := netip.ParseAddr(value); err == nil {
		return address.String()
	}
	return value
}

func DecodeSources(value string) []string {
	result := []string{}
	_ = json.Unmarshal([]byte(value), &result)
	return result
}
