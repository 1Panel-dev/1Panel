package viper

import (
	"1Panel/configs"
	"1Panel/global"
	"fmt"

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
		if err := v.Unmarshal(&global.Config); err != nil {
			panic(err)
		}
	})
	serverConfig := configs.ServerConfig{}
	if err := v.Unmarshal(&serverConfig); err != nil {
		panic(err)
	}
	global.Config = serverConfig
}
