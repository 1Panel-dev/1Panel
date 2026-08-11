package ufw

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

const numberedFixture = `Status: active

     To                         Action      From
     --                         ------      ----
[ 1] 22/tcp                    ALLOW IN    Anywhere                   # 1panel-rule:rule-v4
[ 2] Anywhere                  DENY IN     172.16.10.111
[ 3] OpenSSH                   ALLOW IN    Anywhere
[ 4] 443/tcp (v6)              ALLOW IN    Anywhere (v6)              # web v6
[ 5] Anywhere (v6) on eth0     REJECT IN   2001:db8::/64 (v6)
[ 6] Anywhere                  ALLOW FWD   Anywhere on lxdbr0
[ 7] 53                        ALLOW IN    Anywhere
[ 8] 6000:6010/udp on eth1    ALLOW IN    10.0.0.0/8
[ 9] 22,80,443/tcp             ALLOW IN    Anywhere
[10] 2222/tcp                  LIMIT IN    Anywhere                   # SSH limit
[11] 25/tcp                    DENY IN     2001:db8::/32
[12] a-very-long-application-profile-name ALLOW IN    Anywhere
[13] Anywhere on eth0          ALLOW IN    Anywhere                   (log)
[14] 443/tcp                   ALLOW OUT   Anywhere on eth0           (out)
`

type fakeReader struct {
	outputs map[string]string
	errors  map[string]error
	calls   [][]string
}

type scriptedBackend struct {
	numbered []string
	writes   []filter.NativeCommand
	failAt   int
}

func (b *scriptedBackend) Read(_ context.Context, args ...string) (string, error) {
	if stringsKey(args) == "status verbose" {
		return "Status: active\nDefault: deny (incoming), allow (outgoing), deny (routed)\n", nil
	}
	if stringsKey(args) != "status numbered" || len(b.numbered) == 0 {
		return "", fmt.Errorf("unexpected read: %v", args)
	}
	output := b.numbered[0]
	b.numbered = b.numbered[1:]
	return output, nil
}

func (b *scriptedBackend) Run(_ context.Context, command filter.NativeCommand) error {
	b.writes = append(b.writes, command)
	if b.failAt > 0 && len(b.writes) == b.failAt {
		return errors.New("write failed")
	}
	return nil
}

func (f *fakeReader) Read(_ context.Context, args ...string) (string, error) {
	f.calls = append(f.calls, append([]string(nil), args...))
	key := stringsKey(args)
	return f.outputs[key], f.errors[key]
}

func TestObserveNumberedIncomingIPv4PreservesGlobalPositions(t *testing.T) {
	reader := &fakeReader{outputs: map[string]string{
		"status numbered": numberedFixture,
		"status verbose":  "Status: active\nDefault: deny (incoming), allow (outgoing), deny (routed)\n",
	}}
	adapter := NewAdapterWithReader(reader)
	snapshot, err := adapter.Observe(context.Background(), ufwScope(filter.FamilyIPv4))
	if err != nil {
		t.Fatalf("observe: %v", err)
	}
	if got, want := len(snapshot.Rules), 9; got != want {
		t.Fatalf("expected %d IPv4 rules, got %d", want, got)
	}
	positions := make([]int, 0, len(snapshot.Rules))
	for _, rule := range snapshot.Rules {
		positions = append(positions, *rule.Locator.Position)
	}
	if !reflect.DeepEqual(positions, []int{1, 2, 3, 7, 8, 9, 10, 12, 13}) {
		t.Fatalf("unexpected positions: %v", positions)
	}
	first := snapshot.Rules[0]
	if first.ParseStatus != filter.ParseStatusSupported || first.Marker != "1panel-rule:rule-v4" ||
		first.Rule.DestinationPort != "22" || first.Rule.Protocol != "tcp" || first.Rule.Action != filter.ActionAccept {
		t.Fatalf("unexpected first rule: %#v", first)
	}
	second := snapshot.Rules[1]
	if second.ParseStatus != filter.ParseStatusSupported || second.Rule.SourceAddress != "172.16.10.111/32" || second.Rule.Action != filter.ActionDrop {
		t.Fatalf("unexpected source deny: %#v", second)
	}
	eighth := snapshot.Rules[4]
	if eighth.ParseStatus != filter.ParseStatusSupported || eighth.Rule.DestinationPort != "6000-6010" || eighth.Rule.Interface != "eth1" {
		t.Fatalf("unexpected range rule: %#v", eighth)
	}
	for _, index := range []int{2, 3, 5, 6, 7, 8} {
		if snapshot.Rules[index].ParseStatus != filter.ParseStatusOpaque {
			t.Fatalf("expected rule at slice index %d to be opaque: %#v", index, snapshot.Rules[index])
		}
	}
	if !hasNotice(snapshot.Notices, filter.ScopeNoticeDefaultPolicy) {
		t.Fatalf("expected default policy notice: %#v", snapshot.Notices)
	}
	if !reflect.DeepEqual(reader.calls, [][]string{{"status", "numbered"}, {"status", "verbose"}}) {
		t.Fatalf("unexpected commands: %#v", reader.calls)
	}
}

func TestObserveNumberedIncomingIPv6KeepsFamilyGap(t *testing.T) {
	reader := &fakeReader{outputs: map[string]string{
		"status numbered": numberedFixture,
		"status verbose":  "Status: active\nDefault: deny (incoming), allow (outgoing), deny (routed)\n",
	}}
	snapshot, err := NewAdapterWithReader(reader).Observe(context.Background(), ufwScope(filter.FamilyIPv6))
	if err != nil {
		t.Fatalf("observe: %v", err)
	}
	if len(snapshot.Rules) != 3 || *snapshot.Rules[0].Locator.Position != 4 || *snapshot.Rules[1].Locator.Position != 5 ||
		*snapshot.Rules[2].Locator.Position != 11 {
		t.Fatalf("unexpected IPv6 rules: %#v", snapshot.Rules)
	}
	first := snapshot.Rules[0]
	if first.ParseStatus != filter.ParseStatusSupported || first.Rule.Description != "web v6" || first.Rule.DestinationPort != "443" {
		t.Fatalf("unexpected IPv6 port rule: %#v", first)
	}
	second := snapshot.Rules[1]
	if second.ParseStatus != filter.ParseStatusSupported || second.Rule.Interface != "eth0" ||
		second.Rule.SourceAddress != "2001:db8::/64" || second.Rule.Action != filter.ActionReject {
		t.Fatalf("unexpected IPv6 reject: %#v", second)
	}
	third := snapshot.Rules[2]
	if third.ParseStatus != filter.ParseStatusSupported || third.Rule.SourceAddress != "2001:db8::/32" || third.Rule.DestinationPort != "25" {
		t.Fatalf("unexpected explicit IPv6 rule without family suffix: %#v", third)
	}
}

func TestObserveInactiveUFWReturnsNoticeAndEmptyInventory(t *testing.T) {
	reader := &fakeReader{outputs: map[string]string{
		"status numbered": "Status: inactive\n",
		"status verbose":  "Status: inactive\n",
	}}
	snapshot, err := NewAdapterWithReader(reader).Observe(context.Background(), ufwScope(filter.FamilyIPv4))
	if err != nil {
		t.Fatalf("observe: %v", err)
	}
	if len(snapshot.Rules) != 0 || !hasNotice(snapshot.Notices, filter.ScopeNoticeManagedScopeInactive) {
		t.Fatalf("unexpected inactive snapshot: %#v", snapshot)
	}
}

func TestObserveRejectsScopeOutsideIncoming(t *testing.T) {
	adapter := NewAdapterWithReader(&fakeReader{})
	_, err := adapter.Observe(context.Background(), filter.Scope{
		Provider: filter.ProviderUFW, Family: filter.FamilyIPv4, Chain: "outgoing", Direction: filter.Direction("output"),
	})
	if !errors.Is(err, filter.ErrInvalidScope) {
		t.Fatalf("expected invalid output scope, got %v", err)
	}
}

func TestCapabilitiesOnlyAdvertiseAtomicIncomingFamilies(t *testing.T) {
	capabilities, err := NewAdapterWithReader(&fakeReader{}).Capabilities(context.Background())
	if err != nil {
		t.Fatalf("capabilities: %v", err)
	}
	if !capabilities.Marker || !capabilities.SupportsScope(ufwScope(filter.FamilyIPv4)) ||
		!capabilities.SupportsScope(ufwScope(filter.FamilyIPv6)) ||
		!capabilities.ExplicitPosition ||
		capabilities.SupportsScope(filter.Scope{Provider: filter.ProviderUFW, Family: filter.FamilyIPv4, Chain: "outgoing", Direction: filter.Direction("output")}) {
		t.Fatalf("unexpected capabilities: %#v", capabilities)
	}
}

func TestCompileCreateUsesFamilyExplicitFullSyntax(t *testing.T) {
	tests := []struct {
		name    string
		family  filter.Family
		address string
	}{
		{name: "IPv4", family: filter.FamilyIPv4, address: "0.0.0.0/0"},
		{name: "IPv6", family: filter.FamilyIPv6, address: "::/0"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := mustSnapshot(t, ufwScope(test.family), nil)
			rule := writableRule(test.family, "new-rule", "8080")
			plan, err := NewAdapterWithReader(&fakeReader{}).Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
			if err != nil {
				t.Fatalf("compile: %v", err)
			}
			want := []string{"insert", "1", "allow", "in", "proto", "tcp", "from", test.address, "to", test.address, "port", "8080", "comment", "1panel-rule:new-rule"}
			if !reflect.DeepEqual(plan.Rules[0].Commands[0].Args, want) {
				t.Fatalf("unexpected command:\nwant: %#v\n got: %#v", want, plan.Rules[0].Commands[0].Args)
			}
			if got := plan.Rules[0].RollbackCommands[0].Args[:3]; !reflect.DeepEqual(got, []string{"--force", "delete", "allow"}) {
				t.Fatalf("unexpected rollback prefix: %#v", got)
			}
		})
	}
}

func TestCompileAdoptUpdatesCommentWithoutChangingNumber(t *testing.T) {
	scope := ufwScope(filter.FamilyIPv4)
	observed := parseNumberedRules(scope, "Status: active\n[ 7] Anywhere DENY IN 172.16.10.111 # imported\n")
	if len(observed) != 1 {
		t.Fatalf("expected one observed rule: %#v", observed)
	}
	snapshot := mustSnapshot(t, scope, observed)
	rule := observed[0].Rule
	rule.UUID = "adopted-rule"
	locator := observed[0].Locator
	plan, err := NewAdapterWithReader(&fakeReader{}).Compile(snapshot, []filter.DesiredChange{{
		Operation: filter.ChangeAdopt, After: &rule, Locator: &locator,
	}})
	if err != nil {
		t.Fatalf("compile adoption: %v", err)
	}
	rulePlan := plan.Rules[0]
	if len(rulePlan.Commands) != 1 || rulePlan.Commands[0].Args[0] != "deny" ||
		rulePlan.Commands[0].Args[len(rulePlan.Commands[0].Args)-1] != "1panel-rule:adopted-rule" {
		t.Fatalf("unexpected adoption command: %#v", rulePlan.Commands)
	}
	if rulePlan.RollbackCommands[0].Args[len(rulePlan.RollbackCommands[0].Args)-1] != "imported" ||
		rulePlan.Expected.Locator.Position == nil || *rulePlan.Expected.Locator.Position != 7 {
		t.Fatalf("adoption did not preserve comment and position: %#v", rulePlan)
	}
}

func TestCompileUpdateAndDeleteRequireOwnedNumberedRule(t *testing.T) {
	scope := ufwScope(filter.FamilyIPv4)
	observed := parseNumberedRules(scope, "Status: active\n[ 3] 80/tcp ALLOW IN Anywhere # 1panel-rule:managed\n")
	snapshot := mustSnapshot(t, scope, observed)
	locator := observed[0].Locator
	updated := writableRule(filter.FamilyIPv4, "managed", "443")
	update, err := NewAdapterWithReader(&fakeReader{}).Compile(snapshot, []filter.DesiredChange{{
		Operation: filter.ChangeUpdate, After: &updated, Locator: &locator,
	}})
	if err != nil {
		t.Fatalf("compile update: %v", err)
	}
	if got := update.Rules[0].Commands; len(got) != 2 || !reflect.DeepEqual(got[0].Args, []string{"--force", "delete", "3"}) ||
		!reflect.DeepEqual(got[1].Args[:2], []string{"insert", "3"}) {
		t.Fatalf("unexpected update commands: %#v", got)
	}
	before := observed[0].Rule
	before.UUID = "managed"
	deletePlan, err := NewAdapterWithReader(&fakeReader{}).Compile(snapshot, []filter.DesiredChange{{
		Operation: filter.ChangeDelete, Before: &before, Locator: &locator,
	}})
	if err != nil {
		t.Fatalf("compile delete: %v", err)
	}
	if !reflect.DeepEqual(deletePlan.Rules[0].Commands[0].Args, []string{"--force", "delete", "3"}) {
		t.Fatalf("unexpected delete command: %#v", deletePlan.Rules[0].Commands)
	}

	external := observed[0]
	external.Marker = ""
	external.Rule.Description = "external"
	externalSnapshot := mustSnapshot(t, scope, []filter.ObservedRule{external})
	_, err = NewAdapterWithReader(&fakeReader{}).Compile(externalSnapshot, []filter.DesiredChange{{
		Operation: filter.ChangeDelete, Before: &before, Locator: &locator,
	}})
	if err == nil {
		t.Fatal("expected deletion of an external rule to be rejected")
	}
}

func TestCompileUpdateUsesRequestedGlobalPosition(t *testing.T) {
	scope := ufwScope(filter.FamilyIPv4)
	observed := parseNumberedRules(scope, "Status: active\n[ 3] 80/tcp ALLOW IN Anywhere # 1panel-rule:managed\n")
	snapshot := mustSnapshot(t, scope, observed)
	locator := observed[0].Locator
	updated := writableRule(filter.FamilyIPv4, "managed", "443")
	target := int64(1)
	updated.OrderIndex = &target
	plan, err := NewAdapterWithReader(&fakeReader{}).Compile(snapshot, []filter.DesiredChange{{
		Operation: filter.ChangeUpdate, Before: &observed[0].Rule, After: &updated, Locator: &locator,
	}})
	if err != nil {
		t.Fatalf("compile positioned update: %v", err)
	}
	rulePlan := plan.Rules[0]
	if len(rulePlan.Commands) != 2 || !reflect.DeepEqual(rulePlan.Commands[0].Args, []string{"--force", "delete", "3"}) ||
		!reflect.DeepEqual(rulePlan.Commands[1].Args[:2], []string{"insert", "1"}) ||
		rulePlan.Expected.Locator.Position == nil || *rulePlan.Expected.Locator.Position != 1 {
		t.Fatalf("unexpected positioned update: %#v", rulePlan)
	}
}

func TestCompileRejectsInactiveAndProtected(t *testing.T) {
	scope := ufwScope(filter.FamilyIPv4)
	rule := writableRule(filter.FamilyIPv4, "rule", "8080")
	inactive := mustSnapshot(t, scope, nil)
	inactive.Notices = []filter.ScopeNotice{{Code: filter.ScopeNoticeManagedScopeInactive}}
	_, err := NewAdapterWithReader(&fakeReader{}).Compile(inactive, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
	if !errors.Is(err, filter.ErrProviderUnavailable) {
		t.Fatalf("expected inactive provider error, got %v", err)
	}

	observed := parseNumberedRules(scope, "Status: active\n[ 1] 8080/tcp ALLOW IN Anywhere\n")
	observed[0].Protected = true
	protected := mustSnapshot(t, scope, observed)
	locator := observed[0].Locator
	rule.UUID = "protected"
	_, err = NewAdapterWithReader(&fakeReader{}).Compile(protected, []filter.DesiredChange{{Operation: filter.ChangeAdopt, After: &rule, Locator: &locator}})
	if !errors.Is(err, filter.ErrProtectedRule) {
		t.Fatalf("expected protected rule error, got %v", err)
	}
	_, err = NewAdapterWithReader(&fakeReader{}).Compile(mustSnapshot(t, scope, nil), []filter.DesiredChange{{Operation: filter.ChangeReorder, After: &rule}})
	if !errors.Is(err, filter.ErrUnsupportedScope) {
		t.Fatalf("expected unsupported standalone reorder error, got %v", err)
	}
	broadDeny := filter.FirewallRule{
		UUID: "deny-all", Scope: scope, NativeKind: filter.NativeKindUFWRule, Protocol: "all", Action: filter.ActionDrop,
	}
	_, err = NewAdapterWithReader(&fakeReader{}).Compile(mustSnapshot(t, scope, nil), []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &broadDeny}})
	if !errors.Is(err, filter.ErrLockoutRisk) {
		t.Fatalf("expected broad deny lockout error, got %v", err)
	}
}

func TestApplyVerifiesMarkerAcrossBothFamilies(t *testing.T) {
	scope := ufwScope(filter.FamilyIPv4)
	snapshot := mustSnapshot(t, scope, nil)
	rule := writableRule(filter.FamilyIPv4, "created", "8080")
	backend := &scriptedBackend{numbered: []string{
		"Status: active\n[ 1] 8080/tcp ALLOW IN Anywhere # 1panel-rule:created\n",
		"Status: active\n",
	}}
	adapter := NewAdapterWithBackend(backend, backend)
	plan, err := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if _, err = adapter.Apply(context.Background(), plan); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(backend.writes) != 1 {
		t.Fatalf("unexpected writes: %#v", backend.writes)
	}
}

func TestApplyCompensatesFamilyExpansion(t *testing.T) {
	scope := ufwScope(filter.FamilyIPv4)
	snapshot := mustSnapshot(t, scope, nil)
	rule := writableRule(filter.FamilyIPv4, "expanded", "8080")
	backend := &scriptedBackend{numbered: []string{
		"Status: active\n[ 1] 8080/tcp ALLOW IN Anywhere # 1panel-rule:expanded\n",
		"Status: active\n[ 2] 8080/tcp (v6) ALLOW IN Anywhere (v6) # 1panel-rule:expanded\n",
	}}
	adapter := NewAdapterWithBackend(backend, backend)
	plan, err := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if _, err = adapter.Apply(context.Background(), plan); err == nil {
		t.Fatal("expected one-to-many verification failure")
	}
	if len(backend.writes) != 2 || backend.writes[1].Args[0] != "--force" || backend.writes[1].Args[1] != "delete" {
		t.Fatalf("expected compensating delete, got %#v", backend.writes)
	}
}

func TestRollbackReversesFullyAppliedUFWPlan(t *testing.T) {
	scope := ufwScope(filter.FamilyIPv4)
	snapshot := mustSnapshot(t, scope, nil)
	rule := writableRule(filter.FamilyIPv4, "rollback", "8080")
	backend := &scriptedBackend{}
	adapter := NewAdapterWithBackend(backend, backend)
	plan, err := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
	if err != nil {
		t.Fatalf("compile rollback plan: %v", err)
	}
	if err := adapter.Rollback(context.Background(), plan); err != nil {
		t.Fatalf("rollback applied plan: %v", err)
	}
	if len(backend.writes) != 1 || !reflect.DeepEqual(backend.writes[0].Args[:2], []string{"--force", "delete"}) {
		t.Fatalf("unexpected rollback writes: %#v", backend.writes)
	}
}

func TestApplyCompensatesFailedUpdate(t *testing.T) {
	scope := ufwScope(filter.FamilyIPv4)
	beforeOutput := "Status: active\n[ 3] 80/tcp ALLOW IN Anywhere # 1panel-rule:managed\n"
	observed := parseNumberedRules(scope, beforeOutput)
	snapshot := mustSnapshot(t, scope, observed)
	locator := observed[0].Locator
	updated := writableRule(filter.FamilyIPv4, "managed", "443")
	backend := &scriptedBackend{numbered: []string{beforeOutput, "Status: active\n"}, failAt: 2}
	adapter := NewAdapterWithBackend(backend, backend)
	plan, err := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeUpdate, After: &updated, Locator: &locator}})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if _, err = adapter.Apply(context.Background(), plan); err == nil || !strings.Contains(err.Error(), "write failed") {
		t.Fatalf("expected write failure, got %v", err)
	}
	if len(backend.writes) != 3 || !reflect.DeepEqual(backend.writes[2].Args[:2], []string{"insert", "3"}) {
		t.Fatalf("expected old rule restore after failed update: %#v", backend.writes)
	}
}

func TestApplyUsesCompiledNumberedSnapshotWithoutPreflight(t *testing.T) {
	scope := ufwScope(filter.FamilyIPv4)
	snapshot := mustSnapshot(t, scope, nil)
	rule := writableRule(filter.FamilyIPv4, "stale", "8080")
	backend := &scriptedBackend{numbered: []string{
		"Status: active\n[ 1] 8080/tcp ALLOW IN Anywhere # 1panel-rule:stale\n",
		"Status: active\n",
	}}
	adapter := NewAdapterWithBackend(backend, backend)
	plan, err := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if _, err = adapter.Apply(context.Background(), plan); err != nil {
		t.Fatalf("apply compiled plan: %v", err)
	}
	if len(backend.writes) != 1 {
		t.Fatalf("compiled plan was not written: %#v", backend.writes)
	}
}

func ufwScope(family filter.Family) filter.Scope {
	return filter.Scope{Provider: filter.ProviderUFW, Family: family, Chain: "incoming", Direction: filter.DirectionInput}
}

func writableRule(family filter.Family, uuid, port string) filter.FirewallRule {
	return filter.FirewallRule{
		UUID: uuid, Scope: ufwScope(family), NativeKind: filter.NativeKindUFWRule,
		Protocol: "tcp", DestinationPort: port, Action: filter.ActionAccept,
	}
}

func mustSnapshot(t *testing.T, scope filter.Scope, rules []filter.ObservedRule) filter.Snapshot {
	t.Helper()
	snapshot, err := filter.NewSnapshot(scope, rules)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	return snapshot
}

func hasNotice(notices []filter.ScopeNotice, code filter.ScopeNoticeCode) bool {
	for _, notice := range notices {
		if notice.Code == code {
			return true
		}
	}
	return false
}

func stringsKey(values []string) string {
	result := ""
	for index, value := range values {
		if index != 0 {
			result += " "
		}
		result += value
	}
	return result
}
