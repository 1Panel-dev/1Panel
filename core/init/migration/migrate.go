package migration

import (
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/init/migration/migrations"

	"github.com/go-gormigrate/gormigrate/v2"
)

func Init() {
	m := gormigrate.New(global.DB, gormigrate.DefaultOptions, []*gormigrate.Migration{
		migrations.AddTable,
		migrations.InitSetting,
		migrations.InitOneDrive,
		migrations.InitHost,
		migrations.InitTerminalSetting,
		migrations.AddTaskDB,
		migrations.AddXpackHideMenu,
		migrations.UpdateXpackHideMenu,
		migrations.UpdateOnedrive,
		migrations.AddClusterMenu,
		migrations.DeleteXpackHideMenu,
		migrations.AddCronjobGroup,
		migrations.AddDiskMenu,
	})
	if err := m.Migrate(); err != nil {
		global.LOG.Error(err)
		panic(err)
	}
	global.LOG.Info("Migration run successfully")
}
