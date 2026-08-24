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
