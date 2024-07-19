package service

import "github.com/1Panel-dev/1Panel/core/app/repo"

var (
	commonRepo  = repo.NewCommonRepo()
	settingRepo = repo.NewISettingRepo()
	logRepo     = repo.NewILogRepo()
)
