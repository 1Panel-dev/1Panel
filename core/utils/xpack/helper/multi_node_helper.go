package helper

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/init/proxy"
	"github.com/1Panel-dev/1Panel/core/utils/ssh"
	"github.com/gin-gonic/gin"
)

type multiNodeHelper struct{}

func NewIMultiNodeProvider() *multiNodeHelper {
	return &multiNodeHelper{}
}

func (m *multiNodeHelper) Proxy(c *gin.Context, currentNode string) {
	if currentNode != "local" && currentNode != "" {
		c.Next()
		return
	}
	defer func() {
		if err := recover(); err != nil && err != http.ErrAbortHandler {
			global.LOG.Debug(err)
		}
	}()
	proxy.LocalAgentProxy.ServeHTTP(c.Writer, c.Request)
	c.Abort()
}

func (m *multiNodeHelper) ProxyDocker(proxyURL string) error { return nil }

func (m *multiNodeHelper) UpdateGroup(name string, group, newGroup uint) error { return nil }

func (m *multiNodeHelper) CheckBackupUsed(name string) error { return nil }

func (m *multiNodeHelper) LoadRequestTransport() *http.Transport {
	return &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		IdleConnTimeout:       15 * time.Second,
	}
}

func (m *multiNodeHelper) LoadNodeInfo(currentNode string) (*ssh.ConnInfo, string, error) {
	return nil, "", nil
}

func (m *multiNodeHelper) Sync(dataType string) error { return nil }

func (m *multiNodeHelper) AutoUpgradeWithMaster() {}
