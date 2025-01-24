package hook

import (
	"path"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
)

func Init() {
	initGlobalData()
	handleCronjobStatus()
	handleSnapStatus()
}

func initGlobalData() {
	settingRepo := repo.NewISettingRepo()
	if _, err := settingRepo.Get(settingRepo.WithByKey("SystemStatus")); err != nil {
		_ = settingRepo.Create("SystemStatus", "Free")
	}
	if err := settingRepo.Update("SystemStatus", "Free"); err != nil {
		global.LOG.Fatalf("init service before start failed, err: %v", err)
	}
	global.CONF.Base.EncryptKey, _ = settingRepo.GetValueByKey("EncryptKey")
	_ = service.NewISettingService().ReloadConn()
	if global.IsMaster {
		global.CoreDB = common.LoadDBConnByPath(path.Join(global.Dir.DbDir, "core.db"), "core")
	}
}

func handleSnapStatus() {
	msgFailed := "the task was interrupted due to the restart of the 1panel service"
	_ = global.DB.Model(&model.Snapshot{}).Where("status = ?", "OnSaveData").
		Updates(map[string]interface{}{"status": constant.StatusSuccess}).Error

	_ = global.DB.Model(&model.Snapshot{}).Where("status = ?", constant.StatusWaiting).
		Updates(map[string]interface{}{
			"status":  constant.StatusFailed,
			"message": msgFailed,
		}).Error

	_ = global.DB.Model(&model.Snapshot{}).Where("recover_status = ?", constant.StatusWaiting).
		Updates(map[string]interface{}{
			"recover_status":  constant.StatusFailed,
			"recover_message": msgFailed,
		}).Error

	_ = global.DB.Model(&model.Snapshot{}).Where("rollback_status = ?", constant.StatusWaiting).
		Updates(map[string]interface{}{
			"rollback_status":  constant.StatusFailed,
			"rollback_message": msgFailed,
		}).Error
}

func handleCronjobStatus() {
	var jobRecords []model.JobRecords
	_ = global.DB.Where("status = ?", constant.StatusWaiting).Find(&jobRecords).Error
	for _, record := range jobRecords {
		err := global.DB.Model(&model.JobRecords{}).Where("status = ?", constant.StatusWaiting).
			Updates(map[string]interface{}{
				"status":  constant.StatusFailed,
				"message": "the task was interrupted due to the restart of the 1panel service",
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

func handleCronJobAlert(cronjob *model.Cronjob) {
	pushAlert := dto.PushAlert{
		TaskName:  cronjob.Name,
		AlertType: cronjob.Type,
		EntryID:   cronjob.ID,
		Param:     cronjob.Type,
	}
	err := xpack.PushAlert(pushAlert)
	if err != nil {
		global.LOG.Errorf("cronjob alert push failed, err: %v", err)
		return
	}
}
