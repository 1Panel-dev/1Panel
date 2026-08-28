package filter

import (
	"errors"
	"fmt"
	"strings"
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
	protected, err := NewSnapshot(snapshot.Scope, rules)
	if err != nil {
		return Snapshot{}, err
	}
	protected.Notices = append([]ScopeNotice(nil), snapshot.Notices...)
	return protected, nil
}

func GuardMutation(
	snapshot Snapshot,
	target ObservedRule,
	after FirewallRule,
	clientIP string,
	protectedPorts ...PortWhitelist,
) error {
	if RuleBlocksManagementConnection(after, clientIP, protectedPorts...) {
		return ErrLockoutRisk
	}
	for _, observed := range snapshot.Rules {
		if !observed.Protected || SameLocator(observed.Locator, target.Locator) {
			continue
		}
		if RulesOverlap(observed.Rule, after) && observed.Rule.Action != after.Action {
			return ErrLockoutRisk
		}
	}
	return nil
}

func SameLocator(left, right Locator) bool {
	if left.ScopeKey != right.ScopeKey {
		return false
	}
	if left.Canonical != "" || right.Canonical != "" {
		return left.Canonical != "" && left.Canonical == right.Canonical
	}
	return left.Position != nil && right.Position != nil && *left.Position == *right.Position
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
