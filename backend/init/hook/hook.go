package hook

import (
	"encoding/base64"

	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
)

func Init() {
	settingRepo := repo.NewISettingRepo()
	portSetting, err := settingRepo.Get(settingRepo.WithByKey("ServerPort"))
	if err != nil {
		global.LOG.Errorf("load service port from setting failed, err: %v", err)
	}
	global.CONF.System.Port = portSetting.Value
	encryptSetting, err := settingRepo.Get(settingRepo.WithByKey("EncryptKey"))
	if err != nil {
		global.LOG.Errorf("load service encrypt key from setting failed, err: %v", err)
	}
	global.CONF.System.EncryptKey = encryptSetting.Value
	sslSetting, err := settingRepo.Get(settingRepo.WithByKey("SSL"))
	if err != nil {
		global.LOG.Errorf("load service ssl from setting failed, err: %v", err)
	}
	global.CONF.System.SSL = sslSetting.Value

	OneDriveID, err := settingRepo.Get(settingRepo.WithByKey("OneDriveID"))
	if err != nil {
		global.LOG.Errorf("load onedrive info from setting failed, err: %v", err)
	}
	idItem, _ := base64.StdEncoding.DecodeString(OneDriveID.Value)
	global.CONF.System.OneDriveID = string(idItem)
	OneDriveSc, err := settingRepo.Get(settingRepo.WithByKey("OneDriveSc"))
	if err != nil {
		global.LOG.Errorf("load onedrive info from setting failed, err: %v", err)
	}
	scItem, _ := base64.StdEncoding.DecodeString(OneDriveSc.Value)
	global.CONF.System.OneDriveSc = string(scItem)

	if _, err := settingRepo.Get(settingRepo.WithByKey("SystemStatus")); err != nil {
		_ = settingRepo.Create("SystemStatus", "Free")
	}
	if err := settingRepo.Update("SystemStatus", "Free"); err != nil {
		global.LOG.Fatalf("init service before start failed, err: %v", err)
	}

	if global.CONF.System.ChangeUserInfo {
		if err := settingRepo.Update("UserName", common.RandStrAndNum(10)); err != nil {
			global.LOG.Fatalf("init username before start failed, err: %v", err)
		}
		pass, _ := encrypt.StringEncrypt(common.RandStrAndNum(10))
		if err := settingRepo.Update("Password", pass); err != nil {
			global.LOG.Fatalf("init password before start failed, err: %v", err)
		}
		if err := settingRepo.Update("SecurityEntrance", common.RandStrAndNum(10)); err != nil {
			global.LOG.Fatalf("init entrance before start failed, err: %v", err)
		}

		sudo := cmd.SudoHandleCmd()
		_, _ = cmd.Execf("%s sed -i '/CHANGE_USER_INFO=true/d' /usr/local/bin/1pctl", sudo)
	}

	handleSnapStatus()
}

func handleSnapStatus() {
	snapRepo := repo.NewISnapshotRepo()
	snaps, _ := snapRepo.GetList()
	for _, snap := range snaps {
		if snap.Status == "OnSaveData" {
			_ = snapRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusSuccess})
		}
		if snap.Status == constant.StatusWaiting {
			_ = snapRepo.Update(snap.ID, map[string]interface{}{"status": constant.StatusFailed, "message": "the task was interrupted due to the restart of the 1panel service"})
		}
	}

	status, _ := snapRepo.GetStatusList()
	for _, statu := range status {
		updatas := make(map[string]interface{})
		if statu.Panel == constant.StatusRunning {
			updatas["panel"] = constant.StatusFailed
		}
		if statu.PanelCtl == constant.StatusRunning {
			updatas["panel_ctl"] = constant.StatusFailed
		}
		if statu.PanelService == constant.StatusRunning {
			updatas["panel_service"] = constant.StatusFailed
		}
		if statu.PanelInfo == constant.StatusRunning {
			updatas["panel_info"] = constant.StatusFailed
		}
		if statu.DaemonJson == constant.StatusRunning {
			updatas["daemon_json"] = constant.StatusFailed
		}
		if statu.AppData == constant.StatusRunning {
			updatas["app_data"] = constant.StatusFailed
		}
		if statu.PanelData == constant.StatusRunning {
			updatas["panel_data"] = constant.StatusFailed
		}
		if statu.BackupData == constant.StatusRunning {
			updatas["backup_data"] = constant.StatusFailed
		}
		if statu.Compress == constant.StatusRunning {
			updatas["compress"] = constant.StatusFailed
		}
		if statu.Upload == constant.StatusUploading {
			updatas["upload"] = constant.StatusFailed
		}
		if len(updatas) != 0 {
			_ = snapRepo.UpdateStatus(statu.ID, updatas)
		}
	}
}
