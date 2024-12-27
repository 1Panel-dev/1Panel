package viper

import (
	"bytes"
	"fmt"
	"path"

	"github.com/1Panel-dev/1Panel/agent/cmd/server/conf"
	"github.com/1Panel-dev/1Panel/agent/configs"
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

	config := configs.ServerConfig{}
	if err := yaml.Unmarshal(conf.AppYaml, &config); err != nil {
		panic(err)
	}
	if config.System.Mode != "" {
		mode = config.System.Mode
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
	serverConfig := configs.ServerConfig{}
	if err := v.Unmarshal(&serverConfig); err != nil {
		panic(err)
	}

	global.CONF = serverConfig
	global.CONF.System.IsDemo = v.GetBool("system.is_demo")

	initDir()
	global.Viper = v
}

func initDir() {
	_, nodeInfo, err := xpack.LoadNodeInfo()
	if err != nil {
		panic(err)
	}
	global.CONF.System.BaseDir = nodeInfo.BaseDir

	fileOp := files.NewFileOp()
	_, _ = fileOp.CreateDirWithPath(true, path.Join(global.CONF.System.BaseDir, "1panel/docker/compose/"))
	global.CONF.System.DataDir, _ = fileOp.CreateDirWithPath(true, path.Join(global.CONF.System.BaseDir, "1panel"))
	global.CONF.System.Cache, _ = fileOp.CreateDirWithPath(true, path.Join(global.CONF.System.DataDir, "cache"))
	global.CONF.System.DbPath, _ = fileOp.CreateDirWithPath(true, path.Join(global.CONF.System.DataDir, "db"))
	global.CONF.System.LogPath, _ = fileOp.CreateDirWithPath(true, path.Join(global.CONF.System.DataDir, "log"))
	global.CONF.System.TmpDir, _ = fileOp.CreateDirWithPath(true, path.Join(global.CONF.System.DataDir, "tmp"))
}
