package cron

import (
	"time"

	"github.com/1Panel-dev/1Panel/core/app/service"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/robfig/cron/v3"
)

func Init() {
	nyc, _ := time.LoadLocation(common.LoadTimeZone())
	global.Cron = cron.New(cron.WithLocation(nyc), cron.WithChain(cron.Recover(cron.DefaultLogger)), cron.WithChain(cron.DelayIfStillRunning(cron.DefaultLogger)))

	_ = service.StartRefreshForToken()
	global.Cron.Start()
}
