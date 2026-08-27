package service

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	agenti18n "github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/glebarez/sqlite"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type fakeFirewallDatabaseSyncAdapter struct {
	previewResult dto.FirewallRuleSyncPreview
	syncResult    dto.FirewallRuleSyncResult
	previewErr    error
	syncErr       error
	previewCalls  int
	syncCalls     int
}

func (f *fakeFirewallDatabaseSyncAdapter) previewRuleSync(
	context.Context,
	dto.FirewallRuleSyncRequest,
) (dto.FirewallRuleSyncPreview, error) {
	f.previewCalls++
	return f.previewResult, f.previewErr
}

func (f *fakeFirewallDatabaseSyncAdapter) syncRules(
	context.Context,
	dto.FirewallRuleSyncRequest,
) (dto.FirewallRuleSyncResult, error) {
	f.syncCalls++
	return f.syncResult, f.syncErr
}

func TestFirewallRuleSyncCoordinatorDispatchesSubsystemAdapters(t *testing.T) {
	forwardingErr := errors.New("forwarding preview")
	dockerErr := errors.New("docker sync")
	forwarding := &fakeFirewallDatabaseSyncAdapter{previewErr: forwardingErr}
	docker := &fakeFirewallDatabaseSyncAdapter{syncErr: dockerErr}
	service := &FirewallService{forwardingSync: forwarding, dockerSync: docker}

	_, err := service.PreviewRuleSync(context.Background(), "client-ip", dto.FirewallRuleSyncRequest{
		Subsystem: "forwarding", TargetProvider: filter.ProviderNftables,
	})
	if !errors.Is(err, forwardingErr) {
		t.Fatalf("forwarding preview was not dispatched to its adapter: %v", err)
	}
	_, err = service.SyncRules(context.Background(), "client-ip", dto.FirewallRuleSyncRequest{
		Subsystem: "docker", TargetProvider: filter.ProviderNftables,
	})
	if !errors.Is(err, dockerErr) {
		t.Fatalf("Docker synchronization was not dispatched to its adapter: %v", err)
	}
	if forwarding.previewCalls != 1 || forwarding.syncCalls != 0 || docker.previewCalls != 0 || docker.syncCalls != 1 {
		t.Fatalf("unexpected adapter calls: forwarding=%#v docker=%#v", forwarding, docker)
	}
}

func TestFirewallRuleSyncRequestValidationAllowsDatabaseSource(t *testing.T) {
	validate := validator.New()
	for _, subsystem := range []string{"forwarding", "docker"} {
		request := dto.FirewallRuleSyncRequest{Subsystem: subsystem, TargetProvider: filter.ProviderNftables}
		if err := validate.Struct(request); err != nil {
			t.Fatalf("%s database synchronization rejected missing source provider: %v", subsystem, err)
		}
	}
	if err := validate.Struct(dto.FirewallRuleSyncRequest{Subsystem: "system", SourceProvider: filter.ProviderIptables}); err == nil {
		t.Fatal("synchronization request without target provider was accepted")
	}
}

func TestFirewallRuleSyncFailureMessagesIncludeDetailsAndGroupDuplicates(t *testing.T) {
	messages := firewallRuleSyncFailureMessages([]dto.FirewallRuleSyncFailure{
		{SourceUUID: "rule-1", Error: "iptables-restore failed: invalid port"},
		{SourceUUID: "rule-2", Error: "iptables-restore failed: invalid port"},
		{SourceUUID: "rule-3", Error: "permission denied"},
	})
	want := []string{
		"UUID [rule-1, rule-2]: iptables-restore failed: invalid port",
		"UUID [rule-3]: permission denied",
	}
	if len(messages) != len(want) {
		t.Fatalf("failure messages = %#v, want %#v", messages, want)
	}
	for index := range want {
		if messages[index] != want[index] {
			t.Fatalf("failure message %d = %q, want %q", index, messages[index], want[index])
		}
	}
}

func TestFirewallRuleSyncPreviewApplyAndRetry(t *testing.T) {
	ctx := context.Background()
	db := newFirewallRuleTestDB(t)
	ruleRepo := repo.NewFirewallRuleRepo(db)
	sourceRule := filter.FirewallRule{
		Scope: filter.Scope{
			Provider: filter.ProviderIptables, Family: filter.FamilyIPv4,
			Table: "filter", Chain: filter.IptablesInputChain, Direction: filter.DirectionInput,
		},
		Protocol: "tcp", DestinationPort: "8443", Action: filter.ActionAccept, Description: "sync-test",
	}
	sourceRecord, err := model.FirewallRuleFromDomain(sourceRule)
	if err != nil {
		t.Fatalf("encode source rule: %v", err)
	}
	sourceRecord.Origin = constant.FirewallRuleOriginCreated
	sourceRecord.Owner = model.FirewallRuleOwner(constant.FirewallRuleSourceApp, "test-app")
	if err := ruleRepo.Create(ctx, &sourceRecord); err != nil {
		t.Fatalf("persist source rule: %v", err)
	}

	targetScope := filter.Scope{
		Provider: filter.ProviderNftables, Family: filter.FamilyIPv4,
		Table: "filter", Chain: filter.IptablesInputChain, Direction: filter.DirectionInput,
	}
	adapter := newFakeFilterAdapter(t, targetScope, nil)
	service := &FirewallService{
		rules: ruleRepo,
		adapters: firewallRuleRuntimeRegistry{
			filter.ProviderNftables: newFirewallRuleRuntime(adapter, nil),
		},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderNftables, nil },
	}
	request := dto.FirewallRuleSyncRequest{TargetProvider: filter.ProviderNftables}

	preview, err := service.PreviewRuleSync(ctx, "", request)
	if err != nil {
		t.Fatalf("preview rule sync: %v", err)
	}
	if preview.Total != 1 || preview.Ready != 1 || preview.Existing != 0 || preview.Blocked != 0 {
		t.Fatalf("unexpected preview: %#v", preview)
	}

	result, err := service.syncRules(ctx, "", request)
	if err != nil {
		t.Fatalf("apply rule sync: %v", err)
	}
	if result.Total != 1 || result.Succeeded != 1 || result.Skipped != 0 || result.Failed != 0 {
		t.Fatalf("unexpected sync result: %#v", result)
	}
	targetRecords, err := ruleRepo.List(ctx)
	if err != nil {
		t.Fatalf("list target records: %v", err)
	}
	if len(targetRecords) != 1 || targetRecords[0].Owner != sourceRecord.Owner {
		t.Fatalf("target ownership was not preserved: %#v", targetRecords)
	}

	retry, err := service.syncRules(ctx, "", request)
	if err != nil {
		t.Fatalf("retry rule sync: %v", err)
	}
	if retry.Succeeded != 0 || retry.Skipped != 1 || retry.Failed != 0 {
		t.Fatalf("rule sync is not idempotent: %#v", retry)
	}

	setupFirewallTaskTestDB(t)
	taskResult, err := service.SyncRules(ctx, "", dto.FirewallRuleSyncRequest{
		Subsystem: "system", TargetProvider: filter.ProviderNftables,
		TaskID: "firewall-sync-task",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !taskResult.Queued || taskResult.TaskID != "firewall-sync-task" {
		t.Fatalf("plain synchronization was not queued as a task: %#v", taskResult)
	}
	waitFirewallSyncTask(t, taskResult.TaskID)
	sourceRecords, err := ruleRepo.List(ctx)
	if err != nil || len(sourceRecords) != 1 {
		t.Fatalf("plain synchronization reset the source backend: %#v err=%v", sourceRecords, err)
	}
}

func TestFirewallRuleSyncKeepsExpandedRulesInTheSameScope(t *testing.T) {
	ctx := context.Background()
	ruleRepo := repo.NewFirewallRuleRepo(newFirewallRuleTestDB(t))
	sourceRule := filter.FirewallRule{
		Scope: filter.Scope{
			Provider: filter.ProviderIptables, Family: filter.FamilyIPv4,
			Table: "filter", Chain: filter.IptablesInputChain, Direction: filter.DirectionInput,
		},
		Protocol: "tcp", DestinationPort: "80,443", Action: filter.ActionAccept,
	}
	sourceRecord, err := model.FirewallRuleFromDomain(sourceRule)
	if err != nil {
		t.Fatal(err)
	}
	sourceRecord.Origin = constant.FirewallRuleOriginCreated
	sourceRecord.Owner = constant.FirewallRuleSourceUser
	if err := ruleRepo.Create(ctx, &sourceRecord); err != nil {
		t.Fatal(err)
	}

	targetScope := filter.Scope{
		Provider: filter.ProviderNftables, Family: filter.FamilyIPv4,
		Table: "filter", Chain: filter.IptablesInputChain, Direction: filter.DirectionInput,
	}
	adapter := newFakeFilterAdapter(t, targetScope, nil)
	service := &FirewallService{
		rules: ruleRepo,
		adapters: firewallRuleRuntimeRegistry{
			filter.ProviderNftables: newFirewallRuleRuntime(adapter, nil),
		},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderNftables, nil },
	}
	request := dto.FirewallRuleSyncRequest{TargetProvider: filter.ProviderNftables}

	result, err := service.syncRules(ctx, "", request)
	if err != nil || result.Succeeded != 2 || result.Failed != 0 {
		t.Fatalf("expanded rule synchronization failed: result=%#v err=%v", result, err)
	}
	if len(adapter.snapshot.Rules) != 2 {
		t.Fatalf("expanded rules overwrote each other: %#v", adapter.snapshot.Rules)
	}
	if adapter.applyCount != 1 {
		t.Fatalf("same-scope rules were applied in %d calls, want one batch", adapter.applyCount)
	}
	if adapter.observeCount != len(firewallRuleSyncScopes(filter.ProviderNftables)) {
		t.Fatalf("synchronization observed scopes %d times, want one read per scope", adapter.observeCount)
	}
	ports := make(map[string]struct{}, len(adapter.snapshot.Rules))
	markers := make(map[string]struct{}, len(adapter.snapshot.Rules))
	for _, observed := range adapter.snapshot.Rules {
		ports[observed.Rule.DestinationPort] = struct{}{}
		markers[observed.Marker] = struct{}{}
	}
	if _, exists := ports["80"]; !exists {
		t.Fatalf("expanded port 80 is missing: %#v", adapter.snapshot.Rules)
	}
	if _, exists := ports["443"]; !exists {
		t.Fatalf("expanded port 443 is missing: %#v", adapter.snapshot.Rules)
	}
	if len(markers) != 2 {
		t.Fatalf("expanded rules reused one runtime marker: %#v", adapter.snapshot.Rules)
	}

	retry, err := service.syncRules(ctx, "", request)
	if err != nil || retry.Succeeded != 0 || retry.Skipped != 2 || retry.Failed != 0 {
		t.Fatalf("expanded rule synchronization is not idempotent: result=%#v err=%v", retry, err)
	}
}

func TestFirewallRuleSyncRestoresManagedOrderFromDatabase(t *testing.T) {
	ctx := context.Background()
	ruleRepo := repo.NewFirewallRuleRepo(newFirewallRuleTestDB(t))
	targetScope := filter.Scope{
		Provider: filter.ProviderNftables, Family: filter.FamilyIPv4,
		Table: "filter", Chain: filter.IptablesInputChain, Direction: filter.DirectionInput,
	}
	adapter := newFakeFilterAdapter(t, targetScope, nil)
	service := &FirewallService{
		rules: ruleRepo,
		adapters: firewallRuleRuntimeRegistry{
			filter.ProviderNftables: newFirewallRuleRuntime(adapter, nil),
		},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderNftables, nil },
	}

	records := make([]model.FirewallRule, 0, 2)
	for index, port := range []string{"8080", "8081"} {
		rule := filter.FirewallRule{
			Scope: filter.Scope{
				Provider: filter.ProviderIptables, Family: filter.FamilyIPv4,
				Table: "filter", Chain: filter.IptablesInputChain, Direction: filter.DirectionInput,
			},
			Protocol: "tcp", DestinationPort: port, Action: filter.ActionAccept,
		}
		record, err := model.FirewallRuleFromDomain(rule)
		if err != nil {
			t.Fatal(err)
		}
		record.UUID = []string{"first", "second"}[index]
		record.Origin = constant.FirewallRuleOriginCreated
		record.Owner = constant.FirewallRuleSourceUser
		sequence := int64(index+1) * model.FirewallRuleSequenceStep
		record.Sequence = &sequence
		if err := ruleRepo.Create(ctx, &record); err != nil {
			t.Fatal(err)
		}
		records = append(records, record)
	}
	compiledFirst, err := service.compileStoredFirewallRules(ctx, records[0], filter.ProviderNftables)
	if err != nil {
		t.Fatal(err)
	}
	compiledSecond, err := service.compileStoredFirewallRules(ctx, records[1], filter.ProviderNftables)
	if err != nil {
		t.Fatal(err)
	}
	observedSecond := executorObservedRule(compiledSecond[0].Rule, compiledSecond[0].Marker, 1)
	observedFirst := executorObservedRule(compiledFirst[0].Rule, compiledFirst[0].Marker, 2)
	// Native inventory identifies managed rules through their marker and does
	// not populate the domain rule UUID.
	observedSecond.Rule.UUID = ""
	observedFirst.Rule.UUID = ""
	adapter.snapshot, err = filter.NewSnapshot(targetScope, []filter.ObservedRule{observedSecond, observedFirst})
	if err != nil {
		t.Fatal(err)
	}

	request := dto.FirewallRuleSyncRequest{TargetProvider: filter.ProviderNftables}
	preview, err := service.PreviewRuleSync(ctx, "", request)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Ready != 2 || preview.Existing != 0 || preview.Blocked != 0 {
		t.Fatalf("order drift was not included in preview: %#v", preview)
	}
	result, err := service.syncRules(ctx, "", request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Succeeded != 2 || result.Failed != 0 || adapter.applyCount != 1 {
		t.Fatalf("managed order was not synchronized: result=%#v applies=%d", result, adapter.applyCount)
	}
	if adapter.snapshot.Rules[0].Marker != compiledFirst[0].Marker || adapter.snapshot.Rules[1].Marker != compiledSecond[0].Marker {
		t.Fatalf("runtime order does not match database order: %#v", adapter.snapshot.Rules)
	}
	for index, uuid := range []string{"first", "second"} {
		stored, loadErr := ruleRepo.GetByUUID(ctx, uuid)
		want := int64(index+1) * model.FirewallRuleSequenceStep
		if loadErr != nil || stored.Sequence == nil || *stored.Sequence != want {
			t.Fatalf("synchronization rewrote database sequence for %s: record=%#v err=%v", uuid, stored, loadErr)
		}
	}
}

func TestFirewallRuleSyncBlocksManagedOrderAcrossExternalRule(t *testing.T) {
	ctx := context.Background()
	ruleRepo := repo.NewFirewallRuleRepo(newFirewallRuleTestDB(t))
	targetScope := filter.Scope{
		Provider: filter.ProviderNftables, Family: filter.FamilyIPv4,
		Table: "filter", Chain: filter.IptablesInputChain, Direction: filter.DirectionInput,
	}
	adapter := newFakeFilterAdapter(t, targetScope, nil)
	service := &FirewallService{
		rules: ruleRepo,
		adapters: firewallRuleRuntimeRegistry{
			filter.ProviderNftables: newFirewallRuleRuntime(adapter, nil),
		},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderNftables, nil },
	}

	compiled := make([]filter.DesiredRule, 0, 2)
	for index, port := range []string{"8080", "8081"} {
		rule := filter.FirewallRule{
			Scope: filter.Scope{
				Provider: filter.ProviderIptables, Family: filter.FamilyIPv4,
				Table: "filter", Chain: filter.IptablesInputChain, Direction: filter.DirectionInput,
			},
			Protocol: "tcp", DestinationPort: port, Action: filter.ActionAccept,
		}
		record, err := model.FirewallRuleFromDomain(rule)
		if err != nil {
			t.Fatal(err)
		}
		record.UUID = []string{"first", "second"}[index]
		record.Origin = constant.FirewallRuleOriginCreated
		record.Owner = constant.FirewallRuleSourceUser
		sequence := int64(index+1) * model.FirewallRuleSequenceStep
		record.Sequence = &sequence
		if err := ruleRepo.Create(ctx, &record); err != nil {
			t.Fatal(err)
		}
		rules, compileErr := service.compileStoredFirewallRules(ctx, record, filter.ProviderNftables)
		if compileErr != nil {
			t.Fatal(compileErr)
		}
		compiled = append(compiled, rules[0])
	}
	external := compiled[0].Rule
	external.UUID = ""
	external.DestinationPort = "9090"
	adapter.snapshot, _ = filter.NewSnapshot(targetScope, []filter.ObservedRule{
		executorObservedRule(compiled[1].Rule, compiled[1].Marker, 1),
		executorObservedRule(external, "", 2),
		executorObservedRule(compiled[0].Rule, compiled[0].Marker, 3),
	})

	request := dto.FirewallRuleSyncRequest{TargetProvider: filter.ProviderNftables}
	preview, err := service.PreviewRuleSync(ctx, "", request)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Blocked != 2 || preview.Ready != 0 {
		t.Fatalf("unsafe order change was not blocked: %#v", preview)
	}
	result, err := service.syncRules(ctx, "", request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Failed != 2 || adapter.applyCount != 0 || adapter.snapshot.Rules[1].Marker != "" {
		t.Fatalf("blocked order change mutated runtime rules: result=%#v applies=%d rules=%#v", result, adapter.applyCount, adapter.snapshot.Rules)
	}
}

func TestFirewallRuleSyncRejectsProviderSourceAndKeepsDatabasePolicy(t *testing.T) {
	ctx := context.Background()
	db := newFirewallRuleTestDB(t)
	ruleRepo := repo.NewFirewallRuleRepo(db)
	sourceRule := filter.FirewallRule{
		Scope: filter.Scope{
			Provider: filter.ProviderIptables, Family: filter.FamilyIPv4,
			Table: "filter", Chain: filter.IptablesInputChain, Direction: filter.DirectionInput,
		},
		Protocol: "tcp", DestinationPort: "8443", Action: filter.ActionAccept,
	}
	sourceRecord, err := model.FirewallRuleFromDomain(sourceRule)
	if err != nil {
		t.Fatal(err)
	}
	sourceRecord.Origin = constant.FirewallRuleOriginCreated
	sourceRecord.Owner = constant.FirewallRuleSourceUser
	if err := ruleRepo.Create(ctx, &sourceRecord); err != nil {
		t.Fatal(err)
	}
	targetScope := filter.Scope{
		Provider: filter.ProviderNftables, Family: filter.FamilyIPv4,
		Table: "filter", Chain: filter.IptablesInputChain, Direction: filter.DirectionInput,
	}
	target := newFakeFilterAdapter(t, targetScope, nil)
	service := &FirewallService{
		rules: ruleRepo,
		adapters: firewallRuleRuntimeRegistry{
			filter.ProviderNftables: newFirewallRuleRuntime(target, nil),
		},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderNftables, nil },
	}
	_, err = service.PreviewRuleSync(ctx, "", dto.FirewallRuleSyncRequest{
		Subsystem: "system", SourceProvider: filter.ProviderIptables, TargetProvider: filter.ProviderNftables,
	})
	if err == nil {
		t.Fatal("system database synchronization accepted a source provider")
	}
	result, err := service.syncRules(ctx, "", dto.FirewallRuleSyncRequest{TargetProvider: filter.ProviderNftables})
	if err != nil || result.Succeeded != 1 {
		t.Fatalf("database synchronization failed: result=%#v err=%v", result, err)
	}
	records, err := ruleRepo.List(ctx)
	if err != nil || len(records) != 1 || records[0].UUID != sourceRecord.UUID {
		t.Fatalf("database policy was duplicated or removed: %#v err=%v", records, err)
	}
}

func TestFirewallRuleSyncRemovesManagedRuntimeRuleMissingFromDatabase(t *testing.T) {
	ctx := context.Background()
	ruleRepo := repo.NewFirewallRuleRepo(newFirewallRuleTestDB(t))
	scope := filter.Scope{
		Provider: filter.ProviderFirewalld, Family: filter.FamilyInet,
		Zone: filter.FirewalldInputZone, Direction: filter.DirectionInput,
	}
	rule := filter.FirewallRule{
		UUID: "orphan", Scope: scope, NativeKind: filter.NativeKindRichRule,
		Protocol: "tcp", DestinationPort: "9443", Action: filter.ActionAccept,
	}
	observed := executorObservedRule(rule, "1panel-rule:orphan", 1)
	observed.Rule.UUID = ""
	adapter := newFakeFilterAdapter(t, scope, []filter.ObservedRule{observed})
	service := &FirewallService{
		rules: ruleRepo,
		adapters: firewallRuleRuntimeRegistry{
			filter.ProviderFirewalld: newFirewallRuleRuntime(adapter, nil),
		},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderFirewalld, nil },
	}
	request := dto.FirewallRuleSyncRequest{TargetProvider: filter.ProviderFirewalld}

	preview, err := service.PreviewRuleSync(ctx, "", request)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Total != 0 || preview.Removed != 1 || len(preview.Items) != 1 || preview.Items[0].Status != "remove" {
		t.Fatalf("unexpected orphan preview: %#v", preview)
	}
	result, err := service.syncRules(ctx, "", request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 0 || result.Removed != 1 || result.Failed != 0 || len(adapter.snapshot.Rules) != 0 {
		t.Fatalf("orphan runtime rule was not removed: result=%#v rules=%#v", result, adapter.snapshot.Rules)
	}
	stored, err := ruleRepo.List(ctx)
	if err != nil || len(stored) != 0 {
		t.Fatalf("runtime cleanup changed database policies: %#v err=%v", stored, err)
	}
}

func TestFirewallRuleSyncCompileFailureDoesNotCreateOrphanRemoval(t *testing.T) {
	ctx := context.Background()
	ruleRepo := repo.NewFirewallRuleRepo(newFirewallRuleTestDB(t))
	scope := filter.Scope{
		Provider: filter.ProviderFirewalld, Family: filter.FamilyInet,
		Zone: filter.FirewalldInputZone, Direction: filter.DirectionInput,
	}
	rule := filter.FirewallRule{
		UUID: "incompatible", Scope: scope, NativeKind: filter.NativeKindRichRule,
		Protocol: "tcp", DestinationPort: "9443", Action: filter.ActionAccept,
	}
	record, err := model.FirewallRuleFromDomain(rule)
	if err != nil {
		t.Fatal(err)
	}
	record.UUID = rule.UUID
	record.Origin = constant.FirewallRuleOriginCreated
	record.Owner = constant.FirewallRuleSourceUser
	record.CompatibilityError = "manual recreation required"
	if err := ruleRepo.Create(ctx, &record); err != nil {
		t.Fatal(err)
	}
	adapter := newFakeFilterAdapter(t, scope, []filter.ObservedRule{
		executorObservedRule(rule, "1panel-rule:"+rule.UUID, 1),
	})
	service := &FirewallService{
		rules: ruleRepo,
		adapters: firewallRuleRuntimeRegistry{
			filter.ProviderFirewalld: newFirewallRuleRuntime(adapter, nil),
		},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderFirewalld, nil },
	}

	preview, err := service.PreviewRuleSync(ctx, "", dto.FirewallRuleSyncRequest{TargetProvider: filter.ProviderFirewalld})
	if err != nil {
		t.Fatal(err)
	}
	if preview.Blocked != 1 || preview.Removed != 0 || len(preview.Items) != 1 {
		t.Fatalf("compile failure classified its runtime rule as an orphan: %#v", preview)
	}
}

func TestFirewallRuleSyncBlockedPlanDoesNotRemoveOrphans(t *testing.T) {
	ctx := context.Background()
	ruleRepo := repo.NewFirewallRuleRepo(newFirewallRuleTestDB(t))
	scope := filter.Scope{
		Provider: filter.ProviderNftables, Family: filter.FamilyIPv4,
		Table: "filter", Chain: filter.IptablesInputChain, Direction: filter.DirectionInput,
	}
	blockedRule := filter.FirewallRule{
		Scope: scope, Protocol: "all", Action: filter.ActionDrop,
	}
	record, err := model.FirewallRuleFromDomain(blockedRule)
	if err != nil {
		t.Fatal(err)
	}
	record.Origin = constant.FirewallRuleOriginCreated
	record.Owner = constant.FirewallRuleSourceUser
	if err := ruleRepo.Create(ctx, &record); err != nil {
		t.Fatal(err)
	}
	orphanRule := filter.FirewallRule{
		UUID: "orphan", Scope: scope, Protocol: "tcp", DestinationPort: "9443", Action: filter.ActionAccept,
	}
	adapter := newFakeFilterAdapter(t, scope, []filter.ObservedRule{
		executorObservedRule(orphanRule, "1panel-rule:orphan", 1),
	})
	service := &FirewallService{
		rules: ruleRepo,
		adapters: firewallRuleRuntimeRegistry{
			filter.ProviderNftables: newFirewallRuleRuntime(adapter, nil),
		},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderNftables, nil },
	}

	result, err := service.syncRules(ctx, "203.0.113.10", dto.FirewallRuleSyncRequest{TargetProvider: filter.ProviderNftables})
	if err != nil {
		t.Fatal(err)
	}
	if result.Failed != 1 || result.Removed != 0 || adapter.applyCount != 0 || len(adapter.snapshot.Rules) != 1 {
		t.Fatalf("blocked synchronization mutated target rules: result=%#v rules=%#v applies=%d", result, adapter.snapshot.Rules, adapter.applyCount)
	}
}

func waitFirewallSyncTask(t *testing.T, taskID string) {
	t.Helper()
	taskRepo := repo.NewITaskRepo()
	deadline := time.Now().Add(5 * time.Second)
	for {
		completed := false
		record, loadErr := taskRepo.GetFirst(taskRepo.WithByID(taskID))
		if loadErr == nil && record.Status != constant.StatusExecuting {
			if record.Status != constant.StatusSuccess {
				t.Fatalf("firewall synchronization task failed: %#v", record)
			}
			completed = true
		}
		firewallRuleSyncTaskMu.Lock()
		activeTaskID := firewallRuleSyncTaskID
		firewallRuleSyncTaskMu.Unlock()
		if completed && activeTaskID == "" {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for firewall synchronization task")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func setupFirewallTaskTestDB(t *testing.T) {
	t.Helper()
	agenti18n.Init()
	previousDB := global.TaskDB
	previousDir := global.Dir.TaskDir
	taskDir := t.TempDir()
	db, err := gorm.Open(sqlite.Open(filepath.Join(taskDir, "task.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Task{}); err != nil {
		t.Fatal(err)
	}
	global.TaskDB = db
	global.Dir.TaskDir = taskDir
	t.Cleanup(func() {
		global.TaskDB = previousDB
		global.Dir.TaskDir = previousDir
	})
}

func TestSortFirewallPoliciesUsesProviderPlacement(t *testing.T) {
	sequenceOne, sequenceTwo := model.FirewallRuleSequenceStep, 2*model.FirewallRuleSequenceStep
	priorityLow, priorityHigh := -100, 100
	policies := []model.FirewallRule{
		{UUID: "high", Priority: &priorityHigh, Sequence: &sequenceOne},
		{UUID: "none"},
		{UUID: "low", Priority: &priorityLow, Sequence: &sequenceTwo},
	}

	positional := append([]model.FirewallRule(nil), policies...)
	sortFirewallPolicies(positional, filter.ProviderUFW)
	if positional[0].UUID != "high" || positional[1].UUID != "low" || positional[2].UUID != "none" {
		t.Fatalf("positional policies were not sorted by sequence: %#v", positional)
	}

	weighted := append([]model.FirewallRule(nil), policies...)
	sortFirewallPolicies(weighted, filter.ProviderFirewalld)
	if weighted[0].UUID != "low" || weighted[1].UUID != "high" || weighted[2].UUID != "none" {
		t.Fatalf("firewalld policies were not sorted by priority: %#v", weighted)
	}
}

func TestFirewallRuleSyncRejectsCurrentProviderAsSource(t *testing.T) {
	service := &FirewallService{
		rules:            repo.NewFirewallRuleRepo(newFirewallRuleTestDB(t)),
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderIptables, nil },
	}
	_, err := service.PreviewRuleSync(context.Background(), "", dto.FirewallRuleSyncRequest{
		SourceProvider: filter.ProviderIptables,
		TargetProvider: filter.ProviderIptables,
	})
	if err == nil {
		t.Fatal("expected identical source and target providers to be rejected")
	}
}

func TestDatabaseRuleSyncTargetUsesExplicitTargetOnly(t *testing.T) {
	target, err := databaseRuleSyncTarget(dto.FirewallRuleSyncRequest{TargetProvider: filter.ProviderNftables}, "Docker")
	if err != nil || target != filter.ProviderNftables {
		t.Fatalf("target = %q, err = %v", target, err)
	}
	if _, err := databaseRuleSyncTarget(dto.FirewallRuleSyncRequest{
		SourceProvider: filter.ProviderIptables,
		TargetProvider: filter.ProviderNftables,
	}, "Docker"); err == nil {
		t.Fatal("expected database-backed synchronization to reject a source provider")
	}
	if _, err := databaseRuleSyncTarget(dto.FirewallRuleSyncRequest{TargetProvider: filter.ProviderUFW}, "Docker"); err == nil {
		t.Fatal("expected database-backed synchronization to reject a non-netfilter target")
	}
}
