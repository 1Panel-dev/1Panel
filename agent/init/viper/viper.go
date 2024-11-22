package viper

import (
	"bytes"
	"fmt"
	"path"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/cmd/server/conf"
	"github.com/1Panel-dev/1Panel/agent/configs"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func Init() {
	baseDir := "/opt"
	mode := ""
	version := "v1.0.0"
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
	if mode == "dev" && fileOp.Stat("/opt/1panel/conf/app.yaml") {
		if serverConfig.System.BaseDir != "" {
			baseDir = serverConfig.System.BaseDir
		}
		if serverConfig.System.Version != "" {
			version = serverConfig.System.Version
		}
	}

	global.CONF = serverConfig
	global.CONF.System.BaseDir = baseDir
	global.CONF.System.IsDemo = v.GetBool("system.is_demo")
	global.CONF.System.DataDir = path.Join(global.CONF.System.BaseDir, "1panel")
	global.CONF.System.Cache = path.Join(global.CONF.System.DataDir, "cache")
	global.CONF.System.Backup = path.Join(global.CONF.System.DataDir, "backup")
	global.CONF.System.DbPath = path.Join(global.CONF.System.DataDir, "db")
	global.CONF.System.LogPath = path.Join(global.CONF.System.DataDir, "log")
	global.CONF.System.TmpDir = path.Join(global.CONF.System.DataDir, "tmp")
	global.CONF.System.Version = version
	global.Viper = v
}

func loadParams(param string) string {
	stdout, err := cmd.Execf("grep '^%s=' /usr/local/bin/1pctl | cut -d'=' -f2", param)
	if err != nil {
		panic(err)
	}
	info := strings.ReplaceAll(stdout, "\n", "")
	if len(info) == 0 || info == `""` {
		panic(fmt.Sprintf("error `%s` find in /usr/local/bin/1pctl", param))
	}
	return info
}
