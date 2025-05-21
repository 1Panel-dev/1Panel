package server

import (
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"github.com/1Panel-dev/1Panel/core/init/proxy"
	"net"
	"net/http"
	"os"
	"path"

	"github.com/1Panel-dev/1Panel/core/init/db"
	"github.com/1Panel-dev/1Panel/core/init/geo"
	"github.com/1Panel-dev/1Panel/core/init/log"
	"github.com/1Panel-dev/1Panel/core/init/migration"
	"github.com/1Panel-dev/1Panel/core/init/run"

	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/i18n"
	"github.com/1Panel-dev/1Panel/core/init/cron"
	"github.com/1Panel-dev/1Panel/core/init/hook"
	"github.com/1Panel-dev/1Panel/core/init/router"
	"github.com/1Panel-dev/1Panel/core/init/session"
	"github.com/1Panel-dev/1Panel/core/init/session/psession"
	"github.com/1Panel-dev/1Panel/core/init/validator"
	"github.com/1Panel-dev/1Panel/core/init/viper"

	"github.com/gin-gonic/gin"
)

func Start() {
	viper.Init()
	log.Init()
	db.Init()
	migration.Init()
	i18n.Init()
	validator.Init()
	geo.Init()
	gob.Register(psession.SessionUser{})
	cron.Init()
	session.Init()
	gin.SetMode("debug")
	hook.Init()
	InitOthers()

	run.Init()
	proxy.Init()

	rootRouter := router.Routers()

	tcpItem := "tcp4"
	if global.CONF.Conn.Ipv6 == constant.StatusEnable {
		tcpItem = "tcp"
		global.CONF.Conn.BindAddress = fmt.Sprintf("[%s]", global.CONF.Conn.BindAddress)
	}
	server := &http.Server{
		Addr:    global.CONF.Conn.BindAddress + ":" + global.CONF.Conn.Port,
		Handler: rootRouter,
	}
	ln, err := net.Listen(tcpItem, server.Addr)
	if err != nil {
		panic(err)
	}
	type tcpKeepAliveListener struct {
		*net.TCPListener
	}
	if global.CONF.Conn.SSL == constant.StatusEnable {
		certPath := path.Join(global.CONF.Base.InstallDir, "1panel/secret/server.crt")
		keyPath := path.Join(global.CONF.Base.InstallDir, "1panel/secret/server.key")
		certificate, err := os.ReadFile(certPath)
		if err != nil {
			panic(err)
		}
		key, err := os.ReadFile(keyPath)
		if err != nil {
			panic(err)
		}
		cert, err := tls.X509KeyPair(certificate, key)
		if err != nil {
			panic(err)
		}
		constant.CertStore.Store(&cert)

		server.TLSConfig = &tls.Config{
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return constant.CertStore.Load().(*tls.Certificate), nil
			},
		}
		global.LOG.Infof("listen at https://%s:%s [%s]", global.CONF.Conn.BindAddress, global.CONF.Conn.Port, tcpItem)

		if err := server.ServeTLS(tcpKeepAliveListener{ln.(*net.TCPListener)}, "", ""); err != nil {
			panic(err)
		}
	} else {
		global.LOG.Infof("listen at http://%s:%s [%s]", global.CONF.Conn.BindAddress, global.CONF.Conn.Port, tcpItem)
		if err := server.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)}); err != nil {
			panic(err)
		}
	}
}
