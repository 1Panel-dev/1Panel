package filter

import (
	"errors"
	"testing"
)

func TestRuleKeyUsesSemanticFields(t *testing.T) {
	priority := 10
	base := FirewallRule{
		UUID: "first",
		Scope: Scope{
			Provider:  ProviderFirewalld,
			Family:    FamilyIPv4,
			Zone:      "public",
			Direction: DirectionInput,
		},
		NativeKind:      NativeKindRichRule,
		Protocol:        "TCP",
		SourceAddress:   "172.16.10.111",
		DestinationPort: "3306:3306",
		Action:          "ALLOW",
		Priority:        &priority,
		Description:     "first description",
	}
	other := base
	other.UUID = "second"
	other.Description = "description is not identity"
	other.SourceAddress = "172.16.10.111/32"
	other.DestinationPort = "3306"

	firstKey, err := RuleKey(base)
	if err != nil {
		t.Fatalf("first rule key: %v", err)
	}
	secondKey, err := RuleKey(other)
	if err != nil {
		t.Fatalf("second rule key: %v", err)
	}
	if firstKey != secondKey {
		t.Fatalf("equivalent rules produced different keys:\n%s\n%s", firstKey, secondKey)
	}

	changedPriority := 11
	other.Priority = &changedPriority
	changedKey, err := RuleKey(other)
	if err != nil {
		t.Fatalf("changed rule key: %v", err)
	}
	if changedKey == firstKey {
		t.Fatal("priority change did not change rule key")
	}
}

func TestRuleKeyKeepsFamilyInsideSharedFirewalldScope(t *testing.T) {
	ipv4 := FirewallRule{
		Scope:      Scope{Provider: ProviderFirewalld, Family: FamilyIPv4, Zone: "public", Direction: DirectionInput},
		NativeKind: NativeKindRichRule, Protocol: "tcp", Action: ActionAccept,
	}
	ipv6 := ipv4
	ipv6.Scope.Family = FamilyIPv6
	first, err := RuleKey(ipv4)
	if err != nil {
		t.Fatalf("IPv4 key: %v", err)
	}
	second, err := RuleKey(ipv6)
	if err != nil {
		t.Fatalf("IPv6 key: %v", err)
	}
	if first == second {
		t.Fatal("firewalld family-specific rich rules collapsed to one identity")
	}
}

func TestInstanceKeyIncludesLocator(t *testing.T) {
	positionOne := 1
	positionTwo := 2
	rule := ObservedRule{
		Rule: FirewallRule{
			Scope:           Scope{Provider: ProviderUFW, Family: FamilyIPv4, Direction: DirectionInput},
			Protocol:        "tcp",
			DestinationPort: "22",
			Action:          ActionAccept,
		},
		Marker:      "1panel-rule:test",
		ParseStatus: ParseStatusSupported,
		Locator:     Locator{Position: &positionOne},
	}
	first, err := InstanceKey(rule)
	if err != nil {
		t.Fatalf("first instance key: %v", err)
	}
	rule.Locator.Position = &positionTwo
	second, err := InstanceKey(rule)
	if err != nil {
		t.Fatalf("second instance key: %v", err)
	}
	if first == second {
		t.Fatal("position change did not change instance key")
	}
	rule.Locator.Position = &positionOne
	rule.Persistence = PersistenceStatusRuntimeOnly
	runtimeOnly, err := InstanceKey(rule)
	if err != nil {
		t.Fatalf("runtime-only instance key: %v", err)
	}
	if runtimeOnly == first {
		t.Fatal("runtime/permanent presence did not change instance key")
	}
}

func TestSnapshotRevisionIsInputOrderIndependent(t *testing.T) {
	scope := Scope{Provider: ProviderIptables, Family: FamilyIPv4, Table: "filter", Chain: "1PANEL_BASIC", Direction: DirectionInput}
	positionOne := 1
	positionTwo := 2
	rules := []ObservedRule{
		{
			Rule:        FirewallRule{Scope: scope, Protocol: "tcp", DestinationPort: "22", Action: ActionAccept},
			Locator:     Locator{Position: &positionOne},
			ParseStatus: ParseStatusSupported,
		},
		{
			Rule:        FirewallRule{Scope: scope, Protocol: "tcp", DestinationPort: "80", Action: ActionAccept},
			Locator:     Locator{Position: &positionTwo},
			ParseStatus: ParseStatusSupported,
		},
	}

	first, err := SnapshotRevision(scope, rules)
	if err != nil {
		t.Fatalf("first snapshot revision: %v", err)
	}
	reversed := []ObservedRule{rules[1], rules[0]}
	second, err := SnapshotRevision(scope, reversed)
	if err != nil {
		t.Fatalf("second snapshot revision: %v", err)
	}
	if first != second {
		t.Fatalf("slice order changed snapshot revision:\n%s\n%s", first, second)
	}

	rules[0].Locator.Position = &positionTwo
	changed, err := SnapshotRevision(scope, rules)
	if err != nil {
		t.Fatalf("changed snapshot revision: %v", err)
	}
	if changed == first {
		t.Fatal("locator position change did not change snapshot revision")
	}
}

func TestSnapshotRevisionSupportsOpaqueRules(t *testing.T) {
	scope := Scope{Provider: ProviderFirewalld, Family: FamilyInet, Zone: "public", Direction: DirectionInput}
	position := 1
	rules := []ObservedRule{
		{
			Rule:        FirewallRule{Scope: scope},
			Locator:     Locator{Position: &position, Canonical: "service:ssh"},
			ParseStatus: ParseStatusOpaque,
			Raw:         "services: ssh",
		},
	}
	revision, err := SnapshotRevision(scope, rules)
	if err != nil {
		t.Fatalf("opaque snapshot revision: %v", err)
	}
	if revision == "" {
		t.Fatal("expected non-empty opaque snapshot revision")
	}
	instanceKey, err := InstanceKey(rules[0])
	if err != nil || instanceKey == "" {
		t.Fatalf("opaque instance key: key=%q err=%v", instanceKey, err)
	}
	rules[0].Persistence = PersistenceStatusPermanentOnly
	changed, err := SnapshotRevision(scope, rules)
	if err != nil {
		t.Fatalf("opaque persistence revision: %v", err)
	}
	if changed == revision {
		t.Fatal("opaque runtime/permanent presence did not change snapshot revision")
	}
}

func TestSnapshotRevisionRequiresPositionForOrderedProvider(t *testing.T) {
	scope := Scope{Provider: ProviderUFW, Family: FamilyIPv4, Direction: DirectionInput}
	_, err := SnapshotRevision(scope, []ObservedRule{
		{
			Rule:        FirewallRule{Scope: scope, Protocol: "tcp", DestinationPort: "22", Action: ActionAccept},
			ParseStatus: ParseStatusSupported,
		},
	})
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("expected missing position error, got %v", err)
	}
}
