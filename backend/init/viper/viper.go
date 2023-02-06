package viper

import (
	"bytes"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/configs"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/cmd/server/conf"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func Init() {
	stdout, err := cmd.Exec("grep '^BASE_DIR=' /usr/bin/1pctl | cut -d'=' -f2")
	if err != nil {
		panic(err)
	}
	baseDir := strings.ReplaceAll(stdout, "\n", "")
	v := viper.NewWithOptions()
	v.SetConfigType("yaml")
	reader := bytes.NewReader(conf.AppYaml)
	if err := v.ReadConfig(reader); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
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
	global.CONF.BaseDir = baseDir
	global.CONF.System.DataDir = global.CONF.BaseDir + "/1panel/data"
	global.CONF.System.Cache = global.CONF.BaseDir + "/1panel/data/cache"
	global.CONF.System.Backup = global.CONF.BaseDir + "/1panel/data/backup"
	global.CONF.System.DbPath = global.CONF.BaseDir + "/1panel/data/db"
	global.CONF.System.LogPath = global.CONF.BaseDir + "/1panel/log"
	global.Viper = v
}
