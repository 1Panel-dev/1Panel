package filter

import "testing"

func TestMergeInventoryPreservesObservedOrderAndOwnership(t *testing.T) {
	ssh := inventoryTestRule("22")
	http := inventoryTestRule("80")
	https := inventoryTestRule("443")
	sshObserved := inventoryObservedRule(t, ssh, "onepanel:created:ssh", 1)
	httpObserved := inventoryObservedRule(t, http, "", 2)
	protectedKey, _ := RuleKey(http)

	items, err := MergeInventory(InventoryMergeInput{
		Observed: []ObservedRule{httpObserved, sshObserved},
		Desired: []DesiredRule{
			{UUID: "ssh", Rule: ssh, Origin: RuleOriginCreated, Marker: "onepanel:created:ssh"},
			{UUID: "https", Rule: https, Origin: RuleOriginAdopted},
		},
		ProtectedObservedKeys: map[string]struct{}{protectedKey: {}},
	})
	if err != nil {
		t.Fatalf("merge inventory: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("unexpected item count: %#v", items)
	}
	if items[0].Rule.DestinationPort != "80" || items[0].State != InventoryStateProtected || items[0].Desired != nil {
		t.Fatalf("observed order/protected classification changed: %#v", items[0])
	}
	if items[1].Rule.DestinationPort != "22" || items[1].State != InventoryStateManaged || items[1].Match != InventoryMatchExact {
		t.Fatalf("managed rule was not matched: %#v", items[1])
	}
	if items[2].Rule.DestinationPort != "443" || items[2].State != InventoryStateDrifted || items[2].Match != InventoryMatchMissing || items[2].Observed != nil {
		t.Fatalf("missing desired rule was not appended as drifted: %#v", items[2])
	}
}

func TestMergeInventoryUsesDesiredDescriptionForManagedRule(t *testing.T) {
	rule := inventoryTestRule("8080")
	desired := rule
	desired.Description = "user description"
	observed := inventoryObservedRule(t, rule, "1panel-rule:web", 1)

	items, err := MergeInventory(InventoryMergeInput{
		Observed: []ObservedRule{observed},
		Desired: []DesiredRule{{
			UUID: "web", Rule: desired, Origin: RuleOriginCreated, Marker: observed.Marker,
		}},
	})
	if err != nil {
		t.Fatalf("merge inventory: %v", err)
	}
	if len(items) != 1 || items[0].Rule.Description != desired.Description {
		t.Fatalf("managed description was not taken from desired state: %#v", items)
	}
	if items[0].Observed == nil || items[0].Observed.Rule.Description != "" {
		t.Fatalf("observed runtime rule was modified: %#v", items[0].Observed)
	}
}

func TestMergeInventoryUsesInstanceBeforeSemanticIdentity(t *testing.T) {
	rule := inventoryTestRule("22")
	first := inventoryObservedRule(t, rule, "", 1)
	second := inventoryObservedRule(t, rule, "", 2)
	secondKey, err := InstanceKey(second)
	if err != nil {
		t.Fatalf("instance key: %v", err)
	}

	items, err := MergeInventory(InventoryMergeInput{
		Observed: []ObservedRule{first, second},
		Desired: []DesiredRule{{
			UUID: "adopted", Rule: rule, Origin: RuleOriginAdopted, ObservedInstanceKey: secondKey,
		}},
	})
	if err != nil {
		t.Fatalf("merge inventory: %v", err)
	}
	if items[0].State != InventoryStateExternal || items[1].State != InventoryStateAdopted || items[1].Observed.Locator.Position == nil || *items[1].Observed.Locator.Position != 2 {
		t.Fatalf("instance locator did not select the intended candidate: %#v", items)
	}
}

func TestMergeInventoryDoesNotGuessBetweenEquivalentRules(t *testing.T) {
	rule := inventoryTestRule("3306")
	items, err := MergeInventory(InventoryMergeInput{
		Observed: []ObservedRule{
			inventoryObservedRule(t, rule, "", 1),
			inventoryObservedRule(t, rule, "", 2),
		},
		Desired: []DesiredRule{{UUID: "managed", Rule: rule, Origin: RuleOriginCreated}},
	})
	if err != nil {
		t.Fatalf("merge inventory: %v", err)
	}
	if len(items) != 3 || items[0].State != InventoryStateExternal || items[1].State != InventoryStateExternal || items[2].State != InventoryStateDrifted || items[2].Match != InventoryMatchAmbiguous {
		t.Fatalf("ambiguous candidates were guessed: %#v", items)
	}
}

func TestMergeInventoryUsesMarkerToReportChangedManagedRule(t *testing.T) {
	desiredRule := inventoryTestRule("80")
	changedRule := inventoryTestRule("8080")
	marker := "onepanel:created:web"
	items, err := MergeInventory(InventoryMergeInput{
		Observed: []ObservedRule{inventoryObservedRule(t, changedRule, marker, 1)},
		Desired: []DesiredRule{{
			UUID: "web", Rule: desiredRule, Origin: RuleOriginCreated, Marker: marker,
		}},
	})
	if err != nil {
		t.Fatalf("merge inventory: %v", err)
	}
	if len(items) != 1 || items[0].State != InventoryStateDrifted || items[0].Match != InventoryMatchChanged || items[0].Observed == nil || items[0].Desired == nil {
		t.Fatalf("marker-owned semantic drift was not retained: %#v", items)
	}
}

func TestMergeInventoryMarkerSurvivesTransientLocatorChange(t *testing.T) {
	rule := inventoryTestRule("443")
	marker := "1panel-rule:https"
	previous := inventoryObservedRule(t, rule, marker, 2)
	previousKey, err := InstanceKey(previous)
	if err != nil {
		t.Fatalf("previous instance key: %v", err)
	}
	current := inventoryObservedRule(t, rule, marker, 7)
	items, err := MergeInventory(InventoryMergeInput{
		Observed: []ObservedRule{current},
		Desired: []DesiredRule{{
			UUID: "https", Rule: rule, Origin: RuleOriginCreated,
			Marker: marker, ObservedInstanceKey: previousKey,
		}},
	})
	if err != nil {
		t.Fatalf("merge locator drift: %v", err)
	}
	if len(items) != 1 || items[0].State != InventoryStateManaged || items[0].Match != InventoryMatchExact ||
		items[0].Observed == nil || items[0].Observed.Locator.Position == nil || *items[0].Observed.Locator.Position != 7 {
		t.Fatalf("stable marker did not survive transient locator change: %#v", items)
	}
}

func TestMergeInventoryKeepsOpaqueRulesExternal(t *testing.T) {
	opaque := ObservedRule{
		Rule:        FirewallRule{Scope: Scope{Provider: ProviderFirewalld, Family: FamilyInet, Zone: "public", Direction: DirectionInput}},
		Locator:     Locator{Provider: ProviderFirewalld, ScopeKey: "firewalld:public:input", NativeID: "service:ssh"},
		ParseStatus: ParseStatusOpaque,
		Raw:         "ssh service",
	}
	items, err := MergeInventory(InventoryMergeInput{Observed: []ObservedRule{opaque}})
	if err != nil {
		t.Fatalf("merge opaque inventory: %v", err)
	}
	if len(items) != 1 || items[0].State != InventoryStateExternal || items[0].Match != InventoryMatchOpaque || items[0].Observed.Raw != opaque.Raw {
		t.Fatalf("opaque rule was not preserved: %#v", items)
	}
}

func TestMergeInventoryUsesAdapterProtectedClassification(t *testing.T) {
	rule := inventoryTestRule("22")
	observed := inventoryObservedRule(t, rule, "", 1)
	observed.Protected = true

	items, err := MergeInventory(InventoryMergeInput{Observed: []ObservedRule{observed}})
	if err != nil {
		t.Fatalf("merge protected inventory: %v", err)
	}
	if len(items) != 1 || items[0].State != InventoryStateProtected || items[0].Observed == nil || !items[0].Observed.Protected {
		t.Fatalf("adapter protected classification was lost: %#v", items)
	}
}

func TestMergeInventoryReportsRuntimePermanentDrift(t *testing.T) {
	rule := inventoryTestRule("443")
	observed := inventoryObservedRule(t, rule, "onepanel:created:https", 1)
	observed.Persistence = PersistenceStatusPermanentOnly
	items, err := MergeInventory(InventoryMergeInput{
		Observed: []ObservedRule{observed},
		Desired: []DesiredRule{{
			UUID: "https", Rule: rule, Origin: RuleOriginCreated, Marker: observed.Marker,
		}},
	})
	if err != nil {
		t.Fatalf("merge persistence drift: %v", err)
	}
	if len(items) != 1 || items[0].State != InventoryStateDrifted || items[0].Match != InventoryMatchExact {
		t.Fatalf("runtime/permanent drift was hidden: %#v", items)
	}
}

func TestAttachRuntimeUsageDoesNotChangeOwnership(t *testing.T) {
	rule := inventoryTestRule("8080")
	items := []InventoryItem{{Rule: rule, State: InventoryStateAdopted, Match: InventoryMatchExact}}
	usage := map[string]RuntimeUsage{
		RuntimeUsageKey(rule): {UsedBy: []string{"nginx", "", "nginx", "1panel"}, Reason: "listener"},
	}

	result := AttachRuntimeUsage(items, usage)
	if result[0].State != InventoryStateAdopted || result[0].Usage == nil || !result[0].Usage.Used {
		t.Fatalf("usage changed ownership or was not attached: %#v", result[0])
	}
	if len(result[0].Usage.UsedBy) != 2 || result[0].Usage.UsedBy[0] != "1panel" || result[0].Usage.UsedBy[1] != "nginx" {
		t.Fatalf("usage owners were not normalized: %#v", result[0].Usage)
	}
	if items[0].Usage != nil {
		t.Fatal("AttachRuntimeUsage mutated its input")
	}
}

func TestAttachRuntimeUsageAggregatesPortRanges(t *testing.T) {
	rule := inventoryTestRule("8000-8010")
	items := []InventoryItem{{Rule: rule}}
	usage := map[string]RuntimeUsage{
		"tcp\x008001": {Used: true, UsedBy: []string{"app"}, Reason: "application"},
		"tcp\x008009": {Used: true, UsedBy: []string{"worker"}, Reason: "listener"},
		"udp\x008005": {Used: true, UsedBy: []string{"ignored"}},
	}
	result := AttachRuntimeUsage(items, usage)
	if result[0].Usage == nil || len(result[0].Usage.UsedBy) != 2 || result[0].Usage.Reason != "multiple" {
		t.Fatalf("range usage was not aggregated: %#v", result[0].Usage)
	}
}

func TestMergeInventoryRejectsStoredRuleKeyMismatch(t *testing.T) {
	_, err := MergeInventory(InventoryMergeInput{Desired: []DesiredRule{{
		UUID: "broken", Rule: inventoryTestRule("53"), RuleKey: "sha256:stale", Origin: RuleOriginCreated,
	}}})
	if err == nil {
		t.Fatal("expected stored rule key mismatch to fail")
	}
}

func inventoryTestRule(port string) FirewallRule {
	return FirewallRule{
		Scope:           Scope{Provider: ProviderIptables, Family: FamilyIPv4, Table: "filter", Chain: "1PANEL_BASIC", Direction: DirectionInput},
		NativeKind:      NativeKindRule,
		Protocol:        "tcp",
		DestinationPort: port,
		Action:          ActionAccept,
	}
}

func inventoryObservedRule(t *testing.T, rule FirewallRule, marker string, position int) ObservedRule {
	t.Helper()
	return ObservedRule{
		Rule: rule,
		Locator: Locator{
			Provider: rule.Scope.Provider,
			ScopeKey: rule.Scope.Key(),
			Position: &position,
		},
		Marker:      marker,
		ParseStatus: ParseStatusSupported,
	}
}
