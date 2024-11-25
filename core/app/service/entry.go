package service

import "github.com/1Panel-dev/1Panel/core/app/repo"

var (
	hostRepo     = repo.NewIHostRepo()
	commandRepo  = repo.NewICommandRepo()
	commonRepo   = repo.NewICommonRepo()
	settingRepo  = repo.NewISettingRepo()
	backupRepo   = repo.NewIBackupRepo()
	logRepo      = repo.NewILogRepo()
	groupRepo    = repo.NewIGroupRepo()
	launcherRepo = repo.NewILauncherRepo()

	taskRepo = repo.NewITaskRepo()
)
