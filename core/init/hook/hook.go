package hook

import (
	"os/exec"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/app/service"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/1Panel-dev/1Panel/core/utils/ctl_conf"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
)

func Init() {
	settingRepo := repo.NewISettingRepo()
	global.CONF.Conn.Port, _ = settingRepo.GetValueByKey("ServerPort")
	global.CONF.Conn.Ipv6, _ = settingRepo.GetValueByKey("Ipv6")
	global.CONF.Base.Edition, _ = settingRepo.GetValueByKey("Edition")
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
		settingMap["ServerPort"] = ctl_conf.Load("ORIGINAL_PORT")
		global.CONF.Conn.Port = settingMap["ServerPort"]
		settingMap["UserName"] = global.CONF.Base.Username
		settingMap["Password"] = global.CONF.Base.Password
		settingMap["SecurityEntrance"] = global.CONF.Conn.Entrance
		settingMap["SystemVersion"] = ctl_conf.Load("ORIGINAL_VERSION")
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
	if global.CONF.Base.IsEnterprise {
		if len(settingMap["UserName"]) != 0 || len(settingMap["Password"]) != 0 {
			if err := xpack.AuthProvider.ResetSuperAdminUser(settingMap["UserName"], settingMap["Password"]); err != nil {
				global.LOG.Fatalf("reset enterprise super admin failed, err: %v", err)
				return
			}
		}
	} else {
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
	}

	_ = ctl_conf.RemoveValueFromFile("/usr/local/bin/1pctl", "CHANGE_USER_INFO", global.CONF.Base.ChangeUserInfo)
	_ = ctl_conf.UpdateInFile("/usr/local/bin/1pctl", "ORIGINAL_PASSWORD", "**********")
}

func generateKey() {
	if err := service.NewISettingService().GenerateRSAKey(); err != nil {
		global.LOG.Errorf("generate rsa key error : %s", err.Error())
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
