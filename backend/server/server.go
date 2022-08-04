package server

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/init/db"
	"github.com/1Panel-dev/1Panel/init/log"
	"github.com/1Panel-dev/1Panel/init/migration"
	"github.com/1Panel-dev/1Panel/init/router"
	"github.com/1Panel-dev/1Panel/init/viper"
	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"time"
)

func Start() {
	viper.Init()
	log.Init()
	db.Init()
	migration.Init()
	routers := router.Routers()
	address := fmt.Sprintf(":%d", global.Config.System.Port)
	s := initServer(address, routers)
	global.Logger.Info(fmt.Sprintf("server run success on %d", global.Config.System.Port))
	if err := s.ListenAndServe(); err != nil {
		global.Logger.Error(err)
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
