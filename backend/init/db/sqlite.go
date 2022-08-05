package db

import (
	"1Panel/global"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SqliteGorm() *gorm.DB {
	s := global.Config.Sqlite
	if db, err := gorm.Open(sqlite.Open(s.Dsn()), &gorm.Config{}); err != nil {
		panic(err)
	} else {
		return db
	}
}
