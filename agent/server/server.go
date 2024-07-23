package server

import (
	"net"
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/cron"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/init/app"
	"github.com/1Panel-dev/1Panel/agent/init/business"
	"github.com/1Panel-dev/1Panel/agent/init/db"
	"github.com/1Panel-dev/1Panel/agent/init/hook"
	"github.com/1Panel-dev/1Panel/agent/init/log"
	"github.com/1Panel-dev/1Panel/agent/init/migration"
	"github.com/1Panel-dev/1Panel/agent/init/router"
	"github.com/1Panel-dev/1Panel/agent/init/validator"
	"github.com/1Panel-dev/1Panel/agent/init/viper"

	"github.com/gin-gonic/gin"
)

func Start() {
	viper.Init()
	i18n.Init()
	log.Init()
	db.Init()
	migration.Init()
	app.Init()
	validator.Init()
	gin.SetMode("debug")
	cron.Run()
	InitOthers()
	business.Init()
	hook.Init()

	rootRouter := router.Routers()

	server := &http.Server{
		Addr:    "0.0.0.0:9998",
		Handler: rootRouter,
	}
	ln, err := net.Listen("tcp4", "0.0.0.0:9998")
	if err != nil {
		panic(err)
	}
	type tcpKeepAliveListener struct {
		*net.TCPListener
	}

	global.LOG.Info("listen at http://0.0.0.0:9998")
	if err := server.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)}); err != nil {
		panic(err)
	}
}
