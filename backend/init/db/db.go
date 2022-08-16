package db

import "github.com/1Panel-dev/1Panel/global"

func Init() {
	switch global.CONF.System.DbType {
	case "mysql":
		global.DB = MysqlGorm()
	case "sqlite":
		global.DB = SqliteGorm()
	default:
		global.DB = MysqlGorm()
	}
}
