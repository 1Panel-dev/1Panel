package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"

	"github.com/gin-gonic/gin"

	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/cron"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/init/app"
	"github.com/1Panel-dev/1Panel/agent/init/business"
	"github.com/1Panel-dev/1Panel/agent/init/cache"
	"github.com/1Panel-dev/1Panel/agent/init/db"
	"github.com/1Panel-dev/1Panel/agent/init/dir"
	"github.com/1Panel-dev/1Panel/agent/init/firewall"
	"github.com/1Panel-dev/1Panel/agent/init/hook"
	"github.com/1Panel-dev/1Panel/agent/init/lang"
	"github.com/1Panel-dev/1Panel/agent/init/log"
	"github.com/1Panel-dev/1Panel/agent/init/migration"
	"github.com/1Panel-dev/1Panel/agent/init/router"
	"github.com/1Panel-dev/1Panel/agent/init/validator"
	"github.com/1Panel-dev/1Panel/agent/init/viper"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
)

const (
	masterSocketDir          = "/etc/1panel"
	masterSocketPath         = masterSocketDir + "/agent.sock"
	masterSocketDirPerm      = 0o700
	masterSocketFilePerm     = 0o600
	masterSocketDirPermMask  = 0o077
	masterSocketFilePermMask = 0o077
)

func prepareMasterSocketDir(dir string) error {
	if err := os.MkdirAll(dir, masterSocketDirPerm); err != nil {
		return fmt.Errorf("create master socket dir %s failed: %w", dir, err)
	}
	if err := os.Chmod(dir, masterSocketDirPerm); err != nil {
		return fmt.Errorf("chmod master socket dir %s failed: %w", dir, err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("stat master socket dir %s failed: %w", dir, err)
	}
	if info.Mode().Perm()&masterSocketDirPermMask != 0 {
		return fmt.Errorf("master socket dir %s permission %#o is too permissive", dir, info.Mode().Perm())
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		if int(stat.Uid) != os.Geteuid() {
			return fmt.Errorf(
				"master socket dir %s owner uid %d does not match current process uid %d",
				dir, stat.Uid, os.Geteuid(),
			)
		}
	}
	return nil
}

func secureMasterSocket(sockPath string) error {
	if err := os.Chmod(sockPath, masterSocketFilePerm); err != nil {
		return fmt.Errorf("chmod master socket %s failed: %w", sockPath, err)
	}
	info, err := os.Stat(sockPath)
	if err != nil {
		return fmt.Errorf("stat master socket %s failed: %w", sockPath, err)
	}
	if info.Mode().Perm()&masterSocketFilePermMask != 0 {
		return fmt.Errorf("master socket %s permission %#o is too permissive", sockPath, info.Mode().Perm())
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil
	}
	if int(stat.Uid) != os.Geteuid() {
		return fmt.Errorf(
			"master socket %s owner uid %d does not match current process uid %d",
			sockPath, stat.Uid, os.Geteuid(),
		)
	}
	return nil
}

func Start() {
	re.Init()
	viper.Init()
	dir.Init()
	log.Init()
	db.Init()
	migration.Init()
	i18n.Init()
	cache.Init()
	app.Init()
	lang.Init()
	validator.Init()
	cron.Run()
	hook.Init()
	go firewall.Init()
	InitOthers()

	rootRouter := router.Routers()

	server := &http.Server{
		Handler: rootRouter,
	}

	if global.CONF.Base.Mode != "stable" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	if global.IsMaster {
		if err := prepareMasterSocketDir(masterSocketDir); err != nil {
			panic(err)
		}
		_ = os.Remove(masterSocketPath)
		listener, err := net.Listen("unix", masterSocketPath)
		if err != nil {
			panic(err)
		}
		if err := secureMasterSocket(masterSocketPath); err != nil {
			_ = listener.Close()
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
			ClientAuth:   tls.RequireAndVerifyClientCert,
		}
		caItem, _ := settingRepo.GetValueByKey("RootCrt")
		if len(caItem) != 0 {
			caCertPool := x509.NewCertPool()
			rootCrt, _ := encrypt.StringDecrypt(caItem)
			caCertPool.AppendCertsFromPEM([]byte(rootCrt))
			server.TLSConfig.ClientCAs = caCertPool
		}
		business.Init()
		global.LOG.Infof("listen at https://0.0.0.0:%s", global.CONF.Base.Port)
		if err := server.ListenAndServeTLS("", ""); err != nil {
			panic(err)
		}
	}
}
