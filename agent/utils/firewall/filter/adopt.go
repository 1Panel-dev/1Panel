package filter

import "fmt"

var ErrDuplicateAdoption = fmt.Errorf("%w: duplicate firewall rules prevent adoption; manually delete duplicate rules and retry", ErrRuleOperation)

func CheckAdoptDuplicates(snapshot Snapshot, requested FirewallRule) error {
	count := 0
	for _, observed := range snapshot.Rules {
		if observed.ParseStatus != ParseStatusSupported {
			continue
		}
		same, err := SameRuleContent(observed.Rule, requested)
		if err != nil {
			return err
		}
		if same {
			count++
			if count > 1 {
				return ErrDuplicateAdoption
			}
		}
	}
	return nil
}

func CheckAdopt(snapshot Snapshot, requested FirewallRule, desired []DesiredRule, locator Locator) (RuleCheckResult, error) {
	normalized, err := NormalizeRule(requested)
	if err != nil {
		return RuleCheckResult{}, err
	}
	if normalized.Scope.Key() != snapshot.Scope.Key() {
		return RuleCheckResult{}, ErrInvalidScope
	}
	key, err := RuleKey(normalized)
	if err != nil {
		return RuleCheckResult{}, err
	}
	result := RuleCheckResult{RequestedRule: normalized, RequestedRuleKey: key}
	var selected *ObservedRule
	for index := range snapshot.Rules {
		candidate := &snapshot.Rules[index]
		if SameLocator(candidate.Locator, locator) {
			if selected != nil {
				return RuleCheckResult{}, ErrRuleStale
			}
			selected = candidate
		}
	}
	if selected == nil {
		return RuleCheckResult{}, ErrRuleStale
	}
	selectedKey, err := RuleKey(selected.Rule)
	if err != nil || selected.ParseStatus != ParseStatusSupported || selectedKey != key {
		return RuleCheckResult{}, fmt.Errorf("%w: selected rule changed or cannot be adopted", ErrRuleStale)
	}
	if err := CheckAdoptDuplicates(snapshot, normalized); err != nil {
		if err != ErrDuplicateAdoption {
			return RuleCheckResult{}, err
		}
		result.Decision, result.Classification, result.Reason = CheckDecisionBlocked, CheckClassificationExactExternal, "duplicate_rules"
		return finishCheck(result)
	}
	items, err := MergeInventory(InventoryMergeInput{Observed: snapshot.Rules, Desired: desired})
	if err != nil {
		return RuleCheckResult{}, err
	}
	for _, item := range items {
		if item.Desired != nil && item.Observed != nil && SameLocator(item.Observed.Locator, locator) {
			result.Decision, result.Classification, result.Reason = CheckDecisionNoChange, CheckClassificationExactManaged, "equivalent_managed_rule"
			result.ExistingRuleUUID = item.Desired.UUID
			return result, nil
		}
	}
	for _, owned := range desired {
		same, err := SameRuleContent(owned.Rule, normalized)
		if err != nil {
			return RuleCheckResult{}, err
		}
		if same {
			result.Decision, result.Classification, result.Reason = CheckDecisionBlocked, CheckClassificationExactManaged, "duplicate_rules"
			result.ExistingRuleUUID = owned.UUID
			return result, nil
		}
	}
	result.Candidates = []ObservedRule{*selected}
	switch {
	case selected.Protected:
		result.Decision, result.Classification, result.Reason = CheckDecisionBlocked, CheckClassificationProtected, "protected_rule"
	case containsPersistenceDrift(result.Candidates):
		result.Decision, result.Classification, result.Reason = CheckDecisionBlocked, CheckClassificationConflict, "runtime_permanent_mismatch"
	default:
		result.Decision, result.Classification, result.Reason = CheckDecisionConfirmationRequired, CheckClassificationExactExternal, "equivalent_external_rule"
		result.AllowedActions = []CheckAction{CheckActionAdopt, CheckActionCancel}
	}
	return finishCheck(result)
}
