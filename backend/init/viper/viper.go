package viper

import (
	"bytes"
	"fmt"
	"path"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/configs"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/1Panel-dev/1Panel/cmd/server/conf"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func Init() {
	baseDir := "/opt"
	port := "9999"
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
		stdout, err := cmd.Exec("grep '^BASE_DIR=' /usr/bin/1pctl | cut -d'=' -f2")
		if err != nil {
			panic(err)
		}
		baseDir = strings.ReplaceAll(stdout, "\n", "")
		if len(baseDir) == 0 {
			panic("error `BASE_DIR` find in /usr/bin/1pctl")
		}

		stdoutPort, err := cmd.Exec("grep '^PANEL_PORT=' /usr/bin/1pctl | cut -d'=' -f2")
		if err != nil {
			panic(err)
		}
		port = strings.ReplaceAll(stdoutPort, "\n", "")
		if len(port) == 0 {
			panic("error `PANEL_PORT` find in /usr/bin/1pctl")
		}

		if strings.HasSuffix(baseDir, "/") {
			baseDir = baseDir[:strings.LastIndex(baseDir, "/")]
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
	if mode == "dev" && fileOp.Stat("/opt/1panel/conf/app.yaml") && serverConfig.System.BaseDir != "" {
		baseDir = serverConfig.System.BaseDir
	}
	if mode == "dev" && fileOp.Stat("/opt/1panel/conf/app.yaml") && serverConfig.System.Port != "" {
		port = serverConfig.System.Port
	}

	global.CONF = serverConfig
	global.CONF.BaseDir = baseDir
	global.CONF.System.IsDemo = v.GetBool("system.is_demo")
	global.CONF.System.DataDir = global.CONF.BaseDir + "/1panel"
	global.CONF.System.Cache = global.CONF.System.DataDir + "/cache"
	global.CONF.System.Backup = global.CONF.System.DataDir + "/backup"
	global.CONF.System.DbPath = global.CONF.System.DataDir + "/db"
	global.CONF.System.LogPath = global.CONF.System.DataDir + "/log"
	global.CONF.System.TmpDir = global.CONF.System.DataDir + "/tmp"
	global.CONF.System.Port = port
	global.Viper = v
}
