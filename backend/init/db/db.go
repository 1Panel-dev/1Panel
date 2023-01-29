package db

import (
	"fmt"
	"os"

	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	db, err := gorm.Open(sqlite.Open(fullPath), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	global.DB = db
}
