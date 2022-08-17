package server

import (
	"encoding/gob"
	"fmt"
	"github.com/1Panel-dev/1Panel/init/cache"
	"github.com/1Panel-dev/1Panel/init/session"
	"github.com/1Panel-dev/1Panel/init/session/psession"
	"time"

	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/init/binary"
	"github.com/1Panel-dev/1Panel/init/db"
	"github.com/1Panel-dev/1Panel/init/log"
	"github.com/1Panel-dev/1Panel/init/migration"
	"github.com/1Panel-dev/1Panel/init/router"
	"github.com/1Panel-dev/1Panel/init/validator"
	"github.com/1Panel-dev/1Panel/init/viper"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
)

func Start() {
	viper.Init()
	log.Init()
	db.Init()
	migration.Init()
	validator.Init()
	gob.Register(psession.SessionUser{})
	cache.Init()
	session.Init()
	binary.StartTTY()
	gin.SetMode(global.CONF.System.Level)

	routers := router.Routers()
	address := fmt.Sprintf(":%d", global.CONF.System.Port)
	s := initServer(address, routers)
	global.LOG.Infof("server run success on %d", global.CONF.System.Port)
	if err := s.ListenAndServe(); err != nil {
		global.LOG.Error(err)
		panic(err)
	}
}

type server interface {
	ListenAndServe() error
}

func initServer(address string, router *gin.Engine) server {
	s := endless.NewServer(address, router)
	s.ReadHeaderTimeout = 20 * time.Second
	s.WriteTimeout = 20 * time.Second
	s.MaxHeaderBytes = 1 << 20
	return s
}
