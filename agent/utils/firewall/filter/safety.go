package filter

import (
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
)

var (
	ErrRuleConflict  = errors.New("firewall rule has identical conditions and an opposing action")
	ErrRuleStale     = errors.New("firewall rule state is stale")
	ErrRuleOperation = errors.New("firewall rule operation is not allowed")
)

var ErrVerificationFailed = errors.New("firewall rule verification failed")

func ProtectSnapshot(snapshot Snapshot, ports []PortWhitelist) (Snapshot, error) {
	rules := append([]ObservedRule(nil), snapshot.Rules...)
	for index := range rules {
		rule := rules[index].Rule
		if rules[index].ParseStatus != ParseStatusSupported || rule.Action != ActionAccept ||
			rule.SourceAddress != "" || rule.SourcePort != "" || rule.DestinationAddress != "" || rule.Interface != "" {
			continue
		}
		protected := false
		for _, protectedPort := range ports {
			family := strings.ToLower(strings.TrimSpace(protectedPort.Family))
			if family != "" && rule.Scope.Family != FamilyInet && string(rule.Scope.Family) != family {
				continue
			}
			if rule.Protocol != "all" && rule.Protocol != protectedPort.Protocol {
				continue
			}
			if portCovers(rule.DestinationPort, protectedPort.Port) {
				protected = true
				break
			}
		}
		if protected {
			rules[index].Protected = true
		}
	}
	protected := snapshot
	protected.Rules = rules
	if protected.Revision == "" {
		var err error
		protected, err = NewSnapshot(snapshot.Scope, rules)
		if err != nil {
			return Snapshot{}, err
		}
	}
	protected.Notices = append([]ScopeNotice(nil), snapshot.Notices...)
	return protected, nil
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

func ManagedObserved(snapshot Snapshot, desired DesiredRule) (ObservedRule, error) {
	items, err := MergeInventory(InventoryMergeInput{Observed: snapshot.Rules, Desired: []DesiredRule{desired}})
	if err != nil {
		return ObservedRule{}, err
	}
	for _, item := range items {
		if item.Desired == nil || item.Desired.UUID != desired.UUID {
			continue
		}
		if item.State == InventoryStateProtected || (item.Observed != nil && item.Observed.Protected) {
			return ObservedRule{}, ErrProtectedRule
		}
		if item.Observed == nil || item.Match != InventoryMatchExact || item.State == InventoryStateDrifted {
			return ObservedRule{}, ErrRuleStale
		}
		return *item.Observed, nil
	}
	return ObservedRule{}, ErrRuleStale
}

func FindCommittedObserved(snapshot Snapshot, requested FirewallRule, plan BackendPlan) (ObservedRule, error) {
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
	return familiesOverlap(left.Scope.Family, right.Scope.Family) &&
		protocolsOverlap(left.Protocol, right.Protocol) &&
		addressesOverlap(left.SourceAddress, right.SourceAddress) &&
		addressesOverlap(left.DestinationAddress, right.DestinationAddress) &&
		portsOverlap(left.SourcePort, right.SourcePort) &&
		portsOverlap(left.DestinationPort, right.DestinationPort) &&
		(left.Interface == "" || right.Interface == "" || left.Interface == right.Interface)
}

func familiesOverlap(left, right Family) bool {
	return left == FamilyInet || right == FamilyInet || left == right
}

func protocolsOverlap(left, right string) bool {
	return left == "all" || right == "all" || left == right
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

func portCovers(existing, requested string) bool {
	if existing == "" {
		return true
	}
	if requested == "" {
		return false
	}
	existingIntervals, err := portIntervals(existing)
	if err != nil {
		return false
	}
	requestedIntervals, err := portIntervals(requested)
	if err != nil {
		return false
	}
	for _, requestedInterval := range requestedIntervals {
		covered := false
		for _, existingInterval := range existingIntervals {
			if existingInterval[0] <= requestedInterval[0] && existingInterval[1] >= requestedInterval[1] {
				covered = true
				break
			}
		}
		if !covered {
			return false
		}
	}
	return true
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
