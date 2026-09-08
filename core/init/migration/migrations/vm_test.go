package migrations

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestVMMigrationAddsMenuAndPreservesPreferences(t *testing.T) {
	for _, existing := range []bool{false, true} {
		db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "core.db")), &gorm.Config{})
		if err != nil {
			t.Fatal(err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = sqlDB.Close() })
		if err := db.AutoMigrate(&model.Setting{}); err != nil {
			t.Fatal(err)
		}
		menus := []dto.ShowMenu{{Label: "Xpack-Menu", Children: []dto.ShowMenu{{Label: "Sync"}}}}
		if existing {
			menus[0].Children = append(menus[0].Children, dto.ShowMenu{ID: "123", Label: "VirtualMachine", Path: "/enterprise/vm", IsShow: false, Sort: 555})
		}
		value, err := json.Marshal(menus)
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Create(&model.Setting{Key: "HideMenu", Value: string(value)}).Error; err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 2; i++ {
			if err := MoveVirtualMachineMenuToXpack.Migrate(db); err != nil {
				t.Fatal(err)
			}
		}
		var setting model.Setting
		if err := db.Where("key = ?", "HideMenu").First(&setting).Error; err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(setting.Value), &menus); err != nil {
			t.Fatal(err)
		}
		if len(menus[0].Children) != 2 {
			t.Fatalf("menu duplicated or missing: %+v", menus)
		}
		vm := menus[0].Children[1]
		if vm.Path != "/xpack/vm" || vm.IsShow == existing || (existing && vm.Sort != 555) {
			t.Fatalf("incorrect menu/preferences: %+v", vm)
		}
	}
}
