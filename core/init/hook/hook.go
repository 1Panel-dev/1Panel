package hook

import (
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/app/service"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/cmd"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
)

func Init() {
	settingRepo := repo.NewISettingRepo()
	commonRepo := repo.NewICommonRepo()
	masterSetting, err := settingRepo.Get(commonRepo.WithByKey("MasterAddr"))
	if err != nil {
		global.LOG.Errorf("load master addr from setting failed, err: %v", err)
	}
	global.CONF.System.MasterAddr = masterSetting.Value
	portSetting, err := settingRepo.Get(commonRepo.WithByKey("ServerPort"))
	if err != nil {
		global.LOG.Errorf("load service port from setting failed, err: %v", err)
	}
	global.CONF.System.Port = portSetting.Value
	ipv6Setting, err := settingRepo.Get(commonRepo.WithByKey("Ipv6"))
	if err != nil {
		global.LOG.Errorf("load ipv6 status from setting failed, err: %v", err)
	}
	global.CONF.System.Ipv6 = ipv6Setting.Value
	bindAddressSetting, err := settingRepo.Get(commonRepo.WithByKey("BindAddress"))
	if err != nil {
		global.LOG.Errorf("load bind address from setting failed, err: %v", err)
	}
	global.CONF.System.BindAddress = bindAddressSetting.Value
	sslSetting, err := settingRepo.Get(commonRepo.WithByKey("SSL"))
	if err != nil {
		global.LOG.Errorf("load service ssl from setting failed, err: %v", err)
	}
	global.CONF.System.SSL = sslSetting.Value
	versionSetting, err := settingRepo.Get(commonRepo.WithByKey("SystemVersion"))
	if err != nil {
		global.LOG.Errorf("load version from setting failed, err: %v", err)
	}
	global.CONF.System.Version = versionSetting.Value

	if _, err := settingRepo.Get(commonRepo.WithByKey("SystemStatus")); err != nil {
		_ = settingRepo.Create("SystemStatus", "Free")
	}
	if err := settingRepo.Update("SystemStatus", "Free"); err != nil {
		global.LOG.Fatalf("init service before start failed, err: %v", err)
	}

	handleUserInfo(global.CONF.System.ChangeUserInfo, settingRepo)
	loadLocalDir()
}

func handleUserInfo(tags string, settingRepo repo.ISettingRepo) {
	if len(tags) == 0 {
		return
	}
	if tags == "all" {
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
		return
	}
	if strings.Contains(global.CONF.System.ChangeUserInfo, "username") {
		if err := settingRepo.Update("UserName", common.RandStrAndNum(10)); err != nil {
			global.LOG.Fatalf("init username before start failed, err: %v", err)
		}
	}
	if strings.Contains(global.CONF.System.ChangeUserInfo, "password") {
		pass, _ := encrypt.StringEncrypt(common.RandStrAndNum(10))
		if err := settingRepo.Update("Password", pass); err != nil {
			global.LOG.Fatalf("init password before start failed, err: %v", err)
		}
	}
	if strings.Contains(global.CONF.System.ChangeUserInfo, "entrance") {
		if err := settingRepo.Update("SecurityEntrance", common.RandStrAndNum(10)); err != nil {
			global.LOG.Fatalf("init entrance before start failed, err: %v", err)
		}
	}

	sudo := cmd.SudoHandleCmd()
	_, _ = cmd.Execf("%s sed -i '/CHANGE_USER_INFO=%v/d' /usr/local/bin/1pctl", sudo, global.CONF.System.ChangeUserInfo)
}

func loadLocalDir() {
	var backup model.BackupAccount
	_ = global.DB.Where("type = ?", "LOCAL").First(&backup).Error
	if backup.ID == 0 {
		global.LOG.Errorf("no such backup account `%s` in db", "LOCAL")
		return
	}
	dir, err := service.LoadLocalDirByStr(backup.Vars)
	if err != nil {
		global.LOG.Errorf("load local backup dir failed,err: %v", err)
		return
	}
	global.CONF.System.BackupDir = dir
}
