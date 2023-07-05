package server

import (
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/1Panel-dev/1Panel/backend/init/app"
	"github.com/1Panel-dev/1Panel/backend/init/business"

	"github.com/1Panel-dev/1Panel/backend/cron"
	"github.com/1Panel-dev/1Panel/backend/init/cache"
	"github.com/1Panel-dev/1Panel/backend/init/session"
	"github.com/1Panel-dev/1Panel/backend/init/session/psession"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/init/db"
	"github.com/1Panel-dev/1Panel/backend/init/hook"
	"github.com/1Panel-dev/1Panel/backend/init/log"
	"github.com/1Panel-dev/1Panel/backend/init/migration"
	"github.com/1Panel-dev/1Panel/backend/init/router"
	"github.com/1Panel-dev/1Panel/backend/init/validator"
	"github.com/1Panel-dev/1Panel/backend/init/viper"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
)

func Start() {
	viper.Init()
	log.Init()
	app.Init()
	db.Init()
	migration.Init()
	validator.Init()
	gob.Register(psession.SessionUser{})
	cache.Init()
	session.Init()
	gin.SetMode("debug")
	cron.Run()
	business.Init()
	hook.Init()

	rootRouter := router.Routers()
	address := fmt.Sprintf(":%s", global.CONF.System.Port)
	s := endless.NewServer(address, rootRouter)
	s.ReadHeaderTimeout = 20 * time.Second
	s.WriteTimeout = 60 * time.Second
	s.MaxHeaderBytes = 1 << 20

	if global.CONF.System.SSL == "disable" {
		global.LOG.Infof("server run success on %s with http", global.CONF.System.Port)
		if err := s.ListenAndServe(); err != nil {
			global.LOG.Error(err)
			panic(err)
		}
	} else {
		certificate, err := os.ReadFile(global.CONF.System.BaseDir + "/1panel/secret/server.crt")
		if err != nil {
			panic(err)
		}
		key, err := os.ReadFile(global.CONF.System.BaseDir + "/1panel/secret/server.key")
		if err != nil {
			panic(err)
		}
		cert, err := tls.X509KeyPair(certificate, key)
		if err != nil {
			panic(err)
		}
		s := &http.Server{
			Addr:    address,
			Handler: rootRouter,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		}

		global.LOG.Infof("server run success on %s with https", global.CONF.System.Port)
		if err := s.ListenAndServeTLS("", ""); err != nil {
			global.LOG.Error(err)
			panic(err)
		}
	}
}
