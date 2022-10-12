package viper

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/configs"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/global"
	"path"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func Init() {
	v := viper.NewWithOptions()
	v.SetConfigName("app")
	v.SetConfigType("yml")
	v.AddConfigPath("/opt/1Panel/conf")
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
	InitDir()
}

func InitDir() {
	constant.DefaultDataDir = "/opt/1Panel/data"
	constant.ResourceDir = path.Join(constant.DefaultDataDir, "resource")
	constant.AppResourceDir = path.Join(constant.ResourceDir, "apps")
	constant.AppInstallDir = path.Join(constant.DefaultDataDir, "apps")
	constant.BackupDir = path.Join(constant.DefaultDataDir, "backup")
	constant.AppBackupDir = path.Join(constant.BackupDir, "apps")

	//TODO 创建文件夹
}
