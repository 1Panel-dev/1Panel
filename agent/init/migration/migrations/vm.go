package migrations

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/glebarez/sqlite"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var MigrateVMDatabase = &gormigrate.Migration{
	ID: "20261102-migrate-vm-to-community",
	Migrate: func(db *gorm.DB) error {
		return migrateVMDatabase(db, global.Dir.DbDir)
	},
}

func migrateVMDatabase(db *gorm.DB, dbDir string) error {
	var legacy *gorm.DB
	legacyPath := filepath.Join(dbDir, "xpack.db")
	if _, err := os.Stat(legacyPath); err == nil {
		uri := (&url.URL{Scheme: "file", Path: legacyPath, RawQuery: "mode=ro"}).String()
		legacy, err = gorm.Open(sqlite.Open(uri), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
		if err != nil {
			return fmt.Errorf("open legacy VM database: %w", err)
		}
		sqlDB, err := legacy.DB()
		if err != nil {
			return err
		}
		defer sqlDB.Close()
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect legacy VM database: %w", err)
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.VirtualMachine{}, &model.VMTemplate{}, &model.VMIso{}, &model.VMNetwork{}, &model.VMStorage{}); err != nil {
			return err
		}
		if legacy != nil {
			tables := []struct {
				name    string
				records interface{}
			}{
				{name: "virtual_machines", records: &[]model.VirtualMachine{}},
				{name: "vm_templates", records: &[]model.VMTemplate{}},
				{name: "vm_isos", records: &[]model.VMIso{}},
				{name: "vm_networks", records: &[]model.VMNetwork{}},
				{name: "vm_storages", records: &[]model.VMStorage{}},
			}
			for _, table := range tables {
				if err := migrateVMTable(legacy, tx, table.name, table.records); err != nil {
					return fmt.Errorf("migrate %s: %w", table.name, err)
				}
			}
		}
		return nil
	})
}

func migrateVMTable(source, target *gorm.DB, table string, records interface{}) error {
	if !source.Migrator().HasTable(table) {
		return nil
	}
	if err := target.Table(table).Where("1 = 1").Delete(records).Error; err != nil {
		return err
	}
	query := source.Table(table).Find(records)
	if query.Error != nil {
		return query.Error
	}
	if query.RowsAffected != 0 {
		if err := target.Table(table).CreateInBatches(records, 50).Error; err != nil {
			return err
		}
		var migrated int64
		if err := target.Table(table).Count(&migrated).Error; err != nil {
			return err
		}
		if migrated != query.RowsAffected {
			return fmt.Errorf("expected %d records, migrated %d", query.RowsAffected, migrated)
		}
	}
	return migrateVMSequence(source, target, table)
}

func migrateVMSequence(source, target *gorm.DB, table string) error {
	var sequenceTable int64
	if err := source.Raw("SELECT COUNT(*) FROM sqlite_master WHERE name = 'sqlite_sequence'").Scan(&sequenceTable).Error; err != nil {
		return err
	}
	if sequenceTable == 0 {
		return nil
	}
	var sequence int64
	if err := source.Raw("SELECT COALESCE(MAX(seq), 0) FROM sqlite_sequence WHERE name = ?", table).Scan(&sequence).Error; err != nil {
		return err
	}
	if err := target.Exec("INSERT INTO sqlite_sequence (name, seq) SELECT ?, ? WHERE NOT EXISTS (SELECT 1 FROM sqlite_sequence WHERE name = ?)", table, sequence, table).Error; err != nil {
		return err
	}
	return target.Exec("UPDATE sqlite_sequence SET seq = MAX(seq, ?) WHERE name = ?", sequence, table).Error
}
