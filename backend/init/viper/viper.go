package viper

import (
	"bytes"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"gopkg.in/yaml.v3"
	"path"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/utils/cmd"

	"github.com/1Panel-dev/1Panel/backend/configs"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/cmd/server/conf"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func Init() {
	baseDir := "/opt"
	mode := "dev"
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
	if mode != "release" {
		if !fileOp.Stat("/opt/1panel/conf/app.yaml") {
			panic("conf file is not exist")
		}
		v.SetConfigName("app")
		v.AddConfigPath(path.Join("/opt/1panel/conf"))
		if err := v.ReadInConfig(); err != nil {
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
	} else {
		stdout, err := cmd.Exec("grep '^BASE_DIR=' /usr/bin/1pctl | cut -d'=' -f2")
		if err != nil {
			panic(err)
		}
		baseDir = strings.ReplaceAll(stdout, "\n", "")
		if len(baseDir) == 0 {
			panic("error `BASE_DIR` find in /usr/bin/1pctl")
		}
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
	if mode != "release" && serverConfig.System.BaseDir != "" {
		baseDir = serverConfig.System.BaseDir
	}
	global.CONF = serverConfig
	global.CONF.BaseDir = baseDir
	global.CONF.System.DataDir = global.CONF.BaseDir + "/1panel"
	global.CONF.System.Cache = global.CONF.System.DataDir + "/cache"
	global.CONF.System.Backup = global.CONF.System.DataDir + "/backup"
	global.CONF.System.DbPath = global.CONF.System.DataDir + "/db"
	global.CONF.System.LogPath = global.CONF.System.DataDir + "/log"
	global.Viper = v
}
