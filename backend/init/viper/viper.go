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
	version := ""
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
		baseDir = loadParams("BASE_DIR")
		port = loadParams("PANEL_PORT")
		version = loadParams("ORIGINAL_INSTALLED_VERSION")

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
	if mode == "dev" && fileOp.Stat("/opt/1panel/conf/app.yaml") && serverConfig.System.Version != "" {
		version = serverConfig.System.Version
	}

	global.CONF = serverConfig
	global.CONF.System.BaseDir = baseDir
	global.CONF.System.IsDemo = v.GetBool("system.is_demo")
	global.CONF.System.DataDir = global.CONF.System.BaseDir + "/1panel"
	global.CONF.System.Cache = global.CONF.System.DataDir + "/cache"
	global.CONF.System.Backup = global.CONF.System.DataDir + "/backup"
	global.CONF.System.DbPath = global.CONF.System.DataDir + "/db"
	global.CONF.System.LogPath = global.CONF.System.DataDir + "/log"
	global.CONF.System.TmpDir = global.CONF.System.DataDir + "/tmp"
	global.CONF.System.Port = port
	global.CONF.System.Version = version
	global.Viper = v
}

func loadParams(param string) string {
	stdout, err := cmd.Execf("grep '^%s=' /usr/bin/1pctl | cut -d'=' -f2", param)
	if err != nil {
		panic(err)
	}
	baseDir := strings.ReplaceAll(stdout, "\n", "")
	if len(baseDir) == 0 {
		panic(fmt.Sprintf("error `%s` find in /usr/bin/1pctl", param))
	}
	return baseDir
}
