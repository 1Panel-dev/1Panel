package service

import "github.com/1Panel-dev/1Panel/core/app/repo"

var (
	commonRepo  = repo.NewICommonRepo()
	settingRepo = repo.NewISettingRepo()
	backupRepo  = repo.NewIBackupRepo()
	logRepo     = repo.NewILogRepo()
)
