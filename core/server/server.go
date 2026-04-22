package server

import (
	"bufio"
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/init/auth"
	"github.com/1Panel-dev/1Panel/core/init/db"
	"github.com/1Panel-dev/1Panel/core/init/geo"
	"github.com/1Panel-dev/1Panel/core/init/log"
	"github.com/1Panel-dev/1Panel/core/init/migration"
	"github.com/1Panel-dev/1Panel/core/init/proxy"
	"github.com/1Panel-dev/1Panel/core/init/run"
	"github.com/gin-gonic/gin"
	"github.com/soheilhy/cmux"

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
	"github.com/1Panel-dev/1Panel/core/utils/re"
)

func Start() {
	re.Init()
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
	hook.Init()
	InitOthers()

	run.Init()
	proxy.Init()

	rootRouter := router.Routers()

	if global.CONF.Base.Mode != "stable" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	global.IPTracker = auth.NewIPTracker()

	tcpItem := "tcp4"
	if global.CONF.Conn.Ipv6 == constant.StatusEnable {
		tcpItem = "tcp"
		global.CONF.Conn.BindAddress = fmt.Sprintf("[%s]", global.CONF.Conn.BindAddress)
	}
	server := &http.Server{
		Addr:              global.CONF.Conn.BindAddress + ":" + global.CONF.Conn.Port,
		Handler:           rootRouter,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       600 * time.Second,
		WriteTimeout:      600 * time.Second,
		IdleTimeout:       240 * time.Second,
	}
	ln, err := net.Listen(tcpItem, server.Addr)
	if err != nil {
		panic(err)
	}
	type tcpKeepAliveListener struct {
		*net.TCPListener
	}
	if global.CONF.Conn.SSL == constant.StatusEnable {
		constant.CertStore.Store(loadCert())

		server.TLSConfig = &tls.Config{
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return constant.CertStore.Load().(*tls.Certificate), nil
			},
		}
		global.LOG.Infof("listen at https://%s:%s [%s]", global.CONF.Conn.BindAddress, global.CONF.Conn.Port, tcpItem)

		if err := server.ServeTLS(tcpKeepAliveListener{ln.(*net.TCPListener)}, "", ""); err != nil {
			panic(err)
		}
		return
	} else if global.CONF.Conn.SSL == constant.StatusMux {
		constant.CertStore.Store(loadCert())

		server.TLSConfig = &tls.Config{
			NextProtos: []string{"h2", "http/1.1"},
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return constant.CertStore.Load().(*tls.Certificate), nil
			},
		}

		m := cmux.New(ln)

		httpsL := m.Match(cmux.TLS())
		httpL := m.Match(cmux.HTTP1Fast())
		anyL := m.Match(cmux.Any())

		go func() {
			if err := server.Serve(tls.NewListener(httpsL, server.TLSConfig)); err != nil {
				global.LOG.Errorf("HTTPS Serve Error: %v", err)
			}
		}()

		go func() {
			for {
				conn, err := httpL.Accept()
				if err != nil {
					return
				}
				go handleMuxHttpConn(conn)
			}
		}()

		go func() {
			for {
				conn, err := anyL.Accept()
				if err != nil {
					return
				}
				conn.Close()
			}
		}()

		global.LOG.Infof("listen at mux (http/https)://%s:%s [%s]", global.CONF.Conn.BindAddress, global.CONF.Conn.Port, tcpItem)

		if err := m.Serve(); err != nil {
			panic(err)
		}
		return
	} else {
		global.LOG.Infof("listen at http://%s:%s [%s]", global.CONF.Conn.BindAddress, global.CONF.Conn.Port, tcpItem)
		if err := server.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)}); err != nil {
			panic(err)
		}
		return
	}
}

func loadCert() *tls.Certificate {
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
	return &cert
}

func handleMuxHttpConn(conn net.Conn) {
	defer conn.Close()

	req, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		return
	}

	if req.Host == "" {
		return
	}

	ua := req.Header.Get("User-Agent")
	if ua == "" {
		return
	}

	switch req.Method {
	case http.MethodGet, http.MethodHead, http.MethodPost:
	default:
		return
	}

	if len(req.RequestURI) > 4096 {
		return
	}

	if !strings.HasPrefix(req.URL.Path, "/") {
		return
	}

	target := "https://" + req.Host + req.URL.RequestURI()

	resp := &http.Response{
		StatusCode: http.StatusTemporaryRedirect,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}
	resp.Header.Set("Location", target)
	resp.Header.Set("Connection", "close")

	_ = resp.Write(conn)
	return
}
