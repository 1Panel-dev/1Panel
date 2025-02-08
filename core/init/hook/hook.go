package hook

import (
	"github.com/1Panel-dev/1Panel/core/app/service"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/cmd"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
)

func Init() {
	settingRepo := repo.NewISettingRepo()
	global.CONF.Conn.Port, _ = settingRepo.GetValueByKey("ServerPort")
	global.CONF.Conn.Ipv6, _ = settingRepo.GetValueByKey("Ipv6")
	global.Api.ApiInterfaceStatus, _ = settingRepo.GetValueByKey("ApiInterfaceStatus")
	if global.Api.ApiInterfaceStatus == constant.StatusEnable {
		global.Api.ApiKey, _ = settingRepo.GetValueByKey("ApiKey")
		global.Api.IpWhiteList, _ = settingRepo.GetValueByKey("IpWhiteList")
		global.Api.ApiKeyValidityTime, _ = settingRepo.GetValueByKey("ApiKeyValidityTime")
	}
	global.CONF.Conn.BindAddress, _ = settingRepo.GetValueByKey("BindAddress")
	global.CONF.Conn.SSL, _ = settingRepo.GetValueByKey("SSL")
	global.CONF.Base.Version, _ = settingRepo.GetValueByKey("SystemVersion")
	if err := settingRepo.Update("SystemStatus", "Free"); err != nil {
		global.LOG.Fatalf("init service before start failed, err: %v", err)
	}

	handleUserInfo(global.CONF.Base.ChangeUserInfo, settingRepo)

	generateKey()
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
	if strings.Contains(global.CONF.Base.ChangeUserInfo, "username") {
		if err := settingRepo.Update("UserName", common.RandStrAndNum(10)); err != nil {
			global.LOG.Fatalf("init username before start failed, err: %v", err)
		}
	}
	if strings.Contains(global.CONF.Base.ChangeUserInfo, "password") {
		pass, _ := encrypt.StringEncrypt(common.RandStrAndNum(10))
		if err := settingRepo.Update("Password", pass); err != nil {
			global.LOG.Fatalf("init password before start failed, err: %v", err)
		}
	}
	if strings.Contains(global.CONF.Base.ChangeUserInfo, "entrance") {
		if err := settingRepo.Update("SecurityEntrance", common.RandStrAndNum(10)); err != nil {
			global.LOG.Fatalf("init entrance before start failed, err: %v", err)
		}
	}

	sudo := cmd.SudoHandleCmd()
	_, _ = cmd.Execf("%s sed -i '/CHANGE_USER_INFO=%v/d' /usr/local/bin/1pctl", sudo, global.CONF.Base.ChangeUserInfo)
}

func generateKey() {
	if err := service.NewISettingService().GenerateRSAKey(); err != nil {
		global.LOG.Errorf("generate rsa key error : %s", err.Error())
	}
}
