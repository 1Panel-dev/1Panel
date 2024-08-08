package cron

import (
	"time"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/service"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/robfig/cron/v3"
)

func Init() {
	nyc, _ := time.LoadLocation(common.LoadTimeZone())
	global.Cron = cron.New(cron.WithLocation(nyc), cron.WithChain(cron.Recover(cron.DefaultLogger)), cron.WithChain(cron.DelayIfStillRunning(cron.DefaultLogger)))

	var accounts []model.BackupAccount
	_ = global.DB.Where("type = ?", "OneDrive").Find(&accounts).Error
	for i := 0; i < len(accounts); i++ {
		_ = service.StartRefreshOneDriveToken(&accounts[i])
	}
	global.Cron.Start()
}
