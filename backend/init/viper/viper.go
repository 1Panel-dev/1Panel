package viper

import (
	"fmt"
	"path"

	"github.com/1Panel-dev/1Panel/backend/configs"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func Init() {
	baseDir := "/opt"
	v := viper.NewWithOptions()
	v.SetConfigName("app")
	v.SetConfigType("yaml")
	v.AddConfigPath(path.Dir(baseDir + "/1Panel/conf/app.yaml"))
	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
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
	global.CONF.System.DataDir = global.CONF.BaseDir + "/1Panel/data"
	global.CONF.System.Cache = global.CONF.BaseDir + "/1Panel/data/cache"
	global.CONF.System.Backup = global.CONF.BaseDir + "/1Panel/data/backup"
	global.CONF.System.DbPath = global.CONF.BaseDir + "/1Panel/data/db"
	global.CONF.System.LogPath = global.CONF.BaseDir + "/1Panel/log"
	global.Viper = v
}
