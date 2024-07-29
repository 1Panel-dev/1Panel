package server

import (
	"net"
	"net/http"
	"os"

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
		Handler: rootRouter,
	}
	if len(global.CurrentNode) == 0 || global.CurrentNode == "127.0.0.1" {
		_ = os.Remove("/tmp/agent.sock")
		listener, err := net.Listen("unix", "/tmp/agent.sock")
		if err != nil {
			panic(err)
		}
		_ = server.Serve(listener)
	} else {
		server.Addr = "0.0.0.0:9999"
		type tcpKeepAliveListener struct {
			*net.TCPListener
		}
		ln, err := net.Listen("tcp4", "0.0.0.0:9999")
		if err != nil {
			panic(err)
		}
		global.LOG.Info("listen at http://0.0.0.0:9999")
		if err := server.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)}); err != nil {
			panic(err)
		}
	}
}
