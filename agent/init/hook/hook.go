package hook

import (
	"encoding/json"
	"os"
	"path"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/init/db"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
)

func Init() {
	settingRepo := repo.NewISettingRepo()
	if _, err := settingRepo.Get(settingRepo.WithByKey("SystemStatus")); err != nil {
		_ = settingRepo.Create("SystemStatus", "Free")
	}
	if err := settingRepo.Update("SystemStatus", "Free"); err != nil {
		global.LOG.Fatalf("init service before start failed, err: %v", err)
	}
	node, err := settingRepo.Get(settingRepo.WithByKey("CurrentNode"))
	if err != nil {
		global.LOG.Fatalf("load current node before start failed, err: %v", err)
	}
	baseDir, err := settingRepo.Get(settingRepo.WithByKey("BaseDir"))
	if err != nil {
		global.LOG.Fatalf("load base dir before start failed, err: %v", err)
	}
	global.CONF.System.BaseDir = baseDir.Value
	version, err := settingRepo.Get(settingRepo.WithByKey("SystemVersion"))
	if err != nil {
		global.LOG.Fatalf("load system version before start failed, err: %v", err)
	}
	global.CONF.System.Version = version.Value

	global.IsMaster = node.Value == "127.0.0.1" || len(node.Value) == 0
	if global.IsMaster {
		db.InitCoreDB()
	} else {
		masterAddr, err := settingRepo.Get(settingRepo.WithByKey("MasterAddr"))
		if err != nil {
			global.LOG.Fatalf("load master addr before start failed, err: %v", err)
		}
		global.CONF.System.MasterAddr = masterAddr.Value
	}

	handleCronjobStatus()
	handleSnapStatus()
	loadLocalDir()
	initDir()
	_ = initSSL()
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

func initDir() {
	composePath := path.Join(global.CONF.System.BaseDir, "1panel/docker/compose/")
	if _, err := os.Stat(composePath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(composePath, os.ModePerm); err != nil {
			global.LOG.Errorf("mkdir %s failed, err: %v", composePath, err)
			return
		}
	}
}

func initSSL() error {
	settingRepo := repo.NewISettingRepo()
	if _, err := os.Stat("/opt/1panel/nodeJson"); err != nil {
		return nil
	}
	type nodeInfo struct {
		ServerCrt   string `json:"serverCrt"`
		ServerKey   string `json:"serverKey"`
		CurrentNode string `json:"currentNode"`
	}
	nodeJson, err := os.ReadFile("/opt/1panel/nodeJson")
	if err != nil {
		return err
	}
	var node nodeInfo
	if err := json.Unmarshal(nodeJson, &node); err != nil {
		return err
	}
	itemKey, _ := encrypt.StringEncrypt(node.ServerKey)
	if err := settingRepo.Update("ServerKey", itemKey); err != nil {
		return err
	}
	itemCrt, _ := encrypt.StringEncrypt(node.ServerCrt)
	if err := settingRepo.Update("ServerCrt", itemCrt); err != nil {
		return err
	}
	if err := settingRepo.Update("CurrentNode", node.CurrentNode); err != nil {
		return err
	}
	global.IsMaster = node.CurrentNode == "127.0.0.1" || len(node.CurrentNode) == 0
	_ = os.Remove(("/opt/1panel/nodeJson"))
	return nil
}
