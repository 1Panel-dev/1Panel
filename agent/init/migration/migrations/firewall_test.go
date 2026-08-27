package migrations

import (
	"path/filepath"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAddFirewallRuleTableCreatesUpgradeSchema(t *testing.T) {
	db := newFirewallMigrationTestDB(t)
	for i := 0; i < 2; i++ {
		if err := AddFirewallRuleTable.Migrate(db); err != nil {
			t.Fatalf("migrate firewall v2 tables on pass %d: %v", i+1, err)
		}
	}

	for _, table := range []interface{}{&model.FirewallRule{}, &model.DockerPortGuardPolicy{}, &model.ForwardingRule{}} {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("migration did not create table for %T", table)
		}
	}
	for _, column := range []string{
		"provider", "scope_key", "location", "native_kind", "order_index", "order_bucket", "rule_key", "match_key",
	} {
		if db.Migrator().HasColumn("firewall_rules", column) {
			t.Fatalf("new firewall desired-rule schema persisted derived column %s", column)
		}
	}
	for _, column := range []string{"priority", "sequence"} {
		if !db.Migrator().HasColumn("firewall_rules", column) {
			t.Fatalf("new firewall desired-rule schema omitted placement column %s", column)
		}
	}

	policy := model.DockerPortGuardPolicy{
		UUID: "policy-1", Family: "ipv4", HostIP: "0.0.0.0", HostPort: 8080,
		Protocol: "tcp", Mode: "allow_all",
	}
	if err := db.Create(&policy).Error; err != nil {
		t.Fatalf("insert first Docker guard policy: %v", err)
	}
	duplicatePolicy := policy
	duplicatePolicy.BaseModel = model.BaseModel{}
	duplicatePolicy.UUID = "policy-2"
	if err := db.Create(&duplicatePolicy).Error; err == nil {
		t.Fatal("Docker guard endpoint uniqueness was not created")
	}

	forward := model.ForwardingRule{
		Family: "ipv4", Protocol: "tcp", Port: "8080", TargetIP: "10.0.0.2", TargetPort: "80",
	}
	if err := db.Create(&forward).Error; err != nil {
		t.Fatalf("insert first forwarding rule: %v", err)
	}
	duplicateForward := forward
	duplicateForward.BaseModel = model.BaseModel{}
	if err := db.Create(&duplicateForward).Error; err == nil {
		t.Fatal("forwarding identity uniqueness was not created")
	}
}

func TestInitDockerPortGuardStatusCreatesDefaultOnce(t *testing.T) {
	db := newFirewallMigrationTestDB(t)
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		if err := InitDockerPortGuardStatus.Migrate(db); err != nil {
			t.Fatalf("initialize Docker port guard status on pass %d: %v", i+1, err)
		}
	}

	var settings []model.Setting
	if err := db.Where("key = ?", constant.FirewallDockerPortGuardStatusKey).Find(&settings).Error; err != nil {
		t.Fatal(err)
	}
	if len(settings) != 1 || settings[0].Value != constant.StatusDisable {
		t.Fatalf("unexpected default Docker port guard settings: %#v", settings)
	}
}

func TestInitDockerPortGuardStatusPreservesUpgradeValue(t *testing.T) {
	db := newFirewallMigrationTestDB(t)
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatal(err)
	}
	existing := model.Setting{Key: constant.FirewallDockerPortGuardStatusKey, Value: constant.StatusEnable}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatal(err)
	}

	if err := InitDockerPortGuardStatus.Migrate(db); err != nil {
		t.Fatal(err)
	}
	var after model.Setting
	if err := db.Where("key = ?", constant.FirewallDockerPortGuardStatusKey).First(&after).Error; err != nil {
		t.Fatal(err)
	}
	if after.ID != existing.ID || after.Value != constant.StatusEnable {
		t.Fatalf("migration replaced persisted Docker guard status: before=%#v after=%#v", existing, after)
	}
}

func TestNormalizeFirewallBackendSelections(t *testing.T) {
	db := newFirewallMigrationTestDB(t)
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatal(err)
	}
	settings := []model.Setting{
		{Key: constant.FirewallDockerBackendKey, Value: constant.FirewallProviderNftables},
		{Key: constant.FirewallForwardingBackendKey, Value: constant.FirewallProviderFirewalld},
	}
	if err := db.Create(&settings).Error; err != nil {
		t.Fatal(err)
	}

	for pass := 1; pass <= 2; pass++ {
		if err := NormalizeFirewallBackendSelections.Migrate(db); err != nil {
			t.Fatalf("normalize firewall backend selections on pass %d: %v", pass, err)
		}
	}
	for key, want := range map[string]string{
		constant.FirewallDockerBackendKey:     constant.FirewallProviderNftables,
		constant.FirewallForwardingBackendKey: constant.FirewallProviderIptables,
	} {
		var matches []model.Setting
		if err := db.Where("key = ?", key).Find(&matches).Error; err != nil {
			t.Fatal(err)
		}
		if len(matches) != 1 || matches[0].Value != want {
			t.Fatalf("setting %s after normalization = %#v, want %q", key, matches, want)
		}
	}
}

func TestNormalizeFirewallBackendSelectionsPreservesValidForwardingBackend(t *testing.T) {
	db := newFirewallMigrationTestDB(t)
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatal(err)
	}
	existing := model.Setting{
		Key: constant.FirewallForwardingBackendKey, Value: constant.FirewallProviderNftables,
	}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatal(err)
	}

	for pass := 1; pass <= 2; pass++ {
		if err := NormalizeFirewallBackendSelections.Migrate(db); err != nil {
			t.Fatalf("normalize firewall backend selections on pass %d: %v", pass, err)
		}
	}
	var after model.Setting
	if err := db.Where("key = ?", constant.FirewallForwardingBackendKey).First(&after).Error; err != nil {
		t.Fatal(err)
	}
	if after.ID != existing.ID || after.Value != constant.FirewallProviderNftables {
		t.Fatalf("migration replaced valid forwarding backend: before=%#v after=%#v", existing, after)
	}
}

func TestNormalizeFirewallBackendSelectionsCreatesMissingDefaults(t *testing.T) {
	db := newFirewallMigrationTestDB(t)
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatal(err)
	}
	if err := NormalizeFirewallBackendSelections.Migrate(db); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		constant.FirewallDockerBackendKey:     "",
		constant.FirewallForwardingBackendKey: constant.FirewallProviderIptables,
	} {
		var matches []model.Setting
		if err := db.Where("key = ?", key).Find(&matches).Error; err != nil {
			t.Fatal(err)
		}
		if len(matches) != 1 || matches[0].Value != want {
			t.Fatalf("missing setting %s after normalization = %#v, want %q", key, matches, want)
		}
	}
}

func TestSimplifyFirewallRulePolicyKeepsDesiredRules(t *testing.T) {
	db := newFirewallMigrationTestDB(t)
	if err := db.Exec(`CREATE TABLE firewall_rules (
		uuid text PRIMARY KEY,
		provider text NOT NULL,
		family text NOT NULL,
		protocol text NOT NULL,
		scope_key text NOT NULL,
		location text NOT NULL,
		native_kind text NOT NULL,
		priority integer,
		order_index integer,
		order_bucket text,
		match_key text,
		rule_key text NOT NULL
	)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(
		"INSERT INTO firewall_rules (uuid, provider, family, protocol, scope_key, location, native_kind, priority, order_index, order_bucket, match_key, rule_key) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"policy-1", constant.FirewallProviderIptables, constant.FirewallFamilyIPv4, "tcp",
		"iptables:ipv4:filter:1PANEL_BASIC:input", "1PANEL_BASIC", "rule", 10, 3, "legacy", "instance:legacy", "legacy-key",
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(
		"INSERT INTO firewall_rules (uuid, provider, family, protocol, scope_key, location, native_kind, priority, order_index, order_bucket, match_key, rule_key) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"native-1", constant.FirewallProviderFirewalld, constant.FirewallFamilyInet, "all",
		"firewalld:inet:public:input", "public", "zone_service", nil, nil, "legacy", "", "native-key",
	).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(
		"INSERT INTO firewall_rules (uuid, provider, family, protocol, scope_key, location, native_kind, priority, order_index, order_bucket, match_key, rule_key) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"rich-1", constant.FirewallProviderFirewalld, constant.FirewallFamilyInet, "tcp",
		"firewalld:inet:public:input", "public", "rich_rule", -100, nil, "rich_pre", "", "rich-key",
	).Error; err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		"CREATE UNIQUE INDEX uk_firewall_rules_scope_rule ON firewall_rules(scope_key, rule_key)",
		"CREATE UNIQUE INDEX uk_firewall_rules_scope_match ON firewall_rules(scope_key, match_key) WHERE match_key <> ''",
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	for pass := 1; pass <= 2; pass++ {
		if err := SimplifyFirewallRulePolicy.Migrate(db); err != nil {
			t.Fatalf("simplify firewall rule policy on pass %d: %v", pass, err)
		}
	}
	for _, column := range []string{
		"provider", "scope_key", "location", "native_kind", "order_index", "order_bucket", "rule_key", "match_key",
	} {
		if db.Migrator().HasColumn("firewall_rules", column) {
			t.Fatalf("derived column %s was retained in firewall desired rules", column)
		}
	}
	var positional struct {
		Priority *int
		Sequence *int64
	}
	if err := db.Table("firewall_rules").Where("uuid = ?", "policy-1").First(&positional).Error; err != nil {
		t.Fatal(err)
	}
	if positional.Priority != nil || positional.Sequence != nil {
		t.Fatalf("positional policy migration inferred unsupported placement: %#v", positional)
	}
	var count int64
	if err := db.Table("firewall_rules").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("policy migration removed desired rules, count=%d", count)
	}
	var native struct {
		CompatibilityError string
	}
	if err := db.Table("firewall_rules").Where("uuid = ?", "native-1").First(&native).Error; err != nil {
		t.Fatal(err)
	}
	if native.CompatibilityError == "" {
		t.Fatal("legacy provider-native rule was not quarantined from synchronization")
	}
	var rich struct {
		Priority *int
		Sequence *int64
	}
	if err := db.Table("firewall_rules").Where("uuid = ?", "rich-1").First(&rich).Error; err != nil {
		t.Fatal(err)
	}
	if rich.Priority != nil || rich.Sequence != nil {
		t.Fatalf("firewalld policy migration retained unsupported placement: %#v", rich)
	}
}

func TestSimplifyFirewallRulePolicyContinuesWhenObsoleteColumnCannotBeDropped(t *testing.T) {
	db := newFirewallMigrationTestDB(t)
	if err := db.Exec(`CREATE TABLE firewall_rules (
		uuid text PRIMARY KEY,
		provider text NOT NULL,
		scope_key text NOT NULL,
		location text NOT NULL,
		native_kind text NOT NULL,
		priority integer
	)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE INDEX unexpected_scope_index ON firewall_rules(scope_key)").Error; err != nil {
		t.Fatal(err)
	}

	if err := SimplifyFirewallRulePolicy.Migrate(db); err != nil {
		t.Fatalf("obsolete column cleanup blocked the migration: %v", err)
	}
	for _, column := range []string{"provider", "location", "native_kind"} {
		if db.Migrator().HasColumn("firewall_rules", column) {
			t.Fatalf("obsolete column %s was not dropped after an unrelated drop failure", column)
		}
	}
	for _, column := range []string{"compatibility_error", "sequence"} {
		if !db.Migrator().HasColumn("firewall_rules", column) {
			t.Fatalf("migration omitted required column %s", column)
		}
	}
}

func newFirewallMigrationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "migration.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	return db
}
