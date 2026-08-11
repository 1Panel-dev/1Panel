package firewalld

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

func TestPrepareRuleChoosesNativePortOrRichRule(t *testing.T) {
	adapter := NewAdapterWithReader(newFakeCommandReader())
	zonePort, err := adapter.PrepareRule(filter.FirewallRule{
		Scope: testScope(filter.FamilyInet), Protocol: "tcp", DestinationPort: "443", Action: filter.ActionAccept,
	})
	if err != nil {
		t.Fatalf("prepare native port: %v", err)
	}
	if zonePort.NativeKind != filter.NativeKindZonePort || zonePort.OrderBucket != filter.OrderBucketZonePrimitiveAllow || zonePort.Priority != nil {
		t.Fatalf("simple allow did not remain a native port: %#v", zonePort)
	}
	priority := -100
	rich, err := adapter.PrepareRule(filter.FirewallRule{
		Scope: testScope(filter.FamilyIPv4), Protocol: "tcp", SourceAddress: "172.16.10.111", DestinationPort: "443",
		Action: filter.ActionDrop, Priority: &priority,
	})
	if err != nil {
		t.Fatalf("prepare rich rule: %v", err)
	}
	if rich.NativeKind != filter.NativeKindRichRule || rich.OrderBucket != filter.OrderBucketRichPre || rich.SourceAddress != "172.16.10.111/32" {
		t.Fatalf("address deny did not become a rich rule: %#v", rich)
	}
}

func TestCompileCreateUsesExplicitPublicRuntimeAndPermanentCommands(t *testing.T) {
	adapter := NewAdapterWithReader(newFakeCommandReader())
	snapshot, _ := filter.NewSnapshot(testScope(filter.FamilyInet), nil)
	rule := filter.FirewallRule{
		UUID: "https", Scope: testScope(filter.FamilyInet), Protocol: "tcp", DestinationPort: "443", Action: filter.ActionAccept,
	}
	plan, err := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
	if err != nil {
		t.Fatalf("compile native port: %v", err)
	}
	want := [][]string{
		{"--zone=public", "--add-port=443/tcp"},
		{"--permanent", "--zone=public", "--add-port=443/tcp"},
	}
	if len(plan.Rules) != 1 || len(plan.Rules[0].Commands) != 2 ||
		!reflect.DeepEqual(plan.Rules[0].Commands[0].Args, want[0]) || !reflect.DeepEqual(plan.Rules[0].Commands[1].Args, want[1]) {
		t.Fatalf("unexpected native port plan: %#v", plan)
	}
	if plan.Rules[0].Expected.Rule.NativeKind != filter.NativeKindZonePort || plan.Rules[0].Expected.Marker != "" {
		t.Fatalf("firewalld plan invented marker or representation: %#v", plan.Rules[0].Expected)
	}

	richSnapshot, _ := filter.NewSnapshot(testScope(filter.FamilyIPv4), nil)
	priority := -100
	rich := filter.FirewallRule{
		UUID: "blocked-ip", Scope: testScope(filter.FamilyIPv4), Protocol: "tcp", SourceAddress: "172.16.10.111",
		DestinationPort: "3306", Action: filter.ActionDrop, Priority: &priority,
	}
	plan, err = adapter.Compile(richSnapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rich}})
	if err != nil {
		t.Fatalf("compile rich rule: %v", err)
	}
	option := `--add-rich-rule=rule family="ipv4" priority="-100" source address="172.16.10.111/32" port port="3306" protocol="tcp" drop`
	if plan.Rules[0].Commands[0].Args[1] != option || plan.Rules[0].Expected.Rule.NativeKind != filter.NativeKindRichRule {
		t.Fatalf("unexpected rich rule plan: %#v", plan.Rules[0])
	}
}

func TestCompileAdoptsCanonicalExternalRuleWithoutSystemMutation(t *testing.T) {
	reader := newFakeCommandReader()
	reader.set(false, "--list-ports", "8080/tcp\n")
	reader.set(true, "--list-ports", "8080/tcp\n")
	reader.outputs["--get-default-zone"] = "public\n"
	reader.outputs["--get-active-zones"] = "public\n  interfaces: eth0\n"
	adapter := NewAdapterWithReader(reader)
	snapshot, err := adapter.Observe(context.Background(), testScope(filter.FamilyInet))
	if err != nil {
		t.Fatalf("observe external port: %v", err)
	}
	rule := snapshot.Rules[0].Rule
	rule.UUID = "adopted"
	locator := snapshot.Rules[0].Locator
	plan, err := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeAdopt, After: &rule, Locator: &locator}})
	if err != nil {
		t.Fatalf("compile adoption: %v", err)
	}
	if len(plan.Rules[0].Commands) != 0 || plan.Rules[0].Previous == nil || plan.Rules[0].Expected.Locator.Canonical != "port:8080/tcp" {
		t.Fatalf("canonical adoption mutated firewalld: %#v", plan.Rules[0])
	}
	if _, err := adapter.Apply(context.Background(), plan); err != nil {
		t.Fatalf("no-op canonical adoption required a writer: %v", err)
	}
}

func TestCompileUpdateAndDeleteValidateManagedCanonicalTarget(t *testing.T) {
	reader := newFakeCommandReader()
	reader.set(false, "--list-ports", "8080/tcp\n")
	reader.set(true, "--list-ports", "8080/tcp\n")
	reader.outputs["--get-default-zone"] = "public\n"
	reader.outputs["--get-active-zones"] = "public\n  interfaces: eth0\n"
	adapter := NewAdapterWithReader(reader)
	snapshot, _ := adapter.Observe(context.Background(), testScope(filter.FamilyInet))
	before := snapshot.Rules[0].Rule
	before.UUID = "owned"
	after := before
	after.DestinationPort = "9090"
	locator := snapshot.Rules[0].Locator

	update, err := adapter.Compile(snapshot, []filter.DesiredChange{{
		Operation: filter.ChangeUpdate, Before: &before, After: &after, Locator: &locator,
	}})
	if err != nil {
		t.Fatalf("compile update: %v", err)
	}
	if len(update.Rules[0].Commands) != 4 || len(update.Rules[0].RollbackCommands) != 4 ||
		update.Rules[0].Commands[0].Args[1] != "--remove-port=8080/tcp" || update.Rules[0].Commands[2].Args[1] != "--add-port=9090/tcp" {
		t.Fatalf("unexpected update plan: %#v", update.Rules[0])
	}
	deletePlan, err := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeDelete, Before: &before, Locator: &locator}})
	if err != nil {
		t.Fatalf("compile delete: %v", err)
	}
	if len(deletePlan.Rules[0].Commands) != 2 || deletePlan.Rules[0].Previous == nil || deletePlan.Rules[0].Commands[1].Args[2] != "--remove-port=8080/tcp" {
		t.Fatalf("unexpected delete plan: %#v", deletePlan.Rules[0])
	}
	verified, err := adapter.Verify(context.Background(), deletePlan)
	if err != nil || verified.Matched {
		t.Fatalf("delete verified while canonical target remained: result=%#v err=%v", verified, err)
	}
	reader.set(false, "--list-ports", "")
	reader.set(true, "--list-ports", "")
	verified, err = adapter.Verify(context.Background(), deletePlan)
	if err != nil || !verified.Matched {
		t.Fatalf("deleted canonical target did not verify: result=%#v err=%v", verified, err)
	}

	protected := snapshot
	protected.Rules = append([]filter.ObservedRule(nil), snapshot.Rules...)
	protected.Rules[0].Protected = true
	if _, err := adapter.Compile(protected, []filter.DesiredChange{{Operation: filter.ChangeDelete, Before: &before, Locator: &locator}}); !errors.Is(err, filter.ErrProtectedRule) {
		t.Fatalf("expected protected delete rejection, got %v", err)
	}
}

func TestApplyCompensatesOnlySuccessfulFirewalldSteps(t *testing.T) {
	reader := newFakeCommandReader()
	reader.outputs["--get-default-zone"] = "public\n"
	reader.outputs["--get-active-zones"] = "public\n  interfaces: eth0\n"
	writer := &fakeCommandWriter{failAt: 2, err: errors.New("permanent write failed")}
	adapter := NewAdapterWithBackend(reader, writer)
	snapshot, _ := adapter.Observe(context.Background(), testScope(filter.FamilyInet))
	rule := filter.FirewallRule{UUID: "web", Scope: testScope(filter.FamilyInet), Protocol: "tcp", DestinationPort: "80", Action: filter.ActionAccept}
	plan, _ := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})

	if _, err := adapter.Apply(context.Background(), plan); err == nil || !strings.Contains(err.Error(), "permanent write failed") {
		t.Fatalf("expected apply failure, got %v", err)
	}
	if len(writer.commands) != 3 || writer.commands[2].Args[1] != "--remove-port=80/tcp" {
		t.Fatalf("runtime write was not compensated precisely: %#v", writer.commands)
	}
}

func TestRollbackReversesFullyAppliedFirewalldPlan(t *testing.T) {
	reader := newFakeCommandReader()
	reader.outputs["--get-default-zone"] = "public\n"
	reader.outputs["--get-active-zones"] = "public\n  interfaces: eth0\n"
	writer := &fakeCommandWriter{}
	adapter := NewAdapterWithBackend(reader, writer)
	snapshot, _ := adapter.Observe(context.Background(), testScope(filter.FamilyInet))
	rule := filter.FirewallRule{UUID: "rollback", Scope: testScope(filter.FamilyInet), Protocol: "tcp", DestinationPort: "8080", Action: filter.ActionAccept}
	plan, err := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
	if err != nil {
		t.Fatalf("compile rollback plan: %v", err)
	}
	if err := adapter.Rollback(context.Background(), plan); err != nil {
		t.Fatalf("rollback applied plan: %v", err)
	}
	if len(writer.commands) != 2 || writer.commands[0].Args[0] != "--permanent" ||
		writer.commands[0].Args[2] != "--remove-port=8080/tcp" || writer.commands[1].Args[1] != "--remove-port=8080/tcp" {
		t.Fatalf("unexpected rollback writes: %#v", writer.commands)
	}
}

func TestVerifyRequiresConvergedCanonicalRule(t *testing.T) {
	reader := newFakeCommandReader()
	reader.outputs["--get-default-zone"] = "public\n"
	reader.outputs["--get-active-zones"] = "public\n  interfaces: eth0\n"
	adapter := NewAdapterWithReader(reader)
	snapshot, _ := adapter.Observe(context.Background(), testScope(filter.FamilyInet))
	rule := filter.FirewallRule{UUID: "dns", Scope: testScope(filter.FamilyInet), Protocol: "udp", DestinationPort: "53", Action: filter.ActionAccept}
	plan, _ := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}})
	reader.set(false, "--list-ports", "53/udp\n")

	verified, err := adapter.Verify(context.Background(), plan)
	if err != nil || verified.Matched {
		t.Fatalf("runtime-only rule verified as converged: result=%#v err=%v", verified, err)
	}
	reader.set(true, "--list-ports", "53/udp\n")
	verified, err = adapter.Verify(context.Background(), plan)
	if err != nil || !verified.Matched {
		t.Fatalf("converged rule did not verify: result=%#v err=%v", verified, err)
	}
}

func TestCompileRejectsBroadDenyAndPersistenceDrift(t *testing.T) {
	adapter := NewAdapterWithReader(newFakeCommandReader())
	scope := testScope(filter.FamilyInet)
	snapshot, _ := filter.NewSnapshot(scope, nil)
	rule := filter.FirewallRule{UUID: "deny-all", Scope: scope, Protocol: "all", Action: filter.ActionDrop}
	if _, err := adapter.Compile(snapshot, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &rule}}); !errors.Is(err, filter.ErrLockoutRisk) {
		t.Fatalf("expected broad deny rejection, got %v", err)
	}
	port := filter.FirewallRule{Scope: scope, NativeKind: filter.NativeKindZonePort, Protocol: "tcp", DestinationPort: "80", Action: filter.ActionAccept}
	observed := observedForRule(port)
	observed.Persistence = filter.PersistenceStatusRuntimeOnly
	drifted, _ := filter.NewSnapshot(scope, []filter.ObservedRule{observed})
	create := filter.FirewallRule{UUID: "https", Scope: scope, Protocol: "tcp", DestinationPort: "443", Action: filter.ActionAccept}
	if _, err := adapter.Compile(drifted, []filter.DesiredChange{{Operation: filter.ChangeCreate, After: &create}}); !errors.Is(err, filter.ErrRuleStale) {
		t.Fatalf("expected runtime/permanent drift rejection, got %v", err)
	}
}

func TestObservePublicInetMergesNativeObjectsAndReportsScopeNotices(t *testing.T) {
	reader := newFakeCommandReader()
	reader.set(false, "--list-ports", "22/tcp 53/udp 123/sctp\n")
	reader.set(true, "--list-ports", "22/tcp 80/tcp 123/sctp\n")
	reader.set(false, "--list-rich-rules", `rule port port="8080" protocol="tcp" accept`+"\n")
	reader.set(true, "--list-rich-rules", `rule port port="8080" protocol="tcp" accept`+"\n")
	reader.set(false, "--list-services", "ssh dhcpv6-client\n")
	reader.set(true, "--list-services", "dhcpv6-client ssh\n")
	reader.outputs["--get-default-zone"] = "work\n"
	reader.outputs["--get-active-zones"] = "public\n  interfaces: eth0\ndocker\n  interfaces: docker0\n"

	snapshot, err := NewAdapterWithReader(reader).Observe(context.Background(), testScope(filter.FamilyInet))
	if err != nil {
		t.Fatalf("observe public inet: %v", err)
	}
	if len(snapshot.Rules) != 7 {
		t.Fatalf("unexpected firewalld object count: %#v", snapshot.Rules)
	}
	assertPresence(t, snapshot.Rules, "port:22/tcp", filter.PersistenceStatusConverged)
	assertPresence(t, snapshot.Rules, "port:53/udp", filter.PersistenceStatusRuntimeOnly)
	assertPresence(t, snapshot.Rules, "port:80/tcp", filter.PersistenceStatusPermanentOnly)
	sctp := findObserved(snapshot.Rules, "port:123/sctp")
	if sctp == nil || sctp.ParseStatus != filter.ParseStatusOpaque {
		t.Fatalf("unsupported native port was guessed: %#v", sctp)
	}
	rich := findObserved(snapshot.Rules, `rich:rule port port="8080" protocol="tcp" accept`)
	if rich == nil || rich.ParseStatus != filter.ParseStatusSupported || rich.Rule.NativeKind != filter.NativeKindRichRule || rich.Rule.OrderBucket != filter.OrderBucketRichZeroAllow {
		t.Fatalf("family-neutral rich rule was not normalized: %#v", rich)
	}
	service := findObserved(snapshot.Rules, "service:ssh")
	if service == nil || service.ParseStatus != filter.ParseStatusOpaque || service.Rule.NativeKind != filter.NativeKindZoneService ||
		service.Rule.Protocol != "" || service.Rule.DestinationPort != "" || service.Rule.Description != "ssh" || service.Raw != "ssh" {
		t.Fatalf("service object was not preserved as opaque: %#v", service)
	}
	dhcpv6Client := findObserved(snapshot.Rules, "service:dhcpv6-client")
	if dhcpv6Client == nil || dhcpv6Client.Rule.Protocol != "" || dhcpv6Client.Rule.DestinationPort != "" ||
		dhcpv6Client.Rule.Description != "dhcpv6-client" || dhcpv6Client.Raw != "dhcpv6-client" {
		t.Fatalf("dhcpv6 service was exposed as an allow-all rule: %#v", dhcpv6Client)
	}
	if !hasNotice(snapshot.Notices, filter.ScopeNoticeDefaultScopeMismatch, "work") ||
		!hasNotice(snapshot.Notices, filter.ScopeNoticeUnmanagedActiveScopes, "docker") ||
		!hasNotice(snapshot.Notices, filter.ScopeNoticeRuntimePermanentMismatch, "ports") {
		t.Fatalf("scope notices missing: %#v", snapshot.Notices)
	}
}

func TestObservePublicPipelineKeepsRuleFamiliesInOneZoneScope(t *testing.T) {
	reader := newFakeCommandReader()
	rich := strings.Join([]string{
		`rule family="ipv4" priority="-100" source address="172.16.10.111" port port="3306" protocol="tcp" drop`,
		`rule family="ipv4" destination address="10.0.0.1" reject`,
		`rule family="ipv4" source address="10.0.0.0/8" protocol value="tcp" accept`,
		`rule family="ipv4" log prefix="audit" accept`,
		`rule family="ipv6" source address="2001:db8::1" accept`,
	}, "\n") + "\n"
	reader.set(false, "--list-rich-rules", rich)
	reader.set(true, "--list-rich-rules", rich)
	reader.outputs["--get-default-zone"] = "public\n"
	reader.outputs["--get-active-zones"] = "public\n  interfaces: eth0\n"

	snapshot, err := NewAdapterWithReader(reader).Observe(context.Background(), testScope(filter.FamilyIPv4))
	if err != nil {
		t.Fatalf("observe public IPv4: %v", err)
	}
	if snapshot.Scope.Family != filter.FamilyInet || len(snapshot.Rules) != 5 {
		t.Fatalf("unexpected public pipeline: %#v", snapshot)
	}
	first := snapshot.Rules[0]
	if first.ParseStatus != filter.ParseStatusSupported || first.Rule.SourceAddress != "172.16.10.111/32" || first.Rule.DestinationPort != "3306" ||
		first.Rule.Priority == nil || *first.Rule.Priority != -100 || first.Rule.OrderBucket != filter.OrderBucketRichPre {
		t.Fatalf("priority rich rule was not normalized: %#v", first)
	}
	opaque := findObserved(snapshot.Rules, `rich:rule family="ipv4" log prefix="audit" accept`)
	if opaque == nil || opaque.ParseStatus != filter.ParseStatusOpaque || opaque.Raw == "" {
		t.Fatalf("unsupported rich rule was not kept opaque: %#v", opaque)
	}
	protocolRule := findObserved(snapshot.Rules, `rich:rule family="ipv4" source address="10.0.0.0/8" protocol value="tcp" accept`)
	if protocolRule == nil || protocolRule.ParseStatus != filter.ParseStatusSupported || protocolRule.Rule.Protocol != "tcp" || protocolRule.Rule.DestinationPort != "" {
		t.Fatalf("protocol-only rich rule was not parsed: %#v", protocolRule)
	}
	ipv6 := findObserved(snapshot.Rules, `rich:rule family="ipv6" source address="2001:db8::1/128" accept`)
	if ipv6 == nil || ipv6.Rule.Scope.Family != filter.FamilyIPv6 {
		t.Fatalf("IPv6 rich rule was not retained in the public execution scope: %#v", ipv6)
	}
}

func TestNativeDetailRunsDedicatedFirewalldCommand(t *testing.T) {
	reader := newFakeCommandReader()
	info := "ssh\n  ports: 22/tcp\n  protocols:\n  source-ports:\n  helpers:\n  destination:"
	reader.setServiceInfo(false, "ssh", info)
	reader.setServiceInfo(true, "ssh", info)
	adapter := NewAdapterWithReader(reader)

	runtimeInfo, err := adapter.NativeDetail(context.Background(), "ssh", false)
	if err != nil {
		t.Fatalf("read runtime service info: %v", err)
	}
	permanentInfo, err := adapter.NativeDetail(context.Background(), "ssh", true)
	if err != nil {
		t.Fatalf("read permanent service info: %v", err)
	}
	if runtimeInfo != info || permanentInfo != info {
		t.Fatalf("service info output was not preserved: runtime=%q permanent=%q", runtimeInfo, permanentInfo)
	}
}

func TestNativeDetailRejectsInvalidServiceName(t *testing.T) {
	adapter := NewAdapterWithReader(newFakeCommandReader())
	if _, err := adapter.NativeDetail(context.Background(), "../ssh", false); !errors.Is(err, filter.ErrInvalidRule) {
		t.Fatalf("expected invalid service rejection, got %v", err)
	}
}

func TestObserveRejectsZoneOutsidePublic(t *testing.T) {
	scope := testScope(filter.FamilyInet)
	scope.Zone = "work"
	_, err := NewAdapterWithReader(newFakeCommandReader()).Observe(context.Background(), scope)
	if !errors.Is(err, filter.ErrUnsupportedScope) {
		t.Fatalf("expected unsupported scope, got %v", err)
	}
}

func TestZoneNoticesReportInactivePublic(t *testing.T) {
	notices := zoneNotices("public", []string{"docker"}, zoneOutput{}, zoneOutput{})
	if !hasNotice(notices, filter.ScopeNoticeManagedScopeInactive, "") || !hasNotice(notices, filter.ScopeNoticeUnmanagedActiveScopes, "docker") {
		t.Fatalf("inactive public notices missing: %#v", notices)
	}
}

type fakeCommandReader struct {
	outputs map[string]string
	errors  map[string]error
}

type fakeCommandWriter struct {
	commands []filter.NativeCommand
	failAt   int
	err      error
}

func (f *fakeCommandWriter) Run(_ context.Context, command filter.NativeCommand) error {
	f.commands = append(f.commands, command)
	if f.failAt > 0 && len(f.commands) == f.failAt {
		return f.err
	}
	return nil
}

func newFakeCommandReader() *fakeCommandReader {
	reader := &fakeCommandReader{outputs: make(map[string]string), errors: make(map[string]error)}
	for _, option := range []string{"--list-ports", "--list-rich-rules", "--list-services"} {
		reader.set(false, option, "")
		reader.set(true, option, "")
	}
	return reader
}

func (f *fakeCommandReader) set(permanent bool, option, output string) {
	args := make([]string, 0, 3)
	if permanent {
		args = append(args, "--permanent")
	}
	args = append(args, "--zone=public", option)
	f.outputs[strings.Join(args, "\x00")] = output
}

func (f *fakeCommandReader) setServiceInfo(permanent bool, service, output string) {
	args := make([]string, 0, 2)
	if permanent {
		args = append(args, "--permanent")
	}
	args = append(args, "--info-service="+service)
	f.outputs[strings.Join(args, "\x00")] = output
}

func (f *fakeCommandReader) Read(_ context.Context, args ...string) (string, error) {
	key := strings.Join(args, "\x00")
	if err := f.errors[key]; err != nil {
		return "", err
	}
	output, exists := f.outputs[key]
	if !exists {
		return "", errors.New("unexpected firewall-cmd call: " + strings.Join(args, " "))
	}
	return output, nil
}

func testScope(family filter.Family) filter.Scope {
	return filter.Scope{Provider: filter.ProviderFirewalld, Family: family, Zone: "public", Direction: filter.DirectionInput}
}

func findObserved(rules []filter.ObservedRule, canonical string) *filter.ObservedRule {
	for index := range rules {
		if rules[index].Locator.Canonical == canonical {
			return &rules[index]
		}
	}
	return nil
}

func assertPresence(t *testing.T, rules []filter.ObservedRule, canonical string, expected filter.PersistenceStatus) {
	t.Helper()
	rule := findObserved(rules, canonical)
	if rule == nil || rule.Persistence != expected {
		t.Fatalf("unexpected presence for %s: %#v", canonical, rule)
	}
}

func hasNotice(notices []filter.ScopeNotice, code filter.ScopeNoticeCode, value string) bool {
	for _, notice := range notices {
		if notice.Code != code {
			continue
		}
		if value == "" {
			return true
		}
		for _, candidate := range notice.Values {
			if candidate == value {
				return true
			}
		}
	}
	return false
}
