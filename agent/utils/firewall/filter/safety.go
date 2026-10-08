package filter

import (
	"errors"
	"slices"
	"strings"
)

var (
	ErrRuleStale     = errors.New("firewall rule state is stale")
	ErrRuleOperation = errors.New("firewall rule operation is not allowed")
)

func ProtectRuleSet(snapshot RuleSet, ports []PortWhitelist) (RuleSet, error) {
	rules := slices.Clone(snapshot.Rules)
	whitelist := NewPortWhitelistIndex(ports)
	for index := range rules {
		rules[index].Protected = rules[index].Protected || rules[index].ParseStatus == ParseStatusSupported && whitelist.Matches(rules[index].Rule)
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

func (index PortWhitelistIndex) Matches(rule FirewallRule) bool {
	rule, err := NormalizeRule(rule)
	if err != nil || rule.Action != ActionAccept || rule.SourcePort != "" || rule.DestinationAddress != "" || rule.Interface != "" || len(rule.ConnectionStates) != 0 {
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

func LocateRule(snapshot RuleSet, locator *Locator) (ObservedRule, error) {
	if locator == nil || locator.Provider != snapshot.Scope.Provider || locator.ScopeKey != snapshot.Scope.Key() {
		return ObservedRule{}, ErrInvalidRule
	}
	var matches []ObservedRule
	for _, observed := range snapshot.Rules {
		if SameLocator(observed.Locator, *locator) {
			matches = append(matches, observed)
		}
	}
	if len(matches) != 1 {
		return ObservedRule{}, ErrRuleStale
	}
	if err := GuardMutation(matches[0]); err != nil {
		return ObservedRule{}, err
	}
	return matches[0], nil
}
