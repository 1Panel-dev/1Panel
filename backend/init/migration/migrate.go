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
		migrations.UpdateTableSetting,
		migrations.UpdateTableAppDetail,
		migrations.AddBindAndAllowIPs,
		migrations.UpdateCronjobWithSecond,
		migrations.UpdateWebsite,
		migrations.AddBackupAccountDir,
		migrations.AddMfaInterval,
		migrations.UpdateAppDetail,
		migrations.EncryptHostPassword,
		migrations.AddRemoteDB,
		migrations.UpdateRedisParam,
		migrations.UpdateCronjobWithDb,
		migrations.AddTableFirewall,
		migrations.AddDatabases,
		migrations.UpdateDatabase,
		migrations.UpdateAppInstallResource,
		migrations.DropDatabaseLocal,

		migrations.AddDefaultNetwork,
		migrations.UpdateRuntime,
		migrations.UpdateTag,

		migrations.AddFavorite,
		migrations.AddBindAddress,
		migrations.AddCommandGroup,
		migrations.AddAppSyncStatus,

		migrations.UpdateAcmeAccount,
		migrations.UpdateWebsiteSSL,
		migrations.AddWebsiteCA,
		migrations.AddDockerSockPath,
		migrations.AddDatabaseSSL,
		migrations.AddDefaultCA,
		migrations.AddSettingRecycleBin,
		migrations.UpdateWebsiteBackupRecord,

		migrations.AddTablePHPExtensions,
		migrations.AddTableDatabasePostgresql,
		migrations.AddPostgresqlSuperUser,
		migrations.UpdateCronjobWithWebsite,
		migrations.UpdateOneDriveToken,
		migrations.UpdateCronjobSpec,
		migrations.UpdateBackupRecordPath,
		migrations.UpdateSnapshotRecords,

		migrations.UpdateWebDavConf,

		migrations.AddSnapshotIgnore,
		migrations.AddDatabaseIsDelete,
		migrations.AddXpackHideMenu,
		migrations.AddCronjobCommand,
		migrations.NewMonitorDB,
		migrations.AddNoAuthSetting,
		migrations.UpdateXpackHideMenu,
		migrations.AddMenuTabsSetting,
		migrations.AddDeveloperSetting,

		migrations.AddWebsiteSSLColumn,
		migrations.AddRedisCommand,
		migrations.AddMonitorMenu,
		migrations.AddFtp,
		migrations.AddProxy,
		migrations.AddCronJobColumn,
		migrations.AddForward,
		migrations.AddShellColumn,
		migrations.AddClam,
		migrations.AddClamStatus,
		migrations.AddAlertMenu,
		migrations.AddComposeColumn,

		migrations.AddAutoRestart,
		migrations.AddApiInterfaceConfig,
		migrations.AddApiKeyValidityTime,

		migrations.UpdateAppTag,
		migrations.UpdateApp,
	})
	if err := m.Migrate(); err != nil {
		global.LOG.Error(err)
		panic(err)
	}
	global.LOG.Info("Migration run successfully")
}
