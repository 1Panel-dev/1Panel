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
		migrations.InitBackup,
		migrations.InitGoogle,
		migrations.AddTaskDB,
		migrations.UpdateSettingStatus,
		migrations.RemoveLocalBackup,
		migrations.AddMFAInterval,
		migrations.UpdateXpackHideMemu,
		migrations.AddSystemIP,
		migrations.InitScriptLibrary,
		migrations.AddOperationNode,
		migrations.AddScriptVersion,
		migrations.AddAppStoreSetting,
	})
	if err := m.Migrate(); err != nil {
		global.LOG.Error(err)
		panic(err)
	}
	global.LOG.Info("Migration run successfully")
}
