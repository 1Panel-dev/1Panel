package hook

import (
	"path"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
)

func Init() {
	initWithNodeJson()
	initGlobalData()
	handleCronjobStatus()
	handleSnapStatus()
	loadLocalDir()
}

func initGlobalData() {
	settingRepo := repo.NewISettingRepo()
	if _, err := settingRepo.Get(settingRepo.WithByKey("SystemStatus")); err != nil {
		_ = settingRepo.Create("SystemStatus", "Free")
	}
	if err := settingRepo.Update("SystemStatus", "Free"); err != nil {
		global.LOG.Fatalf("init service before start failed, err: %v", err)
	}

	global.CONF.System.BaseDir, _ = settingRepo.GetValueByKey("BaseDir")
	global.CONF.System.Version, _ = settingRepo.GetValueByKey("SystemVersion")
	global.CONF.System.EncryptKey, _ = settingRepo.GetValueByKey("EncryptKey")
	currentNode, _ := settingRepo.GetValueByKey("CurrentNode")

	global.IsMaster = currentNode == "127.0.0.1" || len(currentNode) == 0
	if global.IsMaster {
		global.CoreDB = common.LoadDBConnByPath(path.Join(global.CONF.System.DbPath, "core.db"), "core")
	} else {
		global.CONF.System.MasterAddr, _ = settingRepo.GetValueByKey("MasterAddr")
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
	_ = global.DB.Model(&model.JobRecords{}).Where("status = ?", constant.StatusWaiting).
		Updates(map[string]interface{}{
			"status":  constant.StatusFailed,
			"message": "the task was interrupted due to the restart of the 1panel service",
		}).Error
}

func loadLocalDir() {
	var vars string
	if global.IsMaster {
		var account model.BackupAccount
		if err := global.CoreDB.Where("id = 1").First(&account).Error; err != nil {
			global.LOG.Errorf("load local backup account info failed, err: %v", err)
			return
		}
		vars = account.Vars
	} else {
		account, _, err := service.NewBackupClientWithID(1)
		if err != nil {
			global.LOG.Errorf("load local backup account info failed, err: %v", err)
			return
		}
		vars = account.Vars
	}
	localDir, err := service.LoadLocalDirByStr(vars)
	if err != nil {
		global.LOG.Errorf("load local backup dir failed, err: %v", err)
		return
	}
	global.CONF.System.Backup = localDir
}

func initWithNodeJson() {
	if global.IsMaster {
		return
	}
	isLocal, nodeInfo, err := xpack.LoadNodeInfo()
	if err != nil {
		global.LOG.Errorf("load new node info failed, err: %v", err)
		return
	}
	if isLocal {
		return
	}

	settingRepo := repo.NewISettingRepo()
	itemKey, _ := encrypt.StringEncrypt(nodeInfo.ServerKey)
	if err := settingRepo.Update("ServerKey", itemKey); err != nil {
		global.LOG.Errorf("update server key failed, err: %v", err)
		return
	}
	itemCrt, _ := encrypt.StringEncrypt(nodeInfo.ServerCrt)
	if err := settingRepo.Update("ServerCrt", itemCrt); err != nil {
		global.LOG.Errorf("update server crt failed, err: %v", err)
		return
	}
	if err := settingRepo.Update("CurrentNode", nodeInfo.CurrentNode); err != nil {
		global.LOG.Errorf("update current node failed, err: %v", err)
		return
	}
	if err := settingRepo.Update("SystemVersion", nodeInfo.Version); err != nil {
		global.LOG.Errorf("update system version failed, err: %v", err)
		return
	}
	if err := settingRepo.Update("BaseDir", nodeInfo.BaseDir); err != nil {
		global.LOG.Errorf("update base dir failed, err: %v", err)
		return
	}
	if err := settingRepo.Update("MasterAddr", nodeInfo.MasterAddr); err != nil {
		global.LOG.Errorf("update master addr failed, err: %v", err)
		return
	}
}
