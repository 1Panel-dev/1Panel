package service

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/glebarez/sqlite"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCronjobUpdateSchedulesPendingJob(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "cronjob.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Cronjob{}, &model.Alert{}); err != nil {
		t.Fatal(err)
	}
	oldDB, oldAlertDB, oldLog, oldI18n, oldCron := global.DB, global.AlertDB, global.LOG, global.I18n, global.Cron
	global.DB, global.AlertDB = db, db
	global.LOG = logrus.New()
	global.LOG.SetOutput(io.Discard)
	global.I18n = goi18n.NewLocalizer(goi18n.NewBundle(language.English), "en")
	global.Cron = cron.New()
	t.Cleanup(func() {
		global.DB, global.AlertDB, global.LOG, global.I18n, global.Cron = oldDB, oldAlertDB, oldLog, oldI18n, oldCron
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	// Import stores a job whose associations are missing as Pending; editing it enables it.
	job := model.Cronjob{Name: "imported", Type: "shell", Spec: "30 1 * * *", ScriptMode: "input", Script: "echo ok", Status: constant.StatusPending}
	if err := db.Create(&job).Error; err != nil {
		t.Fatal(err)
	}
	req := dto.CronjobOperate{ID: job.ID, Name: job.Name, Type: job.Type, Spec: job.Spec, ScriptMode: job.ScriptMode, Script: job.Script, Executor: "bash"}
	if err := NewICronjobService().Update(job.ID, req, "admin"); err != nil {
		t.Fatal(err)
	}

	var stored model.Cronjob
	if err := db.First(&stored, job.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Status != constant.StatusEnable || stored.EntryIDs == "" || len(global.Cron.Entries()) != 1 {
		t.Fatalf("status = %q, entry_ids = %q, scheduled entries = %d; want an enabled, scheduled job", stored.Status, stored.EntryIDs, len(global.Cron.Entries()))
	}
}
