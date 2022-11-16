package db

import (
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Init() {
	s := global.CONF.Sqlite
	db, err := gorm.Open(sqlite.Open(s.Dsn()), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	global.DB = db
}
