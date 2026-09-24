package hook

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/alert_push"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"gorm.io/gorm"
)

func Init() {
	global.LOG.Info("agent hook: init global data start")
	initGlobalData()
	global.LOG.Info("agent hook: init global data done")
	global.LOG.Info("agent hook: handle cronjob status start")
	handleCronjobStatus()
	global.LOG.Info("agent hook: handle cronjob status done")
	global.LOG.Info("agent hook: handle clam status start")
	handleClamStatus()
	global.LOG.Info("agent hook: handle clam status done")
	global.LOG.Info("agent hook: handle record status start")
	handleRecordStatus()
	global.LOG.Info("agent hook: handle record status done")
	global.LOG.Info("agent hook: handle snapshot status start")
	handleSnapStatus()
	global.LOG.Info("agent hook: handle snapshot status done")
	global.LOG.Info("agent hook: handle ollama model status start")
	handleOllamaModelStatus()
	global.LOG.Info("agent hook: handle ollama model status done")

	global.LOG.Info("agent hook: load local dir start")
	loadLocalDir()
	global.LOG.Info("agent hook: load local dir done")

	global.LOG.Info("agent hook: init docker config start")
	initDockerConf()
	global.LOG.Info("agent hook: init docker config done")
	global.LOG.Info("agent hook: init alert task start")
	initAlertTask()
	global.LOG.Info("agent hook: init alert task done")
	global.LOG.Info("agent hook: init monitor db start")
	initMonitorDB()
	global.LOG.Info("agent hook: init monitor db done")
}

func initGlobalData() {
	settingRepo := repo.NewISettingRepo()
	if _, err := settingRepo.GetValueByKey("SystemStatus"); err != nil {
		_ = settingRepo.Create("SystemStatus", "Free")
	}
	if err := settingRepo.Update("SystemStatus", "Free"); err != nil {
		global.LOG.Fatalf("init service before start failed, err: %v", err)
	}
	node, _ := xpack.MultiNodeProvider.LoadNodeInfo(false)
	if len(node.Version) != 0 {
		_ = settingRepo.Update("SystemVersion", node.Version)
	}
	global.CONF.Base.Version = node.Version
	global.CONF.Base.Edition, _ = settingRepo.GetValueByKey("Edition")
	global.CONF.Base.EncryptKey, _ = settingRepo.GetValueByKey("EncryptKey")
}

func handleSnapStatus() {
	_ = global.DB.Model(&model.Snapshot{}).Where("status = ?", "OnSaveData").
		Updates(map[string]interface{}{"status": constant.StatusSuccess}).Error

	_ = global.DB.Model(&model.Snapshot{}).Where("status = ?", constant.StatusWaiting).
		Updates(map[string]interface{}{
			"status":  constant.StatusFailed,
			"message": constant.InterruptedMsg,
		}).Error

	_ = global.DB.Model(&model.Snapshot{}).Where("recover_status = ?", constant.StatusWaiting).
		Updates(map[string]interface{}{
			"recover_status":  constant.StatusFailed,
			"recover_message": constant.InterruptedMsg,
		}).Error

	_ = global.DB.Model(&model.Snapshot{}).Where("rollback_status = ?", constant.StatusWaiting).
		Updates(map[string]interface{}{
			"rollback_status":  constant.StatusFailed,
			"rollback_message": constant.InterruptedMsg,
		}).Error
}

func handleCronjobStatus() {
	var jobRecords []model.JobRecords
	_ = global.DB.Model(&model.Cronjob{}).Where("is_executing = ?", true).Updates(map[string]interface{}{"is_executing": false}).Error
	_ = global.DB.Where("status = ?", constant.StatusWaiting).Find(&jobRecords).Error
	for _, record := range jobRecords {
		err := global.DB.Model(&model.JobRecords{}).Where("status = ?", constant.StatusWaiting).
			Updates(map[string]interface{}{
				"status":  constant.StatusFailed,
				"message": constant.InterruptedMsg,
			}).Error

		if err != nil {
			global.LOG.Errorf("Failed to update job ID: %v, Error:%v", record.ID, err)
			continue
		}

		var cronjob *model.Cronjob
		_ = global.DB.Where("id = ?", record.CronjobID).First(&cronjob).Error
		handleCronJobAlert(cronjob)
	}
}

func handleClamStatus() {
	_ = global.DB.Model(&model.Clam{}).Where("is_executing = ?", true).Updates(map[string]interface{}{"is_executing": false}).Error
	_ = global.DB.Model(&model.ClamRecord{}).Where("status = ?", constant.StatusWaiting).Updates(map[string]interface{}{
		"status":  constant.StatusFailed,
		"message": constant.InterruptedMsg,
	}).Error
}

func handleRecordStatus() {
	_ = global.DB.Model(&model.BackupRecord{}).Where("status = ?", constant.StatusWaiting).Updates(map[string]interface{}{
		"status":  constant.StatusFailed,
		"message": constant.InterruptedMsg,
	}).Error
}

func handleOllamaModelStatus() {
	_ = global.DB.Model(&model.OllamaModel{}).Where("status = ?", constant.StatusWaiting).Updates(map[string]interface{}{
		"status":  constant.StatusCanceled,
		"message": constant.InterruptedMsg,
	}).Error
}

func handleCronJobAlert(cronjob *model.Cronjob) {
	pushAlert := dto.PushAlert{
		TaskName:  cronjob.Name,
		AlertType: cronjob.Type,
		EntryID:   cronjob.ID,
		Param:     cronjob.Type,
	}
	_ = alert_push.PushAlert(pushAlert)
}

func loadLocalDir() {
	var account model.BackupAccount
	if err := global.DB.Where("`type` = ?", constant.Local).First(&account).Error; err != nil {
		global.LOG.Errorf("load local backup account info failed, err: %v", err)
		return
	}
	global.Dir.LocalBackupDir = account.BackupPath

	if _, err := os.Stat(account.BackupPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(account.BackupPath, os.ModePerm); err != nil {
			global.LOG.Errorf("mkdir %s failed, err: %v", account.BackupPath, err)
		}
	}
}

func initDockerConf() {
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return
	}
	if strings.Contains(dockerPath, "snap") {
		constant.DaemonJsonPath = "/var/snap/docker/current/config/daemon.json"
	}
}

func initAlertTask() {
	service.NewIAlertTaskHelper().ResetTask()
}

func initMonitorDB() {
	_ = global.MonitorDB.AutoMigrate(&model.MonitorBase{}, &model.MonitorNetwork{}, &model.MonitorIO{})
	_ = global.GPUMonitorDB.AutoMigrate(&model.MonitorGPU{})
	_ = global.TaskDB.AutoMigrate(&model.Task{})
	// building indexes on large monitor tables can take seconds, keep it off the startup path;
	// WAL mode leaves readers unblocked and busy_timeout covers the collector's inserts meanwhile
	go ensureMonitorIndexes()
}

func ensureMonitorIndexes() {
	indexes := []struct {
		db   *gorm.DB
		stmt string
	}{
		// created_at alone serves unfiltered range queries and retention cleanup
		{global.MonitorDB, "CREATE INDEX IF NOT EXISTS idx_monitor_bases_created ON monitor_bases(created_at)"},
		{global.MonitorDB, "CREATE INDEX IF NOT EXISTS idx_monitor_ios_created ON monitor_ios(created_at)"},
		{global.MonitorDB, "CREATE INDEX IF NOT EXISTS idx_monitor_networks_created ON monitor_networks(created_at)"},
		{global.GPUMonitorDB, "CREATE INDEX IF NOT EXISTS idx_monitor_gpus_created ON monitor_gpus(created_at)"},
		// (name, created_at) serves per-device range queries and distinct name lookups
		{global.MonitorDB, "CREATE INDEX IF NOT EXISTS idx_monitor_ios_name_created ON monitor_ios(name, created_at)"},
		{global.MonitorDB, "CREATE INDEX IF NOT EXISTS idx_monitor_networks_name_created ON monitor_networks(name, created_at)"},
		{global.GPUMonitorDB, "CREATE INDEX IF NOT EXISTS idx_monitor_gpus_product_created ON monitor_gpus(product_name, created_at)"},
	}
	start := time.Now()
	for _, index := range indexes {
		if err := index.db.Exec(index.stmt).Error; err != nil {
			global.LOG.Warnf("create monitor index failed, stmt: %s, err: %v", index.stmt, err)
		}
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		global.LOG.Infof("monitor indexes ready, took %s", elapsed)
	}
}
