package service

import (
	"errors"
	"fmt"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
)

func TestDatabaseSyncPlanMatchesOnceAndSummarizesStates(t *testing.T) {
	desired := []databaseSyncDesired[int]{
		{value: 1, item: dto.FirewallRuleSyncItem{SourceUUID: "existing"}},
		{value: 1, item: dto.FirewallRuleSyncItem{SourceUUID: "duplicate"}},
		{value: 2, item: dto.FirewallRuleSyncItem{SourceUUID: "blocked"}, err: errors.New("invalid policy")},
	}
	plan := buildDatabaseSyncPlan(
		"test", filter.ProviderNftables, desired, []int{1, 3},
		func(value int) string { return fmt.Sprint(value) },
		func(value int) dto.FirewallRuleSyncItem {
			return dto.FirewallRuleSyncItem{SourceUUID: "actual"}
		},
	)

	preview := plan.preview()
	if preview.Total != 3 || preview.Ready != 1 || preview.Existing != 1 ||
		preview.Blocked != 1 || preview.Removed != 1 {
		t.Fatalf("unexpected preview: %#v", preview)
	}
	completed := plan.completedResult()
	if completed.Total != 3 || completed.Succeeded != 1 || completed.Skipped != 1 ||
		completed.Failed != 1 || completed.Removed != 1 || len(completed.Errors) != 1 {
		t.Fatalf("unexpected completed result: %#v", completed)
	}
	failed := plan.failedResult(errors.New("reconcile failed"))
	if failed.Succeeded != 0 || failed.Skipped != 1 || failed.Failed != 2 || failed.Removed != 0 ||
		len(failed.Errors) != 2 {
		t.Fatalf("unexpected failed result: %#v", failed)
	}
}
