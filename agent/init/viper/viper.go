package viper

import (
	"bytes"
	"fmt"
	"path"

	"github.com/1Panel-dev/1Panel/agent/cmd/server/conf"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func Init() {
	mode := ""
	fileOp := files.NewFileOp()
	v := viper.NewWithOptions()
	v.SetConfigType("yaml")

	config := global.ServerConfig{}
	if err := yaml.Unmarshal(conf.AppYaml, &config); err != nil {
		panic(err)
	}
	if config.Base.Mode != "" {
		mode = config.Base.Mode
	}
	if mode == "dev" && fileOp.Stat("/opt/1panel/conf/app.yaml") {
		v.SetConfigName("app")
		v.AddConfigPath(path.Join("/opt/1panel/conf"))
		if err := v.ReadInConfig(); err != nil {
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
	} else {
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

	global.CONF = serverConfig
	global.CONF.Base.IsDemo = v.GetBool("system.is_demo")

	initBaseInfo()
	global.Viper = v
}

func initBaseInfo() {
	nodeInfo, err := xpack.LoadNodeInfo(true)
	if err != nil {
		panic(err)
	}
	global.CONF.Base.InstallDir = nodeInfo.BaseDir
}
