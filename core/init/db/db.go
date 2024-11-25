package db

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Init() {
	initDB()
	initTaskDB()
}

func initDB() {
	dbPath := path.Join(global.CONF.System.BaseDir, "1panel/db")
	if _, err := os.Stat(dbPath); err != nil {
		if err := os.MkdirAll(dbPath, os.ModePerm); err != nil {
			panic(fmt.Errorf("init db dir failed, err: %v", err))
		}
	}
	fullPath := path.Join(dbPath, global.CONF.System.DbCoreFile)
	if _, err := os.Stat(fullPath); err != nil {
		f, err := os.Create(fullPath)
		if err != nil {
			panic(fmt.Errorf("init db file failed, err: %v", err))
		}
		_ = f.Close()
	}

	db, err := NewDBWithPath(fullPath)
	if err != nil {
		panic(err)
	}

	global.DB = db
	global.LOG.Info("init db successfully")
}

func initTaskDB() {
	fullPath := path.Join(global.CONF.System.BaseDir, "1panel/db/task.db")
	if _, err := os.Stat(fullPath); err != nil {
		f, err := os.Create(fullPath)
		if err != nil {
			panic(fmt.Errorf("init task db file failed, err: %v", err))
		}
		_ = f.Close()
	}

	db, err := NewDBWithPath(fullPath)
	if err != nil {
		panic(err)
	}

	global.TaskDB = db
	global.LOG.Info("init task db successfully")
}

func NewDBWithPath(dbPath string) (*gorm.DB, error) {
	db, _ := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   getLogger(),
	})
	sqlDB, dbError := db.DB()
	if dbError != nil {
		return nil, dbError
	}
	sqlDB.SetConnMaxIdleTime(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}
func getLogger() logger.Interface {
	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}
