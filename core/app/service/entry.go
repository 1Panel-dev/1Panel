package service

import "github.com/1Panel-dev/1Panel/core/app/repo"

var (
	hostRepo       = repo.NewIHostRepo()
	commandRepo    = repo.NewICommandRepo()
	settingRepo    = repo.NewISettingRepo()
	backupRepo     = repo.NewIBackupRepo()
	logRepo        = repo.NewILogRepo()
	groupRepo      = repo.NewIGroupRepo()
	launcherRepo   = repo.NewILauncherRepo()
	upgradeLogRepo = repo.NewIUpgradeLogRepo()

	taskRepo   = repo.NewITaskRepo()
	agentRepo  = repo.NewIAgentRepo()
	scriptRepo = repo.NewIScriptRepo()
)
