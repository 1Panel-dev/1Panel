package server

import (
	"1Panel/global"
	"1Panel/init/db"
	"1Panel/init/log"
	"1Panel/init/migration"
	"1Panel/init/router"
	"1Panel/init/validator"
	"1Panel/init/viper"
	"fmt"
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
)

func Start() {
	viper.Init()
	log.Init()
	db.Init()
	migration.Init()
	validator.Init()
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
