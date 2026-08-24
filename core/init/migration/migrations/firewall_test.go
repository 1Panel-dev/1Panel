package migrations

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/init/migration/helper"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestUpdateFirewallMenuPathMigratesCustomizedMenu(t *testing.T) {
	db := newCoreFirewallMigrationTestDB(t)
	menus := []dto.ShowMenu{{
		ID: "custom-parent", Label: "Custom", Children: []dto.ShowMenu{
			{ID: "74", Label: "renamed-firewall", Title: "custom.title", Path: "/hosts/firewall/port", Sort: 987, IsShow: false},
			{ID: "other", Label: "Other", Path: "/hosts/firewall/port", Sort: 123, IsShow: true},
		},
	}}
	seedCoreHideMenu(t, db, menus)

	if err := UpdateFirewallMenuPath.Migrate(db); err != nil {
		t.Fatal(err)
	}
	after := loadCoreHideMenu(t, db)
	firewall := after[0].Children[0]
	if firewall.Path != "/hosts/firewall/rules" || firewall.Title != "custom.title" || firewall.Sort != 987 || firewall.IsShow {
		t.Fatalf("migration did not preserve customized firewall menu: %#v", firewall)
	}
	if other := after[0].Children[1]; other.Path != "/hosts/firewall/port" {
		t.Fatalf("unrelated menu was changed: %#v", other)
	}
}

func TestUpdateFirewallMenuPathSupportsLegacyLabelAndIsIdempotent(t *testing.T) {
	db := newCoreFirewallMigrationTestDB(t)
	menus := []dto.ShowMenu{{Children: []dto.ShowMenu{
		{ID: "legacy-id", Label: "FirewallPort", Path: "/hosts/firewall/port"},
		{ID: "74", Label: "FirewallPort", Path: "/hosts/firewall/rules"},
	}}}
	seedCoreHideMenu(t, db, menus)

	for i := 0; i < 2; i++ {
		if err := UpdateFirewallMenuPath.Migrate(db); err != nil {
			t.Fatalf("migrate firewall menu on pass %d: %v", i+1, err)
		}
	}
	after := loadCoreHideMenu(t, db)
	for _, menu := range after[0].Children {
		if menu.Path != "/hosts/firewall/rules" {
			t.Fatalf("legacy firewall menu path remained after migration: %#v", menu)
		}
	}
}

func TestDefaultMenuUsesFirewallV2Route(t *testing.T) {
	var menus []dto.ShowMenu
	if err := json.Unmarshal([]byte(helper.LoadMenus()), &menus); err != nil {
		t.Fatal(err)
	}
	for _, parent := range menus {
		for _, child := range parent.Children {
			if child.ID == "74" {
				if child.Path != "/hosts/firewall/rules" {
					t.Fatalf("default firewall path = %q", child.Path)
				}
				return
			}
		}
	}
	t.Fatal("default firewall menu was not found")
}

func newCoreFirewallMigrationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "migration.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func seedCoreHideMenu(t *testing.T, db *gorm.DB, menus []dto.ShowMenu) {
	t.Helper()
	value, err := json.Marshal(menus)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.Setting{Key: "HideMenu", Value: string(value)}).Error; err != nil {
		t.Fatal(err)
	}
}

func loadCoreHideMenu(t *testing.T, db *gorm.DB) []dto.ShowMenu {
	t.Helper()
	var setting model.Setting
	if err := db.Where("key = ?", "HideMenu").First(&setting).Error; err != nil {
		t.Fatal(err)
	}
	var menus []dto.ShowMenu
	if err := json.Unmarshal([]byte(setting.Value), &menus); err != nil {
		t.Fatal(err)
	}
	return menus
}
