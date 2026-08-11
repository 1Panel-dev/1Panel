package migrations

import (
	"fmt"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMigrateFirewallRuleTableIsRepeatable(t *testing.T) {
	db := newFirewallMigrationTestDB(t)
	if err := migrateFirewallRuleTable(db); err != nil {
		t.Fatalf("first migration: %v", err)
	}
	if err := migrateFirewallRuleTable(db); err != nil {
		t.Fatalf("repeated migration: %v", err)
	}

	if !db.Migrator().HasTable("firewall_rules") {
		t.Fatal("expected firewall_rules table")
	}

	indexes := []string{
		"uk_firewall_rules_scope_rule",
		"uk_firewall_rules_scope_match",
		"idx_firewall_rules_owner",
	}
	for _, index := range indexes {
		if !db.Migrator().HasIndex(&model.FirewallRule{}, index) {
			t.Fatalf("expected index %s", index)
		}
	}
}

func TestRemoveFirewallLegacyInfluenceIsRepeatable(t *testing.T) {
	db := newFirewallMigrationTestDB(t)
	if err := migrateFirewallRuleTable(db); err != nil {
		t.Fatalf("migrate firewall tables: %v", err)
	}
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatalf("migrate settings: %v", err)
	}

	legacy := newMigratedFirewallRule("legacy", "sha256:legacy", nil)
	legacy.Origin = "legacy"
	managed := newMigratedFirewallRule("managed", "sha256:managed", nil)
	if err := db.Create(&legacy).Error; err != nil {
		t.Fatalf("create legacy metadata: %v", err)
	}
	if err := db.Create(&managed).Error; err != nil {
		t.Fatalf("create managed metadata: %v", err)
	}
	if err := db.Create(&model.Setting{Key: "FirewallV2", Value: "Disable"}).Error; err != nil {
		t.Fatalf("create legacy feature setting: %v", err)
	}

	if err := removeFirewallLegacyInfluence(db); err != nil {
		t.Fatalf("first cleanup: %v", err)
	}
	if err := removeFirewallLegacyInfluence(db); err != nil {
		t.Fatalf("repeated cleanup: %v", err)
	}

	var activeLegacy int64
	if err := db.Model(&model.FirewallRule{}).Where("origin = ?", "legacy").Count(&activeLegacy).Error; err != nil {
		t.Fatalf("count active legacy metadata: %v", err)
	}
	if activeLegacy != 0 {
		t.Fatalf("legacy metadata still affects inventory: count=%d", activeLegacy)
	}
	var deletedLegacy int64
	if err := db.Model(&model.FirewallRule{}).Where("uuid = ?", legacy.UUID).Count(&deletedLegacy).Error; err != nil || deletedLegacy != 0 {
		t.Fatalf("legacy metadata was not hard deleted: count=%d err=%v", deletedLegacy, err)
	}
	var activeManaged int64
	if err := db.Model(&model.FirewallRule{}).Where("uuid = ?", managed.UUID).Count(&activeManaged).Error; err != nil {
		t.Fatalf("count managed metadata: %v", err)
	}
	if activeManaged != 1 {
		t.Fatalf("managed metadata was changed: count=%d", activeManaged)
	}
	var featureSetting int64
	if err := db.Model(&model.Setting{}).Where("key = ?", "FirewallV2").Count(&featureSetting).Error; err != nil {
		t.Fatalf("count legacy feature setting: %v", err)
	}
	if featureSetting != 0 {
		t.Fatalf("legacy feature setting still affects runtime: count=%d", featureSetting)
	}
}

func TestFirewallRuleUniqueIndexes(t *testing.T) {
	db := newFirewallMigrationTestDB(t)
	if err := migrateFirewallRuleTable(db); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}

	marker := "1panel-rule:first"
	first := newMigratedFirewallRule("first", "sha256:same", &marker)
	if err := db.Create(&first).Error; err != nil {
		t.Fatalf("create first rule: %v", err)
	}
	duplicateSemantic := newMigratedFirewallRule("duplicate", "sha256:same", nil)
	if err := db.Create(&duplicateSemantic).Error; err == nil {
		t.Fatal("expected active semantic duplicate to fail")
	}
	otherMarkerRule := newMigratedFirewallRule("marker-duplicate", "sha256:other", &marker)
	if err := db.Create(&otherMarkerRule).Error; err == nil {
		t.Fatal("expected active marker duplicate to fail")
	}

	if err := db.Delete(&first).Error; err != nil {
		t.Fatalf("hard delete first rule: %v", err)
	}
	if err := db.Create(&duplicateSemantic).Error; err != nil {
		t.Fatalf("create semantic replacement after hard delete: %v", err)
	}

}

func newFirewallMigrationTestDB(t *testing.T) *gorm.DB {
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
	return db
}

func newMigratedFirewallRule(name, ruleKey string, marker *string) model.FirewallRule {
	matchKey := ""
	if marker != nil {
		matchKey = "marker:" + *marker
	}
	return model.FirewallRule{
		UUID:            name,
		ScopeKey:        "iptables:ipv4:filter:1PANEL_BASIC:input",
		Provider:        "iptables",
		Family:          "ipv4",
		Location:        "1PANEL_BASIC",
		NativeKind:      "rule",
		Protocol:        "tcp",
		DestinationPort: "22",
		Action:          "accept",
		RuleKey:         ruleKey,
		Origin:          constant.FirewallRuleOriginCreated,
		Owner:           constant.FirewallRuleSourceUser,
		MatchKey:        matchKey,
		Revision:        1,
	}
}
