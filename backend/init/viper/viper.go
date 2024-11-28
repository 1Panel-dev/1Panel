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
	version := "v1.0.0"
	username, password, entrance, language := "", "", "", "zh"
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
		port = loadParams("ORIGINAL_PORT")
		version = loadParams("ORIGINAL_VERSION")
		username = loadParams("ORIGINAL_USERNAME")
		password = loadParams("ORIGINAL_PASSWORD")
		entrance = loadParams("ORIGINAL_ENTRANCE")
		language = loadParams("LANGUAGE")

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
		if serverConfig.System.Port != "" {
			port = serverConfig.System.Port
		}
		if serverConfig.System.Version != "" {
			version = serverConfig.System.Version
		}
		if serverConfig.System.Username != "" {
			username = serverConfig.System.Username
		}
		if serverConfig.System.Password != "" {
			password = serverConfig.System.Password
		}
		if serverConfig.System.Entrance != "" {
			entrance = serverConfig.System.Entrance
		}
		if serverConfig.System.IsIntl {
			language = "en"
		}
	}

	global.CONF = serverConfig
	global.CONF.System.BaseDir = baseDir
	global.CONF.System.IsDemo = v.GetBool("system.is_demo")
	global.CONF.System.IsIntl = v.GetBool("system.is_intl")
	global.CONF.System.DataDir = path.Join(global.CONF.System.BaseDir, "1panel")
	global.CONF.System.Cache = path.Join(global.CONF.System.DataDir, "cache")
	global.CONF.System.Backup = path.Join(global.CONF.System.DataDir, "backup")
	global.CONF.System.DbPath = path.Join(global.CONF.System.DataDir, "db")
	global.CONF.System.LogPath = path.Join(global.CONF.System.DataDir, "log")
	global.CONF.System.TmpDir = path.Join(global.CONF.System.DataDir, "tmp")
	global.CONF.System.Port = port
	global.CONF.System.Version = version
	global.CONF.System.Username = username
	global.CONF.System.Password = password
	global.CONF.System.Entrance = entrance
	global.CONF.System.Language = language
	global.CONF.System.ChangeUserInfo = loadChangeInfo()
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

func loadChangeInfo() string {
	stdout, err := cmd.Exec("grep '^CHANGE_USER_INFO=' /usr/local/bin/1pctl | cut -d'=' -f2")
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(stdout, "\n", "")
}
