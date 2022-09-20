package migration

import (
	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/init/migration/migrations"

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
	})
	if err := m.Migrate(); err != nil {
		global.LOG.Error(err)
		panic(err)
	}
	global.LOG.Info("Migration did run successfully")
}
