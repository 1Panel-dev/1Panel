package filter

import (
	"errors"
	"fmt"
	"net/netip"
	"slices"
	"strconv"
	"strings"
)

var (
	ErrRuleConflict  = errors.New("firewall rule has identical conditions and an opposing action")
	ErrRuleStale     = errors.New("firewall rule state is stale")
	ErrRuleOperation = errors.New("firewall rule operation is not allowed")
)

var ErrVerificationFailed = errors.New("firewall rule verification failed")

func ProtectRuleSet(snapshot RuleSet, ports []PortWhitelist) (RuleSet, error) {
	rules := slices.Clone(snapshot.Rules)
	whitelist := NewPortWhitelistIndex(ports)
	for index := range rules {
		if rules[index].ParseStatus == ParseStatusSupported && whitelist.Matches(rules[index].Rule) {
			rules[index].Protected = true
		}
	}
	protected := snapshot
	protected.Rules = rules

	protected.LastPosition = snapshot.LastPosition
	protected.Notices = append([]ScopeNotice(nil), snapshot.Notices...)
	return protected, nil
}

type portWhitelistKey struct {
	family                 Family
	protocol, port, source string
}

type PortWhitelistIndex map[portWhitelistKey]bool

func NewPortWhitelistIndex(ports []PortWhitelist) PortWhitelistIndex {
	index := make(PortWhitelistIndex)
	for _, port := range ports {
		protocol, err := normalizeProtocol(port.Protocol)
		if err != nil {
			continue
		}
		portRange, err := normalizePortValue(port.Port, false)
		if err != nil {
			continue
		}
		portFamily := Family(strings.ToLower(strings.TrimSpace(port.Family)))
		sources := port.Sources
		if len(sources) == 0 {
			sources = []string{""}
		}
		for _, family := range []Family{FamilyIPv4, FamilyIPv6} {
			if portFamily != "" && portFamily != FamilyInet && portFamily != family {
				continue
			}
			for _, source := range sources {
				normalized, err := normalizeAddress(source, family)
				if err == nil {
					index[portWhitelistKey{family, protocol, portRange, normalized}] = true
				}
			}
		}
	}
	return index
}

func RuleMatchesPortWhitelist(rule FirewallRule, ports []PortWhitelist) bool {
	return NewPortWhitelistIndex(ports).Matches(rule)
}

func (index PortWhitelistIndex) Matches(rule FirewallRule) bool {
	rule, err := NormalizeRule(rule)
	if err != nil || rule.Action != ActionAccept || rule.SourcePort != "" || rule.DestinationAddress != "" || rule.Interface != "" || len(rule.ConnectionStates) != 0 {
		return false
	}
	if rule.Scope.Provider == ProviderFirewalld && (rule.NativeKind == NativeKindZonePort ||
		(rule.NativeKind == NativeKindRule && rule.Scope.Family == FamilyInet && rule.Priority == nil)) {
		return false
	}
	families := []Family{rule.Scope.Family}
	if rule.Scope.Family == FamilyInet {
		families = []Family{FamilyIPv4, FamilyIPv6}
	}
	for _, family := range families {
		if !index[portWhitelistKey{family, rule.Protocol, rule.DestinationPort, rule.SourceAddress}] {
			return false
		}
	}
	return true
}

func IsBuiltinProtectedRule(rule FirewallRule) bool {
	if rule.Scope.Provider != ProviderIptables && rule.Scope.Provider != ProviderNftables {
		return false
	}
	rule, err := NormalizeRule(rule)
	if err != nil || rule.SourceAddress != "" || rule.DestinationAddress != "" || rule.SourcePort != "" || rule.DestinationPort != "" {
		return false
	}
	switch rule.Scope.Chain {
	case BasicBeforeChain:
		if rule.Action != ActionAccept || rule.Protocol != "all" {
			return false
		}
		return rule.Interface == "lo" && len(rule.ConnectionStates) == 0 ||
			rule.Interface == "" && slices.Equal(rule.ConnectionStates, []string{"established", "related"})
	case BasicAfterChain:
		return rule.Action == ActionDrop && (rule.Protocol == "tcp" || rule.Protocol == "udp") &&
			rule.Interface == "" && len(rule.ConnectionStates) == 0
	}
	return false
}

func GuardMutation(target ObservedRule) error {
	if target.Protected {
		return ErrProtectedRule
	}
	return nil
}

func SameLocator(left, right Locator) bool {
	if left.ScopeKey != right.ScopeKey {
		return false
	}
	if left.Provider != "" && right.Provider != "" && left.Provider != right.Provider {
		return false
	}
	if left.Position != nil || right.Position != nil {
		if left.Position == nil || right.Position == nil || *left.Position != *right.Position {
			return false
		}
		if left.NativeID != "" && right.NativeID != "" && left.NativeID != right.NativeID {
			return false
		}
		return left.Canonical == "" || right.Canonical == "" || left.Canonical == right.Canonical
	}
	return left.Canonical != "" && left.Canonical == right.Canonical
}

func MatchObservedByRuleKey(observed []ObservedRule, rule FirewallRule) ([]ObservedRule, error) {
	wanted, err := RuleKey(rule)
	if err != nil {
		return nil, err
	}
	matches := make([]ObservedRule, 0, 1)
	for _, candidate := range observed {
		if candidate.ParseStatus != ParseStatusSupported {
			continue
		}
		candidateKey, keyErr := RuleKey(candidate.Rule)
		if keyErr == nil && candidateKey == wanted {
			matches = append(matches, candidate)
		}
	}
	return matches, nil
}

func FindCommittedObserved(snapshot RuleSet, requested FirewallRule, plan CommandBatch) (ObservedRule, error) {
	if len(plan.Rules) == 1 && plan.Rules[0].Expected.Marker != "" {
		matches := make([]ObservedRule, 0, 1)
		for _, observed := range snapshot.Rules {
			if observed.Marker == plan.Rules[0].Expected.Marker {
				matches = append(matches, observed)
			}
		}
		if len(matches) == 1 {
			return matches[0], nil
		}
	}
	matches, err := MatchObservedByRuleKey(snapshot.Rules, requested)
	if err != nil {
		return ObservedRule{}, err
	}
	if len(matches) != 1 {
		return ObservedRule{}, fmt.Errorf("%w: expected one committed rule, found %d", ErrVerificationFailed, len(matches))
	}
	return matches[0], nil
}

func RulesOverlap(left, right FirewallRule) bool {
	left, leftErr := NormalizeRule(left)
	right, rightErr := NormalizeRule(right)
	if leftErr != nil || rightErr != nil || left.Scope.Key() != right.Scope.Key() {
		return false
	}
	return (left.Scope.Family == FamilyInet || right.Scope.Family == FamilyInet || left.Scope.Family == right.Scope.Family) &&
		(left.Protocol == "all" || right.Protocol == "all" || left.Protocol == right.Protocol) &&
		addressesOverlap(left.SourceAddress, right.SourceAddress) &&
		addressesOverlap(left.DestinationAddress, right.DestinationAddress) &&
		portsOverlap(left.SourcePort, right.SourcePort) &&
		portsOverlap(left.DestinationPort, right.DestinationPort) &&
		(left.Interface == "" || right.Interface == "" || left.Interface == right.Interface)
}

func addressesOverlap(left, right string) bool {
	if left == "" || right == "" {
		return true
	}
	leftPrefix, leftErr := netip.ParsePrefix(left)
	rightPrefix, rightErr := netip.ParsePrefix(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	return leftPrefix.Contains(rightPrefix.Addr()) || rightPrefix.Contains(leftPrefix.Addr())
}

func portsOverlap(left, right string) bool {
	if left == "" || right == "" {
		return true
	}
	leftIntervals, leftErr := portIntervals(left)
	rightIntervals, rightErr := portIntervals(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	for _, leftInterval := range leftIntervals {
		for _, rightInterval := range rightIntervals {
			if leftInterval[0] <= rightInterval[1] && rightInterval[0] <= leftInterval[1] {
				return true
			}
		}
	}
	return false
}

func portIntervals(value string) ([][2]int, error) {
	parts := strings.Split(value, ",")
	intervals := make([][2]int, 0, len(parts))
	for _, part := range parts {
		start, end, err := portInterval(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		intervals = append(intervals, [2]int{start, end})
	}
	return intervals, nil
}

func portInterval(value string) (int, int, error) {
	parts := strings.Split(value, "-")
	if len(parts) == 1 {
		port, err := strconv.Atoi(parts[0])
		return port, port, err
	}
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid port interval %q", value)
	}
	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	end, err := strconv.Atoi(parts[1])
	return start, end, err
}
