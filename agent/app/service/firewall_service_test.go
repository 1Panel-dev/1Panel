package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	filterfirewalld "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/firewalld"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/glebarez/sqlite"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestFirewallBaseInfoDistinguishesMissingAndConflictingProviders(t *testing.T) {
	wantErr := errors.New("firewalld and ufw conflict")
	for _, test := range []struct {
		name      string
		installed []string
		exists    bool
		message   string
	}{
		{name: "missing", installed: nil},
		{name: "conflict", installed: []string{constant.FirewallProviderFirewalld, constant.FirewallProviderUFW}, exists: true, message: wantErr.Error()},
	} {
		t.Run(test.name, func(t *testing.T) {
			service := &FirewallService{
				baseClient: func() (lifecycle.Client, error) { return nil, wantErr },
				installedProviders: func() []string {
					return append([]string(nil), test.installed...)
				},
			}
			base, err := service.LoadBaseInfo("base")
			if err != nil {
				t.Fatal(err)
			}
			if base.IsExist != test.exists || base.Message != test.message {
				t.Fatalf("unexpected firewall failure status: %#v", base)
			}
		})
	}
}

func TestFirewallResetCleansDirectBackendAndKeepsStoredRules(t *testing.T) {
	db := newFirewallRuleTestDB(t)
	ruleRepo := repo.NewFirewallRuleRepo(db)
	stored, err := model.FirewallRuleFromDomain(executorTestAddressRule("172.16.10.111"))
	if err != nil {
		t.Fatal(err)
	}
	stored.UUID = "reset-direct-rule"
	stored.Origin = constant.FirewallRuleOriginCreated
	stored.Owner = constant.FirewallRuleSourceUser
	if err := ruleRepo.Create(context.Background(), &stored); err != nil {
		t.Fatal(err)
	}
	cleaned := ""
	service := &FirewallService{
		rules: ruleRepo,
		selectedProvider: func(context.Context) (filter.Provider, error) {
			return filter.ProviderIptables, nil
		},
		cleanupBackend: func(provider string) error {
			cleaned = provider
			return nil
		},
	}
	result, err := service.Reset(context.Background(), dto.FirewallRuleReset{})
	if err != nil {
		t.Fatal(err)
	}
	if cleaned != string(filter.ProviderIptables) || result.Removed != 1 {
		t.Fatalf("unexpected reset result: cleaned=%q result=%#v", cleaned, result)
	}
	remaining, err := ruleRepo.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 1 || remaining[0].UUID != stored.UUID {
		t.Fatalf("reset removed provider-neutral database policy: %#v", remaining)
	}
}

func TestFirewallResetCleansInactiveDirectBackendWithoutChangingSelectedBackendState(t *testing.T) {
	service := &FirewallService{
		rules: repo.NewFirewallRuleRepo(newFirewallRuleTestDB(t)),
		selectedProvider: func(context.Context) (filter.Provider, error) {
			return filter.ProviderNftables, nil
		},
		cleanupBackend: func(string) error {
			t.Fatal("inactive source cleanup used the selected-backend cleanup path")
			return nil
		},
		cleanupInactiveBackend: func(provider string) error {
			if provider != string(filter.ProviderIptables) {
				t.Fatalf("inactive cleanup provider = %q", provider)
			}
			return nil
		},
	}
	result, err := service.Reset(context.Background(), dto.FirewallRuleReset{Provider: filter.ProviderIptables})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Disabled {
		t.Fatalf("inactive source cleanup result = %#v", result)
	}
}

func TestFirewallResetRestoresUFWDefaultsAndKeepsStoredRules(t *testing.T) {
	scope := filter.Scope{
		Provider: filter.ProviderUFW, Family: filter.FamilyIPv4,
		Chain: filter.UFWInputChain, Direction: filter.DirectionInput,
	}
	protectedRule := filter.FirewallRule{
		Scope: scope, NativeKind: filter.NativeKindUFWRule, Protocol: "tcp",
		DestinationPort: "22", Action: filter.ActionAccept,
	}
	db := newFirewallRuleTestDB(t)
	ruleRepo := repo.NewFirewallRuleRepo(db)
	stored, err := model.FirewallRuleFromDomain(protectedRule)
	if err != nil {
		t.Fatal(err)
	}
	stored.UUID = "protected-reset-rule"
	stored.Origin = constant.FirewallRuleOriginCreated
	stored.Owner = constant.FirewallRuleSourceUser
	if err := ruleRepo.Create(context.Background(), &stored); err != nil {
		t.Fatal(err)
	}
	resetProvider := ""
	service := &FirewallService{
		rules:            ruleRepo,
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderNftables, nil },
		resetBackend: func(provider string) error {
			if provider != string(filter.ProviderUFW) {
				t.Fatalf("reset unexpected provider %q", provider)
			}
			resetProvider = provider
			return nil
		},
	}
	result, err := service.Reset(context.Background(), dto.FirewallRuleReset{Provider: filter.ProviderUFW})
	if err != nil {
		t.Fatal(err)
	}
	if resetProvider != string(filter.ProviderUFW) || result.Removed != 1 || !result.Disabled {
		t.Fatalf("unexpected reset result: provider=%q result=%#v", resetProvider, result)
	}
	remaining, err := ruleRepo.List(context.Background())
	if err != nil || len(remaining) != 1 || remaining[0].UUID != stored.UUID {
		t.Fatalf("reset removed provider-neutral UFW policy: %#v err=%v", remaining, err)
	}
}

func TestFirewallRuleServiceCheckCreateInventoryWorkflow(t *testing.T) {
	rule := executorTestAddressRule("172.16.10.111")
	external := executorObservedRule(rule, "", 1)
	adapter := newFakeFilterAdapter(t, rule.Scope, []filter.ObservedRule{external})
	adapter.snapshot.Notices = []filter.ScopeNotice{{Code: filter.ScopeNoticeManagedScopeInactive}}
	db := newFirewallRuleTestDB(t)
	service := &FirewallService{
		rules:            repo.NewFirewallRuleRepo(db),
		adapters:         firewallRuleRuntimeRegistry{filter.ProviderIptables: newFirewallRuleRuntime(adapter, nil)},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil },
	}
	ctx := context.Background()

	before, err := service.Inventory(ctx, dto.FirewallRuleInventory{Scope: rule.Scope})
	if err != nil {
		t.Fatalf("inventory before adoption: %v", err)
	}
	if len(before.Items) != 1 || before.Items[0].State != filter.InventoryStateExternal {
		t.Fatalf("expected external inventory: %#v", before)
	}
	if len(before.Notices) != 1 || before.Notices[0].Code != filter.ScopeNoticeManagedScopeInactive {
		t.Fatalf("scope notices were not returned: %#v", before.Notices)
	}
	adapter.snapshot.Notices = nil
	check, err := service.checkRule(ctx, "", dto.FirewallRuleCheckItem{Rule: rule})
	if err != nil {
		t.Fatalf("plan adoption: %v", err)
	}
	if check.Classification != filter.CheckClassificationExactExternal {
		t.Fatalf("unexpected check result: %#v", check)
	}
	if check.CheckFlag == "" {
		t.Fatal("check did not return a creation flag")
	}
	err = service.createFirewallRuleItem(ctx, dto.FirewallRuleCreateItem{
		CheckFlag: check.CheckFlag, Action: filter.CheckActionAdopt,
		AdoptInstanceKey: check.Candidates[0].InstanceKey, Rule: check.RequestedRule,
	})
	if err != nil {
		t.Fatalf("commit adoption: %v", err)
	}
	after, err := service.Inventory(ctx, dto.FirewallRuleInventory{Scope: rule.Scope})
	if err != nil {
		t.Fatalf("inventory after adoption: %v", err)
	}
	if len(after.Items) != 1 || after.Items[0].State != filter.InventoryStateAdopted || after.Items[0].Desired == nil {
		t.Fatalf("adopted ownership missing from inventory: %#v", after)
	}
	deleted, err := service.Delete(ctx, dto.FirewallRuleDelete{UUIDs: []string{after.Items[0].Desired.UUID}})
	if err != nil || deleted.Failed > 0 {
		t.Fatalf("delete adopted rule: result=%#v err=%v", deleted, err)
	}
	empty, err := service.Inventory(ctx, dto.FirewallRuleInventory{Scope: rule.Scope})
	if err != nil || len(empty.Items) != 0 {
		t.Fatalf("deleted rule remained in inventory: inventory=%#v err=%v", empty, err)
	}
}

func TestFirewallRuleServiceCombinedUFWInventory(t *testing.T) {
	ipv4Scope := filter.Scope{
		Provider: filter.ProviderUFW, Family: filter.FamilyIPv4,
		Chain: filter.UFWInputChain, Direction: filter.DirectionInput,
	}
	ipv6Scope := ipv4Scope
	ipv6Scope.Family = filter.FamilyIPv6
	ipv4Rule := filter.FirewallRule{
		Scope: ipv4Scope, NativeKind: filter.NativeKindUFWRule,
		Protocol: "tcp", DestinationPort: "8080", Action: filter.ActionAccept,
	}
	ipv6Rule := filter.FirewallRule{
		Scope: ipv6Scope, NativeKind: filter.NativeKindUFWRule,
		Protocol: "udp", SourceAddress: "2001:db8::/64", DestinationPort: "5353", Action: filter.ActionDrop,
	}
	adapter := newFakeFilterAdapter(t, ipv4Scope, []filter.ObservedRule{executorObservedRule(ipv4Rule, "", 1)})
	ipv6Snapshot, err := filter.NewSnapshot(ipv6Scope, []filter.ObservedRule{executorObservedRule(ipv6Rule, "", 2)})
	if err != nil {
		t.Fatalf("create IPv6 snapshot: %v", err)
	}
	adapter.multiSnapshots = []filter.Snapshot{adapter.snapshot, ipv6Snapshot}
	adapter.multiSnapshots[0].Notices = []filter.ScopeNotice{{Code: filter.ScopeNoticeManagedScopeInactive}}
	adapter.multiSnapshots[1].Notices = []filter.ScopeNotice{{Code: filter.ScopeNoticeManagedScopeInactive}}
	service := &FirewallService{
		rules: repo.NewFirewallRuleRepo(newFirewallRuleTestDB(t)),
		adapters: firewallRuleRuntimeRegistry{
			filter.ProviderUFW: newFirewallRuleRuntime(adapter, nil),
		},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderUFW, nil },
	}

	result, err := service.Inventory(context.Background(), dto.FirewallRuleInventory{Scope: filter.Scope{
		Provider: filter.ProviderUFW, Family: filter.FamilyInet,
		Chain: filter.UFWInputChain, Direction: filter.DirectionInput,
	}})
	if err != nil {
		t.Fatalf("load combined UFW inventory: %v", err)
	}
	if adapter.observeScopesCount != 1 || adapter.observeCount != 0 {
		t.Fatalf("unexpected UFW observation counts: multi=%d single=%d", adapter.observeScopesCount, adapter.observeCount)
	}
	if len(result.Items) != 2 || result.Items[0].Rule.Scope.Family != filter.FamilyIPv4 ||
		result.Items[1].Rule.Scope.Family != filter.FamilyIPv6 {
		t.Fatalf("unexpected combined UFW inventory: %#v", result.Items)
	}
	if len(result.Notices) != 1 || result.Notices[0].Code != filter.ScopeNoticeManagedScopeInactive {
		t.Fatalf("duplicate or missing combined UFW notices: %#v", result.Notices)
	}
}

func TestFirewallRuleServiceCheckBlocksConfiguredProtectedPort(t *testing.T) {
	rule := executorTestRule("8443")
	rule.Action = filter.ActionDrop
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	service, _ := newTestFirewallExecutor(t, adapter)
	service.selectedProvider = func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil }
	service.protectedPorts = func() ([]firewall.PortWhitelist, error) {
		return []firewall.PortWhitelist{{Family: "ipv4", Port: "8443", Protocol: "tcp"}}, nil
	}

	result, err := service.checkRule(context.Background(), "203.0.113.9", dto.FirewallRuleCheckItem{Rule: rule})
	if err != nil {
		t.Fatalf("check protected port: %v", err)
	}
	if result.Decision != filter.CheckDecisionBlocked || result.Reason != "current_management_connection" {
		t.Fatalf("configured protected port was not blocked: %#v", result)
	}
}

func TestFirewallRuleServiceLoadsNativeDetailOnDemand(t *testing.T) {
	scope := filter.Scope{
		Provider: filter.ProviderFirewalld, Family: filter.FamilyInet,
		Zone: filter.FirewalldInputZone, Direction: filter.DirectionInput,
	}
	adapter := newFakeFilterAdapter(t, scope, nil)
	adapter.nativeDetail = "ssh\n  ports: 22/tcp\n  protocols:\n  source-ports:\n  helpers:\n  destination:"
	service := &FirewallService{
		adapters: firewallRuleRuntimeRegistry{filter.ProviderFirewalld: newFirewallRuleRuntime(adapter, nil)},
		selectedProvider: func(context.Context) (filter.Provider, error) {
			return filter.ProviderFirewalld, nil
		},
	}

	info, err := service.LoadFirewallNativeDetail(context.Background(), dto.FirewallNativeDetail{
		Provider: filter.ProviderFirewalld, NativeKind: filter.NativeKindZoneService,
		Name: "ssh", Permanent: true,
	})
	if err != nil {
		t.Fatalf("load service info: %v", err)
	}
	if info != adapter.nativeDetail || adapter.nativeDetailName != "ssh" || !adapter.nativeDetailPermanent {
		t.Fatalf("service info request was not passed through: info=%q adapter=%#v", info, adapter)
	}
}

func TestFirewallRuleServiceCreateRequiresCheck(t *testing.T) {
	rule := executorTestRule("8080")
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	service, ruleRepo := newTestFirewallExecutor(t, adapter)
	service.selectedProvider = func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil }

	result, err := service.Create(context.Background(), dto.FirewallRuleCreate{Items: []dto.FirewallRuleCreateItem{{
		Rule: rule, Action: filter.CheckActionCreate,
	}}})
	if err != nil || result.Failed != 1 || len(result.Errors) != 1 ||
		result.Errors[0].Error != filter.ErrRuleCheckRequired.Error() {
		t.Fatalf("expected check-required failure, result=%#v err=%v", result, err)
	}
	rules, _ := ruleRepo.List(context.Background())
	if adapter.applyCount != 0 || len(rules) != 0 {
		t.Fatalf("unchecked rule was created: applyCount=%d rules=%#v", adapter.applyCount, rules)
	}
}

func TestFirewallRuleServiceHonorsExplicitUFWCreatePosition(t *testing.T) {
	scope := filter.Scope{
		Provider: filter.ProviderUFW, Family: filter.FamilyIPv4,
		Chain: filter.UFWInputChain, Direction: filter.DirectionInput,
	}
	existingRule := filter.FirewallRule{
		Scope: scope, NativeKind: filter.NativeKindUFWRule,
		Protocol: "tcp", DestinationPort: "80", Action: filter.ActionAccept,
	}
	existing := executorObservedRule(existingRule, "", 4)
	adapter := newFakeFilterAdapter(t, scope, []filter.ObservedRule{existing})
	service, _ := newTestFirewallExecutor(t, adapter)
	order := int64(2)
	rule := filter.FirewallRule{
		Scope: scope, NativeKind: filter.NativeKindUFWRule,
		Protocol: "tcp", DestinationPort: "8080", Action: filter.ActionAccept, OrderIndex: &order,
	}

	if err := createExecutorRule(service, adapter, dto.FirewallRuleCreateItem{
		Rule: rule, Action: filter.CheckActionCreate, SourceKind: constant.FirewallRuleSourceUser,
	}); err != nil {
		t.Fatalf("create UFW rule at explicit position: %v", err)
	}
	if adapter.lastChange.Append || adapter.lastChange.After == nil || adapter.lastChange.After.OrderIndex == nil ||
		*adapter.lastChange.After.OrderIndex != order {
		t.Fatalf("UFW explicit create position was not preserved: %#v", adapter.lastChange)
	}
}

func TestFirewallRuleServiceDefaultsMissingUFWCreatePositionToAppend(t *testing.T) {
	scope := filter.Scope{
		Provider: filter.ProviderUFW, Family: filter.FamilyIPv4,
		Chain: filter.UFWInputChain, Direction: filter.DirectionInput,
	}
	existingRule := filter.FirewallRule{
		Scope: scope, NativeKind: filter.NativeKindUFWRule,
		Protocol: "tcp", DestinationPort: "80", Action: filter.ActionAccept,
	}
	adapter := newFakeFilterAdapter(t, scope, []filter.ObservedRule{executorObservedRule(existingRule, "", 4)})
	service, _ := newTestFirewallExecutor(t, adapter)
	rule := filter.FirewallRule{
		Scope: scope, NativeKind: filter.NativeKindUFWRule,
		Protocol: "tcp", DestinationPort: "8080", Action: filter.ActionAccept,
	}

	if err := createExecutorRule(service, adapter, dto.FirewallRuleCreateItem{
		Rule: rule, Action: filter.CheckActionCreate, SourceKind: constant.FirewallRuleSourceUser,
	}); err != nil {
		t.Fatalf("append UFW rule: %v", err)
	}
	if !adapter.lastChange.Append || adapter.lastChange.After == nil || adapter.lastChange.After.OrderIndex == nil ||
		*adapter.lastChange.After.OrderIndex != 5 {
		t.Fatalf("missing UFW position did not default to append: %#v", adapter.lastChange)
	}
}

func TestFirewallRuleServiceDefaultsMissingUFWRequestScope(t *testing.T) {
	scope := filter.Scope{
		Provider: filter.ProviderUFW, Family: filter.FamilyIPv4,
		Chain: filter.UFWInputChain, Direction: filter.DirectionInput,
	}
	adapter := newFakeFilterAdapter(t, scope, nil)
	service, _ := newTestFirewallExecutor(t, adapter)
	service.selectedProvider = func(context.Context) (filter.Provider, error) { return filter.ProviderUFW, nil }

	checked, err := service.Check(context.Background(), "", dto.FirewallRuleCheck{Items: []dto.FirewallRuleCheckItem{{
		Rule: filter.FirewallRule{Protocol: "tcp", DestinationPort: "55101", Action: filter.ActionAccept},
	}}})
	if err != nil {
		t.Fatalf("check UFW rule with omitted scope: %v", err)
	}
	if len(checked.Items) != 1 || checked.Items[0].RequestedRule.Scope.Normalize() != scope.Normalize() {
		t.Fatalf("unexpected defaulted UFW scope: %#v", checked.Items)
	}
}

func TestValidateUFWPositionWithinFamilyBounds(t *testing.T) {
	ipv4Scope := filter.Scope{
		Provider: filter.ProviderUFW, Family: filter.FamilyIPv4,
		Chain: filter.UFWInputChain, Direction: filter.DirectionInput,
	}
	ipv4Rule := filter.FirewallRule{
		Scope: ipv4Scope, NativeKind: filter.NativeKindUFWRule,
		Protocol: "tcp", DestinationPort: "80", Action: filter.ActionAccept,
	}
	ipv4Snapshot, err := filter.NewSnapshot(ipv4Scope, []filter.ObservedRule{
		executorObservedRule(ipv4Rule, "", 1),
		executorObservedRule(ipv4Rule, "", 4),
	})
	if err != nil {
		t.Fatalf("create IPv4 snapshot: %v", err)
	}
	if err = validatePositionTarget(context.Background(), nil, ipv4Snapshot, ipv4Rule, 4); err != nil {
		t.Fatalf("valid IPv4 position was rejected: %v", err)
	}
	if err = validatePositionTarget(context.Background(), nil, ipv4Snapshot, ipv4Rule, 5); !errors.Is(err, filter.ErrInvalidRule) {
		t.Fatalf("IPv6 position was accepted for IPv4 rule: %v", err)
	}

	ipv6Scope := ipv4Scope
	ipv6Scope.Family = filter.FamilyIPv6
	ipv6Rule := ipv4Rule
	ipv6Rule.Scope = ipv6Scope
	ipv6Snapshot, err := filter.NewSnapshot(ipv6Scope, []filter.ObservedRule{
		executorObservedRule(ipv6Rule, "", 5),
		executorObservedRule(ipv6Rule, "", 8),
	})
	if err != nil {
		t.Fatalf("create IPv6 snapshot: %v", err)
	}
	if err = validatePositionTarget(context.Background(), nil, ipv6Snapshot, ipv6Rule, 4); !errors.Is(err, filter.ErrInvalidRule) {
		t.Fatalf("IPv4 position was accepted for IPv6 rule: %v", err)
	}
}

func TestUFWAppendPositionUsesFamilyBoundary(t *testing.T) {
	ipv4Scope := filter.Scope{
		Provider: filter.ProviderUFW, Family: filter.FamilyIPv4,
		Chain: filter.UFWInputChain, Direction: filter.DirectionInput,
	}
	ipv4Rule := filter.FirewallRule{
		Scope: ipv4Scope, NativeKind: filter.NativeKindUFWRule,
		Protocol: "tcp", DestinationPort: "80", Action: filter.ActionAccept,
	}
	ipv4Snapshot, err := filter.NewSnapshot(ipv4Scope, []filter.ObservedRule{
		executorObservedRule(ipv4Rule, "", 1),
		executorObservedRule(ipv4Rule, "", 4),
	})
	if err != nil {
		t.Fatalf("create IPv4 snapshot: %v", err)
	}
	if position, positionErr := ufwAppendPosition(context.Background(), nil, ipv4Snapshot, ipv4Rule); positionErr != nil || position != 5 {
		t.Fatalf("unexpected IPv4 append position: position=%d err=%v", position, positionErr)
	}

	ipv6Scope := ipv4Scope
	ipv6Scope.Family = filter.FamilyIPv6
	ipv6Rule := ipv4Rule
	ipv6Rule.Scope = ipv6Scope
	ipv6Snapshot, err := filter.NewSnapshot(ipv6Scope, []filter.ObservedRule{
		executorObservedRule(ipv6Rule, "", 5),
		executorObservedRule(ipv6Rule, "", 8),
	})
	if err != nil {
		t.Fatalf("create IPv6 snapshot: %v", err)
	}
	ipv4Adapter := newFakeFilterAdapter(t, ipv4Scope, ipv4Snapshot.Rules)
	runtime := newFirewallRuleRuntime(ipv4Adapter, nil)
	if position, positionErr := ufwAppendPosition(context.Background(), runtime, ipv6Snapshot, ipv6Rule); positionErr != nil || position != 9 {
		t.Fatalf("unexpected IPv6 append position: position=%d err=%v", position, positionErr)
	}
}

func TestUFWGlobalEndPositionIncludesOtherFamily(t *testing.T) {
	ipv4Scope := filter.Scope{
		Provider: filter.ProviderUFW, Family: filter.FamilyIPv4,
		Chain: filter.UFWInputChain, Direction: filter.DirectionInput,
	}
	ipv4Rule := filter.FirewallRule{
		Scope: ipv4Scope, NativeKind: filter.NativeKindUFWRule,
		Protocol: "tcp", DestinationPort: "4422,8088", Action: filter.ActionAccept,
	}
	ipv4Snapshot, err := filter.NewSnapshot(ipv4Scope, []filter.ObservedRule{
		executorObservedRule(ipv4Rule, "1panel-rule:managed", 4),
	})
	if err != nil {
		t.Fatalf("create IPv4 snapshot: %v", err)
	}

	ipv6Scope := ipv4Scope
	ipv6Scope.Family = filter.FamilyIPv6
	ipv6Rule := ipv4Rule
	ipv6Rule.Scope = ipv6Scope
	ipv6Snapshot, err := filter.NewSnapshot(ipv6Scope, []filter.ObservedRule{
		executorObservedRule(ipv6Rule, "", 5),
		executorObservedRule(ipv6Rule, "", 8),
	})
	if err != nil {
		t.Fatalf("create IPv6 snapshot: %v", err)
	}
	runtime := newFirewallRuleRuntime(newFakeFilterAdapter(t, ipv6Scope, ipv6Snapshot.Rules), nil)
	maxPosition, err := maxPositionForRule(context.Background(), runtime, ipv4Snapshot, ipv4Rule)
	if err != nil {
		t.Fatalf("load UFW global maximum position: %v", err)
	}
	if maxPosition != 8 {
		t.Fatalf("IPv4 family boundary was mistaken for the global end: %d", maxPosition)
	}
	if int64(*ipv4Snapshot.Rules[0].Locator.Position) == maxPosition {
		t.Fatal("last IPv4 rule was incorrectly classified as the global last UFW rule")
	}
}

func TestFirewallRuleServiceBatchCheckAndCreateSameScope(t *testing.T) {
	rules := []filter.FirewallRule{executorTestRule("8080"), executorTestRule("8081")}
	adapter := newFakeFilterAdapter(t, rules[0].Scope, nil)
	db := newFirewallRuleTestDB(t)
	ruleRepo := repo.NewFirewallRuleRepo(db)
	service := &FirewallService{
		rules:            ruleRepo,
		adapters:         firewallRuleRuntimeRegistry{filter.ProviderIptables: newFirewallRuleRuntime(adapter, nil)},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil },
	}
	ctx := context.Background()

	checked, err := service.Check(ctx, "", dto.FirewallRuleCheck{Items: firewallRuleCheckItems(rules)})
	if err != nil || len(checked.Items) != len(rules) {
		t.Fatalf("batch check: result=%#v err=%v", checked, err)
	}
	if adapter.observeCount != 1 {
		t.Fatalf("same-scope batch check observed the chain %d times", adapter.observeCount)
	}
	request := dto.FirewallRuleCreate{Items: make([]dto.FirewallRuleCreateItem, 0, len(rules))}
	for _, item := range checked.Items {
		request.Items = append(request.Items, dto.FirewallRuleCreateItem{
			Rule: item.RequestedRule, CheckFlag: item.CheckFlag, Action: filter.CheckActionCreate,
			SourceKind: constant.FirewallRuleSourceUser,
		})
	}
	created, err := service.Create(ctx, request)
	if err != nil || created.Succeeded != 2 || created.Failed != 0 {
		t.Fatalf("batch create: result=%#v err=%v", created, err)
	}
	stored, _ := ruleRepo.List(ctx)
	if len(stored) != 2 || len(adapter.snapshot.Rules) != 2 || adapter.applyCount != 1 {
		t.Fatalf("same-scope batch did not commit all rules: stored=%#v snapshot=%#v applies=%d", stored, adapter.snapshot, adapter.applyCount)
	}
	if adapter.observeCount != 3 {
		t.Fatalf("same-scope batch create used repeated snapshots: observes=%d", adapter.observeCount)
	}
}

func TestFirewallRuleServiceBatchDeleteSameIptablesScope(t *testing.T) {
	rules := []filter.FirewallRule{executorTestRule("8080"), executorTestRule("8081")}
	adapter := newFakeFilterAdapter(t, rules[0].Scope, nil)
	db := newFirewallRuleTestDB(t)
	ruleRepo := repo.NewFirewallRuleRepo(db)
	service := &FirewallService{
		rules:            ruleRepo,
		adapters:         firewallRuleRuntimeRegistry{filter.ProviderIptables: newFirewallRuleRuntime(adapter, nil)},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil },
	}
	ctx := context.Background()
	checked, err := service.Check(ctx, "", dto.FirewallRuleCheck{Items: firewallRuleCheckItems(rules)})
	if err != nil {
		t.Fatalf("batch check: %v", err)
	}
	createRequest := dto.FirewallRuleCreate{Items: make([]dto.FirewallRuleCreateItem, 0, len(rules))}
	for _, item := range checked.Items {
		createRequest.Items = append(createRequest.Items, dto.FirewallRuleCreateItem{
			Rule: item.RequestedRule, CheckFlag: item.CheckFlag, Action: filter.CheckActionCreate,
		})
	}
	created, err := service.Create(ctx, createRequest)
	if err != nil || created.Succeeded != 2 {
		t.Fatalf("batch create: result=%#v err=%v", created, err)
	}
	stored, err := ruleRepo.List(ctx)
	if err != nil || len(stored) != 2 {
		t.Fatalf("list created rules: %#v err=%v", stored, err)
	}
	deleted, err := service.Delete(ctx, dto.FirewallRuleDelete{UUIDs: []string{stored[0].UUID, stored[1].UUID}})
	if err != nil || deleted.Succeeded != 2 || deleted.Failed != 0 {
		t.Fatalf("batch delete: result=%#v err=%v", deleted, err)
	}
	remaining, listErr := ruleRepo.List(ctx)
	if listErr != nil || len(remaining) != 0 || len(adapter.snapshot.Rules) != 0 || adapter.applyCount != 2 {
		t.Fatalf(
			"same-scope delete did not use one backend apply: remaining=%#v snapshot=%#v applies=%d err=%v",
			remaining, adapter.snapshot, adapter.applyCount, listErr,
		)
	}
}

func TestFirewallRuleServiceBatchCreateAndDeleteSameNftablesScope(t *testing.T) {
	rules := []filter.FirewallRule{executorTestRule("8080"), executorTestRule("8081")}
	for index := range rules {
		rules[index].Scope.Provider = filter.ProviderNftables
	}
	adapter := newFakeFilterAdapter(t, rules[0].Scope, nil)
	db := newFirewallRuleTestDB(t)
	ruleRepo := repo.NewFirewallRuleRepo(db)
	service := &FirewallService{
		rules:            ruleRepo,
		adapters:         firewallRuleRuntimeRegistry{filter.ProviderNftables: newFirewallRuleRuntime(adapter, nil)},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderNftables, nil },
	}
	ctx := context.Background()
	checked, err := service.Check(ctx, "", dto.FirewallRuleCheck{Items: firewallRuleCheckItems(rules)})
	if err != nil {
		t.Fatalf("batch check: %v", err)
	}
	createRequest := dto.FirewallRuleCreate{Items: make([]dto.FirewallRuleCreateItem, 0, len(rules))}
	for _, item := range checked.Items {
		createRequest.Items = append(createRequest.Items, dto.FirewallRuleCreateItem{
			Rule: item.RequestedRule, CheckFlag: item.CheckFlag, Action: filter.CheckActionCreate,
		})
	}
	created, err := service.Create(ctx, createRequest)
	if err != nil || created.Succeeded != 2 || adapter.applyCount != 1 {
		t.Fatalf("nftables batch create: result=%#v applies=%d err=%v", created, adapter.applyCount, err)
	}
	stored, err := ruleRepo.List(ctx)
	if err != nil || len(stored) != 2 {
		t.Fatalf("list nftables rules: %#v err=%v", stored, err)
	}
	deleted, err := service.Delete(ctx, dto.FirewallRuleDelete{UUIDs: []string{stored[0].UUID, stored[1].UUID}})
	if err != nil || deleted.Succeeded != 2 || deleted.Failed != 0 || adapter.applyCount != 2 {
		t.Fatalf("nftables batch delete: result=%#v applies=%d err=%v", deleted, adapter.applyCount, err)
	}
	remaining, err := ruleRepo.List(ctx)
	if err != nil || len(remaining) != 0 || len(adapter.snapshot.Rules) != 0 {
		t.Fatalf("nftables batch delete did not clear rules: stored=%#v snapshot=%#v err=%v", remaining, adapter.snapshot, err)
	}
}

func TestFirewallRuleServiceBatchDeleteRollsBackOnMetadataFailure(t *testing.T) {
	rules := []filter.FirewallRule{executorTestRule("8080"), executorTestRule("8081")}
	adapter := newFakeFilterAdapter(t, rules[0].Scope, nil)
	db := newFirewallRuleTestDB(t)
	ruleRepo := repo.NewFirewallRuleRepo(db)
	failingRepo := &failingFirewallRuleRepo{IFirewallRuleRepo: ruleRepo, deleteErr: errors.New("metadata delete failed")}
	service := &FirewallService{
		rules:            failingRepo,
		adapters:         firewallRuleRuntimeRegistry{filter.ProviderIptables: newFirewallRuleRuntime(adapter, nil)},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil },
	}
	ctx := context.Background()
	checked, err := service.Check(ctx, "", dto.FirewallRuleCheck{Items: firewallRuleCheckItems(rules)})
	if err != nil {
		t.Fatalf("batch check: %v", err)
	}
	createRequest := dto.FirewallRuleCreate{Items: make([]dto.FirewallRuleCreateItem, 0, len(rules))}
	for _, item := range checked.Items {
		createRequest.Items = append(createRequest.Items, dto.FirewallRuleCreateItem{
			Rule: item.RequestedRule, CheckFlag: item.CheckFlag, Action: filter.CheckActionCreate,
		})
	}
	created, err := service.Create(ctx, createRequest)
	if err != nil || created.Succeeded != 2 {
		t.Fatalf("batch create: result=%#v err=%v", created, err)
	}
	stored, err := ruleRepo.List(ctx)
	if err != nil || len(stored) != 2 {
		t.Fatalf("list created rules: %#v err=%v", stored, err)
	}
	deleted, err := service.Delete(ctx, dto.FirewallRuleDelete{UUIDs: []string{stored[0].UUID, stored[1].UUID}})
	if err != nil || deleted.Succeeded != 0 || deleted.Failed != 2 {
		t.Fatalf("batch delete failure: result=%#v err=%v", deleted, err)
	}
	remaining, listErr := ruleRepo.List(ctx)
	if listErr != nil || len(remaining) != 2 || len(adapter.snapshot.Rules) != 2 || adapter.applyCount != 2 || adapter.rollbackCount != 1 {
		t.Fatalf(
			"failed delete batch was not rolled back: remaining=%#v snapshot=%#v applies=%d rollbacks=%d err=%v",
			remaining, adapter.snapshot, adapter.applyCount, adapter.rollbackCount, listErr,
		)
	}
}

func TestFirewallRuleServiceBatchCreateRejectsDuplicateManagedRule(t *testing.T) {
	rule := executorTestRule("8080")
	trailingRule := executorTestRule("8081")
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	db := newFirewallRuleTestDB(t)
	ruleRepo := repo.NewFirewallRuleRepo(db)
	service := &FirewallService{
		rules:            ruleRepo,
		adapters:         firewallRuleRuntimeRegistry{filter.ProviderIptables: newFirewallRuleRuntime(adapter, nil)},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil },
	}
	ctx := context.Background()

	checked, err := service.Check(ctx, "", dto.FirewallRuleCheck{
		Items: firewallRuleCheckItems([]filter.FirewallRule{rule, rule, trailingRule}),
	})
	if err != nil || len(checked.Items) != 3 {
		t.Fatalf("batch check duplicate rules: result=%#v err=%v", checked, err)
	}
	request := dto.FirewallRuleCreate{Items: make([]dto.FirewallRuleCreateItem, 0, 3)}
	for _, item := range checked.Items {
		request.Items = append(request.Items, dto.FirewallRuleCreateItem{
			Rule: item.RequestedRule, CheckFlag: item.CheckFlag, Action: filter.CheckActionCreate,
			SourceKind: constant.FirewallRuleSourceUser,
		})
	}

	created, err := service.Create(ctx, request)
	if err != nil || created.Succeeded != 1 || created.Failed != 1 || created.Skipped != 1 {
		t.Fatalf("batch duplicate create: result=%#v err=%v", created, err)
	}
	if len(created.Errors) != 2 || created.Errors[0].Index != 1 || created.Errors[0].Status != "failed" ||
		created.Errors[0].Error == "" || created.Errors[0].Rule.DestinationPort != "8080" ||
		created.Errors[1].Index != 2 || created.Errors[1].Status != "skipped" || created.Errors[1].Error != "" ||
		created.Errors[1].Rule.DestinationPort != "8081" {
		t.Fatalf("batch failure details missing: %#v", created.Errors)
	}
	stored, listErr := ruleRepo.List(ctx)
	if listErr != nil || len(stored) != 1 || len(adapter.snapshot.Rules) != 1 || adapter.applyCount != 1 {
		t.Fatalf(
			"duplicate managed rule was persisted: stored=%#v snapshot=%#v applies=%d err=%v",
			stored, adapter.snapshot, adapter.applyCount, listErr,
		)
	}
}

func TestFirewallRuleServiceIptablesBatchDoesNotPersistRuntimeMetadata(t *testing.T) {
	rules := []filter.FirewallRule{executorTestRule("8080"), executorTestRule("8081")}
	adapter := newFakeFilterAdapter(t, rules[0].Scope, nil)
	db := newFirewallRuleTestDB(t)
	ruleRepo := repo.NewFirewallRuleRepo(db)
	service := &FirewallService{
		rules: &failingFirewallRuleRepo{
			IFirewallRuleRepo: ruleRepo,
			updateErr:         errors.New("metadata write failed"),
		},
		adapters:         firewallRuleRuntimeRegistry{filter.ProviderIptables: newFirewallRuleRuntime(adapter, nil)},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil },
	}
	ctx := context.Background()
	checked, err := service.Check(ctx, "", dto.FirewallRuleCheck{Items: firewallRuleCheckItems(rules)})
	if err != nil {
		t.Fatalf("batch check: %v", err)
	}
	request := dto.FirewallRuleCreate{Items: make([]dto.FirewallRuleCreateItem, 0, len(rules))}
	for _, item := range checked.Items {
		request.Items = append(request.Items, dto.FirewallRuleCreateItem{
			Rule: item.RequestedRule, CheckFlag: item.CheckFlag, Action: filter.CheckActionCreate,
			SourceKind: constant.FirewallRuleSourceUser,
		})
	}
	created, err := service.Create(ctx, request)
	if err != nil || created.Succeeded != 2 || created.Failed != 0 || created.Skipped != 0 {
		t.Fatalf("batch result: %#v err=%v", created, err)
	}
	stored, listErr := ruleRepo.List(ctx)
	if listErr != nil || len(stored) != 2 || len(adapter.snapshot.Rules) != 2 || adapter.applyCount != 1 || adapter.rollbackCount != 0 {
		t.Fatalf(
			"batch persisted runtime metadata: stored=%#v snapshot=%#v applies=%d rollbacks=%d err=%v",
			stored, adapter.snapshot, adapter.applyCount, adapter.rollbackCount, listErr,
		)
	}
}

func TestFirewallRuleServiceCreateRejectsChangedFirewallState(t *testing.T) {
	rule := executorTestRule("8080")
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	service, ruleRepo := newTestFirewallExecutor(t, adapter)
	service.selectedProvider = func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil }
	ctx := context.Background()

	check, err := service.checkRule(ctx, "", dto.FirewallRuleCheckItem{Rule: rule})
	if err != nil {
		t.Fatalf("check rule: %v", err)
	}
	other := executorTestRule("9090")
	adapter.snapshot, err = filter.NewSnapshot(rule.Scope, []filter.ObservedRule{executorObservedRule(other, "", 1)})
	if err != nil {
		t.Fatalf("change firewall snapshot: %v", err)
	}
	err = service.createFirewallRuleItem(ctx, dto.FirewallRuleCreateItem{
		Rule: check.RequestedRule, CheckFlag: check.CheckFlag, Action: filter.CheckActionCreate,
	})
	if !errors.Is(err, filter.ErrRuleCheckRequired) {
		t.Fatalf("expected changed snapshot to require another check, got %v", err)
	}
	rules, _ := ruleRepo.List(ctx)
	if adapter.applyCount != 0 || len(rules) != 0 {
		t.Fatalf("stale checked rule was created: applyCount=%d rules=%#v", adapter.applyCount, rules)
	}
}

func TestFirewallRuleServiceCreateRejectsRuleChangedAfterCheck(t *testing.T) {
	rule := executorTestRule("8080")
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	service, ruleRepo := newTestFirewallExecutor(t, adapter)
	service.selectedProvider = func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil }
	ctx := context.Background()

	check, err := service.checkRule(ctx, "", dto.FirewallRuleCheckItem{Rule: rule})
	if err != nil {
		t.Fatalf("check rule: %v", err)
	}
	changed := check.RequestedRule
	changed.DestinationPort = "8081"
	err = service.createFirewallRuleItem(ctx, dto.FirewallRuleCreateItem{
		Rule: changed, CheckFlag: check.CheckFlag, Action: filter.CheckActionCreate,
	})
	if !errors.Is(err, filter.ErrRuleCheckRequired) {
		t.Fatalf("expected changed rule to require another check, got %v", err)
	}
	rules, _ := ruleRepo.List(ctx)
	if adapter.applyCount != 0 || len(rules) != 0 {
		t.Fatalf("changed unchecked rule was created: applyCount=%d rules=%#v", adapter.applyCount, rules)
	}
}

func TestFirewallRuleServiceCreateRejectsChangedManagedState(t *testing.T) {
	rule := executorTestRule("8080")
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	service, ruleRepo := newTestFirewallExecutor(t, adapter)
	service.selectedProvider = func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil }
	ctx := context.Background()

	check, err := service.checkRule(ctx, "", dto.FirewallRuleCheckItem{Rule: rule})
	if err != nil {
		t.Fatalf("check rule: %v", err)
	}
	other := executorTestRule("9090")
	record, err := firewallRuleModelForCreate(other, dto.FirewallRuleCreateItem{SourceKind: constant.FirewallRuleSourceUser}, constant.FirewallRuleOriginCreated)
	if err != nil {
		t.Fatalf("build managed record: %v", err)
	}
	if err := ruleRepo.Create(ctx, &record); err != nil {
		t.Fatalf("change managed state: %v", err)
	}
	err = service.createFirewallRuleItem(ctx, dto.FirewallRuleCreateItem{
		Rule: check.RequestedRule, CheckFlag: check.CheckFlag, Action: filter.CheckActionCreate,
	})
	if !errors.Is(err, filter.ErrRuleCheckRequired) {
		t.Fatalf("expected changed managed state to require another check, got %v", err)
	}
	if adapter.applyCount != 0 {
		t.Fatalf("rule was applied with a stale managed-state check: %d", adapter.applyCount)
	}
}

func TestFirewallRuleServiceRejectsUnavailableProductionAdapter(t *testing.T) {
	db := newFirewallRuleTestDB(t)
	service := &FirewallService{
		rules: repo.NewFirewallRuleRepo(db), adapters: firewallRuleRuntimeRegistry{},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil },
	}
	_, err := service.checkRule(context.Background(), "", dto.FirewallRuleCheckItem{Rule: executorTestRule("80")})
	if !errors.Is(err, filter.ErrAdapterUnavailable) {
		t.Fatalf("expected unavailable adapter error, got %v", err)
	}
}

func TestPrepareFirewallRuleUsesProviderRepresentation(t *testing.T) {
	rule, err := filter.NormalizeRule(filter.FirewallRule{
		Scope:    filter.Scope{Provider: filter.ProviderFirewalld, Family: filter.FamilyInet, Zone: "public", Direction: filter.DirectionInput},
		Protocol: "tcp", DestinationPort: "443", Action: filter.ActionAccept,
	})
	if err != nil {
		t.Fatalf("normalize request: %v", err)
	}
	prepared, err := newFirewallRuleRuntime(filterfirewalld.NewAdapterWithReader(nil), nil).Prepare(rule)
	if err != nil {
		t.Fatalf("prepare firewalld request: %v", err)
	}
	if prepared.NativeKind != filter.NativeKindZonePort {
		t.Fatalf("service planned against the wrong native identity: %#v", prepared)
	}
}

func TestNormalizeFirewallRuleScopeDefaultsIptablesChains(t *testing.T) {
	input := filter.Scope{Provider: filter.ProviderIptables, Family: filter.FamilyIPv4, Table: "filter", Direction: filter.DirectionInput}.Normalize()
	if input.Chain != filter.IptablesInputChain {
		t.Fatalf("unexpected input chain: %#v", input)
	}
}

func TestNormalizeSystemPortsKeepsFamilyAndRange(t *testing.T) {
	ports, err := normalizeSystemPorts([]dto.FirewallSystemPort{
		{Family: "ipv4", Protocol: "tcp", Port: "80"},
		{Family: "ipv6", Protocol: "udp", Port: "8000:8100"},
	})
	if err != nil {
		t.Fatalf("normalize system ports: %v", err)
	}
	if _, ok := ports["ipv4/tcp/80"]; !ok {
		t.Fatalf("missing IPv4 port: %#v", ports)
	}
	if port, ok := ports["ipv6/udp/8000-8100"]; !ok || port.Family != "ipv6" || port.Port != "8000-8100" {
		t.Fatalf("missing normalized IPv6 range: %#v", ports)
	}
}

func TestSystemPortRuleUsesRequestedFamily(t *testing.T) {
	port := dto.FirewallSystemPort{Family: "ipv6", Protocol: "tcp", Port: "443"}
	for _, provider := range []filter.Provider{
		filter.ProviderIptables, filter.ProviderNftables, filter.ProviderFirewalld, filter.ProviderUFW,
	} {
		rule := systemPortRule(provider, port)
		if rule.Scope.Family != filter.FamilyIPv6 {
			t.Fatalf("provider %s used family %s", provider, rule.Scope.Family)
		}
	}
}

func TestSyncSystemPortsCreatesTracksAndDeletesAcceptedPort(t *testing.T) {
	scope := filter.Scope{
		Provider: filter.ProviderIptables, Family: filter.FamilyIPv4, Table: "filter",
		Chain: filter.IptablesInputChain, Direction: filter.DirectionInput,
	}
	adapter := newFakeFilterAdapter(t, scope, nil)
	ruleRepo := repo.NewFirewallRuleRepo(newFirewallRuleTestDB(t))
	engine := &FirewallService{
		rules: ruleRepo,
		adapters: firewallRuleRuntimeRegistry{
			filter.ProviderIptables: newFirewallRuleRuntime(adapter, nil),
		},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil },
	}
	port := dto.FirewallSystemPort{Port: "8443", Protocol: "TCP"}

	if err := engine.SyncSystemPorts(context.Background(), nil, []dto.FirewallSystemPort{port}); err != nil {
		t.Fatalf("create accepted port: %v", err)
	}
	stored, err := ruleRepo.List(context.Background(), repo.WithFirewallRuleSource(
		constant.FirewallRuleSourceSecurity, constant.FirewallSystemAcceptedPortSourcePrefix+"tcp/8443",
	))
	if err != nil || len(stored) != 1 {
		t.Fatalf("accepted port ownership was not persisted: rules=%#v err=%v", stored, err)
	}
	if len(adapter.snapshot.Rules) != 1 {
		t.Fatalf("accepted port was not applied: %#v", adapter.snapshot.Rules)
	}

	if err := engine.SyncSystemPorts(context.Background(), []dto.FirewallSystemPort{port}, nil); err != nil {
		t.Fatalf("delete accepted port: %v", err)
	}
	if len(adapter.snapshot.Rules) != 0 {
		t.Fatalf("accepted port remained after deletion: %#v", adapter.snapshot.Rules)
	}
	present, err := ruleRepo.List(context.Background(),
		repo.WithFirewallRuleSource(constant.FirewallRuleSourceSecurity, constant.FirewallSystemAcceptedPortSourcePrefix+"tcp/8443"),
	)
	if err != nil || len(present) != 0 {
		t.Fatalf("accepted port ownership remained present: rules=%#v err=%v", present, err)
	}
}

func TestSyncSystemPortsBatchesNativeAcceptedPortsByScope(t *testing.T) {
	scope := filter.Scope{
		Provider: filter.ProviderIptables, Family: filter.FamilyIPv4, Table: "filter",
		Chain: filter.IptablesInputChain, Direction: filter.DirectionInput,
	}
	adapter := newFakeFilterAdapter(t, scope, nil)
	ruleRepo := repo.NewFirewallRuleRepo(newFirewallRuleTestDB(t))
	engine := &FirewallService{
		rules: ruleRepo,
		adapters: firewallRuleRuntimeRegistry{
			filter.ProviderIptables: newFirewallRuleRuntime(adapter, nil),
		},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil },
	}
	ports := []dto.FirewallSystemPort{
		{Family: "ipv4", Port: "8080", Protocol: "tcp"},
		{Family: "ipv4", Port: "8443", Protocol: "tcp"},
		{Family: "ipv4", Port: "5353", Protocol: "udp"},
	}

	if err := engine.SyncSystemPorts(context.Background(), nil, ports); err != nil {
		t.Fatalf("batch create accepted ports: %v", err)
	}
	if adapter.applyCount != 1 || len(adapter.snapshot.Rules) != len(ports) {
		t.Fatalf("accepted ports were not created in one native apply: applies=%d rules=%d", adapter.applyCount, len(adapter.snapshot.Rules))
	}
	if err := engine.SyncSystemPorts(context.Background(), ports, nil); err != nil {
		t.Fatalf("batch delete accepted ports: %v", err)
	}
	if adapter.applyCount != 2 || len(adapter.snapshot.Rules) != 0 {
		t.Fatalf("accepted ports were not deleted in one native apply: applies=%d rules=%d", adapter.applyCount, len(adapter.snapshot.Rules))
	}
}

func TestSyncSystemPortsCreatesAcceptedPortDespitePartialOppositeOverlap(t *testing.T) {
	scope := filter.Scope{
		Provider: filter.ProviderFirewalld, Family: filter.FamilyInet,
		Zone: filter.FirewalldInputZone, Direction: filter.DirectionInput,
	}
	existing := filter.FirewallRule{
		Scope: filter.Scope{
			Provider: filter.ProviderFirewalld, Family: filter.FamilyIPv4,
			Zone: filter.FirewalldInputZone, Direction: filter.DirectionInput,
		},
		NativeKind: filter.NativeKindRichRule, Protocol: "all",
		SourceAddress: "1.1.1.1", Action: filter.ActionDrop,
	}
	adapter := newFakeFilterAdapter(t, scope, []filter.ObservedRule{executorObservedRule(existing, "", 1)})
	engine := &FirewallService{
		rules: repo.NewFirewallRuleRepo(newFirewallRuleTestDB(t)),
		adapters: firewallRuleRuntimeRegistry{
			filter.ProviderFirewalld: newFirewallRuleRuntime(adapter, nil),
		},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderFirewalld, nil },
	}

	err := engine.SyncSystemPorts(context.Background(), nil, []dto.FirewallSystemPort{{Port: "443", Protocol: "tcp"}})
	if err != nil {
		t.Fatalf("create partially overlapping accepted port: %v", err)
	}
	if adapter.applyCount != 1 || len(adapter.snapshot.Rules) != 2 {
		t.Fatalf("accepted port was not created: applyCount=%d rules=%#v", adapter.applyCount, adapter.snapshot.Rules)
	}
}

func TestSyncSystemPortsRejectsFullyCoveredOppositeRule(t *testing.T) {
	scope := filter.Scope{
		Provider: filter.ProviderFirewalld, Family: filter.FamilyInet,
		Zone: filter.FirewalldInputZone, Direction: filter.DirectionInput,
	}
	existing := filter.FirewallRule{
		Scope: scope, NativeKind: filter.NativeKindRichRule, Protocol: "tcp",
		DestinationPort: "443", Action: filter.ActionDrop,
	}
	adapter := newFakeFilterAdapter(t, scope, []filter.ObservedRule{executorObservedRule(existing, "", 1)})
	engine := &FirewallService{
		rules: repo.NewFirewallRuleRepo(newFirewallRuleTestDB(t)),
		adapters: firewallRuleRuntimeRegistry{
			filter.ProviderFirewalld: newFirewallRuleRuntime(adapter, nil),
		},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderFirewalld, nil },
	}

	err := engine.SyncSystemPorts(context.Background(), nil, []dto.FirewallSystemPort{{Port: "443", Protocol: "tcp"}})
	if err == nil || !strings.Contains(err.Error(), "overlapping_rule_with_different_action") {
		t.Fatalf("fully covered opposite rule returned %v", err)
	}
	if adapter.applyCount != 0 {
		t.Fatalf("fully conflicting accepted port was applied %d times", adapter.applyCount)
	}
}

func TestSyncSystemPortsDoesNotTakeOverExistingManagedRule(t *testing.T) {
	scope := filter.Scope{
		Provider: filter.ProviderIptables, Family: filter.FamilyIPv4, Table: "filter",
		Chain: filter.IptablesInputChain, Direction: filter.DirectionInput,
	}
	port := dto.FirewallSystemPort{Port: "443", Protocol: "tcp"}
	rule := systemPortRule(filter.ProviderIptables, port)
	observed := executorObservedRule(rule, "1panel-rule:user-rule", 1)
	adapter := newFakeFilterAdapter(t, scope, []filter.ObservedRule{observed})
	ruleRepo := repo.NewFirewallRuleRepo(newFirewallRuleTestDB(t))
	userRecord, err := firewallRuleModelForCreate(rule, dto.FirewallRuleCreateItem{SourceKind: constant.FirewallRuleSourceUser}, constant.FirewallRuleOriginCreated)
	if err != nil {
		t.Fatal(err)
	}
	userRecord.UUID = "user-rule"
	if err := ruleRepo.Create(context.Background(), &userRecord); err != nil {
		t.Fatal(err)
	}
	engine := &FirewallService{
		rules: ruleRepo,
		adapters: firewallRuleRuntimeRegistry{
			filter.ProviderIptables: newFirewallRuleRuntime(adapter, nil),
		},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil },
	}

	if err := engine.SyncSystemPorts(context.Background(), nil, []dto.FirewallSystemPort{port}); err != nil {
		t.Fatalf("reuse existing managed rule: %v", err)
	}
	systemOwned, err := ruleRepo.List(context.Background(), repo.WithFirewallRuleSource(
		constant.FirewallRuleSourceSecurity, constant.FirewallSystemAcceptedPortSourcePrefix+"tcp/443",
	))
	if err != nil || len(systemOwned) != 0 || adapter.applyCount != 0 {
		t.Fatalf("existing user rule was taken over: rules=%#v applyCount=%d err=%v", systemOwned, adapter.applyCount, err)
	}
}

func TestSyncSystemPortsAdoptsAndDeletesLegacyAcceptedPort(t *testing.T) {
	scope := filter.Scope{
		Provider: filter.ProviderIptables, Family: filter.FamilyIPv4, Table: "filter",
		Chain: filter.IptablesInputChain, Direction: filter.DirectionInput,
	}
	port := dto.FirewallSystemPort{Port: "8080", Protocol: "tcp"}
	external := executorObservedRule(systemPortRule(filter.ProviderIptables, port), "", 1)
	adapter := newFakeFilterAdapter(t, scope, []filter.ObservedRule{external})
	ruleRepo := repo.NewFirewallRuleRepo(newFirewallRuleTestDB(t))
	engine := &FirewallService{
		rules: ruleRepo,
		adapters: firewallRuleRuntimeRegistry{
			filter.ProviderIptables: newFirewallRuleRuntime(adapter, nil),
		},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil },
	}

	if err := engine.SyncSystemPorts(context.Background(), []dto.FirewallSystemPort{port}, nil); err != nil {
		t.Fatalf("remove legacy accepted port: %v", err)
	}
	if len(adapter.snapshot.Rules) != 0 || adapter.applyCount != 2 {
		t.Fatalf("legacy rule did not use adopt/delete workflow: snapshot=%#v applyCount=%d", adapter.snapshot, adapter.applyCount)
	}
}
func TestFirewallExecutorCreatesAndVerifiesRule(t *testing.T) {
	rule := executorTestRule("8080")
	position := int64(1)
	rule.OrderIndex = &position
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)
	request := dto.FirewallRuleCreateItem{
		Rule: rule, SourceKind: constant.FirewallRuleSourceUser,
	}

	if err := createExecutorRule(executor, adapter, request); err != nil {
		t.Fatalf("commit create: %v", err)
	}
	if adapter.applyCount != 1 {
		t.Fatalf("unexpected apply count: %d", adapter.applyCount)
	}
	rules, _ := ruleRepo.List(context.Background())
	if len(rules) != 1 || rules[0].DestinationPort != "8080" {
		t.Fatalf("rule was not verified and bound: %#v", rules)
	}
}

func TestFirewallExecutorAdoptsWithoutAddingEquivalentRule(t *testing.T) {
	rule := executorTestAddressRule("172.16.10.111")
	external := executorObservedRule(rule, "", 1)
	adapter := newFakeFilterAdapter(t, rule.Scope, []filter.ObservedRule{external})
	instanceKey, _ := filter.InstanceKey(external)
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)
	err := createExecutorRule(executor, adapter, dto.FirewallRuleCreateItem{
		AdoptInstanceKey: instanceKey, Rule: rule, SourceKind: constant.FirewallRuleSourceUser,
	})
	if err != nil {
		t.Fatalf("commit adoption: %v", err)
	}
	if len(adapter.snapshot.Rules) != 1 || adapter.snapshot.Rules[0].Marker == "" {
		t.Fatalf("adoption changed rule count or missed marker: snapshot=%#v", adapter.snapshot)
	}
	rules, _ := ruleRepo.List(context.Background())
	if len(rules) != 1 || rules[0].Origin != constant.FirewallRuleOriginAdopted {
		t.Fatalf("adopted ownership was not persisted: %#v", rules)
	}
}

func TestFirewallExecutorCleansUpVerificationFailure(t *testing.T) {
	rule := executorTestRule("9090")
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	adapter.verifyMatched = false
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)

	err := createExecutorRule(executor, adapter, dto.FirewallRuleCreateItem{
		Rule: rule, SourceKind: constant.FirewallRuleSourceUser,
	})
	if !errors.Is(err, filter.ErrVerificationFailed) {
		t.Fatalf("expected verification failure, got err=%v", err)
	}
	rules, _ := ruleRepo.List(context.Background())
	if len(rules) != 0 {
		t.Fatalf("failed rule metadata was not cleaned up: %#v", rules)
	}
	if len(adapter.snapshot.Rules) != 0 || adapter.rollbackCount != 1 {
		t.Fatalf("failed runtime rule was not rolled back: snapshot=%#v rollbacks=%d", adapter.snapshot, adapter.rollbackCount)
	}
}

func TestFirewallExecutorCreateDoesNotPersistRuntimeMetadata(t *testing.T) {
	rule := executorTestRule("9091")
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)
	executor.rules = &failingFirewallRuleRepo{
		IFirewallRuleRepo: ruleRepo,
		updateErr:         errors.New("commit failed"),
	}

	err := createExecutorRule(executor, adapter, dto.FirewallRuleCreateItem{
		Rule: rule, SourceKind: constant.FirewallRuleSourceUser,
	})
	if err != nil {
		t.Fatalf("runtime metadata update affected create: %v", err)
	}
	stored, _ := ruleRepo.List(context.Background())
	if len(stored) != 1 || len(adapter.snapshot.Rules) != 1 || adapter.rollbackCount != 0 {
		t.Fatalf("create persisted runtime metadata: stored=%#v snapshot=%#v rollbacks=%d", stored, adapter.snapshot, adapter.rollbackCount)
	}
}

func TestFirewallExecutorRollsBackDeleteWhenPersistenceDeleteFails(t *testing.T) {
	rule := executorTestRule("9092")
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)
	if err := createExecutorRule(executor, adapter, dto.FirewallRuleCreateItem{
		Rule: rule, SourceKind: constant.FirewallRuleSourceUser,
	}); err != nil {
		t.Fatalf("create managed rule: %v", err)
	}
	stored, _ := ruleRepo.List(context.Background())
	executor.rules = &failingFirewallRuleRepo{
		IFirewallRuleRepo: ruleRepo,
		deleteErr:         errors.New("delete commit failed"),
	}

	err := executor.deleteRule(context.Background(), stored[0].UUID)
	if err == nil || !strings.Contains(err.Error(), "delete commit failed") {
		t.Fatalf("expected persistence delete failure, got %v", err)
	}
	remaining, _ := ruleRepo.List(context.Background())
	if len(remaining) != 1 || len(adapter.snapshot.Rules) != 1 || adapter.rollbackCount != 1 {
		t.Fatalf("failed delete was not restored: stored=%#v snapshot=%#v rollbacks=%d", remaining, adapter.snapshot, adapter.rollbackCount)
	}
}

func TestFirewallExecutorRollsBackUpdateWhenPersistenceUpdateFails(t *testing.T) {
	rule := executorTestRule("9093")
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)
	if err := createExecutorRule(executor, adapter, dto.FirewallRuleCreateItem{
		Rule: rule, SourceKind: constant.FirewallRuleSourceUser,
	}); err != nil {
		t.Fatalf("create managed rule: %v", err)
	}
	stored, _ := ruleRepo.List(context.Background())
	executor.rules = &failingFirewallRuleRepo{
		IFirewallRuleRepo: ruleRepo,
		updateErr:         errors.New("update commit failed"),
	}
	updated := rule
	updated.DestinationPort = "9443"

	err := executor.updateRule(context.Background(), "", stored[0].UUID, updated)
	if err == nil || !strings.Contains(err.Error(), "update commit failed") {
		t.Fatalf("expected persistence update failure, got %v", err)
	}
	remaining, _ := ruleRepo.List(context.Background())
	if len(remaining) != 1 || remaining[0].DestinationPort != "9093" || len(adapter.snapshot.Rules) != 1 ||
		adapter.snapshot.Rules[0].Rule.DestinationPort != "9093" || adapter.rollbackCount != 1 {
		t.Fatalf("failed update was not restored: stored=%#v snapshot=%#v rollbacks=%d", remaining, adapter.snapshot, adapter.rollbackCount)
	}
}

func TestFirewallExecutorDeletesOnlyVerifiedManagedRule(t *testing.T) {
	rule := executorTestRule("8443")
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)
	err := createExecutorRule(executor, adapter, dto.FirewallRuleCreateItem{
		Rule: rule, SourceKind: constant.FirewallRuleSourceUser,
	})
	if err != nil {
		t.Fatalf("create managed rule: %v", err)
	}
	stored, err := ruleRepo.List(context.Background())
	if err != nil || len(stored) != 1 {
		t.Fatalf("load managed rule: rules=%#v err=%v", stored, err)
	}
	if err := executor.deleteRule(context.Background(), stored[0].UUID); err != nil {
		t.Fatalf("delete managed rule: %v", err)
	}
	if len(adapter.snapshot.Rules) != 0 || adapter.applyCount != 2 {
		t.Fatalf("unexpected delete result: snapshot=%#v applies=%d", adapter.snapshot, adapter.applyCount)
	}
	remaining, err := ruleRepo.List(context.Background())
	if err != nil || len(remaining) != 0 {
		t.Fatalf("deleted rule was not archived: rules=%#v err=%v", remaining, err)
	}
}

func TestFirewallExecutorRefusesDeleteAfterManagedRuleDrifts(t *testing.T) {
	rule := executorTestRule("9443")
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)
	err := createExecutorRule(executor, adapter, dto.FirewallRuleCreateItem{
		Rule: rule, SourceKind: constant.FirewallRuleSourceUser,
	})
	if err != nil {
		t.Fatalf("create managed rule: %v", err)
	}
	stored, _ := ruleRepo.List(context.Background())
	adapter.snapshot.Rules[0].Rule.DestinationPort = "9444"
	adapter.snapshot, _ = filter.NewSnapshot(adapter.snapshot.Scope, adapter.snapshot.Rules)
	err = executor.deleteRule(context.Background(), stored[0].UUID)
	if !errors.Is(err, filter.ErrRuleStale) {
		t.Fatalf("expected drifted delete rejection, got %v", err)
	}
	if adapter.applyCount != 1 {
		t.Fatalf("drifted rule was deleted: applyCount=%d", adapter.applyCount)
	}
}

func TestFirewallExecutorUpdatesManagedRule(t *testing.T) {
	rule := executorTestRule("8080")
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)
	err := createExecutorRule(executor, adapter, dto.FirewallRuleCreateItem{
		Rule: rule, SourceKind: constant.FirewallRuleSourceUser,
	})
	if err != nil {
		t.Fatalf("create managed rule: %v", err)
	}
	stored, _ := ruleRepo.List(context.Background())
	updated := rule
	updated.DestinationPort = "8443"
	updated.Description = "updated"
	if err := executor.updateRule(context.Background(), "", stored[0].UUID, updated); err != nil {
		t.Fatalf("update managed rule: %v", err)
	}
	if adapter.applyCount != 2 {
		t.Fatalf("unexpected update apply count: %d", adapter.applyCount)
	}
	after, _ := ruleRepo.GetByUUID(context.Background(), stored[0].UUID)
	desired, err := desiredFirewallRuleFromModel(after)
	if err != nil {
		t.Fatalf("decode updated rule: %v", err)
	}
	if desired.Rule.DestinationPort != "8443" || desired.Rule.Description != "updated" {
		t.Fatalf("updated semantics were not persisted: %#v", after)
	}
}

func TestFirewallExecutorUpdatesUFWDescriptionWithoutNativeMutation(t *testing.T) {
	rule := executorTestRule("8080")
	rule.Scope = filter.Scope{
		Provider: filter.ProviderUFW, Family: filter.FamilyIPv4,
		Chain: filter.UFWInputChain, Direction: filter.DirectionInput,
	}
	rule.NativeKind = filter.NativeKindUFWRule
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)
	if err := createExecutorRule(executor, adapter, dto.FirewallRuleCreateItem{
		Rule: rule, SourceKind: constant.FirewallRuleSourceUser,
	}); err != nil {
		t.Fatalf("create managed UFW rule: %v", err)
	}
	stored, _ := ruleRepo.List(context.Background())
	updated := rule
	updated.Description = "updated description"
	if err := executor.updateRule(context.Background(), "", stored[0].UUID, updated); err != nil {
		t.Fatalf("update managed UFW description: %v", err)
	}
	if adapter.applyCount != 1 {
		t.Fatalf("description-only update changed UFW: applyCount=%d", adapter.applyCount)
	}
	after, err := ruleRepo.GetByUUID(context.Background(), stored[0].UUID)
	if err != nil {
		t.Fatalf("load updated UFW rule: %v", err)
	}
	if after.Description != "updated description" || after.DestinationPort != "8080" {
		t.Fatalf("unexpected persisted UFW rule: %#v", after)
	}
}

func TestIsUFWMetadataOnlyUpdate(t *testing.T) {
	position := 3
	order := int64(position)
	before := executorTestRule("8080")
	before.Scope = filter.Scope{
		Provider: filter.ProviderUFW, Family: filter.FamilyIPv4,
		Chain: filter.UFWInputChain, Direction: filter.DirectionInput,
	}
	before.NativeKind = filter.NativeKindUFWRule
	before.OrderIndex = &order
	after := before
	after.Description = "new description"

	metadataOnly, err := isUFWMetadataOnlyUpdate(before, after, filter.Locator{Position: &position})
	if err != nil || !metadataOnly {
		t.Fatalf("description-only UFW update = %v, err=%v", metadataOnly, err)
	}

	tests := []struct {
		name    string
		mutate  func(*filter.FirewallRule, *filter.Locator)
		wantErr bool
	}{
		{name: "rule changed", mutate: func(rule *filter.FirewallRule, _ *filter.Locator) { rule.DestinationPort = "8081" }},
		{name: "order changed", mutate: func(rule *filter.FirewallRule, _ *filter.Locator) { changed := int64(4); rule.OrderIndex = &changed }},
		{name: "missing locator", mutate: func(_ *filter.FirewallRule, locator *filter.Locator) { locator.Position = nil }},
		{name: "invalid rule", mutate: func(rule *filter.FirewallRule, _ *filter.Locator) { rule.DestinationPort = "invalid" }, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := after
			locator := filter.Locator{Position: &position}
			test.mutate(&candidate, &locator)
			got, err := isUFWMetadataOnlyUpdate(before, candidate, locator)
			if test.wantErr {
				if err == nil {
					t.Fatal("invalid rule did not return an error")
				}
				return
			}
			if err != nil || got {
				t.Fatalf("metadata-only update = %v, err=%v", got, err)
			}
		})
	}
}

func TestFirewallRuntimeRejectsUnavailableManagedScopeForMutation(t *testing.T) {
	rule := executorTestRule("8080")
	for _, notice := range []filter.ScopeNoticeCode{
		filter.ScopeNoticeManagedScopeInactive,
		filter.ScopeNoticeManagedScopeMissing,
	} {
		t.Run(string(notice), func(t *testing.T) {
			adapter := newFakeFilterAdapter(t, rule.Scope, nil)
			adapter.snapshot.Notices = []filter.ScopeNotice{{Code: notice}}
			runtime := newFirewallRuleRuntime(adapter, nil)
			if _, err := runtime.ObserveMutation(context.Background(), rule.Scope); !errors.Is(err, filter.ErrProviderUnavailable) {
				t.Fatalf("mutation observe error = %v", err)
			}
			if _, err := runtime.Observe(context.Background(), rule.Scope); err != nil {
				t.Fatalf("read-only observe rejected unavailable scope: %v", err)
			}
		})
	}
}

func TestFilterChainOperationValidationAllowsOnlyUnifiedOperations(t *testing.T) {
	validate := validator.New()
	for _, operation := range []string{"init-base", "bind-base", "unbind-base"} {
		request := dto.FilterChainOperation{Name: constant.FirewallBasicChain, Operate: operation}
		if err := validate.Struct(request); err != nil {
			t.Fatalf("operation %q rejected by API contract: %v", operation, err)
		}
	}
	for _, operation := range []string{"", "init-ipv6-base", "init-forward", "repair-anything"} {
		request := dto.FilterChainOperation{Name: constant.FirewallBasicChain, Operate: operation}
		if err := validate.Struct(request); err == nil {
			t.Fatalf("operation %q accepted by API contract", operation)
		}
	}
}

func TestLoadSystemFirewallFamilyInfoExcludesServiceBackends(t *testing.T) {
	for _, provider := range []string{constant.FirewallProviderFirewalld, constant.FirewallProviderUFW, "unsupported"} {
		status := loadSystemFirewallFamilyInfo(provider, constant.FirewallFamilyIPv6)
		if status.Available || status.Initialized || status.Bound {
			t.Fatalf("%s IPv6 status = %#v, want unavailable", provider, status)
		}
	}
}

func TestSupportsManagedFilterChains(t *testing.T) {
	for _, provider := range []string{constant.FirewallProviderIptables, constant.FirewallProviderNftables} {
		if !supportsManagedFilterChains(provider) {
			t.Fatalf("%s should support managed filter chains", provider)
		}
	}
	for _, provider := range []string{constant.FirewallProviderFirewalld, constant.FirewallProviderUFW} {
		if supportsManagedFilterChains(provider) {
			t.Fatalf("%s should not support managed filter chains", provider)
		}
	}
}

func TestFirewallRuleServiceChecksManagedUpdateWithoutApplying(t *testing.T) {
	rule := executorTestRule("8080")
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	service, ruleRepo := newTestFirewallExecutor(t, adapter)
	if err := createExecutorRule(service, adapter, dto.FirewallRuleCreateItem{
		Rule: rule, SourceKind: constant.FirewallRuleSourceUser,
	}); err != nil {
		t.Fatalf("create managed rule: %v", err)
	}
	stored, err := ruleRepo.List(context.Background())
	if err != nil || len(stored) != 1 {
		t.Fatalf("load managed rule: rules=%#v err=%v", stored, err)
	}
	updated := rule
	updated.DestinationPort = "8443"
	result, err := service.checkRule(context.Background(), "", dto.FirewallRuleCheckItem{
		UUID: stored[0].UUID,
		Rule: updated,
	})
	if err != nil {
		t.Fatalf("check managed update: %v", err)
	}
	if result.Decision != filter.CheckDecisionReady || result.Reason != "update_ready" || result.RequestedRule.DestinationPort != "8443" {
		t.Fatalf("unexpected managed update check: %#v", result)
	}
	if adapter.applyCount != 1 {
		t.Fatalf("update check changed the firewall: applyCount=%d", adapter.applyCount)
	}
}

func TestFirewallExecutorUpdatesManagedRuleAndPositionTogether(t *testing.T) {
	first := executorTestRule("8080")
	first.UUID = "first"
	second := executorTestRule("8081")
	second.UUID = "second"
	adapter := newFakeFilterAdapter(t, first.Scope, []filter.ObservedRule{
		executorObservedRule(first, "1panel-rule:first", 1),
		executorObservedRule(second, "1panel-rule:second", 2),
	})
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)
	for _, rule := range []filter.FirewallRule{first, second} {
		record, err := firewallRuleModelForCreate(rule, dto.FirewallRuleCreateItem{SourceKind: constant.FirewallRuleSourceUser}, constant.FirewallRuleOriginCreated)
		if err != nil {
			t.Fatalf("build managed rule: %v", err)
		}
		record.UUID = rule.UUID
		if err := ruleRepo.Create(context.Background(), &record); err != nil {
			t.Fatalf("create managed record: %v", err)
		}
	}
	target := int64(2)
	updated := first
	updated.DestinationPort = "8443"
	updated.OrderIndex = &target
	if err := executor.updateRule(context.Background(), "", first.UUID, updated); err != nil {
		t.Fatalf("positioned update: %v", err)
	}
	if adapter.snapshot.Rules[1].Marker != "1panel-rule:first" || adapter.snapshot.Rules[1].Rule.DestinationPort != "8443" {
		t.Fatalf("rule content and position were not updated together: %#v", adapter.snapshot.Rules)
	}
	stored, err := ruleRepo.GetByUUID(context.Background(), first.UUID)
	if err != nil || stored.DestinationPort != "8443" {
		t.Fatalf("request-only position was persisted: stored=%#v err=%v", stored, err)
	}
}

func TestFirewallExecutorReordersManagedRule(t *testing.T) {
	scope := executorTestRule("8080").Scope
	first := executorTestRule("8080")
	first.UUID = "first"
	second := executorTestRule("8081")
	second.UUID = "second"
	observed := []filter.ObservedRule{
		executorObservedRule(first, "1panel-rule:first", 1),
		executorObservedRule(second, "1panel-rule:second", 2),
	}
	adapter := newFakeFilterAdapter(t, scope, observed)
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)
	for _, rule := range []filter.FirewallRule{first, second} {
		record, err := firewallRuleModelForCreate(rule, dto.FirewallRuleCreateItem{SourceKind: constant.FirewallRuleSourceUser}, constant.FirewallRuleOriginCreated)
		if err != nil {
			t.Fatalf("build managed rule: %v", err)
		}
		record.UUID = rule.UUID
		if err := ruleRepo.Create(context.Background(), &record); err != nil {
			t.Fatalf("create managed record: %v", err)
		}
	}
	target := int64(2)
	err := executor.reorderRule(context.Background(), "", "first", &target, nil)
	if err != nil || adapter.applyCount != 1 || adapter.snapshot.Rules[1].Marker != "1panel-rule:first" {
		t.Fatalf("managed reorder failed: applyCount=%d snapshot=%#v err=%v", adapter.applyCount, adapter.snapshot.Rules, err)
	}
	firstRecord, err := ruleRepo.GetByUUID(context.Background(), "first")
	if err != nil || firstRecord.Sequence == nil || *firstRecord.Sequence != 2*model.FirewallRuleSequenceStep {
		t.Fatalf("reordered rule sequence was not persisted: record=%#v err=%v", firstRecord, err)
	}
	secondRecord, err := ruleRepo.GetByUUID(context.Background(), "second")
	if err != nil || secondRecord.Sequence == nil || *secondRecord.Sequence != model.FirewallRuleSequenceStep {
		t.Fatalf("neighbor sequence was not rebalanced: record=%#v err=%v", secondRecord, err)
	}
}

func TestFirewallExecutorReorderUsesSparseSequenceMidpoint(t *testing.T) {
	first := executorTestRule("8080")
	first.UUID = "first"
	second := executorTestRule("8081")
	second.UUID = "second"
	third := executorTestRule("8082")
	third.UUID = "third"
	adapter := newFakeFilterAdapter(t, first.Scope, []filter.ObservedRule{
		executorObservedRule(first, "1panel-rule:first", 1),
		executorObservedRule(second, "1panel-rule:second", 2),
		executorObservedRule(third, "1panel-rule:third", 3),
	})
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)
	for index, rule := range []filter.FirewallRule{first, second, third} {
		record, err := firewallRuleModelForCreate(rule, dto.FirewallRuleCreateItem{SourceKind: constant.FirewallRuleSourceUser}, constant.FirewallRuleOriginCreated)
		if err != nil {
			t.Fatalf("build managed rule: %v", err)
		}
		record.UUID = rule.UUID
		sequence := int64(index+1) * model.FirewallRuleSequenceStep
		record.Sequence = &sequence
		if err := ruleRepo.Create(context.Background(), &record); err != nil {
			t.Fatalf("create managed record: %v", err)
		}
	}

	target := int64(2)
	if err := executor.reorderRule(context.Background(), "", third.UUID, &target, nil); err != nil {
		t.Fatalf("reorder managed rule: %v", err)
	}
	want := map[string]int64{
		first.UUID:  model.FirewallRuleSequenceStep,
		third.UUID:  model.FirewallRuleSequenceStep + model.FirewallRuleSequenceStep/2,
		second.UUID: 2 * model.FirewallRuleSequenceStep,
	}
	for uuid, wantSequence := range want {
		record, err := ruleRepo.GetByUUID(context.Background(), uuid)
		if err != nil || record.Sequence == nil || *record.Sequence != wantSequence {
			t.Fatalf("sequence for %s = %#v, want %d (err=%v)", uuid, record.Sequence, wantSequence, err)
		}
	}
}

func TestFirewallExecutorUsesExplicitPositionDuringUpdate(t *testing.T) {
	first := executorTestRule("8080")
	first.UUID = "first"
	second := executorTestRule("8081")
	second.UUID = "second"
	adapter := newFakeFilterAdapter(t, first.Scope, []filter.ObservedRule{
		executorObservedRule(first, "1panel-rule:first", 1),
		executorObservedRule(second, "1panel-rule:second", 2),
	})
	adapter.capabilities = filter.Capabilities{
		Scopes: filter.MVPScopePatterns(), Marker: true, ExplicitPosition: true,
	}
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)
	for _, rule := range []filter.FirewallRule{first, second} {
		record, err := firewallRuleModelForCreate(rule, dto.FirewallRuleCreateItem{SourceKind: constant.FirewallRuleSourceUser}, constant.FirewallRuleOriginCreated)
		if err != nil {
			t.Fatalf("build managed rule: %v", err)
		}
		record.UUID = rule.UUID
		if err := ruleRepo.Create(context.Background(), &record); err != nil {
			t.Fatalf("create managed record: %v", err)
		}
	}
	target := int64(2)
	updated := first
	updated.DestinationPort = "8443"
	updated.OrderIndex = &target
	if err := executor.updateRule(context.Background(), "", first.UUID, updated); err != nil {
		t.Fatalf("explicit-position update: %v", err)
	}
	if adapter.snapshot.Rules[1].Marker != "1panel-rule:first" || adapter.snapshot.Rules[1].Rule.DestinationPort != "8443" {
		t.Fatalf("rule was not updated at requested position: snapshot=%#v", adapter.snapshot)
	}
}

func TestFirewallExecutorRejectsPriorityReorderWithoutExplicitPriority(t *testing.T) {
	rule := filter.FirewallRule{
		UUID: "zone-port", Scope: filter.Scope{Provider: filter.ProviderFirewalld, Family: filter.FamilyInet, Zone: "public", Direction: filter.DirectionInput},
		NativeKind: filter.NativeKindZonePort, Protocol: "tcp", DestinationPort: "443", Action: filter.ActionAccept,
	}
	adapter := newFakeFilterAdapter(t, rule.Scope, []filter.ObservedRule{executorObservedRule(rule, "1panel-rule:zone-port", 1)})
	adapter.capabilities = filter.Capabilities{Scopes: filter.MVPScopePatterns(), Marker: true, ExplicitPriority: true}
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)
	record, err := firewallRuleModelForCreate(rule, dto.FirewallRuleCreateItem{SourceKind: constant.FirewallRuleSourceUser}, constant.FirewallRuleOriginCreated)
	if err != nil {
		t.Fatalf("build zone port record: %v", err)
	}
	record.UUID = rule.UUID
	if err := ruleRepo.Create(context.Background(), &record); err != nil {
		t.Fatalf("create zone port record: %v", err)
	}
	priority := -10
	err = executor.reorderRule(context.Background(), "", rule.UUID, nil, &priority)
	if !errors.Is(err, filter.ErrUnsupportedScope) || adapter.applyCount != 0 {
		t.Fatalf("native zone port was converted by reorder: applyCount=%d err=%v", adapter.applyCount, err)
	}
}

func TestMergeFirewallInventoryMapsPersistenceOwnershipAndUsage(t *testing.T) {
	domainRule := filter.FirewallRule{
		Scope:            filter.Scope{Provider: filter.ProviderIptables, Family: filter.FamilyIPv4, Table: "filter", Chain: "1PANEL_BASIC", Direction: filter.DirectionInput},
		NativeKind:       filter.NativeKindRule,
		Protocol:         "tcp",
		DestinationPort:  "22",
		ConnectionStates: []string{"established", "new"},
		Action:           filter.ActionAccept,
		Description:      "ssh",
	}
	position := 1
	marker := "1panel-rule:managed"
	stored, err := model.FirewallRuleFromDomain(domainRule)
	if err != nil {
		t.Fatalf("build stored rule: %v", err)
	}
	stored.UUID = "managed"
	stored.Origin = constant.FirewallRuleOriginCreated
	stored.Owner = constant.FirewallRuleSourceUser
	observed := filter.ObservedRule{
		Rule:    domainRule,
		Locator: filter.Locator{Provider: filter.ProviderIptables, ScopeKey: domainRule.Scope.Key(), Position: &position},
		Marker:  marker, ParseStatus: filter.ParseStatusSupported,
	}
	usage := map[string]filter.RuntimeUsage{filter.RuntimeUsageKey(domainRule): {UsedBy: []string{"sshd"}}}

	items, err := mergeFirewallInventory([]filter.ObservedRule{observed}, []model.FirewallRule{stored}, nil, usage)
	if err != nil {
		t.Fatalf("merge inventory: %v", err)
	}
	if len(items) != 1 || items[0].State != filter.InventoryStateManaged || items[0].Desired == nil || items[0].Usage == nil || !items[0].Usage.Used {
		t.Fatalf("unexpected merged inventory: %#v", items)
	}
	if items[0].Desired.Rule.ConnectionStates[0] != "established" {
		t.Fatalf("persistence metadata was not normalized: %#v", items[0].Desired.Rule)
	}
}

func TestProtectFirewallSnapshotMarksConfiguredAndRequiredPorts(t *testing.T) {
	scope := filter.Scope{Provider: filter.ProviderIptables, Family: filter.FamilyIPv4, Table: "filter", Chain: "1PANEL_BASIC_BEFORE", Direction: filter.DirectionInput}
	positionOne, positionTwo := 1, 2
	rules := []filter.ObservedRule{
		{
			Rule:    filter.FirewallRule{Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: "22", Action: filter.ActionAccept},
			Locator: filter.Locator{Provider: filter.ProviderIptables, ScopeKey: scope.Key(), Position: &positionOne}, ParseStatus: filter.ParseStatusSupported,
		},
		{
			Rule:    filter.FirewallRule{Scope: scope, NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: "80", Action: filter.ActionAccept},
			Locator: filter.Locator{Provider: filter.ProviderIptables, ScopeKey: scope.Key(), Position: &positionTwo}, ParseStatus: filter.ParseStatusSupported,
		},
	}
	snapshot, err := filter.NewSnapshot(scope, rules)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	snapshot.Notices = []filter.ScopeNotice{{Code: filter.ScopeNoticeManagedScopeInactive}}

	protected, err := filter.ProtectSnapshot(snapshot, []firewall.PortWhitelist{{Port: "22", Protocol: "tcp"}})
	if err != nil {
		t.Fatalf("protect snapshot: %v", err)
	}
	if !protected.Rules[0].Protected || protected.Rules[1].Protected {
		t.Fatalf("unexpected protected ports: %#v", protected.Rules)
	}
	if protected.Revision != snapshot.Revision {
		t.Fatal("runtime safety classification changed the provider-state revision")
	}
	if len(protected.Notices) != 1 || protected.Notices[0].Code != filter.ScopeNoticeManagedScopeInactive {
		t.Fatalf("snapshot notices were lost: %#v", protected.Notices)
	}
}

func newTestFirewallExecutor(t *testing.T, adapter *fakeFilterAdapter) (*FirewallService, *repo.FirewallRuleRepo) {
	t.Helper()
	db := newFirewallRuleTestDB(t)
	ruleRepo := repo.NewFirewallRuleRepo(db)
	return &FirewallService{
		rules: ruleRepo,
		adapters: firewallRuleRuntimeRegistry{
			adapter.Provider(): newFirewallRuleRuntime(adapter, nil),
		},
	}, ruleRepo
}

func createExecutorRule(executor *FirewallService, adapter *fakeFilterAdapter, request dto.FirewallRuleCreateItem) error {
	authorization := firewallRuleCreateAuthorization{Operation: filter.ChangeCreate}
	if request.AdoptInstanceKey != "" {
		candidate, err := filter.FindCandidate(adapter.snapshot.Rules, request.AdoptInstanceKey)
		if err != nil {
			return err
		}
		locator := candidate.Locator
		authorization = firewallRuleCreateAuthorization{Operation: filter.ChangeAdopt, Locator: &locator}
	}
	return executor.createRule(
		context.Background(), newFirewallRuleRuntime(adapter, nil), adapter.snapshot, request, authorization,
	)
}

type fakeFilterAdapter struct {
	snapshot              filter.Snapshot
	multiSnapshots        []filter.Snapshot
	rollbackSnapshot      filter.Snapshot
	lastChange            filter.DesiredChange
	applyCount            int
	observeCount          int
	observeScopesCount    int
	rollbackCount         int
	verifyMatched         bool
	capabilities          filter.Capabilities
	nativeDetail          string
	nativeDetailName      string
	nativeDetailPermanent bool
}

type failingFirewallRuleRepo struct {
	repo.IFirewallRuleRepo
	updateErr error
	deleteErr error
}

func (r *failingFirewallRuleRepo) UpdateWithRevision(ctx context.Context, ruleUUID string, revision uint, updates map[string]interface{}) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	return r.IFirewallRuleRepo.UpdateWithRevision(ctx, ruleUUID, revision, updates)
}

func (r *failingFirewallRuleRepo) DeleteWithRevision(ctx context.Context, ruleUUID string, revision uint) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	return r.IFirewallRuleRepo.DeleteWithRevision(ctx, ruleUUID, revision)
}

func newFakeFilterAdapter(t *testing.T, scope filter.Scope, rules []filter.ObservedRule) *fakeFilterAdapter {
	t.Helper()
	snapshot, err := filter.NewSnapshot(scope, rules)
	if err != nil {
		t.Fatalf("new fake snapshot: %v", err)
	}
	return &fakeFilterAdapter{snapshot: snapshot, verifyMatched: true}
}

func (f *fakeFilterAdapter) Provider() filter.Provider { return f.snapshot.Scope.Provider }
func (f *fakeFilterAdapter) Capabilities(context.Context) (filter.Capabilities, error) {
	if f.capabilities.Scopes != nil {
		return f.capabilities, nil
	}
	return filter.Capabilities{Scopes: filter.MVPScopePatterns(), Marker: true, OwnedChains: true}, nil
}
func (f *fakeFilterAdapter) Observe(context.Context, filter.Scope) (filter.Snapshot, error) {
	f.observeCount++
	return f.snapshot, nil
}
func (f *fakeFilterAdapter) ObserveScopes(context.Context, []filter.Scope) ([]filter.Snapshot, error) {
	f.observeScopesCount++
	return append([]filter.Snapshot(nil), f.multiSnapshots...), nil
}
func (f *fakeFilterAdapter) Compile(snapshot filter.Snapshot, changes []filter.DesiredChange) (filter.BackendPlan, error) {
	f.rollbackSnapshot = snapshot
	f.rollbackSnapshot.Rules = append([]filter.ObservedRule(nil), snapshot.Rules...)
	if len(changes) > 1 {
		plan := filter.BackendPlan{
			Provider: f.Provider(), Scope: snapshot.Scope, SnapshotRevision: snapshot.Revision,
			Rules: make([]filter.NativeRulePlan, 0, len(changes)),
		}
		position := len(snapshot.Rules)
		for _, change := range changes {
			if change.Operation != filter.ChangeCreate && change.Operation != filter.ChangeDelete {
				return filter.BackendPlan{}, filter.ErrInvalidRule
			}
			rule := change.After
			if change.Operation == filter.ChangeDelete {
				rule = change.Before
			}
			if rule == nil {
				return filter.BackendPlan{}, filter.ErrInvalidRule
			}
			position++
			if change.Locator != nil && change.Locator.Position != nil {
				position = *change.Locator.Position
			}
			marker := "1panel-rule:" + rule.UUID
			expected := executorObservedRule(*rule, marker, position)
			plan.Rules = append(plan.Rules, filter.NativeRulePlan{
				RuleUUID: rule.UUID, Operation: change.Operation, Expected: expected,
			})
		}
		f.lastChange = changes[len(changes)-1]
		return plan, nil
	}
	change := changes[0]
	f.lastChange = change
	rule := change.After
	if change.Operation == filter.ChangeDelete {
		rule = change.Before
	}
	marker := "1panel-rule:" + rule.UUID
	position := len(snapshot.Rules) + 1
	if change.Locator != nil && change.Locator.Position != nil {
		position = *change.Locator.Position
	}
	if rule.OrderIndex != nil && (change.Operation == filter.ChangeCreate || change.Operation == filter.ChangeReorder || change.Operation == filter.ChangeUpdate) {
		position = int(*rule.OrderIndex)
	}
	expected := executorObservedRule(*rule, marker, position)
	return filter.BackendPlan{
		Provider: f.Provider(), Scope: snapshot.Scope, SnapshotRevision: snapshot.Revision,
		Rules: []filter.NativeRulePlan{{RuleUUID: rule.UUID, Operation: change.Operation, Expected: expected}},
	}, nil
}
func (f *fakeFilterAdapter) Apply(_ context.Context, plan filter.BackendPlan) (filter.ApplyResult, error) {
	f.applyCount++
	if len(plan.Rules) > 1 {
		applied := make([]filter.ObservedRule, 0, len(plan.Rules))
		rules := append([]filter.ObservedRule(nil), f.snapshot.Rules...)
		if plan.Rules[0].Operation == filter.ChangeDelete {
			deleted := make(map[string]struct{}, len(plan.Rules))
			for _, rulePlan := range plan.Rules {
				deleted[rulePlan.Expected.Marker] = struct{}{}
			}
			remaining := make([]filter.ObservedRule, 0, len(rules)-len(deleted))
			for _, observed := range rules {
				if _, exists := deleted[observed.Marker]; !exists {
					remaining = append(remaining, observed)
				}
			}
			f.snapshot, _ = filter.NewSnapshot(f.snapshot.Scope, remaining)
			return filter.ApplyResult{}, nil
		}
		for _, rulePlan := range plan.Rules {
			if rulePlan.Operation != filter.ChangeCreate {
				return filter.ApplyResult{}, filter.ErrInvalidRule
			}
			rules = append(rules, rulePlan.Expected)
			applied = append(applied, rulePlan.Expected)
		}
		f.snapshot, _ = filter.NewSnapshot(f.snapshot.Scope, rules)
		return filter.ApplyResult{Applied: applied}, nil
	}
	expected := plan.Rules[0].Expected
	if plan.Rules[0].Operation == filter.ChangeReorder || plan.Rules[0].Operation == filter.ChangeUpdate {
		current := -1
		for index, observed := range f.snapshot.Rules {
			if observed.Marker == expected.Marker {
				current = index
				break
			}
		}
		if current < 0 || expected.Locator.Position == nil {
			return filter.ApplyResult{}, filter.ErrRuleStale
		}
		rules := append([]filter.ObservedRule(nil), f.snapshot.Rules...)
		moving := rules[current]
		rules = append(rules[:current], rules[current+1:]...)
		target := *expected.Locator.Position - 1
		if target > len(rules) {
			target = len(rules)
		}
		rules = append(rules, filter.ObservedRule{})
		copy(rules[target+1:], rules[target:])
		moving.Rule = expected.Rule
		rules[target] = moving
		for index := range rules {
			position := index + 1
			rules[index].Locator.Position = &position
		}
		f.snapshot, _ = filter.NewSnapshot(f.snapshot.Scope, rules)
		return filter.ApplyResult{Applied: []filter.ObservedRule{rules[target]}}, nil
	}
	if plan.Rules[0].Operation == filter.ChangeDelete {
		remaining := make([]filter.ObservedRule, 0, len(f.snapshot.Rules))
		for _, observed := range f.snapshot.Rules {
			if observed.Marker != expected.Marker {
				remaining = append(remaining, observed)
			}
		}
		f.snapshot, _ = filter.NewSnapshot(f.snapshot.Scope, remaining)
		return filter.ApplyResult{}, nil
	}
	replaced := false
	for index := range f.snapshot.Rules {
		if f.snapshot.Rules[index].Locator.Position != nil && expected.Locator.Position != nil && *f.snapshot.Rules[index].Locator.Position == *expected.Locator.Position {
			f.snapshot.Rules[index] = expected
			replaced = true
			break
		}
	}
	if !replaced {
		f.snapshot.Rules = append(f.snapshot.Rules, expected)
	}
	f.snapshot, _ = filter.NewSnapshot(f.snapshot.Scope, f.snapshot.Rules)
	return filter.ApplyResult{Applied: []filter.ObservedRule{expected}}, nil
}
func (f *fakeFilterAdapter) Verify(context.Context, filter.BackendPlan) (filter.VerifyResult, error) {
	return filter.VerifyResult{Snapshot: f.snapshot, Matched: f.verifyMatched}, nil
}
func (f *fakeFilterAdapter) Rollback(context.Context, filter.BackendPlan) error {
	f.rollbackCount++
	f.snapshot = f.rollbackSnapshot
	return nil
}

func (f *fakeFilterAdapter) NativeDetail(_ context.Context, name string, permanent bool) (string, error) {
	f.nativeDetailName = name
	f.nativeDetailPermanent = permanent
	return f.nativeDetail, nil
}

func firewallRuleCheckItems(rules []filter.FirewallRule) []dto.FirewallRuleCheckItem {
	items := make([]dto.FirewallRuleCheckItem, 0, len(rules))
	for _, rule := range rules {
		items = append(items, dto.FirewallRuleCheckItem{Rule: rule})
	}
	return items
}

func executorTestRule(port string) filter.FirewallRule {
	return filter.FirewallRule{
		Scope:      filter.Scope{Provider: filter.ProviderIptables, Family: filter.FamilyIPv4, Table: "filter", Chain: "1PANEL_BASIC", Direction: filter.DirectionInput},
		NativeKind: filter.NativeKindRule, Protocol: "tcp", DestinationPort: port, Action: filter.ActionAccept,
	}
}

func executorTestAddressRule(address string) filter.FirewallRule {
	rule := executorTestRule("")
	rule.Protocol = "all"
	rule.SourceAddress = address
	rule.Action = filter.ActionDrop
	return rule
}

func executorObservedRule(rule filter.FirewallRule, marker string, position int) filter.ObservedRule {
	return filter.ObservedRule{
		Rule: rule, Marker: marker, ParseStatus: filter.ParseStatusSupported,
		Locator: filter.Locator{Provider: rule.Scope.Provider, ScopeKey: rule.Scope.Key(), Position: &position},
	}
}

func newFirewallRuleTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("load sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&model.FirewallRule{}); err != nil {
		t.Fatalf("migrate firewall rule: %v", err)
	}
	return db
}
