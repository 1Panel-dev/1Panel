package db

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/1Panel-dev/1Panel/backend/global"
)

func Init() {
	if _, err := os.Stat(global.CONF.System.DbPath); err != nil {
		if err := os.MkdirAll(global.CONF.System.DbPath, os.ModePerm); err != nil {
			panic(fmt.Errorf("init db dir falied, err: %v", err))
		}
	}
	fullPath := global.CONF.System.DbPath + "/" + global.CONF.System.DbFile
	if _, err := os.Stat(fullPath); err != nil {
		if _, err := os.Create(fullPath); err != nil {
			panic(fmt.Errorf("init db file falied, err: %v", err))
		}
	}

	db, err := gorm.Open(sqlite.Open(fullPath), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic(err)
	}
	_ = db.Exec("PRAGMA journal_mode = WAL;")
	sqlDB, dbError := db.DB()
	if dbError != nil {
		panic(err)
	}
	sqlDB.SetConnMaxIdleTime(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	global.DB = db
	global.LOG.Info("init db successfully")
}
