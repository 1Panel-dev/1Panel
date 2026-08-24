package migration

import "testing"

func TestCoreMigrationsRegisterFirewallMenuUpgrade(t *testing.T) {
	migrations := coreMigrations()
	seen := make(map[string]int, len(migrations))
	foundFirewallMenu := false
	for index, migration := range migrations {
		if migration == nil || migration.ID == "" {
			t.Fatalf("invalid migration at index %d: %#v", index, migration)
		}
		if previous, exists := seen[migration.ID]; exists {
			t.Fatalf("duplicate migration ID %q at indexes %d and %d", migration.ID, previous, index)
		}
		seen[migration.ID] = index
		if migration.ID == "20260819-update-firewall-menu-path" {
			foundFirewallMenu = true
		}
	}
	if !foundFirewallMenu {
		t.Fatal("firewall menu path migration is not registered")
	}
}
