package cron

import (
	"fmt"
	mathRand "math/rand"
	"time"

	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/init/cron/job"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/robfig/cron/v3"
)

func Init() {
	nyc, _ := time.LoadLocation(common.LoadTimeZoneByCmd())
	global.Cron = cron.New(cron.WithLocation(nyc), cron.WithChain(cron.Recover(cron.DefaultLogger)), cron.WithChain(cron.DelayIfStillRunning(cron.DefaultLogger)))

	if _, err := global.Cron.AddJob("0 3 */31 * *", job.NewBackupJob()); err != nil {
		global.LOG.Errorf("[core] can not add backup token refresh corn job: %s", err.Error())
	}

	scriptJob := job.NewScriptJob()
	scriptJob.Run()
	if _, err := global.Cron.AddJob(fmt.Sprintf("%v %v * * *", mathRand.Intn(60), mathRand.Intn(3)), scriptJob); err != nil {
		global.LOG.Errorf("[core] can not add script sync corn job: %s", err.Error())
	}
	global.Cron.Start()
}
