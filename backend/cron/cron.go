package cron

import (
	"time"

	"github.com/1Panel-dev/1Panel/cron/job"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/robfig/cron/v3"
)

func Run() {
	nyc, _ := time.LoadLocation("Asia/Shanghai")
	Cron := cron.New(cron.WithLocation(nyc))
	_, err := Cron.AddJob("@every 1m", job.NewMonitorJob())
	if err != nil {
		global.LOG.Errorf("can not add corn job: %s", err.Error())
	}
	Cron.Start()

	global.Cron = Cron
}
