package common

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func LoadDBConnByPath(fullPath, dbName string) *gorm.DB {
	if _, err := createDirWhenNotExist(true, path.Dir(fullPath)); err != nil {
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
	if _, err := createDirWhenNotExist(true, path.Dir(fullPath)); err != nil {
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
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   newLogger(),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(0)
	sqlDB.SetConnMaxIdleTime(0)

	if err := db.Exec("PRAGMA journal_mode = WAL;").Error; err != nil {
		return nil, err
	}
	if err := db.Exec("PRAGMA synchronous = NORMAL;").Error; err != nil {
		return nil, err
	}
	if err := db.Exec("PRAGMA busy_timeout = 5000;").Error; err != nil {
		return nil, err
	}
	if err := db.Exec("PRAGMA temp_store = MEMORY;").Error; err != nil {
		return nil, err
	}
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

func createDirWhenNotExist(isDir bool, pathItem string) (string, error) {
	checkPath := pathItem
	if !isDir {
		checkPath = path.Dir(pathItem)
	}
	if _, err := os.Stat(checkPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(checkPath, os.ModePerm); err != nil {
			global.LOG.Errorf("mkdir %s failed, err: %v", checkPath, err)
			return pathItem, err
		}
	}
	return pathItem, nil
}
