package nftables

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

type fakeBackend struct {
	output   string
	commands []filter.NativeCommand
	saves    int
	failAt   int
}

func (f *fakeBackend) ListChain(context.Context, filter.Scope) (string, error) { return f.output, nil }
func (f *fakeBackend) Run(_ context.Context, command filter.NativeCommand) error {
	f.commands = append(f.commands, command)
	if f.failAt != 0 && len(f.commands) == f.failAt {
		return errors.New("run failed")
	}
	return nil
}
func (f *fakeBackend) Save(context.Context) error { f.saves++; return nil }

func TestObserveNativeNftablesRules(t *testing.T) {
	scope := testScope(filter.FamilyIPv4, "1PANEL_BASIC")
	backend := &fakeBackend{output: `table ip nft_1panel_filter {
 chain NFT_1PANEL_BASIC {
	  meta l4proto tcp ip saddr 10.0.0.0/8 tcp dport 443 accept comment "1panel-rule:web" # handle 12
	  meta l4proto udp udp dport 53 drop # handle 13
	  meta l4proto tcp ct state established,related tcp dport 8443 accept comment "1panel-rule:stateful" # handle 14
 }
}`}
	snapshot, err := NewAdapterWithBackend(backend).Observe(context.Background(), scope)
	if err != nil {
		t.Fatalf("observe: %v", err)
	}
	if len(snapshot.Rules) != 3 || snapshot.Rules[0].Marker != "1panel-rule:web" || snapshot.Rules[0].Locator.NativeID != "12" {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
	if snapshot.Rules[0].Rule.SourceAddress != "10.0.0.0/8" || snapshot.Rules[0].Rule.DestinationPort != "443" || snapshot.Rules[1].Rule.Action != filter.ActionDrop {
		t.Fatalf("unexpected parsed rules: %#v", snapshot.Rules)
	}
	if len(snapshot.Rules[2].Rule.ConnectionStates) != 2 || snapshot.Rules[2].Rule.ConnectionStates[0] != "established" {
		t.Fatalf("connection states were not parsed: %#v", snapshot.Rules[2])
	}
}

func TestApplyCompensatesFailedRulesetTransaction(t *testing.T) {
	scope := testScope(filter.FamilyIPv4, "1PANEL_BASIC")
	snapshot, err := filter.NewSnapshot(scope, nil)
	if err != nil {
		t.Fatal(err)
	}
	rule := filter.FirewallRule{UUID: "web", Scope: scope, Protocol: "tcp", DestinationPort: "443", Action: filter.ActionAccept}
	backend := &fakeBackend{failAt: 1}
	adapter := NewAdapterWithBackend(backend)
	plan, err := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if _, err := adapter.Apply(context.Background(), plan); err == nil {
		t.Fatal("expected ruleset transaction failure")
	}
	if len(backend.commands) != 2 {
		t.Fatalf("expected failed transaction and rollback transaction; got %#v", backend.commands)
	}
	if backend.saves != 1 {
		t.Fatalf("rollback was not persisted: saves=%d", backend.saves)
	}
}

func TestCompileApplyAndRollbackRebuildOwnedChain(t *testing.T) {
	scope := testScope(filter.FamilyIPv6, "1PANEL_BASIC")
	snapshot, err := filter.NewSnapshot(scope, nil)
	if err != nil {
		t.Fatal(err)
	}
	rule := filter.FirewallRule{UUID: "dns6", Scope: scope, Protocol: "udp", DestinationPort: "53", Action: filter.ActionAccept}
	backend := &fakeBackend{}
	adapter := NewAdapterWithBackend(backend)
	plan, err := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	commands := plan.Rules[0].Commands
	if len(commands) != 1 || commands[0].Executable != "nft" || strings.Join(commands[0].Args, " ") != "-f -" {
		t.Fatalf("unexpected commands: %#v", commands)
	}
	joined := commands[0].Stdin
	for _, want := range []string{"add rule ip6 nft_1panel_filter NFT_1PANEL_BASIC", "udp dport 53", `comment "1panel-rule:dns6"`} {
		if !strings.Contains(joined, want) {
			t.Fatalf("command %q does not contain %q", joined, want)
		}
	}
	if _, err := adapter.Apply(context.Background(), plan); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(backend.commands) != 1 || backend.saves != 1 {
		t.Fatalf("unexpected apply calls: commands=%#v saves=%d", backend.commands, backend.saves)
	}
	if err := adapter.Rollback(context.Background(), plan); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if len(backend.commands) != 2 || backend.commands[1].Stdin != "flush chain ip6 nft_1panel_filter NFT_1PANEL_BASIC\n" || backend.saves != 2 {
		t.Fatalf("unexpected rollback calls: commands=%#v saves=%d", backend.commands, backend.saves)
	}
}

func TestCompileAndApplyBatchCreateUsesOneRulesetTransaction(t *testing.T) {
	scope := testScope(filter.FamilyIPv4, "1PANEL_BASIC")
	snapshot, err := filter.NewSnapshot(scope, nil)
	if err != nil {
		t.Fatal(err)
	}
	web := filter.FirewallRule{UUID: "web", Scope: scope, Protocol: "tcp", DestinationPort: "80", Action: filter.ActionAccept}
	tls := filter.FirewallRule{UUID: "tls", Scope: scope, Protocol: "tcp", DestinationPort: "443", Action: filter.ActionAccept}
	backend := &fakeBackend{}
	adapter := NewAdapterWithBackend(backend)
	plan, err := adapter.Compile(snapshot, []filter.DesiredChange{
		{Operation: filter.ChangeCreate, After: &web},
		{Operation: filter.ChangeCreate, After: &tls},
	})
	if err != nil {
		t.Fatalf("compile batch create: %v", err)
	}
	if len(plan.Rules) != 2 || len(plan.Rules[0].Commands) != 1 || len(plan.Rules[1].Commands) != 0 {
		t.Fatalf("batch was not compiled into one transaction: %#v", plan.Rules)
	}
	script := plan.Rules[0].Commands[0].Stdin
	if strings.Count(script, "flush chain ") != 1 || strings.Count(script, "add rule ") != 2 ||
		!strings.Contains(script, "1panel-rule:web") || !strings.Contains(script, "1panel-rule:tls") {
		t.Fatalf("unexpected batch ruleset:\n%s", script)
	}
	result, err := adapter.Apply(context.Background(), plan)
	if err != nil {
		t.Fatalf("apply batch create: %v", err)
	}
	if len(result.Applied) != 2 || len(backend.commands) != 1 || backend.saves != 1 {
		t.Fatalf("batch create was not applied once: result=%#v commands=%#v saves=%d", result, backend.commands, backend.saves)
	}
}

func TestCompileBatchDeletePreservesExternalRules(t *testing.T) {
	scope := testScope(filter.FamilyIPv4, "1PANEL_BASIC")
	positionOne, positionTwo, positionThree := 1, 2, 3
	web := filter.FirewallRule{UUID: "web", Scope: scope, Protocol: "tcp", DestinationPort: "80", Action: filter.ActionAccept}
	tls := filter.FirewallRule{UUID: "tls", Scope: scope, Protocol: "tcp", DestinationPort: "443", Action: filter.ActionAccept}
	external := filter.ObservedRule{
		Rule: filter.FirewallRule{Scope: scope, NativeKind: filter.NativeKindOpaque}, Raw: "fib saddr . iif oif missing drop",
		ParseStatus: filter.ParseStatusOpaque, Locator: filter.Locator{Provider: filter.ProviderNftables, ScopeKey: scope.Key(), Position: &positionOne},
	}
	webObserved := observedRule(web, "1panel-rule:web", positionTwo, "")
	tlsObserved := observedRule(tls, "1panel-rule:tls", positionThree, "")
	snapshot, err := filter.NewSnapshot(scope, []filter.ObservedRule{external, webObserved, tlsObserved})
	if err != nil {
		t.Fatal(err)
	}
	adapter := NewAdapterWithBackend(&fakeBackend{})
	plan, err := adapter.Compile(snapshot, []filter.DesiredChange{
		{Operation: filter.ChangeDelete, Before: &tls, Locator: &tlsObserved.Locator},
		{Operation: filter.ChangeDelete, Before: &web, Locator: &webObserved.Locator},
	})
	if err != nil {
		t.Fatalf("compile batch delete: %v", err)
	}
	applyScript := plan.Rules[0].Commands[0].Stdin
	if strings.Count(applyScript, "flush chain ") != 1 || strings.Count(applyScript, "add rule ") != 1 ||
		!strings.Contains(applyScript, external.Raw) || strings.Contains(applyScript, "1panel-rule:web") || strings.Contains(applyScript, "1panel-rule:tls") {
		t.Fatalf("unexpected batch delete ruleset:\n%s", applyScript)
	}
	rollbackScript := plan.Rules[0].RollbackCommands[0].Stdin
	if !strings.Contains(rollbackScript, "1panel-rule:web") || !strings.Contains(rollbackScript, "1panel-rule:tls") ||
		!strings.Contains(rollbackScript, external.Raw) {
		t.Fatalf("rollback does not restore the original ruleset:\n%s", rollbackScript)
	}
}

func TestVerifyBatchCreateChecksEveryRule(t *testing.T) {
	scope := testScope(filter.FamilyIPv4, "1PANEL_BASIC")
	snapshot, err := filter.NewSnapshot(scope, nil)
	if err != nil {
		t.Fatal(err)
	}
	web := filter.FirewallRule{UUID: "web", Scope: scope, Protocol: "tcp", DestinationPort: "80", Action: filter.ActionAccept}
	tls := filter.FirewallRule{UUID: "tls", Scope: scope, Protocol: "tcp", DestinationPort: "443", Action: filter.ActionAccept}
	backend := &fakeBackend{output: strings.Join([]string{
		`meta l4proto tcp tcp dport 80 accept comment "1panel-rule:web" # handle 10`,
		`meta l4proto tcp tcp dport 443 accept comment "1panel-rule:tls" # handle 11`,
	}, "\n")}
	adapter := NewAdapterWithBackend(backend)
	plan, err := adapter.Compile(snapshot, []filter.DesiredChange{
		{Operation: filter.ChangeCreate, After: &web},
		{Operation: filter.ChangeCreate, After: &tls},
	})
	if err != nil {
		t.Fatalf("compile batch create: %v", err)
	}
	verified, err := adapter.Verify(context.Background(), plan)
	if err != nil || !verified.Matched {
		t.Fatalf("verify complete batch: result=%#v err=%v", verified, err)
	}
	backend.output = `meta l4proto tcp tcp dport 80 accept comment "1panel-rule:web" # handle 10`
	verified, err = adapter.Verify(context.Background(), plan)
	if err != nil || verified.Matched {
		t.Fatalf("missing batch member passed verification: result=%#v err=%v", verified, err)
	}
}

func TestParseOpaqueRulePreservesRawExpression(t *testing.T) {
	scope := testScope(filter.FamilyIPv4, "1PANEL_BASIC")
	backend := &fakeBackend{output: "fib saddr . iif oif missing drop # handle 9\n"}
	snapshot, err := NewAdapterWithBackend(backend).Observe(context.Background(), scope)
	if err != nil || len(snapshot.Rules) != 1 {
		t.Fatalf("observe opaque: snapshot=%#v err=%v", snapshot, err)
	}
	if snapshot.Rules[0].ParseStatus != filter.ParseStatusOpaque || snapshot.Rules[0].Raw == "" {
		t.Fatalf("opaque rule was not preserved: %#v", snapshot.Rules[0])
	}
}

func testScope(family filter.Family, chain string) filter.Scope {
	return filter.Scope{Provider: filter.ProviderNftables, Family: family, Table: "filter", Chain: chain, Direction: filter.DirectionInput}
}
