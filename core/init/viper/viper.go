package viper

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/1Panel-dev/1Panel/core/cmd/server/conf"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/cmd"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func Init() {
	baseDir := "/opt"
	port := "9999"
	mode := ""
	version := "v2.0.0"
	username, password, entrance, language := "", "", "", "zh"
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
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
	} else {
		baseDir = common.LoadParams("BASE_DIR")
		port = common.LoadParams("ORIGINAL_PORT")
		version = common.LoadParams("ORIGINAL_VERSION")
		username = common.LoadParams("ORIGINAL_USERNAME")
		password = common.LoadParams("ORIGINAL_PASSWORD")
		entrance = common.LoadParams("ORIGINAL_ENTRANCE")
		language = common.LoadParams("LANGUAGE")

		reader := bytes.NewReader(conf.AppYaml)
		if err := v.ReadConfig(reader); err != nil {
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
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
		if serverConfig.Base.IsIntl {
			language = "en"
		}
	}

	global.CONF = serverConfig
	global.CONF.Base.InstallDir = baseDir
	global.CONF.Base.IsDemo = v.GetBool("base.is_demo")
	global.CONF.Base.IsIntl = v.GetBool("base.is_intl")
	global.CONF.Base.IsOffLine = v.GetBool("base.is_offline")
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
	stdout, err := cmd.RunDefaultWithStdoutBashC("grep '^CHANGE_USER_INFO=' /usr/local/bin/1pctl | cut -d'=' -f2")
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(stdout, "\n", "")
}
