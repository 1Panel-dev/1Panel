package migration

import (
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/init/migration/migrations"

	"github.com/go-gormigrate/gormigrate/v2"
)

func Init() {
	m := gormigrate.New(global.DB, gormigrate.DefaultOptions, []*gormigrate.Migration{
		migrations.AddTableOperationLog,
		migrations.AddTableHost,
		migrations.AddTableMonitor,
		migrations.AddTableSetting,
		migrations.AddTableBackupAccount,
		migrations.AddTableCronjob,
		migrations.AddTableApp,
		migrations.AddTableImageRepo,
		migrations.AddTableWebsite,
		migrations.AddTableDatabaseMysql,
		migrations.AddTableSnap,
		migrations.AddDefaultGroup,
		migrations.AddTableRuntime,
		migrations.UpdateTableApp,
		migrations.UpdateTableHost,
		migrations.UpdateTableWebsite,
		migrations.AddEntranceAndSSL,
	})
	if err := m.Migrate(); err != nil {
		global.LOG.Error(err)
		panic(err)
	}
	global.LOG.Info("Migration run successfully")
}
