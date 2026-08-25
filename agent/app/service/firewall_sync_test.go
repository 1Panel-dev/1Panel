package service

import (
	"context"
	"path/filepath"
	"sync"
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
	request := dto.FirewallRuleSyncRequest{SourceProvider: filter.ProviderIptables, TargetProvider: filter.ProviderNftables}

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
	targetRecords, err := ruleRepo.List(ctx, repo.WithByProvider(string(filter.ProviderNftables)))
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
		Subsystem: "system", SourceProvider: filter.ProviderIptables, TargetProvider: filter.ProviderNftables,
		TaskID: "firewall-sync-task",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !taskResult.Queued || taskResult.TaskID != "firewall-sync-task" {
		t.Fatalf("plain synchronization was not queued as a task: %#v", taskResult)
	}
	waitFirewallSyncTask(t, taskResult.TaskID)
	sourceRecords, err := ruleRepo.List(ctx, repo.WithByProvider(string(filter.ProviderIptables)))
	if err != nil || len(sourceRecords) != 1 {
		t.Fatalf("plain synchronization reset the source backend: %#v err=%v", sourceRecords, err)
	}
}

func TestFirewallRuleSyncTaskResetsSourceAfterSuccessfulMigration(t *testing.T) {
	ctx := context.Background()
	setupFirewallTaskTestDB(t)
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
	cleaned := ""
	cleanupStarted := make(chan struct{})
	continueCleanup := make(chan struct{})
	var releaseCleanup sync.Once
	release := func() { releaseCleanup.Do(func() { close(continueCleanup) }) }
	defer release()
	service := &FirewallService{
		rules: ruleRepo,
		adapters: firewallRuleRuntimeRegistry{
			filter.ProviderNftables: newFirewallRuleRuntime(target, nil),
		},
		selectedProvider: func(context.Context) (filter.Provider, error) { return filter.ProviderNftables, nil },
		cleanupInactiveBackend: func(provider string) error {
			cleaned = provider
			close(cleanupStarted)
			<-continueCleanup
			return nil
		},
	}
	result, err := service.SyncRules(ctx, "", dto.FirewallRuleSyncRequest{
		Subsystem: "system", SourceProvider: filter.ProviderIptables, TargetProvider: filter.ProviderNftables,
		ResetSource: true, TaskID: "firewall-migration-task",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Queued || result.TaskID != "firewall-migration-task" {
		t.Fatalf("unexpected queued result: %#v", result)
	}
	select {
	case <-cleanupStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for source cleanup")
	}
	running, err := service.CurrentRuleSyncTask()
	if err != nil || !running.Executing || running.TaskID != result.TaskID {
		t.Fatalf("unexpected running task: %#v err=%v", running, err)
	}
	duplicate, err := service.SyncRules(ctx, "", dto.FirewallRuleSyncRequest{
		Subsystem: "system", SourceProvider: filter.ProviderIptables, TargetProvider: filter.ProviderNftables,
		ResetSource: true, TaskID: "duplicate-firewall-migration-task",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !duplicate.Queued || duplicate.TaskID != result.TaskID {
		t.Fatalf("duplicate synchronization did not reuse the running task: %#v", duplicate)
	}
	plainSync, err := service.SyncRules(ctx, "", dto.FirewallRuleSyncRequest{
		Subsystem: "system", SourceProvider: filter.ProviderIptables, TargetProvider: filter.ProviderNftables,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !plainSync.Queued || plainSync.TaskID != result.TaskID {
		t.Fatalf("plain synchronization did not reuse the running task: %#v", plainSync)
	}
	release()

	waitFirewallSyncTask(t, result.TaskID)
	if cleaned != string(filter.ProviderIptables) {
		t.Fatalf("source backend was not reset: %q", cleaned)
	}
	sourceRecords, err := ruleRepo.List(ctx, repo.WithByProvider(string(filter.ProviderIptables)))
	if err != nil || len(sourceRecords) != 0 {
		t.Fatalf("source records were retained: %#v err=%v", sourceRecords, err)
	}
	targetRecords, err := ruleRepo.List(ctx, repo.WithByProvider(string(filter.ProviderNftables)))
	if err != nil || len(targetRecords) != 1 {
		t.Fatalf("target records were not retained: %#v err=%v", targetRecords, err)
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

func TestFirewallRulesForSyncProviderSplitsAddresslessInetRule(t *testing.T) {
	source := filter.FirewallRule{
		Scope: filter.Scope{
			Provider: filter.ProviderFirewalld, Family: filter.FamilyInet,
			Zone: filter.FirewalldInputZone, Direction: filter.DirectionInput,
		},
		NativeKind: filter.NativeKindRichRule,
		Protocol:   "tcp", DestinationPort: "443", Action: filter.ActionAccept,
	}
	rules, err := firewallRulesForSyncProvider(source, filter.ProviderUFW)
	if err != nil {
		t.Fatalf("convert inet rule: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("converted rule count = %d, want 2", len(rules))
	}
	for index, family := range []filter.Family{filter.FamilyIPv4, filter.FamilyIPv6} {
		if rules[index].Scope.Provider != filter.ProviderUFW || rules[index].Scope.Family != family ||
			rules[index].Scope.Chain != filter.UFWInputChain {
			t.Fatalf("converted rule %d = %#v", index, rules[index])
		}
		if rules[index].UUID != "" || rules[index].Priority != nil || rules[index].OrderIndex != nil {
			t.Fatalf("provider-native identity leaked into converted rule %d: %#v", index, rules[index])
		}
	}
}

func TestFirewallRulesForSyncProviderSplitsUnsupportedPortSet(t *testing.T) {
	source := filter.FirewallRule{
		Scope: filter.Scope{
			Provider: filter.ProviderIptables, Family: filter.FamilyIPv4,
			Table: "filter", Chain: filter.IptablesInputChain, Direction: filter.DirectionInput,
		},
		Protocol: "tcp", DestinationPort: "80,443", Interface: "*", Action: filter.ActionAccept,
	}
	rules, err := firewallRulesForSyncProvider(source, filter.ProviderNftables)
	if err != nil {
		t.Fatalf("convert iptables port set: %v", err)
	}
	if len(rules) != 2 || rules[0].DestinationPort != "80" || rules[1].DestinationPort != "443" {
		t.Fatalf("nftables port set was not split: %#v", rules)
	}
	for _, rule := range rules {
		if rule.Interface != "" {
			t.Fatalf("wildcard interface leaked into target rule: %#v", rule)
		}
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
