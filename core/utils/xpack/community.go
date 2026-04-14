//go:build !xpack && !xpackee

package xpack

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"

	baseDto "github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/init/proxy"
	"github.com/1Panel-dev/1Panel/core/init/session/psession"
	"github.com/1Panel-dev/1Panel/core/utils/ssh"
	"github.com/1Panel-dev/1Panel/core/xpack/app/model"
	"github.com/gin-gonic/gin"
)

func Proxy(c *gin.Context, currentNode string) {
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

func CoreRBACMiddlewares() []gin.HandlerFunc { return nil }

func ProxyDocker(proxyURL string) error { return nil }

func UpdateGroup(name string, group, newGroup uint) error { return nil }

func CheckBackupUsed(name string) error { return nil }

func LoadRequestTransport() *http.Transport {
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

func LoadNodeInfo(currentNode string) (*ssh.ConnInfo, string, error) {
	return nil, "", nil
}

func Sync(dataType string) error { return nil }

func AutoUpgradeWithMaster() {}

func Login(_ *gin.Context, _ baseDto.Login, _ string) (*baseDto.UserLoginInfo, string, error) {
	return nil, "", errors.New("not xpackee build")
}

func LoadSessionTimeout(sessionUser psession.SessionUser, defaultTTL int) int { return defaultTTL }

func RemoveBindNode(nodeID uint) error { return nil }

func CheckLicenseStatus(isXpack bool, licenseID, nodeID uint) error {
	return nil
}
func LoadLicenseByNodeID(nodeID uint) (bool, model.License, error) {
	return false, model.License{}, nil
}
func Bind(node model.Node, license *model.License, withInit, withSync, withDryRun, withDockerRestart bool) error {
	return nil
}
func BindFree(node model.Node, licenseID uint) error {
	return nil
}
func Unbind(node model.Node, license model.License, withReset, withSync, withDockerRestart bool) error {
	return nil
}
func UnbindFree(node model.Node, licenseID uint) error {
	return nil
}
