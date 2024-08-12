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
	global.CurrentNode = node.Value

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

	snapRepo := repo.NewISnapshotRepo()

	status, _ := snapRepo.GetStatusList()
	for _, item := range status {
		updates := make(map[string]interface{})
		if item.Panel == constant.StatusRunning {
			updates["panel"] = constant.StatusFailed
		}
		if item.PanelInfo == constant.StatusRunning {
			updates["panel_info"] = constant.StatusFailed
		}
		if item.DaemonJson == constant.StatusRunning {
			updates["daemon_json"] = constant.StatusFailed
		}
		if item.AppData == constant.StatusRunning {
			updates["app_data"] = constant.StatusFailed
		}
		if item.PanelData == constant.StatusRunning {
			updates["panel_data"] = constant.StatusFailed
		}
		if item.BackupData == constant.StatusRunning {
			updates["backup_data"] = constant.StatusFailed
		}
		if item.Compress == constant.StatusRunning {
			updates["compress"] = constant.StatusFailed
		}
		if item.Upload == constant.StatusUploading {
			updates["upload"] = constant.StatusFailed
		}
		if len(updates) != 0 {
			_ = snapRepo.UpdateStatus(item.ID, updates)
		}
	}
}

func handleCronjobStatus() {
	_ = global.DB.Model(&model.JobRecords{}).Where("status = ?", constant.StatusWaiting).
		Updates(map[string]interface{}{
			"status":  constant.StatusFailed,
			"message": "the task was interrupted due to the restart of the 1panel service",
		}).Error
}

func loadLocalDir() {
	account, _, err := service.NewBackupClientWithID(1)
	if err != nil {
		global.LOG.Errorf("load local backup account info failed, err: %v", err)
	}
	global.CONF.System.Backup, err = service.LoadLocalDirByStr(account.Vars)
	if err != nil {
		global.LOG.Errorf("load local backup dir failed, err: %v", err)
	}
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
	global.CurrentNode = node.CurrentNode
	_ = os.Remove(("/opt/1panel/nodeJson"))
	return nil
}
