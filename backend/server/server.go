package server

import (
	"net"
	"net/http"

	"github.com/1Panel-dev/1Panel/backend/i18n"

	"github.com/1Panel-dev/1Panel/backend/init/app"
	"github.com/1Panel-dev/1Panel/backend/init/business"

	"github.com/1Panel-dev/1Panel/backend/cron"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/init/db"
	"github.com/1Panel-dev/1Panel/backend/init/hook"
	"github.com/1Panel-dev/1Panel/backend/init/log"
	"github.com/1Panel-dev/1Panel/backend/init/migration"
	"github.com/1Panel-dev/1Panel/backend/init/router"
	"github.com/1Panel-dev/1Panel/backend/init/validator"
	"github.com/1Panel-dev/1Panel/backend/init/viper"

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
		Addr:    ":9999",
		Handler: rootRouter,
	}
	ln, err := net.Listen("tcp4", ":9999")
	if err != nil {
		panic(err)
	}
	type tcpKeepAliveListener struct {
		*net.TCPListener
	}

	global.LOG.Info("listen at http://0.0.0.0:9999")
	if err := server.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)}); err != nil {
		panic(err)
	}
}
