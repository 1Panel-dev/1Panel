package common

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func LoadDBConnByPath(fullPath, dbName string) *gorm.DB {
	if _, err := CreateDirWhenNotExist(true, global.CONF.System.DbPath); err != nil {
		panic(fmt.Errorf("init db dir failed, err: %v", err))
	}
	if _, err := os.Stat(fullPath); err != nil {
		f, err := os.Create(fullPath)
		if err != nil {
			panic(fmt.Errorf("init %s db file failed, err: %v", dbName, err))
		}
		_ = f.Close()
	}

	db, err := GetDBWithPath(fullPath)
	if err != nil {
		panic(err)
	}
	return db
}

func LoadDBConnByPathWithErr(fullPath, dbName string) (*gorm.DB, error) {
	if _, err := CreateDirWhenNotExist(true, global.CONF.System.DbPath); err != nil {
		return nil, fmt.Errorf("init db dir failed, err: %v", err)
	}
	if _, err := os.Stat(fullPath); err != nil {
		f, err := os.Create(fullPath)
		if err != nil {
			return nil, fmt.Errorf("init %s db file failed, err: %v", dbName, err)
		}
		_ = f.Close()
	}

	db, err := GetDBWithPath(fullPath)
	if err != nil {
		return nil, fmt.Errorf("init %s db failed, err: %v", dbName, err)
	}
	return db, nil
}

func CloseDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	_ = sqlDB.Close()
}

func GetDBWithPath(dbPath string) (*gorm.DB, error) {
	db, _ := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   newLogger(),
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

func newLogger() logger.Interface {
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
