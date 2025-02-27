package cron

import (
	"fmt"
	mathRand "math/rand"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/cron/job"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/ntp"
	"github.com/robfig/cron/v3"
)

func Run() {
	nyc, _ := time.LoadLocation(common.LoadTimeZoneByCmd())
	global.Cron = cron.New(cron.WithLocation(nyc), cron.WithChain(cron.Recover(cron.DefaultLogger)), cron.WithChain(cron.DelayIfStillRunning(cron.DefaultLogger)))

	var (
		interval model.Setting
		status   model.Setting
	)
	go syncBeforeStart()
	if err := global.DB.Where("key = ?", "MonitorStatus").Find(&status).Error; err != nil {
		global.LOG.Errorf("load monitor status from db failed, err: %v", err)
	}
	if status.Value == "Enable" {
		if err := global.DB.Where("key = ?", "MonitorInterval").Find(&interval).Error; err != nil {
			global.LOG.Errorf("load monitor interval from db failed, err: %v", err)
		}
		if err := service.StartMonitor(false, interval.Value); err != nil {
			global.LOG.Errorf("can not add monitor corn job: %s", err.Error())
		}
	}

	if _, err := global.Cron.AddJob("@daily", job.NewWebsiteJob()); err != nil {
		global.LOG.Errorf("can not add  website corn job: %s", err.Error())
	}
	if _, err := global.Cron.AddJob("@daily", job.NewSSLJob()); err != nil {
		global.LOG.Errorf("can not add  ssl corn job: %s", err.Error())
	}
	if _, err := global.Cron.AddJob(fmt.Sprintf("%v %v * * *", mathRand.Intn(60), mathRand.Intn(3)), job.NewAppStoreJob()); err != nil {
		global.LOG.Errorf("can not add  appstore corn job: %s", err.Error())
	}
	if _, err := global.Cron.AddJob("0 3 */31 * *", job.NewBackupJob()); err != nil {
		global.LOG.Errorf("can not add  backup token refresh corn job: %s", err.Error())
	}

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
		entryIDs, err := service.NewICronjobService().StartJob(&cronJobs[i], false)
		if err != nil {
			global.LOG.Errorf("start %s job %s failed, err: %v", cronJobs[i].Type, cronJobs[i].Name, err)
		}
		if err := repo.NewICronjobRepo().Update(cronJobs[i].ID, map[string]interface{}{"entry_ids": entryIDs}); err != nil {
			global.LOG.Errorf("update cronjob %s %s failed, err: %v", cronJobs[i].Type, cronJobs[i].Name, err)
		}
	}
	global.Cron.Start()
}

func syncBeforeStart() {
	var ntpSite model.Setting
	if err := global.DB.Where("key = ?", "NtpSite").Find(&ntpSite).Error; err != nil {
		global.LOG.Errorf("load ntp serve from db failed, err: %v", err)
	}
	if len(ntpSite.Value) == 0 {
		ntpSite.Value = "pool.ntp.org"
	}
	ntime, err := ntp.GetRemoteTime(ntpSite.Value)
	if err != nil {
		global.LOG.Errorf("load remote time with [%s] failed, err: %v", ntpSite.Value, err)
		return
	}
	ts := ntime.Format(constant.DateTimeLayout)
	if err := ntp.UpdateSystemTime(ts); err != nil {
		global.LOG.Errorf("failed to synchronize system time with [%s], err: %v", ntpSite.Value, err)
	}
	global.LOG.Debugf("synchronize system time with [%s] successful!", ntpSite.Value)
}
