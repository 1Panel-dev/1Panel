package filter

import (
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
)

var (
	ErrRuleStale         = errors.New("firewall rule state is stale")
	ErrRuleOperation     = errors.New("firewall rule operation is not allowed")
	ErrRuleCheckRequired = errors.New("firewall rule must be checked again")
)

type CheckDecision string

const (
	CheckDecisionReady                CheckDecision = "ready"
	CheckDecisionConfirmationRequired CheckDecision = "confirmation_required"
	CheckDecisionBlocked              CheckDecision = "blocked"
	CheckDecisionNoChange             CheckDecision = "no_change"
)

type CheckClassification string

const (
	CheckClassificationNone          CheckClassification = "none"
	CheckClassificationExactManaged  CheckClassification = "exact_managed"
	CheckClassificationExactExternal CheckClassification = "exact_external"
	CheckClassificationConflict      CheckClassification = "conflict"
	CheckClassificationUnsupported   CheckClassification = "unsupported"
	CheckClassificationProtected     CheckClassification = "protected"
)

type CheckAction string

const (
	CheckActionCreate      CheckAction = "create"
	CheckActionAdopt       CheckAction = "adopt"
	CheckActionSelectAdopt CheckAction = "select_adopt"
	CheckActionCancel      CheckAction = "cancel"
)

type RuleCheckResult struct {
	Decision         CheckDecision       `json:"decision"`
	Classification   CheckClassification `json:"classification"`
	Reason           string              `json:"reason"`
	RequestedRule    FirewallRule        `json:"requestedRule"`
	RequestedRuleKey string              `json:"requestedRuleKey"`
	ExistingRuleUUID string              `json:"existingRuleUUID,omitempty"`
	Candidates       []ObservedRule      `json:"candidates,omitempty"`
	AllowedActions   []CheckAction       `json:"allowedActions,omitempty"`
}

func CheckCreate(
	snapshot Snapshot,
	requested FirewallRule,
	desired []DesiredRule,
	clientIP string,
	protectedPorts ...PortWhitelist,
) (RuleCheckResult, error) {
	normalized, err := NormalizeRule(requested)
	if err != nil {
		return RuleCheckResult{}, err
	}
	if normalized.Scope.Key() != snapshot.Scope.Normalize().Key() {
		return RuleCheckResult{}, fmt.Errorf("%w: requested scope %q does not match snapshot %q", ErrInvalidRule, normalized.Scope.Key(), snapshot.Scope.Key())
	}
	ruleKey, err := RuleKey(normalized)
	if err != nil {
		return RuleCheckResult{}, err
	}
	result := RuleCheckResult{
		Decision:         CheckDecisionReady,
		Classification:   CheckClassificationNone,
		Reason:           "new_rule",
		RequestedRule:    normalized,
		RequestedRuleKey: ruleKey,
	}
	if RuleBlocksManagementConnection(normalized, clientIP, protectedPorts...) {
		result.Decision = CheckDecisionBlocked
		result.Classification = CheckClassificationProtected
		result.Reason = "current_management_connection"
		return finishCheck(result)
	}

	for _, owned := range desired {
		ownedKey, err := desiredRuleKey(owned)
		if err != nil {
			return RuleCheckResult{}, err
		}
		if ownedKey != ruleKey || (owned.Origin != RuleOriginCreated && owned.Origin != RuleOriginAdopted) {
			continue
		}
		owned.RuleKey = ownedKey
		result.ExistingRuleUUID = owned.UUID
		if snapshotHasOwnedRule(snapshot, owned) {
			result.Decision = CheckDecisionNoChange
			result.Classification = CheckClassificationExactManaged
			result.Reason = "equivalent_managed_rule"
		} else {
			result.Decision = CheckDecisionBlocked
			result.Classification = CheckClassificationConflict
			result.Reason = "managed_rule_drifted"
		}
		return finishCheck(result)
	}

	exact := make([]ObservedRule, 0)
	for _, observed := range snapshot.Rules {
		if observed.ParseStatus != ParseStatusSupported {
			if hydrated, ok := hydrateOwnedPartialRule(observed, desired); ok {
				observed = hydrated
			} else {
				if observed.Rule.Scope.Provider == ProviderFirewalld && observed.Rule.NativeKind == NativeKindZoneService {
					continue
				}
				result.Decision = CheckDecisionBlocked
				result.Classification = CheckClassificationUnsupported
				result.Reason = "opaque_rule_in_target_scope"
				result.Candidates = append(result.Candidates, observed)
				continue
			}
		}
		observedRule, err := NormalizeRule(observed.Rule)
		if err != nil {
			return RuleCheckResult{}, err
		}
		observed.Rule = observedRule
		observedKey, err := RuleKey(observedRule)
		if err != nil {
			return RuleCheckResult{}, err
		}
		if observedKey == ruleKey {
			exact = append(exact, observed)
		}
	}

	switch {
	case containsPersistenceDrift(exact):
		result.Decision = CheckDecisionBlocked
		result.Classification = CheckClassificationConflict
		result.Reason = "runtime_permanent_mismatch"
		result.Candidates = exact
	case containsProtectedRule(exact):
		result.Decision = CheckDecisionBlocked
		result.Classification = CheckClassificationProtected
		result.Reason = "protected_rule"
		result.Candidates = exact
	case len(exact) == 1:
		result.Decision = CheckDecisionConfirmationRequired
		result.Classification = CheckClassificationExactExternal
		result.Reason = "equivalent_external_rule"
		result.Candidates = exact
		result.AllowedActions = []CheckAction{CheckActionAdopt, CheckActionCancel}
	case len(exact) > 1:
		result.Decision = CheckDecisionConfirmationRequired
		result.Classification = CheckClassificationExactExternal
		result.Reason = "multiple_equivalent_external_rules"
		result.Candidates = exact
		result.AllowedActions = []CheckAction{CheckActionSelectAdopt, CheckActionCancel}
	case result.Classification == CheckClassificationUnsupported:
	default:
		result.AllowedActions = []CheckAction{CheckActionCreate}
	}
	return finishCheck(result)
}

func hydrateOwnedPartialRule(observed ObservedRule, desired []DesiredRule) (ObservedRule, bool) {
	if observed.ParseStatus != ParseStatusPartial || strings.TrimSpace(observed.Marker) == "" ||
		len(observed.UncertainFields) != 1 || observed.UncertainFields[0] != ObservedFieldProtocol {
		return observed, false
	}
	var matched *DesiredRule
	for index := range desired {
		candidate := &desired[index]
		if candidate.Origin != RuleOriginCreated && candidate.Origin != RuleOriginAdopted {
			continue
		}
		if candidate.Rule.Scope.Normalize().Key() != observed.Rule.Scope.Normalize().Key() ||
			!desiredMarkerMatchesObserved(*candidate, observed.Marker) ||
			!ObservedRuleMatchesExpected(observed, candidate.Rule) {
			continue
		}
		if matched != nil {
			return observed, false
		}
		matched = candidate
	}
	if matched == nil {
		return observed, false
	}
	orderIndex := observed.Rule.OrderIndex
	observed.Rule = matched.Rule
	observed.Rule.OrderIndex = orderIndex
	observed.ParseStatus = ParseStatusSupported
	observed.UncertainFields = nil
	return observed, true
}

func desiredMarkerMatchesObserved(desired DesiredRule, observedMarker string) bool {
	observedMarker = strings.TrimSpace(observedMarker)
	if observedMarker == strings.TrimSpace(desired.Marker) {
		return observedMarker != ""
	}
	return desired.Origin == RuleOriginAdopted && strings.TrimSpace(desired.UUID) != "" &&
		observedMarker == "1panel-rule:"+strings.TrimSpace(desired.UUID)
}

func finishCheck(result RuleCheckResult) (RuleCheckResult, error) {
	for index, candidate := range result.Candidates {
		identity, err := InstanceKey(candidate)
		if err != nil {
			identity, err = RuleKey(candidate.Rule)
			if err != nil {
				return RuleCheckResult{}, err
			}
		}
		result.Candidates[index].InstanceKey = identity
	}
	return result, nil
}

func containsPersistenceDrift(rules []ObservedRule) bool {
	for _, rule := range rules {
		if rule.Persistence != "" && rule.Persistence != PersistenceStatusConverged {
			return true
		}
	}
	return false
}

func containsProtectedRule(rules []ObservedRule) bool {
	for _, rule := range rules {
		if rule.Protected {
			return true
		}
	}
	return false
}

func RuleBlocksManagementConnection(
	rule FirewallRule,
	clientIP string,
	protectedPorts ...PortWhitelist,
) bool {
	if rule.Action == ActionAccept {
		return false
	}
	targetAddress := rule.SourceAddress
	if targetAddress != "" {
		client, err := netip.ParseAddr(strings.TrimSpace(clientIP))
		if err != nil {
			return false
		}
		if rule.Scope.Family == FamilyIPv4 {
			client = client.Unmap()
			if !client.Is4() {
				return false
			}
		} else if rule.Scope.Family == FamilyIPv6 && (!client.Is6() || client.Is4In6()) {
			return false
		}
		prefix, err := netip.ParsePrefix(targetAddress)
		if err != nil || !prefix.Contains(client) {
			return false
		}
	}
	if rule.DestinationPort == "" {
		return true
	}
	for _, protected := range protectedPorts {
		family := strings.ToLower(strings.TrimSpace(protected.Family))
		if family != "" && rule.Scope.Family != FamilyInet && string(rule.Scope.Family) != family {
			continue
		}
		protocol := strings.ToLower(strings.TrimSpace(protected.Protocol))
		if rule.Protocol != "all" && rule.Protocol != protocol {
			continue
		}
		if portsOverlap(rule.DestinationPort, protected.Port) {
			return true
		}
	}
	return false
}

func FindCandidate(candidates []ObservedRule, selected string) (ObservedRule, error) {
	matched := make([]ObservedRule, 0, 1)
	for _, candidate := range candidates {
		identity, err := InstanceKey(candidate)
		if err != nil {
			identity, err = RuleKey(candidate.Rule)
			if err != nil {
				continue
			}
		}
		if selected != "" && identity == selected {
			matched = append(matched, candidate)
		}
	}
	if len(matched) != 1 {
		return ObservedRule{}, ErrRuleOperation
	}
	return matched[0], nil
}

func desiredRuleKey(desired DesiredRule) (string, error) {
	key, err := RuleKey(desired.Rule)
	if err != nil {
		return "", err
	}
	if desired.RuleKey != "" && desired.RuleKey != key {
		return "", fmt.Errorf("%w: desired rule %q key mismatch", ErrInvalidRule, desired.UUID)
	}
	return key, nil
}

func snapshotHasOwnedRule(snapshot Snapshot, desired DesiredRule) bool {
	items, err := MergeInventory(InventoryMergeInput{Observed: snapshot.Rules, Desired: []DesiredRule{desired}})
	if err != nil {
		return false
	}
	for _, item := range items {
		if item.Desired != nil {
			return item.Match == InventoryMatchExact && item.State != InventoryStateDrifted
		}
	}
	return false
}

func RuleCovers(existing, requested FirewallRule) bool {
	existing, existingErr := NormalizeRule(existing)
	requested, requestedErr := NormalizeRule(requested)
	if existingErr != nil || requestedErr != nil || existing.Scope.Key() != requested.Scope.Key() {
		return false
	}
	return familyCovers(existing.Scope.Family, requested.Scope.Family) &&
		protocolCovers(existing.Protocol, requested.Protocol) &&
		addressCovers(existing.SourceAddress, requested.SourceAddress) &&
		addressCovers(existing.DestinationAddress, requested.DestinationAddress) &&
		portCovers(existing.SourcePort, requested.SourcePort) &&
		portCovers(existing.DestinationPort, requested.DestinationPort) &&
		(existing.Interface == "" || existing.Interface == requested.Interface)
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

func familyCovers(existing, requested Family) bool {
	return existing == FamilyInet || existing == requested
}

func familiesOverlap(left, right Family) bool {
	return left == FamilyInet || right == FamilyInet || left == right
}

func protocolCovers(existing, requested string) bool {
	return existing == "all" || existing == requested
}

func protocolsOverlap(left, right string) bool {
	return left == "all" || right == "all" || left == right
}

func addressCovers(existing, requested string) bool {
	if existing == "" {
		return true
	}
	if requested == "" {
		return false
	}
	existingPrefix, err := netip.ParsePrefix(existing)
	if err != nil {
		return false
	}
	requestedPrefix, err := netip.ParsePrefix(requested)
	if err != nil {
		return false
	}
	return existingPrefix.Bits() <= requestedPrefix.Bits() && existingPrefix.Contains(requestedPrefix.Addr())
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
