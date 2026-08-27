package filter

import "testing"

func TestCheckCreateRequestsAdoptionForEquivalentExternalRule(t *testing.T) {
	rule := checkAddressRule("172.16.10.111", ActionDrop)
	snapshot := checkSnapshot(t, rule)
	plan, err := CheckCreate(snapshot, rule, nil, "")
	if err != nil {
		t.Fatalf("plan create: %v", err)
	}
	if plan.Decision != CheckDecisionConfirmationRequired || plan.Classification != CheckClassificationExactExternal || plan.Reason != "equivalent_external_rule" {
		t.Fatalf("unexpected adoption check: %#v", plan)
	}
	if len(plan.Candidates) != 1 || len(plan.AllowedActions) != 2 || plan.AllowedActions[0] != CheckActionAdopt || plan.Candidates[0].InstanceKey == "" {
		t.Fatalf("adoption details missing: %#v", plan)
	}
}

func TestCheckCreateTreatsManagedDuplicateAsIdempotent(t *testing.T) {
	rule := checkPortRule("22", ActionAccept)
	marker := "onepanel:created:ssh"
	observed := checkObservedRule(rule, marker, 1)
	snapshot, err := NewSnapshot(rule.Scope, []ObservedRule{observed})
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	ruleKey, _ := RuleKey(rule)
	desired := DesiredRule{
		UUID: "managed", Rule: rule, RuleKey: ruleKey, Origin: RuleOriginCreated, Marker: marker,
	}

	plan, err := CheckCreate(snapshot, rule, []DesiredRule{desired}, "")
	if err != nil {
		t.Fatalf("plan managed duplicate: %v", err)
	}
	if plan.Decision != CheckDecisionNoChange || plan.Classification != CheckClassificationExactManaged || plan.ExistingRuleUUID != desired.UUID {
		t.Fatalf("managed duplicate was not idempotent: %#v", plan)
	}
}

func TestCheckCreateBlocksMissingManagedRuleInsteadOfCreatingDuplicate(t *testing.T) {
	rule := checkPortRule("22", ActionAccept)
	snapshot, err := NewSnapshot(rule.Scope, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	ruleKey, _ := RuleKey(rule)
	desired := DesiredRule{UUID: "managed", Rule: rule, RuleKey: ruleKey, Origin: RuleOriginCreated}
	plan, err := CheckCreate(snapshot, rule, []DesiredRule{desired}, "")
	if err != nil {
		t.Fatalf("plan missing managed rule: %v", err)
	}
	if plan.Decision != CheckDecisionBlocked || plan.Reason != "managed_rule_drifted" {
		t.Fatalf("missing managed rule was recreated: %#v", plan)
	}
}

func TestCheckCreateRequiresCandidateSelectionForDuplicates(t *testing.T) {
	rule := checkPortRule("3306", ActionAccept)
	first := checkObservedRule(rule, "", 1)
	second := checkObservedRule(rule, "", 2)
	snapshot, err := NewSnapshot(rule.Scope, []ObservedRule{first, second})
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	plan, err := CheckCreate(snapshot, rule, nil, "")
	if err != nil {
		t.Fatalf("plan duplicate candidates: %v", err)
	}
	if plan.Decision != CheckDecisionConfirmationRequired || plan.Reason != "multiple_equivalent_external_rules" || len(plan.Candidates) != 2 || plan.AllowedActions[0] != CheckActionSelectAdopt {
		t.Fatalf("check guessed an equivalent candidate: %#v", plan)
	}
	if plan.Candidates[0].InstanceKey == "" || plan.Candidates[1].InstanceKey == "" || plan.Candidates[0].InstanceKey == plan.Candidates[1].InstanceKey {
		t.Fatalf("candidate instance keys must be present and unique: %+v", plan.Candidates)
	}
	if _, err := FindCandidate(plan.Candidates, ""); err == nil {
		t.Fatalf("expected candidate selection to be required, got %v", err)
	}
	secondKey, _ := InstanceKey(second)
	candidate, err := FindCandidate(plan.Candidates, secondKey)
	if err != nil || candidate.Locator.Position == nil || *candidate.Locator.Position != 2 {
		t.Fatalf("selected candidate was not found: candidate=%#v err=%v", candidate, err)
	}
}

func TestCheckCreateBlocksAdoptionOfProtectedRule(t *testing.T) {
	rule := checkPortRule("22", ActionAccept)
	observed := checkObservedRule(rule, "", 1)
	observed.Protected = true
	snapshot, err := NewSnapshot(rule.Scope, []ObservedRule{observed})
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	plan, err := CheckCreate(snapshot, rule, nil, "")
	if err != nil {
		t.Fatalf("plan protected rule: %v", err)
	}
	if plan.Decision != CheckDecisionBlocked || plan.Classification != CheckClassificationProtected || plan.Reason != "protected_rule" || len(plan.AllowedActions) != 0 {
		t.Fatalf("protected rule could be adopted: %#v", plan)
	}
}

func TestCheckCreateBlocksRuntimePermanentMismatch(t *testing.T) {
	rule := checkPortRule("443", ActionAccept)
	observed := checkObservedRule(rule, "", 1)
	observed.Persistence = PersistenceStatusRuntimeOnly
	snapshot, err := NewSnapshot(rule.Scope, []ObservedRule{observed})
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	plan, err := CheckCreate(snapshot, rule, nil, "")
	if err != nil {
		t.Fatalf("plan persistence drift: %v", err)
	}
	if plan.Decision != CheckDecisionBlocked || plan.Reason != "runtime_permanent_mismatch" {
		t.Fatalf("runtime-only rule could be adopted: %#v", plan)
	}
}

func TestCheckCreateDoesNotLetOpaqueFirewalldServiceBlockUnrelatedRule(t *testing.T) {
	rule := FirewallRule{
		Scope:      Scope{Provider: ProviderFirewalld, Family: FamilyInet, Zone: "public", Direction: DirectionInput},
		NativeKind: NativeKindZonePort, Protocol: "tcp", DestinationPort: "8080", Action: ActionAccept,
	}
	service := ObservedRule{
		Rule:        FirewallRule{Scope: rule.Scope, NativeKind: NativeKindZoneService},
		Locator:     Locator{Provider: ProviderFirewalld, ScopeKey: rule.Scope.Key(), Canonical: "service:ssh"},
		ParseStatus: ParseStatusOpaque, Raw: "ssh", Persistence: PersistenceStatusConverged,
	}
	snapshot, err := NewSnapshot(rule.Scope, []ObservedRule{service})
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	plan, err := CheckCreate(snapshot, rule, nil, "")
	if err != nil {
		t.Fatalf("plan with opaque service: %v", err)
	}
	if plan.Decision != CheckDecisionReady || plan.Classification != CheckClassificationNone {
		t.Fatalf("opaque service blocked unrelated native port: %#v", plan)
	}
	opaqueRich := service
	opaqueRich.Rule.NativeKind = NativeKindRichRule
	opaqueRich.Locator.Canonical = `rich:rule log prefix="audit" accept`
	opaqueRich.Raw = `rule log prefix="audit" accept`
	snapshot, err = NewSnapshot(rule.Scope, []ObservedRule{opaqueRich})
	if err != nil {
		t.Fatalf("opaque rich snapshot: %v", err)
	}
	plan, err = CheckCreate(snapshot, rule, nil, "")
	if err != nil {
		t.Fatalf("plan with opaque rich rule: %v", err)
	}
	if plan.Decision != CheckDecisionBlocked || plan.Classification != CheckClassificationUnsupported {
		t.Fatalf("opaque rich rule did not block an unsafe plan: %#v", plan)
	}
}

func TestCheckCreateClassifiesCoverageAndConflict(t *testing.T) {
	requested := checkAddressRule("172.16.10.111", ActionDrop)
	covering := checkAddressRule("172.16.10.0/24", ActionDrop)
	coveredSnapshot := checkSnapshot(t, covering)
	plan, err := CheckCreate(coveredSnapshot, requested, nil, "")
	if err != nil {
		t.Fatalf("plan covered rule: %v", err)
	}
	if plan.Classification != CheckClassificationCovered || plan.Decision != CheckDecisionConfirmationRequired || plan.AllowedActions[0] != CheckActionCreateAnyway {
		t.Fatalf("unexpected covered plan: %#v", plan)
	}

	conflicting := checkAddressRule("172.16.10.0/24", ActionAccept)
	conflictSnapshot := checkSnapshot(t, conflicting)
	plan, err = CheckCreate(conflictSnapshot, requested, nil, "")
	if err != nil {
		t.Fatalf("plan conflicting rule: %v", err)
	}
	if plan.Classification != CheckClassificationConflict || plan.Decision != CheckDecisionBlocked {
		t.Fatalf("unexpected conflict plan: %#v", plan)
	}
}

func TestCheckCreateAllowsPartialOverlapWithOppositeActionAfterConfirmation(t *testing.T) {
	scope := Scope{Provider: ProviderFirewalld, Family: FamilyInet, Zone: FirewalldInputZone, Direction: DirectionInput}
	existingRule := FirewallRule{
		Scope:      Scope{Provider: ProviderFirewalld, Family: FamilyIPv4, Zone: FirewalldInputZone, Direction: DirectionInput},
		NativeKind: NativeKindRichRule, Protocol: "all", SourceAddress: "1.1.1.1", Action: ActionDrop,
	}
	requested := FirewallRule{
		Scope: scope, NativeKind: NativeKindZonePort, Protocol: "tcp", DestinationPort: "8080", Action: ActionAccept,
	}
	snapshot := checkSnapshot(t, existingRule)
	plan, err := CheckCreate(snapshot, requested, nil, "")
	if err != nil {
		t.Fatalf("plan partially overlapping rule: %v", err)
	}
	if plan.Decision != CheckDecisionConfirmationRequired || plan.Classification != CheckClassificationConflict ||
		plan.Reason != "partially_overlapping_rule_with_different_action" ||
		len(plan.AllowedActions) == 0 || plan.AllowedActions[0] != CheckActionCreateAnyway {
		t.Fatalf("partial overlap was not confirmable: %#v", plan)
	}
}

func TestRuleCoverageAndOverlapRespectAddressFamilies(t *testing.T) {
	base := Scope{Provider: ProviderFirewalld, Zone: FirewalldInputZone, Direction: DirectionInput}
	ipv4 := FirewallRule{
		Scope:      Scope{Provider: ProviderFirewalld, Family: FamilyIPv4, Zone: base.Zone, Direction: base.Direction},
		NativeKind: NativeKindRichRule, Protocol: "all", SourceAddress: "1.1.1.1", Action: ActionDrop,
	}
	ipv6 := FirewallRule{
		Scope:      Scope{Provider: ProviderFirewalld, Family: FamilyIPv6, Zone: base.Zone, Direction: base.Direction},
		NativeKind: NativeKindRichRule, Protocol: "tcp", DestinationPort: "8080", Action: ActionAccept,
	}
	if RulesOverlap(ipv4, ipv6) || RuleCovers(ipv4, ipv6) {
		t.Fatal("disjoint IPv4 and IPv6 rules must not overlap")
	}
	inet := ipv6
	inet.Scope.Family = FamilyInet
	inet.SourceAddress = ""
	if !RulesOverlap(ipv4, inet) {
		t.Fatal("inet rule should overlap an IPv4 rule")
	}
	if RuleCovers(ipv4, inet) {
		t.Fatal("IPv4 rule must not cover a dual-stack inet rule")
	}
}

func TestCheckCreateAllowsOrderedFirewalldDenyBeforeNativePort(t *testing.T) {
	scope := Scope{Provider: ProviderFirewalld, Family: FamilyIPv4, Zone: "public", Direction: DirectionInput}
	existingRule := FirewallRule{
		Scope:      Scope{Provider: ProviderFirewalld, Family: FamilyInet, Zone: "public", Direction: DirectionInput},
		NativeKind: NativeKindZonePort, Protocol: "tcp", DestinationPort: "3306", Action: ActionAccept,
	}
	priority := -100
	requested := FirewallRule{
		Scope: scope, NativeKind: NativeKindRichRule, Protocol: "tcp", SourceAddress: "172.16.10.111",
		DestinationPort: "3306", Action: ActionDrop, Priority: &priority,
	}
	existing := checkObservedRule(existingRule, "", 1)
	existing.Persistence = PersistenceStatusConverged
	snapshot, err := NewSnapshot(scope, []ObservedRule{existing})
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	plan, err := CheckCreate(snapshot, requested, nil, "")
	if err != nil {
		t.Fatalf("plan ordered deny: %v", err)
	}
	if plan.Decision != CheckDecisionReady || plan.Classification != CheckClassificationNone {
		t.Fatalf("negative-priority deny was blocked: %#v", plan)
	}

	existing.Protected = true
	protectedSnapshot, _ := NewSnapshot(scope, []ObservedRule{existing})
	plan, err = CheckCreate(protectedSnapshot, requested, nil, "")
	if err != nil {
		t.Fatalf("plan protected overlap: %v", err)
	}
	if plan.Decision != CheckDecisionBlocked || plan.Classification != CheckClassificationConflict {
		t.Fatalf("protected port overlap was allowed: %#v", plan)
	}
}

func TestRuleCoverageAndOverlapUseNormalizedRanges(t *testing.T) {
	existing := checkPortRule("8000-9000", ActionAccept)
	requested := checkPortRule("8080", ActionAccept)
	if !RuleCovers(existing, requested) || !RulesOverlap(existing, requested) {
		t.Fatal("expected port range to cover and overlap contained port")
	}
	differentProtocol := requested
	differentProtocol.Protocol = "udp"
	if RulesOverlap(existing, differentProtocol) {
		t.Fatal("different transport protocols should not overlap")
	}
}

func TestCheckCreateBlocksRuleTargetingCurrentManagementClient(t *testing.T) {
	rule := checkAddressRule("203.0.113.9", ActionDrop)
	snapshot, err := NewSnapshot(rule.Scope, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	plan, err := CheckCreate(snapshot, rule, nil, "203.0.113.9")
	if err != nil {
		t.Fatalf("plan create: %v", err)
	}
	if plan.Decision != CheckDecisionBlocked || plan.Classification != CheckClassificationProtected || plan.Reason != "current_management_connection" {
		t.Fatalf("unexpected current connection decision: %#v", plan)
	}
}

func TestCheckCreateBlocksProtectedManagementPortWithoutObservedAllowRule(t *testing.T) {
	rule := checkPortRule("22", ActionDrop)
	snapshot, err := NewSnapshot(rule.Scope, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	plan, err := CheckCreate(snapshot, rule, nil, "203.0.113.9", PortWhitelist{
		Family: "ipv4", Port: "22", Protocol: "tcp",
	})
	if err != nil {
		t.Fatalf("plan create: %v", err)
	}
	if plan.Decision != CheckDecisionBlocked || plan.Classification != CheckClassificationProtected || plan.Reason != "current_management_connection" {
		t.Fatalf("protected management port was not blocked: %#v", plan)
	}
}

func TestCheckCreateAllowsProtectedPortDenyForUnrelatedSource(t *testing.T) {
	rule := checkPortRule("22", ActionDrop)
	rule.SourceAddress = "198.51.100.8/32"
	snapshot, err := NewSnapshot(rule.Scope, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	plan, err := CheckCreate(snapshot, rule, nil, "203.0.113.9", PortWhitelist{
		Family: "ipv4", Port: "22", Protocol: "tcp",
	})
	if err != nil {
		t.Fatalf("plan create: %v", err)
	}
	if plan.Decision != CheckDecisionReady {
		t.Fatalf("unrelated source was incorrectly blocked: %#v", plan)
	}
}

func checkSnapshot(t *testing.T, rules ...FirewallRule) Snapshot {
	t.Helper()
	observed := make([]ObservedRule, 0, len(rules))
	for index, rule := range rules {
		observed = append(observed, checkObservedRule(rule, "", index+1))
	}
	snapshot, err := NewSnapshot(rules[0].Scope, observed)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	return snapshot
}

func checkObservedRule(rule FirewallRule, marker string, position int) ObservedRule {
	return ObservedRule{
		Rule:    rule,
		Locator: Locator{Provider: rule.Scope.Provider, ScopeKey: rule.Scope.Key(), Position: &position},
		Marker:  marker, ParseStatus: ParseStatusSupported,
	}
}

func checkAddressRule(address string, action Action) FirewallRule {
	return FirewallRule{
		Scope:         Scope{Provider: ProviderIptables, Family: FamilyIPv4, Table: "filter", Chain: "1PANEL_BASIC", Direction: DirectionInput},
		NativeKind:    NativeKindRule,
		Protocol:      "all",
		SourceAddress: address,
		Action:        action,
	}
}

func checkPortRule(port string, action Action) FirewallRule {
	return FirewallRule{
		Scope:           Scope{Provider: ProviderIptables, Family: FamilyIPv4, Table: "filter", Chain: "1PANEL_BASIC", Direction: DirectionInput},
		NativeKind:      NativeKindRule,
		Protocol:        "tcp",
		DestinationPort: port,
		Action:          action,
	}
}
