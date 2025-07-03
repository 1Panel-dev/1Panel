package hook

import (
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/app/service"
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
	initDockerConf()
}

func handleUserInfo(tags string, settingRepo repo.ISettingRepo) {
	if len(tags) == 0 {
		return
	}
	settingMap := make(map[string]string)
	if tags == "use_existing" {
		settingMap["ServerPort"] = common.LoadParams("ORIGINAL_PORT")
		global.CONF.Conn.Port = settingMap["ServerPort"]
		settingMap["UserName"] = global.CONF.Base.Username
		settingMap["Password"] = global.CONF.Base.Password
		settingMap["SecurityEntrance"] = global.CONF.Conn.Entrance
		settingMap["SystemVersion"] = common.LoadParams("ORIGINAL_VERSION")
		global.CONF.Base.Version = settingMap["SystemVersion"]
		settingMap["Language"] = global.CONF.Base.Language
	}
	if tags == "all" {
		settingMap["UserName"] = common.RandStrAndNum(10)
		settingMap["Password"] = common.RandStrAndNum(10)
		settingMap["SecurityEntrance"] = common.RandStrAndNum(10)
	}
	if strings.Contains(global.CONF.Base.ChangeUserInfo, "username") {
		settingMap["UserName"] = common.RandStrAndNum(10)
	}
	if strings.Contains(global.CONF.Base.ChangeUserInfo, "password") {
		settingMap["Password"] = common.RandStrAndNum(10)
	}
	if strings.Contains(global.CONF.Base.ChangeUserInfo, "entrance") {
		settingMap["SecurityEntrance"] = common.RandStrAndNum(10)
	}
	for key, val := range settingMap {
		if len(val) == 0 {
			continue
		}
		if key == "Password" {
			val, _ = encrypt.StringEncrypt(val)
		}
		if err := settingRepo.Update(key, val); err != nil {
			global.LOG.Errorf("update %s before start failed, err: %v", key, err)
		}
	}

	_, _ = cmd.RunDefaultWithStdoutBashCf("%s sed -i '/CHANGE_USER_INFO=%v/d' /usr/local/bin/1pctl", cmd.SudoHandleCmd(), global.CONF.Base.ChangeUserInfo)
	_, _ = cmd.RunDefaultWithStdoutBashCf("%s sed -i -e 's#ORIGINAL_PASSWORD=.*#ORIGINAL_PASSWORD=**********#g' /usr/local/bin/1pctl", cmd.SudoHandleCmd())
}

func generateKey() {
	if err := service.NewISettingService().GenerateRSAKey(); err != nil {
		global.LOG.Errorf("generate rsa key error : %s", err.Error())
	}
}

func initDockerConf() {
	stdout, err := cmd.RunDefaultWithStdoutBashC("which docker")
	if err != nil {
		return
	}
	dockerPath := stdout
	if strings.Contains(dockerPath, "snap") {
		constant.DaemonJsonPath = "/var/snap/docker/current/config/daemon.json"
	}
}
