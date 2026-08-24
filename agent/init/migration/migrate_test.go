package migration

import "testing"

func TestAgentDBMigrationsRegisterFirewallUpgradeSteps(t *testing.T) {
	migrations := agentDBMigrations()
	indexes := make(map[string]int, len(migrations))
	for index, migration := range migrations {
		if migration == nil || migration.ID == "" {
			t.Fatalf("invalid migration at index %d: %#v", index, migration)
		}
		if previous, exists := indexes[migration.ID]; exists {
			t.Fatalf("duplicate migration ID %q at indexes %d and %d", migration.ID, previous, index)
		}
		indexes[migration.ID] = index
	}

	tables, hasTables := indexes["20260819-add-firewall-v2-tables"]
	status, hasStatus := indexes["20260818-init-docker-port-guard-status"]
	if !hasTables || !hasStatus {
		t.Fatalf("firewall upgrade migrations are not registered: %#v", indexes)
	}
	if tables > status {
		t.Fatalf("firewall table migration index %d runs after status migration index %d", tables, status)
	}
}
