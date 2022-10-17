package cron

import (
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/service"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/cron/job"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/robfig/cron/v3"
)

func Run() {
	nyc, _ := time.LoadLocation("Asia/Shanghai")
	Cron := cron.New(cron.WithLocation(nyc), cron.WithChain(cron.Recover(cron.DefaultLogger)), cron.WithChain(cron.DelayIfStillRunning(cron.DefaultLogger)))
	_, err := Cron.AddJob("@every 1m", job.NewMonitorJob())
	if err != nil {
		global.LOG.Errorf("can not add corn job: %s", err.Error())
	}
	Cron.Start()

	global.Cron = Cron

	var Cronjobs []model.Cronjob
	if err := global.DB.Where("status = ?", constant.StatusEnable).Find(&Cronjobs).Error; err != nil {
		global.LOG.Errorf("start my cronjob failed, err: %v", err)
	}
	if err := global.DB.Model(&model.JobRecords{}).
		Where("status = ?", constant.StatusRunning).
		Updates(map[string]interface{}{
			"status":  constant.StatusFailed,
			"message": "Task Cancel",
			"records": "errHandle",
		}).Error; err != nil {
		global.LOG.Errorf("start my cronjob failed, err: %v", err)
	}
	for _, cronjob := range Cronjobs {
		if err := service.ServiceGroupApp.StartJob(&cronjob); err != nil {
			global.LOG.Errorf("start %s job %s failed, err: %v", cronjob.Type, cronjob.Name, err)
		}
	}
}
