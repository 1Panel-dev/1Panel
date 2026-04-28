package viper

import (
	"bytes"
	"fmt"
	"os"
	"path"

	"github.com/1Panel-dev/1Panel/core/cmd/server/conf"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/ctl_conf"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func Init() {
	baseDir := "/opt"
	port := "9999"
	mode := ""
	version := "v2.0.0"
	username, password, entrance, language, edition := "", "", "", "zh", ""
	v := viper.NewWithOptions()
	v.SetConfigType("yaml")

	config := global.ServerConfig{}
	if err := yaml.Unmarshal(conf.AppYaml, &config); err != nil {
		panic(err)
	}
	if config.Base.Mode != "" {
		mode = config.Base.Mode
	}
	_, err := os.Stat("/opt/1panel/conf/app.yaml")
	if mode == "dev" && err == nil {
		v.SetConfigName("app")
		v.AddConfigPath(path.Join("/opt/1panel/conf"))
		if err := v.ReadInConfig(); err != nil {
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
	} else {
		baseDir = ctl_conf.Load("BASE_DIR")
		port = ctl_conf.Load("ORIGINAL_PORT")
		version = ctl_conf.Load("ORIGINAL_VERSION")
		username = ctl_conf.Load("ORIGINAL_USERNAME")
		password = ctl_conf.Load("ORIGINAL_PASSWORD")
		entrance = ctl_conf.Load("ORIGINAL_ENTRANCE")
		language = ctl_conf.Load("LANGUAGE")
		edition = ctl_conf.LoadWithoutPanic("PANEL_EDITION")

		reader := bytes.NewReader(conf.AppYaml)
		if err := v.ReadConfig(reader); err != nil {
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
	}
	v.OnConfigChange(func(e fsnotify.Event) {
		if err := v.Unmarshal(&global.CONF); err != nil {
			panic(err)
		}
	})
	serverConfig := global.ServerConfig{}
	if err := v.Unmarshal(&serverConfig); err != nil {
		panic(err)
	}
	_, err = os.Stat("/opt/1panel/conf/app.yaml")
	if mode == "dev" && err == nil {
		if serverConfig.Base.InstallDir != "" {
			baseDir = serverConfig.Base.InstallDir
		}
		if serverConfig.Conn.Port != "" {
			port = serverConfig.Conn.Port
		}
		if serverConfig.Base.Version != "" {
			version = serverConfig.Base.Version
		}
		if serverConfig.Base.Username != "" {
			username = serverConfig.Base.Username
		}
		if serverConfig.Base.Password != "" {
			password = serverConfig.Base.Password
		}
		if serverConfig.Conn.Entrance != "" {
			entrance = serverConfig.Conn.Entrance
		}
	}

	global.CONF = serverConfig
	global.CONF.Base.InstallDir = baseDir
	global.CONF.Base.IsDemo = v.GetBool("base.is_demo")
	global.CONF.Base.IsFxplay = v.GetBool("base.is_fxplay")
	global.CONF.Base.IsOffLine = v.GetBool("base.is_offline")
	if edition == "intl" {
		global.CONF.Base.Edition = "intl"
	} else {
		global.CONF.Base.Edition = "cn"
	}
	global.CONF.Base.Version = version
	global.CONF.Base.Username = username
	global.CONF.Base.Password = password
	global.CONF.Base.Language = language
	global.CONF.Base.ChangeUserInfo = loadChangeInfo()
	global.CONF.Conn.Entrance = entrance
	global.CONF.Conn.Port = port
	global.Viper = v
}

func loadChangeInfo() string {
	return ctl_conf.LoadWithoutPanic("CHANGE_USER_INFO")
}
