package migration

import (
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/init/migration/migrations"

	"github.com/go-gormigrate/gormigrate/v2"
)

func Init() {
	InitAgentDB()
	InitTaskDB()
	InitAlertDB()
	global.LOG.Info("Migration run successfully")
}

func InitAgentDB() {
	m := gormigrate.New(global.DB, gormigrate.DefaultOptions, []*gormigrate.Migration{
		migrations.AddTable,
		migrations.AddMonitorTable,
		migrations.InitSetting,
		migrations.InitImageRepo,
		migrations.InitDefaultCA,
		migrations.InitPHPExtensions,
		migrations.InitBackup,
		migrations.InitDefault,
		migrations.UpdateWebsiteExpireDate,
		migrations.UpdateRuntime,
		migrations.AddSnapshotRule,
		migrations.UpdatePHPRuntime,
		migrations.AddSnapshotIgnore,
		migrations.InitAppLauncher,
		migrations.AddTableAlert,
		migrations.InitAlertConfig,
		migrations.AddMethodToAlertLog,
		migrations.AddMethodToAlertTask,
		migrations.UpdateMcpServer,
		migrations.InitCronjobGroup,
		migrations.AddColumnToAlert,
		migrations.UpdateWebsiteSSL,
		migrations.AddQuickJump,
		migrations.UpdateMcpServerAddType,
		migrations.UpdateMcpServerGatewayConfig,
		migrations.InitLocalSSHConn,
		migrations.InitLocalSSHShow,
		migrations.InitRecordStatus,
		migrations.AddShowNameForQuickJump,
		migrations.AddAgentQuickJump,
		migrations.AddTimeoutForClam,
		migrations.UpdateCronjobSpec,
		migrations.UpdateWebsiteSSLAddColumn,
		migrations.AddTensorRTLLMModel,
		migrations.UpdateMonitorInterval,
		migrations.AddMonitorProcess,
		migrations.UpdateCronJob,
		migrations.UpdateTensorrtLLM,
		migrations.AddIptablesFilterRuleTable,
		migrations.AddCommonDescription,
		migrations.UpdateDatabase,
		migrations.AddGPUMonitor,
		migrations.UpdateDatabaseMysql,
		migrations.AddDatabaseMongodb,
		migrations.InitIptablesStatus,
		migrations.UpdateWebsite,
		migrations.AddisIPtoWebsiteSSL,
		migrations.InitPingStatus,
		migrations.UpdateApp,
		migrations.AddCronjobArgs,
		migrations.AddWebsiteAcmeAccountColumn,
		migrations.AddAgentTables,
		migrations.AddAgentCustomModelFields,
		migrations.AddAppInstallSortOrder,
		migrations.AddAgentAccountRememberAPIKey,
		migrations.AddEditionSetting,
		migrations.AddAgentTypeForAgents,
		migrations.NormalizeAgentAccountVerifiedStatus,
		migrations.NormalizeOllamaAccountAPIType,
		migrations.InitAgentAccountModelPool,
		migrations.AddAgentAccountMasterID,
		migrations.NormalizeAgentAccountModelIDs,
		migrations.AddAgentAccountVerifyModel,
		migrations.AddAgentAccountAuthMode,
		migrations.AddHostTable,
		migrations.AddAITerminalSettings,
		migrations.UpdateAgentQuickJumpTitle,
		migrations.FixOpenclaw20260323HTTPPort,
		migrations.AddAgentRemarkColumn,
		migrations.AddAgentWebsiteBinding,
		migrations.AddFileManageAISettings,
		migrations.AddFileShareTable,
		migrations.AddFileHistoryTable,
		migrations.MigrateLegoV5,
		migrations.AddMcpServerGatewayArgs,
		migrations.InitFirewallPortWhiteList,
		migrations.AddDatabaseUserTable,
		migrations.AddBackupRecordArgs,
		migrations.AddFtpIdentity,
		migrations.AddWebsiteTemplateTable,
	})
	if err := m.Migrate(); err != nil {
		global.LOG.Error(err)
		panic(err)
	}
}

func InitTaskDB() {
	m := gormigrate.New(global.TaskDB, gormigrate.DefaultOptions, []*gormigrate.Migration{
		migrations.AddTaskTable,
	})
	if err := m.Migrate(); err != nil {
		global.LOG.Error(err)
		panic(err)
	}
}

func InitAlertDB() {
	m := gormigrate.New(global.AlertDB, gormigrate.DefaultOptions, []*gormigrate.Migration{
		migrations.MigrateAlertMethodConfigIDs,
		migrations.MigrateAlertLogTaskMethodConfigIDs,
		migrations.AddAlertAuditUser,
	})
	if err := m.Migrate(); err != nil {
		global.LOG.Error(err)
		panic(err)
	}
}
