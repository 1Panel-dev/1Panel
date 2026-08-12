package filter

import "testing"

func TestProtectSnapshotTreatsAllProtocolAsCoveringProtectedTransport(t *testing.T) {
	scope := Scope{Provider: ProviderUFW, Family: FamilyIPv4, Chain: UFWInputChain, Direction: DirectionInput}
	rules := []ObservedRule{
		protectedPortTestRule(scope, "all", "22", 1),
		protectedPortTestRule(scope, "udp", "22", 2),
		protectedPortTestRule(scope, "all", "53", 3),
	}
	snapshot, err := NewSnapshot(scope, rules)
	if err != nil {
		t.Fatalf("create snapshot: %v", err)
	}

	protected, err := ProtectSnapshot(snapshot, []ProtectedPort{{Port: "22", Protocol: "tcp"}})
	if err != nil {
		t.Fatalf("protect snapshot: %v", err)
	}
	if !protected.Rules[0].Protected {
		t.Fatal("all-protocol port 22 did not cover protected 22/tcp")
	}
	if protected.Rules[1].Protected {
		t.Fatal("22/udp incorrectly matched protected 22/tcp")
	}
	if protected.Rules[2].Protected {
		t.Fatal("unrelated all-protocol port was protected")
	}
}

func TestProtectSnapshotMarksBareUFWPortForBothFamilies(t *testing.T) {
	for _, family := range []Family{FamilyIPv4, FamilyIPv6} {
		scope := Scope{Provider: ProviderUFW, Family: family, Chain: UFWInputChain, Direction: DirectionInput}
		snapshot, err := NewSnapshot(scope, []ObservedRule{protectedPortTestRule(scope, "all", "22", 1)})
		if err != nil {
			t.Fatalf("create %s snapshot: %v", family, err)
		}
		protected, err := ProtectSnapshot(snapshot, []ProtectedPort{{Port: "22", Protocol: "tcp"}})
		if err != nil {
			t.Fatalf("protect %s snapshot: %v", family, err)
		}
		if !protected.Rules[0].Protected {
			t.Fatalf("bare UFW port 22 was not protected for %s", family)
		}
	}
}

func TestProtectSnapshotMatchesPortSetsAndRanges(t *testing.T) {
	scope := Scope{Provider: ProviderIptables, Family: FamilyIPv4, Table: "filter", Chain: IptablesInputChain, Direction: DirectionInput}
	rules := []ObservedRule{
		protectedPortTestRule(scope, "tcp", "22,80,443", 1),
		protectedPortTestRule(scope, "tcp", "8000-9000", 2),
		protectedPortTestRule(scope, "udp", "22,443", 3),
	}
	snapshot, err := NewSnapshot(scope, rules)
	if err != nil {
		t.Fatalf("create snapshot: %v", err)
	}
	protected, err := ProtectSnapshot(snapshot, []ProtectedPort{{Port: "22", Protocol: "tcp"}, {Port: "8080", Protocol: "tcp"}})
	if err != nil {
		t.Fatalf("protect snapshot: %v", err)
	}
	if !protected.Rules[0].Protected || !protected.Rules[1].Protected || protected.Rules[2].Protected {
		t.Fatalf("unexpected port-set protection: %#v", protected.Rules)
	}
}

func protectedPortTestRule(scope Scope, protocol, port string, position int) ObservedRule {
	return ObservedRule{
		Rule: FirewallRule{
			Scope: scope, NativeKind: NativeKindUFWRule, Protocol: protocol,
			DestinationPort: port, Action: ActionAccept,
		},
		Locator: Locator{
			Provider: scope.Provider, ScopeKey: scope.Key(), Position: &position,
		},
		ParseStatus: ParseStatusSupported,
	}
}
