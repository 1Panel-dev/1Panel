package db

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Init() {
	if _, err := os.Stat(global.CONF.System.DbPath); err != nil {
		if err := os.MkdirAll(global.CONF.System.DbPath, os.ModePerm); err != nil {
			panic(fmt.Errorf("init db dir failed, err: %v", err))
		}
	}
	fullPath := global.CONF.System.DbPath + "/" + global.CONF.System.DbFile
	if _, err := os.Stat(fullPath); err != nil {
		f, err := os.Create(fullPath)
		if err != nil {
			panic(fmt.Errorf("init db file failed, err: %v", err))
		}
		_ = f.Close()
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
	initMonitorDB(newLogger)

	db, err := gorm.Open(sqlite.Open(fullPath), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   newLogger,
	})
	if err != nil {
		panic(err)
	}
	_ = db.Exec("PRAGMA journal_mode = WAL;")
	sqlDB, dbError := db.DB()
	if dbError != nil {
		panic(dbError)
	}
	sqlDB.SetConnMaxIdleTime(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	global.DB = db
	global.LOG.Info("init db successfully")
}

func initMonitorDB(newLogger logger.Interface) {
	if _, err := os.Stat(global.CONF.System.DbPath); err != nil {
		if err := os.MkdirAll(global.CONF.System.DbPath, os.ModePerm); err != nil {
			panic(fmt.Errorf("init db dir failed, err: %v", err))
		}
	}
	fullPath := path.Join(global.CONF.System.DbPath, "monitor.db")
	if _, err := os.Stat(fullPath); err != nil {
		f, err := os.Create(fullPath)
		if err != nil {
			panic(fmt.Errorf("init db file failed, err: %v", err))
		}
		_ = f.Close()
	}

	db, err := gorm.Open(sqlite.Open(fullPath), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   newLogger,
	})
	if err != nil {
		panic(err)
	}
	sqlDB, dbError := db.DB()
	if dbError != nil {
		panic(dbError)
	}
	sqlDB.SetConnMaxIdleTime(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	global.MonitorDB = db
	global.LOG.Info("init monitor db successfully")
}
