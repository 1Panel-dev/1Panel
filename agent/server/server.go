package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/1Panel-dev/1Panel/agent/constant"

	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/cron"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/init/app"
	"github.com/1Panel-dev/1Panel/agent/init/business"
	"github.com/1Panel-dev/1Panel/agent/init/cache"
	"github.com/1Panel-dev/1Panel/agent/init/db"
	"github.com/1Panel-dev/1Panel/agent/init/dir"
	"github.com/1Panel-dev/1Panel/agent/init/hook"
	"github.com/1Panel-dev/1Panel/agent/init/log"
	"github.com/1Panel-dev/1Panel/agent/init/migration"
	"github.com/1Panel-dev/1Panel/agent/init/router"
	"github.com/1Panel-dev/1Panel/agent/init/validator"
	"github.com/1Panel-dev/1Panel/agent/init/viper"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"

	_ "net/http/pprof"

	"github.com/gin-gonic/gin"
)

func Start() {
	viper.Init()
	dir.Init()
	i18n.Init()
	log.Init()
	db.Init()
	cache.Init()
	migration.Init()
	app.Init()
	validator.Init()
	gin.SetMode("debug")
	cron.Run()
	hook.Init()
	InitOthers()

	rootRouter := router.Routers()

	server := &http.Server{
		Handler: rootRouter,
	}

	if global.IsMaster {
		_ = os.Remove("/etc/1panel/agent.sock")
		_ = os.Mkdir("/etc/1panel", constant.DirPerm)
		listener, err := net.Listen("unix", "/etc/1panel/agent.sock")
		if err != nil {
			panic(err)
		}
		business.Init()
		_ = server.Serve(listener)
		return
	} else {
		server.Addr = fmt.Sprintf("0.0.0.0:%s", global.CONF.Base.Port)
		settingRepo := repo.NewISettingRepo()
		certItem, err := settingRepo.Get(settingRepo.WithByKey("ServerCrt"))
		if err != nil {
			panic(err)
		}
		cert, _ := encrypt.StringDecrypt(certItem.Value)
		keyItem, err := settingRepo.Get(settingRepo.WithByKey("ServerKey"))
		if err != nil {
			panic(err)
		}
		key, _ := encrypt.StringDecrypt(keyItem.Value)
		tlsCert, err := tls.X509KeyPair([]byte(cert), []byte(key))
		if err != nil {
			fmt.Printf("failed to load X.509 key pair: %s\n", err)
			return
		}
		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
			ClientAuth:   tls.RequireAnyClientCert,
		}
		business.Init()
		global.LOG.Infof("listen at https://0.0.0.0:%s", global.CONF.Base.Port)
		if err := server.ListenAndServeTLS("", ""); err != nil {
			panic(err)
		}
	}
}
