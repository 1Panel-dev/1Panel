package cron

import (
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
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
		global.LOG.Errorf("can not add monitor corn job: %s", err.Error())
	}
	_, err = Cron.AddJob("@daily", job.NewWebsiteJob())
	if err != nil {
		global.LOG.Errorf("can not add  website corn job: %s", err.Error())
	}
	_, err = Cron.AddJob("@daily", job.NewSSLJob())
	if err != nil {
		global.LOG.Errorf("can not add  ssl corn job: %s", err.Error())
	}
	Cron.Start()

	global.Cron = Cron

	var cronJobs []model.Cronjob
	if err := global.DB.Where("status = ?", constant.StatusEnable).Find(&cronJobs).Error; err != nil {
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
	for i := 0; i < len(cronJobs); i++ {
		entryID, err := service.ServiceGroupApp.StartJob(&cronJobs[i])
		if err != nil {
			global.LOG.Errorf("start %s job %s failed, err: %v", &cronJobs[i].Type, &cronJobs[i].Name, err)
		}
		if err := repo.NewICronjobRepo().Update(cronJobs[i].ID, map[string]interface{}{"entry_id": entryID}); err != nil {
			global.LOG.Errorf("update cronjob %s %s failed, err: %v", cronJobs[i].Type, cronJobs[i].Name, err)
		}
	}
}
