package db

import "1Panel/global"

func Init() {
	switch global.Config.System.DbType {
	case "mysql":
		global.DB = MysqlGorm()
	case "sqlite":
		global.DB = SqliteGorm()
	default:
		global.DB = MysqlGorm()
	}
}
