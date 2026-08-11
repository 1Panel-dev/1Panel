package iptables

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

func TestCompileCreateAndAdoptCommands(t *testing.T) {
	scope := testScope("1PANEL_BASIC")
	snapshot, err := filter.NewSnapshot(scope, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	rule := filter.FirewallRule{
		UUID: "rule-1", Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp",
		SourceAddress: "10.0.0.0/8", DestinationPort: "443", Action: filter.ActionAccept,
	}
	adapter := NewAdapterWithReader(&fakeRuleReader{})
	plan, err := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
	if err != nil {
		t.Fatalf("compile create: %v", err)
	}
	want := []string{"-w", "-t", "filter", "-I", "1PANEL_BASIC", "1", "-p", "tcp", "-s", "10.0.0.0/8", "--dport", "443", "-m", "comment", "--comment", "1panel-rule:rule-1", "-j", "ACCEPT"}
	if len(plan.Rules) != 1 || !reflect.DeepEqual(plan.Rules[0].Commands[0].Args, want) || plan.Rules[0].Expected.Marker != "1panel-rule:rule-1" {
		t.Fatalf("unexpected create plan: %#v", plan)
	}

	position := 1
	adoptSnapshot, _ := filter.NewSnapshot(scope, []filter.ObservedRule{executorObserved(rule, position)})
	plan, err = adapter.Compile(adoptSnapshot, []filter.DesiredChange{{Operation: filter.ChangeAdopt, After: &rule, Locator: &filter.Locator{Position: &position}}})
	if err != nil {
		t.Fatalf("compile adopt: %v", err)
	}
	if plan.Rules[0].Commands[0].Args[3] != "-R" || plan.Rules[0].Commands[0].Args[5] != "1" {
		t.Fatalf("adoption did not replace the selected position: %#v", plan.Rules[0])
	}
}

func TestIPv6ObserveCompileAndCapabilities(t *testing.T) {
	scope := testScopeFamily("1PANEL_BASIC", filter.FamilyIPv6)
	reader := &fakeRuleReader{output: "-A 1PANEL_BASIC -p ipv6-icmp -s 2001:db8::/64 -m comment --comment \"1panel-rule:ping6\" -j ACCEPT"}
	adapter := NewAdapterWithReader(reader)
	snapshot, err := adapter.Observe(context.Background(), scope)
	if err != nil {
		t.Fatalf("observe IPv6 rules: %v", err)
	}
	if len(snapshot.Rules) != 1 || snapshot.Rules[0].Rule.Protocol != "icmpv6" ||
		snapshot.Rules[0].Rule.SourceAddress != "2001:db8::/64" {
		t.Fatalf("unexpected IPv6 snapshot: %#v", snapshot)
	}
	rule := filter.FirewallRule{
		UUID: "ping6", Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "icmpv6",
		SourceAddress: "2001:db8::/64", Action: filter.ActionAccept,
	}
	empty, _ := filter.NewSnapshot(scope, nil)
	plan, err := adapter.Compile(empty, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
	if err != nil {
		t.Fatalf("compile IPv6 rule: %v", err)
	}
	command := plan.Rules[0].Commands[0]
	if command.Executable != "ip6tables" || !reflect.DeepEqual(command.Args[6:10], []string{"-p", "ipv6-icmp", "-s", "2001:db8::/64"}) {
		t.Fatalf("unexpected IPv6 command: %#v", command)
	}
	capabilities, err := adapter.Capabilities(context.Background())
	if err != nil || !capabilities.SupportsScope(scope) || !capabilities.SupportsScope(testScope("1PANEL_BASIC")) {
		t.Fatalf("IPv4/IPv6 capabilities are incomplete: %#v err=%v", capabilities, err)
	}
}

func TestCompileRejectsICMPFamilyMismatch(t *testing.T) {
	for _, test := range []struct {
		family   filter.Family
		protocol string
	}{
		{family: filter.FamilyIPv4, protocol: "icmpv6"},
		{family: filter.FamilyIPv6, protocol: "icmp"},
	} {
		scope := testScopeFamily("1PANEL_BASIC", test.family)
		snapshot, _ := filter.NewSnapshot(scope, nil)
		rule := filter.FirewallRule{
			UUID: "icmp", Scope: scope, NativeKind: filter.NativeKindRule,
			Protocol: test.protocol, Action: filter.ActionAccept,
		}
		_, err := NewAdapterWithReader(&fakeRuleReader{}).Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
		if !errors.Is(err, filter.ErrInvalidRule) {
			t.Fatalf("expected %s/%s to be rejected, got %v", test.family, test.protocol, err)
		}
	}
}

func TestApplyRejectsCrossFamilyExecutable(t *testing.T) {
	scope := testScopeFamily("1PANEL_BASIC", filter.FamilyIPv6)
	snapshot, _ := filter.NewSnapshot(scope, nil)
	rule := filter.FirewallRule{
		UUID: "web6", Scope: scope, NativeKind: filter.NativeKindRule,
		Protocol: "tcp", DestinationPort: "443", Action: filter.ActionAccept,
	}
	writer := &fakeRuleWriter{}
	adapter := NewAdapterWithBackend(&fakeRuleReader{}, writer)
	plan, err := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
	if err != nil {
		t.Fatalf("compile IPv6 rule: %v", err)
	}
	plan.Rules[0].Commands[0].Executable = "iptables"
	if _, err := adapter.Apply(context.Background(), plan); !errors.Is(err, filter.ErrInvalidRule) {
		t.Fatalf("expected cross-family command rejection, got %v", err)
	}
	if len(writer.commands) != 0 {
		t.Fatalf("cross-family command reached the writer: %#v", writer.commands)
	}
}

func TestCompileRejectsBroadIPv4AndIPv6Deny(t *testing.T) {
	for _, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		scope := testScopeFamily("1PANEL_BASIC", family)
		snapshot, _ := filter.NewSnapshot(scope, nil)
		rule := filter.FirewallRule{
			UUID: "deny-all", Scope: scope, NativeKind: filter.NativeKindRule,
			Protocol: "all", Action: filter.ActionDrop,
		}
		_, err := NewAdapterWithReader(&fakeRuleReader{}).Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
		if !errors.Is(err, filter.ErrLockoutRisk) {
			t.Fatalf("expected broad %s deny to be rejected, got %v", family, err)
		}
	}
}

func TestCompileInsertsAllowBeforeTerminalDrop(t *testing.T) {
	scope := testScope("1PANEL_BASIC_AFTER")
	dropTCP := filter.FirewallRule{Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", Action: filter.ActionDrop}
	dropUDP := filter.FirewallRule{Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "udp", Action: filter.ActionDrop}
	snapshot, _ := filter.NewSnapshot(scope, []filter.ObservedRule{executorObserved(dropTCP, 1), executorObserved(dropUDP, 2)})
	allow := filter.FirewallRule{UUID: "dns", Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "udp", DestinationPort: "53", Action: filter.ActionAccept}
	plan, err := NewAdapterWithReader(&fakeRuleReader{}).Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &allow}})
	if err != nil {
		t.Fatalf("compile terminal insert: %v", err)
	}
	if plan.Rules[0].Commands[0].Args[5] != "1" {
		t.Fatalf("allow rule was inserted after terminal drop: %#v", plan.Rules[0].Commands[0])
	}
}

func TestCompileCreateUsesRequestedPosition(t *testing.T) {
	scope := testScope("1PANEL_BASIC")
	first := filter.FirewallRule{Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: "80", Action: filter.ActionAccept}
	second := filter.FirewallRule{Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: "81", Action: filter.ActionAccept}
	snapshot, _ := filter.NewSnapshot(scope, []filter.ObservedRule{executorObserved(first, 1), executorObserved(second, 2)})
	position := int64(2)
	rule := filter.FirewallRule{
		UUID: "inserted", Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp",
		DestinationPort: "443", Action: filter.ActionAccept, OrderIndex: &position,
	}
	plan, err := NewAdapterWithReader(&fakeRuleReader{}).Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
	if err != nil {
		t.Fatalf("compile positioned create: %v", err)
	}
	if got := plan.Rules[0].Commands[0].Args[5]; got != "2" {
		t.Fatalf("create ignored requested position: %#v", plan.Rules[0].Commands[0])
	}
}

func TestCompileReordersManagedRuleWithinChain(t *testing.T) {
	scope := testScope("1PANEL_BASIC")
	first := filter.FirewallRule{UUID: "first", Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: "80", Action: filter.ActionAccept}
	second := filter.FirewallRule{UUID: "second", Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: "81", Action: filter.ActionAccept}
	third := filter.FirewallRule{UUID: "third", Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: "82", Action: filter.ActionAccept}
	rules := []filter.ObservedRule{executorObserved(first, 1), executorObserved(second, 2), executorObserved(third, 3)}
	for index := range rules {
		rules[index].Marker = "1panel-rule:" + rules[index].Rule.UUID
	}
	snapshot, _ := filter.NewSnapshot(scope, rules)
	target := int64(3)
	after := first
	after.OrderIndex = &target
	plan, err := NewAdapterWithReader(&fakeRuleReader{}).Compile(snapshot, []filter.DesiredChange{{
		Operation: filter.ChangeReorder, Before: &first, After: &after, Locator: &rules[0].Locator,
	}})
	if err != nil {
		t.Fatalf("compile reorder: %v", err)
	}
	rulePlan := plan.Rules[0]
	if len(rulePlan.Commands) != 2 || rulePlan.Commands[0].Args[3] != "-D" || rulePlan.Commands[0].Args[5] != "1" ||
		rulePlan.Commands[1].Args[3] != "-I" || rulePlan.Commands[1].Args[5] != "3" ||
		rulePlan.Expected.Locator.Position == nil || *rulePlan.Expected.Locator.Position != 3 {
		t.Fatalf("unexpected reorder plan: %#v", rulePlan)
	}
}

func TestCompileUpdateMovesAndChangesManagedRule(t *testing.T) {
	scope := testScope("1PANEL_BASIC")
	first := filter.FirewallRule{UUID: "first", Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: "80", Action: filter.ActionAccept}
	second := filter.FirewallRule{UUID: "second", Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: "81", Action: filter.ActionAccept}
	third := filter.FirewallRule{UUID: "third", Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: "82", Action: filter.ActionAccept}
	rules := []filter.ObservedRule{executorObserved(first, 1), executorObserved(second, 2), executorObserved(third, 3)}
	for index := range rules {
		rules[index].Marker = "1panel-rule:" + rules[index].Rule.UUID
	}
	snapshot, _ := filter.NewSnapshot(scope, rules)
	target := int64(3)
	after := first
	after.DestinationPort = "443"
	after.OrderIndex = &target
	plan, err := NewAdapterWithReader(&fakeRuleReader{}).Compile(snapshot, []filter.DesiredChange{{
		Operation: filter.ChangeUpdate, Before: &first, After: &after, Locator: &rules[0].Locator,
	}})
	if err != nil {
		t.Fatalf("compile positioned update: %v", err)
	}
	rulePlan := plan.Rules[0]
	if len(rulePlan.Commands) != 2 || rulePlan.Commands[0].Args[3] != "-D" || rulePlan.Commands[1].Args[5] != "3" ||
		!slices.Contains(rulePlan.Commands[1].Args, "443") || rulePlan.Expected.Locator.Position == nil ||
		*rulePlan.Expected.Locator.Position != 3 {
		t.Fatalf("unexpected positioned update plan: %#v", rulePlan)
	}
}

func TestCompileBlocksReorderAcrossExternalOrProtectedRule(t *testing.T) {
	scope := testScope("1PANEL_BASIC")
	first := filter.FirewallRule{UUID: "first", Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: "80", Action: filter.ActionAccept}
	middle := filter.FirewallRule{Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: "81", Action: filter.ActionAccept}
	last := filter.FirewallRule{UUID: "last", Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: "82", Action: filter.ActionAccept}
	rules := []filter.ObservedRule{executorObserved(first, 1), executorObserved(middle, 2), executorObserved(last, 3)}
	rules[0].Marker = "1panel-rule:first"
	rules[2].Marker = "1panel-rule:last"
	snapshot, _ := filter.NewSnapshot(scope, rules)
	target := int64(3)
	after := first
	after.OrderIndex = &target
	change := []filter.DesiredChange{{Operation: filter.ChangeReorder, Before: &first, After: &after, Locator: &rules[0].Locator}}
	if _, err := NewAdapterWithReader(&fakeRuleReader{}).Compile(snapshot, change); !errors.Is(err, filter.ErrUnsupportedScope) {
		t.Fatalf("expected external boundary rejection, got %v", err)
	}
	rules[1].Marker = "1panel-rule:middle"
	rules[1].Protected = true
	snapshot, _ = filter.NewSnapshot(scope, rules)
	change[0].Locator = &snapshot.Rules[0].Locator
	if _, err := NewAdapterWithReader(&fakeRuleReader{}).Compile(snapshot, change); !errors.Is(err, filter.ErrProtectedRule) {
		t.Fatalf("expected protected boundary rejection, got %v", err)
	}
}

func TestApplyUsesCompiledSnapshotWithoutSecondRead(t *testing.T) {
	scope := testScope("1PANEL_BASIC")
	initial, _ := filter.NewSnapshot(scope, nil)
	rule := filter.FirewallRule{UUID: "web", Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: "80", Action: filter.ActionAccept}
	reader := &fakeRuleReader{}
	writer := &fakeRuleWriter{}
	adapter := NewAdapterWithBackend(reader, writer)
	plan, err := adapter.Compile(initial, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
	if err != nil {
		t.Fatalf("compile create: %v", err)
	}
	reader.output = "-A 1PANEL_BASIC -p tcp --dport 22 -j ACCEPT"
	if _, err := adapter.Apply(context.Background(), plan); err != nil {
		t.Fatalf("apply compiled plan: %v", err)
	}
	if len(writer.commands) != 1 || !writer.saved {
		t.Fatalf("compiled plan was not applied: %#v", writer)
	}
}

func TestApplyAndVerifyMarker(t *testing.T) {
	scope := testScope("1PANEL_BASIC")
	reader := &fakeRuleReader{}
	writer := &fakeRuleWriter{}
	adapter := NewAdapterWithBackend(reader, writer)
	snapshot, _ := filter.NewSnapshot(scope, nil)
	rule := filter.FirewallRule{UUID: "ssh", Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: "22", Action: filter.ActionAccept}
	plan, _ := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
	if _, err := adapter.Apply(context.Background(), plan); err != nil {
		t.Fatalf("apply plan: %v", err)
	}
	if len(writer.commands) != 1 || !writer.saved {
		t.Fatalf("command was not executed and persisted: %#v", writer)
	}
	reader.output = `-A 1PANEL_BASIC -p tcp --dport 22 -m comment --comment "1panel-rule:ssh" -j ACCEPT`
	verified, err := adapter.Verify(context.Background(), plan)
	if err != nil || !verified.Matched {
		t.Fatalf("verify marker: result=%#v err=%v", verified, err)
	}
}

func TestApplyCompensatesWhenPersistenceFails(t *testing.T) {
	scope := testScope("1PANEL_BASIC")
	reader := &fakeRuleReader{}
	writer := &fakeRuleWriter{saveErrors: []error{errors.New("disk full"), nil}}
	adapter := NewAdapterWithBackend(reader, writer)
	snapshot, _ := filter.NewSnapshot(scope, nil)
	rule := filter.FirewallRule{UUID: "ssh", Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: "22", Action: filter.ActionAccept}
	plan, _ := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})

	if _, err := adapter.Apply(context.Background(), plan); err == nil || err.Error() != "disk full" {
		t.Fatalf("expected original persistence error, got %v", err)
	}
	if len(writer.commands) != 2 || writer.commands[0].Args[3] != "-I" || writer.commands[1].Args[3] != "-D" || writer.saveCalls != 2 {
		t.Fatalf("failed write was not compensated and persisted: %#v", writer)
	}
}

func TestRollbackReversesFullyAppliedIptablesPlan(t *testing.T) {
	scope := testScope("1PANEL_BASIC")
	writer := &fakeRuleWriter{}
	adapter := NewAdapterWithBackend(&fakeRuleReader{}, writer)
	snapshot, _ := filter.NewSnapshot(scope, nil)
	rule := filter.FirewallRule{UUID: "rollback", Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: "8080", Action: filter.ActionAccept}
	plan, err := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
	if err != nil {
		t.Fatalf("compile rollback plan: %v", err)
	}
	if err := adapter.Rollback(context.Background(), plan); err != nil {
		t.Fatalf("rollback applied plan: %v", err)
	}
	if len(writer.commands) != 1 || writer.commands[0].Args[3] != "-D" || writer.saveCalls != 1 {
		t.Fatalf("unexpected rollback writes: %#v", writer)
	}
}

func TestApplyCompensatesPartiallyExecutedReorder(t *testing.T) {
	scope := testScope("1PANEL_BASIC")
	reader := &fakeRuleReader{output: "-A 1PANEL_BASIC -p tcp --dport 80 -m comment --comment 1panel-rule:first -j ACCEPT\n" +
		"-A 1PANEL_BASIC -p tcp --dport 81 -m comment --comment 1panel-rule:second -j ACCEPT"}
	snapshot, err := NewAdapterWithReader(reader).Observe(context.Background(), scope)
	if err != nil {
		t.Fatalf("observe reorder snapshot: %v", err)
	}
	first := snapshot.Rules[0].Rule
	first.UUID = "first"
	target := int64(2)
	after := first
	after.OrderIndex = &target
	plan, err := NewAdapterWithReader(reader).Compile(snapshot, []filter.DesiredChange{{
		Operation: filter.ChangeReorder, Before: &first, After: &after, Locator: &snapshot.Rules[0].Locator,
	}})
	if err != nil {
		t.Fatalf("compile reorder: %v", err)
	}
	writer := &fakeRuleWriter{runErrors: []error{nil, errors.New("insert failed"), nil}}
	adapter := NewAdapterWithBackend(reader, writer)
	if _, err := adapter.Apply(context.Background(), plan); err == nil || err.Error() != "insert failed" {
		t.Fatalf("expected reorder insert failure, got %v", err)
	}
	if len(writer.commands) != 3 || writer.commands[0].Args[3] != "-D" || writer.commands[2].Args[3] != "-I" || writer.saveCalls != 1 {
		t.Fatalf("partial reorder was not restored: %#v", writer)
	}
}

func TestVerifyDeleteRequiresMarkerToDisappear(t *testing.T) {
	scope := testScope("1PANEL_BASIC")
	rule := filter.FirewallRule{UUID: "ssh", Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: "22", Action: filter.ActionAccept}
	position := 1
	observed := executorObserved(rule, position)
	observed.Marker = "1panel-rule:ssh"
	snapshot, _ := filter.NewSnapshot(scope, []filter.ObservedRule{observed})
	plan, err := NewAdapterWithReader(&fakeRuleReader{}).Compile(snapshot, []filter.DesiredChange{
		{Operation: filter.ChangeDelete, Before: &rule, Locator: &observed.Locator},
	})
	if err != nil {
		t.Fatalf("compile delete: %v", err)
	}
	reader := &fakeRuleReader{output: `-A 1PANEL_BASIC -p tcp --dport 2200 -m comment --comment "1panel-rule:ssh" -j ACCEPT`}
	verified, err := NewAdapterWithReader(reader).Verify(context.Background(), plan)
	if err != nil || verified.Matched {
		t.Fatalf("delete verification ignored a surviving marker: result=%#v err=%v", verified, err)
	}
}

func TestCompileRejectsProtectedMutation(t *testing.T) {
	scope := testScope("1PANEL_BASIC_BEFORE")
	rule := filter.FirewallRule{UUID: "loopback", Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "all", Interface: "lo", Action: filter.ActionAccept}
	position := 1
	observed := executorObserved(rule, position)
	observed.Protected = true
	snapshot, _ := filter.NewSnapshot(scope, []filter.ObservedRule{observed})

	_, err := NewAdapterWithReader(&fakeRuleReader{}).Compile(snapshot, []filter.DesiredChange{
		{Operation: filter.ChangeDelete, Before: &rule, Locator: &observed.Locator},
	})
	if !errors.Is(err, filter.ErrProtectedRule) {
		t.Fatalf("expected protected mutation rejection, got %v", err)
	}
}

func TestObserveMergesIPPortAndCombinedRulesInNativeOrder(t *testing.T) {
	scope := testScope("1PANEL_BASIC")
	reader := &fakeRuleReader{output: `-N 1PANEL_BASIC
-A 1PANEL_BASIC -s 172.16.10.111/32 -j DROP
-A 1PANEL_BASIC -p tcp -m tcp --dport 22 -j ACCEPT -m comment --comment "ssh"
-A 1PANEL_BASIC -p tcp -m tcp -s 10.0.0.0/8 --dport 443 -j ACCEPT -m comment --comment "1panel-rule:managed"
`}
	snapshot, err := NewAdapterWithReader(reader).Observe(context.Background(), scope)
	if err != nil {
		t.Fatalf("observe rules: %v", err)
	}
	if len(snapshot.Rules) != 3 || snapshot.Revision == "" {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
	if snapshot.Rules[0].Rule.SourceAddress != "172.16.10.111/32" || snapshot.Rules[0].Rule.DestinationPort != "" {
		t.Fatalf("IP rule was not preserved: %#v", snapshot.Rules[0])
	}
	if snapshot.Rules[1].Rule.DestinationPort != "22" || snapshot.Rules[1].Rule.Description != "ssh" {
		t.Fatalf("port rule was not preserved: %#v", snapshot.Rules[1])
	}
	if snapshot.Rules[2].Rule.SourceAddress != "10.0.0.0/8" || snapshot.Rules[2].Rule.DestinationPort != "443" || snapshot.Rules[2].Marker != "1panel-rule:managed" {
		t.Fatalf("combined managed rule was not preserved: %#v", snapshot.Rules[2])
	}
	for index, observed := range snapshot.Rules {
		if observed.Locator.Position == nil || *observed.Locator.Position != index+1 {
			t.Fatalf("native position was not preserved: %#v", observed.Locator)
		}
	}
}

func TestObserveKeepsUnsupportedRulesOpaque(t *testing.T) {
	scope := testScope("1PANEL_BASIC_BEFORE")
	reader := &fakeRuleReader{output: `-A 1PANEL_BASIC_BEFORE -m limit --limit 5/min -j ACCEPT
-A 1PANEL_BASIC_BEFORE -p tcp -m multiport --dports 80,443 -j ACCEPT
`}
	snapshot, err := NewAdapterWithReader(reader).Observe(context.Background(), scope)
	if err != nil {
		t.Fatalf("observe opaque rules: %v", err)
	}
	if len(snapshot.Rules) != 2 || snapshot.Rules[0].ParseStatus != filter.ParseStatusOpaque || snapshot.Rules[1].ParseStatus != filter.ParseStatusOpaque {
		t.Fatalf("unsupported rules were guessed: %#v", snapshot.Rules)
	}
	if snapshot.Rules[0].Raw == "" || snapshot.Rules[0].Locator.Canonical == "" {
		t.Fatalf("opaque diagnostics were not retained: %#v", snapshot.Rules[0])
	}
}

func TestObserveAcceptsDefaultIPv6RejectRepresentation(t *testing.T) {
	scope := testScopeFamily("1PANEL_BASIC", filter.FamilyIPv6)
	reader := &fakeRuleReader{output: "-A 1PANEL_BASIC -s 2001:db8::/64 -j REJECT --reject-with icmp6-port-unreachable"}
	snapshot, err := NewAdapterWithReader(reader).Observe(context.Background(), scope)
	if err != nil {
		t.Fatalf("observe IPv6 reject: %v", err)
	}
	if len(snapshot.Rules) != 1 || snapshot.Rules[0].ParseStatus != filter.ParseStatusSupported ||
		snapshot.Rules[0].Rule.Action != filter.ActionReject {
		t.Fatalf("default IPv6 reject became opaque: %#v", snapshot.Rules)
	}
	reader.output = "-A 1PANEL_BASIC -s 2001:db8::/64 -j REJECT --reject-with icmp6-adm-prohibited"
	snapshot, err = NewAdapterWithReader(reader).Observe(context.Background(), scope)
	if err != nil || len(snapshot.Rules) != 1 || snapshot.Rules[0].ParseStatus != filter.ParseStatusOpaque {
		t.Fatalf("non-default IPv6 reject semantics were guessed: snapshot=%#v err=%v", snapshot, err)
	}
}

func TestObserveNormalizesConnectionStateSafetyRule(t *testing.T) {
	scope := testScope("1PANEL_BASIC_BEFORE")
	reader := &fakeRuleReader{output: `-A 1PANEL_BASIC_BEFORE -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT -m comment --comment "ESTABLISHED Whitelist"`}
	snapshot, err := NewAdapterWithReader(reader).Observe(context.Background(), scope)
	if err != nil {
		t.Fatalf("observe state rule: %v", err)
	}
	states := snapshot.Rules[0].Rule.ConnectionStates
	if len(states) != 2 || states[0] != "established" || states[1] != "related" {
		t.Fatalf("connection states were not normalized: %#v", states)
	}
	if !snapshot.Rules[0].Protected {
		t.Fatalf("established whitelist was not protected: %#v", snapshot.Rules[0])
	}
}

func TestObserveProtectsSystemPresetChains(t *testing.T) {
	before := testScope("1PANEL_BASIC_BEFORE")
	beforeSnapshot, err := NewAdapterWithReader(&fakeRuleReader{output: `-A 1PANEL_BASIC_BEFORE -p tcp --dport 8080 -j ACCEPT`}).Observe(context.Background(), before)
	if err != nil || len(beforeSnapshot.Rules) != 1 || !beforeSnapshot.Rules[0].Protected {
		t.Fatalf("BEFORE preset rule was not protected: snapshot=%#v err=%v", beforeSnapshot, err)
	}
	after := testScope("1PANEL_BASIC_AFTER")
	afterSnapshot, err := NewAdapterWithReader(&fakeRuleReader{output: `-A 1PANEL_BASIC_AFTER -p tcp --dport 8080 -j ACCEPT`}).Observe(context.Background(), after)
	if err != nil || len(afterSnapshot.Rules) != 1 || !afterSnapshot.Rules[0].Protected {
		t.Fatalf("AFTER preset rule was not protected: snapshot=%#v err=%v", afterSnapshot, err)
	}
}

func TestObserveRejectsScopeOutsideOwnedChains(t *testing.T) {
	scope := testScope("INPUT")
	_, err := NewAdapterWithReader(&fakeRuleReader{}).Observe(context.Background(), scope)
	if err == nil {
		t.Fatal("expected INPUT scope to be rejected")
	}
}

type fakeRuleReader struct {
	output string
	err    error
}

type fakeRuleWriter struct {
	commands   []filter.NativeCommand
	saved      bool
	saveCalls  int
	saveErrors []error
	runErrors  []error
}

func (f *fakeRuleWriter) Run(_ context.Context, command filter.NativeCommand) error {
	f.commands = append(f.commands, command)
	if len(f.runErrors) >= len(f.commands) {
		return f.runErrors[len(f.commands)-1]
	}
	return nil
}

func (f *fakeRuleWriter) Save(context.Context, filter.Scope) error {
	f.saveCalls++
	f.saved = true
	if len(f.saveErrors) >= f.saveCalls {
		return f.saveErrors[f.saveCalls-1]
	}
	return nil
}

func executorObserved(rule filter.FirewallRule, position int) filter.ObservedRule {
	return filter.ObservedRule{
		Rule: rule, ParseStatus: filter.ParseStatusSupported,
		Locator: filter.Locator{Provider: filter.ProviderIptables, ScopeKey: rule.Scope.Key(), Position: &position},
	}
}

func (f *fakeRuleReader) ListChain(context.Context, filter.Scope) (string, error) {
	return f.output, f.err
}

func testScope(chain string) filter.Scope {
	return testScopeFamily(chain, filter.FamilyIPv4)
}

func testScopeFamily(chain string, family filter.Family) filter.Scope {
	return filter.Scope{
		Provider: filter.ProviderIptables, Family: family, Table: "filter", Chain: chain, Direction: filter.DirectionInput,
	}
}
