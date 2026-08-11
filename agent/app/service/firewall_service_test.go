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
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	filterfirewalld "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/providers/firewalld"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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
	check, err := service.Check(ctx, "", dto.FirewallRuleCheck{Rule: rule})
	if err != nil {
		t.Fatalf("plan adoption: %v", err)
	}
	if check.Classification != filter.CheckClassificationExactExternal {
		t.Fatalf("unexpected check result: %#v", check)
	}
	if check.CheckFlag == "" {
		t.Fatal("check did not return a creation flag")
	}
	err = service.Create(ctx, dto.FirewallRuleCreate{
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
	err = service.Delete(ctx, dto.FirewallRuleDelete{UUID: after.Items[0].Desired.UUID})
	if err != nil {
		t.Fatalf("delete adopted rule: %v", err)
	}
	empty, err := service.Inventory(ctx, dto.FirewallRuleInventory{Scope: rule.Scope})
	if err != nil || len(empty.Items) != 0 {
		t.Fatalf("deleted rule remained in inventory: inventory=%#v err=%v", empty, err)
	}
}

func TestFirewallRuleServiceCreateRequiresCheck(t *testing.T) {
	rule := executorTestRule("8080")
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	service, ruleRepo := newTestFirewallExecutor(t, adapter)
	service.selectedProvider = func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil }

	err := service.Create(context.Background(), dto.FirewallRuleCreate{
		Rule: rule, Action: filter.CheckActionCreate,
	})
	if !errors.Is(err, filter.ErrRuleCheckRequired) {
		t.Fatalf("expected check-required error, got %v", err)
	}
	rules, _ := ruleRepo.List(context.Background())
	if adapter.applyCount != 0 || len(rules) != 0 {
		t.Fatalf("unchecked rule was created: applyCount=%d rules=%#v", adapter.applyCount, rules)
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

	checked, err := service.CheckBatch(ctx, "", dto.FirewallRuleBatchCheck{Rules: rules})
	if err != nil || len(checked.Items) != len(rules) {
		t.Fatalf("batch check: result=%#v err=%v", checked, err)
	}
	request := dto.FirewallRuleBatchCreate{Items: make([]dto.FirewallRuleCreate, 0, len(rules))}
	for _, item := range checked.Items {
		request.Items = append(request.Items, dto.FirewallRuleCreate{
			Rule: item.RequestedRule, CheckFlag: item.CheckFlag, Action: filter.CheckActionCreate,
			SourceKind: constant.FirewallRuleSourceUser,
		})
	}
	created, err := service.CreateBatch(ctx, request)
	if err != nil || created.Succeeded != 2 || created.Failed != 0 {
		t.Fatalf("batch create: result=%#v err=%v", created, err)
	}
	stored, _ := ruleRepo.List(ctx)
	if len(stored) != 2 || len(adapter.snapshot.Rules) != 2 || adapter.applyCount != 2 {
		t.Fatalf("same-scope batch did not commit all rules: stored=%#v snapshot=%#v applies=%d", stored, adapter.snapshot, adapter.applyCount)
	}
}

func TestFirewallRuleServiceCreateRejectsChangedFirewallState(t *testing.T) {
	rule := executorTestRule("8080")
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	service, ruleRepo := newTestFirewallExecutor(t, adapter)
	service.selectedProvider = func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil }
	ctx := context.Background()

	check, err := service.Check(ctx, "", dto.FirewallRuleCheck{Rule: rule})
	if err != nil {
		t.Fatalf("check rule: %v", err)
	}
	other := executorTestRule("9090")
	adapter.snapshot, err = filter.NewSnapshot(rule.Scope, []filter.ObservedRule{executorObservedRule(other, "", 1)})
	if err != nil {
		t.Fatalf("change firewall snapshot: %v", err)
	}
	err = service.Create(ctx, dto.FirewallRuleCreate{
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

	check, err := service.Check(ctx, "", dto.FirewallRuleCheck{Rule: rule})
	if err != nil {
		t.Fatalf("check rule: %v", err)
	}
	changed := check.RequestedRule
	changed.DestinationPort = "8081"
	err = service.Create(ctx, dto.FirewallRuleCreate{
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

	check, err := service.Check(ctx, "", dto.FirewallRuleCheck{Rule: rule})
	if err != nil {
		t.Fatalf("check rule: %v", err)
	}
	other := executorTestRule("9090")
	record, err := firewallRuleModelForCreate(other, dto.FirewallRuleCreate{SourceKind: constant.FirewallRuleSourceUser}, constant.FirewallRuleOriginCreated)
	if err != nil {
		t.Fatalf("build managed record: %v", err)
	}
	if err := ruleRepo.Create(ctx, &record); err != nil {
		t.Fatalf("change managed state: %v", err)
	}
	err = service.Create(ctx, dto.FirewallRuleCreate{
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
	_, err := service.Check(context.Background(), "", dto.FirewallRuleCheck{Rule: executorTestRule("80")})
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

func TestAddFirewallRuntimeUsageMergesOwnersAndReasons(t *testing.T) {
	usage := make(map[string]filter.RuntimeUsage)
	addFirewallRuntimeUsage(usage, "tcp", 8080, "demo", "application")
	addFirewallRuntimeUsage(usage, "tcp", 8080, "server", "listener")
	addFirewallRuntimeUsage(usage, "icmp", 8080, "ignored", "listener")
	value := usage["tcp\x008080"]
	if !value.Used || len(value.UsedBy) != 2 || value.Reason != "application_and_listener" || len(usage) != 1 {
		t.Fatalf("unexpected runtime usage: %#v", usage)
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
	ruleKey, err := filter.RuleKey(rule)
	if err != nil {
		t.Fatal(err)
	}
	userRecord, err := firewallRuleModelForCreate(rule, dto.FirewallRuleCreate{SourceKind: constant.FirewallRuleSourceUser}, constant.FirewallRuleOriginCreated)
	if err != nil {
		t.Fatal(err)
	}
	userRecord.UUID = "user-rule"
	userRecord.RuleKey = ruleKey
	instanceKey, err := filter.InstanceKey(observed)
	if err != nil {
		t.Fatal(err)
	}
	userRecord.MatchKey = firewallRuleMatchKey(observed.Marker, instanceKey)
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
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)
	request := dto.FirewallRuleCreate{
		Rule: rule, SourceKind: constant.FirewallRuleSourceUser,
	}

	if err := createExecutorRule(executor, adapter, request); err != nil {
		t.Fatalf("commit create: %v", err)
	}
	if adapter.applyCount != 1 {
		t.Fatalf("unexpected apply count: %d", adapter.applyCount)
	}
	rules, _ := ruleRepo.List(context.Background())
	if len(rules) != 1 || rules[0].MatchKey != firewallRuleMarkerMatchPrefix+adapter.snapshot.Rules[0].Marker {
		t.Fatalf("rule was not verified and bound: %#v", rules)
	}
}

func TestFirewallExecutorAdoptsWithoutAddingEquivalentRule(t *testing.T) {
	rule := executorTestAddressRule("172.16.10.111")
	external := executorObservedRule(rule, "", 1)
	adapter := newFakeFilterAdapter(t, rule.Scope, []filter.ObservedRule{external})
	instanceKey, _ := filter.InstanceKey(external)
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)
	err := createExecutorRule(executor, adapter, dto.FirewallRuleCreate{
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

	err := createExecutorRule(executor, adapter, dto.FirewallRuleCreate{
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

func TestFirewallExecutorRollsBackCreateWhenPersistenceCommitFails(t *testing.T) {
	rule := executorTestRule("9091")
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)
	executor.rules = &failingFirewallRuleRepo{
		IFirewallRuleRepo: ruleRepo,
		updateErr:         errors.New("commit failed"),
	}

	err := createExecutorRule(executor, adapter, dto.FirewallRuleCreate{
		Rule: rule, SourceKind: constant.FirewallRuleSourceUser,
	})
	if err == nil || !strings.Contains(err.Error(), "commit failed") {
		t.Fatalf("expected persistence commit failure, got %v", err)
	}
	stored, _ := ruleRepo.List(context.Background())
	if len(stored) != 0 || len(adapter.snapshot.Rules) != 0 || adapter.rollbackCount != 1 {
		t.Fatalf("failed create was not fully compensated: stored=%#v snapshot=%#v rollbacks=%d", stored, adapter.snapshot, adapter.rollbackCount)
	}
}

func TestFirewallExecutorRollsBackDeleteWhenPersistenceDeleteFails(t *testing.T) {
	rule := executorTestRule("9092")
	adapter := newFakeFilterAdapter(t, rule.Scope, nil)
	executor, ruleRepo := newTestFirewallExecutor(t, adapter)
	if err := createExecutorRule(executor, adapter, dto.FirewallRuleCreate{
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
	if err := createExecutorRule(executor, adapter, dto.FirewallRuleCreate{
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
	err := createExecutorRule(executor, adapter, dto.FirewallRuleCreate{
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
	err := createExecutorRule(executor, adapter, dto.FirewallRuleCreate{
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
	err := createExecutorRule(executor, adapter, dto.FirewallRuleCreate{
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
		record, err := firewallRuleModelForCreate(rule, dto.FirewallRuleCreate{SourceKind: constant.FirewallRuleSourceUser}, constant.FirewallRuleOriginCreated)
		if err != nil {
			t.Fatalf("build managed rule: %v", err)
		}
		record.UUID = rule.UUID
		marker := "1panel-rule:" + rule.UUID
		record.MatchKey = firewallRuleMatchKey(marker, "")
		if err := ruleRepo.Create(context.Background(), &record); err != nil {
			t.Fatalf("create managed record: %v", err)
		}
	}
	target := int64(2)
	err := executor.reorderRule(context.Background(), "", "first", &target, nil)
	if err != nil {
		t.Fatalf("reorder managed rule: %v", err)
	}
	if adapter.snapshot.Rules[1].Marker != "1panel-rule:first" {
		t.Fatalf("rule was not reordered: snapshot=%#v", adapter.snapshot)
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
	record, err := firewallRuleModelForCreate(rule, dto.FirewallRuleCreate{SourceKind: constant.FirewallRuleSourceUser}, constant.FirewallRuleOriginCreated)
	if err != nil {
		t.Fatalf("build zone port record: %v", err)
	}
	record.UUID = rule.UUID
	marker := "1panel-rule:" + rule.UUID
	record.MatchKey = firewallRuleMatchKey(marker, "")
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
	ruleKey, err := filter.RuleKey(domainRule)
	if err != nil {
		t.Fatalf("rule key: %v", err)
	}
	position := 1
	marker := "onepanel:created:ssh"
	stored, err := firewallRuleSemanticModel(domainRule)
	if err != nil {
		t.Fatalf("build stored rule: %v", err)
	}
	stored.UUID = "managed"
	stored.RuleKey = ruleKey
	stored.Origin = constant.FirewallRuleOriginCreated
	stored.Owner = constant.FirewallRuleSourceUser
	stored.MatchKey = firewallRuleMatchKey(marker, "")
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

func TestFirewallRuleMatchKeyRoundTrip(t *testing.T) {
	tests := []struct {
		marker      string
		instanceKey string
	}{
		{marker: "1panel-rule:ssh", instanceKey: "instance-fallback"},
		{instanceKey: "firewalld-instance"},
	}
	for _, test := range tests {
		marker, instanceKey := firewallRuleMatchValues(firewallRuleMatchKey(test.marker, test.instanceKey))
		if marker != test.marker {
			t.Fatalf("unexpected marker: want=%q got=%q", test.marker, marker)
		}
		if test.marker == "" && instanceKey != test.instanceKey {
			t.Fatalf("unexpected instance key: want=%q got=%q", test.instanceKey, instanceKey)
		}
		if test.marker != "" && instanceKey != "" {
			t.Fatalf("marker match retained fallback instance key: %q", instanceKey)
		}
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

	protected, err := filter.ProtectSnapshot(snapshot, []filter.ProtectedPort{{Port: "22", Protocol: "tcp"}})
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

func createExecutorRule(executor *FirewallService, adapter *fakeFilterAdapter, request dto.FirewallRuleCreate) error {
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
	snapshot         filter.Snapshot
	rollbackSnapshot filter.Snapshot
	applyCount       int
	rollbackCount    int
	verifyMatched    bool
	capabilities     filter.Capabilities
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
	return f.snapshot, nil
}
func (f *fakeFilterAdapter) Compile(snapshot filter.Snapshot, changes []filter.DesiredChange) (filter.BackendPlan, error) {
	f.rollbackSnapshot = snapshot
	f.rollbackSnapshot.Rules = append([]filter.ObservedRule(nil), snapshot.Rules...)
	change := changes[0]
	rule := change.After
	if change.Operation == filter.ChangeDelete {
		rule = change.Before
	}
	marker := "1panel-rule:" + rule.UUID
	position := len(snapshot.Rules) + 1
	if change.Locator != nil && change.Locator.Position != nil {
		position = *change.Locator.Position
	}
	if change.Operation == filter.ChangeReorder && rule.OrderIndex != nil {
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
	expected := plan.Rules[0].Expected
	if plan.Rules[0].Operation == filter.ChangeReorder {
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
