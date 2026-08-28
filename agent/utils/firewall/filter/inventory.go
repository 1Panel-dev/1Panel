package filter

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
)

type RuleOrigin string

const (
	RuleOriginCreated RuleOrigin = constant.FirewallRuleOriginCreated
	RuleOriginAdopted RuleOrigin = constant.FirewallRuleOriginAdopted
)

type InventoryState string

const (
	InventoryStateManaged   InventoryState = "managed"
	InventoryStateAdopted   InventoryState = "adopted"
	InventoryStateExternal  InventoryState = "external"
	InventoryStateDrifted   InventoryState = "drifted"
	InventoryStateProtected InventoryState = "protected"
)

type InventoryMatch string

const (
	InventoryMatchNone      InventoryMatch = "none"
	InventoryMatchExact     InventoryMatch = "exact"
	InventoryMatchChanged   InventoryMatch = "changed"
	InventoryMatchMissing   InventoryMatch = "missing"
	InventoryMatchAmbiguous InventoryMatch = "ambiguous"
	InventoryMatchOpaque    InventoryMatch = "opaque"
)

type DesiredRule struct {
	UUID                string       `json:"uuid"`
	Rule                FirewallRule `json:"rule"`
	RuleKey             string       `json:"ruleKey"`
	Origin              RuleOrigin   `json:"origin"`
	Protected           bool         `json:"protected,omitempty"`
	Marker              string       `json:"marker,omitempty"`
	ObservedInstanceKey string       `json:"observedInstanceKey,omitempty"`
}

type RuntimeUsage struct {
	Used   bool     `json:"used"`
	UsedBy []string `json:"usedBy,omitempty"`
	Reason string   `json:"reason,omitempty"`
}

type InventoryItem struct {
	Rule     FirewallRule   `json:"rule"`
	Observed *ObservedRule  `json:"observed,omitempty"`
	Desired  *DesiredRule   `json:"desired,omitempty"`
	State    InventoryState `json:"state"`
	Match    InventoryMatch `json:"match"`
	Usage    *RuntimeUsage  `json:"usage,omitempty"`
}

type Inventory struct {
	Items   []InventoryItem `json:"items"`
	Notices []ScopeNotice   `json:"notices,omitempty"`
}

type InventoryMergeInput struct {
	Observed              []ObservedRule
	Desired               []DesiredRule
	ProtectedObservedKeys map[string]struct{}
}

type observedInventoryCandidate struct {
	rule        ObservedRule
	ruleKey     string
	instanceKey string
	claimed     bool
}

func MergeInventory(input InventoryMergeInput) ([]InventoryItem, error) {
	candidates := make([]observedInventoryCandidate, len(input.Observed))
	byRuleKey := make(map[string][]int)
	byInstanceKey := make(map[string][]int)
	byMarker := make(map[string][]int)
	for index, observed := range input.Observed {
		candidate := observedInventoryCandidate{rule: observed}
		if marker := strings.TrimSpace(candidate.rule.Marker); marker != "" {
			markerKey := candidate.rule.Rule.Scope.Key() + "\x00" + marker
			byMarker[markerKey] = append(byMarker[markerKey], index)
		}
		if observed.ParseStatus == ParseStatusSupported {
			normalized, err := NormalizeRule(observed.Rule)
			if err != nil {
				return nil, fmt.Errorf("normalize observed firewall rule %d: %w", index, err)
			}
			candidate.rule.Rule = normalized
			candidate.ruleKey, err = RuleKey(normalized)
			if err != nil {
				return nil, err
			}
			byRuleKey[candidate.ruleKey] = append(byRuleKey[candidate.ruleKey], index)
			if instanceKey, err := InstanceKey(candidate.rule); err == nil {
				candidate.instanceKey = instanceKey
				byInstanceKey[instanceKey] = append(byInstanceKey[instanceKey], index)
			}
		}
		candidates[index] = candidate
	}

	normalizedDesired := make([]DesiredRule, 0, len(input.Desired))
	desiredMatches := make(map[int]int)
	desiredMatchStates := make([]InventoryMatch, 0, len(input.Desired))
	for _, desired := range input.Desired {
		normalized, err := NormalizeRule(desired.Rule)
		if err != nil {
			return nil, fmt.Errorf("normalize desired firewall rule %q: %w", desired.UUID, err)
		}
		desired.Rule = normalized
		calculatedKey, err := RuleKey(normalized)
		if err != nil {
			return nil, err
		}
		if desired.RuleKey != "" && desired.RuleKey != calculatedKey {
			return nil, fmt.Errorf("%w: desired rule %q key does not match its semantics", ErrInvalidRule, desired.UUID)
		}
		desired.RuleKey = calculatedKey

		match, matchState := findObservedInventoryMatch(desired, candidates, byRuleKey, byInstanceKey, byMarker)
		normalizedIndex := len(normalizedDesired)
		normalizedDesired = append(normalizedDesired, desired)
		desiredMatchStates = append(desiredMatchStates, matchState)
		if match >= 0 {
			candidates[match].claimed = true
			desiredMatches[match] = normalizedIndex
		}
	}

	items := make([]InventoryItem, 0, len(candidates)+len(normalizedDesired))
	matchedDesired := make(map[int]struct{}, len(desiredMatches))
	for index := range candidates {
		candidate := &candidates[index]
		if desiredIndex, exists := desiredMatches[index]; exists {
			desired := normalizedDesired[desiredIndex]
			observed := candidate.rule
			match := desiredMatchStates[desiredIndex]
			if match == InventoryMatchExact && observed.ParseStatus != ParseStatusSupported {
				orderIndex := observed.Rule.OrderIndex
				observed.Rule = desired.Rule
				observed.Rule.OrderIndex = orderIndex
				observed.ParseStatus = ParseStatusSupported
				observed.UncertainFields = nil
			}
			displayRule := observed.Rule
			displayRule.Description = desired.Rule.Description
			state := inventoryStateForDesired(desired, match)
			if observed.Protected {
				state = InventoryStateProtected
			} else if observed.Persistence != "" && observed.Persistence != PersistenceStatusConverged {
				state = InventoryStateDrifted
			}
			items = append(items, InventoryItem{
				Rule:     displayRule,
				Observed: &observed,
				Desired:  &desired,
				State:    state,
				Match:    match,
			})
			matchedDesired[desiredIndex] = struct{}{}
			continue
		}
		observed := candidate.rule
		state := InventoryStateExternal
		if observed.Protected {
			state = InventoryStateProtected
		} else if _, protected := input.ProtectedObservedKeys[candidate.ruleKey]; protected {
			state = InventoryStateProtected
		}
		match := InventoryMatchNone
		if observed.ParseStatus != ParseStatusSupported {
			match = InventoryMatchOpaque
		}
		items = append(items, InventoryItem{Rule: observed.Rule, Observed: &observed, State: state, Match: match})
	}
	for index, desired := range normalizedDesired {
		if _, matched := matchedDesired[index]; matched {
			continue
		}
		desiredCopy := desired
		match := desiredMatchStates[index]
		items = append(items, InventoryItem{
			Rule:    desired.Rule,
			Desired: &desiredCopy,
			State:   inventoryStateForDesired(desired, match),
			Match:   match,
		})
	}
	return items, nil
}

func findObservedInventoryMatch(
	desired DesiredRule,
	candidates []observedInventoryCandidate,
	byRuleKey map[string][]int,
	byInstanceKey map[string][]int,
	byMarker map[string][]int,
) (int, InventoryMatch) {
	if marker := strings.TrimSpace(desired.Marker); marker != "" {
		markerKey := desired.Rule.Scope.Key() + "\x00" + marker
		match, status := uniqueUnclaimedCandidate(byMarker[markerKey], candidates)
		if match >= 0 && candidates[match].rule.ParseStatus != ParseStatusOpaque &&
			!ObservedRuleMatchesExpected(candidates[match].rule, desired.Rule) {
			return match, InventoryMatchChanged
		}
		return match, status
	}
	if desired.ObservedInstanceKey != "" {
		return uniqueUnclaimedCandidate(byInstanceKey[desired.ObservedInstanceKey], candidates)
	}
	return uniqueUnclaimedCandidate(byRuleKey[desired.RuleKey], candidates)
}

func uniqueUnclaimedCandidate(indices []int, candidates []observedInventoryCandidate) (int, InventoryMatch) {
	match := -1
	count := 0
	for _, index := range indices {
		if candidates[index].claimed {
			continue
		}
		match = index
		count++
	}
	switch count {
	case 0:
		return -1, InventoryMatchMissing
	case 1:
		return match, InventoryMatchExact
	default:
		return -1, InventoryMatchAmbiguous
	}
}

func inventoryStateForDesired(desired DesiredRule, match InventoryMatch) InventoryState {
	if match != InventoryMatchExact {
		return InventoryStateDrifted
	}
	if desired.Protected {
		return InventoryStateProtected
	}
	switch desired.Origin {
	case RuleOriginAdopted:
		return InventoryStateAdopted
	default:
		return InventoryStateManaged
	}
}

func RuntimeUsageKey(rule FirewallRule) string {
	protocol := strings.ToLower(strings.TrimSpace(rule.Protocol))
	port := strings.TrimSpace(rule.DestinationPort)
	if port == "" || (protocol != "tcp" && protocol != "udp") {
		return ""
	}
	return protocol + "\x00" + port
}

func AttachRuntimeUsage(items []InventoryItem, usage map[string]RuntimeUsage) []InventoryItem {
	result := make([]InventoryItem, len(items))
	copy(result, items)
	for index := range result {
		value, exists := runtimeUsageForRule(result[index].Rule, usage)
		if !exists {
			continue
		}
		value.UsedBy = normalizedUsageOwners(value.UsedBy)
		value.Used = value.Used || len(value.UsedBy) > 0
		result[index].Usage = &value
	}
	return result
}

func runtimeUsageForRule(rule FirewallRule, usage map[string]RuntimeUsage) (RuntimeUsage, bool) {
	key := RuntimeUsageKey(rule)
	if key == "" {
		return RuntimeUsage{}, false
	}
	if value, exists := usage[key]; exists {
		return value, true
	}
	intervals, err := portIntervals(rule.DestinationPort)
	if err != nil {
		return RuntimeUsage{}, false
	}
	protocolPrefix := strings.ToLower(strings.TrimSpace(rule.Protocol)) + "\x00"
	combined := RuntimeUsage{}
	found := false
	for usageKey, value := range usage {
		if !strings.HasPrefix(usageKey, protocolPrefix) {
			continue
		}
		port, err := strconv.Atoi(strings.TrimPrefix(usageKey, protocolPrefix))
		if err != nil || !portInIntervals(port, intervals) {
			continue
		}
		found = true
		combined.Used = combined.Used || value.Used
		combined.UsedBy = append(combined.UsedBy, value.UsedBy...)
		if combined.Reason == "" {
			combined.Reason = value.Reason
		} else if value.Reason != "" && combined.Reason != value.Reason {
			combined.Reason = "multiple"
		}
	}
	return combined, found
}

func portInIntervals(port int, intervals [][2]int) bool {
	for _, interval := range intervals {
		if port >= interval[0] && port <= interval[1] {
			return true
		}
	}
	return false
}

func normalizedUsageOwners(owners []string) []string {
	seen := make(map[string]struct{}, len(owners))
	result := make([]string, 0, len(owners))
	for _, owner := range owners {
		owner = strings.TrimSpace(owner)
		if owner == "" {
			continue
		}
		if _, exists := seen[owner]; exists {
			continue
		}
		seen[owner] = struct{}{}
		result = append(result, owner)
	}
	sort.Strings(result)
	return result
}
